package memdb

import (
	"io"
	"runtime"
	"sync"
	"time"

	"github.com/ncruces/go-sqlite3"
	"github.com/ncruces/go-sqlite3/vfs"
)

// Must be a multiple of 64K (the largest page size).
const sectorSize = 65536

type memVFS struct{}

func (memVFS) Open(name string, flags vfs.OpenFlag) (vfs.File, vfs.OpenFlag, error) {
	// For simplicity, we do not support reading or writing data
	// across "sector" boundaries.
	//
	// This is not a problem for most SQLite file types:
	// - databases, which only do page aligned reads/writes;
	// - temp journals, as used by the sorter, which does the same:
	//   https://github.com/sqlite/sqlite/blob/b74eb0/src/vdbesort.c#L409-L412
	//
	// We refuse to open all other file types,
	// but returning OPEN_MEMORY means SQLite won't ask us to.
	const types = vfs.OPEN_MAIN_DB |
		vfs.OPEN_TEMP_DB |
		vfs.OPEN_TEMP_JOURNAL
	if flags&types == 0 {
		// notest // OPEN_MEMORY
		return nil, flags, sqlite3.CANTOPEN
	}

	// A shared database has a name that begins with "/".
	shared := len(name) > 1 && name[0] == '/'

	var db *memDB
	if shared {
		name = name[1:]
		memoryMtx.Lock()
		defer memoryMtx.Unlock()
		db = memoryDBs[name]
	}
	if db == nil {
		if flags&vfs.OPEN_CREATE == 0 {
			return nil, flags, sqlite3.CANTOPEN
		}
		db = &memDB{name: name}
	}
	if shared {
		db.refs++ // +checklocksforce: memoryMtx is held
		memoryDBs[name] = db
	}

	return &memFile{
		memDB:    db,
		readOnly: flags&vfs.OPEN_READONLY != 0,
	}, flags | vfs.OPEN_MEMORY, nil
}

func (memVFS) Delete(name string, dirSync bool) error {
	return sqlite3.IOERR_DELETE
}

func (memVFS) Access(name string, flag vfs.AccessFlag) (bool, error) {
	return false, nil
}

func (memVFS) FullPathname(name string) (string, error) {
	return name, nil
}

type memDB struct {
	name string

	// +checklocks:dataMtx
	data []*[sectorSize]byte
	// +checklocks:dataMtx
	size int64

	// +checklocks:memoryMtx
	refs int32

	shared   int32 // +checklocks:lockMtx
	pending  bool  // +checklocks:lockMtx
	reserved bool  // +checklocks:lockMtx

	lockMtx sync.Mutex
	dataMtx sync.RWMutex
}

func (m *memDB) release() {
	memoryMtx.Lock()
	defer memoryMtx.Unlock()
	if m.refs--; m.refs == 0 && m == memoryDBs[m.name] {
		delete(memoryDBs, m.name)
	}
}

type memFile struct {
	*memDB
	lock     vfs.LockLevel
	readOnly bool
}

var (
	// Ensure these interfaces are implemented:
	_ vfs.FileLockState = &memFile{}
	_ vfs.FileSizeHint  = &memFile{}
)

func (m *memFile) Close() error {
	m.release()
	return m.Unlock(vfs.LOCK_NONE)
}

func (m *memFile) ReadAt(b []byte, off int64) (n int, err error) {
	m.dataMtx.RLock()
	defer m.dataMtx.RUnlock()

	if off >= m.size {
		return 0, io.EOF
	}

	base := off / sectorSize
	rest := off % sectorSize
	have := int64(sectorSize)
	if base == int64(len(m.data))-1 {
		have = modRoundUp(m.size, sectorSize)
	}
	n = copy(b, (*m.data[base])[rest:have])
	if n < len(b) {
		// notest // assume reads are page aligned
		return 0, io.ErrNoProgress
	}
	return n, nil
}

func (m *memFile) WriteAt(b []byte, off int64) (n int, err error) {
	m.dataMtx.Lock()
	defer m.dataMtx.Unlock()

	base := off / sectorSize
	rest := off % sectorSize
	for base >= int64(len(m.data)) {
		m.data = append(m.data, new([sectorSize]byte))
	}
	n = copy((*m.data[base])[rest:], b)
	if n < len(b) {
		// notest // assume writes are page aligned
		return n, io.ErrShortWrite
	}
	if size := off + int64(len(b)); size > m.size {
		m.size = size
	}
	return n, nil
}

func (m *memFile) Truncate(size int64) error {
	m.dataMtx.Lock()
	defer m.dataMtx.Unlock()
	return m.truncate(size)
}

// +checklocks:m.dataMtx
func (m *memFile) truncate(size int64) error {
	if size < m.size {
		base := size / sectorSize
		rest := size % sectorSize
		if rest != 0 {
			clear((*m.data[base])[rest:])
		}
	}
	sectors := divRoundUp(size, sectorSize)
	for sectors > int64(len(m.data)) {
		m.data = append(m.data, new([sectorSize]byte))
	}
	clear(m.data[sectors:])
	m.data = m.data[:sectors]
	m.size = size
	return nil
}

func (m *memFile) Sync(flag vfs.SyncFlag) error {
	return nil
}

func (m *memFile) Size() (int64, error) {
	m.dataMtx.RLock()
	defer m.dataMtx.RUnlock()
	return m.size, nil
}

const spinWait = 25 * time.Microsecond

func (m *memFile) Lock(lock vfs.LockLevel) error {
	if m.lock >= lock {
		return nil
	}

	if m.readOnly && lock >= vfs.LOCK_RESERVED {
		return sqlite3.IOERR_LOCK
	}

	m.lockMtx.Lock()
	defer m.lockMtx.Unlock()

	switch lock {
	case vfs.LOCK_SHARED:
		if m.pending {
			return sqlite3.BUSY
		}
		m.shared++

	case vfs.LOCK_RESERVED:
		if m.reserved {
			return sqlite3.BUSY
		}
		m.reserved = true

	case vfs.LOCK_EXCLUSIVE:
		if m.lock < vfs.LOCK_PENDING {
			m.lock = vfs.LOCK_PENDING
			m.pending = true
		}

		for before := time.Now(); m.shared > 1; {
			if time.Since(before) > spinWait {
				return sqlite3.BUSY
			}
			m.lockMtx.Unlock()
			runtime.Gosched()
			m.lockMtx.Lock()
		}
	}

	m.lock = lock
	return nil
}

func (m *memFile) Unlock(lock vfs.LockLevel) error {
	if m.lock <= lock {
		return nil
	}

	m.lockMtx.Lock()
	defer m.lockMtx.Unlock()

	if m.lock >= vfs.LOCK_RESERVED {
		m.reserved = false
	}
	if m.lock >= vfs.LOCK_PENDING {
		m.pending = false
	}
	if lock < vfs.LOCK_SHARED {
		m.shared--
	}
	m.lock = lock
	return nil
}

func (m *memFile) CheckReservedLock() (bool, error) {
	// notest // OPEN_MEMORY
	if m.lock >= vfs.LOCK_RESERVED {
		return true, nil
	}
	m.lockMtx.Lock()
	defer m.lockMtx.Unlock()
	return m.reserved, nil
}

func (m *memFile) SectorSize() int {
	// notest // IOCAP_POWERSAFE_OVERWRITE
	return sectorSize
}

func (m *memFile) DeviceCharacteristics() vfs.DeviceCharacteristic {
	return vfs.IOCAP_ATOMIC |
		vfs.IOCAP_SEQUENTIAL |
		vfs.IOCAP_SAFE_APPEND |
		vfs.IOCAP_POWERSAFE_OVERWRITE
}

func (m *memFile) SizeHint(size int64) error {
	m.dataMtx.Lock()
	defer m.dataMtx.Unlock()
	if size > m.size {
		return m.truncate(size)
	}
	return nil
}

func (m *memFile) LockState() vfs.LockLevel {
	return m.lock
}

func divRoundUp(a, b int64) int64 {
	return (a + b - 1) / b
}

func modRoundUp(a, b int64) int64 {
	return b - (b-a%b)%b
}

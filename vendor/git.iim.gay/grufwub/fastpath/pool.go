package fastpath

import "sync"

// 1/8 max unix path length
const defaultBufSize = 512

var (
	builderPool sync.Pool
	once        = sync.Once{}
)

func pool() *sync.Pool {
	once.Do(func() {
		builderPool = sync.Pool{
			New: func() interface{} {
				builder := NewBuilder(make([]byte, defaultBufSize))
				return &builder
			},
		}
	})
	return &builderPool
}

func AcquireBuilder() *Builder {
	return pool().Get().(*Builder)
}

func ReleaseBuilder(b *Builder) {
	b.Reset()
	pool().Put(b)
}

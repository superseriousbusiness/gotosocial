package wasm

import (
	"context"

	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/experimental"
)

type snapshotskey struct{}

type snapshotctx struct {
	context.Context
	snaps *snapshots
}

func (ctx snapshotctx) Value(key any) any {
	if _, ok := key.(snapshotskey); ok {
		return ctx.snaps
	}
	return ctx.Context.Value(key)
}

const ringsz uint = 8

type snapshots struct {
	r [ringsz]struct {
		eptr uint32
		snap experimental.Snapshot
	}
	n uint
}

func (s *snapshots) get(envptr uint32) experimental.Snapshot {
	start := (s.n % ringsz)

	for i := start; i != ^uint(0); i-- {
		if s.r[i].eptr == envptr {
			snap := s.r[i].snap
			s.r[i].eptr = 0
			s.r[i].snap = nil
			s.n = i - 1
			return snap
		}
	}

	for i := ringsz - 1; i > start; i-- {
		if s.r[i].eptr == envptr {
			snap := s.r[i].snap
			s.r[i].eptr = 0
			s.r[i].snap = nil
			s.n = i - 1
			return snap
		}
	}

	panic("snapshot not found")
}

func (s *snapshots) set(envptr uint32, snapshot experimental.Snapshot) {
	start := (s.n % ringsz)

	for i := start; i < ringsz; i++ {
		switch s.r[i].eptr {
		case 0, envptr:
			s.r[i].eptr = envptr
			s.r[i].snap = snapshot
			s.n = i
			return
		}
	}

	for i := uint(0); i < start; i++ {
		switch s.r[i].eptr {
		case 0, envptr:
			s.r[i].eptr = envptr
			s.r[i].snap = snapshot
			s.n = i
			return
		}
	}

	panic("snapshots full")
}

// withSetjmpLongjmp updates the context to contain wazero/experimental.Snapshotter{} support,
// and embeds the necessary snapshots map required for later calls to Setjmp() / Longjmp().
func withSetjmpLongjmp(ctx context.Context) context.Context {
	return snapshotctx{Context: experimental.WithSnapshotter(ctx), snaps: new(snapshots)}
}

func getSnapshots(ctx context.Context) *snapshots {
	v, _ := ctx.Value(snapshotskey{}).(*snapshots)
	return v
}

// setjmp implements the C function: setjmp(env jmp_buf)
func setjmp(ctx context.Context, _ api.Module, stack []uint64) {

	// Input arguments.
	envptr := api.DecodeU32(stack[0])

	// Take snapshot of current execution environment.
	snapshotter := experimental.GetSnapshotter(ctx)
	snapshot := snapshotter.Snapshot()

	// Get stored snapshots map.
	snapshots := getSnapshots(ctx)

	// Set latest snapshot in map.
	snapshots.set(envptr, snapshot)

	// Set return.
	stack[0] = 0
}

// longjmp implements the C function: int longjmp(env jmp_buf, value int)
func longjmp(ctx context.Context, _ api.Module, stack []uint64) {

	// Input arguments.
	envptr := api.DecodeU32(stack[0])
	// val := stack[1]

	// Get stored snapshots map.
	snapshots := getSnapshots(ctx)
	if snapshots == nil {
		panic("setjmp / longjmp not supported")
	}

	// Get snapshot stored in map.
	snapshot := snapshots.get(envptr)

	// Set return.
	stack[0] = 0

	// Restore execution and
	// return passed value arg.
	snapshot.Restore(stack[1:])
}

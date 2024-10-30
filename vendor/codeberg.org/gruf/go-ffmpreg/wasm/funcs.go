package wasm

import (
	"context"

	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/experimental"
)

type snapshotskey struct{}

// withSetjmpLongjmp updates the context to contain wazero/experimental.Snapshotter{} support,
// and embeds the necessary snapshots map required for later calls to Setjmp() / Longjmp().
func withSetjmpLongjmp(ctx context.Context) context.Context {
	snapshots := make(map[uint32]experimental.Snapshot, 10)
	ctx = experimental.WithSnapshotter(ctx)
	ctx = context.WithValue(ctx, snapshotskey{}, snapshots)
	return ctx
}

func getSnapshots(ctx context.Context) map[uint32]experimental.Snapshot {
	v, _ := ctx.Value(snapshotskey{}).(map[uint32]experimental.Snapshot)
	return v
}

// setjmp implements the C function: setjmp(env jmp_buf)
func setjmp(ctx context.Context, mod api.Module, stack []uint64) {

	// Input arguments.
	envptr := api.DecodeU32(stack[0])

	// Take snapshot of current execution environment.
	snapshotter := experimental.GetSnapshotter(ctx)
	snapshot := snapshotter.Snapshot()

	// Get stored snapshots map.
	snapshots := getSnapshots(ctx)
	if snapshots == nil {
		panic("setjmp / longjmp not supported")
	}

	// Set latest snapshot in map.
	snapshots[envptr] = snapshot

	// Set return.
	stack[0] = 0
}

// longjmp implements the C function: int longjmp(env jmp_buf, value int)
func longjmp(ctx context.Context, mod api.Module, stack []uint64) {

	// Input arguments.
	envptr := api.DecodeU32(stack[0])
	// val := stack[1]

	// Get stored snapshots map.
	snapshots := getSnapshots(ctx)
	if snapshots == nil {
		panic("setjmp / longjmp not supported")
	}

	// Get snapshot stored in map.
	snapshot := snapshots[envptr]
	if snapshot == nil {
		panic("must first call setjmp")
	}

	// Set return.
	stack[0] = 0

	// Restore execution and
	// return passed value arg.
	snapshot.Restore(stack[1:])
}

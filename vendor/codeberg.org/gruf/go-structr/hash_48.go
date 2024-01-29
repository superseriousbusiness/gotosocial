//go:build structr_48bit_hash
// +build structr_48bit_hash

package structr

// Hash is the current compiler
// flag defined cache key hash
// checksum type. Here; uint48.
type Hash [6]byte

// uint64ToHash converts uint64 to currently Hash type.
func uint64ToHash(u uint64) Hash {
	return Hash{
		0: byte(u),
		1: byte(u >> 8),
		2: byte(u >> 16),
		3: byte(u >> 24),
		4: byte(u >> 32),
		5: byte(u >> 40),
	}
}

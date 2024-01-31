//go:build !structr_32bit_hash && !structr_48bit_hash
// +build !structr_32bit_hash,!structr_48bit_hash

package structr

// Hash is the current compiler
// flag defined cache key hash
// checksum type. Here; uint64.
type Hash uint64

// uint64ToHash converts uint64 to currently Hash type.
func uint64ToHash(u uint64) Hash {
	return Hash(u)
}

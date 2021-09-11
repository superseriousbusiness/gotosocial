package hashenc

import (
	"hash"

	"git.iim.gay/grufwub/go-bytes"
)

// HashEncoder defines an interface for calculating encoded hash sums of binary data
type HashEncoder interface {
	// EncodeSum calculates the hash sum of src and encodes (at most) Size() into dst
	EncodeSum(dst []byte, src []byte)

	// EncodedSum calculates the encoded hash sum of src and returns data in a newly allocated bytes.Bytes
	EncodedSum(src []byte) bytes.Bytes

	// Size returns the expected length of encoded hashes
	Size() int
}

// New returns a new HashEncoder instance based on supplied hash.Hash and Encoder supplying functions
func New(hash hash.Hash, enc Encoder) HashEncoder {
	hashSize := hash.Size()
	return &henc{
		hash: hash,
		hbuf: make([]byte, hashSize),
		enc:  enc,
		size: enc.EncodedLen(hashSize),
	}
}

// henc is the HashEncoder implementation
type henc struct {
	hash hash.Hash
	hbuf []byte
	enc  Encoder
	size int
}

func (henc *henc) EncodeSum(dst []byte, src []byte) {
	// Hash supplied bytes
	henc.hash.Reset()
	henc.hash.Write(src)
	henc.hbuf = henc.hash.Sum(henc.hbuf[:0])

	// Encode the hashsum and return a copy
	henc.enc.Encode(dst, henc.hbuf)
}

func (henc *henc) EncodedSum(src []byte) bytes.Bytes {
	dst := make([]byte, henc.size)
	henc.EncodeSum(dst, src)
	return bytes.ToBytes(dst)
}

func (henc *henc) Size() int {
	return henc.size
}

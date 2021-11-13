package hashenc

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"hash"
	"sync"
)

// Hash defines a pooled hash.Hash implementation
type Hash interface {
	// Hash ensures we implement the base hash.Hash implementation
	hash.Hash

	// Release resets the Hash and places it back in the pool
	Release()
}

// poolHash is our Hash implementation, providing a hash.Hash and a pool to return to
type poolHash struct {
	hash.Hash
	pool *sync.Pool
}

func (h *poolHash) Release() {
	h.Reset()
	h.pool.Put(h)
}

// SHA512Pool defines a pool of SHA512 hashes
type SHA512Pool interface {
	// SHA512 returns a Hash implementing the SHA512 hashing algorithm
	SHA512() Hash
}

// NewSHA512Pool returns a new SHA512Pool implementation
func NewSHA512Pool() SHA512Pool {
	p := &sha512Pool{}
	p.New = func() interface{} {
		return &poolHash{
			Hash: sha512.New(),
			pool: &p.Pool,
		}
	}
	return p
}

// sha512Pool is our SHA512Pool implementation, simply wrapping sync.Pool
type sha512Pool struct {
	sync.Pool
}

func (p *sha512Pool) SHA512() Hash {
	return p.Get().(Hash)
}

// SHA256Pool defines a pool of SHA256 hashes
type SHA256Pool interface {
	// SHA256 returns a Hash implementing the SHA256 hashing algorithm
	SHA256() Hash
}

// NewSHA256Pool returns a new SHA256Pool implementation
func NewSHA256Pool() SHA256Pool {
	p := &sha256Pool{}
	p.New = func() interface{} {
		return &poolHash{
			Hash: sha256.New(),
			pool: &p.Pool,
		}
	}
	return p
}

// sha256Pool is our SHA256Pool implementation, simply wrapping sync.Pool
type sha256Pool struct {
	sync.Pool
}

func (p *sha256Pool) SHA256() Hash {
	return p.Get().(Hash)
}

// SHA1Pool defines a pool of SHA1 hashes
type SHA1Pool interface {
	SHA1() Hash
}

// NewSHA1Pool returns a new SHA1Pool implementation
func NewSHA1Pool() SHA1Pool {
	p := &sha1Pool{}
	p.New = func() interface{} {
		return &poolHash{
			Hash: sha1.New(),
			pool: &p.Pool,
		}
	}
	return p
}

// sha1Pool is our SHA1Pool implementation, simply wrapping sync.Pool
type sha1Pool struct {
	sync.Pool
}

func (p *sha1Pool) SHA1() Hash {
	return p.Get().(Hash)
}

// MD5Pool defines a pool of MD5 hashes
type MD5Pool interface {
	MD5() Hash
}

// NewMD5Pool returns a new MD5 implementation
func NewMD5Pool() MD5Pool {
	p := &md5Pool{}
	p.New = func() interface{} {
		return &poolHash{
			Hash: md5.New(),
			pool: &p.Pool,
		}
	}
	return p
}

// md5Pool is our MD5Pool implementation, simply wrapping sync.Pool
type md5Pool struct {
	sync.Pool
}

func (p *md5Pool) MD5() Hash {
	return p.Get().(Hash)
}

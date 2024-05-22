package internal

import (
	"sync"

	"codeberg.org/gruf/go-fastpath/v2"
)

var pathBuilderPool sync.Pool

func GetPathBuilder() *fastpath.Builder {
	v := pathBuilderPool.Get()
	if v == nil {
		pb := new(fastpath.Builder)
		pb.B = make([]byte, 0, 512)
		v = pb
	}
	return v.(*fastpath.Builder)
}

func PutPathBuilder(pb *fastpath.Builder) {
	pb.Reset()
	pathBuilderPool.Put(pb)
}

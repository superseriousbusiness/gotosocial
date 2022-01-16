asm: internal/alg/hash/hash_avx2/impl_amd64.s internal/alg/compress/compress_sse41/impl_amd64.s

internal/alg/hash/hash_avx2/impl_amd64.s: avo/avx2/*.go
	( cd avo; go run ./avx2 ) > internal/alg/hash/hash_avx2/impl_amd64.s

internal/alg/compress/compress_sse41/impl_amd64.s: avo/sse41/*.go
	( cd avo; go run ./sse41 ) > internal/alg/compress/compress_sse41/impl_amd64.s

.PHONY: test
test:
	go test -race -bench=. -benchtime=1x

package consts

import (
	"os"

	"golang.org/x/sys/cpu"
)

var (
	HasAVX2 = cpu.X86.HasAVX2 &&
		os.Getenv("BLAKE3_DISABLE_AVX2") == "" &&
		os.Getenv("BLAKE3_PUREGO") == ""

	HasSSE41 = cpu.X86.HasSSE41 &&
		os.Getenv("BLAKE3_DISABLE_SSE41") == "" &&
		os.Getenv("BLAKE3_PUREGO") == ""
)

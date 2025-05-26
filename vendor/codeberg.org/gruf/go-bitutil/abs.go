package bitutil

// Abs8 returns the absolute value of i (calculated without branching).
func Abs8(i int8) int8 {
	const bits = 8
	u := uint64(i >> (bits - 1))
	return (i ^ int8(u)) + int8(u&1)
}

// Abs16 returns the absolute value of i (calculated without branching).
func Abs16(i int16) int16 {
	const bits = 16
	u := uint64(i >> (bits - 1))
	return (i ^ int16(u)) + int16(u&1)
}

// Abs32 returns the absolute value of i (calculated without branching).
func Abs32(i int32) int32 {
	const bits = 32
	u := uint64(i >> (bits - 1))
	return (i ^ int32(u)) + int32(u&1)
}

// Abs64 returns the absolute value of i (calculated without branching).
func Abs64(i int64) int64 {
	const bits = 64
	u := uint64(i >> (bits - 1))
	return (i ^ int64(u)) + int64(u&1)
}

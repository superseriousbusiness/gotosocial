package bitutil

// PackInt8s will pack two signed 8bit integers into an unsigned 16bit integer.
func PackInt8s(i1, i2 int8) uint16 {
	const bits = 8
	const mask = (1 << bits) - 1
	return uint16(i1)<<bits | uint16(i2)&mask
}

// UnpackInt8s will unpack two signed 8bit integers from an unsigned 16bit integer.
func UnpackInt8s(i uint16) (int8, int8) {
	const bits = 8
	const mask = (1 << bits) - 1
	return int8(i >> bits), int8(i & mask)
}

// PackInt16s will pack two signed 16bit integers into an unsigned 32bit integer.
func PackInt16s(i1, i2 int16) uint32 {
	const bits = 16
	const mask = (1 << bits) - 1
	return uint32(i1)<<bits | uint32(i2)&mask
}

// UnpackInt16s will unpack two signed 16bit integers from an unsigned 32bit integer.
func UnpackInt16s(i uint32) (int16, int16) {
	const bits = 16
	const mask = (1 << bits) - 1
	return int16(i >> bits), int16(i & mask)
}

// PackInt32s will pack two signed 32bit integers into an unsigned 64bit integer.
func PackInt32s(i1, i2 int32) uint64 {
	const bits = 32
	const mask = (1 << bits) - 1
	return uint64(i1)<<bits | uint64(i2)&mask
}

// UnpackInt32s will unpack two signed 32bit integers from an unsigned 64bit integer.
func UnpackInt32s(i uint64) (int32, int32) {
	const bits = 32
	const mask = (1 << bits) - 1
	return int32(i >> bits), int32(i & mask)
}

// PackUint8s will pack two unsigned 8bit integers into an unsigned 16bit integer.
func PackUint8s(u1, u2 uint8) uint16 {
	const bits = 8
	const mask = (1 << bits) - 1
	return uint16(u1)<<bits | uint16(u2)&mask
}

// UnpackUint8s will unpack two unsigned 8bit integers from an unsigned 16bit integer.
func UnpackUint8s(u uint16) (uint8, uint8) {
	const bits = 8
	const mask = (1 << bits) - 1
	return uint8(u >> bits), uint8(u & mask)
}

// PackUint16s will pack two unsigned 16bit integers into an unsigned 32bit integer.
func PackUint16s(u1, u2 uint16) uint32 {
	const bits = 16
	const mask = (1 << bits) - 1
	return uint32(u1)<<bits | uint32(u2)&mask
}

// UnpackUint16s will unpack two unsigned 16bit integers from an unsigned 32bit integer.
func UnpackUint16s(u uint32) (uint16, uint16) {
	const bits = 16
	const mask = (1 << bits) - 1
	return uint16(u >> bits), uint16(u & mask)
}

// PackUint32s will pack two unsigned 32bit integers into an unsigned 64bit integer.
func PackUint32s(u1, u2 uint32) uint64 {
	const bits = 32
	const mask = (1 << bits) - 1
	return uint64(u1)<<bits | uint64(u2)&mask
}

// UnpackUint32s will unpack two unsigned 32bit integers from an unsigned 64bit integer.
func UnpackUint32s(u uint64) (uint32, uint32) {
	const bits = 32
	const mask = (1 << bits) - 1
	return uint32(u >> bits), uint32(u & mask)
}

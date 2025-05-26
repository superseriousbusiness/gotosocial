package bitutil

import (
	"unsafe"
)

// Flags8 is a type-casted unsigned integer with helper
// methods for easily managing up to 8 bit-flags.
type Flags8 uint8

// Get will fetch the flag bit value at index 'bit'.
func (f Flags8) Get(bit uint8) bool {
	mask := Flags8(1) << bit
	return (f&mask != 0)
}

// Set will set the flag bit value at index 'bit'.
func (f Flags8) Set(bit uint8) Flags8 {
	mask := Flags8(1) << bit
	return f | mask
}

// Unset will unset the flag bit value at index 'bit'.
func (f Flags8) Unset(bit uint8) Flags8 {
	mask := Flags8(1) << bit
	return f & ^mask
}

// Get0 will fetch the flag bit value at index 0.
func (f Flags8) Get0() bool {
	const mask = Flags8(1) << 0
	return (f&mask != 0)
}

// Set0 will set the flag bit value at index 0.
func (f Flags8) Set0() Flags8 {
	const mask = Flags8(1) << 0
	return f | mask
}

// Unset0 will unset the flag bit value at index 0.
func (f Flags8) Unset0() Flags8 {
	const mask = Flags8(1) << 0
	return f & ^mask
}

// Get1 will fetch the flag bit value at index 1.
func (f Flags8) Get1() bool {
	const mask = Flags8(1) << 1
	return (f&mask != 0)
}

// Set1 will set the flag bit value at index 1.
func (f Flags8) Set1() Flags8 {
	const mask = Flags8(1) << 1
	return f | mask
}

// Unset1 will unset the flag bit value at index 1.
func (f Flags8) Unset1() Flags8 {
	const mask = Flags8(1) << 1
	return f & ^mask
}

// Get2 will fetch the flag bit value at index 2.
func (f Flags8) Get2() bool {
	const mask = Flags8(1) << 2
	return (f&mask != 0)
}

// Set2 will set the flag bit value at index 2.
func (f Flags8) Set2() Flags8 {
	const mask = Flags8(1) << 2
	return f | mask
}

// Unset2 will unset the flag bit value at index 2.
func (f Flags8) Unset2() Flags8 {
	const mask = Flags8(1) << 2
	return f & ^mask
}

// Get3 will fetch the flag bit value at index 3.
func (f Flags8) Get3() bool {
	const mask = Flags8(1) << 3
	return (f&mask != 0)
}

// Set3 will set the flag bit value at index 3.
func (f Flags8) Set3() Flags8 {
	const mask = Flags8(1) << 3
	return f | mask
}

// Unset3 will unset the flag bit value at index 3.
func (f Flags8) Unset3() Flags8 {
	const mask = Flags8(1) << 3
	return f & ^mask
}

// Get4 will fetch the flag bit value at index 4.
func (f Flags8) Get4() bool {
	const mask = Flags8(1) << 4
	return (f&mask != 0)
}

// Set4 will set the flag bit value at index 4.
func (f Flags8) Set4() Flags8 {
	const mask = Flags8(1) << 4
	return f | mask
}

// Unset4 will unset the flag bit value at index 4.
func (f Flags8) Unset4() Flags8 {
	const mask = Flags8(1) << 4
	return f & ^mask
}

// Get5 will fetch the flag bit value at index 5.
func (f Flags8) Get5() bool {
	const mask = Flags8(1) << 5
	return (f&mask != 0)
}

// Set5 will set the flag bit value at index 5.
func (f Flags8) Set5() Flags8 {
	const mask = Flags8(1) << 5
	return f | mask
}

// Unset5 will unset the flag bit value at index 5.
func (f Flags8) Unset5() Flags8 {
	const mask = Flags8(1) << 5
	return f & ^mask
}

// Get6 will fetch the flag bit value at index 6.
func (f Flags8) Get6() bool {
	const mask = Flags8(1) << 6
	return (f&mask != 0)
}

// Set6 will set the flag bit value at index 6.
func (f Flags8) Set6() Flags8 {
	const mask = Flags8(1) << 6
	return f | mask
}

// Unset6 will unset the flag bit value at index 6.
func (f Flags8) Unset6() Flags8 {
	const mask = Flags8(1) << 6
	return f & ^mask
}

// Get7 will fetch the flag bit value at index 7.
func (f Flags8) Get7() bool {
	const mask = Flags8(1) << 7
	return (f&mask != 0)
}

// Set7 will set the flag bit value at index 7.
func (f Flags8) Set7() Flags8 {
	const mask = Flags8(1) << 7
	return f | mask
}

// Unset7 will unset the flag bit value at index 7.
func (f Flags8) Unset7() Flags8 {
	const mask = Flags8(1) << 7
	return f & ^mask
}

// String returns a human readable representation of Flags8.
func (f Flags8) String() string {
	var (
		i   int
		val bool
		buf []byte
	)

	// Make a prealloc est. based on longest-possible value
	const prealloc = 1 + (len("false ") * 8) - 1 + 1
	buf = make([]byte, prealloc)

	buf[i] = '{'
	i++

	val = f.Get0()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get1()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get2()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get3()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get4()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get5()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get6()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get7()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	buf[i-1] = '}'
	buf = buf[:i]

	return *(*string)(unsafe.Pointer(&buf))
}

// GoString returns a more verbose human readable representation of Flags8.
func (f Flags8) GoString() string {
	var (
		i   int
		val bool
		buf []byte
	)

	// Make a prealloc est. based on longest-possible value
	const prealloc = len("bitutil.Flags8{") + (len("7=false ") * 8) - 1 + 1
	buf = make([]byte, prealloc)

	i += copy(buf[i:], "bitutil.Flags8{")

	val = f.Get0()
	i += copy(buf[i:], "0=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get1()
	i += copy(buf[i:], "1=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get2()
	i += copy(buf[i:], "2=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get3()
	i += copy(buf[i:], "3=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get4()
	i += copy(buf[i:], "4=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get5()
	i += copy(buf[i:], "5=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get6()
	i += copy(buf[i:], "6=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get7()
	i += copy(buf[i:], "7=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	buf[i-1] = '}'
	buf = buf[:i]

	return *(*string)(unsafe.Pointer(&buf))
}

// Flags16 is a type-casted unsigned integer with helper
// methods for easily managing up to 16 bit-flags.
type Flags16 uint16

// Get will fetch the flag bit value at index 'bit'.
func (f Flags16) Get(bit uint8) bool {
	mask := Flags16(1) << bit
	return (f&mask != 0)
}

// Set will set the flag bit value at index 'bit'.
func (f Flags16) Set(bit uint8) Flags16 {
	mask := Flags16(1) << bit
	return f | mask
}

// Unset will unset the flag bit value at index 'bit'.
func (f Flags16) Unset(bit uint8) Flags16 {
	mask := Flags16(1) << bit
	return f & ^mask
}

// Get0 will fetch the flag bit value at index 0.
func (f Flags16) Get0() bool {
	const mask = Flags16(1) << 0
	return (f&mask != 0)
}

// Set0 will set the flag bit value at index 0.
func (f Flags16) Set0() Flags16 {
	const mask = Flags16(1) << 0
	return f | mask
}

// Unset0 will unset the flag bit value at index 0.
func (f Flags16) Unset0() Flags16 {
	const mask = Flags16(1) << 0
	return f & ^mask
}

// Get1 will fetch the flag bit value at index 1.
func (f Flags16) Get1() bool {
	const mask = Flags16(1) << 1
	return (f&mask != 0)
}

// Set1 will set the flag bit value at index 1.
func (f Flags16) Set1() Flags16 {
	const mask = Flags16(1) << 1
	return f | mask
}

// Unset1 will unset the flag bit value at index 1.
func (f Flags16) Unset1() Flags16 {
	const mask = Flags16(1) << 1
	return f & ^mask
}

// Get2 will fetch the flag bit value at index 2.
func (f Flags16) Get2() bool {
	const mask = Flags16(1) << 2
	return (f&mask != 0)
}

// Set2 will set the flag bit value at index 2.
func (f Flags16) Set2() Flags16 {
	const mask = Flags16(1) << 2
	return f | mask
}

// Unset2 will unset the flag bit value at index 2.
func (f Flags16) Unset2() Flags16 {
	const mask = Flags16(1) << 2
	return f & ^mask
}

// Get3 will fetch the flag bit value at index 3.
func (f Flags16) Get3() bool {
	const mask = Flags16(1) << 3
	return (f&mask != 0)
}

// Set3 will set the flag bit value at index 3.
func (f Flags16) Set3() Flags16 {
	const mask = Flags16(1) << 3
	return f | mask
}

// Unset3 will unset the flag bit value at index 3.
func (f Flags16) Unset3() Flags16 {
	const mask = Flags16(1) << 3
	return f & ^mask
}

// Get4 will fetch the flag bit value at index 4.
func (f Flags16) Get4() bool {
	const mask = Flags16(1) << 4
	return (f&mask != 0)
}

// Set4 will set the flag bit value at index 4.
func (f Flags16) Set4() Flags16 {
	const mask = Flags16(1) << 4
	return f | mask
}

// Unset4 will unset the flag bit value at index 4.
func (f Flags16) Unset4() Flags16 {
	const mask = Flags16(1) << 4
	return f & ^mask
}

// Get5 will fetch the flag bit value at index 5.
func (f Flags16) Get5() bool {
	const mask = Flags16(1) << 5
	return (f&mask != 0)
}

// Set5 will set the flag bit value at index 5.
func (f Flags16) Set5() Flags16 {
	const mask = Flags16(1) << 5
	return f | mask
}

// Unset5 will unset the flag bit value at index 5.
func (f Flags16) Unset5() Flags16 {
	const mask = Flags16(1) << 5
	return f & ^mask
}

// Get6 will fetch the flag bit value at index 6.
func (f Flags16) Get6() bool {
	const mask = Flags16(1) << 6
	return (f&mask != 0)
}

// Set6 will set the flag bit value at index 6.
func (f Flags16) Set6() Flags16 {
	const mask = Flags16(1) << 6
	return f | mask
}

// Unset6 will unset the flag bit value at index 6.
func (f Flags16) Unset6() Flags16 {
	const mask = Flags16(1) << 6
	return f & ^mask
}

// Get7 will fetch the flag bit value at index 7.
func (f Flags16) Get7() bool {
	const mask = Flags16(1) << 7
	return (f&mask != 0)
}

// Set7 will set the flag bit value at index 7.
func (f Flags16) Set7() Flags16 {
	const mask = Flags16(1) << 7
	return f | mask
}

// Unset7 will unset the flag bit value at index 7.
func (f Flags16) Unset7() Flags16 {
	const mask = Flags16(1) << 7
	return f & ^mask
}

// Get8 will fetch the flag bit value at index 8.
func (f Flags16) Get8() bool {
	const mask = Flags16(1) << 8
	return (f&mask != 0)
}

// Set8 will set the flag bit value at index 8.
func (f Flags16) Set8() Flags16 {
	const mask = Flags16(1) << 8
	return f | mask
}

// Unset8 will unset the flag bit value at index 8.
func (f Flags16) Unset8() Flags16 {
	const mask = Flags16(1) << 8
	return f & ^mask
}

// Get9 will fetch the flag bit value at index 9.
func (f Flags16) Get9() bool {
	const mask = Flags16(1) << 9
	return (f&mask != 0)
}

// Set9 will set the flag bit value at index 9.
func (f Flags16) Set9() Flags16 {
	const mask = Flags16(1) << 9
	return f | mask
}

// Unset9 will unset the flag bit value at index 9.
func (f Flags16) Unset9() Flags16 {
	const mask = Flags16(1) << 9
	return f & ^mask
}

// Get10 will fetch the flag bit value at index 10.
func (f Flags16) Get10() bool {
	const mask = Flags16(1) << 10
	return (f&mask != 0)
}

// Set10 will set the flag bit value at index 10.
func (f Flags16) Set10() Flags16 {
	const mask = Flags16(1) << 10
	return f | mask
}

// Unset10 will unset the flag bit value at index 10.
func (f Flags16) Unset10() Flags16 {
	const mask = Flags16(1) << 10
	return f & ^mask
}

// Get11 will fetch the flag bit value at index 11.
func (f Flags16) Get11() bool {
	const mask = Flags16(1) << 11
	return (f&mask != 0)
}

// Set11 will set the flag bit value at index 11.
func (f Flags16) Set11() Flags16 {
	const mask = Flags16(1) << 11
	return f | mask
}

// Unset11 will unset the flag bit value at index 11.
func (f Flags16) Unset11() Flags16 {
	const mask = Flags16(1) << 11
	return f & ^mask
}

// Get12 will fetch the flag bit value at index 12.
func (f Flags16) Get12() bool {
	const mask = Flags16(1) << 12
	return (f&mask != 0)
}

// Set12 will set the flag bit value at index 12.
func (f Flags16) Set12() Flags16 {
	const mask = Flags16(1) << 12
	return f | mask
}

// Unset12 will unset the flag bit value at index 12.
func (f Flags16) Unset12() Flags16 {
	const mask = Flags16(1) << 12
	return f & ^mask
}

// Get13 will fetch the flag bit value at index 13.
func (f Flags16) Get13() bool {
	const mask = Flags16(1) << 13
	return (f&mask != 0)
}

// Set13 will set the flag bit value at index 13.
func (f Flags16) Set13() Flags16 {
	const mask = Flags16(1) << 13
	return f | mask
}

// Unset13 will unset the flag bit value at index 13.
func (f Flags16) Unset13() Flags16 {
	const mask = Flags16(1) << 13
	return f & ^mask
}

// Get14 will fetch the flag bit value at index 14.
func (f Flags16) Get14() bool {
	const mask = Flags16(1) << 14
	return (f&mask != 0)
}

// Set14 will set the flag bit value at index 14.
func (f Flags16) Set14() Flags16 {
	const mask = Flags16(1) << 14
	return f | mask
}

// Unset14 will unset the flag bit value at index 14.
func (f Flags16) Unset14() Flags16 {
	const mask = Flags16(1) << 14
	return f & ^mask
}

// Get15 will fetch the flag bit value at index 15.
func (f Flags16) Get15() bool {
	const mask = Flags16(1) << 15
	return (f&mask != 0)
}

// Set15 will set the flag bit value at index 15.
func (f Flags16) Set15() Flags16 {
	const mask = Flags16(1) << 15
	return f | mask
}

// Unset15 will unset the flag bit value at index 15.
func (f Flags16) Unset15() Flags16 {
	const mask = Flags16(1) << 15
	return f & ^mask
}

// String returns a human readable representation of Flags16.
func (f Flags16) String() string {
	var (
		i   int
		val bool
		buf []byte
	)

	// Make a prealloc est. based on longest-possible value
	const prealloc = 1 + (len("false ") * 16) - 1 + 1
	buf = make([]byte, prealloc)

	buf[i] = '{'
	i++

	val = f.Get0()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get1()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get2()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get3()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get4()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get5()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get6()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get7()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get8()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get9()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get10()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get11()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get12()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get13()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get14()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get15()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	buf[i-1] = '}'
	buf = buf[:i]

	return *(*string)(unsafe.Pointer(&buf))
}

// GoString returns a more verbose human readable representation of Flags16.
func (f Flags16) GoString() string {
	var (
		i   int
		val bool
		buf []byte
	)

	// Make a prealloc est. based on longest-possible value
	const prealloc = len("bitutil.Flags16{") + (len("15=false ") * 16) - 1 + 1
	buf = make([]byte, prealloc)

	i += copy(buf[i:], "bitutil.Flags16{")

	val = f.Get0()
	i += copy(buf[i:], "0=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get1()
	i += copy(buf[i:], "1=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get2()
	i += copy(buf[i:], "2=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get3()
	i += copy(buf[i:], "3=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get4()
	i += copy(buf[i:], "4=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get5()
	i += copy(buf[i:], "5=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get6()
	i += copy(buf[i:], "6=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get7()
	i += copy(buf[i:], "7=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get8()
	i += copy(buf[i:], "8=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get9()
	i += copy(buf[i:], "9=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get10()
	i += copy(buf[i:], "10=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get11()
	i += copy(buf[i:], "11=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get12()
	i += copy(buf[i:], "12=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get13()
	i += copy(buf[i:], "13=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get14()
	i += copy(buf[i:], "14=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get15()
	i += copy(buf[i:], "15=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	buf[i-1] = '}'
	buf = buf[:i]

	return *(*string)(unsafe.Pointer(&buf))
}

// Flags32 is a type-casted unsigned integer with helper
// methods for easily managing up to 32 bit-flags.
type Flags32 uint32

// Get will fetch the flag bit value at index 'bit'.
func (f Flags32) Get(bit uint8) bool {
	mask := Flags32(1) << bit
	return (f&mask != 0)
}

// Set will set the flag bit value at index 'bit'.
func (f Flags32) Set(bit uint8) Flags32 {
	mask := Flags32(1) << bit
	return f | mask
}

// Unset will unset the flag bit value at index 'bit'.
func (f Flags32) Unset(bit uint8) Flags32 {
	mask := Flags32(1) << bit
	return f & ^mask
}

// Get0 will fetch the flag bit value at index 0.
func (f Flags32) Get0() bool {
	const mask = Flags32(1) << 0
	return (f&mask != 0)
}

// Set0 will set the flag bit value at index 0.
func (f Flags32) Set0() Flags32 {
	const mask = Flags32(1) << 0
	return f | mask
}

// Unset0 will unset the flag bit value at index 0.
func (f Flags32) Unset0() Flags32 {
	const mask = Flags32(1) << 0
	return f & ^mask
}

// Get1 will fetch the flag bit value at index 1.
func (f Flags32) Get1() bool {
	const mask = Flags32(1) << 1
	return (f&mask != 0)
}

// Set1 will set the flag bit value at index 1.
func (f Flags32) Set1() Flags32 {
	const mask = Flags32(1) << 1
	return f | mask
}

// Unset1 will unset the flag bit value at index 1.
func (f Flags32) Unset1() Flags32 {
	const mask = Flags32(1) << 1
	return f & ^mask
}

// Get2 will fetch the flag bit value at index 2.
func (f Flags32) Get2() bool {
	const mask = Flags32(1) << 2
	return (f&mask != 0)
}

// Set2 will set the flag bit value at index 2.
func (f Flags32) Set2() Flags32 {
	const mask = Flags32(1) << 2
	return f | mask
}

// Unset2 will unset the flag bit value at index 2.
func (f Flags32) Unset2() Flags32 {
	const mask = Flags32(1) << 2
	return f & ^mask
}

// Get3 will fetch the flag bit value at index 3.
func (f Flags32) Get3() bool {
	const mask = Flags32(1) << 3
	return (f&mask != 0)
}

// Set3 will set the flag bit value at index 3.
func (f Flags32) Set3() Flags32 {
	const mask = Flags32(1) << 3
	return f | mask
}

// Unset3 will unset the flag bit value at index 3.
func (f Flags32) Unset3() Flags32 {
	const mask = Flags32(1) << 3
	return f & ^mask
}

// Get4 will fetch the flag bit value at index 4.
func (f Flags32) Get4() bool {
	const mask = Flags32(1) << 4
	return (f&mask != 0)
}

// Set4 will set the flag bit value at index 4.
func (f Flags32) Set4() Flags32 {
	const mask = Flags32(1) << 4
	return f | mask
}

// Unset4 will unset the flag bit value at index 4.
func (f Flags32) Unset4() Flags32 {
	const mask = Flags32(1) << 4
	return f & ^mask
}

// Get5 will fetch the flag bit value at index 5.
func (f Flags32) Get5() bool {
	const mask = Flags32(1) << 5
	return (f&mask != 0)
}

// Set5 will set the flag bit value at index 5.
func (f Flags32) Set5() Flags32 {
	const mask = Flags32(1) << 5
	return f | mask
}

// Unset5 will unset the flag bit value at index 5.
func (f Flags32) Unset5() Flags32 {
	const mask = Flags32(1) << 5
	return f & ^mask
}

// Get6 will fetch the flag bit value at index 6.
func (f Flags32) Get6() bool {
	const mask = Flags32(1) << 6
	return (f&mask != 0)
}

// Set6 will set the flag bit value at index 6.
func (f Flags32) Set6() Flags32 {
	const mask = Flags32(1) << 6
	return f | mask
}

// Unset6 will unset the flag bit value at index 6.
func (f Flags32) Unset6() Flags32 {
	const mask = Flags32(1) << 6
	return f & ^mask
}

// Get7 will fetch the flag bit value at index 7.
func (f Flags32) Get7() bool {
	const mask = Flags32(1) << 7
	return (f&mask != 0)
}

// Set7 will set the flag bit value at index 7.
func (f Flags32) Set7() Flags32 {
	const mask = Flags32(1) << 7
	return f | mask
}

// Unset7 will unset the flag bit value at index 7.
func (f Flags32) Unset7() Flags32 {
	const mask = Flags32(1) << 7
	return f & ^mask
}

// Get8 will fetch the flag bit value at index 8.
func (f Flags32) Get8() bool {
	const mask = Flags32(1) << 8
	return (f&mask != 0)
}

// Set8 will set the flag bit value at index 8.
func (f Flags32) Set8() Flags32 {
	const mask = Flags32(1) << 8
	return f | mask
}

// Unset8 will unset the flag bit value at index 8.
func (f Flags32) Unset8() Flags32 {
	const mask = Flags32(1) << 8
	return f & ^mask
}

// Get9 will fetch the flag bit value at index 9.
func (f Flags32) Get9() bool {
	const mask = Flags32(1) << 9
	return (f&mask != 0)
}

// Set9 will set the flag bit value at index 9.
func (f Flags32) Set9() Flags32 {
	const mask = Flags32(1) << 9
	return f | mask
}

// Unset9 will unset the flag bit value at index 9.
func (f Flags32) Unset9() Flags32 {
	const mask = Flags32(1) << 9
	return f & ^mask
}

// Get10 will fetch the flag bit value at index 10.
func (f Flags32) Get10() bool {
	const mask = Flags32(1) << 10
	return (f&mask != 0)
}

// Set10 will set the flag bit value at index 10.
func (f Flags32) Set10() Flags32 {
	const mask = Flags32(1) << 10
	return f | mask
}

// Unset10 will unset the flag bit value at index 10.
func (f Flags32) Unset10() Flags32 {
	const mask = Flags32(1) << 10
	return f & ^mask
}

// Get11 will fetch the flag bit value at index 11.
func (f Flags32) Get11() bool {
	const mask = Flags32(1) << 11
	return (f&mask != 0)
}

// Set11 will set the flag bit value at index 11.
func (f Flags32) Set11() Flags32 {
	const mask = Flags32(1) << 11
	return f | mask
}

// Unset11 will unset the flag bit value at index 11.
func (f Flags32) Unset11() Flags32 {
	const mask = Flags32(1) << 11
	return f & ^mask
}

// Get12 will fetch the flag bit value at index 12.
func (f Flags32) Get12() bool {
	const mask = Flags32(1) << 12
	return (f&mask != 0)
}

// Set12 will set the flag bit value at index 12.
func (f Flags32) Set12() Flags32 {
	const mask = Flags32(1) << 12
	return f | mask
}

// Unset12 will unset the flag bit value at index 12.
func (f Flags32) Unset12() Flags32 {
	const mask = Flags32(1) << 12
	return f & ^mask
}

// Get13 will fetch the flag bit value at index 13.
func (f Flags32) Get13() bool {
	const mask = Flags32(1) << 13
	return (f&mask != 0)
}

// Set13 will set the flag bit value at index 13.
func (f Flags32) Set13() Flags32 {
	const mask = Flags32(1) << 13
	return f | mask
}

// Unset13 will unset the flag bit value at index 13.
func (f Flags32) Unset13() Flags32 {
	const mask = Flags32(1) << 13
	return f & ^mask
}

// Get14 will fetch the flag bit value at index 14.
func (f Flags32) Get14() bool {
	const mask = Flags32(1) << 14
	return (f&mask != 0)
}

// Set14 will set the flag bit value at index 14.
func (f Flags32) Set14() Flags32 {
	const mask = Flags32(1) << 14
	return f | mask
}

// Unset14 will unset the flag bit value at index 14.
func (f Flags32) Unset14() Flags32 {
	const mask = Flags32(1) << 14
	return f & ^mask
}

// Get15 will fetch the flag bit value at index 15.
func (f Flags32) Get15() bool {
	const mask = Flags32(1) << 15
	return (f&mask != 0)
}

// Set15 will set the flag bit value at index 15.
func (f Flags32) Set15() Flags32 {
	const mask = Flags32(1) << 15
	return f | mask
}

// Unset15 will unset the flag bit value at index 15.
func (f Flags32) Unset15() Flags32 {
	const mask = Flags32(1) << 15
	return f & ^mask
}

// Get16 will fetch the flag bit value at index 16.
func (f Flags32) Get16() bool {
	const mask = Flags32(1) << 16
	return (f&mask != 0)
}

// Set16 will set the flag bit value at index 16.
func (f Flags32) Set16() Flags32 {
	const mask = Flags32(1) << 16
	return f | mask
}

// Unset16 will unset the flag bit value at index 16.
func (f Flags32) Unset16() Flags32 {
	const mask = Flags32(1) << 16
	return f & ^mask
}

// Get17 will fetch the flag bit value at index 17.
func (f Flags32) Get17() bool {
	const mask = Flags32(1) << 17
	return (f&mask != 0)
}

// Set17 will set the flag bit value at index 17.
func (f Flags32) Set17() Flags32 {
	const mask = Flags32(1) << 17
	return f | mask
}

// Unset17 will unset the flag bit value at index 17.
func (f Flags32) Unset17() Flags32 {
	const mask = Flags32(1) << 17
	return f & ^mask
}

// Get18 will fetch the flag bit value at index 18.
func (f Flags32) Get18() bool {
	const mask = Flags32(1) << 18
	return (f&mask != 0)
}

// Set18 will set the flag bit value at index 18.
func (f Flags32) Set18() Flags32 {
	const mask = Flags32(1) << 18
	return f | mask
}

// Unset18 will unset the flag bit value at index 18.
func (f Flags32) Unset18() Flags32 {
	const mask = Flags32(1) << 18
	return f & ^mask
}

// Get19 will fetch the flag bit value at index 19.
func (f Flags32) Get19() bool {
	const mask = Flags32(1) << 19
	return (f&mask != 0)
}

// Set19 will set the flag bit value at index 19.
func (f Flags32) Set19() Flags32 {
	const mask = Flags32(1) << 19
	return f | mask
}

// Unset19 will unset the flag bit value at index 19.
func (f Flags32) Unset19() Flags32 {
	const mask = Flags32(1) << 19
	return f & ^mask
}

// Get20 will fetch the flag bit value at index 20.
func (f Flags32) Get20() bool {
	const mask = Flags32(1) << 20
	return (f&mask != 0)
}

// Set20 will set the flag bit value at index 20.
func (f Flags32) Set20() Flags32 {
	const mask = Flags32(1) << 20
	return f | mask
}

// Unset20 will unset the flag bit value at index 20.
func (f Flags32) Unset20() Flags32 {
	const mask = Flags32(1) << 20
	return f & ^mask
}

// Get21 will fetch the flag bit value at index 21.
func (f Flags32) Get21() bool {
	const mask = Flags32(1) << 21
	return (f&mask != 0)
}

// Set21 will set the flag bit value at index 21.
func (f Flags32) Set21() Flags32 {
	const mask = Flags32(1) << 21
	return f | mask
}

// Unset21 will unset the flag bit value at index 21.
func (f Flags32) Unset21() Flags32 {
	const mask = Flags32(1) << 21
	return f & ^mask
}

// Get22 will fetch the flag bit value at index 22.
func (f Flags32) Get22() bool {
	const mask = Flags32(1) << 22
	return (f&mask != 0)
}

// Set22 will set the flag bit value at index 22.
func (f Flags32) Set22() Flags32 {
	const mask = Flags32(1) << 22
	return f | mask
}

// Unset22 will unset the flag bit value at index 22.
func (f Flags32) Unset22() Flags32 {
	const mask = Flags32(1) << 22
	return f & ^mask
}

// Get23 will fetch the flag bit value at index 23.
func (f Flags32) Get23() bool {
	const mask = Flags32(1) << 23
	return (f&mask != 0)
}

// Set23 will set the flag bit value at index 23.
func (f Flags32) Set23() Flags32 {
	const mask = Flags32(1) << 23
	return f | mask
}

// Unset23 will unset the flag bit value at index 23.
func (f Flags32) Unset23() Flags32 {
	const mask = Flags32(1) << 23
	return f & ^mask
}

// Get24 will fetch the flag bit value at index 24.
func (f Flags32) Get24() bool {
	const mask = Flags32(1) << 24
	return (f&mask != 0)
}

// Set24 will set the flag bit value at index 24.
func (f Flags32) Set24() Flags32 {
	const mask = Flags32(1) << 24
	return f | mask
}

// Unset24 will unset the flag bit value at index 24.
func (f Flags32) Unset24() Flags32 {
	const mask = Flags32(1) << 24
	return f & ^mask
}

// Get25 will fetch the flag bit value at index 25.
func (f Flags32) Get25() bool {
	const mask = Flags32(1) << 25
	return (f&mask != 0)
}

// Set25 will set the flag bit value at index 25.
func (f Flags32) Set25() Flags32 {
	const mask = Flags32(1) << 25
	return f | mask
}

// Unset25 will unset the flag bit value at index 25.
func (f Flags32) Unset25() Flags32 {
	const mask = Flags32(1) << 25
	return f & ^mask
}

// Get26 will fetch the flag bit value at index 26.
func (f Flags32) Get26() bool {
	const mask = Flags32(1) << 26
	return (f&mask != 0)
}

// Set26 will set the flag bit value at index 26.
func (f Flags32) Set26() Flags32 {
	const mask = Flags32(1) << 26
	return f | mask
}

// Unset26 will unset the flag bit value at index 26.
func (f Flags32) Unset26() Flags32 {
	const mask = Flags32(1) << 26
	return f & ^mask
}

// Get27 will fetch the flag bit value at index 27.
func (f Flags32) Get27() bool {
	const mask = Flags32(1) << 27
	return (f&mask != 0)
}

// Set27 will set the flag bit value at index 27.
func (f Flags32) Set27() Flags32 {
	const mask = Flags32(1) << 27
	return f | mask
}

// Unset27 will unset the flag bit value at index 27.
func (f Flags32) Unset27() Flags32 {
	const mask = Flags32(1) << 27
	return f & ^mask
}

// Get28 will fetch the flag bit value at index 28.
func (f Flags32) Get28() bool {
	const mask = Flags32(1) << 28
	return (f&mask != 0)
}

// Set28 will set the flag bit value at index 28.
func (f Flags32) Set28() Flags32 {
	const mask = Flags32(1) << 28
	return f | mask
}

// Unset28 will unset the flag bit value at index 28.
func (f Flags32) Unset28() Flags32 {
	const mask = Flags32(1) << 28
	return f & ^mask
}

// Get29 will fetch the flag bit value at index 29.
func (f Flags32) Get29() bool {
	const mask = Flags32(1) << 29
	return (f&mask != 0)
}

// Set29 will set the flag bit value at index 29.
func (f Flags32) Set29() Flags32 {
	const mask = Flags32(1) << 29
	return f | mask
}

// Unset29 will unset the flag bit value at index 29.
func (f Flags32) Unset29() Flags32 {
	const mask = Flags32(1) << 29
	return f & ^mask
}

// Get30 will fetch the flag bit value at index 30.
func (f Flags32) Get30() bool {
	const mask = Flags32(1) << 30
	return (f&mask != 0)
}

// Set30 will set the flag bit value at index 30.
func (f Flags32) Set30() Flags32 {
	const mask = Flags32(1) << 30
	return f | mask
}

// Unset30 will unset the flag bit value at index 30.
func (f Flags32) Unset30() Flags32 {
	const mask = Flags32(1) << 30
	return f & ^mask
}

// Get31 will fetch the flag bit value at index 31.
func (f Flags32) Get31() bool {
	const mask = Flags32(1) << 31
	return (f&mask != 0)
}

// Set31 will set the flag bit value at index 31.
func (f Flags32) Set31() Flags32 {
	const mask = Flags32(1) << 31
	return f | mask
}

// Unset31 will unset the flag bit value at index 31.
func (f Flags32) Unset31() Flags32 {
	const mask = Flags32(1) << 31
	return f & ^mask
}

// String returns a human readable representation of Flags32.
func (f Flags32) String() string {
	var (
		i   int
		val bool
		buf []byte
	)

	// Make a prealloc est. based on longest-possible value
	const prealloc = 1 + (len("false ") * 32) - 1 + 1
	buf = make([]byte, prealloc)

	buf[i] = '{'
	i++

	val = f.Get0()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get1()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get2()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get3()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get4()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get5()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get6()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get7()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get8()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get9()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get10()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get11()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get12()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get13()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get14()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get15()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get16()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get17()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get18()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get19()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get20()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get21()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get22()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get23()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get24()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get25()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get26()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get27()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get28()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get29()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get30()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get31()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	buf[i-1] = '}'
	buf = buf[:i]

	return *(*string)(unsafe.Pointer(&buf))
}

// GoString returns a more verbose human readable representation of Flags32.
func (f Flags32) GoString() string {
	var (
		i   int
		val bool
		buf []byte
	)

	// Make a prealloc est. based on longest-possible value
	const prealloc = len("bitutil.Flags32{") + (len("31=false ") * 32) - 1 + 1
	buf = make([]byte, prealloc)

	i += copy(buf[i:], "bitutil.Flags32{")

	val = f.Get0()
	i += copy(buf[i:], "0=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get1()
	i += copy(buf[i:], "1=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get2()
	i += copy(buf[i:], "2=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get3()
	i += copy(buf[i:], "3=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get4()
	i += copy(buf[i:], "4=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get5()
	i += copy(buf[i:], "5=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get6()
	i += copy(buf[i:], "6=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get7()
	i += copy(buf[i:], "7=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get8()
	i += copy(buf[i:], "8=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get9()
	i += copy(buf[i:], "9=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get10()
	i += copy(buf[i:], "10=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get11()
	i += copy(buf[i:], "11=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get12()
	i += copy(buf[i:], "12=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get13()
	i += copy(buf[i:], "13=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get14()
	i += copy(buf[i:], "14=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get15()
	i += copy(buf[i:], "15=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get16()
	i += copy(buf[i:], "16=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get17()
	i += copy(buf[i:], "17=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get18()
	i += copy(buf[i:], "18=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get19()
	i += copy(buf[i:], "19=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get20()
	i += copy(buf[i:], "20=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get21()
	i += copy(buf[i:], "21=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get22()
	i += copy(buf[i:], "22=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get23()
	i += copy(buf[i:], "23=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get24()
	i += copy(buf[i:], "24=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get25()
	i += copy(buf[i:], "25=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get26()
	i += copy(buf[i:], "26=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get27()
	i += copy(buf[i:], "27=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get28()
	i += copy(buf[i:], "28=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get29()
	i += copy(buf[i:], "29=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get30()
	i += copy(buf[i:], "30=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get31()
	i += copy(buf[i:], "31=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	buf[i-1] = '}'
	buf = buf[:i]

	return *(*string)(unsafe.Pointer(&buf))
}

// Flags64 is a type-casted unsigned integer with helper
// methods for easily managing up to 64 bit-flags.
type Flags64 uint64

// Get will fetch the flag bit value at index 'bit'.
func (f Flags64) Get(bit uint8) bool {
	mask := Flags64(1) << bit
	return (f&mask != 0)
}

// Set will set the flag bit value at index 'bit'.
func (f Flags64) Set(bit uint8) Flags64 {
	mask := Flags64(1) << bit
	return f | mask
}

// Unset will unset the flag bit value at index 'bit'.
func (f Flags64) Unset(bit uint8) Flags64 {
	mask := Flags64(1) << bit
	return f & ^mask
}

// Get0 will fetch the flag bit value at index 0.
func (f Flags64) Get0() bool {
	const mask = Flags64(1) << 0
	return (f&mask != 0)
}

// Set0 will set the flag bit value at index 0.
func (f Flags64) Set0() Flags64 {
	const mask = Flags64(1) << 0
	return f | mask
}

// Unset0 will unset the flag bit value at index 0.
func (f Flags64) Unset0() Flags64 {
	const mask = Flags64(1) << 0
	return f & ^mask
}

// Get1 will fetch the flag bit value at index 1.
func (f Flags64) Get1() bool {
	const mask = Flags64(1) << 1
	return (f&mask != 0)
}

// Set1 will set the flag bit value at index 1.
func (f Flags64) Set1() Flags64 {
	const mask = Flags64(1) << 1
	return f | mask
}

// Unset1 will unset the flag bit value at index 1.
func (f Flags64) Unset1() Flags64 {
	const mask = Flags64(1) << 1
	return f & ^mask
}

// Get2 will fetch the flag bit value at index 2.
func (f Flags64) Get2() bool {
	const mask = Flags64(1) << 2
	return (f&mask != 0)
}

// Set2 will set the flag bit value at index 2.
func (f Flags64) Set2() Flags64 {
	const mask = Flags64(1) << 2
	return f | mask
}

// Unset2 will unset the flag bit value at index 2.
func (f Flags64) Unset2() Flags64 {
	const mask = Flags64(1) << 2
	return f & ^mask
}

// Get3 will fetch the flag bit value at index 3.
func (f Flags64) Get3() bool {
	const mask = Flags64(1) << 3
	return (f&mask != 0)
}

// Set3 will set the flag bit value at index 3.
func (f Flags64) Set3() Flags64 {
	const mask = Flags64(1) << 3
	return f | mask
}

// Unset3 will unset the flag bit value at index 3.
func (f Flags64) Unset3() Flags64 {
	const mask = Flags64(1) << 3
	return f & ^mask
}

// Get4 will fetch the flag bit value at index 4.
func (f Flags64) Get4() bool {
	const mask = Flags64(1) << 4
	return (f&mask != 0)
}

// Set4 will set the flag bit value at index 4.
func (f Flags64) Set4() Flags64 {
	const mask = Flags64(1) << 4
	return f | mask
}

// Unset4 will unset the flag bit value at index 4.
func (f Flags64) Unset4() Flags64 {
	const mask = Flags64(1) << 4
	return f & ^mask
}

// Get5 will fetch the flag bit value at index 5.
func (f Flags64) Get5() bool {
	const mask = Flags64(1) << 5
	return (f&mask != 0)
}

// Set5 will set the flag bit value at index 5.
func (f Flags64) Set5() Flags64 {
	const mask = Flags64(1) << 5
	return f | mask
}

// Unset5 will unset the flag bit value at index 5.
func (f Flags64) Unset5() Flags64 {
	const mask = Flags64(1) << 5
	return f & ^mask
}

// Get6 will fetch the flag bit value at index 6.
func (f Flags64) Get6() bool {
	const mask = Flags64(1) << 6
	return (f&mask != 0)
}

// Set6 will set the flag bit value at index 6.
func (f Flags64) Set6() Flags64 {
	const mask = Flags64(1) << 6
	return f | mask
}

// Unset6 will unset the flag bit value at index 6.
func (f Flags64) Unset6() Flags64 {
	const mask = Flags64(1) << 6
	return f & ^mask
}

// Get7 will fetch the flag bit value at index 7.
func (f Flags64) Get7() bool {
	const mask = Flags64(1) << 7
	return (f&mask != 0)
}

// Set7 will set the flag bit value at index 7.
func (f Flags64) Set7() Flags64 {
	const mask = Flags64(1) << 7
	return f | mask
}

// Unset7 will unset the flag bit value at index 7.
func (f Flags64) Unset7() Flags64 {
	const mask = Flags64(1) << 7
	return f & ^mask
}

// Get8 will fetch the flag bit value at index 8.
func (f Flags64) Get8() bool {
	const mask = Flags64(1) << 8
	return (f&mask != 0)
}

// Set8 will set the flag bit value at index 8.
func (f Flags64) Set8() Flags64 {
	const mask = Flags64(1) << 8
	return f | mask
}

// Unset8 will unset the flag bit value at index 8.
func (f Flags64) Unset8() Flags64 {
	const mask = Flags64(1) << 8
	return f & ^mask
}

// Get9 will fetch the flag bit value at index 9.
func (f Flags64) Get9() bool {
	const mask = Flags64(1) << 9
	return (f&mask != 0)
}

// Set9 will set the flag bit value at index 9.
func (f Flags64) Set9() Flags64 {
	const mask = Flags64(1) << 9
	return f | mask
}

// Unset9 will unset the flag bit value at index 9.
func (f Flags64) Unset9() Flags64 {
	const mask = Flags64(1) << 9
	return f & ^mask
}

// Get10 will fetch the flag bit value at index 10.
func (f Flags64) Get10() bool {
	const mask = Flags64(1) << 10
	return (f&mask != 0)
}

// Set10 will set the flag bit value at index 10.
func (f Flags64) Set10() Flags64 {
	const mask = Flags64(1) << 10
	return f | mask
}

// Unset10 will unset the flag bit value at index 10.
func (f Flags64) Unset10() Flags64 {
	const mask = Flags64(1) << 10
	return f & ^mask
}

// Get11 will fetch the flag bit value at index 11.
func (f Flags64) Get11() bool {
	const mask = Flags64(1) << 11
	return (f&mask != 0)
}

// Set11 will set the flag bit value at index 11.
func (f Flags64) Set11() Flags64 {
	const mask = Flags64(1) << 11
	return f | mask
}

// Unset11 will unset the flag bit value at index 11.
func (f Flags64) Unset11() Flags64 {
	const mask = Flags64(1) << 11
	return f & ^mask
}

// Get12 will fetch the flag bit value at index 12.
func (f Flags64) Get12() bool {
	const mask = Flags64(1) << 12
	return (f&mask != 0)
}

// Set12 will set the flag bit value at index 12.
func (f Flags64) Set12() Flags64 {
	const mask = Flags64(1) << 12
	return f | mask
}

// Unset12 will unset the flag bit value at index 12.
func (f Flags64) Unset12() Flags64 {
	const mask = Flags64(1) << 12
	return f & ^mask
}

// Get13 will fetch the flag bit value at index 13.
func (f Flags64) Get13() bool {
	const mask = Flags64(1) << 13
	return (f&mask != 0)
}

// Set13 will set the flag bit value at index 13.
func (f Flags64) Set13() Flags64 {
	const mask = Flags64(1) << 13
	return f | mask
}

// Unset13 will unset the flag bit value at index 13.
func (f Flags64) Unset13() Flags64 {
	const mask = Flags64(1) << 13
	return f & ^mask
}

// Get14 will fetch the flag bit value at index 14.
func (f Flags64) Get14() bool {
	const mask = Flags64(1) << 14
	return (f&mask != 0)
}

// Set14 will set the flag bit value at index 14.
func (f Flags64) Set14() Flags64 {
	const mask = Flags64(1) << 14
	return f | mask
}

// Unset14 will unset the flag bit value at index 14.
func (f Flags64) Unset14() Flags64 {
	const mask = Flags64(1) << 14
	return f & ^mask
}

// Get15 will fetch the flag bit value at index 15.
func (f Flags64) Get15() bool {
	const mask = Flags64(1) << 15
	return (f&mask != 0)
}

// Set15 will set the flag bit value at index 15.
func (f Flags64) Set15() Flags64 {
	const mask = Flags64(1) << 15
	return f | mask
}

// Unset15 will unset the flag bit value at index 15.
func (f Flags64) Unset15() Flags64 {
	const mask = Flags64(1) << 15
	return f & ^mask
}

// Get16 will fetch the flag bit value at index 16.
func (f Flags64) Get16() bool {
	const mask = Flags64(1) << 16
	return (f&mask != 0)
}

// Set16 will set the flag bit value at index 16.
func (f Flags64) Set16() Flags64 {
	const mask = Flags64(1) << 16
	return f | mask
}

// Unset16 will unset the flag bit value at index 16.
func (f Flags64) Unset16() Flags64 {
	const mask = Flags64(1) << 16
	return f & ^mask
}

// Get17 will fetch the flag bit value at index 17.
func (f Flags64) Get17() bool {
	const mask = Flags64(1) << 17
	return (f&mask != 0)
}

// Set17 will set the flag bit value at index 17.
func (f Flags64) Set17() Flags64 {
	const mask = Flags64(1) << 17
	return f | mask
}

// Unset17 will unset the flag bit value at index 17.
func (f Flags64) Unset17() Flags64 {
	const mask = Flags64(1) << 17
	return f & ^mask
}

// Get18 will fetch the flag bit value at index 18.
func (f Flags64) Get18() bool {
	const mask = Flags64(1) << 18
	return (f&mask != 0)
}

// Set18 will set the flag bit value at index 18.
func (f Flags64) Set18() Flags64 {
	const mask = Flags64(1) << 18
	return f | mask
}

// Unset18 will unset the flag bit value at index 18.
func (f Flags64) Unset18() Flags64 {
	const mask = Flags64(1) << 18
	return f & ^mask
}

// Get19 will fetch the flag bit value at index 19.
func (f Flags64) Get19() bool {
	const mask = Flags64(1) << 19
	return (f&mask != 0)
}

// Set19 will set the flag bit value at index 19.
func (f Flags64) Set19() Flags64 {
	const mask = Flags64(1) << 19
	return f | mask
}

// Unset19 will unset the flag bit value at index 19.
func (f Flags64) Unset19() Flags64 {
	const mask = Flags64(1) << 19
	return f & ^mask
}

// Get20 will fetch the flag bit value at index 20.
func (f Flags64) Get20() bool {
	const mask = Flags64(1) << 20
	return (f&mask != 0)
}

// Set20 will set the flag bit value at index 20.
func (f Flags64) Set20() Flags64 {
	const mask = Flags64(1) << 20
	return f | mask
}

// Unset20 will unset the flag bit value at index 20.
func (f Flags64) Unset20() Flags64 {
	const mask = Flags64(1) << 20
	return f & ^mask
}

// Get21 will fetch the flag bit value at index 21.
func (f Flags64) Get21() bool {
	const mask = Flags64(1) << 21
	return (f&mask != 0)
}

// Set21 will set the flag bit value at index 21.
func (f Flags64) Set21() Flags64 {
	const mask = Flags64(1) << 21
	return f | mask
}

// Unset21 will unset the flag bit value at index 21.
func (f Flags64) Unset21() Flags64 {
	const mask = Flags64(1) << 21
	return f & ^mask
}

// Get22 will fetch the flag bit value at index 22.
func (f Flags64) Get22() bool {
	const mask = Flags64(1) << 22
	return (f&mask != 0)
}

// Set22 will set the flag bit value at index 22.
func (f Flags64) Set22() Flags64 {
	const mask = Flags64(1) << 22
	return f | mask
}

// Unset22 will unset the flag bit value at index 22.
func (f Flags64) Unset22() Flags64 {
	const mask = Flags64(1) << 22
	return f & ^mask
}

// Get23 will fetch the flag bit value at index 23.
func (f Flags64) Get23() bool {
	const mask = Flags64(1) << 23
	return (f&mask != 0)
}

// Set23 will set the flag bit value at index 23.
func (f Flags64) Set23() Flags64 {
	const mask = Flags64(1) << 23
	return f | mask
}

// Unset23 will unset the flag bit value at index 23.
func (f Flags64) Unset23() Flags64 {
	const mask = Flags64(1) << 23
	return f & ^mask
}

// Get24 will fetch the flag bit value at index 24.
func (f Flags64) Get24() bool {
	const mask = Flags64(1) << 24
	return (f&mask != 0)
}

// Set24 will set the flag bit value at index 24.
func (f Flags64) Set24() Flags64 {
	const mask = Flags64(1) << 24
	return f | mask
}

// Unset24 will unset the flag bit value at index 24.
func (f Flags64) Unset24() Flags64 {
	const mask = Flags64(1) << 24
	return f & ^mask
}

// Get25 will fetch the flag bit value at index 25.
func (f Flags64) Get25() bool {
	const mask = Flags64(1) << 25
	return (f&mask != 0)
}

// Set25 will set the flag bit value at index 25.
func (f Flags64) Set25() Flags64 {
	const mask = Flags64(1) << 25
	return f | mask
}

// Unset25 will unset the flag bit value at index 25.
func (f Flags64) Unset25() Flags64 {
	const mask = Flags64(1) << 25
	return f & ^mask
}

// Get26 will fetch the flag bit value at index 26.
func (f Flags64) Get26() bool {
	const mask = Flags64(1) << 26
	return (f&mask != 0)
}

// Set26 will set the flag bit value at index 26.
func (f Flags64) Set26() Flags64 {
	const mask = Flags64(1) << 26
	return f | mask
}

// Unset26 will unset the flag bit value at index 26.
func (f Flags64) Unset26() Flags64 {
	const mask = Flags64(1) << 26
	return f & ^mask
}

// Get27 will fetch the flag bit value at index 27.
func (f Flags64) Get27() bool {
	const mask = Flags64(1) << 27
	return (f&mask != 0)
}

// Set27 will set the flag bit value at index 27.
func (f Flags64) Set27() Flags64 {
	const mask = Flags64(1) << 27
	return f | mask
}

// Unset27 will unset the flag bit value at index 27.
func (f Flags64) Unset27() Flags64 {
	const mask = Flags64(1) << 27
	return f & ^mask
}

// Get28 will fetch the flag bit value at index 28.
func (f Flags64) Get28() bool {
	const mask = Flags64(1) << 28
	return (f&mask != 0)
}

// Set28 will set the flag bit value at index 28.
func (f Flags64) Set28() Flags64 {
	const mask = Flags64(1) << 28
	return f | mask
}

// Unset28 will unset the flag bit value at index 28.
func (f Flags64) Unset28() Flags64 {
	const mask = Flags64(1) << 28
	return f & ^mask
}

// Get29 will fetch the flag bit value at index 29.
func (f Flags64) Get29() bool {
	const mask = Flags64(1) << 29
	return (f&mask != 0)
}

// Set29 will set the flag bit value at index 29.
func (f Flags64) Set29() Flags64 {
	const mask = Flags64(1) << 29
	return f | mask
}

// Unset29 will unset the flag bit value at index 29.
func (f Flags64) Unset29() Flags64 {
	const mask = Flags64(1) << 29
	return f & ^mask
}

// Get30 will fetch the flag bit value at index 30.
func (f Flags64) Get30() bool {
	const mask = Flags64(1) << 30
	return (f&mask != 0)
}

// Set30 will set the flag bit value at index 30.
func (f Flags64) Set30() Flags64 {
	const mask = Flags64(1) << 30
	return f | mask
}

// Unset30 will unset the flag bit value at index 30.
func (f Flags64) Unset30() Flags64 {
	const mask = Flags64(1) << 30
	return f & ^mask
}

// Get31 will fetch the flag bit value at index 31.
func (f Flags64) Get31() bool {
	const mask = Flags64(1) << 31
	return (f&mask != 0)
}

// Set31 will set the flag bit value at index 31.
func (f Flags64) Set31() Flags64 {
	const mask = Flags64(1) << 31
	return f | mask
}

// Unset31 will unset the flag bit value at index 31.
func (f Flags64) Unset31() Flags64 {
	const mask = Flags64(1) << 31
	return f & ^mask
}

// Get32 will fetch the flag bit value at index 32.
func (f Flags64) Get32() bool {
	const mask = Flags64(1) << 32
	return (f&mask != 0)
}

// Set32 will set the flag bit value at index 32.
func (f Flags64) Set32() Flags64 {
	const mask = Flags64(1) << 32
	return f | mask
}

// Unset32 will unset the flag bit value at index 32.
func (f Flags64) Unset32() Flags64 {
	const mask = Flags64(1) << 32
	return f & ^mask
}

// Get33 will fetch the flag bit value at index 33.
func (f Flags64) Get33() bool {
	const mask = Flags64(1) << 33
	return (f&mask != 0)
}

// Set33 will set the flag bit value at index 33.
func (f Flags64) Set33() Flags64 {
	const mask = Flags64(1) << 33
	return f | mask
}

// Unset33 will unset the flag bit value at index 33.
func (f Flags64) Unset33() Flags64 {
	const mask = Flags64(1) << 33
	return f & ^mask
}

// Get34 will fetch the flag bit value at index 34.
func (f Flags64) Get34() bool {
	const mask = Flags64(1) << 34
	return (f&mask != 0)
}

// Set34 will set the flag bit value at index 34.
func (f Flags64) Set34() Flags64 {
	const mask = Flags64(1) << 34
	return f | mask
}

// Unset34 will unset the flag bit value at index 34.
func (f Flags64) Unset34() Flags64 {
	const mask = Flags64(1) << 34
	return f & ^mask
}

// Get35 will fetch the flag bit value at index 35.
func (f Flags64) Get35() bool {
	const mask = Flags64(1) << 35
	return (f&mask != 0)
}

// Set35 will set the flag bit value at index 35.
func (f Flags64) Set35() Flags64 {
	const mask = Flags64(1) << 35
	return f | mask
}

// Unset35 will unset the flag bit value at index 35.
func (f Flags64) Unset35() Flags64 {
	const mask = Flags64(1) << 35
	return f & ^mask
}

// Get36 will fetch the flag bit value at index 36.
func (f Flags64) Get36() bool {
	const mask = Flags64(1) << 36
	return (f&mask != 0)
}

// Set36 will set the flag bit value at index 36.
func (f Flags64) Set36() Flags64 {
	const mask = Flags64(1) << 36
	return f | mask
}

// Unset36 will unset the flag bit value at index 36.
func (f Flags64) Unset36() Flags64 {
	const mask = Flags64(1) << 36
	return f & ^mask
}

// Get37 will fetch the flag bit value at index 37.
func (f Flags64) Get37() bool {
	const mask = Flags64(1) << 37
	return (f&mask != 0)
}

// Set37 will set the flag bit value at index 37.
func (f Flags64) Set37() Flags64 {
	const mask = Flags64(1) << 37
	return f | mask
}

// Unset37 will unset the flag bit value at index 37.
func (f Flags64) Unset37() Flags64 {
	const mask = Flags64(1) << 37
	return f & ^mask
}

// Get38 will fetch the flag bit value at index 38.
func (f Flags64) Get38() bool {
	const mask = Flags64(1) << 38
	return (f&mask != 0)
}

// Set38 will set the flag bit value at index 38.
func (f Flags64) Set38() Flags64 {
	const mask = Flags64(1) << 38
	return f | mask
}

// Unset38 will unset the flag bit value at index 38.
func (f Flags64) Unset38() Flags64 {
	const mask = Flags64(1) << 38
	return f & ^mask
}

// Get39 will fetch the flag bit value at index 39.
func (f Flags64) Get39() bool {
	const mask = Flags64(1) << 39
	return (f&mask != 0)
}

// Set39 will set the flag bit value at index 39.
func (f Flags64) Set39() Flags64 {
	const mask = Flags64(1) << 39
	return f | mask
}

// Unset39 will unset the flag bit value at index 39.
func (f Flags64) Unset39() Flags64 {
	const mask = Flags64(1) << 39
	return f & ^mask
}

// Get40 will fetch the flag bit value at index 40.
func (f Flags64) Get40() bool {
	const mask = Flags64(1) << 40
	return (f&mask != 0)
}

// Set40 will set the flag bit value at index 40.
func (f Flags64) Set40() Flags64 {
	const mask = Flags64(1) << 40
	return f | mask
}

// Unset40 will unset the flag bit value at index 40.
func (f Flags64) Unset40() Flags64 {
	const mask = Flags64(1) << 40
	return f & ^mask
}

// Get41 will fetch the flag bit value at index 41.
func (f Flags64) Get41() bool {
	const mask = Flags64(1) << 41
	return (f&mask != 0)
}

// Set41 will set the flag bit value at index 41.
func (f Flags64) Set41() Flags64 {
	const mask = Flags64(1) << 41
	return f | mask
}

// Unset41 will unset the flag bit value at index 41.
func (f Flags64) Unset41() Flags64 {
	const mask = Flags64(1) << 41
	return f & ^mask
}

// Get42 will fetch the flag bit value at index 42.
func (f Flags64) Get42() bool {
	const mask = Flags64(1) << 42
	return (f&mask != 0)
}

// Set42 will set the flag bit value at index 42.
func (f Flags64) Set42() Flags64 {
	const mask = Flags64(1) << 42
	return f | mask
}

// Unset42 will unset the flag bit value at index 42.
func (f Flags64) Unset42() Flags64 {
	const mask = Flags64(1) << 42
	return f & ^mask
}

// Get43 will fetch the flag bit value at index 43.
func (f Flags64) Get43() bool {
	const mask = Flags64(1) << 43
	return (f&mask != 0)
}

// Set43 will set the flag bit value at index 43.
func (f Flags64) Set43() Flags64 {
	const mask = Flags64(1) << 43
	return f | mask
}

// Unset43 will unset the flag bit value at index 43.
func (f Flags64) Unset43() Flags64 {
	const mask = Flags64(1) << 43
	return f & ^mask
}

// Get44 will fetch the flag bit value at index 44.
func (f Flags64) Get44() bool {
	const mask = Flags64(1) << 44
	return (f&mask != 0)
}

// Set44 will set the flag bit value at index 44.
func (f Flags64) Set44() Flags64 {
	const mask = Flags64(1) << 44
	return f | mask
}

// Unset44 will unset the flag bit value at index 44.
func (f Flags64) Unset44() Flags64 {
	const mask = Flags64(1) << 44
	return f & ^mask
}

// Get45 will fetch the flag bit value at index 45.
func (f Flags64) Get45() bool {
	const mask = Flags64(1) << 45
	return (f&mask != 0)
}

// Set45 will set the flag bit value at index 45.
func (f Flags64) Set45() Flags64 {
	const mask = Flags64(1) << 45
	return f | mask
}

// Unset45 will unset the flag bit value at index 45.
func (f Flags64) Unset45() Flags64 {
	const mask = Flags64(1) << 45
	return f & ^mask
}

// Get46 will fetch the flag bit value at index 46.
func (f Flags64) Get46() bool {
	const mask = Flags64(1) << 46
	return (f&mask != 0)
}

// Set46 will set the flag bit value at index 46.
func (f Flags64) Set46() Flags64 {
	const mask = Flags64(1) << 46
	return f | mask
}

// Unset46 will unset the flag bit value at index 46.
func (f Flags64) Unset46() Flags64 {
	const mask = Flags64(1) << 46
	return f & ^mask
}

// Get47 will fetch the flag bit value at index 47.
func (f Flags64) Get47() bool {
	const mask = Flags64(1) << 47
	return (f&mask != 0)
}

// Set47 will set the flag bit value at index 47.
func (f Flags64) Set47() Flags64 {
	const mask = Flags64(1) << 47
	return f | mask
}

// Unset47 will unset the flag bit value at index 47.
func (f Flags64) Unset47() Flags64 {
	const mask = Flags64(1) << 47
	return f & ^mask
}

// Get48 will fetch the flag bit value at index 48.
func (f Flags64) Get48() bool {
	const mask = Flags64(1) << 48
	return (f&mask != 0)
}

// Set48 will set the flag bit value at index 48.
func (f Flags64) Set48() Flags64 {
	const mask = Flags64(1) << 48
	return f | mask
}

// Unset48 will unset the flag bit value at index 48.
func (f Flags64) Unset48() Flags64 {
	const mask = Flags64(1) << 48
	return f & ^mask
}

// Get49 will fetch the flag bit value at index 49.
func (f Flags64) Get49() bool {
	const mask = Flags64(1) << 49
	return (f&mask != 0)
}

// Set49 will set the flag bit value at index 49.
func (f Flags64) Set49() Flags64 {
	const mask = Flags64(1) << 49
	return f | mask
}

// Unset49 will unset the flag bit value at index 49.
func (f Flags64) Unset49() Flags64 {
	const mask = Flags64(1) << 49
	return f & ^mask
}

// Get50 will fetch the flag bit value at index 50.
func (f Flags64) Get50() bool {
	const mask = Flags64(1) << 50
	return (f&mask != 0)
}

// Set50 will set the flag bit value at index 50.
func (f Flags64) Set50() Flags64 {
	const mask = Flags64(1) << 50
	return f | mask
}

// Unset50 will unset the flag bit value at index 50.
func (f Flags64) Unset50() Flags64 {
	const mask = Flags64(1) << 50
	return f & ^mask
}

// Get51 will fetch the flag bit value at index 51.
func (f Flags64) Get51() bool {
	const mask = Flags64(1) << 51
	return (f&mask != 0)
}

// Set51 will set the flag bit value at index 51.
func (f Flags64) Set51() Flags64 {
	const mask = Flags64(1) << 51
	return f | mask
}

// Unset51 will unset the flag bit value at index 51.
func (f Flags64) Unset51() Flags64 {
	const mask = Flags64(1) << 51
	return f & ^mask
}

// Get52 will fetch the flag bit value at index 52.
func (f Flags64) Get52() bool {
	const mask = Flags64(1) << 52
	return (f&mask != 0)
}

// Set52 will set the flag bit value at index 52.
func (f Flags64) Set52() Flags64 {
	const mask = Flags64(1) << 52
	return f | mask
}

// Unset52 will unset the flag bit value at index 52.
func (f Flags64) Unset52() Flags64 {
	const mask = Flags64(1) << 52
	return f & ^mask
}

// Get53 will fetch the flag bit value at index 53.
func (f Flags64) Get53() bool {
	const mask = Flags64(1) << 53
	return (f&mask != 0)
}

// Set53 will set the flag bit value at index 53.
func (f Flags64) Set53() Flags64 {
	const mask = Flags64(1) << 53
	return f | mask
}

// Unset53 will unset the flag bit value at index 53.
func (f Flags64) Unset53() Flags64 {
	const mask = Flags64(1) << 53
	return f & ^mask
}

// Get54 will fetch the flag bit value at index 54.
func (f Flags64) Get54() bool {
	const mask = Flags64(1) << 54
	return (f&mask != 0)
}

// Set54 will set the flag bit value at index 54.
func (f Flags64) Set54() Flags64 {
	const mask = Flags64(1) << 54
	return f | mask
}

// Unset54 will unset the flag bit value at index 54.
func (f Flags64) Unset54() Flags64 {
	const mask = Flags64(1) << 54
	return f & ^mask
}

// Get55 will fetch the flag bit value at index 55.
func (f Flags64) Get55() bool {
	const mask = Flags64(1) << 55
	return (f&mask != 0)
}

// Set55 will set the flag bit value at index 55.
func (f Flags64) Set55() Flags64 {
	const mask = Flags64(1) << 55
	return f | mask
}

// Unset55 will unset the flag bit value at index 55.
func (f Flags64) Unset55() Flags64 {
	const mask = Flags64(1) << 55
	return f & ^mask
}

// Get56 will fetch the flag bit value at index 56.
func (f Flags64) Get56() bool {
	const mask = Flags64(1) << 56
	return (f&mask != 0)
}

// Set56 will set the flag bit value at index 56.
func (f Flags64) Set56() Flags64 {
	const mask = Flags64(1) << 56
	return f | mask
}

// Unset56 will unset the flag bit value at index 56.
func (f Flags64) Unset56() Flags64 {
	const mask = Flags64(1) << 56
	return f & ^mask
}

// Get57 will fetch the flag bit value at index 57.
func (f Flags64) Get57() bool {
	const mask = Flags64(1) << 57
	return (f&mask != 0)
}

// Set57 will set the flag bit value at index 57.
func (f Flags64) Set57() Flags64 {
	const mask = Flags64(1) << 57
	return f | mask
}

// Unset57 will unset the flag bit value at index 57.
func (f Flags64) Unset57() Flags64 {
	const mask = Flags64(1) << 57
	return f & ^mask
}

// Get58 will fetch the flag bit value at index 58.
func (f Flags64) Get58() bool {
	const mask = Flags64(1) << 58
	return (f&mask != 0)
}

// Set58 will set the flag bit value at index 58.
func (f Flags64) Set58() Flags64 {
	const mask = Flags64(1) << 58
	return f | mask
}

// Unset58 will unset the flag bit value at index 58.
func (f Flags64) Unset58() Flags64 {
	const mask = Flags64(1) << 58
	return f & ^mask
}

// Get59 will fetch the flag bit value at index 59.
func (f Flags64) Get59() bool {
	const mask = Flags64(1) << 59
	return (f&mask != 0)
}

// Set59 will set the flag bit value at index 59.
func (f Flags64) Set59() Flags64 {
	const mask = Flags64(1) << 59
	return f | mask
}

// Unset59 will unset the flag bit value at index 59.
func (f Flags64) Unset59() Flags64 {
	const mask = Flags64(1) << 59
	return f & ^mask
}

// Get60 will fetch the flag bit value at index 60.
func (f Flags64) Get60() bool {
	const mask = Flags64(1) << 60
	return (f&mask != 0)
}

// Set60 will set the flag bit value at index 60.
func (f Flags64) Set60() Flags64 {
	const mask = Flags64(1) << 60
	return f | mask
}

// Unset60 will unset the flag bit value at index 60.
func (f Flags64) Unset60() Flags64 {
	const mask = Flags64(1) << 60
	return f & ^mask
}

// Get61 will fetch the flag bit value at index 61.
func (f Flags64) Get61() bool {
	const mask = Flags64(1) << 61
	return (f&mask != 0)
}

// Set61 will set the flag bit value at index 61.
func (f Flags64) Set61() Flags64 {
	const mask = Flags64(1) << 61
	return f | mask
}

// Unset61 will unset the flag bit value at index 61.
func (f Flags64) Unset61() Flags64 {
	const mask = Flags64(1) << 61
	return f & ^mask
}

// Get62 will fetch the flag bit value at index 62.
func (f Flags64) Get62() bool {
	const mask = Flags64(1) << 62
	return (f&mask != 0)
}

// Set62 will set the flag bit value at index 62.
func (f Flags64) Set62() Flags64 {
	const mask = Flags64(1) << 62
	return f | mask
}

// Unset62 will unset the flag bit value at index 62.
func (f Flags64) Unset62() Flags64 {
	const mask = Flags64(1) << 62
	return f & ^mask
}

// Get63 will fetch the flag bit value at index 63.
func (f Flags64) Get63() bool {
	const mask = Flags64(1) << 63
	return (f&mask != 0)
}

// Set63 will set the flag bit value at index 63.
func (f Flags64) Set63() Flags64 {
	const mask = Flags64(1) << 63
	return f | mask
}

// Unset63 will unset the flag bit value at index 63.
func (f Flags64) Unset63() Flags64 {
	const mask = Flags64(1) << 63
	return f & ^mask
}

// String returns a human readable representation of Flags64.
func (f Flags64) String() string {
	var (
		i   int
		val bool
		buf []byte
	)

	// Make a prealloc est. based on longest-possible value
	const prealloc = 1 + (len("false ") * 64) - 1 + 1
	buf = make([]byte, prealloc)

	buf[i] = '{'
	i++

	val = f.Get0()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get1()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get2()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get3()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get4()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get5()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get6()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get7()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get8()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get9()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get10()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get11()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get12()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get13()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get14()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get15()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get16()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get17()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get18()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get19()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get20()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get21()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get22()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get23()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get24()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get25()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get26()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get27()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get28()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get29()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get30()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get31()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get32()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get33()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get34()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get35()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get36()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get37()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get38()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get39()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get40()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get41()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get42()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get43()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get44()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get45()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get46()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get47()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get48()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get49()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get50()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get51()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get52()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get53()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get54()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get55()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get56()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get57()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get58()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get59()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get60()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get61()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get62()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get63()
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	buf[i-1] = '}'
	buf = buf[:i]

	return *(*string)(unsafe.Pointer(&buf))
}

// GoString returns a more verbose human readable representation of Flags64.
func (f Flags64) GoString() string {
	var (
		i   int
		val bool
		buf []byte
	)

	// Make a prealloc est. based on longest-possible value
	const prealloc = len("bitutil.Flags64{") + (len("63=false ") * 64) - 1 + 1
	buf = make([]byte, prealloc)

	i += copy(buf[i:], "bitutil.Flags64{")

	val = f.Get0()
	i += copy(buf[i:], "0=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get1()
	i += copy(buf[i:], "1=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get2()
	i += copy(buf[i:], "2=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get3()
	i += copy(buf[i:], "3=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get4()
	i += copy(buf[i:], "4=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get5()
	i += copy(buf[i:], "5=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get6()
	i += copy(buf[i:], "6=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get7()
	i += copy(buf[i:], "7=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get8()
	i += copy(buf[i:], "8=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get9()
	i += copy(buf[i:], "9=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get10()
	i += copy(buf[i:], "10=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get11()
	i += copy(buf[i:], "11=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get12()
	i += copy(buf[i:], "12=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get13()
	i += copy(buf[i:], "13=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get14()
	i += copy(buf[i:], "14=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get15()
	i += copy(buf[i:], "15=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get16()
	i += copy(buf[i:], "16=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get17()
	i += copy(buf[i:], "17=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get18()
	i += copy(buf[i:], "18=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get19()
	i += copy(buf[i:], "19=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get20()
	i += copy(buf[i:], "20=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get21()
	i += copy(buf[i:], "21=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get22()
	i += copy(buf[i:], "22=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get23()
	i += copy(buf[i:], "23=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get24()
	i += copy(buf[i:], "24=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get25()
	i += copy(buf[i:], "25=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get26()
	i += copy(buf[i:], "26=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get27()
	i += copy(buf[i:], "27=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get28()
	i += copy(buf[i:], "28=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get29()
	i += copy(buf[i:], "29=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get30()
	i += copy(buf[i:], "30=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get31()
	i += copy(buf[i:], "31=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get32()
	i += copy(buf[i:], "32=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get33()
	i += copy(buf[i:], "33=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get34()
	i += copy(buf[i:], "34=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get35()
	i += copy(buf[i:], "35=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get36()
	i += copy(buf[i:], "36=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get37()
	i += copy(buf[i:], "37=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get38()
	i += copy(buf[i:], "38=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get39()
	i += copy(buf[i:], "39=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get40()
	i += copy(buf[i:], "40=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get41()
	i += copy(buf[i:], "41=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get42()
	i += copy(buf[i:], "42=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get43()
	i += copy(buf[i:], "43=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get44()
	i += copy(buf[i:], "44=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get45()
	i += copy(buf[i:], "45=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get46()
	i += copy(buf[i:], "46=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get47()
	i += copy(buf[i:], "47=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get48()
	i += copy(buf[i:], "48=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get49()
	i += copy(buf[i:], "49=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get50()
	i += copy(buf[i:], "50=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get51()
	i += copy(buf[i:], "51=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get52()
	i += copy(buf[i:], "52=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get53()
	i += copy(buf[i:], "53=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get54()
	i += copy(buf[i:], "54=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get55()
	i += copy(buf[i:], "55=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get56()
	i += copy(buf[i:], "56=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get57()
	i += copy(buf[i:], "57=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get58()
	i += copy(buf[i:], "58=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get59()
	i += copy(buf[i:], "59=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get60()
	i += copy(buf[i:], "60=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get61()
	i += copy(buf[i:], "61=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get62()
	i += copy(buf[i:], "62=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	val = f.Get63()
	i += copy(buf[i:], "63=")
	i += copy(buf[i:], bool2str(val))
	buf[i] = ' '
	i++

	buf[i-1] = '}'
	buf = buf[:i]

	return *(*string)(unsafe.Pointer(&buf))
}

func bool2str(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

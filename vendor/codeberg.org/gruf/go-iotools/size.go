package iotools

type Sizer interface {
	Size() int64
}

// SizerFunc is a function signature which allows
// a function to implement the Sizer type.
type SizerFunc func() int64

func (s SizerFunc) Size() int64 {
	return s()
}

type Lengther interface {
	Len() int
}

// LengthFunc is a function signature which allows
// a function to implement the Lengther type.
type LengthFunc func() int

func (l LengthFunc) Len() int {
	return l()
}

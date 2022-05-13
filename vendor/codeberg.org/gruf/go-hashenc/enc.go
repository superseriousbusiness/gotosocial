package hashenc

import (
	"encoding/base32"
	"encoding/base64"
	"encoding/hex"
)

// Encoder defines an interface for encoding binary data.
type Encoder interface {
	// Encode encodes the data at src into dst
	Encode(dst []byte, src []byte)

	// EncodedLen returns the encoded length for input data of supplied length
	EncodedLen(int) int
}

// Base32 returns a new base32 Encoder (StdEncoding, no padding).
func Base32() Encoder {
	return base32.StdEncoding.WithPadding(base64.NoPadding)
}

// Base64 returns a new base64 Encoder (URLEncoding, no padding).
func Base64() Encoder {
	return base64.URLEncoding.WithPadding(base64.NoPadding)
}

// Hex returns a new hex Encoder.
func Hex() Encoder {
	return &hexEncoder{}
}

// hexEncoder simply provides an empty receiver to satisfy Encoder.
type hexEncoder struct{}

func (*hexEncoder) Encode(dst []byte, src []byte) {
	hex.Encode(dst, src)
}

func (*hexEncoder) EncodedLen(len int) int {
	return hex.EncodedLen(len)
}

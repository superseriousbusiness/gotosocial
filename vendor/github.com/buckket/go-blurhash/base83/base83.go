package base83

import (
	"fmt"
	"math"
	"strings"
)

const characters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz#$%*+,-.:;=?@[]^_{|}~"

// An InvalidCharacterError occurs when a characters is found which is not part of the Base83 character set.
type InvalidCharacterError rune

func (e InvalidCharacterError) Error() string {
	return fmt.Sprintf("base83: invalid string (character %q out of range)", rune(e))
}

// An InvalidLengthError occurs when a given value cannot be encoded to a string of given length.
type InvalidLengthError int

func (e InvalidLengthError) Error() string {
	return fmt.Sprintf("base83: invalid length (%d)", int(e))
}

// Encode will encode the given integer value to a Base83 string with given length.
// If length is too short to encode the given value InvalidLengthError will be returned.
func Encode(value, length int) (string, error) {
	divisor := int(math.Pow(83, float64(length)))
	if value/divisor != 0 {
		return "", InvalidLengthError(length)
	}
	divisor /= 83

	var str strings.Builder
	str.Grow(length)
	for i := 0; i < length; i++ {
		if divisor <= 0 {
			return "", InvalidLengthError(length)
		}
		digit := (value / divisor) % 83
		divisor /= 83
		str.WriteRune(rune(characters[digit]))
	}

	return str.String(), nil
}

// Decode will decode the given Base83 string to an integer.
func Decode(str string) (value int, err error) {
	for _, r := range str {
		idx := strings.IndexRune(characters, r)
		if idx == -1 {
			return 0, InvalidCharacterError(r)
		}
		value = value*83 + idx
	}
	return value, nil
}

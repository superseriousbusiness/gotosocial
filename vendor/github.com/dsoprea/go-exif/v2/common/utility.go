package exifcommon

import (
	"bytes"
	"fmt"
	"time"

	"github.com/dsoprea/go-logging"
)

// DumpBytes prints a list of hex-encoded bytes.
func DumpBytes(data []byte) {
	fmt.Printf("DUMP: ")
	for _, x := range data {
		fmt.Printf("%02x ", x)
	}

	fmt.Printf("\n")
}

// DumpBytesClause prints a list like DumpBytes(), but encapsulated in
// "[]byte { ... }".
func DumpBytesClause(data []byte) {
	fmt.Printf("DUMP: ")

	fmt.Printf("[]byte { ")

	for i, x := range data {
		fmt.Printf("0x%02x", x)

		if i < len(data)-1 {
			fmt.Printf(", ")
		}
	}

	fmt.Printf(" }\n")
}

// DumpBytesToString returns a stringified list of hex-encoded bytes.
func DumpBytesToString(data []byte) string {
	b := new(bytes.Buffer)

	for i, x := range data {
		_, err := b.WriteString(fmt.Sprintf("%02x", x))
		log.PanicIf(err)

		if i < len(data)-1 {
			_, err := b.WriteRune(' ')
			log.PanicIf(err)
		}
	}

	return b.String()
}

// DumpBytesClauseToString returns a comma-separated list of hex-encoded bytes.
func DumpBytesClauseToString(data []byte) string {
	b := new(bytes.Buffer)

	for i, x := range data {
		_, err := b.WriteString(fmt.Sprintf("0x%02x", x))
		log.PanicIf(err)

		if i < len(data)-1 {
			_, err := b.WriteString(", ")
			log.PanicIf(err)
		}
	}

	return b.String()
}

// ExifFullTimestampString produces a string like "2018:11:30 13:01:49" from a
// `time.Time` struct. It will attempt to convert to UTC first.
func ExifFullTimestampString(t time.Time) (fullTimestampPhrase string) {
	t = t.UTC()

	return fmt.Sprintf("%04d:%02d:%02d %02d:%02d:%02d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
}

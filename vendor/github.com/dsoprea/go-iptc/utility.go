package iptc

import (
	"bytes"
	"fmt"

	"github.com/dsoprea/go-logging"
)

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

package pngstructure

import (
	"bytes"
	"fmt"
)

func DumpBytes(data []byte) {
	fmt.Printf("DUMP: ")
	for _, x := range data {
		fmt.Printf("%02x ", x)
	}

	fmt.Printf("\n")
}

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

func DumpBytesToString(data []byte) (string, error) {
	b := new(bytes.Buffer)

	for i, x := range data {
		if _, err := b.WriteString(fmt.Sprintf("%02x", x)); err != nil {
			return "", err
		}

		if i < len(data)-1 {
			if _, err := b.WriteRune(' '); err != nil {
				return "", err
			}
		}
	}

	return b.String(), nil
}

func DumpBytesClauseToString(data []byte) (string, error) {
	b := new(bytes.Buffer)

	for i, x := range data {
		if _, err := b.WriteString(fmt.Sprintf("0x%02x", x)); err != nil {
			return "", err
		}

		if i < len(data)-1 {
			if _, err := b.WriteString(", "); err != nil {
				return "", err
			}
		}
	}

	return b.String(), nil
}

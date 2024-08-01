package iptc

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"

	"encoding/binary"

	"github.com/dsoprea/go-logging"
)

var (
	// TODO(dustin): We're still not sure if this is the right endianness. No search to IPTC or IIM seems to state one or the other.

	// DefaultEncoding is the standard encoding for the IPTC format.
	defaultEncoding = binary.BigEndian
)

var (
	// ErrInvalidTagMarker indicates that the tag can not be parsed because the
	// tag boundary marker is not the expected value.
	ErrInvalidTagMarker = errors.New("invalid tag marker")
)

// Tag describes one tag read from the stream.
type Tag struct {
	recordNumber  uint8
	datasetNumber uint8
	dataSize      uint64
}

// String expresses state as a string.
func (tag *Tag) String() string {
	return fmt.Sprintf(
		"Tag<DATASET=(%d:%d) DATA-SIZE=(%d)>",
		tag.recordNumber, tag.datasetNumber, tag.dataSize)
}

// DecodeTag parses one tag from the stream.
func DecodeTag(r io.Reader) (tag Tag, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	tagMarker := uint8(0)
	err = binary.Read(r, defaultEncoding, &tagMarker)
	if err != nil {
		if err == io.EOF {
			return tag, err
		}

		log.Panic(err)
	}

	if tagMarker != 0x1c {
		return tag, ErrInvalidTagMarker
	}

	recordNumber := uint8(0)
	err = binary.Read(r, defaultEncoding, &recordNumber)
	log.PanicIf(err)

	datasetNumber := uint8(0)
	err = binary.Read(r, defaultEncoding, &datasetNumber)
	log.PanicIf(err)

	dataSize16Raw := uint16(0)
	err = binary.Read(r, defaultEncoding, &dataSize16Raw)
	log.PanicIf(err)

	var dataSize uint64

	if dataSize16Raw < 32768 {
		// We only had 16-bits (has the MSB set to (0)).
		dataSize = uint64(dataSize16Raw)
	} else {
		// This field is just the length of the length (has the MSB set to (1)).

		// Clear the MSB.
		lengthLength := dataSize16Raw & 32767

		if lengthLength == 4 {
			dataSize32Raw := uint32(0)
			err := binary.Read(r, defaultEncoding, &dataSize32Raw)
			log.PanicIf(err)

			dataSize = uint64(dataSize32Raw)
		} else if lengthLength == 8 {
			err := binary.Read(r, defaultEncoding, &dataSize)
			log.PanicIf(err)
		} else {
			// No specific sizes or limits are specified in the specification
			// so we need to impose our own limits in order to implement.

			log.Panicf("extended data-set tag size is not supported: (%d)", lengthLength)
		}
	}

	tag = Tag{
		recordNumber:  recordNumber,
		datasetNumber: datasetNumber,
		dataSize:      dataSize,
	}

	return tag, nil
}

// StreamTagKey is a convenience type that lets us key our index with a high-
// level type.
type StreamTagKey struct {
	// RecordNumber is the major classification of the dataset.
	RecordNumber uint8

	// DatasetNumber is the minor classification of the dataset.
	DatasetNumber uint8
}

// String returns a descriptive string.
func (stk StreamTagKey) String() string {
	return fmt.Sprintf("%d:%d", stk.RecordNumber, stk.DatasetNumber)
}

// Data is a convenience wrapper around a byte-slice.
type TagData []byte

// IsPrintable returns true if all characters are printable.
func (tg TagData) IsPrintable() bool {
	for _, b := range tg {
		r := rune(b)

		// Newline characters aren't considered printable.
		if r == 0x0d || r == 0x0a {
			continue
		}

		if unicode.IsGraphic(r) == false || unicode.IsPrint(r) == false {
			return false
		}
	}

	return true
}

// String returns a descriptive string. If the data doesn't include any non-
// printable characters, it will include the value itself.
func (tg TagData) String() string {
	if tg.IsPrintable() == true {
		return string(tg)
	} else {
		return fmt.Sprintf("BINARY<(%d) bytes>", len(tg))
	}
}

// ParsedTags is the complete, unordered set of tags parsed from the stream.
type ParsedTags map[StreamTagKey][]TagData

// ParseStream parses a serial sequence of tags and tag data out of the stream.
func ParseStream(r io.Reader) (tags map[StreamTagKey][]TagData, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	tags = make(ParsedTags)

	for {
		tag, err := DecodeTag(r)
		if err != nil {
			if err == io.EOF {
				break
			}

			log.Panic(err)
		}

		raw := make([]byte, tag.dataSize)

		_, err = io.ReadFull(r, raw)
		log.PanicIf(err)

		data := TagData(raw)

		stk := StreamTagKey{
			RecordNumber:  tag.recordNumber,
			DatasetNumber: tag.datasetNumber,
		}

		if existing, found := tags[stk]; found == true {
			tags[stk] = append(existing, data)
		} else {
			tags[stk] = []TagData{data}
		}
	}

	return tags, nil
}

// GetSimpleDictionaryFromParsedTags returns a dictionary of tag names to tag
// values, where all values are strings and any tag that had a non-printable
// value is omitted. We will also only return the first value, therefore
// dropping any follow-up values for repeatable tags. This will ignore non-
// standard tags. This will trim whitespace from the ends of strings.
//
// This is a convenience function for quickly displaying only the summary IPTC
// metadata that a user might actually be interested in at first glance.
func GetSimpleDictionaryFromParsedTags(pt ParsedTags) (distilled map[string]string) {
	distilled = make(map[string]string)

	for stk, dataSlice := range pt {
		sti, err := GetTagInfo(int(stk.RecordNumber), int(stk.DatasetNumber))
		if err != nil {
			if err == ErrTagNotStandard {
				continue
			} else {
				log.Panic(err)
			}
		}

		data := dataSlice[0]

		if data.IsPrintable() == false {
			continue
		}

		// TODO(dustin): Trim leading whitespace, too.
		distilled[sti.Description] = strings.Trim(string(data), "\r\n")
	}

	return distilled
}

// GetDictionaryFromParsedTags returns all tags. It will keep non-printable
// values, though will not print a placeholder instead. This will keep non-
// standard tags (and print the fully-qualified dataset ID rather than the
// name). It will keep repeated values (with the counter value appended to the
// end).
func GetDictionaryFromParsedTags(pt ParsedTags) (distilled map[string]string) {
	distilled = make(map[string]string)
	for stk, dataSlice := range pt {
		var keyPhrase string

		sti, err := GetTagInfo(int(stk.RecordNumber), int(stk.DatasetNumber))
		if err != nil {
			if err == ErrTagNotStandard {
				keyPhrase = fmt.Sprintf("%s (not a standard tag)", stk.String())
			} else {
				log.Panic(err)
			}
		} else {
			keyPhrase = sti.Description
		}

		for i, data := range dataSlice {
			currentKeyPhrase := keyPhrase
			if len(dataSlice) > 1 {
				currentKeyPhrase = fmt.Sprintf("%s (%d)", currentKeyPhrase, i+1)
			}

			var presentable string
			if data.IsPrintable() == false {
				presentable = fmt.Sprintf("[BINARY] %s", DumpBytesToString(data))
			} else {
				presentable = string(data)
			}

			distilled[currentKeyPhrase] = presentable
		}
	}

	return distilled
}

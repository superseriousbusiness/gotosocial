package exif

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/dsoprea/go-logging"
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

// ParseExifFullTimestamp parses dates like "2018:11:30 13:01:49" into a UTC
// `time.Time` struct.
func ParseExifFullTimestamp(fullTimestampPhrase string) (timestamp time.Time, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	parts := strings.Split(fullTimestampPhrase, " ")
	datestampValue, timestampValue := parts[0], parts[1]

	dateParts := strings.Split(datestampValue, ":")

	year, err := strconv.ParseUint(dateParts[0], 10, 16)
	if err != nil {
		log.Panicf("could not parse year")
	}

	month, err := strconv.ParseUint(dateParts[1], 10, 8)
	if err != nil {
		log.Panicf("could not parse month")
	}

	day, err := strconv.ParseUint(dateParts[2], 10, 8)
	if err != nil {
		log.Panicf("could not parse day")
	}

	timeParts := strings.Split(timestampValue, ":")

	hour, err := strconv.ParseUint(timeParts[0], 10, 8)
	if err != nil {
		log.Panicf("could not parse hour")
	}

	minute, err := strconv.ParseUint(timeParts[1], 10, 8)
	if err != nil {
		log.Panicf("could not parse minute")
	}

	second, err := strconv.ParseUint(timeParts[2], 10, 8)
	if err != nil {
		log.Panicf("could not parse second")
	}

	timestamp = time.Date(int(year), time.Month(month), int(day), int(hour), int(minute), int(second), 0, time.UTC)
	return timestamp, nil
}

// ExifFullTimestampString produces a string like "2018:11:30 13:01:49" from a
// `time.Time` struct. It will attempt to convert to UTC first.
func ExifFullTimestampString(t time.Time) (fullTimestampPhrase string) {
	t = t.UTC()

	return fmt.Sprintf("%04d:%02d:%02d %02d:%02d:%02d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
}

// ExifTag is one simple representation of a tag in a flat list of all of them.
type ExifTag struct {
	IfdPath string `json:"ifd_path"`

	TagId   uint16 `json:"id"`
	TagName string `json:"name"`

	TagTypeId   TagTypePrimitive `json:"type_id"`
	TagTypeName string           `json:"type_name"`
	Value       interface{}      `json:"value"`
	ValueBytes  []byte           `json:"value_bytes"`

	ChildIfdPath string `json:"child_ifd_path"`
}

// String returns a string representation.
func (et ExifTag) String() string {
	return fmt.Sprintf("ExifTag<IFD-PATH=[%s] TAG-ID=(0x%02x) TAG-NAME=[%s] TAG-TYPE=[%s] VALUE=[%v] VALUE-BYTES=(%d) CHILD-IFD-PATH=[%s]", et.IfdPath, et.TagId, et.TagName, et.TagTypeName, et.Value, len(et.ValueBytes), et.ChildIfdPath)
}

// GetFlatExifData returns a simple, flat representation of all tags.
func GetFlatExifData(exifData []byte) (exifTags []ExifTag, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	im := NewIfdMappingWithStandard()
	ti := NewTagIndex()

	_, index, err := Collect(im, ti, exifData)
	log.PanicIf(err)

	q := []*Ifd{index.RootIfd}

	exifTags = make([]ExifTag, 0)

	for len(q) > 0 {
		var ifd *Ifd
		ifd, q = q[0], q[1:]

		ti := NewTagIndex()
		for _, ite := range ifd.Entries {
			tagName := ""

			it, err := ti.Get(ifd.IfdPath, ite.TagId)
			if err != nil {
				// If it's a non-standard tag, just leave the name blank.
				if log.Is(err, ErrTagNotFound) != true {
					log.PanicIf(err)
				}
			} else {
				tagName = it.Name
			}

			value, err := ifd.TagValue(ite)
			if err != nil {
				if err == ErrUnhandledUnknownTypedTag {
					value = UnparseableUnknownTagValuePlaceholder
				} else {
					log.Panic(err)
				}
			}

			valueBytes, err := ifd.TagValueBytes(ite)
			if err != nil && err != ErrUnhandledUnknownTypedTag {
				log.Panic(err)
			}

			et := ExifTag{
				IfdPath:      ifd.IfdPath,
				TagId:        ite.TagId,
				TagName:      tagName,
				TagTypeId:    ite.TagType,
				TagTypeName:  TypeNames[ite.TagType],
				Value:        value,
				ValueBytes:   valueBytes,
				ChildIfdPath: ite.ChildIfdPath,
			}

			exifTags = append(exifTags, et)
		}

		for _, childIfd := range ifd.Children {
			q = append(q, childIfd)
		}

		if ifd.NextIfd != nil {
			q = append(q, ifd.NextIfd)
		}
	}

	return exifTags, nil
}

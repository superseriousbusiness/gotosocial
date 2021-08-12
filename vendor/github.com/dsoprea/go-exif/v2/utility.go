package exif

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/dsoprea/go-logging"

	"github.com/dsoprea/go-exif/v2/common"
	"github.com/dsoprea/go-exif/v2/undefined"
)

var (
	utilityLogger = log.NewLogger("exif.utility")
)

var (
	timeType = reflect.TypeOf(time.Time{})
)

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

	// Normalize the separators.
	datestampValue = strings.ReplaceAll(datestampValue, "-", ":")
	timestampValue = strings.ReplaceAll(timestampValue, "-", ":")

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
	return exifcommon.ExifFullTimestampString(t)
}

// ExifTag is one simple representation of a tag in a flat list of all of them.
type ExifTag struct {
	// IfdPath is the fully-qualified IFD path (even though it is not named as
	// such).
	IfdPath string `json:"ifd_path"`

	// TagId is the tag-ID.
	TagId uint16 `json:"id"`

	// TagName is the tag-name. This is never empty.
	TagName string `json:"name"`

	// UnitCount is the recorded number of units constution of the value.
	UnitCount uint32 `json:"unit_count"`

	// TagTypeId is the type-ID.
	TagTypeId exifcommon.TagTypePrimitive `json:"type_id"`

	// TagTypeName is the type name.
	TagTypeName string `json:"type_name"`

	// Value is the decoded value.
	Value interface{} `json:"value"`

	// ValueBytes is the raw, encoded value.
	ValueBytes []byte `json:"value_bytes"`

	// Formatted is the human representation of the first value (tag values are
	// always an array).
	FormattedFirst string `json:"formatted_first"`

	// Formatted is the human representation of the complete value.
	Formatted string `json:"formatted"`

	// ChildIfdPath is the name of the child IFD this tag represents (if it
	// represents any). Otherwise, this is empty.
	ChildIfdPath string `json:"child_ifd_path"`
}

// String returns a string representation.
func (et ExifTag) String() string {
	return fmt.Sprintf(
		"ExifTag<"+
			"IFD-PATH=[%s] "+
			"TAG-ID=(0x%02x) "+
			"TAG-NAME=[%s] "+
			"TAG-TYPE=[%s] "+
			"VALUE=[%v] "+
			"VALUE-BYTES=(%d) "+
			"CHILD-IFD-PATH=[%s]",
		et.IfdPath, et.TagId, et.TagName, et.TagTypeName, et.FormattedFirst,
		len(et.ValueBytes), et.ChildIfdPath)
}

// GetFlatExifData returns a simple, flat representation of all tags.
func GetFlatExifData(exifData []byte) (exifTags []ExifTag, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	eh, err := ParseExifHeader(exifData)
	log.PanicIf(err)

	im := NewIfdMappingWithStandard()
	ti := NewTagIndex()

	ie := NewIfdEnumerate(im, ti, exifData, eh.ByteOrder)

	exifTags = make([]ExifTag, 0)

	visitor := func(fqIfdPath string, ifdIndex int, ite *IfdTagEntry) (err error) {
		// This encodes down to base64. Since this an example tool and we do not
		// expect to ever decode the output, we are not worried about
		// specifically base64-encoding it in order to have a measure of
		// control.
		valueBytes, err := ite.GetRawBytes()
		if err != nil {
			if err == exifundefined.ErrUnparseableValue {
				return nil
			}

			log.Panic(err)
		}

		value, err := ite.Value()
		if err != nil {
			if err == exifcommon.ErrUnhandledUndefinedTypedTag {
				value = exifundefined.UnparseableUnknownTagValuePlaceholder
			} else {
				log.Panic(err)
			}
		}

		et := ExifTag{
			IfdPath:      fqIfdPath,
			TagId:        ite.TagId(),
			TagName:      ite.TagName(),
			UnitCount:    ite.UnitCount(),
			TagTypeId:    ite.TagType(),
			TagTypeName:  ite.TagType().String(),
			Value:        value,
			ValueBytes:   valueBytes,
			ChildIfdPath: ite.ChildIfdPath(),
		}

		et.Formatted, err = ite.Format()
		log.PanicIf(err)

		et.FormattedFirst, err = ite.FormatFirst()
		log.PanicIf(err)

		exifTags = append(exifTags, et)

		return nil
	}

	_, err = ie.Scan(exifcommon.IfdStandardIfdIdentity, eh.FirstIfdOffset, visitor)
	log.PanicIf(err)

	return exifTags, nil
}

// GpsDegreesEquals returns true if the two `GpsDegrees` are identical.
func GpsDegreesEquals(gi1, gi2 GpsDegrees) bool {
	if gi2.Orientation != gi1.Orientation {
		return false
	}

	degreesRightBound := math.Nextafter(gi1.Degrees, gi1.Degrees+1)
	minutesRightBound := math.Nextafter(gi1.Minutes, gi1.Minutes+1)
	secondsRightBound := math.Nextafter(gi1.Seconds, gi1.Seconds+1)

	if gi2.Degrees < gi1.Degrees || gi2.Degrees >= degreesRightBound {
		return false
	} else if gi2.Minutes < gi1.Minutes || gi2.Minutes >= minutesRightBound {
		return false
	} else if gi2.Seconds < gi1.Seconds || gi2.Seconds >= secondsRightBound {
		return false
	}

	return true
}

// IsTime returns true if the value is a `time.Time`.
func IsTime(v interface{}) bool {
	return reflect.TypeOf(v) == timeType
}

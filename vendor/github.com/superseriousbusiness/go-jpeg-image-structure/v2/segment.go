package jpegstructure

import (
	"bytes"
	"errors"
	"fmt"

	"crypto/sha1"
	"encoding/hex"

	"github.com/dsoprea/go-exif/v3"
	"github.com/dsoprea/go-exif/v3/common"
	"github.com/dsoprea/go-iptc"
	"github.com/dsoprea/go-logging"
	"github.com/dsoprea/go-photoshop-info-format"
	"github.com/dsoprea/go-utility/v2/image"
)

const (
	pirIptcImageResourceId = uint16(0x0404)
)

var (
	// exifPrefix is the prefix found at the top of an EXIF slice. This is JPEG-
	// specific.
	exifPrefix = []byte{'E', 'x', 'i', 'f', 0, 0}

	xmpPrefix = []byte("http://ns.adobe.com/xap/1.0/\000")

	ps30Prefix = []byte("Photoshop 3.0\000")
)

var (
	// ErrNoXmp is returned if XMP data was requested but not found.
	ErrNoXmp = errors.New("no XMP data")

	// ErrNoIptc is returned if IPTC data was requested but not found.
	ErrNoIptc = errors.New("no IPTC data")

	// ErrNoPhotoshopData is returned if Photoshop info was requested but not
	// found.
	ErrNoPhotoshopData = errors.New("no photoshop data")
)

// SofSegment has info read from a SOF segment.
type SofSegment struct {
	// BitsPerSample is the bits-per-sample.
	BitsPerSample byte

	// Width is the image width.
	Width uint16

	// Height is the image height.
	Height uint16

	// ComponentCount is the number of color components.
	ComponentCount byte
}

// String returns a string representation of the SOF segment.
func (ss SofSegment) String() string {

	// TODO(dustin): Add test

	return fmt.Sprintf("SOF<BitsPerSample=(%d) Width=(%d) Height=(%d) ComponentCount=(%d)>", ss.BitsPerSample, ss.Width, ss.Height, ss.ComponentCount)
}

// SegmentVisitor describes a segment-visitor struct.
type SegmentVisitor interface {
	// HandleSegment is triggered for each segment encountered as well as the
	// scan-data.
	HandleSegment(markerId byte, markerName string, counter int, lastIsScanData bool) error
}

// SofSegmentVisitor describes a visitor that is only called for each SOF
// segment.
type SofSegmentVisitor interface {
	// HandleSof is called for each encountered SOF segment.
	HandleSof(sof *SofSegment) error
}

// Segment describes a single segment.
type Segment struct {
	MarkerId   byte
	MarkerName string
	Offset     int
	Data       []byte

	photoshopInfo map[uint16]photoshopinfo.Photoshop30InfoRecord
	iptcTags      map[iptc.StreamTagKey][]iptc.TagData
}

// SetExif encodes and sets EXIF data into this segment.
func (s *Segment) SetExif(ib *exif.IfdBuilder) (err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	ibe := exif.NewIfdByteEncoder()

	exifData, err := ibe.EncodeToExif(ib)
	log.PanicIf(err)

	l := len(exifPrefix)

	s.Data = make([]byte, l+len(exifData))
	copy(s.Data[0:], exifPrefix)
	copy(s.Data[l:], exifData)

	return nil
}

// Exif returns an `exif.Ifd` instance for the EXIF data we currently have.
func (s *Segment) Exif() (rootIfd *exif.Ifd, data []byte, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	l := len(exifPrefix)

	rawExif := s.Data[l:]

	jpegLogger.Debugf(nil, "Attempting to parse (%d) byte EXIF blob (Exif).", len(rawExif))

	im, err := exifcommon.NewIfdMappingWithStandard()
	log.PanicIf(err)

	ti := exif.NewTagIndex()

	_, index, err := exif.Collect(im, ti, rawExif)
	log.PanicIf(err)

	return index.RootIfd, rawExif, nil
}

// FlatExif parses the EXIF data and just returns a list of tags.
func (s *Segment) FlatExif() (exifTags []exif.ExifTag, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	// TODO(dustin): Add test

	l := len(exifPrefix)

	rawExif := s.Data[l:]

	jpegLogger.Debugf(nil, "Attempting to parse (%d) byte EXIF blob (FlatExif).", len(rawExif))

	exifTags, _, err = exif.GetFlatExifData(rawExif, nil)
	log.PanicIf(err)

	return exifTags, nil
}

// EmbeddedString returns a string of properties that can be embedded into an
// longer string of properties.
func (s *Segment) EmbeddedString() string {
	h := sha1.New()
	h.Write(s.Data)

	// TODO(dustin): Add test

	digestString := hex.EncodeToString(h.Sum(nil))

	return fmt.Sprintf("OFFSET=(0x%08x %10d) ID=(0x%02x) NAME=[%-5s] SIZE=(%10d) SHA1=[%s]", s.Offset, s.Offset, s.MarkerId, markerNames[s.MarkerId], len(s.Data), digestString)
}

// String returns a descriptive string.
func (s *Segment) String() string {

	// TODO(dustin): Add test

	return fmt.Sprintf("Segment<%s>", s.EmbeddedString())
}

// IsExif returns true if EXIF data.
func (s *Segment) IsExif() bool {
	if s.MarkerId != MARKER_APP1 {
		return false
	}

	// TODO(dustin): Add test

	l := len(exifPrefix)

	if len(s.Data) < l {
		return false
	}

	if bytes.Equal(s.Data[:l], exifPrefix) == false {
		return false
	}

	return true
}

// IsXmp returns true if XMP data.
func (s *Segment) IsXmp() bool {
	if s.MarkerId != MARKER_APP1 {
		return false
	}

	// TODO(dustin): Add test

	l := len(xmpPrefix)

	if len(s.Data) < l {
		return false
	}

	if bytes.Equal(s.Data[:l], xmpPrefix) == false {
		return false
	}

	return true
}

// FormattedXmp returns a formatted XML string. This only makes sense for a
// segment comprised of XML data (like XMP).
func (s *Segment) FormattedXmp() (formatted string, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	// TODO(dustin): Add test

	if s.IsXmp() != true {
		log.Panicf("not an XMP segment")
	}

	l := len(xmpPrefix)

	raw := string(s.Data[l:])

	formatted, err = FormatXml(raw)
	log.PanicIf(err)

	return formatted, nil
}

func (s *Segment) parsePhotoshopInfo() (photoshopInfo map[uint16]photoshopinfo.Photoshop30InfoRecord, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	if s.photoshopInfo != nil {
		return s.photoshopInfo, nil
	}

	if s.MarkerId != MARKER_APP13 {
		return nil, ErrNoPhotoshopData
	}

	l := len(ps30Prefix)

	if len(s.Data) < l {
		return nil, ErrNoPhotoshopData
	}

	if bytes.Equal(s.Data[:l], ps30Prefix) == false {
		return nil, ErrNoPhotoshopData
	}

	data := s.Data[l:]
	b := bytes.NewBuffer(data)

	// Parse it.

	pirIndex, err := photoshopinfo.ReadPhotoshop30Info(b)
	log.PanicIf(err)

	s.photoshopInfo = pirIndex

	return s.photoshopInfo, nil
}

// IsIptc returns true if XMP data.
func (s *Segment) IsIptc() bool {
	// TODO(dustin): Add test

	// There's a cost to determining if there's IPTC data, so we won't do it
	// more than once.
	if s.iptcTags != nil {
		return true
	}

	photoshopInfo, err := s.parsePhotoshopInfo()
	if err != nil {
		if err == ErrNoPhotoshopData {
			return false
		}

		log.Panic(err)
	}

	// Bail if the Photoshop info doesn't have IPTC data.

	_, found := photoshopInfo[pirIptcImageResourceId]
	if found == false {
		return false
	}

	return true
}

// Iptc parses Photoshop info (if present) and then parses the IPTC info inside
// it (if present).
func (s *Segment) Iptc() (tags map[iptc.StreamTagKey][]iptc.TagData, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	// Cache the parse.
	if s.iptcTags != nil {
		return s.iptcTags, nil
	}

	photoshopInfo, err := s.parsePhotoshopInfo()
	log.PanicIf(err)

	iptcPir, found := photoshopInfo[pirIptcImageResourceId]
	if found == false {
		return nil, ErrNoIptc
	}

	b := bytes.NewBuffer(iptcPir.Data)

	tags, err = iptc.ParseStream(b)
	log.PanicIf(err)

	s.iptcTags = tags

	return tags, nil
}

var (
	// Enforce interface conformance.
	_ riimage.MediaContext = new(Segment)
)

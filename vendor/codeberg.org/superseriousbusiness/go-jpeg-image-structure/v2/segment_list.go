package jpegstructure

import (
	"bytes"
	"fmt"
	"io"

	"crypto/sha1"
	"encoding/binary"

	"github.com/dsoprea/go-exif/v3"
	"github.com/dsoprea/go-exif/v3/common"
	"github.com/dsoprea/go-iptc"
	"github.com/dsoprea/go-logging"
)

// SegmentList contains a slice of segments.
type SegmentList struct {
	segments []*Segment
}

// NewSegmentList returns a new SegmentList struct.
func NewSegmentList(segments []*Segment) (sl *SegmentList) {
	if segments == nil {
		segments = make([]*Segment, 0)
	}

	return &SegmentList{
		segments: segments,
	}
}

// OffsetsEqual returns true is all segments have the same marker-IDs and were
// found at the same offsets.
func (sl *SegmentList) OffsetsEqual(o *SegmentList) bool {
	if len(o.segments) != len(sl.segments) {
		return false
	}

	for i, s := range o.segments {
		if s.MarkerId != sl.segments[i].MarkerId || s.Offset != sl.segments[i].Offset {
			return false
		}
	}

	return true
}

// Segments returns the underlying slice of segments.
func (sl *SegmentList) Segments() []*Segment {
	return sl.segments
}

// Add adds another segment.
func (sl *SegmentList) Add(s *Segment) {
	sl.segments = append(sl.segments, s)
}

// Print prints segment info.
func (sl *SegmentList) Print() {
	if len(sl.segments) == 0 {
		fmt.Printf("No segments.\n")
	} else {
		exifIndex, _, err := sl.FindExif()
		if err != nil {
			if err == exif.ErrNoExif {
				exifIndex = -1
			} else {
				log.Panic(err)
			}
		}

		xmpIndex, _, err := sl.FindXmp()
		if err != nil {
			if err == ErrNoXmp {
				xmpIndex = -1
			} else {
				log.Panic(err)
			}
		}

		iptcIndex, _, err := sl.FindIptc()
		if err != nil {
			if err == ErrNoIptc {
				iptcIndex = -1
			} else {
				log.Panic(err)
			}
		}

		for i, s := range sl.segments {
			fmt.Printf("%2d: %s", i, s.EmbeddedString())

			if i == exifIndex {
				fmt.Printf(" [EXIF]")
			} else if i == xmpIndex {
				fmt.Printf(" [XMP]")
			} else if i == iptcIndex {
				fmt.Printf(" [IPTC]")
			}

			fmt.Printf("\n")
		}
	}
}

// Validate checks that all of the markers are actually located at all of the
// recorded offsets.
func (sl *SegmentList) Validate(data []byte) (err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	if len(sl.segments) < 2 {
		log.Panicf("minimum segments not found")
	}

	if sl.segments[0].MarkerId != MARKER_SOI {
		log.Panicf("first segment not SOI")
	} else if sl.segments[len(sl.segments)-1].MarkerId != MARKER_EOI {
		log.Panicf("last segment not EOI")
	}

	lastOffset := 0
	for i, s := range sl.segments {
		if lastOffset != 0 && s.Offset <= lastOffset {
			log.Panicf("segment offset not greater than the last: SEGMENT=(%d) (0x%08x) <= (0x%08x)", i, s.Offset, lastOffset)
		}

		// The scan-data doesn't start with a marker.
		if s.MarkerId == 0x0 {
			continue
		}

		o := s.Offset
		if bytes.Compare(data[o:o+2], []byte{0xff, s.MarkerId}) != 0 {
			log.Panicf("segment offset does not point to the start of a segment: SEGMENT=(%d) (0x%08x)", i, s.Offset)
		}

		lastOffset = o
	}

	return nil
}

// FindExif returns the the segment that hosts the EXIF data (if present).
func (sl *SegmentList) FindExif() (index int, segment *Segment, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	for i, s := range sl.segments {
		if s.IsExif() == true {
			return i, s, nil
		}
	}

	return -1, nil, exif.ErrNoExif
}

// FindXmp returns the the segment that hosts the XMP data (if present).
func (sl *SegmentList) FindXmp() (index int, segment *Segment, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	for i, s := range sl.segments {
		if s.IsXmp() == true {
			return i, s, nil
		}
	}

	return -1, nil, ErrNoXmp
}

// FindIptc returns the the segment that hosts the IPTC data (if present).
func (sl *SegmentList) FindIptc() (index int, segment *Segment, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	for i, s := range sl.segments {
		if s.IsIptc() == true {
			return i, s, nil
		}
	}

	return -1, nil, ErrNoIptc
}

// Exif returns an `exif.Ifd` instance for the EXIF data we currently have.
func (sl *SegmentList) Exif() (rootIfd *exif.Ifd, rawExif []byte, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	_, s, err := sl.FindExif()
	log.PanicIf(err)

	rootIfd, rawExif, err = s.Exif()
	log.PanicIf(err)

	return rootIfd, rawExif, nil
}

// Iptc returns embedded IPTC data if present.
func (sl *SegmentList) Iptc() (tags map[iptc.StreamTagKey][]iptc.TagData, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	// TODO(dustin): Add comment and return data.

	_, s, err := sl.FindIptc()
	log.PanicIf(err)

	tags, err = s.Iptc()
	log.PanicIf(err)

	return tags, nil
}

// ConstructExifBuilder returns an `exif.IfdBuilder` instance (needed for
// modifying) preloaded with all existing tags.
func (sl *SegmentList) ConstructExifBuilder() (rootIb *exif.IfdBuilder, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	rootIfd, _, err := sl.Exif()
	if log.Is(err, exif.ErrNoExif) == true {
		// No EXIF. Just create a boilerplate builder.

		im := exifcommon.NewIfdMapping()

		err := exifcommon.LoadStandardIfds(im)
		log.PanicIf(err)

		ti := exif.NewTagIndex()

		rootIb :=
			exif.NewIfdBuilder(
				im,
				ti,
				exifcommon.IfdStandardIfdIdentity,
				exifcommon.EncodeDefaultByteOrder)

		return rootIb, nil
	} else if err != nil {
		log.Panic(err)
	}

	rootIb = exif.NewIfdBuilderFromExistingChain(rootIfd)

	return rootIb, nil
}

// DumpExif returns an unstructured list of tags (useful when just reviewing).
func (sl *SegmentList) DumpExif() (segmentIndex int, segment *Segment, exifTags []exif.ExifTag, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	segmentIndex, s, err := sl.FindExif()
	if err != nil {
		if err == exif.ErrNoExif {
			return 0, nil, nil, err
		}

		log.Panic(err)
	}

	exifTags, err = s.FlatExif()
	log.PanicIf(err)

	return segmentIndex, s, exifTags, nil
}

func makeEmptyExifSegment() (s *Segment) {

	// TODO(dustin): Add test

	return &Segment{
		MarkerId: MARKER_APP1,
	}
}

// SetExif encodes and sets EXIF data into the given segment. If `index` is -1,
// append a new segment.
func (sl *SegmentList) SetExif(ib *exif.IfdBuilder) (err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	_, s, err := sl.FindExif()
	if err != nil {
		if log.Is(err, exif.ErrNoExif) == false {
			log.Panic(err)
		}

		s = makeEmptyExifSegment()

		prefix := sl.segments[:1]

		// Install it near the beginning where we know it's safe. We can't
		// insert it after the EOI segment, and there might be more than one
		// depending on implementation and/or lax adherence to the standard.
		tail := append([]*Segment{s}, sl.segments[1:]...)

		sl.segments = append(prefix, tail...)
	}

	err = s.SetExif(ib)
	log.PanicIf(err)

	return nil
}

// DropExif will drop the EXIF data if present.
func (sl *SegmentList) DropExif() (wasDropped bool, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	// TODO(dustin): Add test

	i, _, err := sl.FindExif()
	if err == nil {
		// Found.
		sl.segments = append(sl.segments[:i], sl.segments[i+1:]...)

		return true, nil
	} else if log.Is(err, exif.ErrNoExif) == false {
		log.Panic(err)
	}

	// Not found.
	return false, nil
}

// Write writes the segment data to the given `io.Writer`.
func (sl *SegmentList) Write(w io.Writer) (err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	offset := 0

	for i, s := range sl.segments {
		h := sha1.New()
		h.Write(s.Data)

		// The scan-data will have a marker-ID of (0) because it doesn't have a
		// marker-ID or length.
		if s.MarkerId != 0 {
			_, err := w.Write([]byte{0xff})
			log.PanicIf(err)

			offset++

			_, err = w.Write([]byte{s.MarkerId})
			log.PanicIf(err)

			offset++

			sizeLen, found := markerLen[s.MarkerId]
			if found == false || sizeLen == 2 {
				sizeLen = 2
				l := uint16(len(s.Data) + sizeLen)

				err = binary.Write(w, binary.BigEndian, &l)
				log.PanicIf(err)

				offset += 2
			} else if sizeLen == 4 {
				l := uint32(len(s.Data) + sizeLen)

				err = binary.Write(w, binary.BigEndian, &l)
				log.PanicIf(err)

				offset += 4
			} else if sizeLen != 0 {
				log.Panicf("not a supported marker-size: SEGMENT-INDEX=(%d) MARKER-ID=(0x%02x) MARKER-SIZE-LEN=(%d)", i, s.MarkerId, sizeLen)
			}
		}

		_, err := w.Write(s.Data)
		log.PanicIf(err)

		offset += len(s.Data)
	}

	return nil
}

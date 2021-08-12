package exifremove

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"image/jpeg"
	"image/png"

	"github.com/dsoprea/go-exif"
	jpegstructure "github.com/dsoprea/go-jpeg-image-structure"
	pngstructure "github.com/dsoprea/go-png-image-structure"
	"github.com/h2non/filetype"
)

func Remove(data []byte) ([]byte, error) {

	const (
		JpegMediaType  = "jpeg"
		PngMediaType   = "png"
		OtherMediaType = "other"
		StartBytes     = 0
		EndBytes       = 0
	)

	type MediaContext struct {
		MediaType string
		RootIfd   *exif.Ifd
		RawExif   []byte
		Media     interface{}
	}

	filtered := []byte{}

	head := make([]byte, 261)
	_, err := bytes.NewReader(data).Read(head)
	if err != nil {
		return nil, fmt.Errorf("could not read first 261 bytes of data: %s", err)
	}
	imagetype, err := filetype.Match(head)
	if err != nil {
		return nil, fmt.Errorf("error matching first 261 bytes of image to valid type: %s", err)
	}

	switch imagetype.MIME.Subtype {
	case "jpeg":
		jmp := jpegstructure.NewJpegMediaParser()
		sl, err := jmp.ParseBytes(data)
		if err != nil {
			return nil, err
		}

		_, rawExif, err := sl.Exif()
		if err != nil {
			return data, nil
		}

		startExifBytes := StartBytes
		endExifBytes := EndBytes

		if bytes.Contains(data, rawExif) {
			for i := 0; i < len(data)-len(rawExif); i++ {
				if bytes.Compare(data[i:i+len(rawExif)], rawExif) == 0 {
					startExifBytes = i
					endExifBytes = i + len(rawExif)
					break
				}
			}
			fill := make([]byte, len(data[startExifBytes:endExifBytes]))
			copy(data[startExifBytes:endExifBytes], fill)
		}

		filtered = data

		_, err = jpeg.Decode(bytes.NewReader(filtered))
		if err != nil {
			return nil, errors.New("EXIF removal corrupted " + err.Error())
		}
	case "png":
		pmp := pngstructure.NewPngMediaParser()
		cs, err := pmp.ParseBytes(data)
		if err != nil {
			return nil, err
		}

		_, rawExif, err := cs.Exif()
		if err != nil {
			return data, nil
		}

		startExifBytes := StartBytes
		endExifBytes := EndBytes

		if bytes.Contains(data, rawExif) {
			for i := 0; i < len(data)-len(rawExif); i++ {
				if bytes.Compare(data[i:i+len(rawExif)], rawExif) == 0 {
					startExifBytes = i
					endExifBytes = i + len(rawExif)
					break
				}
			}
			fill := make([]byte, len(data[startExifBytes:endExifBytes]))
			copy(data[startExifBytes:endExifBytes], fill)
		}

		filtered = data

		chunks := readPNGChunks(bytes.NewReader(filtered))

		for _, chunk := range chunks {
			if !chunk.CRCIsValid() {
				offset := int(chunk.Offset) + 8 + int(chunk.Length)
				crc := chunk.CalculateCRC()

				buf := new(bytes.Buffer)
				binary.Write(buf, binary.BigEndian, crc)
				crcBytes := buf.Bytes()

				copy(filtered[offset:], crcBytes)
			}
		}

		chunks = readPNGChunks(bytes.NewReader(filtered))
		for _, chunk := range chunks {
			if !chunk.CRCIsValid() {
				return nil, errors.New("EXIF removal failed CRC")
			}
		}

		_, err = png.Decode(bytes.NewReader(filtered))
		if err != nil {
			return nil, errors.New("EXIF removal corrupted " + err.Error())
		}
	default:
		return nil, errors.New("filetype not recognised")
	}

	return filtered, nil
}

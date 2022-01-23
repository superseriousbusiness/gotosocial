package pngstructure

import (
	"bytes"
	"fmt"

	"encoding/binary"

	"github.com/dsoprea/go-logging"
)

type ChunkDecoder struct {
}

func NewChunkDecoder() *ChunkDecoder {
	return new(ChunkDecoder)
}

func (cd *ChunkDecoder) Decode(c *Chunk) (decoded interface{}, err error) {
	defer func() {
		if state := recover(); state != nil {
			err := log.Wrap(state.(error))
			log.Panic(err)
		}
	}()

	switch c.Type {
	case "IHDR":
		ihdr, err := cd.decodeIHDR(c)
		log.PanicIf(err)

		return ihdr, nil
	}

	// We don't decode this particular type.
	return nil, nil
}

type ChunkIHDR struct {
	Width             uint32
	Height            uint32
	BitDepth          uint8
	ColorType         uint8
	CompressionMethod uint8
	FilterMethod      uint8
	InterlaceMethod   uint8
}

func (ihdr *ChunkIHDR) String() string {
	return fmt.Sprintf("IHDR<WIDTH=(%d) HEIGHT=(%d) DEPTH=(%d) COLOR-TYPE=(%d) COMP-METHOD=(%d) FILTER-METHOD=(%d) INTRLC-METHOD=(%d)>", ihdr.Width, ihdr.Height, ihdr.BitDepth, ihdr.ColorType, ihdr.CompressionMethod, ihdr.FilterMethod, ihdr.InterlaceMethod)
}

func (cd *ChunkDecoder) decodeIHDR(c *Chunk) (ihdr *ChunkIHDR, err error) {
	defer func() {
		if state := recover(); state != nil {
			err := log.Wrap(state.(error))
			log.Panic(err)
		}
	}()

	b := bytes.NewBuffer(c.Data)

	ihdr = new(ChunkIHDR)

	err = binary.Read(b, binary.BigEndian, &ihdr.Width)
	log.PanicIf(err)

	err = binary.Read(b, binary.BigEndian, &ihdr.Height)
	log.PanicIf(err)

	err = binary.Read(b, binary.BigEndian, &ihdr.BitDepth)
	log.PanicIf(err)

	err = binary.Read(b, binary.BigEndian, &ihdr.ColorType)
	log.PanicIf(err)

	err = binary.Read(b, binary.BigEndian, &ihdr.CompressionMethod)
	log.PanicIf(err)

	err = binary.Read(b, binary.BigEndian, &ihdr.FilterMethod)
	log.PanicIf(err)

	err = binary.Read(b, binary.BigEndian, &ihdr.InterlaceMethod)
	log.PanicIf(err)

	return ihdr, nil
}

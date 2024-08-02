package pngstructure

import (
	"bytes"
	"fmt"

	"encoding/binary"
)

type ChunkDecoder struct {
}

func NewChunkDecoder() *ChunkDecoder {
	return new(ChunkDecoder)
}

func (cd *ChunkDecoder) Decode(c *Chunk) (decoded interface{}, err error) {
	switch c.Type {
	case "IHDR":
		return cd.decodeIHDR(c)
	}

	// We don't decode this type.
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
	return fmt.Sprintf("IHDR<WIDTH=(%d) HEIGHT=(%d) DEPTH=(%d) COLOR-TYPE=(%d) COMP-METHOD=(%d) FILTER-METHOD=(%d) INTRLC-METHOD=(%d)>",
		ihdr.Width, ihdr.Height, ihdr.BitDepth, ihdr.ColorType, ihdr.CompressionMethod, ihdr.FilterMethod, ihdr.InterlaceMethod,
	)
}

func (cd *ChunkDecoder) decodeIHDR(c *Chunk) (*ChunkIHDR, error) {
	var (
		b     = bytes.NewBuffer(c.Data)
		ihdr  = new(ChunkIHDR)
		readf = func(data interface{}) error {
			return binary.Read(b, binary.BigEndian, data)
		}
	)

	if err := readf(&ihdr.Width); err != nil {
		return nil, err
	}

	if err := readf(&ihdr.Height); err != nil {
		return nil, err
	}

	if err := readf(&ihdr.BitDepth); err != nil {
		return nil, err
	}

	if err := readf(&ihdr.ColorType); err != nil {
		return nil, err
	}

	if err := readf(&ihdr.CompressionMethod); err != nil {
		return nil, err
	}

	if err := readf(&ihdr.FilterMethod); err != nil {
		return nil, err
	}

	if err := readf(&ihdr.InterlaceMethod); err != nil {
		return nil, err
	}

	return ihdr, nil
}

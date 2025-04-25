package jpegstructure

import (
	"github.com/dsoprea/go-logging"
)

const (
	// MARKER_SOI marker
	MARKER_SOI = 0xd8

	// MARKER_EOI marker
	MARKER_EOI = 0xd9

	// MARKER_SOS marker
	MARKER_SOS = 0xda

	// MARKER_SOD marker
	MARKER_SOD = 0x93

	// MARKER_DQT marker
	MARKER_DQT = 0xdb

	// MARKER_APP0 marker
	MARKER_APP0 = 0xe0

	// MARKER_APP1 marker
	MARKER_APP1 = 0xe1

	// MARKER_APP2 marker
	MARKER_APP2 = 0xe2

	// MARKER_APP3 marker
	MARKER_APP3 = 0xe3

	// MARKER_APP4 marker
	MARKER_APP4 = 0xe4

	// MARKER_APP5 marker
	MARKER_APP5 = 0xe5

	// MARKER_APP6 marker
	MARKER_APP6 = 0xe6

	// MARKER_APP7 marker
	MARKER_APP7 = 0xe7

	// MARKER_APP8 marker
	MARKER_APP8 = 0xe8

	// MARKER_APP10 marker
	MARKER_APP10 = 0xea

	// MARKER_APP12 marker
	MARKER_APP12 = 0xec

	// MARKER_APP13 marker
	MARKER_APP13 = 0xed

	// MARKER_APP14 marker
	MARKER_APP14 = 0xee

	// MARKER_APP15 marker
	MARKER_APP15 = 0xef

	// MARKER_COM marker
	MARKER_COM = 0xfe

	// MARKER_CME marker
	MARKER_CME = 0x64

	// MARKER_SIZ marker
	MARKER_SIZ = 0x51

	// MARKER_DHT marker
	MARKER_DHT = 0xc4

	// MARKER_JPG marker
	MARKER_JPG = 0xc8

	// MARKER_DAC marker
	MARKER_DAC = 0xcc

	// MARKER_SOF0 marker
	MARKER_SOF0 = 0xc0

	// MARKER_SOF1 marker
	MARKER_SOF1 = 0xc1

	// MARKER_SOF2 marker
	MARKER_SOF2 = 0xc2

	// MARKER_SOF3 marker
	MARKER_SOF3 = 0xc3

	// MARKER_SOF5 marker
	MARKER_SOF5 = 0xc5

	// MARKER_SOF6 marker
	MARKER_SOF6 = 0xc6

	// MARKER_SOF7 marker
	MARKER_SOF7 = 0xc7

	// MARKER_SOF9 marker
	MARKER_SOF9 = 0xc9

	// MARKER_SOF10 marker
	MARKER_SOF10 = 0xca

	// MARKER_SOF11 marker
	MARKER_SOF11 = 0xcb

	// MARKER_SOF13 marker
	MARKER_SOF13 = 0xcd

	// MARKER_SOF14 marker
	MARKER_SOF14 = 0xce

	// MARKER_SOF15 marker
	MARKER_SOF15 = 0xcf
)

var (
	jpegLogger        = log.NewLogger("jpegstructure.jpeg")
	jpegMagicStandard = []byte{0xff, MARKER_SOI, 0xff}
	jpegMagic2000     = []byte{0xff, 0x4f, 0xff}

	markerLen = map[byte]int{
		0x00: 0,
		0x01: 0,
		0xd0: 0,
		0xd1: 0,
		0xd2: 0,
		0xd3: 0,
		0xd4: 0,
		0xd5: 0,
		0xd6: 0,
		0xd7: 0,
		0xd8: 0,
		0xd9: 0,
		0xda: 0,

		// J2C
		0x30: 0,
		0x31: 0,
		0x32: 0,
		0x33: 0,
		0x34: 0,
		0x35: 0,
		0x36: 0,
		0x37: 0,
		0x38: 0,
		0x39: 0,
		0x3a: 0,
		0x3b: 0,
		0x3c: 0,
		0x3d: 0,
		0x3e: 0,
		0x3f: 0,
		0x4f: 0,
		0x92: 0,
		0x93: 0,

		// J2C extensions
		0x74: 4,
		0x75: 4,
		0x77: 4,
	}

	markerNames = map[byte]string{
		MARKER_SOI:   "SOI",
		MARKER_EOI:   "EOI",
		MARKER_SOS:   "SOS",
		MARKER_SOD:   "SOD",
		MARKER_DQT:   "DQT",
		MARKER_APP0:  "APP0",
		MARKER_APP1:  "APP1",
		MARKER_APP2:  "APP2",
		MARKER_APP3:  "APP3",
		MARKER_APP4:  "APP4",
		MARKER_APP5:  "APP5",
		MARKER_APP6:  "APP6",
		MARKER_APP7:  "APP7",
		MARKER_APP8:  "APP8",
		MARKER_APP10: "APP10",
		MARKER_APP12: "APP12",
		MARKER_APP13: "APP13",
		MARKER_APP14: "APP14",
		MARKER_APP15: "APP15",
		MARKER_COM:   "COM",
		MARKER_CME:   "CME",
		MARKER_SIZ:   "SIZ",

		MARKER_DHT: "DHT",
		MARKER_JPG: "JPG",
		MARKER_DAC: "DAC",

		MARKER_SOF0:  "SOF0",
		MARKER_SOF1:  "SOF1",
		MARKER_SOF2:  "SOF2",
		MARKER_SOF3:  "SOF3",
		MARKER_SOF5:  "SOF5",
		MARKER_SOF6:  "SOF6",
		MARKER_SOF7:  "SOF7",
		MARKER_SOF9:  "SOF9",
		MARKER_SOF10: "SOF10",
		MARKER_SOF11: "SOF11",
		MARKER_SOF13: "SOF13",
		MARKER_SOF14: "SOF14",
		MARKER_SOF15: "SOF15",
	}
)

package mp4

import (
	"fmt"

	"github.com/abema/go-mp4/internal/util"
)

/*************************** ilst ****************************/

func BoxTypeIlst() BoxType { return StrToBoxType("ilst") }
func BoxTypeData() BoxType { return StrToBoxType("data") }

var ilstMetaBoxTypes = []BoxType{
	StrToBoxType("----"),
	StrToBoxType("aART"),
	StrToBoxType("akID"),
	StrToBoxType("apID"),
	StrToBoxType("atID"),
	StrToBoxType("cmID"),
	StrToBoxType("cnID"),
	StrToBoxType("covr"),
	StrToBoxType("cpil"),
	StrToBoxType("cprt"),
	StrToBoxType("desc"),
	StrToBoxType("disk"),
	StrToBoxType("egid"),
	StrToBoxType("geID"),
	StrToBoxType("gnre"),
	StrToBoxType("pcst"),
	StrToBoxType("pgap"),
	StrToBoxType("plID"),
	StrToBoxType("purd"),
	StrToBoxType("purl"),
	StrToBoxType("rtng"),
	StrToBoxType("sfID"),
	StrToBoxType("soaa"),
	StrToBoxType("soal"),
	StrToBoxType("soar"),
	StrToBoxType("soco"),
	StrToBoxType("sonm"),
	StrToBoxType("sosn"),
	StrToBoxType("stik"),
	StrToBoxType("tmpo"),
	StrToBoxType("trkn"),
	StrToBoxType("tven"),
	StrToBoxType("tves"),
	StrToBoxType("tvnn"),
	StrToBoxType("tvsh"),
	StrToBoxType("tvsn"),
	{0xA9, 'A', 'R', 'T'},
	{0xA9, 'a', 'l', 'b'},
	{0xA9, 'c', 'm', 't'},
	{0xA9, 'c', 'o', 'm'},
	{0xA9, 'd', 'a', 'y'},
	{0xA9, 'g', 'e', 'n'},
	{0xA9, 'g', 'r', 'p'},
	{0xA9, 'n', 'a', 'm'},
	{0xA9, 't', 'o', 'o'},
	{0xA9, 'w', 'r', 't'},
}

func IsIlstMetaBoxType(boxType BoxType) bool {
	for _, bt := range ilstMetaBoxTypes {
		if boxType == bt {
			return true
		}
	}
	return false
}

func init() {
	AddBoxDef(&Ilst{})
	AddBoxDefEx(&Data{}, isUnderIlstMeta)
	for _, bt := range ilstMetaBoxTypes {
		AddAnyTypeBoxDefEx(&IlstMetaContainer{}, bt, isIlstMetaContainer)
	}
	AddAnyTypeBoxDefEx(&StringData{}, StrToBoxType("mean"), isUnderIlstFreeFormat)
	AddAnyTypeBoxDefEx(&StringData{}, StrToBoxType("name"), isUnderIlstFreeFormat)
}

type Ilst struct {
	Box
}

// GetType returns the BoxType
func (*Ilst) GetType() BoxType {
	return BoxTypeIlst()
}

type IlstMetaContainer struct {
	AnyTypeBox
}

func isIlstMetaContainer(ctx Context) bool {
	return ctx.UnderIlst && !ctx.UnderIlstMeta
}

const (
	DataTypeBinary             = 0
	DataTypeStringUTF8         = 1
	DataTypeStringUTF16        = 2
	DataTypeStringMac          = 3
	DataTypeStringJPEG         = 14
	DataTypeSignedIntBigEndian = 21
	DataTypeFloat32BigEndian   = 22
	DataTypeFloat64BigEndian   = 23
)

type Data struct {
	Box
	DataType uint32 `mp4:"0,size=32"`
	DataLang uint32 `mp4:"1,size=32"`
	Data     []byte `mp4:"2,size=8"`
}

// GetType returns the BoxType
func (*Data) GetType() BoxType {
	return BoxTypeData()
}

func isUnderIlstMeta(ctx Context) bool {
	return ctx.UnderIlstMeta
}

// StringifyField returns field value as string
func (data *Data) StringifyField(name string, indent string, depth int, ctx Context) (string, bool) {
	switch name {
	case "DataType":
		switch data.DataType {
		case DataTypeBinary:
			return "BINARY", true
		case DataTypeStringUTF8:
			return "UTF8", true
		case DataTypeStringUTF16:
			return "UTF16", true
		case DataTypeStringMac:
			return "MAC_STR", true
		case DataTypeStringJPEG:
			return "JPEG", true
		case DataTypeSignedIntBigEndian:
			return "INT", true
		case DataTypeFloat32BigEndian:
			return "FLOAT32", true
		case DataTypeFloat64BigEndian:
			return "FLOAT64", true
		}
	case "Data":
		switch data.DataType {
		case DataTypeStringUTF8:
			return fmt.Sprintf("\"%s\"", util.EscapeUnprintables(string(data.Data))), true
		}
	}
	return "", false
}

type StringData struct {
	AnyTypeBox
	Data []byte `mp4:"0,size=8"`
}

// StringifyField returns field value as string
func (sd *StringData) StringifyField(name string, indent string, depth int, ctx Context) (string, bool) {
	if name == "Data" {
		return fmt.Sprintf("\"%s\"", util.EscapeUnprintables(string(sd.Data))), true
	}
	return "", false
}

func isUnderIlstFreeFormat(ctx Context) bool {
	return ctx.UnderIlstFreeMeta
}

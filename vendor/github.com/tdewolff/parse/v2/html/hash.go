package html

// uses github.com/tdewolff/hasher
//go:generate hasher -type=Hash -file=hash.go

// Hash defines perfect hashes for a predefined list of strings
type Hash uint32

// Identifiers for the hashes associated with the text in the comments.
const (
	Iframe    Hash = 0x6    // iframe
	Math      Hash = 0x604  // math
	Plaintext Hash = 0x2109 // plaintext
	Script    Hash = 0xa06  // script
	Style     Hash = 0x1405 // style
	Svg       Hash = 0x1903 // svg
	Textarea  Hash = 0x2608 // textarea
	Title     Hash = 0xf05  // title
	Xml       Hash = 0x1c03 // xml
	Xmp       Hash = 0x1f03 // xmp
)

//var HashMap = map[string]Hash{
//	"iframe": Iframe,
//	"math": Math,
//	"plaintext": Plaintext,
//	"script": Script,
//	"style": Style,
//	"svg": Svg,
//	"textarea": Textarea,
//	"title": Title,
//	"xml": Xml,
//	"xmp": Xmp,
//}

// String returns the text associated with the hash.
func (i Hash) String() string {
	return string(i.Bytes())
}

// Bytes returns the text associated with the hash.
func (i Hash) Bytes() []byte {
	start := uint32(i >> 8)
	n := uint32(i & 0xff)
	if start+n > uint32(len(_Hash_text)) {
		return []byte{}
	}
	return _Hash_text[start : start+n]
}

// ToHash returns a hash Hash for a given []byte. Hash is a uint32 that is associated with the text in []byte. It returns zero if no match found.
func ToHash(s []byte) Hash {
	if len(s) == 0 || len(s) > _Hash_maxLen {
		return 0
	}
	//if 3 < len(s) {
	//	return HashMap[string(s)]
	//}
	h := uint32(_Hash_hash0)
	for i := 0; i < len(s); i++ {
		h ^= uint32(s[i])
		h *= 16777619
	}
	if i := _Hash_table[h&uint32(len(_Hash_table)-1)]; int(i&0xff) == len(s) {
		t := _Hash_text[i>>8 : i>>8+i&0xff]
		for i := 0; i < len(s); i++ {
			if t[i] != s[i] {
				goto NEXT
			}
		}
		return i
	}
NEXT:
	if i := _Hash_table[(h>>16)&uint32(len(_Hash_table)-1)]; int(i&0xff) == len(s) {
		t := _Hash_text[i>>8 : i>>8+i&0xff]
		for i := 0; i < len(s); i++ {
			if t[i] != s[i] {
				return 0
			}
		}
		return i
	}
	return 0
}

const _Hash_hash0 = 0xb4b790b3
const _Hash_maxLen = 9

var _Hash_text = []byte("" +
	"iframemathscriptitlestylesvgxmlxmplaintextarea")

var _Hash_table = [1 << 4]Hash{
	0x2: 0xa06,  // script
	0x3: 0xf05,  // title
	0x4: 0x1405, // style
	0x5: 0x604,  // math
	0x6: 0x6,    // iframe
	0x8: 0x1c03, // xml
	0x9: 0x2608, // textarea
	0xc: 0x1f03, // xmp
	0xe: 0x2109, // plaintext
	0xf: 0x1903, // svg
}

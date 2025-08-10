package html

// uses github.com/tdewolff/hasher
//go:generate hasher -type=Hash -file=hash.go

// Hash defines perfect hashes for a predefined list of strings
type Hash uint32

// Identifiers for the hashes associated with the text in the comments.
const (
	A                        Hash = 0x1     // a
	Abbr                     Hash = 0x40004 // abbr
	About                    Hash = 0x5     // about
	Accept                   Hash = 0xc06   // accept
	Accept_Charset           Hash = 0xc0e   // accept-charset
	Accesskey                Hash = 0x2c09  // accesskey
	Acronym                  Hash = 0x3507  // acronym
	Action                   Hash = 0x26006 // action
	Address                  Hash = 0x6d07  // address
	Allow                    Hash = 0x31f05 // allow
	Allowfullscreen          Hash = 0x31f0f // allowfullscreen
	Amp_Boilerplate          Hash = 0x5e0f  // amp-boilerplate
	Applet                   Hash = 0xee06  // applet
	Area                     Hash = 0x2c304 // area
	Article                  Hash = 0x22507 // article
	As                       Hash = 0x2102  // as
	Aside                    Hash = 0x9205  // aside
	Async                    Hash = 0x8a05  // async
	Audio                    Hash = 0x9d05  // audio
	Autocapitalize           Hash = 0xc30e  // autocapitalize
	Autocomplete             Hash = 0xd10c  // autocomplete
	Autofocus                Hash = 0xe309  // autofocus
	Autoplay                 Hash = 0xfc08  // autoplay
	B                        Hash = 0x101   // b
	Base                     Hash = 0x2004  // base
	Basefont                 Hash = 0x2008  // basefont
	Bb                       Hash = 0x40102 // bb
	Bdi                      Hash = 0x8303  // bdi
	Bdo                      Hash = 0x3dc03 // bdo
	Big                      Hash = 0x12f03 // big
	Blocking                 Hash = 0x13208 // blocking
	Blockquote               Hash = 0x13a0a // blockquote
	Body                     Hash = 0x804   // body
	Br                       Hash = 0x14b02 // br
	Button                   Hash = 0x14406 // button
	Canvas                   Hash = 0x8e06  // canvas
	Caption                  Hash = 0x23707 // caption
	Capture                  Hash = 0x10d07 // capture
	Center                   Hash = 0x24f06 // center
	Charset                  Hash = 0x1307  // charset
	Checked                  Hash = 0x37707 // checked
	Cite                     Hash = 0x14d04 // cite
	Class                    Hash = 0x15a05 // class
	Code                     Hash = 0x17604 // code
	Col                      Hash = 0x17f03 // col
	Colgroup                 Hash = 0x17f08 // colgroup
	Color                    Hash = 0x19e05 // color
	Cols                     Hash = 0x1a304 // cols
	Colspan                  Hash = 0x1a307 // colspan
	Content                  Hash = 0x1b107 // content
	Contenteditable          Hash = 0x1b10f // contenteditable
	Controls                 Hash = 0x1cc08 // controls
	Coords                   Hash = 0x1e306 // coords
	Crossorigin              Hash = 0x2160b // crossorigin
	Data                     Hash = 0xad04  // data
	Datalist                 Hash = 0xad08  // datalist
	Datatype                 Hash = 0x11908 // datatype
	Datetime                 Hash = 0x28508 // datetime
	Dd                       Hash = 0x6e02  // dd
	Decoding                 Hash = 0x9508  // decoding
	Default                  Hash = 0x17807 // default
	Defer                    Hash = 0x4405  // defer
	Del                      Hash = 0x1f203 // del
	Details                  Hash = 0x20b07 // details
	Dfn                      Hash = 0x16a03 // dfn
	Dialog                   Hash = 0x28d06 // dialog
	Dir                      Hash = 0x8403  // dir
	Disabled                 Hash = 0x19208 // disabled
	Div                      Hash = 0x19903 // div
	Dl                       Hash = 0x1c302 // dl
	Draggable                Hash = 0x1da09 // draggable
	Dt                       Hash = 0x40902 // dt
	Em                       Hash = 0xdc02  // em
	Embed                    Hash = 0x16605 // embed
	Enctype                  Hash = 0x26a07 // enctype
	Enterkeyhint             Hash = 0x2500c // enterkeyhint
	Fetchpriority            Hash = 0x1220d // fetchpriority
	Fieldset                 Hash = 0x22c08 // fieldset
	Figcaption               Hash = 0x2340a // figcaption
	Figure                   Hash = 0x24506 // figure
	Font                     Hash = 0x2404  // font
	Footer                   Hash = 0x1a06  // footer
	For                      Hash = 0x25c03 // for
	Form                     Hash = 0x25c04 // form
	Formaction               Hash = 0x25c0a // formaction
	Formenctype              Hash = 0x2660b // formenctype
	Formmethod               Hash = 0x2710a // formmethod
	Formnovalidate           Hash = 0x27b0e // formnovalidate
	Formtarget               Hash = 0x2930a // formtarget
	Frame                    Hash = 0x16e05 // frame
	Frameset                 Hash = 0x16e08 // frameset
	H1                       Hash = 0x2d502 // h1
	H2                       Hash = 0x38602 // h2
	H3                       Hash = 0x39502 // h3
	H4                       Hash = 0x40b02 // h4
	H5                       Hash = 0x29d02 // h5
	H6                       Hash = 0x29f02 // h6
	Head                     Hash = 0x36c04 // head
	Header                   Hash = 0x36c06 // header
	Headers                  Hash = 0x36c07 // headers
	Height                   Hash = 0x2a106 // height
	Hgroup                   Hash = 0x2b506 // hgroup
	Hidden                   Hash = 0x2cc06 // hidden
	High                     Hash = 0x2d204 // high
	Hr                       Hash = 0x2d702 // hr
	Href                     Hash = 0x2d704 // href
	Hreflang                 Hash = 0x2d708 // hreflang
	Html                     Hash = 0x2a504 // html
	Http_Equiv               Hash = 0x2df0a // http-equiv
	I                        Hash = 0x2801  // i
	Id                       Hash = 0x9402  // id
	Iframe                   Hash = 0x2f206 // iframe
	Image                    Hash = 0x30005 // image
	Imagesizes               Hash = 0x3000a // imagesizes
	Imagesrcset              Hash = 0x30d0b // imagesrcset
	Img                      Hash = 0x31803 // img
	Inert                    Hash = 0x10805 // inert
	Inlist                   Hash = 0x21f06 // inlist
	Input                    Hash = 0x3d05  // input
	Inputmode                Hash = 0x3d09  // inputmode
	Ins                      Hash = 0x31b03 // ins
	Is                       Hash = 0xb202  // is
	Ismap                    Hash = 0x32e05 // ismap
	Itemid                   Hash = 0x2fa06 // itemid
	Itemprop                 Hash = 0x14e08 // itemprop
	Itemref                  Hash = 0x34507 // itemref
	Itemscope                Hash = 0x35709 // itemscope
	Itemtype                 Hash = 0x36108 // itemtype
	Kbd                      Hash = 0x8203  // kbd
	Kind                     Hash = 0xaa04  // kind
	Label                    Hash = 0x1c405 // label
	Lang                     Hash = 0x2db04 // lang
	Legend                   Hash = 0x1be06 // legend
	Li                       Hash = 0xb102  // li
	Link                     Hash = 0x1c804 // link
	List                     Hash = 0xb104  // list
	Loading                  Hash = 0x3ad07 // loading
	Loop                     Hash = 0x2a804 // loop
	Low                      Hash = 0x32103 // low
	Main                     Hash = 0x3b04  // main
	Map                      Hash = 0xed03  // map
	Mark                     Hash = 0x7f04  // mark
	Marquee                  Hash = 0x3e407 // marquee
	Math                     Hash = 0x36904 // math
	Max                      Hash = 0x37e03 // max
	Maxlength                Hash = 0x37e09 // maxlength
	Media                    Hash = 0x28b05 // media
	Menu                     Hash = 0x2f604 // menu
	Menuitem                 Hash = 0x2f608 // menuitem
	Meta                     Hash = 0x5004  // meta
	Meter                    Hash = 0x38805 // meter
	Method                   Hash = 0x27506 // method
	Min                      Hash = 0x38d03 // min
	Minlength                Hash = 0x38d09 // minlength
	Multiple                 Hash = 0x39708 // multiple
	Muted                    Hash = 0x39f05 // muted
	Name                     Hash = 0x4e04  // name
	Nav                      Hash = 0xbc03  // nav
	Nobr                     Hash = 0x14904 // nobr
	Noembed                  Hash = 0x16407 // noembed
	Noframes                 Hash = 0x16c08 // noframes
	Nomodule                 Hash = 0x1a908 // nomodule
	Noscript                 Hash = 0x23d08 // noscript
	Novalidate               Hash = 0x27f0a // novalidate
	Object                   Hash = 0xa106  // object
	Ol                       Hash = 0x18002 // ol
	Open                     Hash = 0x35d04 // open
	Optgroup                 Hash = 0x2aa08 // optgroup
	Optimum                  Hash = 0x3de07 // optimum
	Option                   Hash = 0x2ec06 // option
	Output                   Hash = 0x206   // output
	P                        Hash = 0x501   // p
	Param                    Hash = 0x7b05  // param
	Pattern                  Hash = 0xb607  // pattern
	Picture                  Hash = 0x18607 // picture
	Ping                     Hash = 0x2b104 // ping
	Plaintext                Hash = 0x2ba09 // plaintext
	Playsinline              Hash = 0x1000b // playsinline
	Popover                  Hash = 0x33207 // popover
	Popovertarget            Hash = 0x3320d // popovertarget
	Popovertargetaction      Hash = 0x33213 // popovertargetaction
	Portal                   Hash = 0x3f406 // portal
	Poster                   Hash = 0x41006 // poster
	Pre                      Hash = 0x3a403 // pre
	Prefix                   Hash = 0x3a406 // prefix
	Preload                  Hash = 0x3aa07 // preload
	Profile                  Hash = 0x3b407 // profile
	Progress                 Hash = 0x3bb08 // progress
	Property                 Hash = 0x15208 // property
	Q                        Hash = 0x11401 // q
	Rb                       Hash = 0x1f02  // rb
	Readonly                 Hash = 0x2c408 // readonly
	Referrerpolicy           Hash = 0x3490e // referrerpolicy
	Rel                      Hash = 0x3ab03 // rel
	Required                 Hash = 0x11208 // required
	Resource                 Hash = 0x24908 // resource
	Rev                      Hash = 0x18b03 // rev
	Reversed                 Hash = 0x18b08 // reversed
	Rows                     Hash = 0x4804  // rows
	Rowspan                  Hash = 0x4807  // rowspan
	Rp                       Hash = 0x6702  // rp
	Rt                       Hash = 0x10b02 // rt
	Rtc                      Hash = 0x10b03 // rtc
	Ruby                     Hash = 0x8604  // ruby
	S                        Hash = 0x1701  // s
	Samp                     Hash = 0x5d04  // samp
	Sandbox                  Hash = 0x7307  // sandbox
	Scope                    Hash = 0x35b05 // scope
	Script                   Hash = 0x23f06 // script
	Section                  Hash = 0x15e07 // section
	Select                   Hash = 0x1d306 // select
	Selected                 Hash = 0x1d308 // selected
	Shadowrootdelegatesfocus Hash = 0x1e818 // shadowrootdelegatesfocus
	Shadowrootmode           Hash = 0x1ff0e // shadowrootmode
	Shape                    Hash = 0x21105 // shape
	Size                     Hash = 0x30504 // size
	Sizes                    Hash = 0x30505 // sizes
	Slot                     Hash = 0x30904 // slot
	Small                    Hash = 0x31d05 // small
	Source                   Hash = 0x24b06 // source
	Span                     Hash = 0x4b04  // span
	Spellcheck               Hash = 0x3720a // spellcheck
	Src                      Hash = 0x31203 // src
	Srclang                  Hash = 0x3c207 // srclang
	Srcset                   Hash = 0x31206 // srcset
	Start                    Hash = 0x22305 // start
	Step                     Hash = 0xb304  // step
	Strike                   Hash = 0x3c906 // strike
	Strong                   Hash = 0x3cf06 // strong
	Style                    Hash = 0x3d505 // style
	Sub                      Hash = 0x3da03 // sub
	Summary                  Hash = 0x3eb07 // summary
	Sup                      Hash = 0x3f203 // sup
	Svg                      Hash = 0x3fa03 // svg
	Tabindex                 Hash = 0x5208  // tabindex
	Table                    Hash = 0x1bb05 // table
	Target                   Hash = 0x29706 // target
	Tbody                    Hash = 0x705   // tbody
	Td                       Hash = 0x1f102 // td
	Template                 Hash = 0xdb08  // template
	Text                     Hash = 0x2bf04 // text
	Textarea                 Hash = 0x2bf08 // textarea
	Tfoot                    Hash = 0x1905  // tfoot
	Th                       Hash = 0x27702 // th
	Thead                    Hash = 0x36b05 // thead
	Time                     Hash = 0x28904 // time
	Title                    Hash = 0x2705  // title
	Tr                       Hash = 0xa602  // tr
	Track                    Hash = 0xa605  // track
	Translate                Hash = 0xf309  // translate
	Tt                       Hash = 0xb802  // tt
	Type                     Hash = 0x11d04 // type
	Typeof                   Hash = 0x11d06 // typeof
	U                        Hash = 0x301   // u
	Ul                       Hash = 0x17c02 // ul
	Usemap                   Hash = 0xea06  // usemap
	Value                    Hash = 0xbe05  // value
	Var                      Hash = 0x19b03 // var
	Video                    Hash = 0x2e805 // video
	Vocab                    Hash = 0x3fd05 // vocab
	Wbr                      Hash = 0x40403 // wbr
	Width                    Hash = 0x40705 // width
	Wrap                     Hash = 0x40d04 // wrap
	Xmlns                    Hash = 0x5905  // xmlns
	Xmp                      Hash = 0x7903  // xmp
)

//var HashMap = map[string]Hash{
//	"a": A,
//	"abbr": Abbr,
//	"about": About,
//	"accept": Accept,
//	"accept-charset": Accept_Charset,
//	"accesskey": Accesskey,
//	"acronym": Acronym,
//	"action": Action,
//	"address": Address,
//	"allow": Allow,
//	"allowfullscreen": Allowfullscreen,
//	"amp-boilerplate": Amp_Boilerplate,
//	"applet": Applet,
//	"area": Area,
//	"article": Article,
//	"as": As,
//	"aside": Aside,
//	"async": Async,
//	"audio": Audio,
//	"autocapitalize": Autocapitalize,
//	"autocomplete": Autocomplete,
//	"autofocus": Autofocus,
//	"autoplay": Autoplay,
//	"b": B,
//	"base": Base,
//	"basefont": Basefont,
//	"bb": Bb,
//	"bdi": Bdi,
//	"bdo": Bdo,
//	"big": Big,
//	"blocking": Blocking,
//	"blockquote": Blockquote,
//	"body": Body,
//	"br": Br,
//	"button": Button,
//	"canvas": Canvas,
//	"caption": Caption,
//	"capture": Capture,
//	"center": Center,
//	"charset": Charset,
//	"checked": Checked,
//	"cite": Cite,
//	"class": Class,
//	"code": Code,
//	"col": Col,
//	"colgroup": Colgroup,
//	"color": Color,
//	"cols": Cols,
//	"colspan": Colspan,
//	"content": Content,
//	"contenteditable": Contenteditable,
//	"controls": Controls,
//	"coords": Coords,
//	"crossorigin": Crossorigin,
//	"data": Data,
//	"datalist": Datalist,
//	"datatype": Datatype,
//	"datetime": Datetime,
//	"dd": Dd,
//	"decoding": Decoding,
//	"default": Default,
//	"defer": Defer,
//	"del": Del,
//	"details": Details,
//	"dfn": Dfn,
//	"dialog": Dialog,
//	"dir": Dir,
//	"disabled": Disabled,
//	"div": Div,
//	"dl": Dl,
//	"draggable": Draggable,
//	"dt": Dt,
//	"em": Em,
//	"embed": Embed,
//	"enctype": Enctype,
//	"enterkeyhint": Enterkeyhint,
//	"fetchpriority": Fetchpriority,
//	"fieldset": Fieldset,
//	"figcaption": Figcaption,
//	"figure": Figure,
//	"font": Font,
//	"footer": Footer,
//	"for": For,
//	"form": Form,
//	"formaction": Formaction,
//	"formenctype": Formenctype,
//	"formmethod": Formmethod,
//	"formnovalidate": Formnovalidate,
//	"formtarget": Formtarget,
//	"frame": Frame,
//	"frameset": Frameset,
//	"h1": H1,
//	"h2": H2,
//	"h3": H3,
//	"h4": H4,
//	"h5": H5,
//	"h6": H6,
//	"head": Head,
//	"header": Header,
//	"headers": Headers,
//	"height": Height,
//	"hgroup": Hgroup,
//	"hidden": Hidden,
//	"high": High,
//	"hr": Hr,
//	"href": Href,
//	"hreflang": Hreflang,
//	"html": Html,
//	"http-equiv": Http_Equiv,
//	"i": I,
//	"id": Id,
//	"iframe": Iframe,
//	"image": Image,
//	"imagesizes": Imagesizes,
//	"imagesrcset": Imagesrcset,
//	"img": Img,
//	"inert": Inert,
//	"inlist": Inlist,
//	"input": Input,
//	"inputmode": Inputmode,
//	"ins": Ins,
//	"is": Is,
//	"ismap": Ismap,
//	"itemid": Itemid,
//	"itemprop": Itemprop,
//	"itemref": Itemref,
//	"itemscope": Itemscope,
//	"itemtype": Itemtype,
//	"kbd": Kbd,
//	"kind": Kind,
//	"label": Label,
//	"lang": Lang,
//	"legend": Legend,
//	"li": Li,
//	"link": Link,
//	"list": List,
//	"loading": Loading,
//	"loop": Loop,
//	"low": Low,
//	"main": Main,
//	"map": Map,
//	"mark": Mark,
//	"marquee": Marquee,
//	"math": Math,
//	"max": Max,
//	"maxlength": Maxlength,
//	"media": Media,
//	"menu": Menu,
//	"menuitem": Menuitem,
//	"meta": Meta,
//	"meter": Meter,
//	"method": Method,
//	"min": Min,
//	"minlength": Minlength,
//	"multiple": Multiple,
//	"muted": Muted,
//	"name": Name,
//	"nav": Nav,
//	"nobr": Nobr,
//	"noembed": Noembed,
//	"noframes": Noframes,
//	"nomodule": Nomodule,
//	"noscript": Noscript,
//	"novalidate": Novalidate,
//	"object": Object,
//	"ol": Ol,
//	"open": Open,
//	"optgroup": Optgroup,
//	"optimum": Optimum,
//	"option": Option,
//	"output": Output,
//	"p": P,
//	"param": Param,
//	"pattern": Pattern,
//	"picture": Picture,
//	"ping": Ping,
//	"plaintext": Plaintext,
//	"playsinline": Playsinline,
//	"popover": Popover,
//	"popovertarget": Popovertarget,
//	"popovertargetaction": Popovertargetaction,
//	"portal": Portal,
//	"poster": Poster,
//	"pre": Pre,
//	"prefix": Prefix,
//	"preload": Preload,
//	"profile": Profile,
//	"progress": Progress,
//	"property": Property,
//	"q": Q,
//	"rb": Rb,
//	"readonly": Readonly,
//	"referrerpolicy": Referrerpolicy,
//	"rel": Rel,
//	"required": Required,
//	"resource": Resource,
//	"rev": Rev,
//	"reversed": Reversed,
//	"rows": Rows,
//	"rowspan": Rowspan,
//	"rp": Rp,
//	"rt": Rt,
//	"rtc": Rtc,
//	"ruby": Ruby,
//	"s": S,
//	"samp": Samp,
//	"sandbox": Sandbox,
//	"scope": Scope,
//	"script": Script,
//	"section": Section,
//	"select": Select,
//	"selected": Selected,
//	"shadowrootdelegatesfocus": Shadowrootdelegatesfocus,
//	"shadowrootmode": Shadowrootmode,
//	"shape": Shape,
//	"size": Size,
//	"sizes": Sizes,
//	"slot": Slot,
//	"small": Small,
//	"source": Source,
//	"span": Span,
//	"spellcheck": Spellcheck,
//	"src": Src,
//	"srclang": Srclang,
//	"srcset": Srcset,
//	"start": Start,
//	"step": Step,
//	"strike": Strike,
//	"strong": Strong,
//	"style": Style,
//	"sub": Sub,
//	"summary": Summary,
//	"sup": Sup,
//	"svg": Svg,
//	"tabindex": Tabindex,
//	"table": Table,
//	"target": Target,
//	"tbody": Tbody,
//	"td": Td,
//	"template": Template,
//	"text": Text,
//	"textarea": Textarea,
//	"tfoot": Tfoot,
//	"th": Th,
//	"thead": Thead,
//	"time": Time,
//	"title": Title,
//	"tr": Tr,
//	"track": Track,
//	"translate": Translate,
//	"tt": Tt,
//	"type": Type,
//	"typeof": Typeof,
//	"u": U,
//	"ul": Ul,
//	"usemap": Usemap,
//	"value": Value,
//	"var": Var,
//	"video": Video,
//	"vocab": Vocab,
//	"wbr": Wbr,
//	"width": Width,
//	"wrap": Wrap,
//	"xmlns": Xmlns,
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

const _Hash_hash0 = 0x87d8a7d9
const _Hash_maxLen = 24

var _Hash_text = []byte("" +
	"aboutputbodyaccept-charsetfooterbasefontitleaccesskeyacronym" +
	"ainputmodeferowspanametabindexmlnsamp-boilerplateaddressandb" +
	"oxmparamarkbdirubyasyncanvasidecodingaudiobjectrackindatalis" +
	"tepatternavalueautocapitalizeautocompletemplateautofocusemap" +
	"pletranslateautoplaysinlinertcapturequiredatatypeofetchprior" +
	"itybigblockingblockquotebuttonobrcitempropertyclassectionoem" +
	"bedfnoframesetcodefaultcolgroupictureversedisabledivarcolorc" +
	"olspanomodulecontenteditablegendlabelinkcontrolselectedragga" +
	"blecoordshadowrootdelegatesfocushadowrootmodetailshapecrosso" +
	"riginlistarticlefieldsetfigcaptionoscriptfiguresourcenterkey" +
	"hintformactionformenctypeformmethodformnovalidatetimedialogf" +
	"ormtargeth5h6heightmlooptgroupinghgrouplaintextareadonlyhidd" +
	"enhigh1hreflanghttp-equivideoptioniframenuitemidimagesizeslo" +
	"timagesrcsetimginsmallowfullscreenismapopovertargetactionite" +
	"mreferrerpolicyitemscopenitemtypematheaderspellcheckedmaxlen" +
	"gth2meterminlength3multiplemutedprefixpreloadingprofileprogr" +
	"essrclangstrikestrongstylesubdoptimumarqueesummarysuportalsv" +
	"gvocabbrwbrwidth4wraposter")

var _Hash_table = [1 << 9]Hash{
	0x3:   0xb304,  // step
	0x4:   0x2004,  // base
	0x5:   0xb607,  // pattern
	0x8:   0x8403,  // dir
	0xa:   0xe309,  // autofocus
	0xc:   0x3b04,  // main
	0xf:   0x2801,  // i
	0x10:  0x1,     // a
	0x12:  0x40004, // abbr
	0x13:  0x40705, // width
	0x15:  0x24506, // figure
	0x16:  0x23f06, // script
	0x17:  0x5e0f,  // amp-boilerplate
	0x18:  0x3d09,  // inputmode
	0x19:  0xb802,  // tt
	0x1c:  0x2d704, // href
	0x1d:  0x22305, // start
	0x21:  0x4807,  // rowspan
	0x23:  0x1e306, // coords
	0x25:  0xb104,  // list
	0x28:  0x3fa03, // svg
	0x29:  0x2d502, // h1
	0x2a:  0x15a05, // class
	0x2b:  0x2e805, // video
	0x2c:  0x3490e, // referrerpolicy
	0x2d:  0x2f608, // menuitem
	0x2e:  0x38805, // meter
	0x30:  0x17604, // code
	0x33:  0x2c408, // readonly
	0x35:  0x3c207, // srclang
	0x37:  0x3320d, // popovertarget
	0x39:  0x2db04, // lang
	0x3a:  0x3a403, // pre
	0x3d:  0x2f206, // iframe
	0x3e:  0x1b107, // content
	0x3f:  0x2fa06, // itemid
	0x40:  0x27f0a, // novalidate
	0x41:  0x1d306, // select
	0x43:  0x3c906, // strike
	0x44:  0x1a304, // cols
	0x46:  0x36b05, // thead
	0x48:  0x32103, // low
	0x4b:  0x1000b, // playsinline
	0x4d:  0x31206, // srcset
	0x51:  0x1c405, // label
	0x52:  0x3bb08, // progress
	0x53:  0x6702,  // rp
	0x54:  0x19903, // div
	0x55:  0xad08,  // datalist
	0x5b:  0x28d06, // dialog
	0x5c:  0x5208,  // tabindex
	0x5d:  0x40d04, // wrap
	0x61:  0x16e05, // frame
	0x64:  0x3000a, // imagesizes
	0x67:  0x6d07,  // address
	0x69:  0x3da03, // sub
	0x6d:  0x4b04,  // span
	0x6f:  0x16a03, // dfn
	0x70:  0xf309,  // translate
	0x71:  0x1f203, // del
	0x72:  0x705,   // tbody
	0x74:  0x15208, // property
	0x7b:  0x38d09, // minlength
	0x7d:  0x2cc06, // hidden
	0x7e:  0x18b03, // rev
	0x7f:  0xdb08,  // template
	0x81:  0x20b07, // details
	0x82:  0x8303,  // bdi
	0x86:  0x22507, // article
	0x88:  0x2ec06, // option
	0x89:  0x40902, // dt
	0x8b:  0x31b03, // ins
	0x8d:  0x18607, // picture
	0x8f:  0x18b08, // reversed
	0x92:  0x19b03, // var
	0x93:  0xad04,  // data
	0x95:  0x8e06,  // canvas
	0x96:  0x7b05,  // param
	0x97:  0x3eb07, // summary
	0x98:  0x15e07, // section
	0x9a:  0x2c09,  // accesskey
	0x9b:  0x26006, // action
	0x9c:  0x9402,  // id
	0x9e:  0x1701,  // s
	0x9f:  0x10b02, // rt
	0xa0:  0x2c304, // area
	0xa2:  0x3b407, // profile
	0xa5:  0x31203, // src
	0xa6:  0xea06,  // usemap
	0xa8:  0x1be06, // legend
	0xa9:  0x8604,  // ruby
	0xaf:  0x26a07, // enctype
	0xb0:  0x2a106, // height
	0xb1:  0x2340a, // figcaption
	0xb2:  0x3aa07, // preload
	0xb4:  0x10b03, // rtc
	0xb5:  0x40b02, // h4
	0xb6:  0xa106,  // object
	0xb8:  0x3fd05, // vocab
	0xb9:  0x19208, // disabled
	0xba:  0x16605, // embed
	0xbc:  0x9508,  // decoding
	0xc1:  0x2102,  // as
	0xc2:  0x14904, // nobr
	0xc4:  0x16c08, // noframes
	0xc5:  0x3507,  // acronym
	0xc6:  0x2930a, // formtarget
	0xc7:  0x35b05, // scope
	0xc8:  0x30504, // size
	0xcb:  0x3ad07, // loading
	0xcd:  0x17f03, // col
	0xd0:  0x2a804, // loop
	0xd1:  0x1307,  // charset
	0xd2:  0x1bb05, // table
	0xd5:  0x3a406, // prefix
	0xd6:  0x3de07, // optimum
	0xd8:  0x24f06, // center
	0xdb:  0xdc02,  // em
	0xdc:  0x2aa08, // optgroup
	0xde:  0x40403, // wbr
	0xe2:  0x3cf06, // strong
	0xe6:  0xbe05,  // value
	0xe9:  0x14b02, // br
	0xed:  0xee06,  // applet
	0xf0:  0x206,   // output
	0xf1:  0x22c08, // fieldset
	0xfb:  0x14406, // button
	0xfc:  0x30d0b, // imagesrcset
	0xfd:  0xc06,   // accept
	0x100: 0x31d05, // small
	0x102: 0x3f406, // portal
	0x103: 0x8a05,  // async
	0x104: 0x11208, // required
	0x105: 0x35d04, // open
	0x107: 0xaa04,  // kind
	0x108: 0x33213, // popovertargetaction
	0x109: 0x2a504, // html
	0x10b: 0x501,   // p
	0x10c: 0x7f04,  // mark
	0x10d: 0x32e05, // ismap
	0x10f: 0x1cc08, // controls
	0x110: 0xa605,  // track
	0x112: 0x38d03, // min
	0x113: 0x16407, // noembed
	0x116: 0x21f06, // inlist
	0x118: 0x1da09, // draggable
	0x119: 0x14e08, // itemprop
	0x11a: 0x1f02,  // rb
	0x11c: 0x17c02, // ul
	0x11e: 0xa602,  // tr
	0x11f: 0x27702, // th
	0x122: 0x29d02, // h5
	0x126: 0x1905,  // tfoot
	0x127: 0x37e03, // max
	0x129: 0x2d702, // hr
	0x12b: 0x1ff0e, // shadowrootmode
	0x12c: 0x29706, // target
	0x12f: 0x3f203, // sup
	0x134: 0x11d06, // typeof
	0x136: 0x18002, // ol
	0x137: 0x36c04, // head
	0x138: 0x7307,  // sandbox
	0x13a: 0x2b506, // hgroup
	0x13f: 0x5004,  // meta
	0x141: 0x5905,  // xmlns
	0x143: 0x38602, // h2
	0x144: 0xc0e,   // accept-charset
	0x146: 0x2bf04, // text
	0x147: 0x13a0a, // blockquote
	0x149: 0x1f102, // td
	0x14a: 0x37707, // checked
	0x14d: 0x2b104, // ping
	0x14e: 0x2f604, // menu
	0x150: 0x5d04,  // samp
	0x151: 0x2008,  // basefont
	0x152: 0x2710a, // formmethod
	0x155: 0xed03,  // map
	0x156: 0x27b0e, // formnovalidate
	0x159: 0x6e02,  // dd
	0x15c: 0xc30e,  // autocapitalize
	0x15d: 0x2660b, // formenctype
	0x15e: 0xbc03,  // nav
	0x161: 0x101,   // b
	0x163: 0x1a06,  // footer
	0x164: 0x24b06, // source
	0x166: 0x35709, // itemscope
	0x16a: 0x10d07, // capture
	0x16c: 0x36c06, // header
	0x16d: 0x1c804, // link
	0x171: 0x2160b, // crossorigin
	0x172: 0x4405,  // defer
	0x175: 0x2705,  // title
	0x177: 0x28b05, // media
	0x178: 0x11401, // q
	0x179: 0x21105, // shape
	0x17c: 0x25c03, // for
	0x17d: 0x30904, // slot
	0x17e: 0x7903,  // xmp
	0x184: 0x2404,  // font
	0x187: 0x13208, // blocking
	0x188: 0x8203,  // kbd
	0x18a: 0x1a908, // nomodule
	0x18b: 0x4e04,  // name
	0x18f: 0x29f02, // h6
	0x191: 0x31f05, // allow
	0x194: 0x39708, // multiple
	0x196: 0x30505, // sizes
	0x199: 0x23707, // caption
	0x19b: 0x34507, // itemref
	0x19c: 0x19e05, // color
	0x19f: 0x1220d, // fetchpriority
	0x1a7: 0xd10c,  // autocomplete
	0x1a8: 0x1a307, // colspan
	0x1aa: 0x16e08, // frameset
	0x1ab: 0x31f0f, // allowfullscreen
	0x1ac: 0x14d04, // cite
	0x1ae: 0x3ab03, // rel
	0x1b0: 0x39502, // h3
	0x1b1: 0x25c0a, // formaction
	0x1b3: 0x36904, // math
	0x1b4: 0x39f05, // muted
	0x1b5: 0x1e818, // shadowrootdelegatesfocus
	0x1b6: 0x24908, // resource
	0x1b9: 0x40102, // bb
	0x1ba: 0x2df0a, // http-equiv
	0x1be: 0x30005, // image
	0x1bf: 0x2bf08, // textarea
	0x1c1: 0x28904, // time
	0x1c2: 0x5,     // about
	0x1c3: 0x25c04, // form
	0x1c4: 0x301,   // u
	0x1c5: 0x41006, // poster
	0x1c8: 0x1d308, // selected
	0x1c9: 0x2d204, // high
	0x1ca: 0x3d505, // style
	0x1cc: 0x4804,  // rows
	0x1cd: 0x36c07, // headers
	0x1cf: 0x3720a, // spellcheck
	0x1d1: 0x11d04, // type
	0x1d3: 0xfc08,  // autoplay
	0x1d4: 0x28508, // datetime
	0x1d7: 0x9d05,  // audio
	0x1d9: 0xb202,  // is
	0x1de: 0x3dc03, // bdo
	0x1df: 0x3d05,  // input
	0x1e0: 0x31803, // img
	0x1e1: 0x11908, // datatype
	0x1e2: 0x36108, // itemtype
	0x1e3: 0x33207, // popover
	0x1e4: 0x2ba09, // plaintext
	0x1e6: 0x12f03, // big
	0x1e9: 0x2500c, // enterkeyhint
	0x1ea: 0x17807, // default
	0x1ec: 0x27506, // method
	0x1ed: 0x37e09, // maxlength
	0x1f0: 0x2d708, // hreflang
	0x1f1: 0x1c302, // dl
	0x1f2: 0xb102,  // li
	0x1f4: 0x17f08, // colgroup
	0x1f6: 0x1b10f, // contenteditable
	0x1f7: 0x3e407, // marquee
	0x1f9: 0x9205,  // aside
	0x1fa: 0x804,   // body
	0x1fb: 0x10805, // inert
	0x1fd: 0x23d08, // noscript
}

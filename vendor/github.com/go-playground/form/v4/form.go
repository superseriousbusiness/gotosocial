package form

import (
	"reflect"
	"time"
)

const (
	blank     = ""
	ignore    = "-"
	fieldNS   = "Field Namespace:"
	errorText = " ERROR:"
)

var (
	timeType = reflect.TypeOf(time.Time{})
)

// Mode specifies which mode the form decoder is to run
type Mode uint8

const (

	// ModeImplicit tries to parse values for all
	// fields that do not have an ignore '-' tag
	ModeImplicit Mode = iota

	// ModeExplicit only parses values for field with a field tag
	// and that tag is not the ignore '-' tag
	ModeExplicit
)

// AnonymousMode specifies how data should be rolled up
// or separated from anonymous structs
type AnonymousMode uint8

const (
	// AnonymousEmbed embeds anonymous data when encoding
	// eg. type A struct { Field string }
	//     type B struct { A, Field string }
	//     encode results: url.Values{"Field":[]string{"B FieldVal", "A FieldVal"}}
	AnonymousEmbed AnonymousMode = iota

	// AnonymousSeparate does not embed anonymous data when encoding
	// eg. type A struct { Field string }
	//     type B struct { A, Field string }
	//     encode results: url.Values{"Field":[]string{"B FieldVal"}, "A.Field":[]string{"A FieldVal"}}
	AnonymousSeparate
)

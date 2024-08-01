package iptc

import (
	"errors"
)

type StreamTagInfo struct {
	Description string
}

var (
	standardTags = map[StreamTagKey]StreamTagInfo{
		StreamTagKey{1, 120}: StreamTagInfo{"ARM Identifier"},

		StreamTagKey{1, 122}: StreamTagInfo{"ARM Version"},
		StreamTagKey{2, 0}:   StreamTagInfo{"Record Version"},
		StreamTagKey{2, 3}:   StreamTagInfo{"Object Type Reference"},
		StreamTagKey{2, 4}:   StreamTagInfo{"Object Attribute Reference"},
		StreamTagKey{2, 5}:   StreamTagInfo{"Object Name"},
		StreamTagKey{2, 7}:   StreamTagInfo{"Edit Status"},
		StreamTagKey{2, 8}:   StreamTagInfo{"Editorial Update"},
		StreamTagKey{2, 10}:  StreamTagInfo{"Urgency"},
		StreamTagKey{2, 12}:  StreamTagInfo{"Subject Reference"},
		StreamTagKey{2, 15}:  StreamTagInfo{"Category"},
		StreamTagKey{2, 20}:  StreamTagInfo{"Supplemental Category"},
		StreamTagKey{2, 22}:  StreamTagInfo{"Fixture Identifier"},
		StreamTagKey{2, 25}:  StreamTagInfo{"Keywords"},
		StreamTagKey{2, 26}:  StreamTagInfo{"Content Location Code"},
		StreamTagKey{2, 27}:  StreamTagInfo{"Content Location Name"},
		StreamTagKey{2, 30}:  StreamTagInfo{"Release Date"},
		StreamTagKey{2, 35}:  StreamTagInfo{"Release Time"},
		StreamTagKey{2, 37}:  StreamTagInfo{"Expiration Date"},
		StreamTagKey{2, 38}:  StreamTagInfo{"Expiration Time"},
		StreamTagKey{2, 40}:  StreamTagInfo{"Special Instructions"},
		StreamTagKey{2, 42}:  StreamTagInfo{"Action Advised"},
		StreamTagKey{2, 45}:  StreamTagInfo{"Reference Service"},
		StreamTagKey{2, 47}:  StreamTagInfo{"Reference Date"},
		StreamTagKey{2, 50}:  StreamTagInfo{"Reference Number"},
		StreamTagKey{2, 55}:  StreamTagInfo{"Date Created"},
		StreamTagKey{2, 60}:  StreamTagInfo{"Time Created"},
		StreamTagKey{2, 62}:  StreamTagInfo{"Digital Creation Date"},
		StreamTagKey{2, 63}:  StreamTagInfo{"Digital Creation Time"},
		StreamTagKey{2, 65}:  StreamTagInfo{"Originating Program"},
		StreamTagKey{2, 70}:  StreamTagInfo{"Program Version"},
		StreamTagKey{2, 75}:  StreamTagInfo{"Object Cycle"},
		StreamTagKey{2, 80}:  StreamTagInfo{"By-line"},
		StreamTagKey{2, 85}:  StreamTagInfo{"By-line Title"},
		StreamTagKey{2, 90}:  StreamTagInfo{"City"},
		StreamTagKey{2, 92}:  StreamTagInfo{"Sublocation"},
		StreamTagKey{2, 95}:  StreamTagInfo{"Province/State"},
		StreamTagKey{2, 100}: StreamTagInfo{"Country/Primary Location Code"},
		StreamTagKey{2, 101}: StreamTagInfo{"Country/Primary Location Name"},
		StreamTagKey{2, 103}: StreamTagInfo{"Original Transmission Reference"},
		StreamTagKey{2, 105}: StreamTagInfo{"Headline"},
		StreamTagKey{2, 110}: StreamTagInfo{"Credit"},
		StreamTagKey{2, 115}: StreamTagInfo{"Source"},
		StreamTagKey{2, 116}: StreamTagInfo{"Copyright Notice"},
		StreamTagKey{2, 118}: StreamTagInfo{"Contact"},
		StreamTagKey{2, 120}: StreamTagInfo{"Caption/Abstract"},
		StreamTagKey{2, 122}: StreamTagInfo{"Writer/Editor"},
		StreamTagKey{2, 125}: StreamTagInfo{"Rasterized Caption"},
		StreamTagKey{2, 130}: StreamTagInfo{"Image Type"},
		StreamTagKey{2, 131}: StreamTagInfo{"Image Orientation"},
		StreamTagKey{2, 135}: StreamTagInfo{"Language Identifier"},
		StreamTagKey{2, 150}: StreamTagInfo{"Audio Type"},
		StreamTagKey{2, 151}: StreamTagInfo{"Audio Sampling Rate"},
		StreamTagKey{2, 152}: StreamTagInfo{"Audio Sampling Resolution"},
		StreamTagKey{2, 153}: StreamTagInfo{"Audio Duration"},
		StreamTagKey{2, 154}: StreamTagInfo{"Audio Outcue"},
		StreamTagKey{2, 200}: StreamTagInfo{"ObjectData Preview File Format"},
		StreamTagKey{2, 201}: StreamTagInfo{"ObjectData Preview File Format Version"},
		StreamTagKey{2, 202}: StreamTagInfo{"ObjectData Preview Data"},
		StreamTagKey{7, 10}:  StreamTagInfo{"Size Mode"},
		StreamTagKey{7, 20}:  StreamTagInfo{"Max Subfile Size"},
		StreamTagKey{7, 90}:  StreamTagInfo{"ObjectData Size Announced"},
		StreamTagKey{7, 95}:  StreamTagInfo{"Maximum ObjectData Size"},
		StreamTagKey{8, 10}:  StreamTagInfo{"Subfile"},
		StreamTagKey{9, 10}:  StreamTagInfo{"Confirmed ObjectData Size"},
	}
)

var (
	// ErrTagNotStandard indicates that the given tag is not known among the
	// documented standard set.
	ErrTagNotStandard = errors.New("not a standard tag")
)

// GetTagInfo return the info for the given tag. Returns ErrTagNotStandard if
// not known.
func GetTagInfo(recordNumber, datasetNumber int) (sti StreamTagInfo, err error) {
	stk := StreamTagKey{uint8(recordNumber), uint8(datasetNumber)}

	sti, found := standardTags[stk]
	if found == false {
		return sti, ErrTagNotStandard
	}

	return sti, nil
}

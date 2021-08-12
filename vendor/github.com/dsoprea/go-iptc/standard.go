package iptc

import (
	"errors"
)

// StreamTagInfo encapsulates the properties of each tag.
type StreamTagInfo struct {
	// Description is the human-readable description of the tag.
	Description string
}

var (
	standardTags = map[StreamTagKey]StreamTagInfo{
		{1, 120}: {"ARM Identifier"},

		{1, 122}: {"ARM Version"},
		{2, 0}:   {"Record Version"},
		{2, 3}:   {"Object Type Reference"},
		{2, 4}:   {"Object Attribute Reference"},
		{2, 5}:   {"Object Name"},
		{2, 7}:   {"Edit Status"},
		{2, 8}:   {"Editorial Update"},
		{2, 10}:  {"Urgency"},
		{2, 12}:  {"Subject Reference"},
		{2, 15}:  {"Category"},
		{2, 20}:  {"Supplemental Category"},
		{2, 22}:  {"Fixture Identifier"},
		{2, 25}:  {"Keywords"},
		{2, 26}:  {"Content Location Code"},
		{2, 27}:  {"Content Location Name"},
		{2, 30}:  {"Release Date"},
		{2, 35}:  {"Release Time"},
		{2, 37}:  {"Expiration Date"},
		{2, 38}:  {"Expiration Time"},
		{2, 40}:  {"Special Instructions"},
		{2, 42}:  {"Action Advised"},
		{2, 45}:  {"Reference Service"},
		{2, 47}:  {"Reference Date"},
		{2, 50}:  {"Reference Number"},
		{2, 55}:  {"Date Created"},
		{2, 60}:  {"Time Created"},
		{2, 62}:  {"Digital Creation Date"},
		{2, 63}:  {"Digital Creation Time"},
		{2, 65}:  {"Originating Program"},
		{2, 70}:  {"Program Version"},
		{2, 75}:  {"Object Cycle"},
		{2, 80}:  {"By-line"},
		{2, 85}:  {"By-line Title"},
		{2, 90}:  {"City"},
		{2, 92}:  {"Sublocation"},
		{2, 95}:  {"Province/State"},
		{2, 100}: {"Country/Primary Location Code"},
		{2, 101}: {"Country/Primary Location Name"},
		{2, 103}: {"Original Transmission Reference"},
		{2, 105}: {"Headline"},
		{2, 110}: {"Credit"},
		{2, 115}: {"Source"},
		{2, 116}: {"Copyright Notice"},
		{2, 118}: {"Contact"},
		{2, 120}: {"Caption/Abstract"},
		{2, 122}: {"Writer/Editor"},
		{2, 125}: {"Rasterized Caption"},
		{2, 130}: {"Image Type"},
		{2, 131}: {"Image Orientation"},
		{2, 135}: {"Language Identifier"},
		{2, 150}: {"Audio Type"},
		{2, 151}: {"Audio Sampling Rate"},
		{2, 152}: {"Audio Sampling Resolution"},
		{2, 153}: {"Audio Duration"},
		{2, 154}: {"Audio Outcue"},
		{2, 200}: {"ObjectData Preview File Format"},
		{2, 201}: {"ObjectData Preview File Format Version"},
		{2, 202}: {"ObjectData Preview Data"},
		{7, 10}:  {"Size Mode"},
		{7, 20}:  {"Max Subfile Size"},
		{7, 90}:  {"ObjectData Size Announced"},
		{7, 95}:  {"Maximum ObjectData Size"},
		{8, 10}:  {"Subfile"},
		{9, 10}:  {"Confirmed ObjectData Size"},
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

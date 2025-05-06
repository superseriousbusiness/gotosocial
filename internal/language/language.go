// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package language

import (
	"errors"

	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"golang.org/x/text/language"
	"golang.org/x/text/language/display"
)

var namer display.Namer

// InitLangs parses languages from the
// given slice of tags, and sets the `namer`
// display.Namer for the instance.
//
// This function should only be called once,
// since setting the namer is not thread safe.
func InitLangs(tagStrs []string) (Languages, error) {
	var (
		languages = make(Languages, len(tagStrs))
		tags      = make([]language.Tag, len(tagStrs))
	)

	// Reset namer.
	namer = nil

	// Parse all tags first.
	for i, tagStr := range tagStrs {
		tag, err := language.Parse(tagStr)
		if err != nil {
			return nil, gtserror.Newf(
				"error parsing %s as BCP47 language tag: %w",
				tagStr, err,
			)
		}
		tags[i] = tag
	}

	// Check if we can set a namer.
	if len(tags) != 0 {
		namer = display.Languages(tags[0])
	}

	// Fall namer back to English.
	if namer == nil {
		namer = display.Languages(language.English)
	}

	// Parse nice language models from tags
	// (this will use the namer we just set).
	for i, tag := range tags {
		languages[i] = ParseTag(tag)
	}

	return languages, nil
}

// Language models a BCP47 language tag
// along with helper strings for the tag.
type Language struct {
	// BCP47 language tag
	Tag language.Tag
	// Normalized string
	// of BCP47 tag.
	TagStr string
	// Human-readable
	// language name(s).
	DisplayStr string
}

// MarshalText implements encoding.TextMarshaler{}.
func (l *Language) MarshalText() ([]byte, error) {
	return []byte(l.TagStr), nil
}

// UnmarshalText implements encoding.TextUnmarshaler{}.
func (l *Language) UnmarshalText(text []byte) error {
	lang, err := Parse(string(text))
	if err != nil {
		return err
	}
	*l = *lang
	return nil
}

type Languages []*Language

func (l *Languages) Set(in string) error {
	if l == nil {
		return errors.New("nil receiver")
	}
	prefix, err := Parse(in)
	if err != nil {
		return err
	}
	(*l) = append((*l), prefix)
	return nil
}

func (l *Languages) Strings() []string {
	if l == nil || len(*l) == 0 {
		return nil
	}
	strs := make([]string, len(*l))
	for i, lang := range *l {
		strs[i] = lang.TagStr
	}
	return strs
}

func (l Languages) Tags() []language.Tag {
	tags := make([]language.Tag, len(l))
	for i, lang := range l {
		tags[i] = lang.Tag
	}

	return tags
}

func (l Languages) TagStrs() []string {
	tagStrs := make([]string, len(l))
	for i, lang := range l {
		tagStrs[i] = lang.TagStr
	}

	return tagStrs
}

func (l Languages) DisplayStrs() []string {
	displayStrs := make([]string, len(l))
	for i, lang := range l {
		displayStrs[i] = lang.DisplayStr
	}

	return displayStrs
}

// ParseTag parses and nicely formats the input language BCP47 tag,
// returning a Language with ready-to-use display and tag strings.
func ParseTag(tag language.Tag) *Language {
	l := new(Language)
	l.Tag = tag
	l.TagStr = tag.String()

	var (
		// Our name for the language.
		name string
		// Language's name for itself.
		selfName = display.Self.Name(tag)
	)

	// Try to use namer
	// (if initialized).
	if namer != nil {
		name = namer.Name(tag)
	}

	switch {
	case name == "":
		// We don't have a name for
		// this language, just use
		// its own name for itself.
		l.DisplayStr = selfName

	case name == selfName:
		// Avoid repeating ourselves:
		// showing "English (English)"
		// is not useful.
		l.DisplayStr = name

	default:
		// Include our name for the
		// language, and its own
		// name for itself.
		l.DisplayStr = name + " " + "(" + selfName + ")"
	}

	return l
}

// Parse parses and nicely formats the input language BCP47 tag,
// returning a Language with ready-to-use display and tag strings.
func Parse(lang string) (*Language, error) {
	tag, err := language.Parse(lang)
	if err != nil {
		return nil, err
	}

	return ParseTag(tag), nil
}

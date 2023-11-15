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

package langs

import (
	"slices"

	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"golang.org/x/text/language"
	"golang.org/x/text/language/display"
)

var (
	instanceLangTags []language.Tag
	instanceLangStrs []string
	namer            display.Namer
)

// InstanceLangTags returns a copy of the global
// parsed language tags preferred by this
// GoToSocial instance.
func InstanceLangTags() []language.Tag {
	return slices.Clone(instanceLangTags)
}

// InstanceLangStrings returns a copy of the global
// parsed language tags preferred by this
// GoToSocial instance, as nicely formatted strings.
func InstanceLangStrings() []string {
	return slices.Clone(instanceLangStrs)
}

// DisplayLang parses and nicely formats the input language BCP47 tag.
//
// First return value: the normalized string version of the tag.
//
// Second return value: a human-readable name for that language, in the format:
// `[language name in the instance language] ([language name in the tag language])`
//
// Returns empty strings if the tag could not be parsed.
func DisplayLang(lang string) (string, string) {
	if lang == "" {
		// Unknown.
		return "", ""
	}

	tag, err := language.Parse(lang)
	if err != nil {
		// Unknown.
		return "", ""
	}

	// Create nice human-readable
	// language name.
	var nice string
	name := namer.Name(tag)
	selfName := display.Self.Name(tag)

	// Avoid repeating ourselves.
	if name == selfName {
		nice = name
	} else {
		nice = name + " " + "(" + selfName + ")"
	}

	return tag.String(), nice
}

// InitLangs sets global language variables in this
// package using the given BCP47 language tags.
func InitLangs(tagStrs []string) error {
	instanceLangTags = make([]language.Tag, len(tagStrs))
	instanceLangStrs = make([]string, len(tagStrs))

	for i, tagStr := range tagStrs {
		tag, err := language.Parse(tagStr)
		if err != nil {
			return gtserror.Newf(
				"error parsing %s as BCP47 language tag: %w",
				tagStr, err,
			)
		}

		instanceLangTags[i] = tag

		// Try to set a nice version of
		// this tag as lang string, fall
		// back to normalized tag.
		langStr := display.Self.Name(tag)
		if langStr != "" {
			instanceLangStrs[i] = langStr
		} else {
			instanceLangStrs[i] = tag.String()
		}
	}

	// Pick a display language for naming
	// other languages. Prefer primary instance
	// lang if it's set, fall back to English
	// if it's not set, or if we can't create
	// a namer using the primary instance language.
	var displayLang language.Tag
	if len(instanceLangTags) != 0 {
		displayLang = instanceLangTags[0]
	} else {
		displayLang = language.English
	}

	namer = display.Languages(displayLang)
	if namer == nil {
		namer = display.Languages(language.English)
	}

	return nil
}

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

package language_test

import (
	"slices"
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/language"
	golanguage "golang.org/x/text/language"
)

func TestInstanceLangs(t *testing.T) {
	for i, test := range []struct {
		InstanceLangs       []string
		expectedLangs       []golanguage.Tag
		expectedLangStrs    []string
		expectedErr         error
		parseDisplayLang    string
		expectedDisplayLang string
	}{
		{
			InstanceLangs: []string{"en-us", "fr"},
			expectedLangs: []golanguage.Tag{
				golanguage.AmericanEnglish,
				golanguage.French,
			},
			expectedLangStrs: []string{
				"American English",
				"French (français)",
			},
			parseDisplayLang:    "de",
			expectedDisplayLang: "German (Deutsch)",
		},
		{
			InstanceLangs: []string{"fr", "en-us"},
			expectedLangs: []golanguage.Tag{
				golanguage.French,
				golanguage.AmericanEnglish,
			},
			expectedLangStrs: []string{
				"français",
				"anglais américain (American English)",
			},
			parseDisplayLang:    "de",
			expectedDisplayLang: "allemand (Deutsch)",
		},
		{
			InstanceLangs:       []string{},
			expectedLangs:       []golanguage.Tag{},
			expectedLangStrs:    []string{},
			parseDisplayLang:    "de",
			expectedDisplayLang: "German (Deutsch)",
		},
		{
			InstanceLangs: []string{"zh"},
			expectedLangs: []golanguage.Tag{
				golanguage.Chinese,
			},
			expectedLangStrs: []string{
				"中文",
			},
			parseDisplayLang:    "de",
			expectedDisplayLang: "德语 (Deutsch)",
		},
		{
			InstanceLangs: []string{"ar", "en"},
			expectedLangs: []golanguage.Tag{
				golanguage.Arabic,
				golanguage.English,
			},
			expectedLangStrs: []string{
				"العربية",
				"الإنجليزية (English)",
			},
			parseDisplayLang:    "fi",
			expectedDisplayLang: "الفنلندية (suomi)",
		},
		{
			InstanceLangs: []string{"en-us"},
			expectedLangs: []golanguage.Tag{
				golanguage.AmericanEnglish,
			},
			expectedLangStrs: []string{
				"American English",
			},
			parseDisplayLang:    "en-us",
			expectedDisplayLang: "American English",
		},
		{
			InstanceLangs: []string{"en-us"},
			expectedLangs: []golanguage.Tag{
				golanguage.AmericanEnglish,
			},
			expectedLangStrs: []string{
				"American English",
			},
			parseDisplayLang:    "en-gb",
			expectedDisplayLang: "British English",
		},
	} {
		languages, err := language.InitLangs(test.InstanceLangs)
		if err != test.expectedErr {
			t.Errorf("test %d expected error %v, got %v", i, test.expectedErr, err)
		}

		parsedTags := languages.Tags()
		if !slices.Equal(test.expectedLangs, parsedTags) {
			t.Errorf("test %d expected language tags %v, got %v", i, test.expectedLangs, parsedTags)
		}

		parsedLangStrs := languages.DisplayStrs()
		if !slices.Equal(test.expectedLangStrs, parsedLangStrs) {
			t.Errorf("test %d expected language strings %v, got %v", i, test.expectedLangStrs, parsedLangStrs)
		}

		parsedLang, err := language.Parse(test.parseDisplayLang)
		if err != nil {
			t.Errorf("unexpected error %v", err)
			return
		}

		if test.expectedDisplayLang != parsedLang.DisplayStr {
			t.Errorf("test %d expected to parse language %v, got %v", i, test.expectedDisplayLang, parsedLang.DisplayStr)
		}
	}
}

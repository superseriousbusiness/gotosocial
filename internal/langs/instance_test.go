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

package langs_test

import (
	"slices"
	"testing"

	"github.com/superseriousbusiness/gotosocial/internal/langs"
	"golang.org/x/text/language"
)

func TestInstanceLangs(t *testing.T) {
	for i, test := range []struct {
		InstanceLangs       []string
		expectedLangs       []language.Tag
		expectedLangStrs    []string
		expectedErr         error
		parseDisplayLang    string
		expectedDisplayLang string
	}{
		{
			InstanceLangs: []string{"en-us", "fr"},
			expectedLangs: []language.Tag{
				language.AmericanEnglish,
				language.French,
			},
			expectedLangStrs: []string{
				"American English",
				"français",
			},
			parseDisplayLang:    "de",
			expectedDisplayLang: "German (Deutsch)",
		},
		{
			InstanceLangs: []string{"fr", "en-us"},
			expectedLangs: []language.Tag{
				language.French,
				language.AmericanEnglish,
			},
			expectedLangStrs: []string{
				"français",
				"American English",
			},
			parseDisplayLang:    "de",
			expectedDisplayLang: "allemand (Deutsch)",
		},
		{
			InstanceLangs:       []string{},
			expectedLangs:       []language.Tag{},
			expectedLangStrs:    []string{},
			parseDisplayLang:    "de",
			expectedDisplayLang: "German (Deutsch)",
		},
		{
			InstanceLangs: []string{"zh"},
			expectedLangs: []language.Tag{
				language.Chinese,
			},
			expectedLangStrs: []string{
				"中文",
			},
			parseDisplayLang:    "de",
			expectedDisplayLang: "德语 (Deutsch)",
		},
		{
			InstanceLangs: []string{"ar", "en"},
			expectedLangs: []language.Tag{
				language.Arabic,
				language.English,
			},
			expectedLangStrs: []string{
				"العربية",
				"English",
			},
			parseDisplayLang:    "fi",
			expectedDisplayLang: "الفنلندية (suomi)",
		},
		{
			InstanceLangs: []string{"en-us"},
			expectedLangs: []language.Tag{
				language.AmericanEnglish,
			},
			expectedLangStrs: []string{
				"American English",
			},
			parseDisplayLang:    "en-us",
			expectedDisplayLang: "American English",
		},
		{
			InstanceLangs: []string{"en-us"},
			expectedLangs: []language.Tag{
				language.AmericanEnglish,
			},
			expectedLangStrs: []string{
				"American English",
			},
			parseDisplayLang:    "en-gb",
			expectedDisplayLang: "British English",
		},
	} {
		if err := langs.InitLangs(test.InstanceLangs); err != test.expectedErr {
			t.Errorf("test %d expected error %v, got %v", i, test.expectedErr, err)
		}

		parsedLangs := langs.InstanceLangTags()
		if !slices.Equal(test.expectedLangs, parsedLangs) {
			t.Errorf("test %d expected language tags %v, got %v", i, test.expectedLangs, parsedLangs)
		}

		parsedLangStrs := langs.InstanceLangStrings()
		if !slices.Equal(test.expectedLangStrs, parsedLangStrs) {
			t.Errorf("test %d expected language strings %v, got %v", i, test.expectedLangStrs, parsedLangStrs)
		}

		_, displayLang := langs.DisplayLang(test.parseDisplayLang)
		if test.expectedDisplayLang != displayLang {
			t.Errorf("test %d expected to parse language %v, got %v", i, test.expectedDisplayLang, displayLang)
		}
	}
}

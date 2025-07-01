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

package typeutils

import (
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/language"
)

func TestMisskeyReportContentURLs1(t *testing.T) {
	content := `Note: https://bad.instance/@tobi/statuses/01GPB56GPJ37JTK9HW308HQKBQ
Note: https://bad.instance/@tobi/statuses/01GPB56GPJ37JTK9HW308HQKBQ
Note: https://bad.instance/@tobi/statuses/01GPB56GPJ37JTK9HW308HQKBQ
-----
Test report from Calckey`

	urls := misskeyReportInlineURLs(content)
	if l := len(urls); l != 3 {
		t.Fatalf("wanted 3 urls, got %d", l)
	}
}

func TestMisskeyReportContentURLs2(t *testing.T) {
	content := `This is a report
with just a normal url in it: https://example.org, and is not
misskey-formatted`

	urls := misskeyReportInlineURLs(content)
	if l := len(urls); l != 0 {
		t.Fatalf("wanted 0 urls, got %d", l)
	}
}

func TestContentToContentLanguage(t *testing.T) {
	type testcase struct {
		content           gtsmodel.Content
		instanceLanguages language.Languages
		expectedContent   string
		expectedLang      string
	}

	ctx := t.Context()

	for i, testcase := range []testcase{
		{
			content: gtsmodel.Content{
				Content:    "hello world",
				ContentMap: nil,
			},
			expectedContent: "hello world",
			expectedLang:    "",
		},
		{
			content: gtsmodel.Content{
				Content: "",
				ContentMap: map[string]string{
					"en": "hello world",
				},
			},
			expectedContent: "hello world",
			expectedLang:    "en",
		},
		{
			content: gtsmodel.Content{
				Content: "bonjour le monde",
				ContentMap: map[string]string{
					"en": "hello world",
					"fr": "bonjour le monde",
				},
			},
			expectedContent: "bonjour le monde",
			expectedLang:    "fr",
		},
		{
			content: gtsmodel.Content{
				Content: "bonjour le monde",
				ContentMap: map[string]string{
					"en": "hello world",
				},
			},
			expectedContent: "bonjour le monde",
			expectedLang:    "",
		},
		{
			content: gtsmodel.Content{
				Content: "",
				ContentMap: map[string]string{
					"en": "hello world",
					"ru": "Привет, мир!",
					"nl": "hallo wereld!",
					"ca": "Hola món!",
				},
			},
			instanceLanguages: language.Languages{
				{TagStr: "en"},
				{TagStr: "ca"},
			},
			expectedContent: "hello world",
			expectedLang:    "en",
		},
		{
			content: gtsmodel.Content{
				Content: "",
				ContentMap: map[string]string{
					"en": "hello world",
					"ru": "Привет, мир!",
					"nl": "hallo wereld!",
					"ca": "Hola món!",
				},
			},
			instanceLanguages: language.Languages{
				{TagStr: "ca"},
				{TagStr: "en"},
			},
			expectedContent: "Hola món!",
			expectedLang:    "ca",
		},
	} {
		langs, err := language.InitLangs(testcase.instanceLanguages.TagStrs())
		if err != nil {
			t.Fatal(err)
		}
		config.SetInstanceLanguages(langs)

		content, language := ContentToContentLanguage(ctx, testcase.content)
		if content != testcase.expectedContent {
			t.Errorf(
				"test %d expected content '%s' got '%s'",
				i, testcase.expectedContent, content,
			)
		}

		if language != testcase.expectedLang {
			t.Errorf(
				"test %d expected language '%s' got '%s'",
				i, testcase.expectedLang, language,
			)
		}
	}
}

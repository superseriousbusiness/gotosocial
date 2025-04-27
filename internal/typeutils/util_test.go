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
	"context"
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/language"
	"github.com/stretchr/testify/assert"
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

	ctx, cncl := context.WithCancel(context.Background())
	defer cncl()

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

func TestFilterableText(t *testing.T) {
	type testcase struct {
		status         *gtsmodel.Status
		expectedFields []string
	}

	for _, testcase := range []testcase{
		{
			status: &gtsmodel.Status{
				ContentWarning: "This is a test status",
				Content:        `<p>Import / export of account data via CSV files will be coming in 0.17.0 :) No more having to run scripts + CLI tools to import a list of accounts you follow, after doing a migration to a <a href="https://gts.superseriousbusiness.org/tags/gotosocial" class="mention hashtag" rel="tag nofollow noreferrer noopener" target="_blank">#<span>GoToSocial</span></a> instance.</p>`,
			},
			expectedFields: []string{
				"This is a test status",
				"Import / export of account data via CSV files will be coming in 0.17.0 :) No more having to run scripts + CLI tools to import a list of accounts you follow, after doing a migration to a #GoToSocial <https://gts.superseriousbusiness.org/tags/gotosocial> instance.",
			},
		},
		{
			status: &gtsmodel.Status{
				Content: `<p><span class="h-card"><a href="https://example.org/@zlatko" class="u-url mention" rel="nofollow noreferrer noopener" target="_blank">@<span>zlatko</span></a></span> currently we used modernc/sqlite3 for our sqlite driver, but we've been experimenting with wasm sqlite, and will likely move to that permanently in future; in the meantime, both options are available (the latter with a build tag)</p><p><a href="https://codeberg.org/superseriousbusiness/gotosocial/pulls/2863" rel="nofollow noreferrer noopener" target="_blank">https://codeberg.org/superseriousbusiness/gotosocial/pulls/2863</a></p>`,
			},
			expectedFields: []string{
				"@zlatko <https://example.org/@zlatko> currently we used modernc/sqlite3 for our sqlite driver, but we've been experimenting with wasm sqlite, and will likely move to that permanently in future; in the meantime, both options are available (the latter with a build tag)\n\nhttps://codeberg.org/superseriousbusiness/gotosocial/pulls/2863 <https://codeberg.org/superseriousbusiness/gotosocial/pulls/2863>",
			},
		},
		{
			status: &gtsmodel.Status{
				ContentWarning: "Nerd stuff",
				Content:        `<p>Latest graphs for <a href="https://gts.superseriousbusiness.org/tags/gotosocial" class="mention hashtag" rel="tag nofollow noreferrer noopener" target="_blank">#<span>GoToSocial</span></a> on <a href="https://github.com/ncruces/go-sqlite3" rel="nofollow noreferrer noopener" target="_blank">Wasm sqlite3</a> with <a href="https://codeberg.org/gruf/go-ffmpreg" rel="nofollow noreferrer noopener" target="_blank">embedded Wasm ffmpeg</a>, both running on <a href="https://wazero.io/" rel="nofollow noreferrer noopener" target="_blank">Wazero</a>, and configured with a <a href="https://codeberg.org/superseriousbusiness/gotosocial/src/commit/20fe430ef9ff3012a7a4dc2d01b68020c20e13bb/example/config.yaml#L259-L266" rel="nofollow noreferrer noopener" target="_blank">50MiB db cache target</a>. This is the version we'll be releasing soonish, now we're happy with how we've tamed everything.</p>`,
				Attachments: []*gtsmodel.MediaAttachment{
					{
						Description: `Graph showing GtS using between 150-300 MiB of memory, steadily, over a few days.`,
					},
					{
						Description: `Another media attachment`,
					},
				},
				Poll: &gtsmodel.Poll{
					Options: []string{
						"Poll option 1",
						"Poll option 2",
					},
				},
			},
			expectedFields: []string{
				"Nerd stuff",
				"Latest graphs for #GoToSocial <https://gts.superseriousbusiness.org/tags/gotosocial> on Wasm sqlite3 <https://github.com/ncruces/go-sqlite3> with embedded Wasm ffmpeg <https://codeberg.org/gruf/go-ffmpreg>, both running on Wazero <https://wazero.io/>, and configured with a 50MiB db cache target <https://codeberg.org/superseriousbusiness/gotosocial/src/commit/20fe430ef9ff3012a7a4dc2d01b68020c20e13bb/example/config.yaml#L259-L266>. This is the version we'll be releasing soonish, now we're happy with how we've tamed everything.",
				"Graph showing GtS using between 150-300 MiB of memory, steadily, over a few days.",
				"Another media attachment",
				"Poll option 1",
				"Poll option 2",
			},
		},
	} {
		fields := filterableFields(testcase.status)
		assert.Equal(t, testcase.expectedFields, fields)
	}
}

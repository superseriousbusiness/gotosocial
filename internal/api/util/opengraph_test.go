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

package util

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
)

type OpenGraphTestSuite struct {
	suite.Suite
}

func (suite *OpenGraphTestSuite) TestParseDescription() {
	tests := []struct {
		name, in, exp string
	}{
		{name: "shellcmd", in: `echo '\e]8;;http://example.com\e\This is a link\e]8;;\e'`, exp: `echo &#39;&bsol;e]8;;http://example.com&bsol;e&bsol;This is a link&bsol;e]8;;&bsol;e&#39;`},
		{name: "newlines", in: "test\n\ntest\ntest", exp: "test test test"},
	}

	for _, tt := range tests {
		tt := tt
		suite.Run(tt.name, func() {
			suite.Equal(fmt.Sprintf("content=\"%s\"", tt.exp), ParseDescription(tt.in))
		})
	}
}

func (suite *OpenGraphTestSuite) TestWithAccountWithNote() {
	baseMeta := OGBase(&apimodel.InstanceV1{
		AccountDomain: "example.org",
		Languages:     []string{"en"},
	})

	acct := &apimodel.Account{
		Acct:        "example_account",
		DisplayName: "example person!!",
		URL:         "https://example.org/@example_account",
		Note:        "<p>This is my profile, read it and weep! Weep then!</p>",
		Username:    "example_account",
	}

	accountMeta := baseMeta.WithAccount(&apimodel.WebAccount{Account: acct})

	suite.EqualValues(OGMeta{
		Title:                "example person!!, @example_account@example.org",
		Type:                 "profile",
		Locale:               "en",
		URL:                  "https://example.org/@example_account",
		SiteName:             "example.org",
		Description:          "content=\"This is my profile, read it and weep! Weep then!\"",
		Image:                "",
		ImageWidth:           "",
		ImageHeight:          "",
		ImageAlt:             "Avatar for example_account",
		ArticlePublisher:     "",
		ArticleAuthor:        "",
		ArticleModifiedTime:  "",
		ArticlePublishedTime: "",
		ProfileUsername:      "example_account",
	}, *accountMeta)
}

func (suite *OpenGraphTestSuite) TestWithAccountNoNote() {
	baseMeta := OGBase(&apimodel.InstanceV1{
		AccountDomain: "example.org",
		Languages:     []string{"en"},
	})

	acct := &apimodel.Account{
		Acct:        "example_account",
		DisplayName: "example person!!",
		URL:         "https://example.org/@example_account",
		Note:        "", // <- empty
		Username:    "example_account",
	}

	accountMeta := baseMeta.WithAccount(&apimodel.WebAccount{Account: acct})

	suite.EqualValues(OGMeta{
		Title:                "example person!!, @example_account@example.org",
		Type:                 "profile",
		Locale:               "en",
		URL:                  "https://example.org/@example_account",
		SiteName:             "example.org",
		Description:          "content=\"This GoToSocial user hasn't written a bio yet!\"",
		Image:                "",
		ImageWidth:           "",
		ImageHeight:          "",
		ImageAlt:             "Avatar for example_account",
		ArticlePublisher:     "",
		ArticleAuthor:        "",
		ArticleModifiedTime:  "",
		ArticlePublishedTime: "",
		ProfileUsername:      "example_account",
	}, *accountMeta)
}

func TestOpenGraphTestSuite(t *testing.T) {
	suite.Run(t, &OpenGraphTestSuite{})
}

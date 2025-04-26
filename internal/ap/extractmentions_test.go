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

package ap_test

import (
	"testing"

	"code.superseriousbusiness.org/activity/streams"
	"code.superseriousbusiness.org/gotosocial/internal/ap"
	"code.superseriousbusiness.org/gotosocial/testrig"
	"github.com/stretchr/testify/suite"
)

type ExtractMentionsTestSuite struct {
	APTestSuite
}

func (suite *ExtractMentionsTestSuite) TestExtractMentionsFromNote() {
	note := suite.noteWithMentions1

	mentions, err := ap.ExtractMentions(note)
	suite.NoError(err)
	suite.Len(mentions, 2)

	m1 := mentions[0]
	suite.Equal("@dumpsterqueer@superseriousbusiness.org", m1.NameString)
	suite.Equal("https://gts.superseriousbusiness.org/users/dumpsterqueer", m1.TargetAccountURI)

	m2 := mentions[1]
	suite.Equal("@f0x@superseriousbusiness.org", m2.NameString)
	suite.Equal("https://gts.superseriousbusiness.org/users/f0x", m2.TargetAccountURI)
}

func (suite *ExtractMentionsTestSuite) TestExtractMentions() {
	newMention := func(nameString string, href string) ap.Mentionable {
		mention := streams.NewActivityStreamsMention()

		if nameString != "" {
			nameProp := streams.NewActivityStreamsNameProperty()
			nameProp.AppendXMLSchemaString(nameString)
			mention.SetActivityStreamsName(nameProp)
		}

		if href != "" {
			hrefProp := streams.NewActivityStreamsHrefProperty()
			hrefProp.SetIRI(testrig.URLMustParse(href))
			mention.SetActivityStreamsHref(hrefProp)
		}

		return mention
	}

	type test struct {
		nameString         string
		href               string
		expectedNameString string
		expectedHref       string
		expectedErr        string
	}

	for i, t := range []test{
		{
			// Mention with both Name and Href set, should be fine.
			nameString:         "@someone@example.org",
			href:               "https://example.org/@someone",
			expectedNameString: "@someone@example.org",
			expectedHref:       "https://example.org/@someone",
			expectedErr:        "",
		},
		{
			// Mention with just Href set, should be fine.
			nameString:         "",
			href:               "https://example.org/@someone",
			expectedNameString: "",
			expectedHref:       "https://example.org/@someone",
			expectedErr:        "",
		},
		{
			// Mention with just Name set, should be fine.
			nameString:         "@someone@example.org",
			href:               "",
			expectedNameString: "@someone@example.org",
			expectedHref:       "",
			expectedErr:        "",
		},
		{
			// Mention with nothing set, not fine!
			nameString:         "",
			href:               "",
			expectedNameString: "",
			expectedHref:       "",
			expectedErr:        "ExtractMention: neither Name nor Href were set",
		},
	} {
		apMention := newMention(t.nameString, t.href)
		mention, err := ap.ExtractMention(apMention)

		if err != nil {
			if errString := err.Error(); errString != t.expectedErr {
				suite.Fail("",
					"test %d expected error %s, got %s",
					i+1, t.expectedErr, errString,
				)
			}
			continue
		} else if t.expectedErr != "" {
			suite.Fail("",
				"test %d expected error %s, got no error",
				i+1, t.expectedErr,
			)
		}

		if mention.NameString != t.expectedNameString {
			suite.Fail("",
				"test %d expected nameString %s, got %s",
				i+1, t.expectedNameString, mention.NameString,
			)
		}

		if mention.TargetAccountURI != t.expectedHref {
			suite.Fail("",
				"test %d expected href %s, got %s",
				i+1, t.expectedHref, mention.TargetAccountURI,
			)
		}
	}
}

func TestExtractMentionsTestSuite(t *testing.T) {
	suite.Run(t, &ExtractMentionsTestSuite{})
}

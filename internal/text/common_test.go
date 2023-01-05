/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package text_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

const (
	replaceMentionsString                 = "Another test @foss_satan@fossbros-anonymous.io\n\n#Hashtag\n\nText"
	replaceMentionsExpected               = "Another test <span class=\"h-card\"><a href=\"http://fossbros-anonymous.io/@foss_satan\" class=\"u-url mention\">@<span>foss_satan</span></a></span>\n\n#Hashtag\n\nText"
	replaceHashtagsExpected               = "Another test @foss_satan@fossbros-anonymous.io\n\n<a href=\"http://localhost:8080/tags/Hashtag\" class=\"mention hashtag\" rel=\"tag\">#<span>Hashtag</span></a>\n\nText"
	replaceHashtagsAfterMentionsExpected  = "Another test <span class=\"h-card\"><a href=\"http://fossbros-anonymous.io/@foss_satan\" class=\"u-url mention\">@<span>foss_satan</span></a></span>\n\n<a href=\"http://localhost:8080/tags/Hashtag\" class=\"mention hashtag\" rel=\"tag\">#<span>Hashtag</span></a>\n\nText"
	replaceMentionsWithLinkString         = "Another test @foss_satan@fossbros-anonymous.io\n\nhttp://fossbros-anonymous.io/@foss_satan/statuses/6675ee73-fccc-4562-a46a-3e8cd9798060"
	replaceMentionsWithLinkStringExpected = "Another test <span class=\"h-card\"><a href=\"http://fossbros-anonymous.io/@foss_satan\" class=\"u-url mention\">@<span>foss_satan</span></a></span>\n\nhttp://fossbros-anonymous.io/@foss_satan/statuses/6675ee73-fccc-4562-a46a-3e8cd9798060"
	replaceMentionsWithLinkSelfString     = "Mentioning myself: @the_mighty_zork\n\nand linking to my own status: https://localhost:8080/@the_mighty_zork/statuses/01FGXKJRX2PMERJQ9EQF8Y6HCR"
	replaceMemtionsWithLinkSelfExpected   = "Mentioning myself: <span class=\"h-card\"><a href=\"http://localhost:8080/@the_mighty_zork\" class=\"u-url mention\">@<span>the_mighty_zork</span></a></span>\n\nand linking to my own status: https://localhost:8080/@the_mighty_zork/statuses/01FGXKJRX2PMERJQ9EQF8Y6HCR"
)

type CommonTestSuite struct {
	TextStandardTestSuite
}

func (suite *CommonTestSuite) TestReplaceMentions() {
	foundMentions := []*gtsmodel.Mention{
		suite.testMentions["zork_mention_foss_satan"],
	}

	f := suite.formatter.ReplaceMentions(context.Background(), replaceMentionsString, foundMentions)
	suite.Equal(replaceMentionsExpected, f)
}

func (suite *CommonTestSuite) TestReplaceHashtags() {
	foundTags := []*gtsmodel.Tag{
		suite.testTags["Hashtag"],
	}

	f := suite.formatter.ReplaceTags(context.Background(), replaceMentionsString, foundTags)

	suite.Equal(replaceHashtagsExpected, f)
}

func (suite *CommonTestSuite) TestReplaceHashtagsAfterReplaceMentions() {
	foundTags := []*gtsmodel.Tag{
		suite.testTags["Hashtag"],
	}

	f := suite.formatter.ReplaceTags(context.Background(), replaceMentionsExpected, foundTags)

	suite.Equal(replaceHashtagsAfterMentionsExpected, f)
}

func (suite *CommonTestSuite) TestReplaceMentionsWithLink() {
	foundMentions := []*gtsmodel.Mention{
		suite.testMentions["zork_mention_foss_satan"],
	}

	f := suite.formatter.ReplaceMentions(context.Background(), replaceMentionsWithLinkString, foundMentions)
	suite.Equal(replaceMentionsWithLinkStringExpected, f)
}

func (suite *CommonTestSuite) TestReplaceMentionsWithLinkSelf() {
	mentioningAccount := suite.testAccounts["local_account_1"]

	foundMentions := []*gtsmodel.Mention{
		{
			ID:               "01FGXKN5F815DVFVD53PN9NYM6",
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
			StatusID:         "01FGXKP0S5THQXFC1D9R141DDR",
			OriginAccountID:  mentioningAccount.ID,
			TargetAccountID:  mentioningAccount.ID,
			NameString:       "@the_mighty_zork",
			TargetAccountURI: mentioningAccount.URI,
			TargetAccountURL: mentioningAccount.URL,
		},
	}

	f := suite.formatter.ReplaceMentions(context.Background(), replaceMentionsWithLinkSelfString, foundMentions)
	suite.Equal(replaceMemtionsWithLinkSelfExpected, f)
}

func TestCommonTestSuite(t *testing.T) {
	suite.Run(t, new(CommonTestSuite))
}

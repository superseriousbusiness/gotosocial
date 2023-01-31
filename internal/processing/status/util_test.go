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

package status_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

const (
	statusText1         = "Another test @foss_satan@fossbros-anonymous.io\n\n#Hashtag\n\nText"
	statusText1Expected = "<p>Another test <span class=\"h-card\"><a href=\"http://fossbros-anonymous.io/@foss_satan\" class=\"u-url mention\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">@<span>foss_satan</span></a></span><br><br><a href=\"http://localhost:8080/tags/Hashtag\" class=\"mention hashtag\" rel=\"tag nofollow noreferrer noopener\" target=\"_blank\">#<span>Hashtag</span></a><br><br>Text</p>"
	statusText2         = "Another test @foss_satan@fossbros-anonymous.io\n\n#Hashtag\n\n#hashTAG"
	status2TextExpected = "<p>Another test <span class=\"h-card\"><a href=\"http://fossbros-anonymous.io/@foss_satan\" class=\"u-url mention\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">@<span>foss_satan</span></a></span><br><br><a href=\"http://localhost:8080/tags/Hashtag\" class=\"mention hashtag\" rel=\"tag nofollow noreferrer noopener\" target=\"_blank\">#<span>Hashtag</span></a><br><br><a href=\"http://localhost:8080/tags/Hashtag\" class=\"mention hashtag\" rel=\"tag nofollow noreferrer noopener\" target=\"_blank\">#<span>hashTAG</span></a></p>"
)

type UtilTestSuite struct {
	StatusStandardTestSuite
}

func (suite *UtilTestSuite) TestProcessContent1() {
	/*
		TEST PREPARATION
	*/
	// we need to partially process the status first since processContent expects a status with some stuff already set on it
	creatingAccount := suite.testAccounts["local_account_1"]
	mentionedAccount := suite.testAccounts["remote_account_1"]
	form := &apimodel.AdvancedStatusCreateForm{
		StatusCreateRequest: apimodel.StatusCreateRequest{
			Status:      statusText1,
			MediaIDs:    []string{},
			Poll:        nil,
			InReplyToID: "",
			Sensitive:   false,
			SpoilerText: "",
			Visibility:  apimodel.VisibilityPublic,
			ScheduledAt: "",
			Language:    "en",
			Format:      apimodel.StatusFormatPlain,
		},
		AdvancedVisibilityFlagsForm: apimodel.AdvancedVisibilityFlagsForm{
			Federated: nil,
			Boostable: nil,
			Replyable: nil,
			Likeable:  nil,
		},
	}

	status := &gtsmodel.Status{
		ID: "01FCTDD78JJMX3K9KPXQ7ZQ8BJ",
	}

	/*
		ACTUAL TEST
	*/

	err := suite.status.ProcessContent(context.Background(), form, creatingAccount.ID, status)
	suite.NoError(err)
	suite.Equal(statusText1Expected, status.Content)

	suite.Len(status.Mentions, 1)
	newMention := status.Mentions[0]
	suite.Equal(mentionedAccount.ID, newMention.TargetAccountID)
	suite.Equal(creatingAccount.ID, newMention.OriginAccountID)
	suite.Equal(creatingAccount.URI, newMention.OriginAccountURI)
	suite.Equal(status.ID, newMention.StatusID)
	suite.Equal(fmt.Sprintf("@%s@%s", mentionedAccount.Username, mentionedAccount.Domain), newMention.NameString)
	suite.Equal(mentionedAccount.URI, newMention.TargetAccountURI)
	suite.Equal(mentionedAccount.URL, newMention.TargetAccountURL)
	suite.NotNil(newMention.OriginAccount)

	suite.Len(status.MentionIDs, 1)
	suite.Equal(newMention.ID, status.MentionIDs[0])
}

func (suite *UtilTestSuite) TestProcessContent2() {
	/*
		TEST PREPARATION
	*/
	// we need to partially process the status first since processContent expects a status with some stuff already set on it
	creatingAccount := suite.testAccounts["local_account_1"]
	mentionedAccount := suite.testAccounts["remote_account_1"]
	form := &apimodel.AdvancedStatusCreateForm{
		StatusCreateRequest: apimodel.StatusCreateRequest{
			Status:      statusText2,
			MediaIDs:    []string{},
			Poll:        nil,
			InReplyToID: "",
			Sensitive:   false,
			SpoilerText: "",
			Visibility:  apimodel.VisibilityPublic,
			ScheduledAt: "",
			Language:    "en",
			Format:      apimodel.StatusFormatPlain,
		},
		AdvancedVisibilityFlagsForm: apimodel.AdvancedVisibilityFlagsForm{
			Federated: nil,
			Boostable: nil,
			Replyable: nil,
			Likeable:  nil,
		},
	}

	status := &gtsmodel.Status{
		ID: "01FCTDD78JJMX3K9KPXQ7ZQ8BJ",
	}

	/*
		ACTUAL TEST
	*/

	err := suite.status.ProcessContent(context.Background(), form, creatingAccount.ID, status)
	suite.NoError(err)

	suite.Equal(status2TextExpected, status.Content)

	suite.Len(status.Mentions, 1)
	newMention := status.Mentions[0]
	suite.Equal(mentionedAccount.ID, newMention.TargetAccountID)
	suite.Equal(creatingAccount.ID, newMention.OriginAccountID)
	suite.Equal(creatingAccount.URI, newMention.OriginAccountURI)
	suite.Equal(status.ID, newMention.StatusID)
	suite.Equal(fmt.Sprintf("@%s@%s", mentionedAccount.Username, mentionedAccount.Domain), newMention.NameString)
	suite.Equal(mentionedAccount.URI, newMention.TargetAccountURI)
	suite.Equal(mentionedAccount.URL, newMention.TargetAccountURL)
	suite.NotNil(newMention.OriginAccount)

	suite.Len(status.MentionIDs, 1)
	suite.Equal(newMention.ID, status.MentionIDs[0])
}

func TestUtilTestSuite(t *testing.T) {
	suite.Run(t, new(UtilTestSuite))
}

/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

const statusText1 = `Another test @foss_satan@fossbros-anonymous.io

#Hashtag

Text`
const statusText1ExpectedFull = "<p>Another test <span class=\"h-card\"><a href=\"http://fossbros-anonymous.io/@foss_satan\" class=\"u-url mention\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">@<span>foss_satan</span></a></span><br><br><a href=\"http://localhost:8080/tags/Hashtag\" class=\"mention hashtag\" rel=\"tag nofollow noreferrer noopener\" target=\"_blank\">#<span>Hashtag</span></a><br><br>Text</p>"
const statusText1ExpectedPartial = "<p>Another test <span class=\"h-card\"><a href=\"http://fossbros-anonymous.io/@foss_satan\" class=\"u-url mention\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">@<span>foss_satan</span></a></span><br><br>#Hashtag<br><br>Text</p>"

const statusText2 = `Another test @foss_satan@fossbros-anonymous.io

#Hashtag

#hashTAG`

const status2TextExpectedFull = "<p>Another test <span class=\"h-card\"><a href=\"http://fossbros-anonymous.io/@foss_satan\" class=\"u-url mention\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">@<span>foss_satan</span></a></span><br><br><a href=\"http://localhost:8080/tags/Hashtag\" class=\"mention hashtag\" rel=\"tag nofollow noreferrer noopener\" target=\"_blank\">#<span>Hashtag</span></a><br><br><a href=\"http://localhost:8080/tags/Hashtag\" class=\"mention hashtag\" rel=\"tag nofollow noreferrer noopener\" target=\"_blank\">#<span>hashTAG</span></a></p>"

type UtilTestSuite struct {
	StatusStandardTestSuite
}

func (suite *UtilTestSuite) TestProcessMentions1() {
	creatingAccount := suite.testAccounts["local_account_1"]
	mentionedAccount := suite.testAccounts["remote_account_1"]

	form := &model.AdvancedStatusCreateForm{
		StatusCreateRequest: model.StatusCreateRequest{
			Status:      statusText1,
			MediaIDs:    []string{},
			Poll:        nil,
			InReplyToID: "",
			Sensitive:   false,
			SpoilerText: "",
			Visibility:  model.VisibilityPublic,
			ScheduledAt: "",
			Language:    "en",
			Format:      model.StatusFormatPlain,
		},
		AdvancedVisibilityFlagsForm: model.AdvancedVisibilityFlagsForm{
			Federated: nil,
			Boostable: nil,
			Replyable: nil,
			Likeable:  nil,
		},
	}

	status := &gtsmodel.Status{
		ID: "01FCTDD78JJMX3K9KPXQ7ZQ8BJ",
	}

	err := suite.status.ProcessMentions(context.Background(), form, creatingAccount.ID, status)
	assert.NoError(suite.T(), err)

	assert.Len(suite.T(), status.Mentions, 1)
	newMention := status.Mentions[0]
	assert.Equal(suite.T(), mentionedAccount.ID, newMention.TargetAccountID)
	assert.Equal(suite.T(), creatingAccount.ID, newMention.OriginAccountID)
	assert.Equal(suite.T(), creatingAccount.URI, newMention.OriginAccountURI)
	assert.Equal(suite.T(), status.ID, newMention.StatusID)
	assert.Equal(suite.T(), fmt.Sprintf("@%s@%s", mentionedAccount.Username, mentionedAccount.Domain), newMention.NameString)
	assert.Equal(suite.T(), mentionedAccount.URI, newMention.TargetAccountURI)
	assert.Equal(suite.T(), mentionedAccount.URL, newMention.TargetAccountURL)
	assert.NotNil(suite.T(), newMention.OriginAccount)

	assert.Len(suite.T(), status.MentionIDs, 1)
	assert.Equal(suite.T(), newMention.ID, status.MentionIDs[0])
}

func (suite *UtilTestSuite) TestProcessContentFull1() {

	/*
		TEST PREPARATION
	*/
	// we need to partially process the status first since processContent expects a status with some stuff already set on it
	creatingAccount := suite.testAccounts["local_account_1"]
	form := &model.AdvancedStatusCreateForm{
		StatusCreateRequest: model.StatusCreateRequest{
			Status:      statusText1,
			MediaIDs:    []string{},
			Poll:        nil,
			InReplyToID: "",
			Sensitive:   false,
			SpoilerText: "",
			Visibility:  model.VisibilityPublic,
			ScheduledAt: "",
			Language:    "en",
			Format:      model.StatusFormatPlain,
		},
		AdvancedVisibilityFlagsForm: model.AdvancedVisibilityFlagsForm{
			Federated: nil,
			Boostable: nil,
			Replyable: nil,
			Likeable:  nil,
		},
	}

	status := &gtsmodel.Status{
		ID: "01FCTDD78JJMX3K9KPXQ7ZQ8BJ",
	}

	err := suite.status.ProcessMentions(context.Background(), form, creatingAccount.ID, status)
	assert.NoError(suite.T(), err)
	assert.Empty(suite.T(), status.Content) // shouldn't be set yet

	err = suite.status.ProcessTags(context.Background(), form, creatingAccount.ID, status)
	assert.NoError(suite.T(), err)
	assert.Empty(suite.T(), status.Content) // shouldn't be set yet

	/*
		ACTUAL TEST
	*/

	err = suite.status.ProcessContent(context.Background(), form, creatingAccount.ID, status)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), statusText1ExpectedFull, status.Content)
}

func (suite *UtilTestSuite) TestProcessContentPartial1() {

	/*
		TEST PREPARATION
	*/
	// we need to partially process the status first since processContent expects a status with some stuff already set on it
	creatingAccount := suite.testAccounts["local_account_1"]
	form := &model.AdvancedStatusCreateForm{
		StatusCreateRequest: model.StatusCreateRequest{
			Status:      statusText1,
			MediaIDs:    []string{},
			Poll:        nil,
			InReplyToID: "",
			Sensitive:   false,
			SpoilerText: "",
			Visibility:  model.VisibilityPublic,
			ScheduledAt: "",
			Language:    "en",
			Format:      model.StatusFormatPlain,
		},
		AdvancedVisibilityFlagsForm: model.AdvancedVisibilityFlagsForm{
			Federated: nil,
			Boostable: nil,
			Replyable: nil,
			Likeable:  nil,
		},
	}

	status := &gtsmodel.Status{
		ID: "01FCTDD78JJMX3K9KPXQ7ZQ8BJ",
	}

	err := suite.status.ProcessMentions(context.Background(), form, creatingAccount.ID, status)
	assert.NoError(suite.T(), err)
	assert.Empty(suite.T(), status.Content) // shouldn't be set yet

	/*
		ACTUAL TEST
	*/

	err = suite.status.ProcessContent(context.Background(), form, creatingAccount.ID, status)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), statusText1ExpectedPartial, status.Content)
}

func (suite *UtilTestSuite) TestProcessMentions2() {
	creatingAccount := suite.testAccounts["local_account_1"]
	mentionedAccount := suite.testAccounts["remote_account_1"]

	form := &model.AdvancedStatusCreateForm{
		StatusCreateRequest: model.StatusCreateRequest{
			Status:      statusText2,
			MediaIDs:    []string{},
			Poll:        nil,
			InReplyToID: "",
			Sensitive:   false,
			SpoilerText: "",
			Visibility:  model.VisibilityPublic,
			ScheduledAt: "",
			Language:    "en",
			Format:      model.StatusFormatPlain,
		},
		AdvancedVisibilityFlagsForm: model.AdvancedVisibilityFlagsForm{
			Federated: nil,
			Boostable: nil,
			Replyable: nil,
			Likeable:  nil,
		},
	}

	status := &gtsmodel.Status{
		ID: "01FCTDD78JJMX3K9KPXQ7ZQ8BJ",
	}

	err := suite.status.ProcessMentions(context.Background(), form, creatingAccount.ID, status)
	assert.NoError(suite.T(), err)

	assert.Len(suite.T(), status.Mentions, 1)
	newMention := status.Mentions[0]
	assert.Equal(suite.T(), mentionedAccount.ID, newMention.TargetAccountID)
	assert.Equal(suite.T(), creatingAccount.ID, newMention.OriginAccountID)
	assert.Equal(suite.T(), creatingAccount.URI, newMention.OriginAccountURI)
	assert.Equal(suite.T(), status.ID, newMention.StatusID)
	assert.Equal(suite.T(), fmt.Sprintf("@%s@%s", mentionedAccount.Username, mentionedAccount.Domain), newMention.NameString)
	assert.Equal(suite.T(), mentionedAccount.URI, newMention.TargetAccountURI)
	assert.Equal(suite.T(), mentionedAccount.URL, newMention.TargetAccountURL)
	assert.NotNil(suite.T(), newMention.OriginAccount)

	assert.Len(suite.T(), status.MentionIDs, 1)
	assert.Equal(suite.T(), newMention.ID, status.MentionIDs[0])
}

func (suite *UtilTestSuite) TestProcessContentFull2() {

	/*
		TEST PREPARATION
	*/
	// we need to partially process the status first since processContent expects a status with some stuff already set on it
	creatingAccount := suite.testAccounts["local_account_1"]
	form := &model.AdvancedStatusCreateForm{
		StatusCreateRequest: model.StatusCreateRequest{
			Status:      statusText2,
			MediaIDs:    []string{},
			Poll:        nil,
			InReplyToID: "",
			Sensitive:   false,
			SpoilerText: "",
			Visibility:  model.VisibilityPublic,
			ScheduledAt: "",
			Language:    "en",
			Format:      model.StatusFormatPlain,
		},
		AdvancedVisibilityFlagsForm: model.AdvancedVisibilityFlagsForm{
			Federated: nil,
			Boostable: nil,
			Replyable: nil,
			Likeable:  nil,
		},
	}

	status := &gtsmodel.Status{
		ID: "01FCTDD78JJMX3K9KPXQ7ZQ8BJ",
	}

	err := suite.status.ProcessMentions(context.Background(), form, creatingAccount.ID, status)
	assert.NoError(suite.T(), err)
	assert.Empty(suite.T(), status.Content) // shouldn't be set yet

	err = suite.status.ProcessTags(context.Background(), form, creatingAccount.ID, status)
	assert.NoError(suite.T(), err)
	assert.Empty(suite.T(), status.Content) // shouldn't be set yet

	/*
		ACTUAL TEST
	*/

	err = suite.status.ProcessContent(context.Background(), form, creatingAccount.ID, status)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), status2TextExpectedFull, status.Content)
}

func (suite *UtilTestSuite) TestProcessContentPartial2() {

	/*
		TEST PREPARATION
	*/
	// we need to partially process the status first since processContent expects a status with some stuff already set on it
	creatingAccount := suite.testAccounts["local_account_1"]
	form := &model.AdvancedStatusCreateForm{
		StatusCreateRequest: model.StatusCreateRequest{
			Status:      statusText2,
			MediaIDs:    []string{},
			Poll:        nil,
			InReplyToID: "",
			Sensitive:   false,
			SpoilerText: "",
			Visibility:  model.VisibilityPublic,
			ScheduledAt: "",
			Language:    "en",
			Format:      model.StatusFormatPlain,
		},
		AdvancedVisibilityFlagsForm: model.AdvancedVisibilityFlagsForm{
			Federated: nil,
			Boostable: nil,
			Replyable: nil,
			Likeable:  nil,
		},
	}

	status := &gtsmodel.Status{
		ID: "01FCTDD78JJMX3K9KPXQ7ZQ8BJ",
	}

	err := suite.status.ProcessMentions(context.Background(), form, creatingAccount.ID, status)
	assert.NoError(suite.T(), err)
	assert.Empty(suite.T(), status.Content) // shouldn't be set yet

	/*
		ACTUAL TEST
	*/

	err = suite.status.ProcessContent(context.Background(), form, creatingAccount.ID, status)
	assert.NoError(suite.T(), err)

	fmt.Println(status.Content)
	// assert.Equal(suite.T(), statusText2ExpectedPartial, status.Content)
}

func TestUtilTestSuite(t *testing.T) {
	suite.Run(t, new(UtilTestSuite))
}

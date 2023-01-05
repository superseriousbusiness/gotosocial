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
	"testing"

	"github.com/stretchr/testify/suite"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

type StatusCreateTestSuite struct {
	StatusStandardTestSuite
}

func (suite *StatusCreateTestSuite) TestProcessContentWarningWithQuotationMarks() {
	ctx := context.Background()

	creatingAccount := suite.testAccounts["local_account_1"]
	creatingApplication := suite.testApplications["application_1"]

	statusCreateForm := &apimodel.AdvancedStatusCreateForm{
		StatusCreateRequest: apimodel.StatusCreateRequest{
			Status:      "poopoo peepee",
			MediaIDs:    []string{},
			Poll:        nil,
			InReplyToID: "",
			Sensitive:   false,
			SpoilerText: "\"test\"", // these should not be html-escaped when the final text is rendered
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

	apiStatus, err := suite.status.Create(ctx, creatingAccount, creatingApplication, statusCreateForm)
	suite.NoError(err)
	suite.NotNil(apiStatus)

	suite.Equal("\"test\"", apiStatus.SpoilerText)
}

func (suite *StatusCreateTestSuite) TestProcessContentWarningWithHTMLEscapedQuotationMarks() {
	ctx := context.Background()

	creatingAccount := suite.testAccounts["local_account_1"]
	creatingApplication := suite.testApplications["application_1"]

	statusCreateForm := &apimodel.AdvancedStatusCreateForm{
		StatusCreateRequest: apimodel.StatusCreateRequest{
			Status:      "poopoo peepee",
			MediaIDs:    []string{},
			Poll:        nil,
			InReplyToID: "",
			Sensitive:   false,
			SpoilerText: "&#34test&#34", // the html-escaped quotation marks should appear as normal quotation marks in the finished text
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

	apiStatus, err := suite.status.Create(ctx, creatingAccount, creatingApplication, statusCreateForm)
	suite.NoError(err)
	suite.NotNil(apiStatus)

	suite.Equal("\"test\"", apiStatus.SpoilerText)
}

func (suite *StatusCreateTestSuite) TestProcessStatusMarkdownWithUnderscoreEmoji() {
	ctx := context.Background()

	// update the shortcode of the rainbow emoji to surround it in underscores
	if err := suite.db.UpdateWhere(ctx, []db.Where{{Key: "shortcode", Value: "rainbow"}}, "shortcode", "_rainbow_", &gtsmodel.Emoji{}); err != nil {
		suite.FailNow(err.Error())
	}

	creatingAccount := suite.testAccounts["local_account_1"]
	creatingApplication := suite.testApplications["application_1"]

	statusCreateForm := &apimodel.AdvancedStatusCreateForm{
		StatusCreateRequest: apimodel.StatusCreateRequest{
			Status:      "poopoo peepee :_rainbow_:",
			MediaIDs:    []string{},
			Poll:        nil,
			InReplyToID: "",
			Sensitive:   false,
			Visibility:  apimodel.VisibilityPublic,
			ScheduledAt: "",
			Language:    "en",
			Format:      apimodel.StatusFormatMarkdown,
		},
		AdvancedVisibilityFlagsForm: apimodel.AdvancedVisibilityFlagsForm{
			Federated: nil,
			Boostable: nil,
			Replyable: nil,
			Likeable:  nil,
		},
	}

	apiStatus, err := suite.status.Create(ctx, creatingAccount, creatingApplication, statusCreateForm)
	suite.NoError(err)
	suite.NotNil(apiStatus)

	suite.Equal("<p>poopoo peepee :_rainbow_:</p>", apiStatus.Content)
	suite.NotEmpty(apiStatus.Emojis)
}

func (suite *StatusCreateTestSuite) TestProcessStatusMarkdownWithSpoilerTextEmoji() {
	ctx := context.Background()
	creatingAccount := suite.testAccounts["local_account_1"]
	creatingApplication := suite.testApplications["application_1"]

	statusCreateForm := &apimodel.AdvancedStatusCreateForm{
		StatusCreateRequest: apimodel.StatusCreateRequest{
			Status:      "poopoo peepee",
			SpoilerText: "testing something :rainbow:",
			MediaIDs:    []string{},
			Poll:        nil,
			InReplyToID: "",
			Sensitive:   false,
			Visibility:  apimodel.VisibilityPublic,
			ScheduledAt: "",
			Language:    "en",
			Format:      apimodel.StatusFormatMarkdown,
		},
		AdvancedVisibilityFlagsForm: apimodel.AdvancedVisibilityFlagsForm{
			Federated: nil,
			Boostable: nil,
			Replyable: nil,
			Likeable:  nil,
		},
	}

	apiStatus, err := suite.status.Create(ctx, creatingAccount, creatingApplication, statusCreateForm)
	suite.NoError(err)
	suite.NotNil(apiStatus)

	suite.Equal("<p>poopoo peepee</p>", apiStatus.Content)
	suite.Equal("testing something :rainbow:", apiStatus.SpoilerText)
	suite.NotEmpty(apiStatus.Emojis)
}

func (suite *StatusCreateTestSuite) TestProcessMediaDescriptionTooShort() {
	ctx := context.Background()

	config.SetMediaDescriptionMinChars(100)

	creatingAccount := suite.testAccounts["local_account_1"]
	creatingApplication := suite.testApplications["application_1"]

	statusCreateForm := &apimodel.AdvancedStatusCreateForm{
		StatusCreateRequest: apimodel.StatusCreateRequest{
			Status:      "poopoo peepee",
			MediaIDs:    []string{suite.testAttachments["local_account_1_unattached_1"].ID},
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

	apiStatus, err := suite.status.Create(ctx, creatingAccount, creatingApplication, statusCreateForm)
	suite.EqualError(err, "ProcessMediaIDs: description too short! media description of at least 100 chararacters is required but 15 was provided for media with id 01F8MH8RMYQ6MSNY3JM2XT1CQ5")
	suite.Nil(apiStatus)
}

func TestStatusCreateTestSuite(t *testing.T) {
	suite.Run(t, new(StatusCreateTestSuite))
}

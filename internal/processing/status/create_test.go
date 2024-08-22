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

package status_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

type StatusCreateTestSuite struct {
	StatusStandardTestSuite
}

func (suite *StatusCreateTestSuite) TestProcessContentWarningWithQuotationMarks() {
	ctx := context.Background()

	creatingAccount := suite.testAccounts["local_account_1"]
	creatingApplication := suite.testApplications["application_1"]

	statusCreateForm := &apimodel.StatusCreateRequest{
		Status:      "poopoo peepee",
		MediaIDs:    []string{},
		Poll:        nil,
		InReplyToID: "",
		Sensitive:   false,
		SpoilerText: "\"test\"", // these should not be html-escaped when the final text is rendered
		Visibility:  apimodel.VisibilityPublic,
		LocalOnly:   util.Ptr(false),
		ScheduledAt: "",
		Language:    "en",
		ContentType: apimodel.StatusContentTypePlain,
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

	statusCreateForm := &apimodel.StatusCreateRequest{
		Status:      "poopoo peepee",
		MediaIDs:    []string{},
		Poll:        nil,
		InReplyToID: "",
		Sensitive:   false,
		SpoilerText: "&#34test&#34", // the html-escaped quotation marks should appear as normal quotation marks in the finished text
		Visibility:  apimodel.VisibilityPublic,
		LocalOnly:   util.Ptr(false),
		ScheduledAt: "",
		Language:    "en",
		ContentType: apimodel.StatusContentTypePlain,
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

	statusCreateForm := &apimodel.StatusCreateRequest{
		Status:      "poopoo peepee :_rainbow_:",
		MediaIDs:    []string{},
		Poll:        nil,
		InReplyToID: "",
		Sensitive:   false,
		Visibility:  apimodel.VisibilityPublic,
		LocalOnly:   util.Ptr(false),
		ScheduledAt: "",
		Language:    "en",
		ContentType: apimodel.StatusContentTypeMarkdown,
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

	statusCreateForm := &apimodel.StatusCreateRequest{
		Status:      "poopoo peepee",
		SpoilerText: "testing something :rainbow:",
		MediaIDs:    []string{},
		Poll:        nil,
		InReplyToID: "",
		Sensitive:   false,
		Visibility:  apimodel.VisibilityPublic,
		LocalOnly:   util.Ptr(false),
		ScheduledAt: "",
		Language:    "en",
		ContentType: apimodel.StatusContentTypeMarkdown,
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

	statusCreateForm := &apimodel.StatusCreateRequest{
		Status:      "poopoo peepee",
		MediaIDs:    []string{suite.testAttachments["local_account_1_unattached_1"].ID},
		Poll:        nil,
		InReplyToID: "",
		Sensitive:   false,
		SpoilerText: "",
		Visibility:  apimodel.VisibilityPublic,
		LocalOnly:   util.Ptr(false),
		ScheduledAt: "",
		Language:    "en",
		ContentType: apimodel.StatusContentTypePlain,
	}

	apiStatus, err := suite.status.Create(ctx, creatingAccount, creatingApplication, statusCreateForm)
	suite.EqualError(err, "media 01F8MH8RMYQ6MSNY3JM2XT1CQ5 description too short, at least 100 required")
	suite.Nil(apiStatus)
}

func (suite *StatusCreateTestSuite) TestProcessLanguageWithScriptPart() {
	ctx := context.Background()

	creatingAccount := suite.testAccounts["local_account_1"]
	creatingApplication := suite.testApplications["application_1"]

	statusCreateForm := &apimodel.StatusCreateRequest{
		Status:      "你好世界", // hello world
		MediaIDs:    []string{},
		Poll:        nil,
		InReplyToID: "",
		Sensitive:   false,
		SpoilerText: "",
		Visibility:  apimodel.VisibilityPublic,
		LocalOnly:   util.Ptr(false),
		ScheduledAt: "",
		Language:    "zh-Hans",
		ContentType: apimodel.StatusContentTypePlain,
	}

	apiStatus, err := suite.status.Create(ctx, creatingAccount, creatingApplication, statusCreateForm)
	suite.NoError(err)
	suite.NotNil(apiStatus)

	suite.Equal("zh-Hans", *apiStatus.Language)
}

func (suite *StatusCreateTestSuite) TestProcessReplyToUnthreadedRemoteStatus() {
	ctx := context.Background()

	creatingAccount := suite.testAccounts["local_account_1"]
	creatingApplication := suite.testApplications["application_1"]
	inReplyTo := suite.testStatuses["remote_account_1_status_1"]

	// Reply to a remote status that
	// doesn't have a threadID set on it.
	statusCreateForm := &apimodel.StatusCreateRequest{
		Status:      "boobies",
		MediaIDs:    []string{},
		Poll:        nil,
		InReplyToID: inReplyTo.ID,
		Sensitive:   false,
		SpoilerText: "this is a reply",
		Visibility:  apimodel.VisibilityPublic,
		LocalOnly:   util.Ptr(false),
		ScheduledAt: "",
		Language:    "en",
		ContentType: apimodel.StatusContentTypePlain,
	}

	apiStatus, err := suite.status.Create(ctx, creatingAccount, creatingApplication, statusCreateForm)
	suite.NoError(err)
	suite.NotNil(apiStatus)

	// ThreadID should be set on the status,
	// even though the replied-to status does
	// not have a threadID.
	dbStatus, dbErr := suite.state.DB.GetStatusByID(ctx, apiStatus.ID)
	if dbErr != nil {
		suite.FailNow(err.Error())
	}
	suite.NotEmpty(dbStatus.ThreadID)
}

func TestStatusCreateTestSuite(t *testing.T) {
	suite.Run(t, new(StatusCreateTestSuite))
}

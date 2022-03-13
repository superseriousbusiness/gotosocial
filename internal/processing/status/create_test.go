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
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/api/model"
)

type StatusCreateTestSuite struct {
	StatusStandardTestSuite
}

func (suite *StatusCreateTestSuite) TestProcessContentWarningWithQuotationMarks() {
	ctx := context.Background()

	creatingAccount := suite.testAccounts["local_account_1"]
	creatingApplication := suite.testApplications["application_1"]

	statusCreateForm := &model.AdvancedStatusCreateForm{
		StatusCreateRequest: model.StatusCreateRequest{
			Status:      "poopoo peepee",
			MediaIDs:    []string{},
			Poll:        nil,
			InReplyToID: "",
			Sensitive:   false,
			SpoilerText: "\"test\"", // these should not be html-escaped when the final text is rendered
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

	apiStatus, err := suite.status.Create(ctx, creatingAccount, creatingApplication, statusCreateForm)
	suite.NoError(err)
	suite.NotNil(apiStatus)

	suite.Equal("\"test\"", apiStatus.SpoilerText)
}

func (suite *StatusCreateTestSuite) TestProcessContentWarningWithHTMLEscapedQuotationMarks() {
	ctx := context.Background()

	creatingAccount := suite.testAccounts["local_account_1"]
	creatingApplication := suite.testApplications["application_1"]

	statusCreateForm := &model.AdvancedStatusCreateForm{
		StatusCreateRequest: model.StatusCreateRequest{
			Status:      "poopoo peepee",
			MediaIDs:    []string{},
			Poll:        nil,
			InReplyToID: "",
			Sensitive:   false,
			SpoilerText: "&#34test&#34", // the html-escaped quotation marks should appear as normal quotation marks in the finished text
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

	apiStatus, err := suite.status.Create(ctx, creatingAccount, creatingApplication, statusCreateForm)
	suite.NoError(err)
	suite.NotNil(apiStatus)

	suite.Equal("\"test\"", apiStatus.SpoilerText)
}

func TestStatusCreateTestSuite(t *testing.T) {
	suite.Run(t, new(StatusCreateTestSuite))
}

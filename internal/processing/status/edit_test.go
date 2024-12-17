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
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

type StatusEditTestSuite struct {
	StatusStandardTestSuite
}

func (suite *StatusEditTestSuite) TestSimpleEdit() {
	ctx, cncl := context.WithCancel(context.Background())
	defer cncl()

	requester := suite.testAccounts["local_account_1"]
	requester, _ = suite.state.DB.GetAccountByID(ctx, requester.ID)

	status := suite.testStatuses["local_account_1_status_9"]
	status, _ = suite.state.DB.GetStatusByID(ctx, status.ID)

	form := &apimodel.StatusEditRequest{
		Status:          "<p>this is some edited status text!</p>",
		SpoilerText:     "shhhhh",
		Sensitive:       true,
		Language:        "fr", // hoh hoh hoh
		MediaIDs:        nil,
		MediaAttributes: nil,
		Poll:            nil,
	}

	apiStatus, errWithCode := suite.status.Edit(ctx, requester, status.ID, form)
	suite.NotNil(apiStatus)
	suite.Nil(errWithCode)

	suite.Equal(form.Status, apiStatus.Text)
	suite.Equal(form.SpoilerText, apiStatus.SpoilerText)
	suite.Equal(form.Sensitive, apiStatus.Sensitive)
	suite.Equal(form.Language, *apiStatus.Language)
	suite.NotEqual(util.FormatISO8601(status.UpdatedAt), *apiStatus.EditedAt)

	latestStatus, err := suite.state.DB.GetStatusByID(ctx, status.ID)
	suite.NoError(err)

	suite.Equal(form.Status, latestStatus.Text)
	suite.Equal(form.SpoilerText, latestStatus.ContentWarning)
	suite.Equal(form.Sensitive, *latestStatus.Sensitive)
	suite.Equal(form.Language, latestStatus.Language)
	suite.Equal(len(status.EditIDs)+1, len(latestStatus.EditIDs))
	suite.NotEqual(status.UpdatedAt, latestStatus.UpdatedAt)

	err = suite.state.DB.PopulateStatusEdits(ctx, latestStatus)
	suite.NoError(err)

	previousEdit := latestStatus.Edits[len(latestStatus.Edits)-1]
	suite.Equal(status.Content, previousEdit.Content)
	suite.Equal(status.Text, previousEdit.Text)
	suite.Equal(status.ContentWarning, previousEdit.ContentWarning)
	suite.Equal(*status.Sensitive, *previousEdit.Sensitive)
	suite.Equal(status.Language, previousEdit.Language)
}

func (suite *StatusEditTestSuite) TestEditOthersStatus1() {
	ctx, cncl := context.WithCancel(context.Background())
	defer cncl()

	requester := suite.testAccounts["local_account_1"]
	requester, _ = suite.state.DB.GetAccountByID(ctx, requester.ID)

	status := suite.testStatuses["remote_account_1_status_1"]
	status, _ = suite.state.DB.GetStatusByID(ctx, status.ID)

	form := &apimodel.StatusEditRequest{}

	apiStatus, errWithCode := suite.status.Edit(ctx, requester, status.ID, form)
	suite.Nil(apiStatus)
	suite.Equal(http.StatusNotFound, errWithCode.Code())
	suite.Equal("status does not belong to requester", errWithCode.Error())
	suite.Equal("Not Found: target status not found", errWithCode.Safe())
}

func (suite *StatusEditTestSuite) TestEditOthersStatus2() {
	ctx, cncl := context.WithCancel(context.Background())
	defer cncl()

	requester := suite.testAccounts["local_account_1"]
	requester, _ = suite.state.DB.GetAccountByID(ctx, requester.ID)

	status := suite.testStatuses["local_account_2_status_1"]
	status, _ = suite.state.DB.GetStatusByID(ctx, status.ID)

	form := &apimodel.StatusEditRequest{}

	apiStatus, errWithCode := suite.status.Edit(ctx, requester, status.ID, form)
	suite.Nil(apiStatus)
	suite.Equal(http.StatusNotFound, errWithCode.Code())
	suite.Equal("status does not belong to requester", errWithCode.Error())
	suite.Equal("Not Found: target status not found", errWithCode.Safe())
}

func TestStatusEditTestSuite(t *testing.T) {
	suite.Run(t, new(StatusEditTestSuite))
}

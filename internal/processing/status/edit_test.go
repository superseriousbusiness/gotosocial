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
	"time"

	"github.com/stretchr/testify/suite"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/util"
	"github.com/superseriousbusiness/gotosocial/internal/util/xslices"
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
	suite.Equal(status.UpdatedAt, previousEdit.CreatedAt)
}

func (suite *StatusEditTestSuite) TestEditAddPoll() {
	ctx, cncl := context.WithCancel(context.Background())
	defer cncl()

	requester := suite.testAccounts["local_account_1"]
	requester, _ = suite.state.DB.GetAccountByID(ctx, requester.ID)

	status := suite.testStatuses["local_account_1_status_9"]
	status, _ = suite.state.DB.GetStatusByID(ctx, status.ID)

	form := &apimodel.StatusEditRequest{
		Status:          "<p>this is some edited status text!</p>",
		SpoilerText:     "",
		Sensitive:       true,
		Language:        "fr", // hoh hoh hoh
		MediaIDs:        nil,
		MediaAttributes: nil,
		Poll: &apimodel.PollRequest{
			Options:    []string{"yes", "no", "spiderman"},
			ExpiresIn:  int(time.Minute),
			Multiple:   true,
			HideTotals: false,
		},
	}

	apiStatus, errWithCode := suite.status.Edit(ctx, requester, status.ID, form)
	suite.NotNil(apiStatus)
	suite.Nil(errWithCode)

	suite.Equal(form.Status, apiStatus.Text)
	suite.Equal(form.SpoilerText, apiStatus.SpoilerText)
	suite.Equal(form.Sensitive, apiStatus.Sensitive)
	suite.Equal(form.Language, *apiStatus.Language)
	suite.NotEqual(util.FormatISO8601(status.UpdatedAt), *apiStatus.EditedAt)
	suite.NotNil(apiStatus.Poll)
	suite.Equal(form.Poll.Options, xslices.Gather(nil, apiStatus.Poll.Options, func(opt apimodel.PollOption) string {
		return opt.Title
	}))

	latestStatus, err := suite.state.DB.GetStatusByID(ctx, status.ID)
	suite.NoError(err)

	suite.Equal(form.Status, latestStatus.Text)
	suite.Equal(form.SpoilerText, latestStatus.ContentWarning)
	suite.Equal(form.Sensitive, *latestStatus.Sensitive)
	suite.Equal(form.Language, latestStatus.Language)
	suite.Equal(len(status.EditIDs)+1, len(latestStatus.EditIDs))
	suite.NotEqual(status.UpdatedAt, latestStatus.UpdatedAt)
	suite.NotNil(latestStatus.Poll)
	suite.Equal(form.Poll.Options, latestStatus.Poll.Options)

	expiryWorker := suite.state.Workers.Scheduler.Cancel(latestStatus.PollID)
	suite.Equal(form.Poll.ExpiresIn > 0, expiryWorker)

	err = suite.state.DB.PopulateStatusEdits(ctx, latestStatus)
	suite.NoError(err)

	previousEdit := latestStatus.Edits[len(latestStatus.Edits)-1]
	suite.Equal(status.Content, previousEdit.Content)
	suite.Equal(status.Text, previousEdit.Text)
	suite.Equal(status.ContentWarning, previousEdit.ContentWarning)
	suite.Equal(*status.Sensitive, *previousEdit.Sensitive)
	suite.Equal(status.Language, previousEdit.Language)
	suite.Equal(status.UpdatedAt, previousEdit.CreatedAt)
	suite.Equal(status.Poll != nil, len(previousEdit.PollOptions) > 0)
}

func (suite *StatusEditTestSuite) TestEditAddPollNoExpiry() {
	ctx, cncl := context.WithCancel(context.Background())
	defer cncl()

	requester := suite.testAccounts["local_account_1"]
	requester, _ = suite.state.DB.GetAccountByID(ctx, requester.ID)

	status := suite.testStatuses["local_account_1_status_9"]
	status, _ = suite.state.DB.GetStatusByID(ctx, status.ID)

	form := &apimodel.StatusEditRequest{
		Status:          "<p>this is some edited status text!</p>",
		SpoilerText:     "",
		Sensitive:       true,
		Language:        "fr", // hoh hoh hoh
		MediaIDs:        nil,
		MediaAttributes: nil,
		Poll: &apimodel.PollRequest{
			Options:    []string{"yes", "no", "spiderman"},
			ExpiresIn:  0,
			Multiple:   true,
			HideTotals: false,
		},
	}

	apiStatus, errWithCode := suite.status.Edit(ctx, requester, status.ID, form)
	suite.NotNil(apiStatus)
	suite.Nil(errWithCode)

	suite.Equal(form.Status, apiStatus.Text)
	suite.Equal(form.SpoilerText, apiStatus.SpoilerText)
	suite.Equal(form.Sensitive, apiStatus.Sensitive)
	suite.Equal(form.Language, *apiStatus.Language)
	suite.NotEqual(util.FormatISO8601(status.UpdatedAt), *apiStatus.EditedAt)
	suite.NotNil(apiStatus.Poll)
	suite.Equal(form.Poll.Options, xslices.Gather(nil, apiStatus.Poll.Options, func(opt apimodel.PollOption) string {
		return opt.Title
	}))

	latestStatus, err := suite.state.DB.GetStatusByID(ctx, status.ID)
	suite.NoError(err)

	suite.Equal(form.Status, latestStatus.Text)
	suite.Equal(form.SpoilerText, latestStatus.ContentWarning)
	suite.Equal(form.Sensitive, *latestStatus.Sensitive)
	suite.Equal(form.Language, latestStatus.Language)
	suite.Equal(len(status.EditIDs)+1, len(latestStatus.EditIDs))
	suite.NotEqual(status.UpdatedAt, latestStatus.UpdatedAt)
	suite.NotNil(latestStatus.Poll)
	suite.Equal(form.Poll.Options, latestStatus.Poll.Options)

	expiryWorker := suite.state.Workers.Scheduler.Cancel(latestStatus.PollID)
	suite.Equal(form.Poll.ExpiresIn > 0, expiryWorker)

	err = suite.state.DB.PopulateStatusEdits(ctx, latestStatus)
	suite.NoError(err)

	previousEdit := latestStatus.Edits[len(latestStatus.Edits)-1]
	suite.Equal(status.Content, previousEdit.Content)
	suite.Equal(status.Text, previousEdit.Text)
	suite.Equal(status.ContentWarning, previousEdit.ContentWarning)
	suite.Equal(*status.Sensitive, *previousEdit.Sensitive)
	suite.Equal(status.Language, previousEdit.Language)
	suite.Equal(status.UpdatedAt, previousEdit.CreatedAt)
	suite.Equal(status.Poll != nil, len(previousEdit.PollOptions) > 0)
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

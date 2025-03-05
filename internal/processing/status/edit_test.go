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
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/internal/util"
	"github.com/superseriousbusiness/gotosocial/internal/util/xslices"
)

type StatusEditTestSuite struct {
	StatusStandardTestSuite
}

func (suite *StatusEditTestSuite) TestSimpleEdit() {
	// Create cancellable context to use for test.
	ctx, cncl := context.WithCancel(context.Background())
	defer cncl()

	// Get a local account to use as test requester.
	requester := suite.testAccounts["local_account_1"]
	requester, _ = suite.state.DB.GetAccountByID(ctx, requester.ID)

	// Get requester's existing status to perform an edit on.
	status := suite.testStatuses["local_account_1_status_9"]
	status, _ = suite.state.DB.GetStatusByID(ctx, status.ID)

	// Prepare a simple status edit.
	form := &apimodel.StatusEditRequest{
		Status:          "<p>this is some edited status text!</p>",
		SpoilerText:     "shhhhh",
		Sensitive:       true,
		Language:        "fr", // hoh hoh hoh
		MediaIDs:        nil,
		MediaAttributes: nil,
		Poll:            nil,
	}

	// Pass the prepared form to the status processor to perform the edit.
	apiStatus, errWithCode := suite.status.Edit(ctx, requester, status.ID, form)
	suite.NotNil(apiStatus)
	suite.NoError(errWithCode)

	// Check response against input form data.
	suite.Equal(form.Status, apiStatus.Text)
	suite.Equal(form.SpoilerText, apiStatus.SpoilerText)
	suite.Equal(form.Sensitive, apiStatus.Sensitive)
	suite.Equal(form.Language, *apiStatus.Language)
	suite.NotEqual(util.FormatISO8601(status.EditedAt), *apiStatus.EditedAt)

	// Fetched the latest version of edited status from the database.
	latestStatus, err := suite.state.DB.GetStatusByID(ctx, status.ID)
	suite.NoError(err)

	// Check latest status against input form data.
	suite.Equal(form.Status, latestStatus.Text)
	suite.Equal(form.SpoilerText, latestStatus.ContentWarning)
	suite.Equal(form.Sensitive, *latestStatus.Sensitive)
	suite.Equal(form.Language, latestStatus.Language)
	suite.Equal(len(status.EditIDs)+1, len(latestStatus.EditIDs))
	suite.NotEqual(status.UpdatedAt(), latestStatus.UpdatedAt())

	// Populate all historical edits for this status.
	err = suite.state.DB.PopulateStatusEdits(ctx, latestStatus)
	suite.NoError(err)

	// Check previous status edit matches original status content.
	previousEdit := latestStatus.Edits[len(latestStatus.Edits)-1]
	suite.Equal(status.Content, previousEdit.Content)
	suite.Equal(status.Text, previousEdit.Text)
	suite.Equal(status.ContentType, previousEdit.ContentType)
	suite.Equal(status.ContentWarning, previousEdit.ContentWarning)
	suite.Equal(*status.Sensitive, *previousEdit.Sensitive)
	suite.Equal(status.Language, previousEdit.Language)
	suite.Equal(status.UpdatedAt(), previousEdit.CreatedAt)
}

func (suite *StatusEditTestSuite) TestEditChangeContentType() {
	// Create cancellable context to use for test.
	ctx, cncl := context.WithCancel(context.Background())
	defer cncl()

	// Get a local account to use as test requester.
	requester := suite.testAccounts["local_account_1"]
	requester, _ = suite.state.DB.GetAccountByID(ctx, requester.ID)

	// Get requester's existing plain text status to perform an edit on.
	status := suite.testStatuses["local_account_1_status_6"]
	status, _ = suite.state.DB.GetStatusByID(ctx, status.ID)

	// Prepare edit with a Markdown body.
	form := &apimodel.StatusEditRequest{
		Status:          "ooh the status is *fancy* now!",
		ContentType:     apimodel.StatusContentTypeMarkdown,
		SpoilerText:     "shhhhh",
		Sensitive:       true,
		Language:        "fr", // hoh hoh hoh
		MediaIDs:        nil,
		MediaAttributes: nil,
		Poll:            nil,
	}

	// Pass the prepared form to the status processor to perform the edit.
	apiStatus, errWithCode := suite.status.Edit(ctx, requester, status.ID, form)
	suite.NotNil(apiStatus)
	suite.NoError(errWithCode)

	// Check response against input form data.
	suite.Equal(form.Status, apiStatus.Text)
	suite.Equal(form.ContentType, apiStatus.ContentType)
	suite.Equal(form.SpoilerText, apiStatus.SpoilerText)
	suite.Equal(form.Sensitive, apiStatus.Sensitive)
	suite.Equal(form.Language, *apiStatus.Language)
	suite.NotEqual(util.FormatISO8601(status.EditedAt), *apiStatus.EditedAt)

	// Fetched the latest version of edited status from the database.
	latestStatus, err := suite.state.DB.GetStatusByID(ctx, status.ID)
	suite.NoError(err)

	// Check latest status against input form data.
	suite.Equal(form.Status, latestStatus.Text)
	suite.Equal(typeutils.APIContentTypeToContentType(form.ContentType), latestStatus.ContentType)
	suite.Equal(form.SpoilerText, latestStatus.ContentWarning)
	suite.Equal(form.Sensitive, *latestStatus.Sensitive)
	suite.Equal(form.Language, latestStatus.Language)
	suite.Equal(len(status.EditIDs)+1, len(latestStatus.EditIDs))
	suite.NotEqual(status.UpdatedAt(), latestStatus.UpdatedAt())

	// Populate all historical edits for this status.
	err = suite.state.DB.PopulateStatusEdits(ctx, latestStatus)
	suite.NoError(err)

	// Check previous status edit matches original status content.
	previousEdit := latestStatus.Edits[len(latestStatus.Edits)-1]
	suite.Equal(status.Content, previousEdit.Content)
	suite.Equal(status.Text, previousEdit.Text)
	suite.Equal(status.ContentType, previousEdit.ContentType)
	suite.Equal(status.ContentWarning, previousEdit.ContentWarning)
	suite.Equal(*status.Sensitive, *previousEdit.Sensitive)
	suite.Equal(status.Language, previousEdit.Language)
	suite.Equal(status.UpdatedAt(), previousEdit.CreatedAt)
}

func (suite *StatusEditTestSuite) TestEditOnStatusWithNoContentType() {
	// Create cancellable context to use for test.
	ctx, cncl := context.WithCancel(context.Background())
	defer cncl()

	// Get a local account to use as test requester.
	requester := suite.testAccounts["local_account_1"]
	requester, _ = suite.state.DB.GetAccountByID(ctx, requester.ID)

	// Get requester's existing status, which has no
	// stored content type, to perform an edit on.
	status := suite.testStatuses["local_account_1_status_2"]
	status, _ = suite.state.DB.GetStatusByID(ctx, status.ID)

	// Prepare edit without setting a new content type.
	form := &apimodel.StatusEditRequest{
		Status:          "how will this text be parsed? it is a mystery",
		SpoilerText:     "shhhhh",
		Sensitive:       true,
		Language:        "fr", // hoh hoh hoh
		MediaIDs:        nil,
		MediaAttributes: nil,
		Poll:            nil,
	}

	// Pass the prepared form to the status processor to perform the edit.
	apiStatus, errWithCode := suite.status.Edit(ctx, requester, status.ID, form)
	suite.NotNil(apiStatus)
	suite.NoError(errWithCode)

	// Check response against input form data.
	suite.Equal(form.Status, apiStatus.Text)
	suite.NotEqual(util.FormatISO8601(status.EditedAt), *apiStatus.EditedAt)

	// Check response against requester's default content type setting
	// (the test accounts don't actually have settings on them, so
	// instead we check that the global default content type is used)
	suite.Equal(apimodel.StatusContentTypeDefault, apiStatus.ContentType)

	// Fetched the latest version of edited status from the database.
	latestStatus, err := suite.state.DB.GetStatusByID(ctx, status.ID)
	suite.NoError(err)

	// Check latest status against input form data
	suite.Equal(form.Status, latestStatus.Text)
	suite.Equal(form.Sensitive, *latestStatus.Sensitive)
	suite.Equal(form.Language, latestStatus.Language)
	suite.Equal(len(status.EditIDs)+1, len(latestStatus.EditIDs))
	suite.NotEqual(status.UpdatedAt(), latestStatus.UpdatedAt())

	// Check latest status against requester's default content
	// type (again, actually just checking for the global default)
	suite.Equal(gtsmodel.StatusContentTypeDefault, latestStatus.ContentType)

	// Populate all historical edits for this status.
	err = suite.state.DB.PopulateStatusEdits(ctx, latestStatus)
	suite.NoError(err)

	// Check previous status edit matches original status content.
	previousEdit := latestStatus.Edits[len(latestStatus.Edits)-1]
	suite.Equal(status.Content, previousEdit.Content)
	suite.Equal(status.Text, previousEdit.Text)
	suite.Equal(status.ContentType, previousEdit.ContentType)
	suite.Equal(status.ContentWarning, previousEdit.ContentWarning)
	suite.Equal(*status.Sensitive, *previousEdit.Sensitive)
	suite.Equal(status.Language, previousEdit.Language)
	suite.Equal(status.UpdatedAt(), previousEdit.CreatedAt)
}

func (suite *StatusEditTestSuite) TestEditAddPoll() {
	// Create cancellable context to use for test.
	ctx, cncl := context.WithCancel(context.Background())
	defer cncl()

	// Get a local account to use as test requester.
	requester := suite.testAccounts["local_account_1"]
	requester, _ = suite.state.DB.GetAccountByID(ctx, requester.ID)

	// Get requester's existing status to perform an edit on.
	status := suite.testStatuses["local_account_1_status_9"]
	status, _ = suite.state.DB.GetStatusByID(ctx, status.ID)

	// Prepare edit adding a status poll.
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

	// Pass the prepared form to the status processor to perform the edit.
	apiStatus, errWithCode := suite.status.Edit(ctx, requester, status.ID, form)
	suite.NotNil(apiStatus)
	suite.NoError(errWithCode)

	// Check response against input form data.
	suite.Equal(form.Status, apiStatus.Text)
	suite.Equal(form.SpoilerText, apiStatus.SpoilerText)
	suite.Equal(form.Sensitive, apiStatus.Sensitive)
	suite.Equal(form.Language, *apiStatus.Language)
	suite.NotEqual(util.FormatISO8601(status.EditedAt), *apiStatus.EditedAt)
	suite.NotNil(apiStatus.Poll)
	suite.Equal(form.Poll.Options, xslices.Gather(nil, apiStatus.Poll.Options, func(opt apimodel.PollOption) string {
		return opt.Title
	}))

	// Fetched the latest version of edited status from the database.
	latestStatus, err := suite.state.DB.GetStatusByID(ctx, status.ID)
	suite.NoError(err)

	// Check latest status against input form data.
	suite.Equal(form.Status, latestStatus.Text)
	suite.Equal(form.SpoilerText, latestStatus.ContentWarning)
	suite.Equal(form.Sensitive, *latestStatus.Sensitive)
	suite.Equal(form.Language, latestStatus.Language)
	suite.Equal(len(status.EditIDs)+1, len(latestStatus.EditIDs))
	suite.NotEqual(status.UpdatedAt(), latestStatus.UpdatedAt())
	suite.NotNil(latestStatus.Poll)
	suite.Equal(form.Poll.Options, latestStatus.Poll.Options)

	// Ensure that a poll expiry handler was scheduled on status edit.
	expiryWorker := suite.state.Workers.Scheduler.Cancel(latestStatus.PollID)
	suite.Equal(form.Poll.ExpiresIn > 0, expiryWorker)

	// Populate all historical edits for this status.
	err = suite.state.DB.PopulateStatusEdits(ctx, latestStatus)
	suite.NoError(err)

	// Check previous status edit matches original status content.
	previousEdit := latestStatus.Edits[len(latestStatus.Edits)-1]
	suite.Equal(status.Content, previousEdit.Content)
	suite.Equal(status.Text, previousEdit.Text)
	suite.Equal(status.ContentWarning, previousEdit.ContentWarning)
	suite.Equal(*status.Sensitive, *previousEdit.Sensitive)
	suite.Equal(status.Language, previousEdit.Language)
	suite.Equal(status.UpdatedAt(), previousEdit.CreatedAt)
	suite.Equal(status.Poll != nil, len(previousEdit.PollOptions) > 0)
}

func (suite *StatusEditTestSuite) TestEditAddPollNoExpiry() {
	// Create cancellable context to use for test.
	ctx, cncl := context.WithCancel(context.Background())
	defer cncl()

	// Get a local account to use as test requester.
	requester := suite.testAccounts["local_account_1"]
	requester, _ = suite.state.DB.GetAccountByID(ctx, requester.ID)

	// Get requester's existing status to perform an edit on.
	status := suite.testStatuses["local_account_1_status_9"]
	status, _ = suite.state.DB.GetStatusByID(ctx, status.ID)

	// Prepare edit adding an endless poll.
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

	// Pass the prepared form to the status processor to perform the edit.
	apiStatus, errWithCode := suite.status.Edit(ctx, requester, status.ID, form)
	suite.NotNil(apiStatus)
	suite.NoError(errWithCode)

	// Check response against input form data.
	suite.Equal(form.Status, apiStatus.Text)
	suite.Equal(form.SpoilerText, apiStatus.SpoilerText)
	suite.Equal(form.Sensitive, apiStatus.Sensitive)
	suite.Equal(form.Language, *apiStatus.Language)
	suite.NotEqual(util.FormatISO8601(status.EditedAt), *apiStatus.EditedAt)
	suite.NotNil(apiStatus.Poll)
	suite.Equal(form.Poll.Options, xslices.Gather(nil, apiStatus.Poll.Options, func(opt apimodel.PollOption) string {
		return opt.Title
	}))

	// Fetched the latest version of edited status from the database.
	latestStatus, err := suite.state.DB.GetStatusByID(ctx, status.ID)
	suite.NoError(err)

	// Check latest status against input form data.
	suite.Equal(form.Status, latestStatus.Text)
	suite.Equal(form.SpoilerText, latestStatus.ContentWarning)
	suite.Equal(form.Sensitive, *latestStatus.Sensitive)
	suite.Equal(form.Language, latestStatus.Language)
	suite.Equal(len(status.EditIDs)+1, len(latestStatus.EditIDs))
	suite.NotEqual(status.UpdatedAt(), latestStatus.UpdatedAt())
	suite.NotNil(latestStatus.Poll)
	suite.Equal(form.Poll.Options, latestStatus.Poll.Options)

	// Ensure that a poll expiry handler was *not* scheduled on status edit.
	expiryWorker := suite.state.Workers.Scheduler.Cancel(latestStatus.PollID)
	suite.Equal(form.Poll.ExpiresIn > 0, expiryWorker)

	// Populate all historical edits for this status.
	err = suite.state.DB.PopulateStatusEdits(ctx, latestStatus)
	suite.NoError(err)

	// Check previous status edit matches original status content.
	previousEdit := latestStatus.Edits[len(latestStatus.Edits)-1]
	suite.Equal(status.Content, previousEdit.Content)
	suite.Equal(status.Text, previousEdit.Text)
	suite.Equal(status.ContentWarning, previousEdit.ContentWarning)
	suite.Equal(*status.Sensitive, *previousEdit.Sensitive)
	suite.Equal(status.Language, previousEdit.Language)
	suite.Equal(status.UpdatedAt(), previousEdit.CreatedAt)
	suite.Equal(status.Poll != nil, len(previousEdit.PollOptions) > 0)
}

func (suite *StatusEditTestSuite) TestEditMediaDescription() {
	// Create cancellable context to use for test.
	ctx, cncl := context.WithCancel(context.Background())
	defer cncl()

	// Get a local account to use as test requester.
	requester := suite.testAccounts["local_account_1"]
	requester, _ = suite.state.DB.GetAccountByID(ctx, requester.ID)

	// Get requester's existing status to perform an edit on.
	status := suite.testStatuses["local_account_1_status_4"]
	status, _ = suite.state.DB.GetStatusByID(ctx, status.ID)

	// Prepare edit changing media description.
	form := &apimodel.StatusEditRequest{
		Status:      "<p>this is some edited status text!</p>",
		SpoilerText: "this status is now missing media",
		Sensitive:   true,
		Language:    "en",
		MediaIDs:    status.AttachmentIDs,
		MediaAttributes: []apimodel.AttachmentAttributesRequest{
			{ID: status.AttachmentIDs[0], Description: "hello world!"},
			{ID: status.AttachmentIDs[1], Description: "media attachment numero two"},
		},
	}

	// Pass the prepared form to the status processor to perform the edit.
	apiStatus, errWithCode := suite.status.Edit(ctx, requester, status.ID, form)
	suite.NoError(errWithCode)

	// Check response against input form data.
	suite.Equal(form.Status, apiStatus.Text)
	suite.Equal(form.SpoilerText, apiStatus.SpoilerText)
	suite.Equal(form.Sensitive, apiStatus.Sensitive)
	suite.Equal(form.Language, *apiStatus.Language)
	suite.NotEqual(util.FormatISO8601(status.EditedAt), *apiStatus.EditedAt)
	suite.Equal(form.MediaIDs, xslices.Gather(nil, apiStatus.MediaAttachments, func(media *apimodel.Attachment) string {
		return media.ID
	}))
	suite.Equal(
		xslices.Gather(nil, form.MediaAttributes, func(attr apimodel.AttachmentAttributesRequest) string {
			return attr.Description
		}),
		xslices.Gather(nil, apiStatus.MediaAttachments, func(media *apimodel.Attachment) string {
			return *media.Description
		}),
	)

	// Fetched the latest version of edited status from the database.
	latestStatus, err := suite.state.DB.GetStatusByID(ctx, status.ID)
	suite.NoError(err)

	// Check latest status against input form data.
	suite.Equal(form.Status, latestStatus.Text)
	suite.Equal(form.SpoilerText, latestStatus.ContentWarning)
	suite.Equal(form.Sensitive, *latestStatus.Sensitive)
	suite.Equal(form.Language, latestStatus.Language)
	suite.Equal(len(status.EditIDs)+1, len(latestStatus.EditIDs))
	suite.NotEqual(status.UpdatedAt(), latestStatus.UpdatedAt())
	suite.Equal(form.MediaIDs, latestStatus.AttachmentIDs)
	suite.Equal(
		xslices.Gather(nil, form.MediaAttributes, func(attr apimodel.AttachmentAttributesRequest) string {
			return attr.Description
		}),
		xslices.Gather(nil, latestStatus.Attachments, func(media *gtsmodel.MediaAttachment) string {
			return media.Description
		}),
	)

	// Populate all historical edits for this status.
	err = suite.state.DB.PopulateStatusEdits(ctx, latestStatus)
	suite.NoError(err)

	// Further populate edits to get attachments.
	for _, edit := range latestStatus.Edits {
		err = suite.state.DB.PopulateStatusEdit(ctx, edit)
		suite.NoError(err)
	}

	// Check previous status edit matches original status content.
	previousEdit := latestStatus.Edits[len(latestStatus.Edits)-1]
	suite.Equal(status.Content, previousEdit.Content)
	suite.Equal(status.Text, previousEdit.Text)
	suite.Equal(status.ContentWarning, previousEdit.ContentWarning)
	suite.Equal(*status.Sensitive, *previousEdit.Sensitive)
	suite.Equal(status.Language, previousEdit.Language)
	suite.Equal(status.UpdatedAt(), previousEdit.CreatedAt)
	suite.Equal(status.AttachmentIDs, previousEdit.AttachmentIDs)
	suite.Equal(
		xslices.Gather(nil, status.Attachments, func(media *gtsmodel.MediaAttachment) string {
			return media.Description
		}),
		previousEdit.AttachmentDescriptions,
	)
}

func (suite *StatusEditTestSuite) TestEditAddMedia() {
	// Create cancellable context to use for test.
	ctx, cncl := context.WithCancel(context.Background())
	defer cncl()

	// Get a local account to use as test requester.
	requester := suite.testAccounts["local_account_1"]
	requester, _ = suite.state.DB.GetAccountByID(ctx, requester.ID)

	// Get some of requester's existing media, and unattach from existing status.
	media1 := suite.testAttachments["local_account_1_status_4_attachment_1"]
	media2 := suite.testAttachments["local_account_1_status_4_attachment_2"]
	media1.StatusID, media2.StatusID = "", ""
	suite.NoError(suite.state.DB.UpdateAttachment(ctx, media1, "status_id"))
	suite.NoError(suite.state.DB.UpdateAttachment(ctx, media2, "status_id"))
	media1, _ = suite.state.DB.GetAttachmentByID(ctx, media1.ID)
	media2, _ = suite.state.DB.GetAttachmentByID(ctx, media2.ID)

	// Get requester's existing status to perform an edit on.
	status := suite.testStatuses["local_account_1_status_9"]
	status, _ = suite.state.DB.GetStatusByID(ctx, status.ID)

	// Prepare edit addding status media.
	form := &apimodel.StatusEditRequest{
		Status:          "<p>this is some edited status text!</p>",
		SpoilerText:     "this status now has media",
		Sensitive:       true,
		Language:        "en",
		MediaIDs:        []string{media1.ID, media2.ID},
		MediaAttributes: nil,
	}

	// Pass the prepared form to the status processor to perform the edit.
	apiStatus, errWithCode := suite.status.Edit(ctx, requester, status.ID, form)
	suite.NotNil(apiStatus)
	suite.NoError(errWithCode)

	// Check response against input form data.
	suite.Equal(form.Status, apiStatus.Text)
	suite.Equal(form.SpoilerText, apiStatus.SpoilerText)
	suite.Equal(form.Sensitive, apiStatus.Sensitive)
	suite.Equal(form.Language, *apiStatus.Language)
	suite.NotEqual(util.FormatISO8601(status.EditedAt), *apiStatus.EditedAt)
	suite.Equal(form.MediaIDs, xslices.Gather(nil, apiStatus.MediaAttachments, func(media *apimodel.Attachment) string {
		return media.ID
	}))

	// Fetched the latest version of edited status from the database.
	latestStatus, err := suite.state.DB.GetStatusByID(ctx, status.ID)
	suite.NoError(err)

	// Check latest status against input form data.
	suite.Equal(form.Status, latestStatus.Text)
	suite.Equal(form.SpoilerText, latestStatus.ContentWarning)
	suite.Equal(form.Sensitive, *latestStatus.Sensitive)
	suite.Equal(form.Language, latestStatus.Language)
	suite.Equal(len(status.EditIDs)+1, len(latestStatus.EditIDs))
	suite.NotEqual(status.UpdatedAt(), latestStatus.UpdatedAt())
	suite.Equal(form.MediaIDs, latestStatus.AttachmentIDs)

	// Populate all historical edits for this status.
	err = suite.state.DB.PopulateStatusEdits(ctx, latestStatus)
	suite.NoError(err)

	// Check previous status edit matches original status content.
	previousEdit := latestStatus.Edits[len(latestStatus.Edits)-1]
	suite.Equal(status.Content, previousEdit.Content)
	suite.Equal(status.Text, previousEdit.Text)
	suite.Equal(status.ContentWarning, previousEdit.ContentWarning)
	suite.Equal(*status.Sensitive, *previousEdit.Sensitive)
	suite.Equal(status.Language, previousEdit.Language)
	suite.Equal(status.UpdatedAt(), previousEdit.CreatedAt)
	suite.Equal(status.AttachmentIDs, previousEdit.AttachmentIDs)
}

func (suite *StatusEditTestSuite) TestEditRemoveMedia() {
	// Create cancellable context to use for test.
	ctx, cncl := context.WithCancel(context.Background())
	defer cncl()

	// Get a local account to use as test requester.
	requester := suite.testAccounts["local_account_1"]
	requester, _ = suite.state.DB.GetAccountByID(ctx, requester.ID)

	// Get requester's existing status to perform an edit on.
	status := suite.testStatuses["local_account_1_status_4"]
	status, _ = suite.state.DB.GetStatusByID(ctx, status.ID)

	// Prepare edit removing status media.
	form := &apimodel.StatusEditRequest{
		Status:          "<p>this is some edited status text!</p>",
		SpoilerText:     "this status is now missing media",
		Sensitive:       true,
		Language:        "en",
		MediaIDs:        nil,
		MediaAttributes: nil,
	}

	// Pass the prepared form to the status processor to perform the edit.
	apiStatus, errWithCode := suite.status.Edit(ctx, requester, status.ID, form)
	suite.NotNil(apiStatus)
	suite.NoError(errWithCode)

	// Check response against input form data.
	suite.Equal(form.Status, apiStatus.Text)
	suite.Equal(form.SpoilerText, apiStatus.SpoilerText)
	suite.Equal(form.Sensitive, apiStatus.Sensitive)
	suite.Equal(form.Language, *apiStatus.Language)
	suite.NotEqual(util.FormatISO8601(status.EditedAt), *apiStatus.EditedAt)
	suite.Equal(form.MediaIDs, xslices.Gather(nil, apiStatus.MediaAttachments, func(media *apimodel.Attachment) string {
		return media.ID
	}))

	// Fetched the latest version of edited status from the database.
	latestStatus, err := suite.state.DB.GetStatusByID(ctx, status.ID)
	suite.NoError(err)

	// Check latest status against input form data.
	suite.Equal(form.Status, latestStatus.Text)
	suite.Equal(form.SpoilerText, latestStatus.ContentWarning)
	suite.Equal(form.Sensitive, *latestStatus.Sensitive)
	suite.Equal(form.Language, latestStatus.Language)
	suite.Equal(len(status.EditIDs)+1, len(latestStatus.EditIDs))
	suite.NotEqual(status.UpdatedAt(), latestStatus.UpdatedAt())
	suite.Equal(form.MediaIDs, latestStatus.AttachmentIDs)

	// Populate all historical edits for this status.
	err = suite.state.DB.PopulateStatusEdits(ctx, latestStatus)
	suite.NoError(err)

	// Check previous status edit matches original status content.
	previousEdit := latestStatus.Edits[len(latestStatus.Edits)-1]
	suite.Equal(status.Content, previousEdit.Content)
	suite.Equal(status.Text, previousEdit.Text)
	suite.Equal(status.ContentWarning, previousEdit.ContentWarning)
	suite.Equal(*status.Sensitive, *previousEdit.Sensitive)
	suite.Equal(status.Language, previousEdit.Language)
	suite.Equal(status.UpdatedAt(), previousEdit.CreatedAt)
	suite.Equal(status.AttachmentIDs, previousEdit.AttachmentIDs)
}

func (suite *StatusEditTestSuite) TestEditOthersStatus1() {
	// Create cancellable context to use for test.
	ctx, cncl := context.WithCancel(context.Background())
	defer cncl()

	// Get a local account to use as test requester.
	requester := suite.testAccounts["local_account_1"]
	requester, _ = suite.state.DB.GetAccountByID(ctx, requester.ID)

	// Get remote accounts's status to attempt an edit on.
	status := suite.testStatuses["remote_account_1_status_1"]
	status, _ = suite.state.DB.GetStatusByID(ctx, status.ID)

	// Prepare an empty request form, this
	// should be all we need to trigger it.
	form := &apimodel.StatusEditRequest{}

	// Attempt to edit other remote account's status, this should return an error.
	apiStatus, errWithCode := suite.status.Edit(ctx, requester, status.ID, form)
	suite.Nil(apiStatus)
	suite.Equal(http.StatusNotFound, errWithCode.Code())
	suite.Equal("status does not belong to requester", errWithCode.Error())
	suite.Equal("Not Found: target status not found", errWithCode.Safe())
}

func (suite *StatusEditTestSuite) TestEditOthersStatus2() {
	// Create cancellable context to use for test.
	ctx, cncl := context.WithCancel(context.Background())
	defer cncl()

	// Get a local account to use as test requester.
	requester := suite.testAccounts["local_account_1"]
	requester, _ = suite.state.DB.GetAccountByID(ctx, requester.ID)

	// Get other local accounts's status to attempt edit on.
	status := suite.testStatuses["local_account_2_status_1"]
	status, _ = suite.state.DB.GetStatusByID(ctx, status.ID)

	// Prepare an empty request form, this
	// should be all we need to trigger it.
	form := &apimodel.StatusEditRequest{}

	// Attempt to edit other local account's status, this should return an error.
	apiStatus, errWithCode := suite.status.Edit(ctx, requester, status.ID, form)
	suite.Nil(apiStatus)
	suite.Equal(http.StatusNotFound, errWithCode.Code())
	suite.Equal("status does not belong to requester", errWithCode.Error())
	suite.Equal("Not Found: target status not found", errWithCode.Safe())
}

func TestStatusEditTestSuite(t *testing.T) {
	suite.Run(t, new(StatusEditTestSuite))
}

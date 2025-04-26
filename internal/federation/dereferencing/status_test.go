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

package dereferencing_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"code.superseriousbusiness.org/activity/streams"
	"code.superseriousbusiness.org/gotosocial/internal/ap"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/federation/dereferencing"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/util"
	"code.superseriousbusiness.org/gotosocial/testrig"
	"github.com/stretchr/testify/suite"
)

// instantFreshness is the shortest possible freshness window.
var instantFreshness = util.Ptr(dereferencing.FreshnessWindow(0))

type StatusTestSuite struct {
	DereferencerStandardTestSuite
}

func (suite *StatusTestSuite) TestDereferenceSimpleStatus() {
	fetchingAccount := suite.testAccounts["local_account_1"]

	statusURL := testrig.URLMustParse("https://unknown-instance.com/users/brand_new_person/statuses/01FE4NTHKWW7THT67EF10EB839")
	status, _, err := suite.dereferencer.GetStatusByURI(context.Background(), fetchingAccount.Username, statusURL)
	suite.NoError(err)
	suite.NotNil(status)

	// status values should be set
	suite.Equal("https://unknown-instance.com/users/brand_new_person/statuses/01FE4NTHKWW7THT67EF10EB839", status.URI)
	suite.Equal("https://unknown-instance.com/users/@brand_new_person/01FE4NTHKWW7THT67EF10EB839", status.URL)
	suite.Equal("Hello world!", status.Content)
	suite.Equal("https://unknown-instance.com/users/brand_new_person", status.AccountURI)
	suite.False(*status.Local)
	suite.Empty(status.ContentWarning)
	suite.Equal(gtsmodel.VisibilityPublic, status.Visibility)
	suite.Equal(ap.ObjectNote, status.ActivityStreamsType)

	// status should be in the database
	dbStatus, err := suite.db.GetStatusByURI(context.Background(), status.URI)
	suite.NoError(err)
	suite.Equal(status.ID, dbStatus.ID)
	suite.True(*dbStatus.Federated)

	// account should be in the database now too
	account, err := suite.db.GetAccountByURI(context.Background(), status.AccountURI)
	suite.NoError(err)
	suite.NotNil(account)
	suite.True(*account.Discoverable)
	suite.Equal("https://unknown-instance.com/users/brand_new_person", account.URI)
	suite.Equal("hey I'm a new person, your instance hasn't seen me yet uwu", account.Note)
	suite.Equal("Geoff Brando New Personson", account.DisplayName)
	suite.Equal("brand_new_person", account.Username)
	suite.NotNil(account.PublicKey)
	suite.Nil(account.PrivateKey)
}

func (suite *StatusTestSuite) TestDereferenceStatusWithMention() {
	fetchingAccount := suite.testAccounts["local_account_1"]

	statusURL := testrig.URLMustParse("https://unknown-instance.com/users/brand_new_person/statuses/01FE5Y30E3W4P7TRE0R98KAYQV")
	status, _, err := suite.dereferencer.GetStatusByURI(context.Background(), fetchingAccount.Username, statusURL)
	suite.NoError(err)
	suite.NotNil(status)

	// status values should be set
	suite.Equal("https://unknown-instance.com/users/brand_new_person/statuses/01FE5Y30E3W4P7TRE0R98KAYQV", status.URI)
	suite.Equal("https://unknown-instance.com/users/@brand_new_person/01FE5Y30E3W4P7TRE0R98KAYQV", status.URL)
	suite.Equal("Hey @the_mighty_zork@localhost:8080 how's it going?", status.Content)
	suite.Equal("https://unknown-instance.com/users/brand_new_person", status.AccountURI)
	suite.False(*status.Local)
	suite.Empty(status.ContentWarning)
	suite.Equal(gtsmodel.VisibilityPublic, status.Visibility)
	suite.Equal(ap.ObjectNote, status.ActivityStreamsType)

	// status should be in the database
	dbStatus, err := suite.db.GetStatusByURI(context.Background(), status.URI)
	suite.NoError(err)
	suite.Equal(status.ID, dbStatus.ID)
	suite.True(*dbStatus.Federated)

	// account should be in the database now too
	account, err := suite.db.GetAccountByURI(context.Background(), status.AccountURI)
	suite.NoError(err)
	suite.NotNil(account)
	suite.True(*account.Discoverable)
	suite.Equal("https://unknown-instance.com/users/brand_new_person", account.URI)
	suite.Equal("hey I'm a new person, your instance hasn't seen me yet uwu", account.Note)
	suite.Equal("Geoff Brando New Personson", account.DisplayName)
	suite.Equal("brand_new_person", account.Username)
	suite.NotNil(account.PublicKey)
	suite.Nil(account.PrivateKey)

	// we should have a mention in the database
	m := &gtsmodel.Mention{}
	err = suite.db.GetWhere(context.Background(), []db.Where{{Key: "status_id", Value: status.ID}}, m)
	suite.NoError(err)
	suite.NotNil(m)
	suite.Equal(status.ID, m.StatusID)
	suite.Equal(account.ID, m.OriginAccountID)
	suite.Equal(fetchingAccount.ID, m.TargetAccountID)
	suite.Equal(account.URI, m.OriginAccountURI)
	suite.False(*m.Silent)
}

func (suite *StatusTestSuite) TestDereferenceStatusWithTag() {
	fetchingAccount := suite.testAccounts["local_account_1"]

	statusURL := testrig.URLMustParse("https://unknown-instance.com/users/brand_new_person/statuses/01H641QSRS3TCXSVC10X4GPKW7")
	status, _, err := suite.dereferencer.GetStatusByURI(context.Background(), fetchingAccount.Username, statusURL)
	suite.NoError(err)
	suite.NotNil(status)

	// status values should be set
	suite.Equal("https://unknown-instance.com/users/brand_new_person/statuses/01H641QSRS3TCXSVC10X4GPKW7", status.URI)
	suite.Equal("https://unknown-instance.com/users/@brand_new_person/01H641QSRS3TCXSVC10X4GPKW7", status.URL)
	suite.Equal("<p>Babe are you okay, you've hardly touched your <a href=\"https://unknown-instance.com/tags/piss\" class=\"mention hashtag\" rel=\"tag nofollow noreferrer noopener\" target=\"_blank\">#<span>piss</span></a></p>", status.Content)
	suite.Equal("https://unknown-instance.com/users/brand_new_person", status.AccountURI)
	suite.False(*status.Local)
	suite.Empty(status.ContentWarning)
	suite.Equal(gtsmodel.VisibilityPublic, status.Visibility)
	suite.Equal(ap.ObjectNote, status.ActivityStreamsType)

	// Ensure tags set + ID'd.
	suite.Len(status.Tags, 1)
	suite.Len(status.TagIDs, 1)

	// status should be in the database
	dbStatus, err := suite.db.GetStatusByURI(context.Background(), status.URI)
	suite.NoError(err)
	suite.Equal(status.ID, dbStatus.ID)
	suite.True(*dbStatus.Federated)

	// account should be in the database now too
	account, err := suite.db.GetAccountByURI(context.Background(), status.AccountURI)
	suite.NoError(err)
	suite.NotNil(account)
	suite.True(*account.Discoverable)
	suite.Equal("https://unknown-instance.com/users/brand_new_person", account.URI)
	suite.Equal("hey I'm a new person, your instance hasn't seen me yet uwu", account.Note)
	suite.Equal("Geoff Brando New Personson", account.DisplayName)
	suite.Equal("brand_new_person", account.Username)
	suite.NotNil(account.PublicKey)
	suite.Nil(account.PrivateKey)

	// we should have a tag in the database
	t := &gtsmodel.Tag{}
	err = suite.db.GetWhere(context.Background(), []db.Where{{Key: "name", Value: "piss"}}, t)
	suite.NoError(err)
	suite.NotNil(t)
}

func (suite *StatusTestSuite) TestDereferenceStatusWithImageAndNoContent() {
	fetchingAccount := suite.testAccounts["local_account_1"]

	statusURL := testrig.URLMustParse("https://turnip.farm/users/turniplover6969/statuses/70c53e54-3146-42d5-a630-83c8b6c7c042")
	status, _, err := suite.dereferencer.GetStatusByURI(context.Background(), fetchingAccount.Username, statusURL)
	suite.NoError(err)
	suite.NotNil(status)

	// status values should be set
	suite.Equal("https://turnip.farm/users/turniplover6969/statuses/70c53e54-3146-42d5-a630-83c8b6c7c042", status.URI)
	suite.Equal("https://turnip.farm/@turniplover6969/70c53e54-3146-42d5-a630-83c8b6c7c042", status.URL)
	suite.Equal("", status.Content)
	suite.Equal("https://turnip.farm/users/turniplover6969", status.AccountURI)
	suite.False(*status.Local)
	suite.Empty(status.ContentWarning)
	suite.Equal(gtsmodel.VisibilityPublic, status.Visibility)
	suite.Equal(ap.ObjectNote, status.ActivityStreamsType)

	// status should be in the database
	dbStatus, err := suite.db.GetStatusByURI(context.Background(), status.URI)
	suite.NoError(err)
	suite.Equal(status.ID, dbStatus.ID)
	suite.True(*dbStatus.Federated)

	// account should be in the database now too
	account, err := suite.db.GetAccountByURI(context.Background(), status.AccountURI)
	suite.NoError(err)
	suite.NotNil(account)
	suite.True(*account.Discoverable)
	suite.Equal("https://turnip.farm/users/turniplover6969", account.URI)
	suite.Equal("I just think they're neat", account.Note)
	suite.Equal("Turnip Lover 6969", account.DisplayName)
	suite.Equal("turniplover6969", account.Username)
	suite.NotNil(account.PublicKey)
	suite.Nil(account.PrivateKey)

	// we should have an attachment in the database
	a := &gtsmodel.MediaAttachment{}
	err = suite.db.GetWhere(context.Background(), []db.Where{{Key: "status_id", Value: status.ID}}, a)
	suite.NoError(err)
}

func (suite *StatusTestSuite) TestDereferenceStatusWithNonMatchingURI() {
	fetchingAccount := suite.testAccounts["local_account_1"]

	const (
		remoteURI    = "https://turnip.farm/users/turniplover6969/statuses/70c53e54-3146-42d5-a630-83c8b6c7c042"
		remoteAltURI = "https://turnip.farm/users/turniphater420/statuses/70c53e54-3146-42d5-a630-83c8b6c7c042"
	)

	// Create a copy of this remote account at alternative URI.
	remoteStatus := suite.client.TestRemoteStatuses[remoteURI]
	suite.client.TestRemoteStatuses[remoteAltURI] = remoteStatus

	// Attempt to fetch account at alternative URI, it should fail!
	fetchedStatus, _, err := suite.dereferencer.GetStatusByURI(
		context.Background(),
		fetchingAccount.Username,
		testrig.URLMustParse(remoteAltURI),
	)
	suite.Equal(err.Error(), fmt.Sprintf("enrichStatus: dereferenced status uri %s does not match %s", remoteURI, remoteAltURI))
	suite.Nil(fetchedStatus)
}

func (suite *StatusTestSuite) TestDereferencerRefreshStatusUpdated() {
	// Create a new context for this test.
	ctx, cncl := context.WithCancel(context.Background())
	defer cncl()

	// The local account we will be fetching statuses as.
	fetchingAccount := suite.testAccounts["local_account_1"]

	// The test status in question that we will be dereferencing from "remote".
	testURIStr := "https://unknown-instance.com/users/brand_new_person/statuses/01FE4NTHKWW7THT67EF10EB839"
	testURI := testrig.URLMustParse(testURIStr)
	testStatusable := suite.client.TestRemoteStatuses[testURIStr]

	// Fetch the remote status first to load it into instance.
	testStatus, statusable, err := suite.dereferencer.GetStatusByURI(ctx,
		fetchingAccount.Username,
		testURI,
	)
	suite.NotNil(statusable)
	suite.NoError(err)

	// Run through multiple possible edits.
	for _, testCase := range []struct {
		editedContent        string
		editedContentWarning string
		editedLanguage       string
		editedSensitive      bool
		editedAttachmentIDs  []string
		editedPollOptions    []string
		editedPollVotes      []int
		editedAt             time.Time
	}{
		{
			editedContent:        "updated status content!",
			editedContentWarning: "CW: edited status content",
			editedLanguage:       testStatus.Language,        // no change
			editedSensitive:      *testStatus.Sensitive,      // no change
			editedAttachmentIDs:  testStatus.AttachmentIDs,   // no change
			editedPollOptions:    getPollOptions(testStatus), // no change
			editedPollVotes:      getPollVotes(testStatus),   // no change
			editedAt:             time.Now(),
		},
	} {
		// Take a snapshot of current
		// state of the test status.
		testStatus = copyStatus(testStatus)

		// Edit the "remote" statusable obj.
		suite.editStatusable(testStatusable,
			testCase.editedContent,
			testCase.editedContentWarning,
			testCase.editedLanguage,
			testCase.editedSensitive,
			testCase.editedAttachmentIDs,
			testCase.editedPollOptions,
			testCase.editedPollVotes,
			testCase.editedAt,
		)

		// Refresh with a given statusable to updated to edited copy.
		latest, statusable, err := suite.dereferencer.RefreshStatus(ctx,
			fetchingAccount.Username,
			testStatus,
			nil, // NOTE: can provide testStatusable here to test as being received (not deref'd)
			instantFreshness,
		)
		suite.NotNil(statusable)
		suite.NoError(err)

		// verify updated status details.
		suite.verifyEditedStatusUpdate(

			// the original status
			// before any changes.
			testStatus,

			// latest status
			// being tested.
			latest,

			// expected current state.
			&gtsmodel.StatusEdit{
				Content:        testCase.editedContent,
				ContentWarning: testCase.editedContentWarning,
				Language:       testCase.editedLanguage,
				Sensitive:      &testCase.editedSensitive,
				AttachmentIDs:  testCase.editedAttachmentIDs,
				PollOptions:    testCase.editedPollOptions,
				PollVotes:      testCase.editedPollVotes,
				// createdAt never changes
			},

			// expected historic edit.
			&gtsmodel.StatusEdit{
				Content:        testStatus.Content,
				ContentWarning: testStatus.ContentWarning,
				Language:       testStatus.Language,
				Sensitive:      testStatus.Sensitive,
				AttachmentIDs:  testStatus.AttachmentIDs,
				PollOptions:    getPollOptions(testStatus),
				PollVotes:      getPollVotes(testStatus),
				CreatedAt:      testStatus.UpdatedAt(),
			},
		)
	}
}

// editStatusable updates the given statusable attributes.
// note that this acts on the original object, no copying.
func (suite *StatusTestSuite) editStatusable(
	statusable ap.Statusable,
	content string,
	contentWarning string,
	language string,
	sensitive bool,
	attachmentIDs []string, // TODO: this will require some thinking as to how ...
	pollOptions []string, // TODO: this will require changing statusable type to question
	pollVotes []int, // TODO: this will require changing statusable type to question
	editedAt time.Time,
) {
	// simply reset all mentions / emojis / tags
	statusable.SetActivityStreamsTag(nil)

	// Update the statusable content property + language (if set).
	contentProp := streams.NewActivityStreamsContentProperty()
	statusable.SetActivityStreamsContent(contentProp)
	contentProp.AppendXMLSchemaString(content)
	if language != "" {
		contentProp.AppendRDFLangString(map[string]string{
			language: content,
		})
	}

	// Update the statusable content-warning property.
	summaryProp := streams.NewActivityStreamsSummaryProperty()
	statusable.SetActivityStreamsSummary(summaryProp)
	summaryProp.AppendXMLSchemaString(contentWarning)

	// Update the statusable sensitive property.
	sensitiveProp := streams.NewActivityStreamsSensitiveProperty()
	statusable.SetActivityStreamsSensitive(sensitiveProp)
	sensitiveProp.AppendXMLSchemaBoolean(sensitive)

	// Update the statusable updated property.
	ap.SetUpdated(statusable, editedAt)
}

// verifyEditedStatusUpdate verifies that a given status has
// the expected number of historic edits, the 'current' status
// attributes (encapsulated as an edit for minimized no. args),
// and the last given 'historic' status edit attributes.
func (suite *StatusTestSuite) verifyEditedStatusUpdate(
	testStatus *gtsmodel.Status, // the original model
	status *gtsmodel.Status, // the status to check
	current *gtsmodel.StatusEdit, // expected current state
	historic *gtsmodel.StatusEdit, // historic edit we expect to have
) {
	// don't use this func
	// name in error msgs.
	suite.T().Helper()

	// Check we have expected number of edits.
	previousEdits := len(testStatus.Edits)
	suite.Len(status.Edits, previousEdits+1)
	suite.Len(status.EditIDs, previousEdits+1)

	// Check current state of status.
	suite.Equal(current.Content, status.Content)
	suite.Equal(current.ContentWarning, status.ContentWarning)
	suite.Equal(current.Language, status.Language)
	suite.Equal(*current.Sensitive, *status.Sensitive)
	suite.Equal(current.AttachmentIDs, status.AttachmentIDs)
	suite.Equal(current.PollOptions, getPollOptions(status))
	suite.Equal(current.PollVotes, getPollVotes(status))

	// Check the latest historic edit matches expected.
	latestEdit := status.Edits[len(status.Edits)-1]
	suite.Equal(historic.Content, latestEdit.Content)
	suite.Equal(historic.ContentWarning, latestEdit.ContentWarning)
	suite.Equal(historic.Language, latestEdit.Language)
	suite.Equal(*historic.Sensitive, *latestEdit.Sensitive)
	suite.Equal(historic.AttachmentIDs, latestEdit.AttachmentIDs)
	suite.Equal(historic.PollOptions, latestEdit.PollOptions)
	suite.Equal(historic.PollVotes, latestEdit.PollVotes)
	suite.Equal(historic.CreatedAt, latestEdit.CreatedAt)

	// The status creation date should never change.
	suite.Equal(testStatus.CreatedAt, status.CreatedAt)
}

func TestStatusTestSuite(t *testing.T) {
	suite.Run(t, new(StatusTestSuite))
}

// copyStatus returns a copy of the given status model (not including sub-structs).
func copyStatus(status *gtsmodel.Status) *gtsmodel.Status {
	copy := new(gtsmodel.Status)
	*copy = *status
	return copy
}

// getPollOptions extracts poll option strings from status (if poll is set).
func getPollOptions(status *gtsmodel.Status) []string {
	if status.Poll != nil {
		return status.Poll.Options
	}
	return nil
}

// getPollVotes extracts poll vote counts from status (if poll is set).
func getPollVotes(status *gtsmodel.Status) []int {
	if status.Poll != nil {
		return status.Poll.Votes
	}
	return nil
}

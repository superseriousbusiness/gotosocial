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

package status

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/superseriousbusiness/gotosocial/internal/ap"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	"github.com/superseriousbusiness/gotosocial/internal/util/xslices"
)

// Edit ...
func (p *Processor) Edit(
	ctx context.Context,
	requester *gtsmodel.Account,
	statusID string,
	form *apimodel.StatusEditRequest,
) (
	*apimodel.Status,
	gtserror.WithCode,
) {
	// Fetch status and ensure it's owned by requesting account.
	status, errWithCode := p.c.GetOwnStatus(ctx, requester, statusID)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// Ensure this isn't a boost.
	if status.BoostOfID != "" {
		return nil, gtserror.NewErrorNotFound(
			errors.New("status is a boost wrapper"),
			"target status not found",
		)
	}

	// Ensure account populated; we'll need their settings.
	if err := p.state.DB.PopulateAccount(ctx, requester); err != nil {
		log.Errorf(ctx, "error(s) populating account, will continue: %s", err)
	}

	// We need the status populated including all historical edits.
	if err := p.state.DB.PopulateStatusEdits(ctx, status); err != nil {
		err := gtserror.Newf("error getting status edits from db: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Time of edit.
	now := time.Now()

	// Validate incoming form edit content.
	if errWithCode := validateStatusContent(
		form.Status,
		form.SpoilerText,
		form.MediaIDs,
		form.Poll,
	); errWithCode != nil {
		return nil, errWithCode
	}

	// Process incoming status edit content fields.
	content, errWithCode := p.processContent(ctx,
		requester,
		statusID,
		string(form.ContentType),
		form.Status,
		form.SpoilerText,
		form.Language,
		form.Poll,
	)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// Process new status attachments to use.
	media, errWithCode := p.processMedia(ctx,
		requester.ID,
		statusID,
		form.MediaIDs,
	)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// Process incoming edits of any attached media.
	mediaEdited, errWithCode := p.processMediaEdits(ctx,
		media,
		form.MediaAttributes,
	)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// Process incoming edits of any attached status poll.
	poll, pollEdited, errWithCode := p.processPollEdit(ctx,
		statusID,
		status.Poll,
		form.Poll,
		now,
	)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// Check if new status poll was set.
	pollChanged := (poll != status.Poll)

	// Determine whether there were any changes possibly
	// causing a change to embedded mentions, tags, emojis.
	contentChanged := (status.Content != content.Content)
	warningChanged := (status.ContentWarning != content.ContentWarning)
	languageChanged := (status.Language != content.Language)
	anyContentChanged := contentChanged || warningChanged ||
		pollEdited // encapsulates pollChanged too

	// Check if status media attachments have changed.
	mediaChanged := !slices.Equal(status.AttachmentIDs,
		form.MediaIDs,
	)

	// Track status columns we
	// need to update in database.
	cols := make([]string, 2, 13)
	cols[0] = "edited_at"
	cols[1] = "edits"

	if contentChanged {
		// Update status text.
		//
		// Note we don't update these
		// status fields right away so
		// we can save current version.
		cols = append(cols, "content")
		cols = append(cols, "text")
	}

	if warningChanged {
		// Update status content warning.
		//
		// Note we don't update these
		// status fields right away so
		// we can save current version.
		cols = append(cols, "content_warning")
	}

	if languageChanged {
		// Update status language pref.
		//
		// Note we don't update these
		// status fields right away so
		// we can save current version.
		cols = append(cols, "language")
	}

	if *status.Sensitive != form.Sensitive {
		// Update status sensitivity pref.
		//
		// Note we don't update these
		// status fields right away so
		// we can save current version.
		cols = append(cols, "sensitive")
	}

	if mediaChanged {
		// Updated status media attachments.
		//
		// Note we don't update these
		// status fields right away so
		// we can save current version.
		cols = append(cols, "attachments")
	}

	if pollChanged {
		// Updated attached status poll.
		//
		// Note we don't update these
		// status fields right away so
		// we can save current version.
		cols = append(cols, "poll_id")

		if status.Poll == nil || poll == nil {
			// Went from with-poll to without-poll
			// or vice-versa. This changes AP type.
			cols = append(cols, "activity_streams_type")
		}
	}

	if anyContentChanged {
		if !slices.Equal(status.MentionIDs, content.MentionIDs) {
			// Update attached status mentions.
			cols = append(cols, "mentions")
			status.MentionIDs = content.MentionIDs
			status.Mentions = content.Mentions
		}

		if !slices.Equal(status.TagIDs, content.TagIDs) {
			// Updated attached status tags.
			cols = append(cols, "tags")
			status.TagIDs = content.TagIDs
			status.Tags = content.Tags
		}

		if !slices.Equal(status.EmojiIDs, content.EmojiIDs) {
			// We specifically store both *new* AND *old* edit
			// revision emojis in the statuses.emojis column.
			emojiByID := func(e *gtsmodel.Emoji) string { return e.ID }
			status.Emojis = append(status.Emojis, content.Emojis...)
			status.Emojis = xslices.DeduplicateFunc(status.Emojis, emojiByID)
			status.EmojiIDs = xslices.Gather(status.EmojiIDs[:0], status.Emojis, emojiByID)

			// Update attached status emojis.
			cols = append(cols, "emojis")
		}
	}

	// If no status columns were updated, no media and
	// no poll were edited, there's nothing to do!
	if len(cols) == 2 && !mediaEdited && !pollEdited {
		const text = "status was not changed"
		return nil, gtserror.NewErrorUnprocessableEntity(
			errors.New(text),
			text,
		)
	}

	// Create an edit to store a
	// historical snapshot of status.
	var edit gtsmodel.StatusEdit
	edit.ID = id.NewULIDFromTime(now)
	edit.Content = status.Content
	edit.ContentWarning = status.ContentWarning
	edit.Text = status.Text
	edit.Language = status.Language
	edit.Sensitive = status.Sensitive
	edit.StatusID = status.ID
	edit.CreatedAt = status.UpdatedAt()

	// Copy existing media and descriptions.
	edit.AttachmentIDs = status.AttachmentIDs
	if l := len(status.Attachments); l > 0 {
		edit.AttachmentDescriptions = make([]string, l)
		for i, attach := range status.Attachments {
			edit.AttachmentDescriptions[i] = attach.Description
		}
	}

	if status.Poll != nil {
		// Poll only set if existed previously.
		edit.PollOptions = status.Poll.Options

		if pollChanged || !*status.Poll.HideCounts ||
			!status.Poll.ClosedAt.IsZero() {
			// If the counts are allowed to be
			// shown, or poll has changed, then
			// include poll vote counts in edit.
			edit.PollVotes = status.Poll.Votes
		}
	}

	// Insert this new edit of existing status into database.
	if err := p.state.DB.PutStatusEdit(ctx, &edit); err != nil {
		err := gtserror.Newf("error putting edit in database: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Add edit to list of edits on the status.
	status.EditIDs = append(status.EditIDs, edit.ID)
	status.Edits = append(status.Edits, &edit)

	// Now historical status data is stored,
	// update the other necessary status fields.
	status.Content = content.Content
	status.ContentWarning = content.ContentWarning
	status.Text = form.Status
	status.Language = content.Language
	status.Sensitive = &form.Sensitive
	status.AttachmentIDs = form.MediaIDs
	status.Attachments = media
	status.EditedAt = now

	if poll != nil {
		// Set relevent fields for latest with poll.
		status.ActivityStreamsType = ap.ActivityQuestion
		status.PollID = poll.ID
		status.Poll = poll
	} else {
		// Set relevant fields for latest without poll.
		status.ActivityStreamsType = ap.ObjectNote
		status.PollID = ""
		status.Poll = nil
	}

	// Finally update the existing status model in the database.
	if err := p.state.DB.UpdateStatus(ctx, status, cols...); err != nil {
		err := gtserror.Newf("error updating status in db: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if pollChanged && status.Poll != nil && !status.Poll.ExpiresAt.IsZero() {
		// Now the status is updated, attempt to schedule
		// an expiry handler for the changed status poll.
		if err := p.polls.ScheduleExpiry(ctx, status.Poll); err != nil {
			log.Errorf(ctx, "error scheduling poll expiry: %v", err)
		}
	}

	// Send it to the client API worker for async side-effects.
	p.state.Workers.Client.Queue.Push(&messages.FromClientAPI{
		APObjectType:   ap.ObjectNote,
		APActivityType: ap.ActivityUpdate,
		GTSModel:       status,
		Origin:         requester,
	})

	// Return an API model of the updated status.
	return p.c.GetAPIStatus(ctx, requester, status)
}

// HistoryGet gets edit history for the target status, taking account of privacy settings and blocks etc.
func (p *Processor) HistoryGet(ctx context.Context, requester *gtsmodel.Account, targetStatusID string) ([]*apimodel.StatusEdit, gtserror.WithCode) {
	target, errWithCode := p.c.GetVisibleTargetStatus(ctx,
		requester,
		targetStatusID,
		nil, // default freshness
	)
	if errWithCode != nil {
		return nil, errWithCode
	}

	if err := p.state.DB.PopulateStatusEdits(ctx, target); err != nil {
		err := gtserror.Newf("error getting status edits from db: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	edits, err := p.converter.StatusToAPIEdits(ctx, target)
	if err != nil {
		err := gtserror.Newf("error converting status edits: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return edits, nil
}

func (p *Processor) processMediaEdits(
	ctx context.Context,
	attachs []*gtsmodel.MediaAttachment,
	attrs []apimodel.AttachmentAttributesRequest,
) (
	bool,
	gtserror.WithCode,
) {
	var edited bool

	for _, attr := range attrs {
		// Search the media attachments slice for index of media with attr.ID.
		i := slices.IndexFunc(attachs, func(m *gtsmodel.MediaAttachment) bool {
			return m.ID == attr.ID
		})
		if i == -1 {
			text := fmt.Sprintf("media not found: %s", attr.ID)
			return false, gtserror.NewErrorBadRequest(errors.New(text), text)
		}

		// Get attach at index.
		attach := attachs[i]

		// Track which columns need
		// updating in database query.
		cols := make([]string, 0, 2)

		// Check for description change.
		if attr.Description != attach.Description {
			attach.Description = attr.Description
			cols = append(cols, "description")
		}

		if attr.Focus != "" {
			// Parse provided media focus parameters from string.
			fx, fy, errWithCode := apiutil.ParseFocus(attr.Focus)
			if errWithCode != nil {
				return false, errWithCode
			}

			// Check for change in focus coords.
			if attach.FileMeta.Focus.X != fx ||
				attach.FileMeta.Focus.Y != fy {
				attach.FileMeta.Focus.X = fx
				attach.FileMeta.Focus.Y = fy
				cols = append(cols, "focus_x", "focus_y")
			}
		}

		if len(cols) > 0 {
			// Media attachment was changed, update this in database.
			err := p.state.DB.UpdateAttachment(ctx, attach, cols...)
			if err != nil {
				err := gtserror.Newf("error updating attachment in db: %w", err)
				return false, gtserror.NewErrorInternalError(err)
			}

			// Set edited.
			edited = true
		}
	}

	return edited, nil
}

func (p *Processor) processPollEdit(
	ctx context.Context,
	statusID string,
	original *gtsmodel.Poll,
	form *apimodel.PollRequest,
	now time.Time, // used for expiry time
) (
	*gtsmodel.Poll,
	bool,
	gtserror.WithCode,
) {
	if form == nil {
		if original != nil {
			// No poll was given but there's an existing poll,
			// this indicates the original needs to be deleted.
			if err := p.deletePoll(ctx, original); err != nil {
				return nil, true, gtserror.NewErrorInternalError(err)
			}

			// Existing was deleted.
			return nil, true, nil
		}

		// No change in poll.
		return nil, false, nil
	}

	switch {
	// No existing poll.
	case original == nil:

	// Any change that effects voting, i.e. options, allow multiple
	// or re-opening a closed poll requires deleting the existing poll.
	case !slices.Equal(form.Options, original.Options) ||
		(form.Multiple != *original.Multiple) ||
		(!original.ClosedAt.IsZero() && form.ExpiresIn != 0):
		if err := p.deletePoll(ctx, original); err != nil {
			return nil, true, gtserror.NewErrorInternalError(err)
		}

	// Any other changes only require a model
	// update, and at-most a new expiry handler.
	default:
		var cols []string

		// Check if the hide counts field changed.
		if form.HideTotals != *original.HideCounts {
			cols = append(cols, "hide_counts")
			original.HideCounts = &form.HideTotals
		}

		var expiresAt time.Time

		// Determine expiry time if given.
		if in := form.ExpiresIn; in > 0 {
			expiresIn := time.Duration(in)
			expiresAt = now.Add(expiresIn * time.Second)
		}

		// Check for expiry time.
		if !expiresAt.IsZero() {

			if !original.ExpiresAt.IsZero() {
				// Existing had expiry, cancel scheduled handler.
				_ = p.state.Workers.Scheduler.Cancel(original.ID)
			}

			// Since expiry is given as a duration
			// we always treat > 0 as a change as
			// we can't know otherwise unfortunately.
			cols = append(cols, "expires_at")
			original.ExpiresAt = expiresAt
		}

		if len(cols) == 0 {
			// Were no changes to poll.
			return original, false, nil
		}

		// Update the original poll model in the database with these columns.
		if err := p.state.DB.UpdatePoll(ctx, original, cols...); err != nil {
			err := gtserror.Newf("error updating poll.expires_at in db: %w", err)
			return nil, true, gtserror.NewErrorInternalError(err)
		}

		if !expiresAt.IsZero() {
			// Updated poll has an expiry, schedule a new expiry handler.
			if err := p.polls.ScheduleExpiry(ctx, original); err != nil {
				log.Errorf(ctx, "error scheduling poll expiry: %v", err)
			}
		}

		// Existing poll was updated.
		return original, true, nil
	}

	// If we reached here then an entirely
	// new status poll needs to be created.
	poll, errWithCode := p.processPoll(ctx,
		statusID,
		form,
		now,
	)
	return poll, true, errWithCode
}

func (p *Processor) deletePoll(ctx context.Context, poll *gtsmodel.Poll) error {
	if !poll.ExpiresAt.IsZero() && !poll.ClosedAt.IsZero() {
		// Poll has an expiry and has not yet closed,
		// cancel any expiry handler before deletion.
		_ = p.state.Workers.Scheduler.Cancel(poll.ID)
	}

	// Delete the given poll from the database.
	err := p.state.DB.DeletePollByID(ctx, poll.ID)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return gtserror.Newf("error deleting poll from db: %w", err)
	}

	return nil
}

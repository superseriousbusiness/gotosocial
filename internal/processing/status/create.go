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
	"time"

	"github.com/superseriousbusiness/gotosocial/internal/ap"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	"github.com/superseriousbusiness/gotosocial/internal/text"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/internal/uris"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// Create processes the given form to create a new status, returning the api model representation of that status if it's OK.
//
// Precondition: the form's fields should have already been validated and normalized by the caller.
func (p *Processor) Create(ctx context.Context, requestingAccount *gtsmodel.Account, application *gtsmodel.Application, form *apimodel.AdvancedStatusCreateForm) (*apimodel.Status, gtserror.WithCode) {
	// Generate new ID for status.
	statusID := id.NewULID()

	// Generate necessary URIs for username, to build status URIs.
	accountURIs := uris.GenerateURIsForAccount(requestingAccount.Username)

	// Get current time.
	now := time.Now()

	status := &gtsmodel.Status{
		ID:                       statusID,
		URI:                      accountURIs.StatusesURI + "/" + statusID,
		URL:                      accountURIs.StatusesURL + "/" + statusID,
		CreatedAt:                now,
		UpdatedAt:                now,
		Local:                    util.Ptr(true),
		Account:                  requestingAccount,
		AccountID:                requestingAccount.ID,
		AccountURI:               requestingAccount.URI,
		ActivityStreamsType:      ap.ObjectNote,
		Sensitive:                &form.Sensitive,
		CreatedWithApplicationID: application.ID,
		Text:                     form.Status,
	}

	if form.Poll != nil {
		// Update the status AS type to "Question".
		status.ActivityStreamsType = ap.ActivityQuestion

		// Create new poll for status from form.
		secs := time.Duration(form.Poll.ExpiresIn)
		status.Poll = &gtsmodel.Poll{
			ID:         id.NewULID(),
			Multiple:   &form.Poll.Multiple,
			HideCounts: &form.Poll.HideTotals,
			Options:    form.Poll.Options,
			StatusID:   statusID,
			Status:     status,
			ExpiresAt:  now.Add(secs * time.Second),
		}

		// Set poll ID on the status.
		status.PollID = status.Poll.ID
	}

	if errWithCode := p.processReplyToID(ctx, form, requestingAccount.ID, status); errWithCode != nil {
		return nil, errWithCode
	}

	if errWithCode := p.processThreadID(ctx, status); errWithCode != nil {
		return nil, errWithCode
	}

	if errWithCode := p.processMediaIDs(ctx, form, requestingAccount.ID, status); errWithCode != nil {
		return nil, errWithCode
	}

	if err := processVisibility(form, requestingAccount.Privacy, status); err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	if err := processLanguage(form, requestingAccount.Language, status); err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	if err := p.processContent(ctx, p.parseMention, form, status); err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	if status.Poll != nil {
		// Try to insert the new status poll in the database.
		if err := p.state.DB.PutPoll(ctx, status.Poll); err != nil {
			err := gtserror.Newf("error inserting poll in db: %w", err)
			return nil, gtserror.NewErrorInternalError(err)
		}
	}

	// Insert this new status in the database.
	if err := p.state.DB.PutStatus(ctx, status); err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	// send it back to the client API worker for async side-effects.
	p.state.Workers.EnqueueClientAPI(ctx, messages.FromClientAPI{
		APObjectType:   ap.ObjectNote,
		APActivityType: ap.ActivityCreate,
		GTSModel:       status,
		OriginAccount:  requestingAccount,
	})

	if status.Poll != nil {
		// Now that the status is inserted, and side effects queued,
		// attempt to schedule an expiry handler for the status poll.
		if err := p.polls.ScheduleExpiry(ctx, status.Poll); err != nil {
			err := gtserror.Newf("error scheduling poll expiry: %w", err)
			return nil, gtserror.NewErrorInternalError(err)
		}
	}

	return p.c.GetAPIStatus(ctx, requestingAccount, status)
}

func (p *Processor) processReplyToID(ctx context.Context, form *apimodel.AdvancedStatusCreateForm, thisAccountID string, status *gtsmodel.Status) gtserror.WithCode {
	if form.InReplyToID == "" {
		return nil
	}

	// If this status is a reply to another status, we need to do a bit of work to establish whether or not this status can be posted:
	//
	// 1. Does the replied status exist in the database?
	// 2. Is the replied status marked as replyable?
	// 3. Does a block exist between either the current account or the account that posted the status it's replying to?
	//
	// If this is all OK, then we fetch the repliedStatus and the repliedAccount for later processing.

	inReplyTo, err := p.state.DB.GetStatusByID(ctx, form.InReplyToID)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("error fetching status %s from db: %w", form.InReplyToID, err)
		return gtserror.NewErrorInternalError(err)
	}

	if inReplyTo == nil {
		const text = "cannot reply to status that does not exist"
		return gtserror.NewErrorBadRequest(errors.New(text), text)
	}

	if !*inReplyTo.Replyable {
		text := fmt.Sprintf("status %s is marked as not replyable", form.InReplyToID)
		return gtserror.NewErrorForbidden(errors.New(text), text)
	}

	if blocked, err := p.state.DB.IsEitherBlocked(ctx, thisAccountID, inReplyTo.AccountID); err != nil {
		err := gtserror.Newf("error checking block in db: %w", err)
		return gtserror.NewErrorInternalError(err)
	} else if blocked {
		text := fmt.Sprintf("status %s is not replyable", form.InReplyToID)
		return gtserror.NewErrorNotFound(errors.New(text), text)
	}

	// Set status fields from inReplyTo.
	status.InReplyToID = inReplyTo.ID
	status.InReplyTo = inReplyTo
	status.InReplyToURI = inReplyTo.URI
	status.InReplyToAccountID = inReplyTo.AccountID

	return nil
}

func (p *Processor) processThreadID(ctx context.Context, status *gtsmodel.Status) gtserror.WithCode {
	// Status takes the thread ID of
	// whatever it replies to, if set.
	//
	// Might not be set if status is local
	// and replies to a remote status that
	// doesn't have a thread ID yet.
	//
	// If so, we can just thread from this
	// status onwards instead, since this
	// is where the relevant part of the
	// thread starts, from the perspective
	// of our instance at least.
	if status.InReplyTo != nil &&
		status.InReplyTo.ThreadID != "" {
		// Just inherit threadID from parent.
		status.ThreadID = status.InReplyTo.ThreadID
		return nil
	}

	// Mark new thread (or threaded
	// subsection) starting from here.
	threadID := id.NewULID()
	if err := p.state.DB.PutThread(
		ctx,
		&gtsmodel.Thread{
			ID: threadID,
		},
	); err != nil {
		err := gtserror.Newf("error inserting new thread in db: %w", err)
		return gtserror.NewErrorInternalError(err)
	}

	// Future replies to this status
	// (if any) will inherit this thread ID.
	status.ThreadID = threadID

	return nil
}

func (p *Processor) processMediaIDs(ctx context.Context, form *apimodel.AdvancedStatusCreateForm, thisAccountID string, status *gtsmodel.Status) gtserror.WithCode {
	if form.MediaIDs == nil {
		return nil
	}

	// Get minimum allowed char descriptions.
	minChars := config.GetMediaDescriptionMinChars()

	attachments := []*gtsmodel.MediaAttachment{}
	attachmentIDs := []string{}

	for _, mediaID := range form.MediaIDs {
		attachment, err := p.state.DB.GetAttachmentByID(ctx, mediaID)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			err := gtserror.Newf("error fetching media from db: %w", err)
			return gtserror.NewErrorInternalError(err)
		}

		if attachment == nil {
			text := fmt.Sprintf("media %s not found", mediaID)
			return gtserror.NewErrorBadRequest(errors.New(text), text)
		}

		if attachment.AccountID != thisAccountID {
			text := fmt.Sprintf("media %s does not belong to account", mediaID)
			return gtserror.NewErrorBadRequest(errors.New(text), text)
		}

		if attachment.StatusID != "" || attachment.ScheduledStatusID != "" {
			text := fmt.Sprintf("media %s already attached to status", mediaID)
			return gtserror.NewErrorBadRequest(errors.New(text), text)
		}

		if length := len([]rune(attachment.Description)); length < minChars {
			text := fmt.Sprintf("media %s description too short, at least %d required", mediaID, minChars)
			return gtserror.NewErrorBadRequest(errors.New(text), text)
		}

		attachments = append(attachments, attachment)
		attachmentIDs = append(attachmentIDs, attachment.ID)
	}

	status.Attachments = attachments
	status.AttachmentIDs = attachmentIDs
	return nil
}

func processVisibility(form *apimodel.AdvancedStatusCreateForm, accountDefaultVis gtsmodel.Visibility, status *gtsmodel.Status) error {
	// by default all flags are set to true
	federated := true
	boostable := true
	replyable := true
	likeable := true

	// If visibility isn't set on the form, then just take the account default.
	// If that's also not set, take the default for the whole instance.
	var vis gtsmodel.Visibility
	switch {
	case form.Visibility != "":
		vis = typeutils.APIVisToVis(form.Visibility)
	case accountDefaultVis != "":
		vis = accountDefaultVis
	default:
		vis = gtsmodel.VisibilityDefault
	}

	switch vis {
	case gtsmodel.VisibilityPublic:
		// for public, there's no need to change any of the advanced flags from true regardless of what the user filled out
		break
	case gtsmodel.VisibilityUnlocked:
		// for unlocked the user can set any combination of flags they like so look at them all to see if they're set and then apply them
		if form.Federated != nil {
			federated = *form.Federated
		}

		if form.Boostable != nil {
			boostable = *form.Boostable
		}

		if form.Replyable != nil {
			replyable = *form.Replyable
		}

		if form.Likeable != nil {
			likeable = *form.Likeable
		}

	case gtsmodel.VisibilityFollowersOnly, gtsmodel.VisibilityMutualsOnly:
		// for followers or mutuals only, boostable will *always* be false, but the other fields can be set so check and apply them
		boostable = false

		if form.Federated != nil {
			federated = *form.Federated
		}

		if form.Replyable != nil {
			replyable = *form.Replyable
		}

		if form.Likeable != nil {
			likeable = *form.Likeable
		}

	case gtsmodel.VisibilityDirect:
		// direct is pretty easy: there's only one possible setting so return it
		federated = true
		boostable = false
		replyable = true
		likeable = true
	}

	status.Visibility = vis
	status.Federated = &federated
	status.Boostable = &boostable
	status.Replyable = &replyable
	status.Likeable = &likeable
	return nil
}

func processLanguage(form *apimodel.AdvancedStatusCreateForm, accountDefaultLanguage string, status *gtsmodel.Status) error {
	if form.Language != "" {
		status.Language = form.Language
	} else {
		status.Language = accountDefaultLanguage
	}
	if status.Language == "" {
		return errors.New("no language given either in status create form or account default")
	}
	return nil
}

func (p *Processor) processContent(ctx context.Context, parseMention gtsmodel.ParseMentionFunc, form *apimodel.AdvancedStatusCreateForm, status *gtsmodel.Status) error {
	if form.ContentType == "" {
		// If content type wasn't specified, use the author's preferred content-type.
		contentType := apimodel.StatusContentType(status.Account.StatusContentType)
		form.ContentType = contentType
	}

	// format is the currently set text formatting
	// function, according to the provided content-type.
	var format text.FormatFunc

	// formatInput is a shorthand function to format the given input string with the
	// currently set 'formatFunc', passing in all required args and returning result.
	formatInput := func(formatFunc text.FormatFunc, input string) *text.FormatResult {
		return formatFunc(ctx, parseMention, status.AccountID, status.ID, input)
	}

	switch form.ContentType {
	// None given / set,
	// use default (plain).
	case "":
		fallthrough

	// Format status according to text/plain.
	case apimodel.StatusContentTypePlain:
		format = p.formatter.FromPlain

	// Format status according to text/markdown.
	case apimodel.StatusContentTypeMarkdown:
		format = p.formatter.FromMarkdown

	// Unknown.
	default:
		return fmt.Errorf("invalid status format: %q", form.ContentType)
	}

	// Sanitize status text and format.
	contentRes := formatInput(format, form.Status)

	// Collect formatted results.
	status.Content = contentRes.HTML
	status.Mentions = append(status.Mentions, contentRes.Mentions...)
	status.Emojis = append(status.Emojis, contentRes.Emojis...)
	status.Tags = append(status.Tags, contentRes.Tags...)

	// From here-on-out just use emoji-only
	// plain-text formatting as the FormatFunc.
	format = p.formatter.FromPlainEmojiOnly

	// Sanitize content warning and format.
	spoiler := text.SanitizeToPlaintext(form.SpoilerText)
	warningRes := formatInput(format, spoiler)

	// Collect formatted results.
	status.ContentWarning = warningRes.HTML
	status.Emojis = append(status.Emojis, warningRes.Emojis...)

	if status.Poll != nil {
		for i := range status.Poll.Options {
			// Sanitize each option title name and format.
			option := text.SanitizeToPlaintext(status.Poll.Options[i])
			optionRes := formatInput(format, option)

			// Collect each formatted result.
			status.Poll.Options[i] = optionRes.HTML
			status.Emojis = append(status.Emojis, optionRes.Emojis...)
		}
	}

	// Gather all the database IDs from each of the gathered status mentions, tags, and emojis.
	status.MentionIDs = gatherIDs(status.Mentions, func(mention *gtsmodel.Mention) string { return mention.ID })
	status.TagIDs = gatherIDs(status.Tags, func(tag *gtsmodel.Tag) string { return tag.ID })
	status.EmojiIDs = gatherIDs(status.Emojis, func(emoji *gtsmodel.Emoji) string { return emoji.ID })

	return nil
}

// gatherIDs is a small utility function to gather IDs from a slice of type T.
func gatherIDs[T any](in []T, getID func(T) string) []string {
	if getID == nil {
		// move nil check out loop.
		panic("nil getID function")
	}
	ids := make([]string, len(in))
	for i, t := range in {
		ids[i] = getID(t)
	}
	return ids
}

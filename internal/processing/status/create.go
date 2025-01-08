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
	"time"

	"github.com/superseriousbusiness/gotosocial/internal/ap"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/internal/uris"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// Create processes the given form to create a new status, returning the api model representation of that status if it's OK.
// Note this also handles validation of incoming form field data.
func (p *Processor) Create(
	ctx context.Context,
	requester *gtsmodel.Account,
	application *gtsmodel.Application,
	form *apimodel.StatusCreateRequest,
) (
	*apimodel.Status,
	gtserror.WithCode,
) {
	// Validate incoming form status content.
	if errWithCode := validateStatusContent(
		form.Status,
		form.SpoilerText,
		form.MediaIDs,
		form.Poll,
	); errWithCode != nil {
		return nil, errWithCode
	}

	// Ensure account populated; we'll need their settings.
	if err := p.state.DB.PopulateAccount(ctx, requester); err != nil {
		log.Errorf(ctx, "error(s) populating account, will continue: %s", err)
	}

	// Generate new ID for status.
	statusID := id.NewULID()

	// Process incoming status content fields.
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

	// Process incoming status attachments.
	media, errWithCode := p.processMedia(ctx,
		requester.ID,
		statusID,
		form.MediaIDs,
	)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// Generate necessary URIs for username, to build status URIs.
	accountURIs := uris.GenerateURIsForAccount(requester.Username)

	// Get current time.
	now := time.Now()

	status := &gtsmodel.Status{
		ID:                       statusID,
		URI:                      accountURIs.StatusesURI + "/" + statusID,
		URL:                      accountURIs.StatusesURL + "/" + statusID,
		CreatedAt:                now,
		Local:                    util.Ptr(true),
		Account:                  requester,
		AccountID:                requester.ID,
		AccountURI:               requester.URI,
		ActivityStreamsType:      ap.ObjectNote,
		Sensitive:                &form.Sensitive,
		CreatedWithApplicationID: application.ID,

		// Set validated language.
		Language: content.Language,

		// Set formatted status content.
		Content:        content.Content,
		ContentWarning: content.ContentWarning,
		Text:           form.Status, // raw

		// Set gathered mentions.
		MentionIDs: content.MentionIDs,
		Mentions:   content.Mentions,

		// Set gathered emojis.
		EmojiIDs: content.EmojiIDs,
		Emojis:   content.Emojis,

		// Set gathered tags.
		TagIDs: content.TagIDs,
		Tags:   content.Tags,

		// Set gathered media.
		AttachmentIDs: form.MediaIDs,
		Attachments:   media,

		// Assume not pending approval; this may
		// change when permissivity is checked.
		PendingApproval: util.Ptr(false),
	}

	// Check + attach in-reply-to status.
	if errWithCode := p.processInReplyTo(ctx,
		requester,
		status,
		form.InReplyToID,
	); errWithCode != nil {
		return nil, errWithCode
	}

	if errWithCode := p.processThreadID(ctx, status); errWithCode != nil {
		return nil, errWithCode
	}

	if err := p.processVisibility(ctx, form, requester.Settings.Privacy, status); err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Process policy AFTER visibility as it relies
	// on status.Visibility and form.Visibility being set.
	if errWithCode := processInteractionPolicy(form, requester.Settings, status); errWithCode != nil {
		return nil, errWithCode
	}

	if status.ContentWarning != "" && len(status.AttachmentIDs) > 0 {
		// If a content-warning is set, and
		// the status contains media, always
		// set the status sensitive flag.
		status.Sensitive = util.Ptr(true)
	}

	if form.Poll != nil {
		// Process poll, inserting into database.
		poll, errWithCode := p.processPoll(ctx,
			statusID,
			form.Poll,
			now,
		)
		if errWithCode != nil {
			return nil, errWithCode
		}

		// Set poll and its ID
		// on status before insert.
		status.PollID = poll.ID
		status.Poll = poll
		poll.Status = status

		// Update the status' ActivityPub type to Question.
		status.ActivityStreamsType = ap.ActivityQuestion
	}

	// Insert this newly prepared status into the database.
	if err := p.state.DB.PutStatus(ctx, status); err != nil {
		err := gtserror.Newf("error inserting status in db: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if status.Poll != nil && !status.Poll.ExpiresAt.IsZero() {
		// Now that the status is inserted, attempt to
		// schedule an expiry handler for the status poll.
		if err := p.polls.ScheduleExpiry(ctx, status.Poll); err != nil {
			log.Errorf(ctx, "error scheduling poll expiry: %v", err)
		}
	}

	// Send it to the client API worker for async side-effects.
	p.state.Workers.Client.Queue.Push(&messages.FromClientAPI{
		APObjectType:   ap.ObjectNote,
		APActivityType: ap.ActivityCreate,
		GTSModel:       status,
		Origin:         requester,
	})

	// If the new status replies to a status that
	// replies to us, use our reply as an implicit
	// accept of any pending interaction.
	implicitlyAccepted, errWithCode := p.implicitlyAccept(ctx,
		requester, status,
	)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// If we ended up implicitly accepting, mark the
	// replied-to status as no longer pending approval
	// so it's serialized properly via the API.
	if implicitlyAccepted {
		status.InReplyTo.PendingApproval = util.Ptr(false)
	}

	return p.c.GetAPIStatus(ctx, requester, status)
}

func (p *Processor) processInReplyTo(ctx context.Context, requester *gtsmodel.Account, status *gtsmodel.Status, inReplyToID string) gtserror.WithCode {
	if inReplyToID == "" {
		// Not a reply.
		// Nothing to do.
		return nil
	}

	// Fetch target in-reply-to status (checking visibility).
	inReplyTo, errWithCode := p.c.GetVisibleTargetStatus(ctx,
		requester,
		inReplyToID,
		nil,
	)
	if errWithCode != nil {
		return errWithCode
	}

	// If this is a boost, unwrap it to get source status.
	inReplyTo, errWithCode = p.c.UnwrapIfBoost(ctx,
		requester,
		inReplyTo,
	)
	if errWithCode != nil {
		return errWithCode
	}

	// Ensure valid reply target for requester.
	policyResult, err := p.intFilter.StatusReplyable(ctx,
		requester,
		inReplyTo,
	)
	if err != nil {
		err := gtserror.Newf("error seeing if status %s is replyable: %w", status.ID, err)
		return gtserror.NewErrorInternalError(err)
	}

	if policyResult.Forbidden() {
		const errText = "you do not have permission to reply to this status"
		err := gtserror.New(errText)
		return gtserror.NewErrorForbidden(err, errText)
	}

	// Derive pendingApproval status.
	var pendingApproval bool
	switch {
	case policyResult.WithApproval():
		// We're allowed to do
		// this pending approval.
		pendingApproval = true

	case policyResult.MatchedOnCollection():
		// We're permitted to do this, but since
		// we matched due to presence in a followers
		// or following collection, we should mark
		// as pending approval and wait until we can
		// prove it's been Accepted by the target.
		pendingApproval = true

		if *inReplyTo.Local {
			// If the target is local we don't need
			// to wait for an Accept from remote,
			// we can just preapprove it and have
			// the processor create the Accept.
			status.PreApproved = true
		}

	case policyResult.Permitted():
		// We're permitted to do this
		// based on another kind of match.
		pendingApproval = false
	}

	status.PendingApproval = &pendingApproval

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

func (p *Processor) processVisibility(
	ctx context.Context,
	form *apimodel.StatusCreateRequest,
	accountDefaultVis gtsmodel.Visibility,
	status *gtsmodel.Status,
) error {
	switch {
	// Visibility set on form, use that.
	case form.Visibility != "":
		status.Visibility = typeutils.APIVisToVis(form.Visibility)

	// Fall back to account default, set
	// this back on the form for later use.
	case accountDefaultVis != 0:
		status.Visibility = accountDefaultVis
		form.Visibility = p.converter.VisToAPIVis(ctx, accountDefaultVis)

	// What? Fall back to global default, set
	// this back on the form for later use.
	default:
		status.Visibility = gtsmodel.VisibilityDefault
		form.Visibility = p.converter.VisToAPIVis(ctx, gtsmodel.VisibilityDefault)
	}

	// Set federated according to "local_only" field,
	// assuming federated (ie., not local-only) by default.
	localOnly := util.PtrOrValue(form.LocalOnly, false)
	status.Federated = util.Ptr(!localOnly)

	return nil
}

func processInteractionPolicy(
	form *apimodel.StatusCreateRequest,
	settings *gtsmodel.AccountSettings,
	status *gtsmodel.Status,
) gtserror.WithCode {

	// If policy is set on the
	// form then prefer this.
	//
	// TODO: prevent scope widening by
	// limiting interaction policy if
	// inReplyTo status has a stricter
	// interaction policy than this one.
	if form.InteractionPolicy != nil {
		p, err := typeutils.APIInteractionPolicyToInteractionPolicy(
			form.InteractionPolicy,
			form.Visibility,
		)

		if err != nil {
			errWithCode := gtserror.NewErrorBadRequest(err, err.Error())
			return errWithCode
		}

		status.InteractionPolicy = p
		return nil
	}

	switch status.Visibility {

	case gtsmodel.VisibilityPublic:
		// Take account's default "public" policy if set.
		if p := settings.InteractionPolicyPublic; p != nil {
			status.InteractionPolicy = p
		}

	case gtsmodel.VisibilityUnlocked:
		// Take account's default "unlisted" policy if set.
		if p := settings.InteractionPolicyUnlocked; p != nil {
			status.InteractionPolicy = p
		}

	case gtsmodel.VisibilityFollowersOnly,
		gtsmodel.VisibilityMutualsOnly:
		// Take account's default followers-only policy if set.
		// TODO: separate policy for mutuals-only vis.
		if p := settings.InteractionPolicyFollowersOnly; p != nil {
			status.InteractionPolicy = p
		}

	case gtsmodel.VisibilityDirect:
		// Take account's default direct policy if set.
		if p := settings.InteractionPolicyDirect; p != nil {
			status.InteractionPolicy = p
		}
	}

	// If no policy set by now, status interaction
	// policy will be stored as nil, which just means
	// "fall back to global default policy". We avoid
	// setting it explicitly to save space.
	return nil
}

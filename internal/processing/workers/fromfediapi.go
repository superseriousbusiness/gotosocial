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

package workers

import (
	"context"
	"errors"
	"net/url"
	"time"

	"codeberg.org/gruf/go-kv"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/federation/dereferencing"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/uris"

	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	"github.com/superseriousbusiness/gotosocial/internal/processing/account"
	"github.com/superseriousbusiness/gotosocial/internal/processing/common"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// fediAPI wraps processing functions
// specifically for messages originating
// from the federation/ActivityPub API.
type fediAPI struct {
	state    *state.State
	surface  *Surface
	federate *federate
	account  *account.Processor
	common   *common.Processor
	utils    *utils
}

func (p *Processor) ProcessFromFediAPI(ctx context.Context, fMsg *messages.FromFediAPI) error {
	// Allocate new log fields slice
	fields := make([]kv.Field, 3, 5)
	fields[0] = kv.Field{"activityType", fMsg.APActivityType}
	fields[1] = kv.Field{"objectType", fMsg.APObjectType}
	fields[2] = kv.Field{"toAccount", fMsg.Receiving.Username}

	if fMsg.APIRI != nil {
		// An IRI was supplied, append to log
		fields = append(fields, kv.Field{
			"iri", fMsg.APIRI,
		})
	}

	// Include GTSModel in logs if appropriate.
	if fMsg.GTSModel != nil &&
		log.Level() >= log.DEBUG {
		fields = append(fields, kv.Field{
			"model", fMsg.GTSModel,
		})
	}

	l := log.WithContext(ctx).WithFields(fields...)
	l.Info("processing from fedi API")

	switch fMsg.APActivityType {

	// CREATE SOMETHING
	case ap.ActivityCreate:
		switch fMsg.APObjectType {

		// CREATE NOTE/STATUS
		case ap.ObjectNote:
			return p.fediAPI.CreateStatus(ctx, fMsg)

		// CREATE FOLLOW (request)
		case ap.ActivityFollow:
			return p.fediAPI.CreateFollowReq(ctx, fMsg)

		// CREATE LIKE/FAVE
		case ap.ActivityLike:
			return p.fediAPI.CreateLike(ctx, fMsg)

		// CREATE ANNOUNCE/BOOST
		case ap.ActivityAnnounce:
			return p.fediAPI.CreateAnnounce(ctx, fMsg)

		// CREATE BLOCK
		case ap.ActivityBlock:
			return p.fediAPI.CreateBlock(ctx, fMsg)

		// CREATE FLAG/REPORT
		case ap.ActivityFlag:
			return p.fediAPI.CreateFlag(ctx, fMsg)

		// CREATE QUESTION
		case ap.ActivityQuestion:
			return p.fediAPI.CreatePollVote(ctx, fMsg)
		}

	// UPDATE SOMETHING
	case ap.ActivityUpdate:
		switch fMsg.APObjectType {

		// UPDATE NOTE/STATUS
		case ap.ObjectNote:
			return p.fediAPI.UpdateStatus(ctx, fMsg)

		// UPDATE ACCOUNT
		case ap.ActorPerson:
			return p.fediAPI.UpdateAccount(ctx, fMsg)

		// UPDATE QUESTION
		case ap.ActivityQuestion:
			return p.fediAPI.UpdatePollVote(ctx, fMsg)
		}

	// ACCEPT SOMETHING
	case ap.ActivityAccept:
		switch fMsg.APObjectType {

		// ACCEPT (pending) FOLLOW
		case ap.ActivityFollow:
			return p.fediAPI.AcceptFollow(ctx, fMsg)

		// ACCEPT (pending) LIKE
		case ap.ActivityLike:
			return p.fediAPI.AcceptLike(ctx, fMsg)

		// ACCEPT (pending) REPLY
		case ap.ObjectNote:
			return p.fediAPI.AcceptReply(ctx, fMsg)

		// ACCEPT (pending) ANNOUNCE
		case ap.ActivityAnnounce:
			return p.fediAPI.AcceptAnnounce(ctx, fMsg)

		// ACCEPT (remote) REPLY or ANNOUNCE
		case ap.ObjectUnknown:
			return p.fediAPI.AcceptRemoteStatus(ctx, fMsg)
		}

	// REJECT SOMETHING
	case ap.ActivityReject:
		switch fMsg.APObjectType {

		// REJECT LIKE
		case ap.ActivityLike:
			return p.fediAPI.RejectLike(ctx, fMsg)

		// REJECT NOTE/STATUS (ie., reject a reply)
		case ap.ObjectNote:
			return p.fediAPI.RejectReply(ctx, fMsg)

		// REJECT BOOST
		case ap.ActivityAnnounce:
			return p.fediAPI.RejectAnnounce(ctx, fMsg)
		}

	// DELETE SOMETHING
	case ap.ActivityDelete:
		switch fMsg.APObjectType {

		// DELETE NOTE/STATUS
		case ap.ObjectNote:
			return p.fediAPI.DeleteStatus(ctx, fMsg)

		// DELETE ACCOUNT
		case ap.ActorPerson:
			return p.fediAPI.DeleteAccount(ctx, fMsg)
		}

	// MOVE SOMETHING
	case ap.ActivityMove:

		// MOVE ACCOUNT
		// fromfediapi_move.go.
		if fMsg.APObjectType == ap.ActorPerson {
			return p.fediAPI.MoveAccount(ctx, fMsg)
		}

	// UNDO SOMETHING
	case ap.ActivityUndo:

		// UNDO ANNOUNCE
		if fMsg.APObjectType == ap.ActivityAnnounce {
			return p.fediAPI.UndoAnnounce(ctx, fMsg)
		}
	}

	return gtserror.Newf("unhandled: %s %s", fMsg.APActivityType, fMsg.APObjectType)
}

func (p *fediAPI) CreateStatus(ctx context.Context, fMsg *messages.FromFediAPI) error {
	var (
		status     *gtsmodel.Status
		statusable ap.Statusable
		err        error
	)

	var ok bool

	switch {
	case fMsg.APObject != nil:
		// A model was provided, extract this from message.
		statusable, ok = fMsg.APObject.(ap.Statusable)
		if !ok {
			return gtserror.Newf("cannot cast %T -> ap.Statusable", fMsg.APObject)
		}

		// Create bare-bones model to pass
		// into RefreshStatus(), which it will
		// further populate and insert as new.
		bareStatus := new(gtsmodel.Status)
		bareStatus.Local = util.Ptr(false)
		bareStatus.URI = ap.GetJSONLDId(statusable).String()

		// Call RefreshStatus() to parse and process the provided
		// statusable model, which it will use to further flesh out
		// the bare bones model and insert it into the database.
		status, statusable, err = p.federate.RefreshStatus(ctx,
			fMsg.Receiving.Username,
			bareStatus,
			statusable,
			// Force refresh within 5min window.
			dereferencing.Fresh,
		)
		if err != nil {
			return gtserror.Newf("error processing new status %s: %w", bareStatus.URI, err)
		}

	case fMsg.APIRI != nil:
		// Model was not set, deref with IRI (this is a forward).
		// This will also cause the status to be inserted into the db.
		status, statusable, err = p.federate.GetStatusByURI(ctx,
			fMsg.Receiving.Username,
			fMsg.APIRI,
		)
		if err != nil {
			return gtserror.Newf("error dereferencing forwarded status %s: %w", fMsg.APIRI, err)
		}

	default:
		return gtserror.New("neither APObjectModel nor APIri set")
	}

	if statusable == nil {
		// Another thread beat us to
		// creating this status! Return
		// here and let the other thread
		// handle timelining + notifying.
		return nil
	}

	// If pending approval is true then
	// status must reply to a LOCAL status
	// that requires approval for the reply.
	pendingApproval := util.PtrOrZero(status.PendingApproval)

	switch {
	case pendingApproval && !status.PreApproved:
		// If approval is required and status isn't
		// preapproved, then just notify the account
		// that's being interacted with: they can
		// approve or deny the interaction later.
		if err := p.utils.requestReply(ctx, status); err != nil {
			return gtserror.Newf("error pending reply: %w", err)
		}

		// Return early.
		return nil

	case pendingApproval && status.PreApproved:
		// If approval is required and status is
		// preapproved, that means this is a reply
		// to one of our statuses with permission
		// that matched on a following/followers
		// collection. Do the Accept immediately and
		// then process everything else as normal.

		// Store an already-accepted interaction request.
		id := id.NewULID()
		approval := &gtsmodel.InteractionRequest{
			ID:                   id,
			StatusID:             status.InReplyToID,
			TargetAccountID:      status.InReplyToAccountID,
			TargetAccount:        status.InReplyToAccount,
			InteractingAccountID: status.AccountID,
			InteractingAccount:   status.Account,
			InteractionURI:       status.URI,
			InteractionType:      gtsmodel.InteractionLike,
			Reply:                status,
			URI:                  uris.GenerateURIForAccept(status.InReplyToAccount.Username, id),
			AcceptedAt:           time.Now(),
		}
		if err := p.state.DB.PutInteractionRequest(ctx, approval); err != nil {
			return gtserror.Newf("db error putting pre-approved interaction request: %w", err)
		}

		// Mark the status as now approved.
		status.PendingApproval = util.Ptr(false)
		status.PreApproved = false
		status.ApprovedByURI = approval.URI
		if err := p.state.DB.UpdateStatus(
			ctx,
			status,
			"pending_approval",
			"approved_by_uri",
		); err != nil {
			return gtserror.Newf("db error updating status: %w", err)
		}

		// Send out the approval as Accept.
		if err := p.federate.AcceptInteraction(ctx, approval); err != nil {
			return gtserror.Newf("error federating pre-approval of reply: %w", err)
		}

		// Don't return, just continue as normal.
	}

	// Update stats for the remote account.
	if err := p.utils.incrementStatusesCount(ctx, fMsg.Requesting, status); err != nil {
		log.Errorf(ctx, "error updating account stats: %v", err)
	}

	if err := p.surface.timelineAndNotifyStatus(ctx, status); err != nil {
		log.Errorf(ctx, "error timelining and notifying status: %v", err)
	}

	if status.InReplyToID != "" {
		// Interaction counts changed on the replied status; uncache the
		// prepared version from all timelines. The status dereferencer
		// functions will ensure necessary ancestors exist before this point.
		p.surface.invalidateStatusFromTimelines(ctx, status.InReplyToID)
	}

	return nil
}

func (p *fediAPI) CreatePollVote(ctx context.Context, fMsg *messages.FromFediAPI) error {
	// Cast poll vote type from the worker message.
	vote, ok := fMsg.GTSModel.(*gtsmodel.PollVote)
	if !ok {
		return gtserror.Newf("cannot cast %T -> *gtsmodel.PollVote", fMsg.GTSModel)
	}

	// Insert the new poll vote in the database, note this
	// will handle updating votes on the poll model itself.
	if err := p.state.DB.PutPollVote(ctx, vote); err != nil {
		return gtserror.Newf("error inserting poll vote in db: %w", err)
	}

	// Ensure the poll vote is fully populated at this point.
	if err := p.state.DB.PopulatePollVote(ctx, vote); err != nil {
		return gtserror.Newf("error populating poll vote from db: %w", err)
	}

	// Ensure the poll on the vote is fully populated to get origin status.
	if err := p.state.DB.PopulatePoll(ctx, vote.Poll); err != nil {
		return gtserror.Newf("error populating poll from db: %w", err)
	}

	// Get the origin status,
	// (also set the poll on it).
	status := vote.Poll.Status
	status.Poll = vote.Poll

	if *status.Local {
		// Before federating it, increment the poll vote
		// and voter counts, *only on our local copy*.
		status.Poll.IncrementVotes(vote.Choices, true)

		// These were poll votes in a local status, we need to
		// federate the updated status model with latest vote counts.
		if err := p.federate.UpdateStatus(ctx, status); err != nil {
			log.Errorf(ctx, "error federating status update: %v", err)
		}
	}

	// Interaction counts changed, uncache from timelines.
	p.surface.invalidateStatusFromTimelines(ctx, status.ID)

	return nil
}

func (p *fediAPI) UpdatePollVote(ctx context.Context, fMsg *messages.FromFediAPI) error {
	// Cast poll vote type from the worker message.
	vote, ok := fMsg.GTSModel.(*gtsmodel.PollVote)
	if !ok {
		return gtserror.Newf("cannot cast %T -> *gtsmodel.PollVote", fMsg.GTSModel)
	}

	// Update poll vote model (specifically only choices) in the database.
	if err := p.state.DB.UpdatePollVote(ctx, vote, "choices"); err != nil {
		return gtserror.Newf("error updating poll vote in db: %w", err)
	}

	// Update the vote counts on the poll model itself. These will have
	// been updated by message pusher as we can't know which were new.
	if err := p.state.DB.UpdatePoll(ctx, vote.Poll, "votes"); err != nil {
		return gtserror.Newf("error updating poll in db: %w", err)
	}

	// Get the origin status.
	status := vote.Poll.Status

	if *status.Local {
		// These were poll votes in a local status, we need to
		// federate the updated status model with latest vote counts.
		if err := p.federate.UpdateStatus(ctx, status); err != nil {
			log.Errorf(ctx, "error federating status update: %v", err)
		}
	}

	// Interaction counts changed, uncache from timelines.
	p.surface.invalidateStatusFromTimelines(ctx, status.ID)

	return nil
}

func (p *fediAPI) CreateFollowReq(ctx context.Context, fMsg *messages.FromFediAPI) error {
	followRequest, ok := fMsg.GTSModel.(*gtsmodel.FollowRequest)
	if !ok {
		return gtserror.Newf("%T not parseable as *gtsmodel.FollowRequest", fMsg.GTSModel)
	}

	if err := p.state.DB.PopulateFollowRequest(ctx, followRequest); err != nil {
		return gtserror.Newf("error populating follow request: %w", err)
	}

	if *followRequest.TargetAccount.Locked {
		// Local account is locked: just notify the follow request.
		if err := p.surface.notifyFollowRequest(ctx, followRequest); err != nil {
			log.Errorf(ctx, "error notifying follow request: %v", err)
		}

		// And update stats for the local account.
		if err := p.utils.incrementFollowRequestsCount(ctx, fMsg.Receiving); err != nil {
			log.Errorf(ctx, "error updating account stats: %v", err)
		}

		return nil
	}

	// Local account is not locked:
	// Automatically accept the follow request
	// and notify about the new follower.
	follow, err := p.state.DB.AcceptFollowRequest(
		ctx,
		followRequest.AccountID,
		followRequest.TargetAccountID,
	)
	if err != nil {
		return gtserror.Newf("error accepting follow request: %w", err)
	}

	// Update stats for the local account.
	if err := p.utils.incrementFollowersCount(ctx, fMsg.Receiving); err != nil {
		log.Errorf(ctx, "error updating account stats: %v", err)
	}

	// Update stats for the remote account.
	if err := p.utils.incrementFollowingCount(ctx, fMsg.Requesting); err != nil {
		log.Errorf(ctx, "error updating account stats: %v", err)
	}

	if err := p.federate.AcceptFollow(ctx, follow); err != nil {
		log.Errorf(ctx, "error federating follow request accept: %v", err)
	}

	if err := p.surface.notifyFollow(ctx, follow); err != nil {
		log.Errorf(ctx, "error notifying follow: %v", err)
	}

	return nil
}

func (p *fediAPI) CreateLike(ctx context.Context, fMsg *messages.FromFediAPI) error {
	fave, ok := fMsg.GTSModel.(*gtsmodel.StatusFave)
	if !ok {
		return gtserror.Newf("%T not parseable as *gtsmodel.StatusFave", fMsg.GTSModel)
	}

	// Ensure fave populated.
	if err := p.state.DB.PopulateStatusFave(ctx, fave); err != nil {
		return gtserror.Newf("error populating status fave: %w", err)
	}

	// If pending approval is true then
	// fave must target a LOCAL status
	// that requires approval for the fave.
	pendingApproval := util.PtrOrZero(fave.PendingApproval)

	switch {
	case pendingApproval && !fave.PreApproved:
		// If approval is required and fave isn't
		// preapproved, then just notify the account
		// that's being interacted with: they can
		// approve or deny the interaction later.
		if err := p.utils.requestFave(ctx, fave); err != nil {
			return gtserror.Newf("error pending fave: %w", err)
		}

		// Return early.
		return nil

	case pendingApproval && fave.PreApproved:
		// If approval is required and fave is
		// preapproved, that means this is a fave
		// of one of our statuses with permission
		// that matched on a following/followers
		// collection. Do the Accept immediately and
		// then process everything else as normal.

		// Store an already-accepted interaction request.
		id := id.NewULID()
		approval := &gtsmodel.InteractionRequest{
			ID:                   id,
			StatusID:             fave.StatusID,
			TargetAccountID:      fave.TargetAccountID,
			TargetAccount:        fave.TargetAccount,
			InteractingAccountID: fave.AccountID,
			InteractingAccount:   fave.Account,
			InteractionURI:       fave.URI,
			InteractionType:      gtsmodel.InteractionLike,
			Like:                 fave,
			URI:                  uris.GenerateURIForAccept(fave.TargetAccount.Username, id),
			AcceptedAt:           time.Now(),
		}
		if err := p.state.DB.PutInteractionRequest(ctx, approval); err != nil {
			return gtserror.Newf("db error putting pre-approved interaction request: %w", err)
		}

		// Mark the fave itself as now approved.
		fave.PendingApproval = util.Ptr(false)
		fave.PreApproved = false
		fave.ApprovedByURI = approval.URI
		if err := p.state.DB.UpdateStatusFave(
			ctx,
			fave,
			"pending_approval",
			"approved_by_uri",
		); err != nil {
			return gtserror.Newf("db error updating status fave: %w", err)
		}

		// Send out the approval as Accept.
		if err := p.federate.AcceptInteraction(ctx, approval); err != nil {
			return gtserror.Newf("error federating pre-approval of fave: %w", err)
		}

		// Don't return, just continue as normal.
	}

	if err := p.surface.notifyFave(ctx, fave); err != nil {
		log.Errorf(ctx, "error notifying fave: %v", err)
	}

	// Interaction counts changed on the faved status;
	// uncache the prepared version from all timelines.
	p.surface.invalidateStatusFromTimelines(ctx, fave.StatusID)

	return nil
}

func (p *fediAPI) CreateAnnounce(ctx context.Context, fMsg *messages.FromFediAPI) error {
	boost, ok := fMsg.GTSModel.(*gtsmodel.Status)
	if !ok {
		return gtserror.Newf("%T not parseable as *gtsmodel.Status", fMsg.GTSModel)
	}

	// Dereference into a boost wrapper status.
	//
	// Note: this will handle storing the boost in
	// the db, and dereferencing the target status
	// ancestors / descendants where appropriate.
	var err error
	boost, err = p.federate.EnrichAnnounce(
		ctx,
		boost,
		fMsg.Receiving.Username,
	)
	if err != nil {
		if gtserror.IsUnretrievable(err) ||
			gtserror.NotPermitted(err) {
			// Boosted status domain blocked, or
			// otherwise not permitted, nothing to do.
			log.Debugf(ctx, "skipping announce: %v", err)
			return nil
		}

		// Actual error.
		return gtserror.Newf("error dereferencing announce: %w", err)
	}

	// If pending approval is true then
	// boost must target a LOCAL status
	// that requires approval for the boost.
	pendingApproval := util.PtrOrZero(boost.PendingApproval)

	switch {
	case pendingApproval && !boost.PreApproved:
		// If approval is required and boost isn't
		// preapproved, then just notify the account
		// that's being interacted with: they can
		// approve or deny the interaction later.
		if err := p.utils.requestAnnounce(ctx, boost); err != nil {
			return gtserror.Newf("error pending boost: %w", err)
		}

		// Return early.
		return nil

	case pendingApproval && boost.PreApproved:
		// If approval is required and status is
		// preapproved, that means this is a boost
		// of one of our statuses with permission
		// that matched on a following/followers
		// collection. Do the Accept immediately and
		// then process everything else as normal.

		// Store an already-accepted interaction request.
		id := id.NewULID()
		approval := &gtsmodel.InteractionRequest{
			ID:                   id,
			StatusID:             boost.BoostOfID,
			TargetAccountID:      boost.BoostOfAccountID,
			TargetAccount:        boost.BoostOfAccount,
			InteractingAccountID: boost.AccountID,
			InteractingAccount:   boost.Account,
			InteractionURI:       boost.URI,
			InteractionType:      gtsmodel.InteractionLike,
			Announce:             boost,
			URI:                  uris.GenerateURIForAccept(boost.BoostOfAccount.Username, id),
			AcceptedAt:           time.Now(),
		}
		if err := p.state.DB.PutInteractionRequest(ctx, approval); err != nil {
			return gtserror.Newf("db error putting pre-approved interaction request: %w", err)
		}

		// Mark the boost itself as now approved.
		boost.PendingApproval = util.Ptr(false)
		boost.PreApproved = false
		boost.ApprovedByURI = approval.URI
		if err := p.state.DB.UpdateStatus(
			ctx,
			boost,
			"pending_approval",
			"approved_by_uri",
		); err != nil {
			return gtserror.Newf("db error updating status: %w", err)
		}

		// Send out the approval as Accept.
		if err := p.federate.AcceptInteraction(ctx, approval); err != nil {
			return gtserror.Newf("error federating pre-approval of boost: %w", err)
		}

		// Don't return, just continue as normal.
	}

	// Update stats for the remote account.
	if err := p.utils.incrementStatusesCount(ctx, fMsg.Requesting, boost); err != nil {
		log.Errorf(ctx, "error updating account stats: %v", err)
	}

	// Timeline and notify the announce.
	if err := p.surface.timelineAndNotifyStatus(ctx, boost); err != nil {
		log.Errorf(ctx, "error timelining and notifying status: %v", err)
	}

	if err := p.surface.notifyAnnounce(ctx, boost); err != nil {
		log.Errorf(ctx, "error notifying announce: %v", err)
	}

	// Interaction counts changed on the original status;
	// uncache the prepared version from all timelines.
	p.surface.invalidateStatusFromTimelines(ctx, boost.BoostOfID)

	return nil
}

func (p *fediAPI) CreateBlock(ctx context.Context, fMsg *messages.FromFediAPI) error {
	block, ok := fMsg.GTSModel.(*gtsmodel.Block)
	if !ok {
		return gtserror.Newf("%T not parseable as *gtsmodel.Block", fMsg.GTSModel)
	}

	// Remove each account's posts from the other's timelines.
	//
	// First home timelines.
	if err := p.state.Timelines.Home.WipeItemsFromAccountID(
		ctx,
		block.AccountID,
		block.TargetAccountID,
	); err != nil {
		log.Errorf(ctx, "error wiping items from block -> target's home timeline: %v", err)
	}

	if err := p.state.Timelines.Home.WipeItemsFromAccountID(
		ctx,
		block.TargetAccountID,
		block.AccountID,
	); err != nil {
		log.Errorf(ctx, "error wiping items from target -> block's home timeline: %v", err)
	}

	// Now list timelines.
	if err := p.state.Timelines.List.WipeItemsFromAccountID(
		ctx,
		block.AccountID,
		block.TargetAccountID,
	); err != nil {
		log.Errorf(ctx, "error wiping items from block -> target's list timeline(s): %v", err)
	}

	if err := p.state.Timelines.List.WipeItemsFromAccountID(
		ctx,
		block.TargetAccountID,
		block.AccountID,
	); err != nil {
		log.Errorf(ctx, "error wiping items from target -> block's list timeline(s): %v", err)
	}

	// Remove any follows that existed between blocker + blockee.
	if err := p.state.DB.DeleteFollow(
		ctx,
		block.AccountID,
		block.TargetAccountID,
	); err != nil {
		log.Errorf(ctx, "error deleting follow from block -> target: %v", err)
	}

	if err := p.state.DB.DeleteFollow(
		ctx,
		block.TargetAccountID,
		block.AccountID,
	); err != nil {
		log.Errorf(ctx, "error deleting follow from target -> block: %v", err)
	}

	// Remove any follow requests that existed between blocker + blockee.
	if err := p.state.DB.DeleteFollowRequest(
		ctx,
		block.AccountID,
		block.TargetAccountID,
	); err != nil {
		log.Errorf(ctx, "error deleting follow request from block -> target: %v", err)
	}

	if err := p.state.DB.DeleteFollowRequest(
		ctx,
		block.TargetAccountID,
		block.AccountID,
	); err != nil {
		log.Errorf(ctx, "error deleting follow request from target -> block: %v", err)
	}

	return nil
}

func (p *fediAPI) CreateFlag(ctx context.Context, fMsg *messages.FromFediAPI) error {
	incomingReport, ok := fMsg.GTSModel.(*gtsmodel.Report)
	if !ok {
		return gtserror.Newf("%T not parseable as *gtsmodel.Report", fMsg.GTSModel)
	}

	// TODO: handle additional side effects of flag creation:
	// - notify admins by dm / notification

	if err := p.surface.emailAdminReportOpened(ctx, incomingReport); err != nil {
		log.Errorf(ctx, "error emailing report opened: %v", err)
	}

	return nil
}

func (p *fediAPI) UpdateAccount(ctx context.Context, fMsg *messages.FromFediAPI) error {
	// Parse the old/existing account model.
	account, ok := fMsg.GTSModel.(*gtsmodel.Account)
	if !ok {
		return gtserror.Newf("cannot cast %T -> *gtsmodel.Account", fMsg.GTSModel)
	}

	// Because this was an Update, the new Accountable should be set on the message.
	apubAcc, ok := fMsg.APObject.(ap.Accountable)
	if !ok {
		return gtserror.Newf("cannot cast %T -> ap.Accountable", fMsg.APObject)
	}

	// Fetch up-to-date bio, avatar, header, etc.
	_, _, err := p.federate.RefreshAccount(
		ctx,
		fMsg.Receiving.Username,
		account,
		apubAcc,

		// Force refresh within 5s window.
		//
		// Missing account updates could be
		// detrimental to federation if they
		// include public key changes.
		dereferencing.Freshest,
	)
	if err != nil {
		log.Errorf(ctx, "error refreshing account: %v", err)
	}

	return nil
}

func (p *fediAPI) AcceptFollow(ctx context.Context, fMsg *messages.FromFediAPI) error {
	// Update stats for the remote account.
	if err := p.utils.decrementFollowRequestsCount(ctx, fMsg.Requesting); err != nil {
		log.Errorf(ctx, "error updating account stats: %v", err)
	}

	if err := p.utils.incrementFollowersCount(ctx, fMsg.Requesting); err != nil {
		log.Errorf(ctx, "error updating account stats: %v", err)
	}

	// Update stats for the local account.
	if err := p.utils.incrementFollowingCount(ctx, fMsg.Receiving); err != nil {
		log.Errorf(ctx, "error updating account stats: %v", err)
	}

	return nil
}

func (p *fediAPI) AcceptLike(ctx context.Context, fMsg *messages.FromFediAPI) error {
	// TODO: Add something here if we ever implement sending out Likes to
	// followers more broadly and not just the owner of the Liked status.
	return nil
}

func (p *fediAPI) AcceptReply(ctx context.Context, fMsg *messages.FromFediAPI) error {
	status, ok := fMsg.GTSModel.(*gtsmodel.Status)
	if !ok {
		return gtserror.Newf("%T not parseable as *gtsmodel.Status", fMsg.GTSModel)
	}

	// Update stats for the actor account.
	if err := p.utils.incrementStatusesCount(ctx, status.Account, status); err != nil {
		log.Errorf(ctx, "error updating account stats: %v", err)
	}

	// Timeline and notify the status.
	if err := p.surface.timelineAndNotifyStatus(ctx, status); err != nil {
		log.Errorf(ctx, "error timelining and notifying status: %v", err)
	}

	// Send out the reply again, fully this time.
	if err := p.federate.CreateStatus(ctx, status); err != nil {
		log.Errorf(ctx, "error federating announce: %v", err)
	}

	// Interaction counts changed on the replied-to status;
	// uncache the prepared version from all timelines.
	p.surface.invalidateStatusFromTimelines(ctx, status.InReplyToID)

	return nil
}

func (p *fediAPI) AcceptRemoteStatus(ctx context.Context, fMsg *messages.FromFediAPI) error {
	// See if we can accept a remote
	// status we don't have stored yet.
	objectIRI, ok := fMsg.APObject.(*url.URL)
	if !ok {
		return gtserror.Newf("%T not parseable as *url.URL", fMsg.APObject)
	}

	approvedByURI := fMsg.APIRI
	if approvedByURI == nil {
		return gtserror.New("approvedByURI was nil")
	}

	// Assume we're accepting a status; create a
	// barebones status for dereferencing purposes.
	bareStatus := &gtsmodel.Status{
		URI:           objectIRI.String(),
		ApprovedByURI: approvedByURI.String(),
	}

	// Call RefreshStatus() to process the provided
	// barebones status and insert it into the database,
	// if indeed it's actually a status URI we can fetch.
	//
	// This will also check whether the given AcceptIRI
	// actually grants permission for this status.
	status, _, err := p.federate.RefreshStatus(ctx,
		fMsg.Receiving.Username,
		bareStatus,
		nil, nil,
	)
	if err != nil {
		return gtserror.Newf("error processing accepted status %s: %w", bareStatus.URI, err)
	}

	// No error means it was indeed a remote status, and the
	// given approvedByURI permitted it. Timeline and notify it.
	if err := p.surface.timelineAndNotifyStatus(ctx, status); err != nil {
		log.Errorf(ctx, "error timelining and notifying status: %v", err)
	}

	// Interaction counts changed on the interacted status;
	// uncache the prepared version from all timelines.
	if status.InReplyToID != "" {
		p.surface.invalidateStatusFromTimelines(ctx, status.InReplyToID)
	}

	if status.BoostOfID != "" {
		p.surface.invalidateStatusFromTimelines(ctx, status.BoostOfID)
	}

	return nil
}

func (p *fediAPI) AcceptAnnounce(ctx context.Context, fMsg *messages.FromFediAPI) error {
	boost, ok := fMsg.GTSModel.(*gtsmodel.Status)
	if !ok {
		return gtserror.Newf("%T not parseable as *gtsmodel.Status", fMsg.GTSModel)
	}

	// Update stats for the actor account.
	if err := p.utils.incrementStatusesCount(ctx, boost.Account, boost); err != nil {
		log.Errorf(ctx, "error updating account stats: %v", err)
	}

	// Timeline and notify the boost wrapper status.
	if err := p.surface.timelineAndNotifyStatus(ctx, boost); err != nil {
		log.Errorf(ctx, "error timelining and notifying status: %v", err)
	}

	// Send out the boost again, fully this time.
	if err := p.federate.Announce(ctx, boost); err != nil {
		log.Errorf(ctx, "error federating announce: %v", err)
	}

	// Interaction counts changed on the boosted status;
	// uncache the prepared version from all timelines.
	p.surface.invalidateStatusFromTimelines(ctx, boost.BoostOfID)

	return nil
}

func (p *fediAPI) UpdateStatus(ctx context.Context, fMsg *messages.FromFediAPI) error {
	// Cast the existing Status model attached to msg.
	existing, ok := fMsg.GTSModel.(*gtsmodel.Status)
	if !ok {
		return gtserror.Newf("cannot cast %T -> *gtsmodel.Status", fMsg.GTSModel)
	}

	var freshness *dereferencing.FreshnessWindow

	// Cast the updated ActivityPub statusable object .
	apStatus, _ := fMsg.APObject.(ap.Statusable)

	if apStatus != nil {
		// If an AP object was provided, we
		// allow very fast refreshes that likely
		// indicate a status edit after post.
		freshness = dereferencing.Freshest
	}

	// Fetch up-to-date attach status attachments, etc.
	status, _, err := p.federate.RefreshStatus(
		ctx,
		fMsg.Receiving.Username,
		existing,
		apStatus,
		freshness,
	)
	if err != nil {
		log.Errorf(ctx, "error refreshing status: %v", err)
	}

	if status.Poll != nil && status.Poll.Closing {

		// If the latest status has a newly closed poll, at least compared
		// to the existing version, then notify poll close to all voters.
		if err := p.surface.notifyPollClose(ctx, status); err != nil {
			log.Errorf(ctx, "error sending poll notification: %v", err)
		}
	}

	// Push message that the status has been edited to streams.
	if err := p.surface.timelineStatusUpdate(ctx, status); err != nil {
		log.Errorf(ctx, "error streaming status edit: %v", err)
	}

	// Status representation was refetched, uncache from timelines.
	p.surface.invalidateStatusFromTimelines(ctx, status.ID)

	return nil
}

func (p *fediAPI) DeleteStatus(ctx context.Context, fMsg *messages.FromFediAPI) error {
	status, ok := fMsg.GTSModel.(*gtsmodel.Status)
	if !ok {
		return gtserror.Newf("%T not parseable as *gtsmodel.Status", fMsg.GTSModel)
	}

	// Try to populate status structs if possible,
	// in order to more thoroughly remove them.
	if err := p.state.DB.PopulateStatus(
		ctx, status,
	); err != nil && !errors.Is(err, db.ErrNoEntries) {
		return gtserror.Newf("db error populating status: %w", err)
	}

	// Drop any outgoing queued AP requests about / targeting
	// this status, (stops queued likes, boosts, creates etc).
	p.state.Workers.Delivery.Queue.Delete("ObjectID", status.URI)
	p.state.Workers.Delivery.Queue.Delete("TargetID", status.URI)

	// Drop any incoming queued client messages about / targeting
	// status, (stops processing of local origin data for status).
	p.state.Workers.Client.Queue.Delete("TargetURI", status.URI)

	// Drop any incoming queued federator messages targeting status,
	// (stops processing of remote origin data targeting this status).
	p.state.Workers.Federator.Queue.Delete("TargetURI", status.URI)

	// Delete attachments from this status, since this request
	// comes from the federating API, and there's no way the
	// poster can do a delete + redraft for it on our instance.
	const deleteAttachments = true

	// This is just a deletion, not a Reject,
	// we don't need to take a copy of this status.
	const copyToSinBin = false

	// Perform the actual status deletion.
	if err := p.utils.wipeStatus(
		ctx,
		status,
		deleteAttachments,
		copyToSinBin,
	); err != nil {
		log.Errorf(ctx, "error wiping status: %v", err)
	}

	// Update stats for the remote account.
	if err := p.utils.decrementStatusesCount(ctx, fMsg.Requesting, status); err != nil {
		log.Errorf(ctx, "error updating account stats: %v", err)
	}

	if status.InReplyToID != "" {
		// Interaction counts changed on the replied status;
		// uncache the prepared version from all timelines.
		p.surface.invalidateStatusFromTimelines(ctx, status.InReplyToID)
	}

	return nil
}

func (p *fediAPI) DeleteAccount(ctx context.Context, fMsg *messages.FromFediAPI) error {
	account, ok := fMsg.GTSModel.(*gtsmodel.Account)
	if !ok {
		return gtserror.Newf("%T not parseable as *gtsmodel.Account", fMsg.GTSModel)
	}

	// Drop any outgoing queued AP requests to / from / targeting
	// this account, (stops queued likes, boosts, creates etc).
	p.state.Workers.Delivery.Queue.Delete("ObjectID", account.URI)
	p.state.Workers.Delivery.Queue.Delete("TargetID", account.URI)

	// Drop any incoming queued client messages to / from this
	// account, (stops processing of local origin data for acccount).
	p.state.Workers.Client.Queue.Delete("Target.ID", account.ID)
	p.state.Workers.Client.Queue.Delete("TargetURI", account.URI)

	// Drop any incoming queued federator messages to this account,
	// (stops processing of remote origin data targeting this account).
	p.state.Workers.Federator.Queue.Delete("Requesting.ID", account.ID)
	p.state.Workers.Federator.Queue.Delete("TargetURI", account.URI)

	// First perform the actual account deletion.
	if err := p.account.Delete(ctx, account, account.ID); err != nil {
		log.Errorf(ctx, "error deleting account: %v", err)
	}

	return nil
}

func (p *fediAPI) RejectLike(ctx context.Context, fMsg *messages.FromFediAPI) error {
	req, ok := fMsg.GTSModel.(*gtsmodel.InteractionRequest)
	if !ok {
		return gtserror.Newf("%T not parseable as *gtsmodel.InteractionRequest", fMsg.GTSModel)
	}

	// At this point the InteractionRequest should already
	// be in the database, we just need to do side effects.

	// Send out the Reject.
	if err := p.federate.RejectInteraction(ctx, req); err != nil {
		log.Errorf(ctx, "error federating rejection of like: %v", err)
	}

	// Get the rejected fave.
	fave, err := p.state.DB.GetStatusFaveByURI(
		gtscontext.SetBarebones(ctx),
		req.InteractionURI,
	)
	if err != nil {
		return gtserror.Newf("db error getting rejected fave: %w", err)
	}

	// Delete the fave.
	if err := p.state.DB.DeleteStatusFaveByID(ctx, fave.ID); err != nil {
		return gtserror.Newf("db error deleting fave: %w", err)
	}

	return nil
}

func (p *fediAPI) RejectReply(ctx context.Context, fMsg *messages.FromFediAPI) error {
	req, ok := fMsg.GTSModel.(*gtsmodel.InteractionRequest)
	if !ok {
		return gtserror.Newf("%T not parseable as *gtsmodel.InteractionRequest", fMsg.GTSModel)
	}

	// At this point the InteractionRequest should already
	// be in the database, we just need to do side effects.

	// Get the rejected status.
	status, err := p.state.DB.GetStatusByURI(
		gtscontext.SetBarebones(ctx),
		req.InteractionURI,
	)
	if err != nil {
		return gtserror.Newf("db error getting rejected reply: %w", err)
	}

	// Delete attachments from this status.
	// It's rejected so there's no possibility
	// for the poster to delete + redraft it.
	const deleteAttachments = true

	// Keep a copy of the status in
	// the sin bin for future review.
	const copyToSinBin = true

	// Perform the actual status deletion.
	if err := p.utils.wipeStatus(
		ctx,
		status,
		deleteAttachments,
		copyToSinBin,
	); err != nil {
		log.Errorf(ctx, "error wiping reply: %v", err)
	}

	return nil
}

func (p *fediAPI) RejectAnnounce(ctx context.Context, fMsg *messages.FromFediAPI) error {
	req, ok := fMsg.GTSModel.(*gtsmodel.InteractionRequest)
	if !ok {
		return gtserror.Newf("%T not parseable as *gtsmodel.InteractionRequest", fMsg.GTSModel)
	}

	// At this point the InteractionRequest should already
	// be in the database, we just need to do side effects.

	// Get the rejected boost.
	boost, err := p.state.DB.GetStatusByURI(
		gtscontext.SetBarebones(ctx),
		req.InteractionURI,
	)
	if err != nil {
		return gtserror.Newf("db error getting rejected announce: %w", err)
	}

	// Boosts don't have attachments anyway
	// so it doesn't matter what we set here.
	const deleteAttachments = true

	// This is just a boost, don't
	// keep a copy in the sin bin.
	const copyToSinBin = true

	// Perform the actual status deletion.
	if err := p.utils.wipeStatus(
		ctx,
		boost,
		deleteAttachments,
		copyToSinBin,
	); err != nil {
		log.Errorf(ctx, "error wiping announce: %v", err)
	}

	return nil
}

func (p *fediAPI) UndoAnnounce(
	ctx context.Context,
	fMsg *messages.FromFediAPI,
) error {
	boost, ok := fMsg.GTSModel.(*gtsmodel.Status)
	if !ok {
		return gtserror.Newf("%T not parseable as *gtsmodel.Status", fMsg.GTSModel)
	}

	// Delete the boost wrapper itself.
	if err := p.state.DB.DeleteStatusByID(ctx, boost.ID); err != nil {
		return gtserror.Newf("db error deleting boost: %w", err)
	}

	// Update statuses count for the requesting account.
	if err := p.utils.decrementStatusesCount(ctx, fMsg.Requesting, boost); err != nil {
		log.Errorf(ctx, "error updating account stats: %v", err)
	}

	// Remove the boost wrapper from all timelines.
	if err := p.surface.deleteStatusFromTimelines(ctx, boost.ID); err != nil {
		log.Errorf(ctx, "error removing timelined boost: %v", err)
	}

	// Interaction counts changed on the boosted status;
	// uncache the prepared version from all timelines.
	p.surface.invalidateStatusFromTimelines(ctx, boost.BoostOfID)

	return nil
}

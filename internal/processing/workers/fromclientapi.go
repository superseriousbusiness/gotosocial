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
	"time"

	"codeberg.org/gruf/go-kv"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	"github.com/superseriousbusiness/gotosocial/internal/processing/account"
	"github.com/superseriousbusiness/gotosocial/internal/processing/common"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/internal/uris"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// clientAPI wraps processing functions
// specifically for messages originating
// from the client/REST API.
type clientAPI struct {
	state     *state.State
	converter *typeutils.Converter
	surface   *Surface
	federate  *federate
	account   *account.Processor
	common    *common.Processor
	utils     *utils
}

func (p *Processor) ProcessFromClientAPI(ctx context.Context, cMsg *messages.FromClientAPI) error {
	// Allocate new log fields slice
	fields := make([]kv.Field, 3, 4)
	fields[0] = kv.Field{"activityType", cMsg.APActivityType}
	fields[1] = kv.Field{"objectType", cMsg.APObjectType}
	fields[2] = kv.Field{"fromAccount", cMsg.Origin.Username}

	// Include GTSModel in logs if appropriate.
	if cMsg.GTSModel != nil &&
		log.Level() >= log.DEBUG {
		fields = append(fields, kv.Field{
			"model", cMsg.GTSModel,
		})
	}

	l := log.WithContext(ctx).WithFields(fields...)
	l.Info("processing from client API")

	switch cMsg.APActivityType {

	// CREATE SOMETHING
	case ap.ActivityCreate:
		switch cMsg.APObjectType {

		// CREATE USER (ie., new user+account sign-up)
		case ap.ObjectProfile:
			return p.clientAPI.CreateUser(ctx, cMsg)

		// CREATE NOTE/STATUS
		case ap.ObjectNote:
			return p.clientAPI.CreateStatus(ctx, cMsg)

		// CREATE QUESTION
		// (note we don't handle poll *votes* as AS
		// question type when federating (just notes),
		// but it makes for a nicer type switch here.
		case ap.ActivityQuestion:
			return p.clientAPI.CreatePollVote(ctx, cMsg)

		// CREATE FOLLOW (request)
		case ap.ActivityFollow:
			return p.clientAPI.CreateFollowReq(ctx, cMsg)

		// CREATE LIKE/FAVE
		case ap.ActivityLike:
			return p.clientAPI.CreateLike(ctx, cMsg)

		// CREATE ANNOUNCE/BOOST
		case ap.ActivityAnnounce:
			return p.clientAPI.CreateAnnounce(ctx, cMsg)

		// CREATE BLOCK
		case ap.ActivityBlock:
			return p.clientAPI.CreateBlock(ctx, cMsg)
		}

	// UPDATE SOMETHING
	case ap.ActivityUpdate:
		switch cMsg.APObjectType {

		// UPDATE NOTE/STATUS
		case ap.ObjectNote:
			return p.clientAPI.UpdateStatus(ctx, cMsg)

		// UPDATE ACCOUNT (ie., bio, settings, etc)
		case ap.ActorPerson:
			return p.clientAPI.UpdateAccount(ctx, cMsg)

		// UPDATE A FLAG/REPORT (mark as resolved/closed)
		case ap.ActivityFlag:
			return p.clientAPI.UpdateReport(ctx, cMsg)

		// UPDATE USER (ie., email address)
		case ap.ObjectProfile:
			return p.clientAPI.UpdateUser(ctx, cMsg)
		}

	// ACCEPT SOMETHING
	case ap.ActivityAccept:
		switch cMsg.APObjectType { //nolint:gocritic

		// ACCEPT FOLLOW (request)
		case ap.ActivityFollow:
			return p.clientAPI.AcceptFollow(ctx, cMsg)

		// ACCEPT USER (ie., new user+account sign-up)
		case ap.ObjectProfile:
			return p.clientAPI.AcceptUser(ctx, cMsg)

		// ACCEPT NOTE/STATUS (ie., accept a reply)
		case ap.ObjectNote:
			return p.clientAPI.AcceptReply(ctx, cMsg)

		// ACCEPT LIKE
		case ap.ActivityLike:
			return p.clientAPI.AcceptLike(ctx, cMsg)

		// ACCEPT BOOST
		case ap.ActivityAnnounce:
			return p.clientAPI.AcceptAnnounce(ctx, cMsg)
		}

	// REJECT SOMETHING
	case ap.ActivityReject:
		switch cMsg.APObjectType { //nolint:gocritic

		// REJECT FOLLOW (request)
		case ap.ActivityFollow:
			return p.clientAPI.RejectFollowRequest(ctx, cMsg)

		// REJECT USER (ie., new user+account sign-up)
		case ap.ObjectProfile:
			return p.clientAPI.RejectUser(ctx, cMsg)

		// REJECT NOTE/STATUS (ie., reject a reply)
		case ap.ObjectNote:
			return p.clientAPI.RejectReply(ctx, cMsg)

		// REJECT LIKE
		case ap.ActivityLike:
			return p.clientAPI.RejectLike(ctx, cMsg)

		// REJECT BOOST
		case ap.ActivityAnnounce:
			return p.clientAPI.RejectAnnounce(ctx, cMsg)
		}

	// UNDO SOMETHING
	case ap.ActivityUndo:
		switch cMsg.APObjectType {

		// UNDO FOLLOW (request)
		case ap.ActivityFollow:
			return p.clientAPI.UndoFollow(ctx, cMsg)

		// UNDO BLOCK
		case ap.ActivityBlock:
			return p.clientAPI.UndoBlock(ctx, cMsg)

		// UNDO LIKE/FAVE
		case ap.ActivityLike:
			return p.clientAPI.UndoFave(ctx, cMsg)

		// UNDO ANNOUNCE/BOOST
		case ap.ActivityAnnounce:
			return p.clientAPI.UndoAnnounce(ctx, cMsg)
		}

	// DELETE SOMETHING
	case ap.ActivityDelete:
		switch cMsg.APObjectType {

		// DELETE NOTE/STATUS
		case ap.ObjectNote:
			return p.clientAPI.DeleteStatus(ctx, cMsg)

		// DELETE REMOTE ACCOUNT or LOCAL USER+ACCOUNT
		case ap.ActorPerson, ap.ObjectProfile:
			return p.clientAPI.DeleteAccountOrUser(ctx, cMsg)
		}

	// FLAG/REPORT SOMETHING
	case ap.ActivityFlag:
		switch cMsg.APObjectType { //nolint:gocritic

		// FLAG/REPORT ACCOUNT
		case ap.ActorPerson:
			return p.clientAPI.ReportAccount(ctx, cMsg)
		}

	// MOVE SOMETHING
	case ap.ActivityMove:
		switch cMsg.APObjectType { //nolint:gocritic

		// MOVE ACCOUNT
		case ap.ActorPerson:
			return p.clientAPI.MoveAccount(ctx, cMsg)
		}
	}

	return gtserror.Newf("unhandled: %s %s", cMsg.APActivityType, cMsg.APObjectType)
}

func (p *clientAPI) CreateUser(ctx context.Context, cMsg *messages.FromClientAPI) error {
	newUser, ok := cMsg.GTSModel.(*gtsmodel.User)
	if !ok {
		return gtserror.Newf("%T not parseable as *gtsmodel.User", cMsg.GTSModel)
	}

	// Notify mods of the new signup.
	if err := p.surface.notifySignup(ctx, newUser); err != nil {
		log.Errorf(ctx, "error notifying mods of new sign-up: %v", err)
	}

	// Send "new sign up" email to mods.
	if err := p.surface.emailAdminNewSignup(ctx, newUser); err != nil {
		log.Errorf(ctx, "error emailing new signup: %v", err)
	}

	// Send "please confirm your address" email to the new user.
	if err := p.surface.emailUserPleaseConfirm(ctx, newUser, true); err != nil {
		log.Errorf(ctx, "error emailing confirm: %v", err)
	}

	return nil
}

func (p *clientAPI) CreateStatus(ctx context.Context, cMsg *messages.FromClientAPI) error {
	var status *gtsmodel.Status
	var backfill bool

	// Check passed client message model type.
	switch model := cMsg.GTSModel.(type) {
	case *gtsmodel.Status:
		status = model
	case *gtsmodel.BackfillStatus:
		status = model.Status
		backfill = true
	default:
		return gtserror.Newf("%T not parseable as *gtsmodel.Status or *gtsmodel.BackfillStatus", cMsg.GTSModel)
	}

	// If pending approval is true then status must
	// reply to a status (either one of ours or a
	// remote) that requires approval for the reply.
	pendingApproval := util.PtrOrZero(status.PendingApproval)

	switch {
	case pendingApproval && !status.PreApproved:
		// If approval is required and status isn't
		// preapproved, then send out the Create to
		// only the replied-to account (if it's remote),
		// and/or notify the account that's being
		// interacted with (if it's local): they can
		// approve or deny the interaction later.
		if err := p.utils.requestReply(ctx, status); err != nil {
			return gtserror.Newf("error pending reply: %w", err)
		}

		// Send Create to *remote* account inbox ONLY.
		if err := p.federate.CreateStatus(ctx, status); err != nil {
			return gtserror.Newf("error federating pending reply: %w", err)
		}

		// Return early.
		return nil

	case pendingApproval && status.PreApproved:
		// If approval is required and status is
		// preapproved, that means this is a reply
		// to one of our statuses with permission
		// that matched on a following/followers
		// collection. Do the Accept immediately and
		// then process everything else as normal,
		// sending out the Create with the approval
		// URI attached.

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

		if err := p.federate.AcceptInteraction(ctx, approval); err != nil {
			return gtserror.Newf("error federating pre-approval of reply: %w", err)
		}

		// Don't return, just continue as normal.
	}

	// Update stats for the actor account.
	if err := p.utils.incrementStatusesCount(ctx, cMsg.Origin, status); err != nil {
		log.Errorf(ctx, "error updating account stats: %v", err)
	}

	// We specifically do not timeline
	// or notify for backfilled statuses,
	// as these are more for archival than
	// newly posted content for user feeds.
	if !backfill {

		if err := p.surface.timelineAndNotifyStatus(ctx, status); err != nil {
			log.Errorf(ctx, "error timelining and notifying status: %v", err)
		}

		if err := p.federate.CreateStatus(ctx, status); err != nil {
			log.Errorf(ctx, "error federating status: %v", err)
		}
	}

	if status.InReplyToID != "" {
		// Interaction counts changed on the replied status;
		// uncache the prepared version from all timelines.
		p.surface.invalidateStatusFromTimelines(ctx, status.InReplyToID)
	}

	return nil
}

func (p *clientAPI) CreatePollVote(ctx context.Context, cMsg *messages.FromClientAPI) error {
	// Cast the create poll vote attached to message.
	vote, ok := cMsg.GTSModel.(*gtsmodel.PollVote)
	if !ok {
		return gtserror.Newf("cannot cast %T -> *gtsmodel.Pollvote", cMsg.GTSModel)
	}

	// Ensure the vote is fully populated in order to get original poll.
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
		// These are poll votes in a local status, we only need to
		// federate the updated status model with latest vote counts.
		if err := p.federate.UpdateStatus(ctx, status); err != nil {
			log.Errorf(ctx, "error federating status update: %v", err)
		}
	} else {
		// These are votes in a remote poll, federate to origin the new poll vote(s).
		if err := p.federate.CreatePollVote(ctx, vote.Poll, vote); err != nil {
			log.Errorf(ctx, "error federating poll vote: %v", err)
		}
	}

	// Interaction counts changed on the source status, uncache from timelines.
	p.surface.invalidateStatusFromTimelines(ctx, vote.Poll.StatusID)

	return nil
}

func (p *clientAPI) CreateFollowReq(ctx context.Context, cMsg *messages.FromClientAPI) error {
	followRequest, ok := cMsg.GTSModel.(*gtsmodel.FollowRequest)
	if !ok {
		return gtserror.Newf("%T not parseable as *gtsmodel.FollowRequest", cMsg.GTSModel)
	}

	// If target is a local, unlocked account,
	// we can skip side effects for the follow
	// request and accept the follow immediately.
	if cMsg.Target.IsLocal() && !*cMsg.Target.Locked {
		// Accept the FR first to get the Follow.
		follow, err := p.state.DB.AcceptFollowRequest(
			ctx,
			cMsg.Origin.ID,
			cMsg.Target.ID,
		)
		if err != nil {
			return gtserror.Newf("db error accepting follow req: %w", err)
		}

		// Use AcceptFollow to do side effects.
		return p.AcceptFollow(ctx, &messages.FromClientAPI{
			APObjectType:   ap.ActivityFollow,
			APActivityType: ap.ActivityAccept,
			GTSModel:       follow,
			Origin:         cMsg.Origin,
			Target:         cMsg.Target,
		})
	}

	// Update stats for the target account.
	if err := p.utils.incrementFollowRequestsCount(ctx, cMsg.Target); err != nil {
		log.Errorf(ctx, "error updating account stats: %v", err)
	}

	if err := p.surface.notifyFollowRequest(ctx, followRequest); err != nil {
		log.Errorf(ctx, "error notifying follow request: %v", err)
	}

	// Convert the follow request to follow model (requests are sent as follows).
	follow := p.converter.FollowRequestToFollow(ctx, followRequest)

	if err := p.federate.Follow(
		ctx,
		follow,
	); err != nil {
		log.Errorf(ctx, "error federating follow request: %v", err)
	}

	return nil
}

func (p *clientAPI) CreateLike(ctx context.Context, cMsg *messages.FromClientAPI) error {
	fave, ok := cMsg.GTSModel.(*gtsmodel.StatusFave)
	if !ok {
		return gtserror.Newf("%T not parseable as *gtsmodel.StatusFave", cMsg.GTSModel)
	}

	// Ensure fave populated.
	if err := p.state.DB.PopulateStatusFave(ctx, fave); err != nil {
		return gtserror.Newf("error populating status fave: %w", err)
	}

	// If pending approval is true then fave must
	// target a status (either one of ours or a
	// remote) that requires approval for the fave.
	pendingApproval := util.PtrOrZero(fave.PendingApproval)

	switch {
	case pendingApproval && !fave.PreApproved:
		// If approval is required and fave isn't
		// preapproved, then send out the Like to
		// only the faved account (if it's remote),
		// and/or notify the account that's being
		// interacted with (if it's local): they can
		// approve or deny the interaction later.
		if err := p.utils.requestFave(ctx, fave); err != nil {
			return gtserror.Newf("error pending fave: %w", err)
		}

		// Send Like to *remote* account inbox ONLY.
		if err := p.federate.Like(ctx, fave); err != nil {
			return gtserror.Newf("error federating pending Like: %v", err)
		}

		// Return early.
		return nil

	case pendingApproval && fave.PreApproved:
		// If approval is required and fave is
		// preapproved, that means this is a fave
		// of one of our statuses with permission
		// that matched on a following/followers
		// collection. Do the Accept immediately and
		// then process everything else as normal,
		// sending out the Like with the approval
		// URI attached.

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

		if err := p.federate.AcceptInteraction(ctx, approval); err != nil {
			return gtserror.Newf("error federating pre-approval of fave: %w", err)
		}

		// Don't return, just continue as normal.
	}

	if err := p.surface.notifyFave(ctx, fave); err != nil {
		log.Errorf(ctx, "error notifying fave: %v", err)
	}

	if err := p.federate.Like(ctx, fave); err != nil {
		log.Errorf(ctx, "error federating like: %v", err)
	}

	// Interaction counts changed on the faved status;
	// uncache the prepared version from all timelines.
	p.surface.invalidateStatusFromTimelines(ctx, fave.StatusID)

	return nil
}

func (p *clientAPI) CreateAnnounce(ctx context.Context, cMsg *messages.FromClientAPI) error {
	boost, ok := cMsg.GTSModel.(*gtsmodel.Status)
	if !ok {
		return gtserror.Newf("%T not parseable as *gtsmodel.Status", cMsg.GTSModel)
	}

	// If pending approval is true then status must
	// boost a status (either one of ours or a
	// remote) that requires approval for the boost.
	pendingApproval := util.PtrOrZero(boost.PendingApproval)

	switch {
	case pendingApproval && !boost.PreApproved:
		// If approval is required and boost isn't
		// preapproved, then send out the Announce to
		// only the boosted account (if it's remote),
		// and/or notify the account that's being
		// interacted with (if it's local): they can
		// approve or deny the interaction later.
		if err := p.utils.requestAnnounce(ctx, boost); err != nil {
			return gtserror.Newf("error pending boost: %w", err)
		}

		// Send Announce to *remote* account inbox ONLY.
		if err := p.federate.Announce(ctx, boost); err != nil {
			return gtserror.Newf("error federating pending Announce: %v", err)
		}

		// Return early.
		return nil

	case pendingApproval && boost.PreApproved:
		// If approval is required and boost is
		// preapproved, that means this is a boost
		// of one of our statuses with permission
		// that matched on a following/followers
		// collection. Do the Accept immediately and
		// then process everything else as normal,
		// sending out the Create with the approval
		// URI attached.

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

		if err := p.federate.AcceptInteraction(ctx, approval); err != nil {
			return gtserror.Newf("error federating pre-approval of boost: %w", err)
		}

		// Don't return, just continue as normal.
	}

	// Update stats for the actor account.
	if err := p.utils.incrementStatusesCount(ctx, cMsg.Origin, boost); err != nil {
		log.Errorf(ctx, "error updating account stats: %v", err)
	}

	// Timeline and notify the boost wrapper status.
	if err := p.surface.timelineAndNotifyStatus(ctx, boost); err != nil {
		log.Errorf(ctx, "error timelining and notifying status: %v", err)
	}

	// Notify the boost target account.
	if err := p.surface.notifyAnnounce(ctx, boost); err != nil {
		log.Errorf(ctx, "error notifying boost: %v", err)
	}

	if err := p.federate.Announce(ctx, boost); err != nil {
		log.Errorf(ctx, "error federating announce: %v", err)
	}

	// Interaction counts changed on the boosted status;
	// uncache the prepared version from all timelines.
	p.surface.invalidateStatusFromTimelines(ctx, boost.BoostOfID)

	return nil
}

func (p *clientAPI) CreateBlock(ctx context.Context, cMsg *messages.FromClientAPI) error {
	block, ok := cMsg.GTSModel.(*gtsmodel.Block)
	if !ok {
		return gtserror.Newf("%T not parseable as *gtsmodel.Block", cMsg.GTSModel)
	}

	// Remove blockee's statuses from blocker's timeline.
	if err := p.state.Timelines.Home.WipeItemsFromAccountID(
		ctx,
		block.AccountID,
		block.TargetAccountID,
	); err != nil {
		return gtserror.Newf("error wiping timeline items for block: %w", err)
	}

	// Remove blocker's statuses from blockee's timeline.
	if err := p.state.Timelines.Home.WipeItemsFromAccountID(
		ctx,
		block.TargetAccountID,
		block.AccountID,
	); err != nil {
		return gtserror.Newf("error wiping timeline items for block: %w", err)
	}

	// TODO: same with notifications?
	// TODO: same with bookmarks?

	if err := p.federate.Block(ctx, block); err != nil {
		log.Errorf(ctx, "error federating block: %v", err)
	}

	return nil
}

func (p *clientAPI) UpdateStatus(ctx context.Context, cMsg *messages.FromClientAPI) error {
	// Cast the updated Status model attached to msg.
	status, ok := cMsg.GTSModel.(*gtsmodel.Status)
	if !ok {
		return gtserror.Newf("cannot cast %T -> *gtsmodel.Status", cMsg.GTSModel)
	}

	// Federate the updated status changes out remotely.
	if err := p.federate.UpdateStatus(ctx, status); err != nil {
		log.Errorf(ctx, "error federating status update: %v", err)
	}

	if status.Poll != nil && status.Poll.Closing {

		// If the latest status has a newly closed poll, at least compared
		// to the existing version, then notify poll close to all voters.
		if err := p.surface.notifyPollClose(ctx, status); err != nil {
			log.Errorf(ctx, "error notifying poll close: %v", err)
		}
	}

	// Push message that the status has been edited to streams.
	if err := p.surface.timelineStatusUpdate(ctx, status); err != nil {
		log.Errorf(ctx, "error streaming status edit: %v", err)
	}

	// Status representation has changed, invalidate from timelines.
	p.surface.invalidateStatusFromTimelines(ctx, status.ID)

	return nil
}

func (p *clientAPI) UpdateAccount(ctx context.Context, cMsg *messages.FromClientAPI) error {
	account, ok := cMsg.GTSModel.(*gtsmodel.Account)
	if !ok {
		return gtserror.Newf("cannot cast %T -> *gtsmodel.Account", cMsg.GTSModel)
	}

	if err := p.federate.UpdateAccount(ctx, account); err != nil {
		log.Errorf(ctx, "error federating account update: %v", err)
	}

	return nil
}

func (p *clientAPI) UpdateReport(ctx context.Context, cMsg *messages.FromClientAPI) error {
	report, ok := cMsg.GTSModel.(*gtsmodel.Report)
	if !ok {
		return gtserror.Newf("%T not parseable as *gtsmodel.Report", cMsg.GTSModel)
	}

	if report.Account.IsRemote() {
		// Report creator is a remote account,
		// we shouldn't try to email them!
		return nil
	}

	if err := p.surface.emailUserReportClosed(ctx, report); err != nil {
		log.Errorf(ctx, "error emailing report closed: %v", err)
	}

	return nil
}

func (p *clientAPI) UpdateUser(ctx context.Context, cMsg *messages.FromClientAPI) error {
	user, ok := cMsg.GTSModel.(*gtsmodel.User)
	if !ok {
		return gtserror.Newf("cannot cast %T -> *gtsmodel.User", cMsg.GTSModel)
	}

	// The only possible "UpdateUser" action is to update the
	// user's email address, so we can safely assume by this
	// point that a new unconfirmed email address has been set.
	if err := p.surface.emailUserPleaseConfirm(ctx, user, false); err != nil {
		log.Errorf(ctx, "error emailing report closed: %v", err)
	}

	return nil
}

func (p *clientAPI) AcceptFollow(ctx context.Context, cMsg *messages.FromClientAPI) error {
	follow, ok := cMsg.GTSModel.(*gtsmodel.Follow)
	if !ok {
		return gtserror.Newf("%T not parseable as *gtsmodel.Follow", cMsg.GTSModel)
	}

	// Update stats for the target account.
	if err := p.utils.decrementFollowRequestsCount(ctx, cMsg.Target); err != nil {
		log.Errorf(ctx, "error updating account stats: %v", err)
	}

	if err := p.utils.incrementFollowersCount(ctx, cMsg.Target); err != nil {
		log.Errorf(ctx, "error updating account stats: %v", err)
	}

	// Update stats for the origin account.
	if err := p.utils.incrementFollowingCount(ctx, cMsg.Origin); err != nil {
		log.Errorf(ctx, "error updating account stats: %v", err)
	}

	if err := p.surface.notifyFollow(ctx, follow); err != nil {
		log.Errorf(ctx, "error notifying follow: %v", err)
	}

	if err := p.federate.AcceptFollow(ctx, follow); err != nil {
		log.Errorf(ctx, "error federating follow accept: %v", err)
	}

	return nil
}

func (p *clientAPI) RejectFollowRequest(ctx context.Context, cMsg *messages.FromClientAPI) error {
	followReq, ok := cMsg.GTSModel.(*gtsmodel.FollowRequest)
	if !ok {
		return gtserror.Newf("%T not parseable as *gtsmodel.FollowRequest", cMsg.GTSModel)
	}

	// Update stats for the target account.
	if err := p.utils.decrementFollowRequestsCount(ctx, cMsg.Target); err != nil {
		log.Errorf(ctx, "error updating account stats: %v", err)
	}

	if err := p.federate.RejectFollow(
		ctx,
		p.converter.FollowRequestToFollow(ctx, followReq),
	); err != nil {
		log.Errorf(ctx, "error federating follow reject: %v", err)
	}

	return nil
}

func (p *clientAPI) UndoFollow(ctx context.Context, cMsg *messages.FromClientAPI) error {
	follow, ok := cMsg.GTSModel.(*gtsmodel.Follow)
	if !ok {
		return gtserror.Newf("%T not parseable as *gtsmodel.Follow", cMsg.GTSModel)
	}

	// Update stats for the origin account.
	if err := p.utils.decrementFollowingCount(ctx, cMsg.Origin); err != nil {
		log.Errorf(ctx, "error updating account stats: %v", err)
	}

	// Update stats for the target account.
	if err := p.utils.decrementFollowersCount(ctx, cMsg.Target); err != nil {
		log.Errorf(ctx, "error updating account stats: %v", err)
	}

	if err := p.federate.UndoFollow(ctx, follow); err != nil {
		log.Errorf(ctx, "error federating follow undo: %v", err)
	}

	return nil
}

func (p *clientAPI) UndoBlock(ctx context.Context, cMsg *messages.FromClientAPI) error {
	block, ok := cMsg.GTSModel.(*gtsmodel.Block)
	if !ok {
		return gtserror.Newf("%T not parseable as *gtsmodel.Block", cMsg.GTSModel)
	}

	if err := p.federate.UndoBlock(ctx, block); err != nil {
		log.Errorf(ctx, "error federating block undo: %v", err)
	}

	return nil
}

func (p *clientAPI) UndoFave(ctx context.Context, cMsg *messages.FromClientAPI) error {
	statusFave, ok := cMsg.GTSModel.(*gtsmodel.StatusFave)
	if !ok {
		return gtserror.Newf("%T not parseable as *gtsmodel.StatusFave", cMsg.GTSModel)
	}

	if err := p.federate.UndoLike(ctx, statusFave); err != nil {
		log.Errorf(ctx, "error federating like undo: %v", err)
	}

	// Interaction counts changed on the faved status;
	// uncache the prepared version from all timelines.
	p.surface.invalidateStatusFromTimelines(ctx, statusFave.StatusID)

	return nil
}

func (p *clientAPI) UndoAnnounce(ctx context.Context, cMsg *messages.FromClientAPI) error {
	status, ok := cMsg.GTSModel.(*gtsmodel.Status)
	if !ok {
		return gtserror.Newf("%T not parseable as *gtsmodel.Status", cMsg.GTSModel)
	}

	if err := p.state.DB.DeleteStatusByID(ctx, status.ID); err != nil {
		return gtserror.Newf("db error deleting status: %w", err)
	}

	// Update stats for the origin account.
	if err := p.utils.decrementStatusesCount(ctx, cMsg.Origin, status); err != nil {
		log.Errorf(ctx, "error updating account stats: %v", err)
	}

	if err := p.surface.deleteStatusFromTimelines(ctx, status.ID); err != nil {
		log.Errorf(ctx, "error removing timelined status: %v", err)
	}

	if err := p.federate.UndoAnnounce(ctx, status); err != nil {
		log.Errorf(ctx, "error federating announce undo: %v", err)
	}

	// Interaction counts changed on the boosted status;
	// uncache the prepared version from all timelines.
	p.surface.invalidateStatusFromTimelines(ctx, status.BoostOfID)

	return nil
}

func (p *clientAPI) DeleteStatus(ctx context.Context, cMsg *messages.FromClientAPI) error {
	status, ok := cMsg.GTSModel.(*gtsmodel.Status)
	if !ok {
		return gtserror.Newf("%T not parseable as *gtsmodel.Status", cMsg.GTSModel)
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

	// Don't delete attachments, just unattach them:
	// this request comes from the client API and the
	// poster may want to use attachments again later.
	const deleteAttachments = false

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

	// Update stats for the origin account.
	if err := p.utils.decrementStatusesCount(ctx, cMsg.Origin, status); err != nil {
		log.Errorf(ctx, "error updating account stats: %v", err)
	}

	if err := p.federate.DeleteStatus(ctx, status); err != nil {
		log.Errorf(ctx, "error federating status delete: %v", err)
	}

	if status.InReplyToID != "" {
		// Interaction counts changed on the replied status;
		// uncache the prepared version from all timelines.
		p.surface.invalidateStatusFromTimelines(ctx, status.InReplyToID)
	}

	return nil
}

func (p *clientAPI) DeleteAccountOrUser(ctx context.Context, cMsg *messages.FromClientAPI) error {
	// The originID of the delete, one of:
	//   - ID of a domain block, for which
	//     this account delete is a side effect.
	//   - ID of the deleted account itself (self delete).
	//   - ID of an admin account (account suspension).
	var originID string

	if domainBlock, ok := cMsg.GTSModel.(*gtsmodel.DomainBlock); ok {
		// Origin is a domain block.
		originID = domainBlock.ID
	} else {
		// Origin is whichever account
		// originated this message.
		originID = cMsg.Origin.ID
	}

	// Extract target account.
	account := cMsg.Target

	// Drop any outgoing queued AP requests to / from / targeting
	// this account, (stops queued likes, boosts, creates etc).
	p.state.Workers.Delivery.Queue.Delete("ActorID", account.URI)
	p.state.Workers.Delivery.Queue.Delete("ObjectID", account.URI)
	p.state.Workers.Delivery.Queue.Delete("TargetID", account.URI)

	// Drop any incoming queued client messages to / from this
	// account, (stops processing of local origin data for acccount).
	p.state.Workers.Client.Queue.Delete("Origin.ID", account.ID)
	p.state.Workers.Client.Queue.Delete("Target.ID", account.ID)
	p.state.Workers.Client.Queue.Delete("TargetURI", account.URI)

	// Drop any incoming queued federator messages to this account,
	// (stops processing of remote origin data targeting this account).
	p.state.Workers.Federator.Queue.Delete("Receiving.ID", account.ID)
	p.state.Workers.Federator.Queue.Delete("TargetURI", account.URI)

	if err := p.federate.DeleteAccount(ctx, cMsg.Target); err != nil {
		log.Errorf(ctx, "error federating account delete: %v", err)
	}

	if err := p.account.Delete(ctx, cMsg.Target, originID); err != nil {
		log.Errorf(ctx, "error deleting account: %v", err)
	}

	return nil
}

func (p *clientAPI) ReportAccount(ctx context.Context, cMsg *messages.FromClientAPI) error {
	report, ok := cMsg.GTSModel.(*gtsmodel.Report)
	if !ok {
		return gtserror.Newf("%T not parseable as *gtsmodel.Report", cMsg.GTSModel)
	}

	// Federate this report to the
	// remote instance if desired.
	if *report.Forwarded {
		if err := p.federate.Flag(ctx, report); err != nil {
			log.Errorf(ctx, "error federating flag: %v", err)
		}
	}

	if err := p.surface.emailAdminReportOpened(ctx, report); err != nil {
		log.Errorf(ctx, "error emailing report opened: %v", err)
	}

	return nil
}

func (p *clientAPI) MoveAccount(ctx context.Context, cMsg *messages.FromClientAPI) error {
	// Redirect each local follower of
	// OriginAccount to follow move target.
	p.utils.redirectFollowers(ctx, cMsg.Origin, cMsg.Target)

	// At this point, we know OriginAccount has the
	// Move set on it. Just make sure it's populated.
	if err := p.state.DB.PopulateMove(ctx, cMsg.Origin.Move); err != nil {
		return gtserror.Newf("error populating Move: %w", err)
	}

	// Now send the Move message out to
	// OriginAccount's (remote) followers.
	if err := p.federate.MoveAccount(ctx, cMsg.Origin); err != nil {
		return gtserror.Newf("error federating account move: %w", err)
	}

	// Mark the move attempt as successful.
	cMsg.Origin.Move.SucceededAt = cMsg.Origin.Move.AttemptedAt
	if err := p.state.DB.UpdateMove(
		ctx,
		cMsg.Origin.Move,
		"succeeded_at",
	); err != nil {
		return gtserror.Newf("error marking move as successful: %w", err)
	}

	return nil
}

func (p *clientAPI) AcceptUser(ctx context.Context, cMsg *messages.FromClientAPI) error {
	newUser, ok := cMsg.GTSModel.(*gtsmodel.User)
	if !ok {
		return gtserror.Newf("%T not parseable as *gtsmodel.User", cMsg.GTSModel)
	}

	// Mark user as approved + clear sign-up IP.
	newUser.Approved = util.Ptr(true)
	newUser.SignUpIP = nil
	if err := p.state.DB.UpdateUser(ctx, newUser, "approved", "sign_up_ip"); err != nil {
		// Error now means we should return without
		// sending email + let admin try to approve again.
		return gtserror.Newf("db error updating user %s: %w", newUser.ID, err)
	}

	// Send "your sign-up has been approved" email to the new user.
	if err := p.surface.emailUserSignupApproved(ctx, newUser); err != nil {
		log.Errorf(ctx, "error emailing: %v", err)
	}

	return nil
}

func (p *clientAPI) RejectUser(ctx context.Context, cMsg *messages.FromClientAPI) error {
	deniedUser, ok := cMsg.GTSModel.(*gtsmodel.DeniedUser)
	if !ok {
		return gtserror.Newf("%T not parseable as *gtsmodel.DeniedUser", cMsg.GTSModel)
	}

	// Remove the account.
	if err := p.state.DB.DeleteAccount(ctx, cMsg.Target.ID); err != nil {
		log.Errorf(ctx,
			"db error deleting account %s: %v",
			cMsg.Target.ID, err,
		)
	}

	// Remove the user.
	if err := p.state.DB.DeleteUserByID(ctx, deniedUser.ID); err != nil {
		log.Errorf(ctx,
			"db error deleting user %s: %v",
			deniedUser.ID, err,
		)
	}

	// Store the deniedUser entry.
	if err := p.state.DB.PutDeniedUser(ctx, deniedUser); err != nil {
		log.Errorf(ctx,
			"db error putting denied user %s: %v",
			deniedUser.ID, err,
		)
	}

	if *deniedUser.SendEmail {
		// Send "your sign-up has been rejected" email to the denied user.
		if err := p.surface.emailUserSignupRejected(ctx, deniedUser); err != nil {
			log.Errorf(ctx, "error emailing: %v", err)
		}
	}

	return nil
}

func (p *clientAPI) AcceptLike(ctx context.Context, cMsg *messages.FromClientAPI) error {
	req, ok := cMsg.GTSModel.(*gtsmodel.InteractionRequest)
	if !ok {
		return gtserror.Newf("%T not parseable as *gtsmodel.InteractionRequest", cMsg.GTSModel)
	}

	// Notify the fave (distinct from the notif for the pending fave).
	if err := p.surface.notifyFave(ctx, req.Like); err != nil {
		log.Errorf(ctx, "error notifying fave: %v", err)
	}

	// Send out the Accept.
	if err := p.federate.AcceptInteraction(ctx, req); err != nil {
		log.Errorf(ctx, "error federating approval of like: %v", err)
	}

	// Interaction counts changed on the faved status;
	// uncache the prepared version from all timelines.
	p.surface.invalidateStatusFromTimelines(ctx, req.Like.StatusID)

	return nil
}

func (p *clientAPI) AcceptReply(ctx context.Context, cMsg *messages.FromClientAPI) error {
	req, ok := cMsg.GTSModel.(*gtsmodel.InteractionRequest)
	if !ok {
		return gtserror.Newf("%T not parseable as *gtsmodel.InteractionRequest", cMsg.GTSModel)
	}

	var (
		interactingAcct = req.InteractingAccount
		reply           = req.Reply
	)

	// Update stats for the reply author account.
	if err := p.utils.incrementStatusesCount(ctx, interactingAcct, reply); err != nil {
		log.Errorf(ctx, "error updating account stats: %v", err)
	}

	// Timeline the reply + notify relevant accounts.
	if err := p.surface.timelineAndNotifyStatus(ctx, reply); err != nil {
		log.Errorf(ctx, "error timelining and notifying status reply: %v", err)
	}

	// Send out the Accept.
	if err := p.federate.AcceptInteraction(ctx, req); err != nil {
		log.Errorf(ctx, "error federating approval of reply: %v", err)
	}

	// Interaction counts changed on the replied status;
	// uncache the prepared version from all timelines.
	p.surface.invalidateStatusFromTimelines(ctx, reply.InReplyToID)

	return nil
}

func (p *clientAPI) AcceptAnnounce(ctx context.Context, cMsg *messages.FromClientAPI) error {
	req, ok := cMsg.GTSModel.(*gtsmodel.InteractionRequest)
	if !ok {
		return gtserror.Newf("%T not parseable as *gtsmodel.InteractionRequest", cMsg.GTSModel)
	}

	var (
		interactingAcct = req.InteractingAccount
		boost           = req.Announce
	)

	// Update stats for the boost author account.
	if err := p.utils.incrementStatusesCount(ctx, interactingAcct, boost); err != nil {
		log.Errorf(ctx, "error updating account stats: %v", err)
	}

	// Timeline and notify the announce.
	if err := p.surface.timelineAndNotifyStatus(ctx, boost); err != nil {
		log.Errorf(ctx, "error timelining and notifying status: %v", err)
	}

	// Notify the announce (distinct from the notif for the pending announce).
	if err := p.surface.notifyAnnounce(ctx, boost); err != nil {
		log.Errorf(ctx, "error notifying announce: %v", err)
	}

	// Send out the Accept.
	if err := p.federate.AcceptInteraction(ctx, req); err != nil {
		log.Errorf(ctx, "error federating approval of announce: %v", err)
	}

	// Interaction counts changed on the original status;
	// uncache the prepared version from all timelines.
	p.surface.invalidateStatusFromTimelines(ctx, boost.BoostOfID)

	return nil
}

func (p *clientAPI) RejectLike(ctx context.Context, cMsg *messages.FromClientAPI) error {
	req, ok := cMsg.GTSModel.(*gtsmodel.InteractionRequest)
	if !ok {
		return gtserror.Newf("%T not parseable as *gtsmodel.InteractionRequest", cMsg.GTSModel)
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

	// Delete the status fave.
	if err := p.state.DB.DeleteStatusFaveByID(ctx, fave.ID); err != nil {
		return gtserror.Newf("db error deleting status fave: %w", err)
	}

	return nil
}

func (p *clientAPI) RejectReply(ctx context.Context, cMsg *messages.FromClientAPI) error {
	req, ok := cMsg.GTSModel.(*gtsmodel.InteractionRequest)
	if !ok {
		return gtserror.Newf("%T not parseable as *gtsmodel.InteractionRequest", cMsg.GTSModel)
	}

	// At this point the InteractionRequest should already
	// be in the database, we just need to do side effects.

	// Send out the Reject.
	if err := p.federate.RejectInteraction(ctx, req); err != nil {
		log.Errorf(ctx, "error federating rejection of reply: %v", err)
	}

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

func (p *clientAPI) RejectAnnounce(ctx context.Context, cMsg *messages.FromClientAPI) error {
	req, ok := cMsg.GTSModel.(*gtsmodel.InteractionRequest)
	if !ok {
		return gtserror.Newf("%T not parseable as *gtsmodel.InteractionRequest", cMsg.GTSModel)
	}

	// At this point the InteractionRequest should already
	// be in the database, we just need to do side effects.

	// Send out the Reject.
	if err := p.federate.RejectInteraction(ctx, req); err != nil {
		log.Errorf(ctx, "error federating rejection of announce: %v", err)
	}

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

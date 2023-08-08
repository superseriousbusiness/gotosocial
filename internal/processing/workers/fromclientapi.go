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

	"codeberg.org/gruf/go-kv"
	"codeberg.org/gruf/go-logger/v2/level"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
)

func (p *Processor) EnqueueClientAPI(ctx context.Context, msgs ...messages.FromClientAPI) {
	log.Trace(ctx, "enqueuing")
	_ = p.state.Workers.ClientAPI.MustEnqueueCtx(ctx, func(ctx context.Context) {
		for _, msg := range msgs {
			log.Trace(ctx, "processing: %+v", msg)
			if err := p.ProcessFromClientAPI(ctx, msg); err != nil {
				log.Errorf(ctx, "error processing client API message: %v", err)
			}
		}
	})
}

func (p *Processor) ProcessFromClientAPI(ctx context.Context, cMsg messages.FromClientAPI) error {
	// Allocate new log fields slice
	fields := make([]kv.Field, 3, 4)
	fields[0] = kv.Field{"activityType", cMsg.APActivityType}
	fields[1] = kv.Field{"objectType", cMsg.APObjectType}
	fields[2] = kv.Field{"fromAccount", cMsg.OriginAccount.Username}

	// Include GTSModel in logs if appropriate.
	if cMsg.GTSModel != nil &&
		log.Level() >= level.DEBUG {
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

		// CREATE PROFILE/ACCOUNT
		case ap.ObjectProfile, ap.ActorPerson:
			return p.cAPICreateAccount(ctx, cMsg)

		// CREATE NOTE/STATUS
		case ap.ObjectNote:
			return p.cAPICreateStatus(ctx, cMsg)

		// CREATE FOLLOW (request)
		case ap.ActivityFollow:
			return p.cAPICreateFollowReq(ctx, cMsg)

		// CREATE LIKE/FAVE
		case ap.ActivityLike:
			return p.cAPICreateLike(ctx, cMsg)

		// CREATE ANNOUNCE/BOOST
		case ap.ActivityAnnounce:
			return p.cAPICreateAnnounce(ctx, cMsg)

		// CREATE BLOCK
		case ap.ActivityBlock:
			return p.cAPICreateBlock(ctx, cMsg)
		}

	// UPDATE SOMETHING
	case ap.ActivityUpdate:
		switch cMsg.APObjectType {

		// UPDATE PROFILE/ACCOUNT
		case ap.ObjectProfile, ap.ActorPerson:
			return p.cAPIUpdateAccount(ctx, cMsg)

		// UPDATE A FLAG/REPORT (mark as resolved/closed)
		case ap.ActivityFlag:
			return p.cAPIUpdateReport(ctx, cMsg)
		}

	// ACCEPT SOMETHING
	case ap.ActivityAccept:
		switch cMsg.APObjectType { //nolint:gocritic

		// ACCEPT FOLLOW (request)
		case ap.ActivityFollow:
			return p.cAPIAcceptFollow(ctx, cMsg)
		}

	// REJECT SOMETHING
	case ap.ActivityReject:
		switch cMsg.APObjectType { //nolint:gocritic

		// REJECT FOLLOW (request)
		case ap.ActivityFollow:
			return p.cAPIRejectFollowRequest(ctx, cMsg)
		}

	// UNDO SOMETHING
	case ap.ActivityUndo:
		switch cMsg.APObjectType {

		// UNDO FOLLOW (request)
		case ap.ActivityFollow:
			return p.cAPIUndoFollow(ctx, cMsg)

		// UNDO BLOCK
		case ap.ActivityBlock:
			return p.cAPIUndoBlock(ctx, cMsg)

		// UNDO LIKE/FAVE
		case ap.ActivityLike:
			return p.cAPIUndoFave(ctx, cMsg)

		// UNDO ANNOUNCE/BOOST
		case ap.ActivityAnnounce:
			return p.cAPIUndoAnnounce(ctx, cMsg)
		}

	// DELETE SOMETHING
	case ap.ActivityDelete:
		switch cMsg.APObjectType {

		// DELETE NOTE/STATUS
		case ap.ObjectNote:
			return p.cAPIDeleteStatus(ctx, cMsg)

		// DELETE PROFILE/ACCOUNT
		case ap.ObjectProfile, ap.ActorPerson:
			return p.cAPIDeleteAccount(ctx, cMsg)
		}

	// FLAG/REPORT SOMETHING
	case ap.ActivityFlag:
		switch cMsg.APObjectType { //nolint:gocritic

		// FLAG/REPORT A PROFILE
		case ap.ObjectProfile:
			return p.cAPIReportAccount(ctx, cMsg)
		}
	}

	return nil
}

func (p *Processor) cAPICreateAccount(ctx context.Context, cMsg messages.FromClientAPI) error {
	account, ok := cMsg.GTSModel.(*gtsmodel.Account)
	if !ok {
		return gtserror.Newf("%T not parseable as *gtsmodel.Account", cMsg.GTSModel)
	}

	// Send a confirmation email to the newly created account.
	user, err := p.state.DB.GetUserByAccountID(ctx, account.ID)
	if err != nil {
		return gtserror.Newf("db error getting user for account id %s: %w", account.ID, err)
	}

	if err := p.user.EmailSendConfirmation(ctx, user, account.Username); err != nil {
		return gtserror.Newf("error emailing %s: %w", account.Username, err)
	}

	return nil
}

func (p *Processor) cAPICreateStatus(ctx context.Context, cMsg messages.FromClientAPI) error {
	status, ok := cMsg.GTSModel.(*gtsmodel.Status)
	if !ok {
		return gtserror.Newf("%T not parseable as *gtsmodel.Status", cMsg.GTSModel)
	}

	if err := p.timelineAndNotifyStatus(ctx, status); err != nil {
		return gtserror.Newf("error timelining status: %w", err)
	}

	if status.InReplyToID != "" {
		// Interaction counts changed on the replied status;
		// uncache the prepared version from all timelines.
		p.invalidateStatusFromTimelines(ctx, status.InReplyToID)
	}

	if err := p.federateCreateStatus(ctx, status); err != nil {
		return gtserror.Newf("error federating status: %w", err)
	}

	return nil
}

func (p *Processor) cAPICreateFollowReq(ctx context.Context, cMsg messages.FromClientAPI) error {
	followRequest, ok := cMsg.GTSModel.(*gtsmodel.FollowRequest)
	if !ok {
		return gtserror.Newf("%T not parseable as *gtsmodel.FollowRequest", cMsg.GTSModel)
	}

	if err := p.notifyFollowRequest(ctx, followRequest); err != nil {
		return gtserror.Newf("error notifying follow request: %w", err)
	}

	if err := p.federateFollow(
		ctx,
		p.tc.FollowRequestToFollow(ctx, followRequest),
	); err != nil {
		return gtserror.Newf("error federating follow: %w", err)
	}

	return nil
}

func (p *Processor) cAPICreateLike(ctx context.Context, cMsg messages.FromClientAPI) error {
	fave, ok := cMsg.GTSModel.(*gtsmodel.StatusFave)
	if !ok {
		return gtserror.Newf("%T not parseable as *gtsmodel.StatusFave", cMsg.GTSModel)
	}

	if err := p.notifyFave(ctx, fave); err != nil {
		return gtserror.Newf("error notifying fave: %w", err)
	}

	// Interaction counts changed on the faved status;
	// uncache the prepared version from all timelines.
	p.invalidateStatusFromTimelines(ctx, fave.StatusID)

	if err := p.federateLike(ctx, fave); err != nil {
		return gtserror.Newf("error federating like: %w", err)
	}

	return nil
}

func (p *Processor) cAPICreateAnnounce(ctx context.Context, cMsg messages.FromClientAPI) error {
	boost, ok := cMsg.GTSModel.(*gtsmodel.Status)
	if !ok {
		return gtserror.Newf("%T not parseable as *gtsmodel.Status", cMsg.GTSModel)
	}

	// Timeline and notify the boost wrapper status.
	if err := p.timelineAndNotifyStatus(ctx, boost); err != nil {
		return gtserror.Newf("error timelining boost: %w", err)
	}

	// Notify the boost target account.
	if err := p.notifyAnnounce(ctx, boost); err != nil {
		return gtserror.Newf("error notifying boost: %w", err)
	}

	// Interaction counts changed on the boosted status;
	// uncache the prepared version from all timelines.
	p.invalidateStatusFromTimelines(ctx, boost.BoostOfID)

	if err := p.federateAnnounce(ctx, boost); err != nil {
		return gtserror.Newf("error federating announce: %w", err)
	}

	return nil
}

func (p *Processor) cAPICreateBlock(ctx context.Context, cMsg messages.FromClientAPI) error {
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

	if err := p.federateBlock(ctx, block); err != nil {
		return gtserror.Newf("error federating block: %w", err)
	}

	return nil
}

func (p *Processor) cAPIUpdateAccount(ctx context.Context, cMsg messages.FromClientAPI) error {
	account, ok := cMsg.GTSModel.(*gtsmodel.Account)
	if !ok {
		return gtserror.Newf("%T not parseable as *gtsmodel.Account", cMsg.GTSModel)
	}

	if err := p.federateUpdateAccount(ctx, account); err != nil {
		return gtserror.Newf("error federating account update: %w", err)
	}

	return nil
}

func (p *Processor) cAPIUpdateReport(ctx context.Context, cMsg messages.FromClientAPI) error {
	report, ok := cMsg.GTSModel.(*gtsmodel.Report)
	if !ok {
		return gtserror.Newf("%T not parseable as *gtsmodel.Report", cMsg.GTSModel)
	}

	if report.Account.IsRemote() {
		// Report creator is a remote account,
		// we shouldn't try to email them!
		return nil
	}

	if err := p.emailReportClosed(ctx, report); err != nil {
		return gtserror.Newf("error sending report closed email: %w", err)
	}

	return nil
}

func (p *Processor) cAPIAcceptFollow(ctx context.Context, cMsg messages.FromClientAPI) error {
	follow, ok := cMsg.GTSModel.(*gtsmodel.Follow)
	if !ok {
		return gtserror.Newf("%T not parseable as *gtsmodel.Follow", cMsg.GTSModel)
	}

	if err := p.notifyFollow(ctx, follow); err != nil {
		return gtserror.Newf("error notifying follow: %w", err)
	}

	if err := p.federateAcceptFollow(ctx, follow); err != nil {
		return gtserror.Newf("error federating follow request accept: %w", err)
	}

	return nil
}

func (p *Processor) cAPIRejectFollowRequest(ctx context.Context, cMsg messages.FromClientAPI) error {
	followReq, ok := cMsg.GTSModel.(*gtsmodel.FollowRequest)
	if !ok {
		return gtserror.Newf("%T not parseable as *gtsmodel.FollowRequest", cMsg.GTSModel)
	}

	if err := p.federateRejectFollow(
		ctx,
		p.tc.FollowRequestToFollow(ctx, followReq),
	); err != nil {
		return gtserror.Newf("error federating reject follow: %w", err)
	}

	return nil
}

func (p *Processor) cAPIUndoFollow(ctx context.Context, cMsg messages.FromClientAPI) error {
	follow, ok := cMsg.GTSModel.(*gtsmodel.Follow)
	if !ok {
		return gtserror.Newf("%T not parseable as *gtsmodel.Follow", cMsg.GTSModel)
	}

	if err := p.federateUndoFollow(ctx, follow); err != nil {
		return gtserror.Newf("error federating undo follow: %w", err)
	}

	return nil
}

func (p *Processor) cAPIUndoBlock(ctx context.Context, cMsg messages.FromClientAPI) error {
	block, ok := cMsg.GTSModel.(*gtsmodel.Block)
	if !ok {
		return gtserror.Newf("%T not parseable as *gtsmodel.Block", cMsg.GTSModel)
	}

	if err := p.federateUndoBlock(ctx, block); err != nil {
		return gtserror.Newf("error federating undo block: %w", err)
	}

	return nil
}

func (p *Processor) cAPIUndoFave(ctx context.Context, cMsg messages.FromClientAPI) error {
	statusFave, ok := cMsg.GTSModel.(*gtsmodel.StatusFave)
	if !ok {
		return gtserror.Newf("%T not parseable as *gtsmodel.StatusFave", cMsg.GTSModel)
	}

	// Interaction counts changed on the faved status;
	// uncache the prepared version from all timelines.
	p.invalidateStatusFromTimelines(ctx, statusFave.StatusID)

	if err := p.federateUndoLike(ctx, statusFave); err != nil {
		return gtserror.Newf("error federating undo like: %w", err)
	}

	return nil
}

func (p *Processor) cAPIUndoAnnounce(ctx context.Context, cMsg messages.FromClientAPI) error {
	status, ok := cMsg.GTSModel.(*gtsmodel.Status)
	if !ok {
		return gtserror.Newf("%T not parseable as *gtsmodel.Status", cMsg.GTSModel)
	}

	if err := p.state.DB.DeleteStatusByID(ctx, status.ID); err != nil {
		return gtserror.Newf("db error deleting status: %w", err)
	}

	if err := p.deleteStatusFromTimelines(ctx, status.ID); err != nil {
		return gtserror.Newf("error removing status from timelines: %w", err)
	}

	// Interaction counts changed on the boosted status;
	// uncache the prepared version from all timelines.
	p.invalidateStatusFromTimelines(ctx, status.BoostOfID)

	if err := p.federateUndoAnnounce(ctx, status); err != nil {
		return gtserror.Newf("error federating undo announce: %w", err)
	}

	return nil
}

func (p *Processor) cAPIDeleteStatus(ctx context.Context, cMsg messages.FromClientAPI) error {
	// Don't delete attachments, just unattach them:
	// this request comes from the client API and the
	// poster may want to use attachments again later.
	const deleteAttachments = false

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

	if err := p.wipeStatus(ctx, status, deleteAttachments); err != nil {
		return gtserror.Newf("error wiping status: %w", err)
	}

	if status.InReplyToID != "" {
		// Interaction counts changed on the replied status;
		// uncache the prepared version from all timelines.
		p.invalidateStatusFromTimelines(ctx, status.InReplyToID)
	}

	if err := p.federateDeleteStatus(ctx, status); err != nil {
		return gtserror.Newf("error federating status delete: %w", err)
	}

	return nil
}

func (p *Processor) cAPIDeleteAccount(ctx context.Context, cMsg messages.FromClientAPI) error {
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
		originID = cMsg.OriginAccount.ID
	}

	if err := p.federateDeleteAccount(ctx, cMsg.TargetAccount); err != nil {
		return gtserror.Newf("error federating account delete: %w", err)
	}

	if err := p.account.Delete(ctx, cMsg.TargetAccount, originID); err != nil {
		return gtserror.Newf("error deleting account: %w", err)
	}

	return nil
}

func (p *Processor) cAPIReportAccount(ctx context.Context, cMsg messages.FromClientAPI) error {
	report, ok := cMsg.GTSModel.(*gtsmodel.Report)
	if !ok {
		return gtserror.Newf("%T not parseable as *gtsmodel.Report", cMsg.GTSModel)
	}

	// Federate this report to the
	// remote instance if desired.
	if *report.Forwarded {
		if err := p.federateFlag(ctx, report); err != nil {
			return gtserror.Newf("error federating report: %w", err)
		}
	}

	if err := p.emailReportOpened(ctx, report); err != nil {
		return gtserror.Newf("error sending report opened email: %w", err)
	}

	return nil
}

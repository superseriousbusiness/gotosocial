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
	"net/url"

	"codeberg.org/gruf/go-kv"
	"codeberg.org/gruf/go-logger/v2/level"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
)

func (p *Processor) EnqueueFediAPI(ctx context.Context, msgs ...messages.FromFediAPI) {
	log.Trace(ctx, "enqueuing")
	_ = p.state.Workers.Federator.MustEnqueueCtx(ctx, func(ctx context.Context) {
		for _, msg := range msgs {
			log.Trace(ctx, "processing: %+v", msg)
			if err := p.ProcessFromFediAPI(ctx, msg); err != nil {
				log.Errorf(ctx, "error processing fedi API message: %v", err)
			}
		}
	})
}

func (p *Processor) ProcessFromFediAPI(ctx context.Context, fMsg messages.FromFediAPI) error {
	// Allocate new log fields slice
	fields := make([]kv.Field, 3, 5)
	fields[0] = kv.Field{"activityType", fMsg.APActivityType}
	fields[1] = kv.Field{"objectType", fMsg.APObjectType}
	fields[2] = kv.Field{"toAccount", fMsg.ReceivingAccount.Username}

	if fMsg.APIri != nil {
		// An IRI was supplied, append to log
		fields = append(fields, kv.Field{
			"iri", fMsg.APIri,
		})
	}

	// Include GTSModel in logs if appropriate.
	if fMsg.GTSModel != nil &&
		log.Level() >= level.DEBUG {
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
			return p.fAPICreateStatus(ctx, fMsg)

		// CREATE FOLLOW (request)
		case ap.ActivityFollow:
			return p.fAPICreateFollowReq(ctx, fMsg)

		// CREATE LIKE/FAVE
		case ap.ActivityLike:
			return p.fAPICreateLike(ctx, fMsg)

		// CREATE ANNOUNCE/BOOST
		case ap.ActivityAnnounce:
			return p.fAPICreateAnnounce(ctx, fMsg)

		// CREATE BLOCK
		case ap.ActivityBlock:
			return p.fAPICreateBlock(ctx, fMsg)

		// CREATE FLAG/REPORT
		case ap.ActivityFlag:
			return p.fAPICreateFlag(ctx, fMsg)
		}

	// UPDATE SOMETHING
	case ap.ActivityUpdate:
		switch fMsg.APObjectType { //nolint:gocritic

		// UPDATE PROFILE/ACCOUNT
		case ap.ObjectProfile:
			return p.fAPIUpdateAccount(ctx, fMsg)
		}

	// DELETE SOMETHING
	case ap.ActivityDelete:
		switch fMsg.APObjectType {

		// DELETE NOTE/STATUS
		case ap.ObjectNote:
			return p.fAPIDeleteStatus(ctx, fMsg)

		// DELETE PROFILE/ACCOUNT
		case ap.ObjectProfile:
			return p.fAPIDeleteAccount(ctx, fMsg)
		}
	}

	return nil
}

func (p *Processor) fAPICreateStatus(ctx context.Context, fMsg messages.FromFediAPI) error {
	// Check the message for either an already
	// dereferenced and converted status, or an
	// ActivityPub IRI that we need to deref.
	var (
		status *gtsmodel.Status
		err    error
	)

	if fMsg.GTSModel != nil {
		// Model is set, use that.
		status, err = p.statusFromGTSModel(ctx, fMsg)
	} else {
		// Model is not set, use IRI.
		status, err = p.statusFromAPIRI(ctx, fMsg)
	}

	if err != nil {
		return gtserror.Newf("error extracting status from federatorMsg: %w", err)
	}

	if status.Account == nil || status.Account.IsRemote() {
		// Either no account attached yet, or a remote account.
		// Both situations we need to parse account URI to fetch it.
		accountURI, err := url.Parse(status.AccountURI)
		if err != nil {
			return err
		}

		// Ensure that account for this status has been deref'd.
		status.Account, _, err = p.federator.GetAccountByURI(
			ctx,
			fMsg.ReceivingAccount.Username,
			accountURI,
		)
		if err != nil {
			return err
		}
	}

	// Ensure status ancestors dereferenced. We need at least the
	// immediate parent (if present) to ascertain timelineability.
	if err := p.federator.DereferenceStatusAncestors(
		ctx,
		fMsg.ReceivingAccount.Username,
		status,
	); err != nil {
		return err
	}

	if status.InReplyToID != "" {
		// Interaction counts changed on the replied status;
		// uncache the prepared version from all timelines.
		p.invalidateStatusFromTimelines(ctx, status.InReplyToID)
	}

	if err := p.timelineAndNotifyStatus(ctx, status); err != nil {
		return gtserror.Newf("error timelining status: %w", err)
	}

	return nil
}

func (p *Processor) statusFromGTSModel(ctx context.Context, fMsg messages.FromFediAPI) (*gtsmodel.Status, error) {
	// There should be a status pinned to the message:
	// we've already checked to ensure this is not nil.
	status, ok := fMsg.GTSModel.(*gtsmodel.Status)
	if !ok {
		err := gtserror.New("Note was not parseable as *gtsmodel.Status")
		return nil, err
	}

	// AP statusable representation may have also
	// been set on message (no problem if not).
	statusable, _ := fMsg.APObjectModel.(ap.Statusable)

	// Call refresh on status to update
	// it (deref remote) if necessary.
	var err error
	status, _, err = p.federator.RefreshStatus(
		ctx,
		fMsg.ReceivingAccount.Username,
		status,
		statusable,
		false, // Don't force refresh.
	)
	if err != nil {
		return nil, gtserror.Newf("%w", err)
	}

	return status, nil
}

func (p *Processor) statusFromAPIRI(ctx context.Context, fMsg messages.FromFediAPI) (*gtsmodel.Status, error) {
	// There should be a status IRI pinned to
	// the federatorMsg for us to dereference.
	if fMsg.APIri == nil {
		err := gtserror.New(
			"status was not pinned to federatorMsg, " +
				"and neither was an IRI for us to dereference",
		)
		return nil, err
	}

	// Get the status + ensure we have
	// the most up-to-date version.
	status, _, err := p.federator.GetStatusByURI(
		ctx,
		fMsg.ReceivingAccount.Username,
		fMsg.APIri,
	)
	if err != nil {
		return nil, gtserror.Newf("%w", err)
	}

	return status, nil
}

func (p *Processor) fAPICreateFollowReq(ctx context.Context, fMsg messages.FromFediAPI) error {
	followRequest, ok := fMsg.GTSModel.(*gtsmodel.FollowRequest)
	if !ok {
		return gtserror.Newf("%T not parseable as *gtsmodel.FollowRequest", fMsg.GTSModel)
	}

	if *followRequest.TargetAccount.Locked {
		// Account on our instance is locked:
		// just notify the follow request.
		if err := p.notifyFollowRequest(ctx, followRequest); err != nil {
			return gtserror.Newf("error notifying follow request: %w", err)
		}

		return nil
	}

	// Account on our instance is not locked:
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

	if err := p.federateAcceptFollow(ctx, follow); err != nil {
		return gtserror.Newf("error federating accept follow request: %w", err)
	}

	if err := p.notifyFollow(ctx, follow); err != nil {
		return gtserror.Newf("error notifying follow: %w", err)
	}

	return nil
}

func (p *Processor) fAPICreateLike(ctx context.Context, fMsg messages.FromFediAPI) error {
	fave, ok := fMsg.GTSModel.(*gtsmodel.StatusFave)
	if !ok {
		return gtserror.Newf("%T not parseable as *gtsmodel.StatusFave", fMsg.GTSModel)
	}

	if err := p.notifyFave(ctx, fave); err != nil {
		return gtserror.Newf("error notifying fave: %w", err)
	}

	// Interaction counts changed on the faved status;
	// uncache the prepared version from all timelines.
	p.invalidateStatusFromTimelines(ctx, fave.StatusID)

	return nil
}

func (p *Processor) fAPICreateAnnounce(ctx context.Context, fMsg messages.FromFediAPI) error {
	status, ok := fMsg.GTSModel.(*gtsmodel.Status)
	if !ok {
		return gtserror.Newf("%T not parseable as *gtsmodel.Status", fMsg.GTSModel)
	}

	// Dereference status that this status boosts.
	if err := p.federator.DereferenceAnnounce(
		ctx,
		status,
		fMsg.ReceivingAccount.Username,
	); err != nil {
		return gtserror.Newf("error dereferencing announce: %w", err)
	}

	// Generate an ID for the boost wrapper status.
	statusID, err := id.NewULIDFromTime(status.CreatedAt)
	if err != nil {
		return gtserror.Newf("error generating id: %w", err)
	}
	status.ID = statusID

	// Store the boost wrapper status.
	if err := p.state.DB.PutStatus(ctx, status); err != nil {
		return gtserror.Newf("db error inserting status: %w", err)
	}

	// Ensure boosted status ancestors dereferenced. We need at least
	// the immediate parent (if present) to ascertain timelineability.
	if err := p.federator.DereferenceStatusAncestors(ctx,
		fMsg.ReceivingAccount.Username,
		status.BoostOf,
	); err != nil {
		return err
	}

	// Timeline and notify the announce.
	if err := p.timelineAndNotifyStatus(ctx, status); err != nil {
		return gtserror.Newf("error timelining status: %w", err)
	}

	if err := p.notifyAnnounce(ctx, status); err != nil {
		return gtserror.Newf("error notifying status: %w", err)
	}

	// Interaction counts changed on the boosted status;
	// uncache the prepared version from all timelines.
	p.invalidateStatusFromTimelines(ctx, status.ID)

	return nil
}

func (p *Processor) fAPICreateBlock(ctx context.Context, fMsg messages.FromFediAPI) error {
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
		return gtserror.Newf("%w", err)
	}

	if err := p.state.Timelines.Home.WipeItemsFromAccountID(
		ctx,
		block.TargetAccountID,
		block.AccountID,
	); err != nil {
		return gtserror.Newf("%w", err)
	}

	// Now list timelines.
	if err := p.state.Timelines.List.WipeItemsFromAccountID(
		ctx,
		block.AccountID,
		block.TargetAccountID,
	); err != nil {
		return gtserror.Newf("%w", err)
	}

	if err := p.state.Timelines.List.WipeItemsFromAccountID(
		ctx,
		block.TargetAccountID,
		block.AccountID,
	); err != nil {
		return gtserror.Newf("%w", err)
	}

	// Remove any follows that existed between blocker + blockee.
	if err := p.state.DB.DeleteFollow(
		ctx,
		block.AccountID,
		block.TargetAccountID,
	); err != nil {
		return gtserror.Newf(
			"db error deleting follow from %s targeting %s: %w",
			block.AccountID, block.TargetAccountID, err,
		)
	}

	if err := p.state.DB.DeleteFollow(
		ctx,
		block.TargetAccountID,
		block.AccountID,
	); err != nil {
		return gtserror.Newf(
			"db error deleting follow from %s targeting %s: %w",
			block.TargetAccountID, block.AccountID, err,
		)
	}

	// Remove any follow requests that existed between blocker + blockee.
	if err := p.state.DB.DeleteFollowRequest(
		ctx,
		block.AccountID,
		block.TargetAccountID,
	); err != nil {
		return gtserror.Newf(
			"db error deleting follow request from %s targeting %s: %w",
			block.AccountID, block.TargetAccountID, err,
		)
	}

	if err := p.state.DB.DeleteFollowRequest(
		ctx,
		block.TargetAccountID,
		block.AccountID,
	); err != nil {
		return gtserror.Newf(
			"db error deleting follow request from %s targeting %s: %w",
			block.TargetAccountID, block.AccountID, err,
		)
	}

	return nil
}

func (p *Processor) fAPICreateFlag(ctx context.Context, fMsg messages.FromFediAPI) error {
	incomingReport, ok := fMsg.GTSModel.(*gtsmodel.Report)
	if !ok {
		return gtserror.Newf("%T not parseable as *gtsmodel.Report", fMsg.GTSModel)
	}

	// TODO: handle additional side effects of flag creation:
	// - notify admins by dm / notification

	if err := p.emailReportOpened(ctx, incomingReport); err != nil {
		return gtserror.Newf("error sending report opened email: %w", err)
	}

	return nil
}

func (p *Processor) fAPIUpdateAccount(ctx context.Context, fMsg messages.FromFediAPI) error {
	// Parse the old/existing account model.
	account, ok := fMsg.GTSModel.(*gtsmodel.Account)
	if !ok {
		return gtserror.Newf("%T not parseable as *gtsmodel.Account", fMsg.GTSModel)
	}

	// Because this was an Update, the new Accountable should be set on the message.
	apubAcc, ok := fMsg.APObjectModel.(ap.Accountable)
	if !ok {
		return gtserror.Newf("%T not parseable as ap.Accountable", fMsg.APObjectModel)
	}

	// Fetch up-to-date bio, avatar, header, etc.
	_, _, err := p.federator.RefreshAccount(
		ctx,
		fMsg.ReceivingAccount.Username,
		account,
		apubAcc,
		true, // Force refresh.
	)
	if err != nil {
		return gtserror.Newf("error refreshing updated account: %w", err)
	}

	return nil
}

func (p *Processor) fAPIDeleteStatus(ctx context.Context, fMsg messages.FromFediAPI) error {
	// Delete attachments from this status, since this request
	// comes from the federating API, and there's no way the
	// poster can do a delete + redraft for it on our instance.
	const deleteAttachments = true

	status, ok := fMsg.GTSModel.(*gtsmodel.Status)
	if !ok {
		return gtserror.Newf("%T not parseable as *gtsmodel.Status", fMsg.GTSModel)
	}

	if err := p.wipeStatus(ctx, status, deleteAttachments); err != nil {
		return gtserror.Newf("error wiping status: %w", err)
	}

	if status.InReplyToID != "" {
		// Interaction counts changed on the replied status;
		// uncache the prepared version from all timelines.
		p.invalidateStatusFromTimelines(ctx, status.InReplyToID)
	}

	return nil
}

func (p *Processor) fAPIDeleteAccount(ctx context.Context, fMsg messages.FromFediAPI) error {
	account, ok := fMsg.GTSModel.(*gtsmodel.Account)
	if !ok {
		return gtserror.Newf("%T not parseable as *gtsmodel.Account", fMsg.GTSModel)
	}

	if err := p.account.Delete(ctx, account, account.ID); err != nil {
		return gtserror.Newf("error deleting account: %w", err)
	}

	return nil
}

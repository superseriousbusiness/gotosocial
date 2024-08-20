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

package common

import (
	"context"

	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/uris"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// ApproveFave stores + returns an interactionApproval for a fave,
// and updates the fave to unset pending_approval and set approval uri.
//
// Callers to this function should ensure they have a lock on the
// fave URI, and make sure they delete any pending interaction
// requests for the fave before or after calling.
func (p *Processor) ApproveFave(
	ctx context.Context,
	fave *gtsmodel.StatusFave,
) (*gtsmodel.InteractionApproval, error) {
	// Create + store the approval.
	id := id.NewULID()
	approval := &gtsmodel.InteractionApproval{
		ID:                   id,
		StatusID:             fave.StatusID,
		AccountID:            fave.TargetAccountID,
		Account:              fave.TargetAccount,
		InteractingAccountID: fave.AccountID,
		InteractingAccount:   fave.Account,
		InteractionURI:       fave.URI,
		InteractionType:      gtsmodel.InteractionLike,
		Like:                 fave,
		URI:                  uris.GenerateURIForAccept(fave.TargetAccount.Username, id),
	}

	if err := p.state.DB.PutInteractionApproval(ctx, approval); err != nil {
		err := gtserror.Newf("db error inserting interaction approval: %w", err)
		return nil, err
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
		err := gtserror.Newf("db error updating status fave: %w", err)
		return nil, err
	}

	return approval, nil
}

// RejectFave stores + returns an interactionRejection for a fave.
//
// Callers to this function should ensure they have a lock on the
// fave URI, and make sure they delete any pending interaction
// requests (and the fave itself if necessary) before or after calling.
func (p *Processor) RejectFave(
	ctx context.Context,
	fave *gtsmodel.StatusFave,
) (*gtsmodel.InteractionRejection, error) {
	// Create + store the rejection.
	id := id.NewULID()
	rejection := &gtsmodel.InteractionRejection{
		ID:                   id,
		StatusID:             fave.StatusID,
		AccountID:            fave.TargetAccountID,
		Account:              fave.TargetAccount,
		InteractingAccountID: fave.AccountID,
		InteractingAccount:   fave.Account,
		InteractionURI:       fave.URI,
		InteractionType:      gtsmodel.InteractionLike,
		URI:                  uris.GenerateURIForReject(fave.TargetAccount.Username, id),
	}

	if err := p.state.DB.PutInteractionRejection(ctx, rejection); err != nil {
		err := gtserror.Newf("db error inserting interaction rejection: %w", err)
		return nil, err
	}

	return rejection, nil
}

// ApproveReply stores + returns an interactionApproval for a reply,
// and updates the reply to unset pending_approval and set approval uri.
//
// Callers to this function should ensure they have a lock on the
// reply URI, and make sure they delete any pending interaction
// requests for the reply before or after calling.
func (p *Processor) ApproveReply(
	ctx context.Context,
	reply *gtsmodel.Status,
) (*gtsmodel.InteractionApproval, error) {
	// Create + store the approval.
	id := id.NewULID()
	approval := &gtsmodel.InteractionApproval{
		ID:                   id,
		StatusID:             reply.InReplyToID,
		AccountID:            reply.InReplyToAccountID,
		Account:              reply.InReplyToAccount,
		InteractingAccountID: reply.AccountID,
		InteractingAccount:   reply.Account,
		InteractionURI:       reply.URI,
		InteractionType:      gtsmodel.InteractionReply,
		Reply:                reply,
		URI:                  uris.GenerateURIForAccept(reply.InReplyToAccount.Username, id),
	}

	if err := p.state.DB.PutInteractionApproval(ctx, approval); err != nil {
		err := gtserror.Newf("db error inserting interaction approval: %w", err)
		return nil, err
	}

	// Mark the reply itself as now approved.
	reply.PendingApproval = util.Ptr(false)
	reply.PreApproved = false
	reply.ApprovedByURI = approval.URI

	if err := p.state.DB.UpdateStatus(
		ctx,
		reply,
		"pending_approval",
		"approved_by_uri",
	); err != nil {
		err := gtserror.Newf("db error updating status: %w", err)
		return nil, err
	}

	return approval, nil
}

// RejectReply stores + returns an interactionRejection for a reply.
//
// Callers to this function should ensure they have a lock on the
// reply URI, and make sure they delete any pending interaction
// requests (and the reply itself if necessary) before or after calling.
func (p *Processor) RejectReply(
	ctx context.Context,
	reply *gtsmodel.Status,
) (*gtsmodel.InteractionRejection, error) {
	// Create + store the rejection.
	id := id.NewULID()
	rejection := &gtsmodel.InteractionRejection{
		ID:                   id,
		StatusID:             reply.InReplyToID,
		AccountID:            reply.InReplyToAccountID,
		Account:              reply.InReplyToAccount,
		InteractingAccountID: reply.AccountID,
		InteractingAccount:   reply.Account,
		InteractionURI:       reply.URI,
		InteractionType:      gtsmodel.InteractionReply,
		URI:                  uris.GenerateURIForReject(reply.InReplyToAccount.Username, id),
	}

	if err := p.state.DB.PutInteractionRejection(ctx, rejection); err != nil {
		err := gtserror.Newf("db error inserting interaction rejection: %w", err)
		return nil, err
	}

	return rejection, nil
}

// ApproveAnnounce stores + returns an interactionApproval for an announce,
// and updates the announce to unset pending_approval and set approval uri.
//
// Callers to this function should ensure they have a lock on the
// announce wrapper URI, and make sure they delete any pending interaction
// requests for the announce before or after calling.
func (p *Processor) ApproveAnnounce(
	ctx context.Context,
	boost *gtsmodel.Status,
) (*gtsmodel.InteractionApproval, error) {
	// Create + store the approval.
	id := id.NewULID()
	approval := &gtsmodel.InteractionApproval{
		ID:                   id,
		StatusID:             boost.BoostOfID,
		AccountID:            boost.BoostOfAccountID,
		Account:              boost.BoostOfAccount,
		InteractingAccountID: boost.AccountID,
		InteractingAccount:   boost.Account,
		InteractionURI:       boost.URI,
		InteractionType:      gtsmodel.InteractionAnnounce,
		Announce:             boost,
		URI:                  uris.GenerateURIForAccept(boost.BoostOfAccount.Username, id),
	}

	if err := p.state.DB.PutInteractionApproval(ctx, approval); err != nil {
		err := gtserror.Newf("db error inserting interaction approval: %w", err)
		return nil, err
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
		err := gtserror.Newf("db error updating boost wrapper status: %w", err)
		return nil, err
	}

	return approval, nil
}

// RejectAnnounce stores + returns an interactionRejection for an announce.
//
// Callers to this function should ensure they have a lock on the
// announce URI, and make sure they delete any pending interaction
// requests (and the announce itself if necessary) before or after calling.
func (p *Processor) RejectAnnounce(
	ctx context.Context,
	boost *gtsmodel.Status,
) (*gtsmodel.InteractionRejection, error) {
	// Create + store the rejection.
	id := id.NewULID()
	Rejection := &gtsmodel.InteractionRejection{
		ID:                   id,
		StatusID:             boost.BoostOfID,
		AccountID:            boost.BoostOfAccountID,
		Account:              boost.BoostOfAccount,
		InteractingAccountID: boost.AccountID,
		InteractingAccount:   boost.Account,
		InteractionURI:       boost.URI,
		InteractionType:      gtsmodel.InteractionAnnounce,
		URI:                  uris.GenerateURIForReject(boost.BoostOfAccount.Username, id),
	}

	if err := p.state.DB.PutInteractionRejection(ctx, Rejection); err != nil {
		err := gtserror.Newf("db error inserting interaction rejection: %w", err)
		return nil, err
	}

	return Rejection, nil
}

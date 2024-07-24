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

package dereferencing

import (
	"context"

	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

// isPermittedStatus returns whether the given status
// is permitted to be stored on this instance, checking:
//
//   - author is not suspended
//   - status passes visibility checks
//   - status passes interaction policy checks
//
// If status is not permitted to be stored, the function
// will clean up after itself by removing the status.
//
// If status is a reply or a boost, and the author of
// the given status is only permitted to reply or boost
// pending approval, then "PendingApproval" will be set
// to "true" on status. Callers should check this
// and handle it as appropriate.
func (d *Dereferencer) isPermittedStatus(
	ctx context.Context,
	requestUser string,
	existing *gtsmodel.Status,
	status *gtsmodel.Status,
) (
	bool, // is permitted?
	error,
) {
	// our failure condition handling
	// at the end of this function for
	// the case of permission = false.
	onFalse := func() (bool, error) {
		if existing != nil {
			log.Infof(ctx, "deleting unpermitted: %s", existing.URI)

			// Delete existing status from database as it's no longer permitted.
			if err := d.state.DB.DeleteStatusByID(ctx, existing.ID); err != nil {
				log.Errorf(ctx, "error deleting %s after permissivity fail: %v", existing.URI, err)
			}
		}
		return false, nil
	}

	if status.Account.IsSuspended() {
		// The status author is suspended,
		// this shouldn't have reached here
		// but it's a fast check anyways.
		log.Debugf(ctx,
			"status author %s is suspended",
			status.AccountURI,
		)
		return onFalse()
	}

	if inReplyTo := status.InReplyTo; inReplyTo != nil {
		return d.isPermittedReply(
			ctx,
			requestUser,
			status,
			inReplyTo,
			onFalse,
		)
	} else if boostOf := status.BoostOf; boostOf != nil {
		return d.isPermittedBoost(
			ctx,
			requestUser,
			status,
			boostOf,
			onFalse,
		)
	}

	// Nothing else stopping this.
	return true, nil
}

func (d *Dereferencer) isPermittedReply(
	ctx context.Context,
	requestUser string,
	status *gtsmodel.Status,
	inReplyTo *gtsmodel.Status,
	onFalse func() (bool, error),
) (bool, error) {
	if inReplyTo.BoostOfID != "" {
		// We do not permit replies to
		// boost wrapper statuses. (this
		// shouldn't be able to happen).
		log.Info(ctx, "rejecting reply to boost wrapper status")
		return onFalse()
	}

	// Check visibility of local
	// inReplyTo to replying account.
	if inReplyTo.IsLocal() {
		visible, err := d.visFilter.StatusVisible(ctx,
			status.Account,
			inReplyTo,
		)
		if err != nil {
			err := gtserror.Newf("error checking inReplyTo visibility: %w", err)
			return false, err
		}

		// Our status is not visible to the
		// account trying to do the reply.
		if !visible {
			return onFalse()
		}
	}

	// Check interaction policy of inReplyTo.
	replyable, err := d.intFilter.StatusReplyable(ctx,
		status.Account,
		inReplyTo,
	)
	if err != nil {
		err := gtserror.Newf("error checking status replyability: %w", err)
		return false, err
	}

	if replyable.Forbidden() {
		// Replier is not permitted
		// to do this interaction.
		return onFalse()
	}

	// TODO in next PR: check conditional /
	// with approval and deref Accept.
	if !replyable.Permitted() {
		return onFalse()
	}

	return true, nil
}

func (d *Dereferencer) isPermittedBoost(
	ctx context.Context,
	requestUser string,
	status *gtsmodel.Status,
	boostOf *gtsmodel.Status,
	onFalse func() (bool, error),
) (bool, error) {
	if boostOf.BoostOfID != "" {
		// We do not permit boosts of
		// boost wrapper statuses. (this
		// shouldn't be able to happen).
		log.Info(ctx, "rejecting boost of boost wrapper status")
		return onFalse()
	}

	// Check visibility of local
	// boostOf to boosting account.
	if boostOf.IsLocal() {
		visible, err := d.visFilter.StatusVisible(ctx,
			status.Account,
			boostOf,
		)
		if err != nil {
			err := gtserror.Newf("error checking boostOf visibility: %w", err)
			return false, err
		}

		// Our status is not visible to the
		// account trying to do the boost.
		if !visible {
			return onFalse()
		}
	}

	// Check interaction policy of boostOf.
	boostable, err := d.intFilter.StatusBoostable(ctx,
		status.Account,
		boostOf,
	)
	if err != nil {
		err := gtserror.Newf("error checking status boostability: %w", err)
		return false, err
	}

	if boostable.Forbidden() {
		// Booster is not permitted
		// to do this interaction.
		return onFalse()
	}

	// TODO in next PR: check conditional /
	// with approval and deref Accept.
	if !boostable.Permitted() {
		return onFalse()
	}

	return true, nil
}

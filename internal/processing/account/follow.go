/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package account

import (
	"context"
	"errors"
	"fmt"

	"github.com/superseriousbusiness/gotosocial/internal/ap"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	"github.com/superseriousbusiness/gotosocial/internal/uris"
)

// FollowCreate handles a follow request to an account, either remote or local.
func (p *Processor) FollowCreate(ctx context.Context, requestingAccount *gtsmodel.Account, form *apimodel.AccountFollowRequest) (*apimodel.Relationship, gtserror.WithCode) {
	targetAccount, errWithCode := p.getFollowTarget(ctx, requestingAccount.ID, form.ID)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// Check if a follow exists already.
	if follows, err := p.state.DB.IsFollowing(ctx, requestingAccount, targetAccount); err != nil {
		err = fmt.Errorf("FollowCreate: db error checking follow: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	} else if follows {
		// Already follows, just return current relationship.
		return p.RelationshipGet(ctx, requestingAccount, form.ID)
	}

	// Check if a follow request exists already.
	if followRequested, err := p.state.DB.IsFollowRequested(ctx, requestingAccount, targetAccount); err != nil {
		err = fmt.Errorf("FollowCreate: db error checking follow request: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	} else if followRequested {
		// Already follow requested, just return current relationship.
		return p.RelationshipGet(ctx, requestingAccount, form.ID)
	}

	// Create and store a new follow request.
	followID, err := id.NewRandomULID()
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}
	followURI := uris.GenerateURIForFollow(requestingAccount.Username, followID)

	fr := &gtsmodel.FollowRequest{
		ID:              followID,
		URI:             followURI,
		AccountID:       requestingAccount.ID,
		Account:         requestingAccount,
		TargetAccountID: form.ID,
		TargetAccount:   targetAccount,
		ShowReblogs:     form.Reblogs,
		Notify:          form.Notify,
	}

	if err := p.state.DB.Put(ctx, fr); err != nil {
		err = fmt.Errorf("FollowCreate: error creating follow request in db: %s", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if targetAccount.IsLocal() && !*targetAccount.Locked {
		// If the target account is local and not locked,
		// we can already accept the follow request and
		// skip any further processing.
		//
		// Because we know the requestingAccount is also
		// local, we don't need to federate the accept out.
		if _, err := p.state.DB.AcceptFollowRequest(ctx, requestingAccount.ID, form.ID); err != nil {
			err = fmt.Errorf("FollowCreate: error accepting follow request for local unlocked account: %w", err)
			return nil, gtserror.NewErrorInternalError(err)
		}
	} else if targetAccount.IsRemote() {
		// Otherwise we leave the follow request as it is,
		// and we handle the rest of the process async.
		p.state.Workers.EnqueueClientAPI(ctx, messages.FromClientAPI{
			APObjectType:   ap.ActivityFollow,
			APActivityType: ap.ActivityCreate,
			GTSModel:       fr,
			OriginAccount:  requestingAccount,
			TargetAccount:  targetAccount,
		})
	}

	return p.RelationshipGet(ctx, requestingAccount, form.ID)
}

// FollowRemove handles the removal of a follow/follow request to an account, either remote or local.
func (p *Processor) FollowRemove(ctx context.Context, requestingAccount *gtsmodel.Account, targetAccountID string) (*apimodel.Relationship, gtserror.WithCode) {
	targetAccount, errWithCode := p.getFollowTarget(ctx, requestingAccount.ID, targetAccountID)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// Unfollow and deal with side effects.
	msgs, err := p.unfollow(ctx, requestingAccount, targetAccount)
	if err != nil {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("FollowRemove: account %s not found in the db: %s", targetAccountID, err))
	}

	// Batch queue accreted client api messages.
	p.state.Workers.EnqueueClientAPI(ctx, msgs...)

	return p.RelationshipGet(ctx, requestingAccount, targetAccountID)
}

/*
	Utility functions.
*/

// getFollowTarget is a convenience function which:
//   - Checks if account is trying to follow/unfollow itself.
//   - Returns not found if there's a block in place between accounts.
//   - Returns target account according to its id.
func (p *Processor) getFollowTarget(ctx context.Context, requestingAccountID string, targetAccountID string) (*gtsmodel.Account, gtserror.WithCode) {
	// Account can't follow or unfollow itself.
	if requestingAccountID == targetAccountID {
		err := errors.New("account can't follow or unfollow itself")
		return nil, gtserror.NewErrorNotAcceptable(err)
	}

	// Do nothing if a block exists in either direction between accounts.
	if blocked, err := p.state.DB.IsBlocked(ctx, requestingAccountID, targetAccountID, true); err != nil {
		err = fmt.Errorf("db error checking block between accounts: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	} else if blocked {
		err = errors.New("block exists between accounts")
		return nil, gtserror.NewErrorNotFound(err)
	}

	// Ensure target account retrievable.
	targetAccount, err := p.state.DB.GetAccountByID(ctx, targetAccountID)
	if err != nil {
		if !errors.Is(err, db.ErrNoEntries) {
			// Real db error.
			err = fmt.Errorf("db error looking for target account %s: %w", targetAccountID, err)
			return nil, gtserror.NewErrorInternalError(err)
		}
		// Account not found.
		err = fmt.Errorf("target account %s not found in the db", targetAccountID)
		return nil, gtserror.NewErrorNotFound(err, err.Error())
	}

	return targetAccount, nil
}

// unfollow is a convenience function for having requesting account
// unfollow (and un follow request) target account, if follows and/or
// follow requests exist.
//
// If a follow and/or follow request was removed this way, one or two
// messages will be returned which should then be processed by a client
// api worker.
func (p *Processor) unfollow(ctx context.Context, requestingAccount *gtsmodel.Account, targetAccount *gtsmodel.Account) ([]messages.FromClientAPI, error) {
	msgs := []messages.FromClientAPI{}

	if fURI, err := p.state.DB.Unfollow(ctx, requestingAccount.ID, targetAccount.ID); err != nil {
		err = fmt.Errorf("unfollow: error deleting follow from %s targeting %s: %w", requestingAccount.ID, targetAccount.ID, err)
		return nil, err
	} else if fURI != "" {
		// Follow status changed, process side effects.
		msgs = append(msgs, messages.FromClientAPI{
			APObjectType:   ap.ActivityFollow,
			APActivityType: ap.ActivityUndo,
			GTSModel: &gtsmodel.Follow{
				AccountID:       requestingAccount.ID,
				TargetAccountID: targetAccount.ID,
				URI:             fURI,
			},
			OriginAccount: requestingAccount,
			TargetAccount: targetAccount,
		})
	}

	if frURI, err := p.state.DB.UnfollowRequest(ctx, requestingAccount.ID, targetAccount.ID); err != nil {
		err = fmt.Errorf("unfollow: error deleting follow request from %s targeting %s: %w", requestingAccount.ID, targetAccount.ID, err)
		return nil, err
	} else if frURI != "" {
		// Follow request status changed, process side effects.
		msgs = append(msgs, messages.FromClientAPI{
			APObjectType:   ap.ActivityFollow,
			APActivityType: ap.ActivityUndo,
			GTSModel: &gtsmodel.Follow{
				AccountID:       requestingAccount.ID,
				TargetAccountID: targetAccount.ID,
				URI:             frURI,
			},
			OriginAccount: requestingAccount,
			TargetAccount: targetAccount,
		})
	}

	return msgs, nil
}

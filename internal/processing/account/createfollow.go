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

func (p *processor) FollowCreate(ctx context.Context, requestingAccount *gtsmodel.Account, form *apimodel.AccountFollowRequest) (*apimodel.Relationship, gtserror.WithCode) {
	// if there's a block between the accounts we shouldn't create the request ofc
	if blocked, err := p.db.IsBlocked(ctx, requestingAccount.ID, form.ID, true); err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	} else if blocked {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("block exists between accounts"))
	}

	// make sure the target account actually exists in our db
	targetAcct, err := p.db.GetAccountByID(ctx, form.ID)
	if err != nil {
		if err == db.ErrNoEntries {
			return nil, gtserror.NewErrorNotFound(fmt.Errorf("accountfollowcreate: account %s not found in the db: %s", form.ID, err))
		}
		return nil, gtserror.NewErrorInternalError(err)
	}

	// check if a follow exists already
	if follows, err := p.db.IsFollowing(ctx, requestingAccount, targetAcct); err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("accountfollowcreate: error checking follow in db: %s", err))
	} else if follows {
		// already follows so just return the relationship
		return p.RelationshipGet(ctx, requestingAccount, form.ID)
	}

	// check if a follow request exists already
	if followRequested, err := p.db.IsFollowRequested(ctx, requestingAccount, targetAcct); err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("accountfollowcreate: error checking follow request in db: %s", err))
	} else if followRequested {
		// already follow requested so just return the relationship
		return p.RelationshipGet(ctx, requestingAccount, form.ID)
	}

	// check for attempt to follow self
	if requestingAccount.ID == targetAcct.ID {
		return nil, gtserror.NewErrorNotAcceptable(fmt.Errorf("accountfollowcreate: account %s cannot follow itself", requestingAccount.ID))
	}

	// make the follow request
	newFollowID, err := id.NewRandomULID()
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	showReblogs := true
	notify := false
	fr := &gtsmodel.FollowRequest{
		ID:              newFollowID,
		AccountID:       requestingAccount.ID,
		TargetAccountID: form.ID,
		ShowReblogs:     &showReblogs,
		URI:             uris.GenerateURIForFollow(requestingAccount.Username, newFollowID),
		Notify:          &notify,
	}
	if form.Reblogs != nil {
		fr.ShowReblogs = form.Reblogs
	}
	if form.Notify != nil {
		fr.Notify = form.Notify
	}

	// whack it in the database
	if err := p.db.Put(ctx, fr); err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("accountfollowcreate: error creating follow request in db: %s", err))
	}

	// if it's a local account that's not locked we can just straight up accept the follow request
	if !*targetAcct.Locked && targetAcct.Domain == "" {
		if _, err := p.db.AcceptFollowRequest(ctx, requestingAccount.ID, form.ID); err != nil {
			return nil, gtserror.NewErrorInternalError(fmt.Errorf("accountfollowcreate: error accepting folow request for local unlocked account: %s", err))
		}
		// return the new relationship
		return p.RelationshipGet(ctx, requestingAccount, form.ID)
	}

	// otherwise we leave the follow request as it is and we handle the rest of the process asynchronously
	p.clientWorker.Queue(messages.FromClientAPI{
		APObjectType:   ap.ActivityFollow,
		APActivityType: ap.ActivityCreate,
		GTSModel:       fr,
		OriginAccount:  requestingAccount,
		TargetAccount:  targetAcct,
	})

	// return whatever relationship results from this
	return p.RelationshipGet(ctx, requestingAccount, form.ID)
}

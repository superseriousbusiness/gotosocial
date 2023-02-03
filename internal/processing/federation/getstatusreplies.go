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

package federation

import (
	"context"
	"fmt"
	"net/url"

	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/transport"
)

func (p *processor) GetStatusReplies(ctx context.Context, requestedUsername string, requestedStatusID string, page bool, onlyOtherAccounts bool, minID string, requestURL *url.URL) (interface{}, gtserror.WithCode) {
	// get the account the request is referring to
	requestedAccount, err := p.db.GetAccountByUsernameDomain(ctx, requestedUsername, "")
	if err != nil {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("database error getting account with username %s: %s", requestedUsername, err))
	}

	// authenticate the request
	requestingAccountURI, errWithCode := p.federator.AuthenticateFederatedRequest(ctx, requestedUsername)
	if errWithCode != nil {
		return nil, errWithCode
	}

	requestingAccount, err := p.federator.GetAccountByURI(
		transport.WithFastfail(ctx), requestedUsername, requestingAccountURI, false,
	)
	if err != nil {
		return nil, gtserror.NewErrorUnauthorized(err)
	}

	// authorize the request:
	// 1. check if a block exists between the requester and the requestee
	blocked, err := p.db.IsBlocked(ctx, requestedAccount.ID, requestingAccount.ID, true)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	if blocked {
		return nil, gtserror.NewErrorUnauthorized(fmt.Errorf("block exists between accounts %s and %s", requestedAccount.ID, requestingAccount.ID))
	}

	// get the status out of the database here
	s := &gtsmodel.Status{}
	if err := p.db.GetWhere(ctx, []db.Where{
		{Key: "id", Value: requestedStatusID},
		{Key: "account_id", Value: requestedAccount.ID},
	}, s); err != nil {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("database error getting status with id %s and account id %s: %s", requestedStatusID, requestedAccount.ID, err))
	}

	visible, err := p.filter.StatusVisible(ctx, s, requestingAccount)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}
	if !visible {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("status with id %s not visible to user with id %s", s.ID, requestingAccount.ID))
	}

	var data map[string]interface{}

	// now there are three scenarios:
	// 1. we're asked for the whole collection and not a page -- we can just return the collection, with no items, but a link to 'first' page.
	// 2. we're asked for a page but only_other_accounts has not been set in the query -- so we should just return the first page of the collection, with no items.
	// 3. we're asked for a page, and only_other_accounts has been set, and min_id has optionally been set -- so we need to return some actual items!
	switch {
	case !page:
		// scenario 1
		// get the collection
		collection, err := p.tc.StatusToASRepliesCollection(ctx, s, onlyOtherAccounts)
		if err != nil {
			return nil, gtserror.NewErrorInternalError(err)
		}

		data, err = streams.Serialize(collection)
		if err != nil {
			return nil, gtserror.NewErrorInternalError(err)
		}
	case page && requestURL.Query().Get("only_other_accounts") == "":
		// scenario 2
		// get the collection
		collection, err := p.tc.StatusToASRepliesCollection(ctx, s, onlyOtherAccounts)
		if err != nil {
			return nil, gtserror.NewErrorInternalError(err)
		}
		// but only return the first page
		data, err = streams.Serialize(collection.GetActivityStreamsFirst().GetActivityStreamsCollectionPage())
		if err != nil {
			return nil, gtserror.NewErrorInternalError(err)
		}
	default:
		// scenario 3
		// get immediate children
		replies, err := p.db.GetStatusChildren(ctx, s, true, minID)
		if err != nil {
			return nil, gtserror.NewErrorInternalError(err)
		}

		// filter children and extract URIs
		replyURIs := map[string]*url.URL{}
		for _, r := range replies {
			// only show public or unlocked statuses as replies
			if r.Visibility != gtsmodel.VisibilityPublic && r.Visibility != gtsmodel.VisibilityUnlocked {
				continue
			}

			// respect onlyOtherAccounts parameter
			if onlyOtherAccounts && r.AccountID == requestedAccount.ID {
				continue
			}

			// only show replies that the status owner can see
			visibleToStatusOwner, err := p.filter.StatusVisible(ctx, r, requestedAccount)
			if err != nil || !visibleToStatusOwner {
				continue
			}

			// only show replies that the requester can see
			visibleToRequester, err := p.filter.StatusVisible(ctx, r, requestingAccount)
			if err != nil || !visibleToRequester {
				continue
			}

			rURI, err := url.Parse(r.URI)
			if err != nil {
				continue
			}

			replyURIs[r.ID] = rURI
		}

		repliesPage, err := p.tc.StatusURIsToASRepliesPage(ctx, s, onlyOtherAccounts, minID, replyURIs)
		if err != nil {
			return nil, gtserror.NewErrorInternalError(err)
		}
		data, err = streams.Serialize(repliesPage)
		if err != nil {
			return nil, gtserror.NewErrorInternalError(err)
		}
	}

	return data, nil
}

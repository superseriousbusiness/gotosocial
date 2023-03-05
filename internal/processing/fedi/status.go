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

package fedi

import (
	"context"
	"fmt"
	"net/url"

	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

// StatusGet handles the getting of a fedi/activitypub representation of a particular status, performing appropriate
// authentication before returning a JSON serializable interface to the caller.
func (p *Processor) StatusGet(ctx context.Context, requestedUsername string, requestedStatusID string) (interface{}, gtserror.WithCode) {
	requestedAccount, requestingAccount, errWithCode := p.authenticate(ctx, requestedUsername)
	if errWithCode != nil {
		return nil, errWithCode
	}

	status, err := p.state.DB.GetStatusByID(ctx, requestedStatusID)
	if err != nil {
		return nil, gtserror.NewErrorNotFound(err)
	}

	if status.AccountID != requestedAccount.ID {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("status with id %s does not belong to account with id %s", status.ID, requestedAccount.ID))
	}

	visible, err := p.filter.StatusVisible(ctx, requestingAccount, status)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}
	if !visible {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("status with id %s not visible to user with id %s", status.ID, requestingAccount.ID))
	}

	asStatus, err := p.tc.StatusToAS(ctx, status)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	data, err := streams.Serialize(asStatus)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	return data, nil
}

// GetStatus handles the getting of a fedi/activitypub representation of replies to a status, performing appropriate
// authentication before returning a JSON serializable interface to the caller.
func (p *Processor) StatusRepliesGet(ctx context.Context, requestedUsername string, requestedStatusID string, page bool, onlyOtherAccounts bool, onlyOtherAccountsSet bool, minID string) (interface{}, gtserror.WithCode) {
	requestedAccount, requestingAccount, errWithCode := p.authenticate(ctx, requestedUsername)
	if errWithCode != nil {
		return nil, errWithCode
	}

	status, err := p.state.DB.GetStatusByID(ctx, requestedStatusID)
	if err != nil {
		return nil, gtserror.NewErrorNotFound(err)
	}

	if status.AccountID != requestedAccount.ID {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("status with id %s does not belong to account with id %s", status.ID, requestedAccount.ID))
	}

	visible, err := p.filter.StatusVisible(ctx, requestedAccount, status)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}
	if !visible {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("status with id %s not visible to user with id %s", status.ID, requestingAccount.ID))
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
		collection, err := p.tc.StatusToASRepliesCollection(ctx, status, onlyOtherAccounts)
		if err != nil {
			return nil, gtserror.NewErrorInternalError(err)
		}

		data, err = streams.Serialize(collection)
		if err != nil {
			return nil, gtserror.NewErrorInternalError(err)
		}
	case page && !onlyOtherAccountsSet:
		// scenario 2
		// get the collection
		collection, err := p.tc.StatusToASRepliesCollection(ctx, status, onlyOtherAccounts)
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
		replies, err := p.state.DB.GetStatusChildren(ctx, status, true, minID)
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
			visibleToStatusOwner, err := p.filter.StatusVisible(ctx, requestedAccount, r)
			if err != nil || !visibleToStatusOwner {
				continue
			}

			// only show replies that the requester can see
			visibleToRequester, err := p.filter.StatusVisible(ctx, requestingAccount, r)
			if err != nil || !visibleToRequester {
				continue
			}

			rURI, err := url.Parse(r.URI)
			if err != nil {
				continue
			}

			replyURIs[r.ID] = rURI
		}

		repliesPage, err := p.tc.StatusURIsToASRepliesPage(ctx, status, onlyOtherAccounts, minID, replyURIs)
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

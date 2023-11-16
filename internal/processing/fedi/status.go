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
	"errors"
	"fmt"
	"net/url"
	"slices"
	"strconv"

	"github.com/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/paging"
)

// StatusGet handles the getting of a fedi/activitypub representation of a local status.
// It performs appropriate authentication before returning a JSON serializable interface.
func (p *Processor) StatusGet(ctx context.Context, requestedUsername string, requestedStatusID string) (interface{}, gtserror.WithCode) {
	// Authenticate using http signature.
	requestedAccount, requestingAccount, errWithCode := p.authenticate(ctx, requestedUsername)
	if errWithCode != nil {
		return nil, errWithCode
	}

	status, err := p.state.DB.GetStatusByID(ctx, requestedStatusID)
	if err != nil {
		return nil, gtserror.NewErrorNotFound(err)
	}

	if status.AccountID != requestedAccount.ID {
		err := fmt.Errorf("status with id %s does not belong to account with id %s", status.ID, requestedAccount.ID)
		return nil, gtserror.NewErrorNotFound(err)
	}

	visible, err := p.filter.StatusVisible(ctx, requestingAccount, status)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	if !visible {
		err := fmt.Errorf("status with id %s not visible to user with id %s", status.ID, requestingAccount.ID)
		return nil, gtserror.NewErrorNotFound(err)
	}

	statusable, err := p.converter.StatusToAS(ctx, status)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	data, err := ap.Serialize(statusable)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	return data, nil
}

// GetStatus handles the getting of a fedi/activitypub representation of replies to a status,
// performing appropriate authentication before returning a JSON serializable interface to the caller.
func (p *Processor) StatusRepliesGet(
	ctx context.Context,
	requestedUser string,
	statusID string,
	page *paging.Page,
	onlyOtherAccounts *bool,
) (interface{}, gtserror.WithCode) {
	requested, _, errWithCode := p.authenticate(ctx, requestedUser)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// Get target status and ensure visible to requester.
	status, errWithCode := p.c.GetVisibleTargetStatus(ctx,
		requested,
		statusID,
	)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// Ensure status is by requested account.
	if status.AccountID != requested.ID {
		const text = "status does not belong to requested account"
		return nil, gtserror.NewErrorNotFound(errors.New(text), text)
	}

	// Parse collection ID from status' URI.
	collectionID, err := url.Parse(status.URI)
	if err != nil {
		err := gtserror.Newf("error parsing status uri %s: %w", status.URI, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	var obj vocab.Type

	// Start AS collection params.
	var params ap.CollectionParams
	params.ID = collectionID

	switch {
	case page == nil:
		// i.e. paging disabled.
		//
		// Just build collection object from params.
		obj = ap.NewASOrderedCollection(params)

	case onlyOtherAccounts == nil:
		// i.e. paging enabled, but only first page.

		// Start AS collection page params.
		var pageParams ap.CollectionPageParams
		pageParams.CollectionParams = params
		pageParams.Append = func(int, ap.ItemsPropertyBuilder) {
			panic("this should not be called!")
		}

		// Build AS collection page from params.
		obj = ap.NewASOrderedCollectionPage(pageParams)

	default:
		// i.e. paging enabled, with an onlyOtherAccounts param.
		//
		// Get all immediate children (replies) of status in question.
		replies, err := p.state.DB.GetStatusReplies(ctx, status.ID, page)
		if err != nil {
			err := gtserror.Newf("error getting status children: %w", err)
			return nil, gtserror.NewErrorInternalError(err)
		}

		// Filter replies so we only show those visible to requester.
		replies, err = p.filter.StatusesVisible(ctx, requested, replies)
		if err != nil {
			err := gtserror.Newf("error filtering status children: %w", err)
			return nil, gtserror.NewErrorInternalError(err)
		}

		if *onlyOtherAccounts {
			// If 'onlyOtherAccounts' is set, drop all by the requested account.
			replies = slices.DeleteFunc(replies, func(status *gtsmodel.Status) bool {
				return status.AccountID == requested.ID
			})
		}

		// Get the lowest and highest
		// ID values, used for paging.
		lo := replies[len(replies)-1].ID
		hi := replies[0].ID

		// Start AS collection page params.
		var pageParams ap.CollectionPageParams
		pageParams.CollectionParams = params

		// Add the 'onlyOtherAccounts' query param.
		pageParams.Query = make(url.Values, 1)
		onlyOtherAcc := strconv.FormatBool(*onlyOtherAccounts)
		pageParams.Query.Set("onlyOtherAccounts", onlyOtherAcc)

		// Current page details.
		pageParams.Current = page
		pageParams.Count = len(replies)

		// Set linked next/prev parameters.
		pageParams.Next = page.Next(lo, hi)
		pageParams.Prev = page.Prev(lo, hi)

		// Set the collection item property builder function.
		pageParams.Append = func(i int, itemsProp ap.ItemsPropertyBuilder) {
			// Get follower URI at index.
			status := replies[i]
			uri := status.URI

			// Parse URL object from URI.
			iri, err := url.Parse(uri)
			if err != nil {
				log.Errorf(ctx, "error parsing status uri %s: %v", uri, err)
				return
			}

			// Add to item property.
			itemsProp.AppendIRI(iri)
		}

		// Build AS collection page object from params.
		obj = ap.NewASOrderedCollectionPage(pageParams)
	}

	// Serialized the prepared object.
	data, err := ap.Serialize(obj)
	if err != nil {
		err := gtserror.Newf("error serializing: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return data, nil
}

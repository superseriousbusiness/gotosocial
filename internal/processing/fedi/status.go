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
func (p *Processor) StatusGet(ctx context.Context, requestedUser string, statusID string) (interface{}, gtserror.WithCode) {
	// Authenticate the incoming request, getting related user accounts.
	requester, receiver, errWithCode := p.authenticate(ctx, requestedUser)
	if errWithCode != nil {
		return nil, errWithCode
	}

	status, err := p.state.DB.GetStatusByID(ctx, statusID)
	if err != nil {
		return nil, gtserror.NewErrorNotFound(err)
	}

	if status.AccountID != receiver.ID {
		const text = "status does not belong to receiving account"
		return nil, gtserror.NewErrorNotFound(errors.New(text))
	}

	visible, err := p.filter.StatusVisible(ctx, requester, status)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	if !visible {
		const text = "status not vising to requesting account"
		return nil, gtserror.NewErrorNotFound(errors.New(text))
	}

	statusable, err := p.converter.StatusToAS(ctx, status)
	if err != nil {
		err := gtserror.Newf("error converting status: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	data, err := ap.Serialize(statusable)
	if err != nil {
		err := gtserror.Newf("error serializing status: %w", err)
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
	onlyOtherAccounts bool,
) (interface{}, gtserror.WithCode) {
	// Authenticate the incoming request, getting related user accounts.
	requester, receiver, errWithCode := p.authenticate(ctx, requestedUser)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// Get target status and ensure visible to requester.
	status, errWithCode := p.c.GetVisibleTargetStatus(ctx,
		requester,
		statusID,
	)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// Ensure status is by receiving account.
	if status.AccountID != receiver.ID {
		const text = "status does not belong to receiving account"
		return nil, gtserror.NewErrorNotFound(errors.New(text))
	}

	// Parse replies collection ID from status' URI with onlyOtherAccounts param.
	onlyOtherAccStr := "only_other_accounts=" + strconv.FormatBool(onlyOtherAccounts)
	collectionID, err := url.Parse(status.URI + "/replies?" + onlyOtherAccStr)
	if err != nil {
		err := gtserror.Newf("error parsing status uri %s: %w", status.URI, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Get *all* available replies for status (i.e. without paging).
	replies, err := p.state.DB.GetStatusReplies(ctx, status.ID)
	if err != nil {
		err := gtserror.Newf("error getting status replies: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if onlyOtherAccounts {
		// If 'onlyOtherAccounts' is set, drop all by original status author.
		replies = slices.DeleteFunc(replies, func(reply *gtsmodel.Status) bool {
			return reply.AccountID == status.AccountID
		})
	}

	// Reslice replies dropping all those invisible to requester.
	replies, err = p.filter.StatusesVisible(ctx, requester, replies)
	if err != nil {
		err := gtserror.Newf("error filtering status replies: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	var obj vocab.Type

	// Start AS collection params.
	var params ap.CollectionParams
	params.ID = collectionID
	params.Total = len(replies)

	if page == nil {
		// i.e. paging disabled, return collection
		// that links to first page (i.e. path below).
		params.Query = make(url.Values, 1)
		params.Query.Set("limit", "20") // enables paging
		obj = ap.NewASOrderedCollection(params)
	} else {
		// i.e. paging enabled

		// Page and reslice the replies according to given parameters.
		replies = paging.Page_PageFunc(page, replies, func(reply *gtsmodel.Status) string {
			return reply.ID
		})

		// page ID values.
		var lo, hi string

		if len(replies) > 0 {
			// Get the lowest and highest
			// ID values, used for paging.
			lo = replies[len(replies)-1].ID
			hi = replies[0].ID
		}

		// Start AS collection page params.
		var pageParams ap.CollectionPageParams
		pageParams.CollectionParams = params

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

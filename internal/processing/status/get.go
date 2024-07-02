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

package status

import (
	"context"
	"slices"
	"strings"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	statusfilter "github.com/superseriousbusiness/gotosocial/internal/filter/status"
	"github.com/superseriousbusiness/gotosocial/internal/filter/usermute"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// HistoryGet gets edit history for the target status, taking account of privacy settings and blocks etc.
// TODO: currently this just returns the latest version of the status.
func (p *Processor) HistoryGet(ctx context.Context, requestingAccount *gtsmodel.Account, targetStatusID string) ([]*apimodel.StatusEdit, gtserror.WithCode) {
	targetStatus, errWithCode := p.c.GetVisibleTargetStatus(ctx,
		requestingAccount,
		targetStatusID,
		nil, // default freshness
	)
	if errWithCode != nil {
		return nil, errWithCode
	}

	apiStatus, errWithCode := p.c.GetAPIStatus(ctx, requestingAccount, targetStatus)
	if errWithCode != nil {
		return nil, errWithCode
	}

	return []*apimodel.StatusEdit{
		{
			Content:          apiStatus.Content,
			SpoilerText:      apiStatus.SpoilerText,
			Sensitive:        apiStatus.Sensitive,
			CreatedAt:        util.FormatISO8601(targetStatus.UpdatedAt),
			Account:          apiStatus.Account,
			Poll:             apiStatus.Poll,
			MediaAttachments: apiStatus.MediaAttachments,
			Emojis:           apiStatus.Emojis,
		},
	}, nil
}

// Get gets the given status, taking account of privacy settings and blocks etc.
func (p *Processor) Get(ctx context.Context, requestingAccount *gtsmodel.Account, targetStatusID string) (*apimodel.Status, gtserror.WithCode) {
	targetStatus, errWithCode := p.c.GetVisibleTargetStatus(ctx,
		requestingAccount,
		targetStatusID,
		nil, // default freshness
	)
	if errWithCode != nil {
		return nil, errWithCode
	}

	return p.c.GetAPIStatus(ctx, requestingAccount, targetStatus)
}

// SourceGet returns the *apimodel.StatusSource version of the targetStatusID.
// Status must belong to the requester, and must not be a boost.
func (p *Processor) SourceGet(ctx context.Context, requestingAccount *gtsmodel.Account, targetStatusID string) (*apimodel.StatusSource, gtserror.WithCode) {
	targetStatus, errWithCode := p.c.GetVisibleTargetStatus(ctx,
		requestingAccount,
		targetStatusID,
		nil, // default freshness
	)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// Redirect to wrapped status if boost.
	targetStatus, errWithCode = p.c.UnwrapIfBoost(
		ctx,
		requestingAccount,
		targetStatus,
	)
	if errWithCode != nil {
		return nil, errWithCode
	}

	if targetStatus.AccountID != requestingAccount.ID {
		err := gtserror.Newf(
			"status %s does not belong to account %s",
			targetStatusID, requestingAccount.ID,
		)
		return nil, gtserror.NewErrorNotFound(err)
	}

	statusSource, err := p.converter.StatusToAPIStatusSource(ctx, targetStatus)
	if err != nil {
		err = gtserror.Newf("error converting status: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}
	return statusSource, nil
}

// WebGet gets the given status for web use, taking account of privacy settings.
func (p *Processor) WebGet(ctx context.Context, targetStatusID string) (*apimodel.Status, gtserror.WithCode) {
	targetStatus, errWithCode := p.c.GetVisibleTargetStatus(ctx,
		nil, // requester
		targetStatusID,
		nil, // default freshness
	)
	if errWithCode != nil {
		return nil, errWithCode
	}

	webStatus, err := p.converter.StatusToWebStatus(ctx, targetStatus, nil)
	if err != nil {
		err = gtserror.Newf("error converting status: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}
	return webStatus, nil
}

func (p *Processor) contextGet(
	ctx context.Context,
	requestingAccount *gtsmodel.Account,
	targetStatusID string,
	convert func(context.Context, *gtsmodel.Status, *gtsmodel.Account) (*apimodel.Status, error),
) (*apimodel.Context, gtserror.WithCode) {
	targetStatus, errWithCode := p.c.GetVisibleTargetStatus(ctx,
		requestingAccount,
		targetStatusID,
		nil, // default freshness
	)
	if errWithCode != nil {
		return nil, errWithCode
	}

	parents, err := p.state.DB.GetStatusParents(ctx, targetStatus)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	var ancestors []*apimodel.Status
	for _, status := range parents {
		if v, err := p.visFilter.StatusVisible(ctx, requestingAccount, status); err == nil && v {
			apiStatus, err := convert(ctx, status, requestingAccount)
			if err == nil {
				ancestors = append(ancestors, apiStatus)
			}
		}
	}

	slices.SortFunc(ancestors, func(lhs, rhs *apimodel.Status) int {
		return strings.Compare(lhs.ID, rhs.ID)
	})

	children, err := p.state.DB.GetStatusChildren(ctx, targetStatus.ID)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	var descendants []*apimodel.Status
	for _, status := range children {
		if v, err := p.visFilter.StatusVisible(ctx, requestingAccount, status); err == nil && v {
			apiStatus, err := convert(ctx, status, requestingAccount)
			if err == nil {
				descendants = append(descendants, apiStatus)
			}
		}
	}

	TopoSort(descendants, targetStatus.AccountID)

	context := &apimodel.Context{
		Ancestors:   make([]apimodel.Status, 0, len(ancestors)),
		Descendants: make([]apimodel.Status, 0, len(descendants)),
	}
	for _, ancestor := range ancestors {
		context.Ancestors = append(context.Ancestors, *ancestor)
	}
	for _, descendant := range descendants {
		context.Descendants = append(context.Descendants, *descendant)
	}

	return context, nil
}

// TopoSort sorts statuses topologically, by self-reply, and by ID.
// Can handle cycles but the output order will be arbitrary.
// (But if there are cycles, something went wrong upstream.)
func TopoSort(apiStatuses []*apimodel.Status, targetAccountID string) {
	if len(apiStatuses) == 0 {
		return
	}

	// Map of status IDs to statuses.
	lookup := make(map[string]*apimodel.Status, len(apiStatuses))
	for _, apiStatus := range apiStatuses {
		lookup[apiStatus.ID] = apiStatus
	}

	// Tree of statuses to their children.
	// The nil status may have children: any who don't have a parent, or whose parent isn't in the input.
	tree := make(map[*apimodel.Status][]*apimodel.Status, len(apiStatuses))
	for _, apiStatus := range apiStatuses {
		var parent *apimodel.Status
		if apiStatus.InReplyToID != nil {
			parent = lookup[*apiStatus.InReplyToID]
		}
		tree[parent] = append(tree[parent], apiStatus)
	}

	// Sort children of each status by self-reply status and then ID, *in reverse*.
	isSelfReply := func(apiStatus *apimodel.Status) bool {
		return apiStatus.GetAccountID() == targetAccountID &&
			apiStatus.InReplyToAccountID != nil &&
			*apiStatus.InReplyToAccountID == targetAccountID
	}
	for id, children := range tree {
		slices.SortFunc(children, func(lhs, rhs *apimodel.Status) int {
			lhsIsContextSelfReply := isSelfReply(lhs)
			rhsIsContextSelfReply := isSelfReply(rhs)

			if lhsIsContextSelfReply && !rhsIsContextSelfReply {
				return 1
			} else if !lhsIsContextSelfReply && rhsIsContextSelfReply {
				return -1
			}

			return -strings.Compare(lhs.ID, rhs.ID)
		})
		tree[id] = children
	}

	// Traverse the tree using preorder depth-first search, topologically sorting the statuses.
	stack := make([]*apimodel.Status, 1, len(tree))
	apiStatusIndex := 0
	for len(stack) > 0 {
		parent := stack[len(stack)-1]
		children := tree[parent]

		if len(children) == 0 {
			// Remove this node from the tree.
			delete(tree, parent)
			// Go back to this node's parent.
			stack = stack[:len(stack)-1]
			continue
		}

		// Remove the last child entry (the first in sorted order).
		child := children[len(children)-1]
		tree[parent] = children[:len(children)-1]

		// Explore its children next.
		stack = append(stack, child)

		// Overwrite the next entry of the input slice.
		apiStatuses[apiStatusIndex] = child
		apiStatusIndex++
	}

	// There should only be nodes left in the tree in the event of a cycle.
	// Append them to the end in arbitrary order.
	// This ensures that the slice of statuses has no duplicates.
	for node := range tree {
		apiStatuses[apiStatusIndex] = node
		apiStatusIndex++
	}
}

// ContextGet returns the context (previous and following posts) from the given status ID.
func (p *Processor) ContextGet(ctx context.Context, requestingAccount *gtsmodel.Account, targetStatusID string) (*apimodel.Context, gtserror.WithCode) {
	filters, err := p.state.DB.GetFiltersForAccountID(ctx, requestingAccount.ID)
	if err != nil {
		err = gtserror.Newf("couldn't retrieve filters for account %s: %w", requestingAccount.ID, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	mutes, err := p.state.DB.GetAccountMutes(gtscontext.SetBarebones(ctx), requestingAccount.ID, nil)
	if err != nil {
		err = gtserror.Newf("couldn't retrieve mutes for account %s: %w", requestingAccount.ID, err)
		return nil, gtserror.NewErrorInternalError(err)
	}
	compiledMutes := usermute.NewCompiledUserMuteList(mutes)

	convert := func(ctx context.Context, status *gtsmodel.Status, requestingAccount *gtsmodel.Account) (*apimodel.Status, error) {
		return p.converter.StatusToAPIStatus(ctx, status, requestingAccount, statusfilter.FilterContextThread, filters, compiledMutes)
	}
	return p.contextGet(ctx, requestingAccount, targetStatusID, convert)
}

// WebContextGet is like ContextGet, but is explicitly
// for viewing statuses via the unauthenticated web UI.
//
// TODO: a more advanced threading model could be implemented here.
func (p *Processor) WebContextGet(ctx context.Context, targetStatusID string) (*apimodel.Context, gtserror.WithCode) {
	return p.contextGet(ctx, nil, targetStatusID, p.converter.StatusToWebStatus)
}

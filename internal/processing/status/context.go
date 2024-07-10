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
)

// internalThreadContext is like
// *apimodel.ThreadContext, but
// for internal use only.
type internalThreadContext struct {
	targetStatus *gtsmodel.Status
	ancestors    []*gtsmodel.Status
	descendants  []*gtsmodel.Status
}

func (p *Processor) contextGet(
	ctx context.Context,
	requester *gtsmodel.Account,
	targetStatusID string,
) (*internalThreadContext, gtserror.WithCode) {
	targetStatus, errWithCode := p.c.GetVisibleTargetStatus(ctx,
		requester,
		targetStatusID,
		nil, // default freshness
	)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// Don't generate thread for boosts/reblogs.
	if targetStatus.BoostOfID != "" {
		err := gtserror.New("target status is a boost wrapper / reblog")
		return nil, gtserror.NewErrorNotFound(err)
	}

	// Fetch up to the top of the thread.
	ancestors, err := p.state.DB.GetStatusParents(ctx, targetStatus)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Do a simple ID sort of ancestors
	// to arrange them by creation time.
	slices.SortFunc(ancestors, func(lhs, rhs *gtsmodel.Status) int {
		return strings.Compare(lhs.ID, rhs.ID)
	})

	// Fetch down to the bottom of the thread.
	descendants, err := p.state.DB.GetStatusChildren(ctx, targetStatus.ID)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Topographically sort descendants,
	// to place them in sub-threads.
	TopoSort(descendants, targetStatus.AccountID)

	return &internalThreadContext{
		targetStatus: targetStatus,
		ancestors:    ancestors,
		descendants:  descendants,
	}, nil
}

// Returns true if status counts as a self-reply
// *within the current context*, ie., status is a
// self-reply by contextAcctID to contextAcctID.
func isSelfReply(
	status *gtsmodel.Status,
	contextAcctID string,
) bool {
	if status.AccountID != contextAcctID {
		// Doesn't belong
		// to context acct.
		return false
	}

	return status.InReplyToAccountID == contextAcctID
}

// TopoSort sorts the given slice of *descendant*
// statuses topologically, by self-reply, and by ID.
//
// "contextAcctID" should be the ID of the account that owns
// the status the thread context is being constructed around.
//
// Can handle cycles but the output order will be arbitrary.
// (But if there are cycles, something went wrong upstream.)
func TopoSort(
	statuses []*gtsmodel.Status,
	contextAcctID string,
) {
	if len(statuses) == 0 {
		return
	}

	// Simple map of status IDs to statuses.
	//
	// Eg.,
	//
	//	01J2BC6DQ37A6SQPAVCZ2BYSTN: *gtsmodel.Status
	//	01J2BC8GT9THMPWMCAZYX48PXJ: *gtsmodel.Status
	//	01J2BC8M56C5ZAH76KN93D7F0W: *gtsmodel.Status
	//	01J2BC90QNW65SM2F89R5M0NGE: *gtsmodel.Status
	//	01J2BC916YVX6D6Q0SA30JV82D: *gtsmodel.Status
	//	01J2BC91J2Y75D4Z3EEDF3DYAV: *gtsmodel.Status
	//	01J2BC91VBVPBZACZMDA7NEZY9: *gtsmodel.Status
	//	01J2BCMM3CXQE70S831YPWT48T: *gtsmodel.Status
	lookup := make(map[string]*gtsmodel.Status, len(statuses))
	for _, status := range statuses {
		lookup[status.ID] = status
	}

	// Tree of statuses to their children.
	//
	// The nil status may have children: any who don't
	// have a parent, or whose parent isn't in the input.
	//
	// Eg.,
	//
	//	*gtsmodel.Status (01J2BC916YVX6D6Q0SA30JV82D): [     <- parent2 (1 child)
	//		*gtsmodel.Status (01J2BC91J2Y75D4Z3EEDF3DYAV)    <- p2 child1
	//	],
	//	*gtsmodel.Status (01J2BC6DQ37A6SQPAVCZ2BYSTN): [     <- parent1 (3 children)
	//		*gtsmodel.Status (01J2BC8M56C5ZAH76KN93D7F0W)    <- p1 child3  |
	//		*gtsmodel.Status (01J2BC90QNW65SM2F89R5M0NGE)    <- p1 child1  |- Not sorted
	//		*gtsmodel.Status (01J2BC8GT9THMPWMCAZYX48PXJ)    <- p1 child2  |
	//	],
	//	*gtsmodel.Status (01J2BC91VBVPBZACZMDA7NEZY9): [     <- parent3 (no children 😢)
	//	]
	//	*gtsmodel.Status (nil): [                            <- parent4 (nil status)
	//		*gtsmodel.Status (01J2BCMM3CXQE70S831YPWT48T)    <- p4 child1 (no parent 😢)
	//	]
	tree := make(map[*gtsmodel.Status][]*gtsmodel.Status, len(statuses))
	for _, status := range statuses {
		var parent *gtsmodel.Status
		if status.InReplyToID != "" {
			// May be nil if reply is missing.
			parent = lookup[status.InReplyToID]
		}

		tree[parent] = append(tree[parent], status)
	}

	// Sort children of each parent by self-reply status and then ID, *in reverse*.
	// This results in the tree looking something like:
	//
	//	*gtsmodel.Status (01J2BC916YVX6D6Q0SA30JV82D): [     <- parent2 (1 child)
	//		*gtsmodel.Status (01J2BC91J2Y75D4Z3EEDF3DYAV)    <- p2 child1
	//	],
	//	*gtsmodel.Status (01J2BC6DQ37A6SQPAVCZ2BYSTN): [     <- parent1 (3 children)
	//		*gtsmodel.Status (01J2BC90QNW65SM2F89R5M0NGE)    <- p1 child1  |
	//		*gtsmodel.Status (01J2BC8GT9THMPWMCAZYX48PXJ)    <- p1 child2  |- Sorted
	//		*gtsmodel.Status (01J2BC8M56C5ZAH76KN93D7F0W)    <- p1 child3  |
	//	],
	//	*gtsmodel.Status (01J2BC91VBVPBZACZMDA7NEZY9): [     <- parent3 (no children 😢)
	//	],
	//	*gtsmodel.Status (nil): [                            <- parent4 (nil status)
	//		*gtsmodel.Status (01J2BCMM3CXQE70S831YPWT48T)    <- p4 child1 (no parent 😢)
	//	]
	for id, children := range tree {
		slices.SortFunc(children, func(lhs, rhs *gtsmodel.Status) int {
			lhsIsSelfReply := isSelfReply(lhs, contextAcctID)
			rhsIsSelfReply := isSelfReply(rhs, contextAcctID)

			if lhsIsSelfReply && !rhsIsSelfReply {
				// lhs is the end
				// of a sub-thread.
				return 1
			} else if !lhsIsSelfReply && rhsIsSelfReply {
				// lhs is the start
				// of a sub-thread.
				return -1
			}

			// Sort by created-at descending.
			return -strings.Compare(lhs.ID, rhs.ID)
		})
		tree[id] = children
	}

	// Traverse the tree using preorder depth-first
	// search, topologically sorting the statuses
	// until the stack is empty.
	//
	// The stack starts with one nil status in it
	// to account for potential nil key in the tree,
	// which means the below "for" loop will always
	// iterate at least once.
	//
	// The result will look something like:
	//
	//	*gtsmodel.Status (01J2BC6DQ37A6SQPAVCZ2BYSTN)   <- parent1 (3 children)
	//	*gtsmodel.Status (01J2BC90QNW65SM2F89R5M0NGE)   <- p1 child1  |
	//	*gtsmodel.Status (01J2BC8GT9THMPWMCAZYX48PXJ)   <- p1 child2  |- Sorted
	//	*gtsmodel.Status (01J2BC8M56C5ZAH76KN93D7F0W)   <- p1 child3  |
	//	*gtsmodel.Status (01J2BC916YVX6D6Q0SA30JV82D)   <- parent2 (1 child)
	//	*gtsmodel.Status (01J2BC91J2Y75D4Z3EEDF3DYAV)   <- p2 child1
	//	*gtsmodel.Status (01J2BC91VBVPBZACZMDA7NEZY9)   <- parent3 (no children 😢)
	//	*gtsmodel.Status (01J2BCMM3CXQE70S831YPWT48T)   <- p4 child1 (no parent 😢)

	stack := make([]*gtsmodel.Status, 1, len(tree))
	statusIndex := 0
	for len(stack) > 0 {
		parent := stack[len(stack)-1]
		children := tree[parent]

		if len(children) == 0 {
			// No (more) children so we're
			// done with this node.
			// Remove it from the tree.
			delete(tree, parent)

			// Also remove this node from
			// the stack, then continue
			// from its parent.
			stack = stack[:len(stack)-1]

			continue
		}

		// Pop the last child entry
		// (the first in sorted order).
		child := children[len(children)-1]
		tree[parent] = children[:len(children)-1]

		// Explore its children next.
		stack = append(stack, child)

		// Overwrite the next entry of the input slice.
		statuses[statusIndex] = child
		statusIndex++
	}

	// There should only be orphan nodes remaining
	// (or other nodes in the event of a cycle).
	// Append them to the end in arbitrary order.
	//
	// The fact we put them in a map first just
	// ensures the slice of statuses has no duplicates.
	for orphan := range tree {
		statuses[statusIndex] = orphan
		statusIndex++
	}
}

// ContextGet returns the context (previous
// and following posts) from the given status ID.
func (p *Processor) ContextGet(
	ctx context.Context,
	requester *gtsmodel.Account,
	targetStatusID string,
) (*apimodel.ThreadContext, gtserror.WithCode) {
	// Retrieve filters as they affect
	// what should be shown to requester.
	filters, err := p.state.DB.GetFiltersForAccountID(
		ctx, // Populate filters.
		requester.ID,
	)
	if err != nil {
		err = gtserror.Newf(
			"couldn't retrieve filters for account %s: %w",
			requester.ID, err,
		)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Retrieve mutes as they affect
	// what should be shown to requester.
	mutes, err := p.state.DB.GetAccountMutes(
		// No need to populate mutes,
		// IDs are enough here.
		gtscontext.SetBarebones(ctx),
		requester.ID,
		nil, // No paging - get all.
	)
	if err != nil {
		err = gtserror.Newf(
			"couldn't retrieve mutes for account %s: %w",
			requester.ID, err,
		)
		return nil, gtserror.NewErrorInternalError(err)
	}

	convert := func(
		ctx context.Context,
		status *gtsmodel.Status,
		requestingAccount *gtsmodel.Account,
	) (*apimodel.Status, error) {
		return p.converter.StatusToAPIStatus(
			ctx,
			status,
			requestingAccount,
			statusfilter.FilterContextThread,
			filters,
			usermute.NewCompiledUserMuteList(mutes),
		)
	}

	// Retrieve the thread context.
	threadContext, errWithCode := p.contextGet(
		ctx,
		requester,
		targetStatusID,
	)
	if errWithCode != nil {
		return nil, errWithCode
	}

	apiContext := &apimodel.ThreadContext{
		Ancestors:   make([]apimodel.Status, 0, len(threadContext.ancestors)),
		Descendants: make([]apimodel.Status, 0, len(threadContext.descendants)),
	}

	// Convert ancestors + filter
	// out ones that aren't visible.
	for _, status := range threadContext.ancestors {
		if v, err := p.filter.StatusVisible(ctx, requester, status); err == nil && v {
			status, err := convert(ctx, status, requester)
			if err == nil {
				apiContext.Ancestors = append(apiContext.Ancestors, *status)
			}
		}
	}

	// Convert descendants + filter
	// out ones that aren't visible.
	for _, status := range threadContext.descendants {
		if v, err := p.filter.StatusVisible(ctx, requester, status); err == nil && v {
			status, err := convert(ctx, status, requester)
			if err == nil {
				apiContext.Descendants = append(apiContext.Descendants, *status)
			}
		}
	}

	return apiContext, nil
}

// WebContextGet is like ContextGet, but is explicitly
// for viewing statuses via the unauthenticated web UI.
//
// The returned statuses in the ThreadContext will be
// populated with ThreadMeta annotations for more easily
// positioning the status in a web view of a thread.
func (p *Processor) WebContextGet(
	ctx context.Context,
	targetStatusID string,
) (*apimodel.ThreadContext, gtserror.WithCode) {
	// Retrieve the thread context.
	threadContext, errWithCode := p.contextGet(
		ctx,
		nil, // No authed requester.
		targetStatusID,
	)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// Recreate the whole thread so we can go
	// through it again add ThreadMeta annotations
	// from the perspective of the OG status.
	// nolint:gocritic
	wholeThread := append(
		// Ancestors at the beginning.
		threadContext.ancestors,
		append(
			// Target status in the middle.
			[]*gtsmodel.Status{threadContext.targetStatus},
			// Descendants at the end.
			threadContext.descendants...,
		)...,
	)

	// Start preparing API context.
	apiContext := &apimodel.ThreadContext{
		Ancestors:   make([]apimodel.Status, 0, len(threadContext.ancestors)),
		Descendants: make([]apimodel.Status, 0, len(threadContext.descendants)),
	}

	var (
		// Metadata about the thread.
		meta = new(apimodel.WebThreadMeta)

		// Who the current thread "belongs" to,
		// ie., who created first post in the thread.
		contextAcctID = wholeThread[0].AccountID

		// Position of target status in wholeThread,
		// we put it on top of ancestors.
		targetStatusIdx = len(threadContext.ancestors)

		// Position from which we should add
		// to descendants and not to ancestors.
		descendantsIdx = targetStatusIdx + 1

		// Whether we've reached
		// end of "main" thread yet.
		foundMainThreadEnd bool

		// Index in wholeThread where
		// the "main" thread ends.
		firstReplyIdx int

		// We should mark the next **VISIBLE**
		// reply as the first reply.
		shouldMarkNextReply bool
	)

	for idx, status := range wholeThread {
		if !foundMainThreadEnd {
			// Haven't reached end
			// of "main" thread yet.
			//
			// First post in wholeThread can't
			// be a self reply, so ignore it.
			//
			// That aside, first non-self-reply
			// in wholeThread means the "main"
			// thread is now over.
			if idx != 0 && !isSelfReply(status, contextAcctID) {
				// Jot some stuff down.
				foundMainThreadEnd = true
				firstReplyIdx = idx
				shouldMarkNextReply = true
			}
		}

		// Ensure status is actually
		// visible to just anyone.
		v, err := p.filter.StatusVisible(ctx, nil, status)
		if err != nil || !v {
			// Skip this one.
			if !foundMainThreadEnd {
				meta.WebThreadHidden++
			} else {
				meta.WebThreadRepliesHidden++
			}
			continue
		}

		// Prepare status to add to thread context.
		apiStatus, err := p.converter.StatusToWebStatus(ctx, status, nil)
		if err != nil {
			continue
		}

		if shouldMarkNextReply {
			// This is the first visible
			// "reply / comment".
			apiStatus.WebThreadFirstReply = true
			shouldMarkNextReply = false
		}

		switch {
		case idx == targetStatusIdx:
			// This is the target status itself.
			apiContext.WebTargetStatus = apiStatus

		case idx < descendantsIdx:
			// Haven't reached descendants yet,
			// so this must be an ancestor.
			apiContext.Ancestors = append(
				apiContext.Ancestors,
				*apiStatus,
			)

		default:
			// We're in descendants town now.
			apiContext.Descendants = append(
				apiContext.Descendants,
				*apiStatus,
			)
		}
	}

	// Now we've gone through the whole
	// thread, we can add some additional info.

	// Length of the "main" thread. If there are
	// replies then it's up to where the replies
	// start, otherwise it's the whole thing.
	if foundMainThreadEnd {
		meta.WebThreadLength = firstReplyIdx
	} else {
		meta.WebThreadLength = len(wholeThread)
	}

	// Jot down number of hidden posts so template doesn't have to do it.
	meta.WebThreadShown = meta.WebThreadLength - meta.WebThreadHidden

	// Number of replies is equal to number
	// of statuses in the thread that aren't
	// part of the "main" thread.
	meta.WebThreadReplies = len(wholeThread) - meta.WebThreadLength

	// Jot down number of hidden replies so template doesn't have to do it.
	meta.WebThreadRepliesShown = meta.WebThreadReplies - meta.WebThreadRepliesHidden

	// Return the finished context.
	apiContext.WebThreadMeta = meta
	return apiContext, nil
}

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
	"errors"
	"slices"
	"strings"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	statusfilter "github.com/superseriousbusiness/gotosocial/internal/filter/status"
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

	// Retrieve the full thread context.
	threadContext, errWithCode := p.contextGet(
		ctx,
		requester,
		targetStatusID,
	)
	if errWithCode != nil {
		return nil, errWithCode
	}

	var apiContext apimodel.ThreadContext

	// Convert and filter the thread context ancestors.
	apiContext.Ancestors = p.c.GetVisibleAPIStatuses(ctx,
		requester,
		threadContext.ancestors,
		statusfilter.FilterContextThread,
		filters,
		mutes,
	)

	// Convert and filter the thread context descendants
	apiContext.Descendants = p.c.GetVisibleAPIStatuses(ctx,
		requester,
		threadContext.descendants,
		statusfilter.FilterContextThread,
		filters,
		mutes,
	)

	return &apiContext, nil
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
) (*apimodel.WebThreadContext, gtserror.WithCode) {
	// Retrieve the internal thread context.
	iCtx, errWithCode := p.contextGet(
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
		iCtx.ancestors,
		append(
			// Target status in the middle.
			[]*gtsmodel.Status{iCtx.targetStatus},
			// Descendants at the end.
			iCtx.descendants...,
		)...,
	)

	// Start preparing web context.
	wCtx := &apimodel.WebThreadContext{
		Statuses: make([]*apimodel.WebStatus, 0, len(wholeThread)),
	}

	var (
		threadLength = len(wholeThread)

		// Track how much each reply status
		// should be indented (if at all).
		statusIndents = make(map[string]int, threadLength)

		// Who the current thread "belongs" to,
		// ie., who created first post in the thread.
		contextAcctID = wholeThread[0].AccountID

		// Whether we've reached end of "main"
		// thread and are now looking at replies.
		inReplies bool

		// Index in wholeThread
		// where replies begin.
		firstReplyIdx int

		// We should mark the next **VISIBLE**
		// reply as the first reply.
		markNextVisibleAsFirstReply bool

		// Map of statuses that didn't pass visi
		// checks and won't be shown via the web.
		hiddenStatuses = make(map[string]struct{})
	)

	for idx, status := range wholeThread {
		if !inReplies {
			// Check if we've reached replies
			// by looking for the first status
			// that's not a self-reply, ie.,
			// not a post in the "main" thread.
			switch {
			case idx == 0:
				// First post in wholeThread can't
				// be a self reply anyway because
				// it (very likely) doesn't reply
				// to anything, so ignore it.

			case !isSelfReply(status, contextAcctID):
				// This is not a self-reply, which
				// means it's a reply from another
				// account. So, replies start here.
				inReplies = true
				firstReplyIdx = idx
				markNextVisibleAsFirstReply = true
			}
		}

		// Ensure status is actually visible to just
		// anyone, and hide / don't include it if not.
		//
		// Include a check to see if the parent status
		// is hidden; if so, we shouldn't show the child
		// as it leads to weird-looking threading where
		// a status seems to reply to nothing.
		_, parentHidden := hiddenStatuses[status.InReplyToID]
		v, err := p.visFilter.StatusVisible(ctx, nil, status)
		if err != nil || !v || parentHidden {
			// If this is the main status whose
			// context we're looking for, and it's
			// not visible for whatever reason, we
			// should just return a 404 here, as we
			// can't meaningfully render the thread.
			if status.ID == targetStatusID {
				var thisErr error
				switch {
				case err != nil:
					thisErr = gtserror.Newf("error checking visibility of target status: %w", err)

				case !v:
					const errText = "target status not visible"
					thisErr = gtserror.New(errText)

				case parentHidden:
					const errText = "target status parent is hidden"
					thisErr = gtserror.New(errText)
				}

				return nil, gtserror.NewErrorNotFound(thisErr)
			}

			// This isn't the main status whose
			// context we're looking for, just
			// your standard not-visible status,
			// so add it to the count + map.
			if !inReplies {
				// Main thread entry hidden.
				wCtx.ThreadHidden++
			} else {
				// Reply hidden.
				wCtx.ThreadRepliesHidden++
			}

			hiddenStatuses[status.ID] = struct{}{}
			continue
		}

		// Prepare visible status to add to thread context.
		webStatus, err := p.converter.StatusToWebStatus(ctx, status)
		if err != nil {
			hiddenStatuses[status.ID] = struct{}{}
			continue
		}

		if markNextVisibleAsFirstReply {
			// This is the first visible
			// "reply / comment", so the
			// little "x amount of replies"
			// header should go above this.
			webStatus.ThreadFirstReply = true
			markNextVisibleAsFirstReply = false
		}

		// If this is a reply, work out the indent of
		// this status based on its parent's indent.
		if inReplies {
			parentIndent, ok := statusIndents[status.InReplyToID]
			switch {
			case !ok:
				// No parent with
				// indent, start at 0.
				webStatus.Indent = 0

			case isSelfReply(status, status.AccountID):
				// Self reply, so indent at same
				// level as own replied-to status.
				webStatus.Indent = parentIndent

			case parentIndent == 5:
				// Already indented as far as we
				// can go to keep things readable
				// on thin screens, so just keep
				// parent's indent.
				webStatus.Indent = parentIndent

			default:
				// Reply to someone else who's
				// indented, but not to TO THE MAX.
				// Indent by another one.
				webStatus.Indent = parentIndent + 1
			}

			// Store the indent for this status.
			statusIndents[status.ID] = webStatus.Indent
		}

		if webStatus.ID == targetStatusID {
			// This is the og
			// thread context status.
			webStatus.ThreadContextStatus = true
			wCtx.Status = webStatus
		}

		wCtx.Statuses = append(wCtx.Statuses, webStatus)
	}

	// Now we've gone through the whole
	// thread, we can add some additional info.

	// Length of the "main" thread. If there are
	// visible replies then it's up to where the
	// replies start, else it's the whole thing.
	if inReplies {
		wCtx.ThreadLength = firstReplyIdx
	} else {
		wCtx.ThreadLength = threadLength
	}

	// Jot down number of "main" thread entries shown.
	wCtx.ThreadShown = wCtx.ThreadLength - wCtx.ThreadHidden

	// If there's no posts visible in the
	// "main" thread we shouldn't show replies
	// via the web as that's just weird.
	if wCtx.ThreadShown < 1 {
		const text = "no statuses visible in main thread"
		return nil, gtserror.NewErrorNotFound(errors.New(text))
	}

	// Mark the last "main" visible status.
	wCtx.Statuses[wCtx.ThreadShown-1].ThreadLastMain = true

	// Number of replies is equal to number
	// of statuses in the thread that aren't
	// part of the "main" thread.
	wCtx.ThreadReplies = threadLength - wCtx.ThreadLength

	// Jot down number of "replies" shown.
	wCtx.ThreadRepliesShown = wCtx.ThreadReplies - wCtx.ThreadRepliesHidden

	// Return the finished context.
	return wCtx, nil
}

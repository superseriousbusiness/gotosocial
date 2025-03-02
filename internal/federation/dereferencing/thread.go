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
	"net/http"
	"net/url"

	"codeberg.org/gruf/go-kv"
	"codeberg.org/superseriousbusiness/activity/pub"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

// maxIter defines how many iterations of descendants or
// ancesters we are willing to follow before returning error.
const maxIter = 512

// dereferenceThread handles dereferencing status thread after
// fetch. Passing off appropriate parts to be enqueued for async
// processing, or handling some parts synchronously when required.
func (d *Dereferencer) dereferenceThread(
	ctx context.Context,
	requestUser string,
	uri *url.URL,
	status *gtsmodel.Status,
	statusable ap.Statusable,
	isNew bool,
) {
	if isNew {
		// This is a new status that we need the ancestors of in
		// order to determine visibility. Perform the initial part
		// of thread dereferencing, i.e. parents, synchronously.
		err := d.DereferenceStatusAncestors(ctx, requestUser, status)
		if err != nil {
			log.Error(ctx, err)
		}

		// Enqueue dereferencing remaining status thread, (children), asychronously .
		d.state.Workers.Dereference.Queue.Push(func(ctx context.Context) {
			if err := d.DereferenceStatusDescendants(ctx, requestUser, uri, statusable); err != nil {
				log.Error(ctx, err)
			}
		})
	} else {
		// This is an existing status, dereference the WHOLE thread asynchronously.
		d.state.Workers.Dereference.Queue.Push(func(ctx context.Context) {
			if err := d.DereferenceStatusAncestors(ctx, requestUser, status); err != nil {
				log.Error(ctx, err)
			}
			if err := d.DereferenceStatusDescendants(ctx, requestUser, uri, statusable); err != nil {
				log.Error(ctx, err)
			}
		})
	}
}

// DereferenceStatusAncestors iterates upwards from the given status, using InReplyToURI, to ensure that as many parent statuses as possible are dereferenced.
func (d *Dereferencer) DereferenceStatusAncestors(ctx context.Context, username string, status *gtsmodel.Status) error {
	// Start log entry with fields
	l := log.WithContext(ctx).
		WithFields(kv.Fields{
			{"username", username},
			{"original", status.URI},
		}...)

	// Keep track of already dereferenced statuses
	// for this ancestor thread to prevent recursion.
	derefdStatuses := make(map[string]struct{}, 10)

	// Mark given status as the one
	// we're currently working on.
	current := status

	for i := 0; i < maxIter; i++ {
		if current.InReplyToURI == "" {
			// Status has no parent, we've
			// reached the top of the chain.
			return nil
		}

		// Apparent current parent URI to log fields.
		l = l.WithField("parent", current.InReplyToURI)
		l.Trace("following status ancestor")

		// Check whether this parent has already been deref'd.
		if _, ok := derefdStatuses[current.InReplyToURI]; ok {
			l.Warn("self referencing status ancestor")
			return nil
		}

		// Add this status's parent URI to map of deref'd.
		derefdStatuses[current.InReplyToURI] = struct{}{}

		// Parse status parent URI for later use.
		uri, err := url.Parse(current.InReplyToURI)
		if err != nil {
			l.Warnf("invalid uri: %v", err)
			return nil
		}

		// Fetch parent status by current's reply URI, this handles
		// case of existing (updating if necessary) or a new status.
		parent, _, _, err := d.getStatusByURI(ctx, username, uri)

		// Check for a returned HTTP code via error.
		switch code := gtserror.StatusCode(err); {

		// 404 may indicate deletion, but can also
		// indicate that we don't have permission to
		// view the status (it's followers-only and
		// we don't follow, for example).
		case code == http.StatusNotFound:

			// If this reply is followers-only or stricter,
			// we can safely assume the status it replies
			// to is also followers only or stricter.
			//
			// In this case we should leave the inReplyTo
			// URI in place for visibility filtering,
			// and just return since we can go no further.
			if status.Visibility == gtsmodel.VisibilityFollowersOnly ||
				status.Visibility == gtsmodel.VisibilityMutualsOnly ||
				status.Visibility == gtsmodel.VisibilityDirect {
				return nil
			}

			// If the reply is public or unlisted then
			// likely the replied-to status is/was public
			// or unlisted and has indeed been deleted,
			// fall through to the Gone case to clean up.
			fallthrough

		// Gone (410) definitely indicates deletion.
		// Update the status to remove references to
		// the now-gone parent.
		case code == http.StatusGone:
			l.Trace("status orphaned")
			current.InReplyTo = nil
			current.InReplyToAccount = nil
			return d.updateStatusParent(ctx,
				current,
				"", // status ID
				"", // status URI
				"", // account ID
			)

		// An error was returned for a status during
		// an attempted NEW dereference, return here.
		//
		// NOTE: this will catch all cases of a nil
		// parent, all cases below can safely assume
		// a non-nil parent in their code logic.
		case err != nil && parent == nil:
			return gtserror.Newf("error dereferencing new %s: %w", current.InReplyToURI, err)

		// An error was returned for an existing parent,
		// we simply treat this as a temporary situation.
		case err != nil:
			l.Errorf("error getting parent: %v", err)
		}

		// Start a new switch case
		// as the following scenarios
		// are possible with / without
		// any returned error.
		switch {

		// The current status is using an indirect URL
		// in order to reference the parent. This is just
		// weird and broken... Leave the URI in place but
		// don't link the statuses via database IDs as it
		// could cause all sorts of unexpected situations.
		case current.InReplyToURI != parent.URI:
			l.Errorf("indirect in_reply_to_uri => %s", parent.URI)

		// The ID has changed for currently stored parent ID
		// (which may be empty, if new!) and fetched version.
		//
		// Update the current's inReplyTo fields to parent.
		case current.InReplyToID != parent.ID:
			l.Tracef("parent changed %s => %s", current.InReplyToID, parent.ID)
			current.InReplyToAccount = parent.Account
			if err := d.updateStatusParent(ctx,
				current,
				parent.ID,
				parent.URI,
				parent.AccountID,
			); err != nil {
				return err
			}
		}

		// Set next parent to use.
		current.InReplyTo = parent
		current = current.InReplyTo
	}

	return gtserror.Newf("reached %d ancestor iterations for %q", maxIter, status.URI)
}

// DereferenceStatusDescendents iterates downwards from the given status, using its replies, to ensure that as many children statuses as possible are dereferenced.
func (d *Dereferencer) DereferenceStatusDescendants(ctx context.Context, username string, statusIRI *url.URL, parent ap.Statusable) error {
	statusIRIStr := statusIRI.String()

	// Start log entry with fields
	l := log.WithContext(ctx).
		WithFields(kv.Fields{
			{"username", username},
			{"status", statusIRIStr},
		}...)

	// Log function start
	l.Trace("beginning")

	// OUR instance hostname.
	localhost := config.GetHost()

	// Keep track of already dereferenced collection
	// pages for this thread to prevent recursion.
	derefdPages := make(map[string]struct{}, 16)

	// frame represents a single stack frame when
	// iteratively derefencing status descendants.
	type frame struct {
		// page is the current activity streams
		// collection page we are on (as we often
		// push a frame to stack mid-paging).
		page ap.CollectionPageIterator

		// pageURI is the URI string of
		// the frame's collection page
		// (is useful for logging).
		pageURI string
	}

	var (
		// current stack frame
		current *frame

		// stack is a list of "shelved" descendand iterator
		// frames. this is pushed to when a child status frame
		// is found that we need to further iterate down, and
		// popped from into 'current' when that child's tree
		// of further descendants is exhausted.
		stack = []*frame{
			func() *frame {
				// Start input frame is built from the first input.
				page, pageURI := getAttachedStatusCollectionPage(parent)
				if page == nil {
					return nil
				}
				return &frame{page: page, pageURI: pageURI}
			}(),
		}

		// popStack will remove and return the top frame
		// from the stack, or nil if currently empty.
		popStack = func() *frame {
			if len(stack) == 0 {
				return nil
			}

			// Get frame index
			idx := len(stack) - 1

			// Pop last frame
			frame := stack[idx]
			stack = stack[:idx]

			return frame
		}
	)

stackLoop:
	for i := 0; i < maxIter; i++ {
		// Pop next frame, nil means we are at end
		if current = popStack(); current == nil {
			return nil
		}

	pageLoop:
		for {
			l.Tracef("following collection page: %s", current.pageURI)

		itemLoop:
			for {
				// Get next item from page iter.
				next := current.page.NextItem()
				if next == nil {
					break itemLoop
				}

				// Check for available IRI.
				itemIRI, _ := pub.ToId(next)
				if itemIRI == nil {
					continue itemLoop
				}

				if itemIRI.Host == localhost {
					// This child is one of ours,
					continue itemLoop
				}

				// Dereference the remote status and store in the database.
				// getStatusByURI guards against the following conditions:
				//   - refetching recently fetched statuses (recursion!)
				//   - remote domain is blocked (will return unretrievable)
				//   - any http type error for a new status returns unretrievable
				_, statusable, _, err := d.getStatusByURI(ctx, username, itemIRI)
				if err != nil {
					l.Errorf("error dereferencing remote status %s: %v", itemIRI, err)
					continue itemLoop
				}

				if statusable == nil {
					// A nil statusable return from
					// getStatusByURI() indicates a
					// remote status that was already
					// dereferenced recently (so no
					// need to go through descendents).
					continue itemLoop
				}

				// Extract any attached collection + ID URI from status.
				page, pageURI := getAttachedStatusCollectionPage(statusable)
				if page == nil {
					continue itemLoop
				}

				// Put current and next frame at top of stack
				stack = append(stack, current, &frame{
					pageURI: pageURI,
					page:    page,
				})

				// Now start at top of loop
				continue stackLoop
			}

			// Get the next page from iterator.
			next := current.page.NextPage()
			if next == nil || !next.IsIRI() {
				continue stackLoop
			}

			// Get the next page IRI.
			nextURI := next.GetIRI()
			nextURIStr := nextURI.String()

			// Check whether this page has already been deref'd.
			if _, ok := derefdPages[nextURIStr]; ok {
				l.Warnf("self referencing collection page(s): %s", nextURIStr)
				continue stackLoop
			}

			// Mark this collection page as deref'd.
			derefdPages[nextURIStr] = struct{}{}

			// Dereference this next collection page by its IRI.
			collectionPage, err := d.dereferenceCollectionPage(ctx,
				username,
				nextURI,
			)
			if err != nil {
				l.Errorf("error dereferencing collection page %q: %s", nextURIStr, err)
				continue stackLoop
			}

			// Set the next collection page.
			current.page = collectionPage
			current.pageURI = nextURIStr
			continue pageLoop
		}
	}

	return gtserror.Newf("reached %d descendant iterations for %q", maxIter, statusIRIStr)
}

// updateStatusParent updates the given status' parent
// status URI, ID and account ID to given values in DB.
func (d *Dereferencer) updateStatusParent(
	ctx context.Context,
	status *gtsmodel.Status,
	parentStatusID string,
	parentStatusURI string,
	parentAccountID string,
) error {
	status.InReplyToAccountID = parentAccountID
	status.InReplyToURI = parentStatusURI
	status.InReplyToID = parentStatusID
	if err := d.state.DB.UpdateStatus(ctx,
		status,
		"in_reply_to_id",
		"in_reply_to_uri",
		"in_reply_to_account_id",
	); err != nil {
		return gtserror.Newf("error updating status %s: %w", status.URI, err)
	}
	return nil
}

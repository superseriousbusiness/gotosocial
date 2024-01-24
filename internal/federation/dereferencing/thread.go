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
	"github.com/superseriousbusiness/activity/pub"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

// maxIter defines how many iterations of descendants or
// ancesters we are willing to follow before returning error.
const maxIter = 1000

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
		d.state.Workers.Federator.MustEnqueueCtx(ctx, func(ctx context.Context) {
			if err := d.DereferenceStatusDescendants(ctx, requestUser, uri, statusable); err != nil {
				log.Error(ctx, err)
			}
		})
	} else {
		// This is an existing status, dereference the WHOLE thread asynchronously.
		d.state.Workers.Federator.MustEnqueueCtx(ctx, func(ctx context.Context) {
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

		l = l.WithField("parent", current.InReplyToURI)
		l.Trace("following status ancestor")

		// Parse status parent URI for later use.
		uri, err := url.Parse(current.InReplyToURI)
		if err != nil {
			l.Warnf("invalid uri: %v", err)
			return nil
		}

		// Check whether this parent has already been deref'd.
		if _, ok := derefdStatuses[current.InReplyToURI]; ok {
			l.Warn("self referencing status ancestor")
			return nil
		}

		// Add this status's parent URI to map of deref'd.
		derefdStatuses[current.InReplyToURI] = struct{}{}

		// Fetch parent status by current's reply URI, this handles
		// case of existing (updating if necessary) or a new status.
		parent, update, _, err := d.getStatusByURI(ctx, username, uri)

		if err == nil && update == nil {
			// A parent status already existed
			// and was up-to-date, return here.
			return nil
		}

		// Check for a returned HTTP code via error.
		switch code := gtserror.StatusCode(err); {

		// Status codes 404 and 410 incicate the status does not exist anymore.
		// Gone (410) is the preferred for deletion, but we accept NotFound too.
		case code == http.StatusNotFound || code == http.StatusGone:
			l.Trace("status orphaned")
			current.InReplyToID = ""
			current.InReplyToURI = ""
			current.InReplyToAccountID = ""
			current.InReplyTo = nil
			current.InReplyToAccount = nil
			if err := d.state.DB.UpdateStatus(ctx,
				current,
				"in_reply_to_id",
				"in_reply_to_uri",
				"in_reply_to_account_id",
			); err != nil {
				return gtserror.Newf("db error updating status %s: %w", current.ID, err)
			}
			return nil

		// An error was returned for a status during
		// an attempted NEW dereference, return here.
		case err != nil && current.InReplyToID == "":
			return gtserror.Newf("error dereferencing new %s: %w", current.InReplyToURI, err)

		// An error was returned for an existing parent,
		// we simply treat this as a temporary situation.
		// (we fallback to using existing parent status).
		case err != nil:
			l.Errorf("error getting parent: %v", err)

		// The ID has changed for currently stored parent ID
		// (which may be empty, if new!) and fetched version.
		//
		// Update the current's inReplyTo fields to parent.
		case current.InReplyToID != parent.ID:
			l.Tracef("parent changed %s => %s", current.InReplyToID, parent.ID)
			current.InReplyToAccountID = parent.AccountID
			current.InReplyToAccount = parent.Account
			current.InReplyToURI = parent.URI
			current.InReplyToID = parent.ID
			current.InReplyTo = parent
			if err := d.state.DB.UpdateStatus(ctx,
				current,
				"in_reply_to_id",
				"in_reply_to_uri",
				"in_reply_to_account_id",
			); err != nil {
				return gtserror.Newf("db error updating status %s: %w", current.ID, err)
			}
		}

		// Set next parent to use.
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
	derefdPages := make(map[string]struct{}, 10)

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

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
	"errors"
	"net/http"
	"net/url"

	"codeberg.org/gruf/go-kv"
	"github.com/superseriousbusiness/activity/pub"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

// maxIter defines how many iterations of descendants or
// ancesters we are willing to follow before returning error.
const maxIter = 1000

func (d *deref) dereferenceThread(ctx context.Context, username string, statusIRI *url.URL, status *gtsmodel.Status, statusable ap.Statusable) {
	// Ensure that ancestors have been fully dereferenced
	if err := d.DereferenceStatusAncestors(ctx, username, status); err != nil {
		log.Error(ctx, err)
	}

	// Ensure that descendants have been fully dereferenced
	if err := d.DereferenceStatusDescendants(ctx, username, statusIRI, statusable); err != nil {
		log.Error(ctx, err)
	}
}

func (d *deref) DereferenceStatusAncestors(
	ctx context.Context,
	username string,
	status *gtsmodel.Status,
) error {
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

		// Add new log fields for this iteration.
		l = l.WithFields(kv.Fields{
			{"current", current.URI},
			{"parent", current.InReplyToURI},
		}...)
		l.Trace("following status ancestors")

		// Check whether this parent has already been deref'd.
		if _, ok := derefdStatuses[current.InReplyToURI]; ok {
			l.Warn("self referencing status ancestors")
			return nil
		}

		// Add this status URI to map of deref'd.
		derefdStatuses[current.URI] = struct{}{}

		if current.InReplyToID != "" {
			// We already have an InReplyToID set. This means
			// the status's parent has, at some point, been
			// inserted into the database, either because it
			// is a status from our instance, or a status from
			// remote that we've dereferenced before, or found
			// out about in some other way.
			//
			// Working on this assumption, check if the parent
			// status exists, either as a copy pinned on the
			// current status, or in the database.

			if current.InReplyTo != nil {
				// We have the parent already, and the child
				// doesn't need to be updated; keep iterating
				// from this parent upwards.
				current = current.InReplyTo
				continue
			}

			// Parent isn't pinned to this status (yet), see
			// if we can get it from the db (we should be
			// able to, since it has an ID already).
			parent, err := d.state.DB.GetStatusByID(
				gtscontext.SetBarebones(ctx),
				current.InReplyToID,
			)
			if err != nil && !errors.Is(err, db.ErrNoEntries) {
				// Real db error, stop.
				return gtserror.Newf("db error getting status %s: %w", current.InReplyToID, err)
			}

			if parent != nil {
				// We got the parent from the db, and the child
				// doesn't need to be updated; keep iterating
				// from this parent upwards.
				current.InReplyTo = parent
				current = parent
				continue
			}

			// If we arrive here, we know this child *did* have
			// a parent at some point, but it no longer exists in
			// the database, presumably because it's been deleted
			// by another action.
			//
			// TODO: clean this up in a nightly task.
			l.Warn("orphaned status (parent no longer exists)")
			return nil // Cannot iterate further.
		}

		// If we reach this point, we know the status has
		// an InReplyToURI set, but it doesn't yet have an
		// InReplyToID, which means that the parent status
		// has not yet been dereferenced.
		inReplyToURI, err := url.Parse(current.InReplyToURI)
		if err != nil || inReplyToURI == nil {
			// Parent URI is not something we can handle.
			l.Warn("orphaned status (invalid InReplyToURI)")
			return nil //nolint:nilerr
		}

		// Parent URI is valid, try to get it.
		// getStatusByURI guards against the following conditions:
		//   - refetching recently fetched statuses (recursion!)
		//   - remote domain is blocked (will return unretrievable)
		//   - any http type error for a new status returns unretrievable
		parent, _, err := d.getStatusByURI(ctx, username, inReplyToURI)
		if err == nil {
			// We successfully fetched the parent.
			// Update current status with new info.
			current.InReplyToID = parent.ID
			current.InReplyToAccountID = parent.AccountID
			if err := d.state.DB.UpdateStatus(
				ctx, current,
				"in_reply_to_id",
				"in_reply_to_account_id",
			); err != nil {
				return gtserror.Newf("db error updating status %s: %w", current.ID, err)
			}

			// Mark parent as next status to
			// work on, and keep iterating.
			current = parent
			continue
		}

		// We could not fetch the parent, check if we can do anything
		// useful with the error. For example, HTTP status code returned
		// from remote may indicate that the parent has been deleted.
		switch code := gtserror.StatusCode(err); {
		case code == http.StatusGone:
			// 410 means the status has definitely been deleted.
			// Update this status to reflect that, then bail.
			l.Debug("orphaned status: parent returned 410 Gone")

			current.InReplyToURI = ""
			if err := d.state.DB.UpdateStatus(
				ctx, current,
				"in_reply_to_uri",
			); err != nil {
				return gtserror.Newf("db error updating status %s: %w", current.ID, err)
			}

			return nil

		case code != 0:
			// We had a code, but not one indicating deletion, log the code
			// but don't return error or update the status; we can try again later.
			l.Warnf("orphaned status: http error dereferencing parent: %v)", err)
			return nil

		case gtserror.Unretrievable(err):
			// Not retrievable for some other reason, so just
			// bail for now; we can try again later if necessary.
			l.Warnf("orphaned status: parent unretrievable: %v)", err)
			return nil

		default:
			// Some other error that stops us in our tracks.
			return gtserror.Newf("error dereferencing parent %s: %w", current.InReplyToURI, err)
		}
	}

	return gtserror.Newf("reached %d ancestor iterations for %q", maxIter, status.URI)
}

func (d *deref) DereferenceStatusDescendants(ctx context.Context, username string, statusIRI *url.URL, parent ap.Statusable) error {
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

				// Check for available IRI on item
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
				_, statusable, err := d.getStatusByURI(ctx, username, itemIRI)
				if err != nil {
					if !gtserror.Unretrievable(err) {
						l.Errorf("error dereferencing remote status %s: %v", itemIRI, err)
					}
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

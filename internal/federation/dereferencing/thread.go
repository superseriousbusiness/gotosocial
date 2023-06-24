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
	"github.com/superseriousbusiness/activity/streams/vocab"
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
	// Mark given status as the one
	// we're currently working on.
	var current = status

	for i := 0; i < maxIter; i++ {
		if current.InReplyToURI == "" {
			// Status has no parent, we've
			// reached the top of the chain.
			return nil
		}

		l := log.
			WithContext(ctx).
			WithFields(kv.Fields{
				{"username", username},
				{"originalStatusIRI", status.URI},
				{"currentStatusURI", current.URI},
				{"currentInReplyToURI", current.InReplyToURI},
			}...)

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
			l.Warnf("current status has been orphaned (parent %s no longer exists in database)", current.InReplyToID)
			return nil // Cannot iterate further.
		}

		// If we reach this point, we know the status has
		// an InReplyToURI set, but it doesn't yet have an
		// InReplyToID, which means that the parent status
		// has not yet been dereferenced.
		inReplyToURI, err := url.Parse(current.InReplyToURI)
		if err != nil || inReplyToURI == nil {
			// Parent URI is not something we can handle.
			l.Debug("current status has been orphaned (invalid InReplyToURI)")
			return nil //nolint:nilerr
		}

		// Parent URI is valid, try to get it.
		// getStatusByURI guards against the following conditions:
		//
		//   - remote domain is blocked (will return unretrievable)
		//   - domain is local (will try to return something, or
		//     return unretrievable).
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
		case code == http.StatusGone || code == http.StatusNotFound:
			// 410 means the status has definitely been deleted.
			// 404 means the status has *probably* been deleted.
			// Update this status to reflect that, then bail.
			l.Debugf("current status has been orphaned (call to parent returned code %d)", code)

			current.InReplyToURI = ""
			if err := d.state.DB.UpdateStatus(
				ctx, current,
				"in_reply_to_uri",
			); err != nil {
				return gtserror.Newf("db error updating status %s: %w", current.ID, err)
			}
			return nil

		case code != 0:
			// We had a code, but not one indicating deletion,
			// log the code but don't return error or update the
			// status; we can try again later.
			l.Warnf("cannot dereference parent (%q)", err)
			return nil

		case gtserror.Unretrievable(err):
			// Not retrievable for some other reason, so just
			// bail; we can try again later if necessary.
			l.Debugf("parent unretrievable (%q)", err)
			return nil

		default:
			// Some other error that stops us in our tracks.
			return gtserror.Newf("error dereferencing parent %s: %w", current.InReplyToURI, err)
		}
	}

	return gtserror.Newf("reached %d ancestor iterations for %q", maxIter, status.URI)
}

func (d *deref) DereferenceStatusDescendants(ctx context.Context, username string, statusIRI *url.URL, parent ap.Statusable) error {
	// Take ref to original
	ogIRI := statusIRI

	// Start log entry with fields
	l := log.WithContext(ctx).
		WithFields(kv.Fields{
			{"username", username},
			{"statusIRI", ogIRI},
		}...)

	// Log function start
	l.Trace("beginning")

	// frame represents a single stack frame when iteratively
	// dereferencing status descendants. where statusIRI and
	// statusable are of the status whose children we are to
	// descend, page is the current activity streams collection
	// page of entities we are on (as we often push a frame to
	// stack mid-paging), and item___ are entity iterators for
	// this activity streams collection page.
	type frame struct {
		statusIRI  *url.URL
		statusable ap.Statusable
		page       ap.CollectionPageable
		itemIter   vocab.ActivityStreamsItemsPropertyIterator
	}

	var (
		// current is the current stack frame
		current *frame

		// stack is a list of "shelved" descendand iterator
		// frames. this is pushed to when a child status frame
		// is found that we need to further iterate down, and
		// popped from into 'current' when that child's tree
		// of further descendants is exhausted.
		stack = []*frame{
			{
				// Starting input is first frame
				statusIRI:  statusIRI,
				statusable: parent,
			},
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

		if current.page == nil {
			if current.statusIRI.Host == config.GetHost() {
				// This is a local status, no looping to do
				continue stackLoop
			}

			l.Tracef("following remote status descendants: %s", current.statusIRI)

			// Look for an attached status replies (as collection)
			replies := current.statusable.GetActivityStreamsReplies()
			if replies == nil {
				continue stackLoop
			}

			// Get the status replies collection
			collection := replies.GetActivityStreamsCollection()
			if collection == nil {
				continue stackLoop
			}

			// Get the "first" property of the replies collection
			first := collection.GetActivityStreamsFirst()
			if first == nil {
				continue stackLoop
			}

			// Set the first activity stream collection page
			current.page = first.GetActivityStreamsCollectionPage()
			if current.page == nil {
				continue stackLoop
			}
		}

	pageLoop:
		for {
			if current.itemIter == nil {
				// Get the items associated with this page
				items := current.page.GetActivityStreamsItems()
				if items == nil {
					continue stackLoop
				}

				// Start off the item iterator
				current.itemIter = items.Begin()
			}

		itemLoop:
			for {
				// Check for remaining iter
				if current.itemIter == nil {
					break itemLoop
				}

				// Get current item iterator
				itemIter := current.itemIter

				// Set the next available iterator
				current.itemIter = itemIter.Next()

				// Check for available IRI on item
				itemIRI, _ := pub.ToId(itemIter)
				if itemIRI == nil {
					continue itemLoop
				}

				if itemIRI.Host == config.GetHost() {
					// This child is one of ours,
					continue itemLoop
				}

				// Dereference the remote status and store in the database.
				// getStatusByURI guards against the following conditions:
				//
				//   - remote domain is blocked (will return unretrievable)
				//   - domain is local (will try to return something, or
				//     return unretrievable).
				_, statusable, err := d.getStatusByURI(ctx, username, itemIRI)
				if err != nil {
					if !gtserror.Unretrievable(err) {
						l.Errorf("error dereferencing remote status %s: %v", itemIRI, err)
					}

					continue itemLoop
				}

				if statusable == nil {
					// Already up-to-date.
					continue itemLoop
				}

				// Put current and next frame at top of stack
				stack = append(stack, current, &frame{
					statusIRI:  itemIRI,
					statusable: statusable,
				})

				// Now start at top of loop
				continue stackLoop
			}

			// Get the current page's "next" property
			pageNext := current.page.GetActivityStreamsNext()
			if pageNext == nil || !pageNext.IsIRI() {
				continue stackLoop
			}

			// Get the IRI of the "next" property.
			pageNextIRI := pageNext.GetIRI()

			// Ensure this isn't a self-referencing page...
			// We don't need to store / check against a map of IRIs
			// as our getStatusByIRI() function above prevents iter'ing
			// over statuses that have been dereferenced recently, due to
			// the `fetched_at` field preventing frequent refetches.
			if id := current.page.GetJSONLDId(); id != nil &&
				pageNextIRI.String() == id.Get().String() {
				log.Warnf(ctx, "self referencing collection page: %s", pageNextIRI)
				continue stackLoop
			}

			// Dereference this next collection page by its IRI
			collectionPage, err := d.dereferenceCollectionPage(ctx,
				username,
				pageNextIRI,
			)
			if err != nil {
				l.Errorf("error dereferencing remote collection page %q: %s", pageNextIRI.String(), err)
				continue stackLoop
			}

			// Set the updated collection page
			current.page = collectionPage
			continue pageLoop
		}
	}

	return gtserror.Newf("reached %d descendant iterations for %q", maxIter, ogIRI.String())
}

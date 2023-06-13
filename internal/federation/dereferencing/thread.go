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
	"net/url"

	"codeberg.org/gruf/go-kv"
	"github.com/superseriousbusiness/activity/pub"
	"github.com/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/uris"
)

// maxIter defines how many iterations of descendants or
// ancesters we are willing to follow before returning error.
const maxIter = 1000

// dereferenceThread will dereference statuses both above and below the given status in a thread, it returns no error and is intended to be called asychronously.
func (d *deref) dereferenceThread(ctx context.Context, username string, statusIRI *url.URL, status *gtsmodel.Status, statusable ap.Statusable) {
	// Ensure that ancestors have been fully dereferenced
	if err := d.dereferenceStatusAncestors(ctx, username, status); err != nil {
		log.Error(ctx, err) // log entry and error will include caller prefixes
	}

	// Ensure that descendants have been fully dereferenced
	if err := d.dereferenceStatusDescendants(ctx, username, statusIRI, statusable); err != nil {
		log.Error(ctx, err) // log entry and error will include caller prefixes
	}
}

// dereferenceAncestors has the goal of reaching the oldest ancestor of a given status, and stashing all statuses along the way.
func (d *deref) dereferenceStatusAncestors(ctx context.Context, username string, status *gtsmodel.Status) error {
	// Take ref to original
	ogIRI := status.URI

	// Start log entry with fields
	l := log.WithContext(ctx).
		WithFields(kv.Fields{
			{"username", username},
			{"statusIRI", ogIRI},
		}...)

	// Log function start
	l.Trace("beginning")

	for i := 0; i < maxIter; i++ {
		if status.InReplyToURI == "" {
			// status doesn't reply to anything
			return nil
		}

		// Parse this status's replied IRI
		replyIRI, err := url.Parse(status.InReplyToURI)
		if err != nil {
			return gtserror.Newf("invalid status InReplyToURI %q: %w", status.InReplyToURI, err)
		}

		if replyIRI.Host == config.GetHost() {
			l.Tracef("following local status ancestors: %s", status.InReplyToURI)

			// This is our status, extract ID from path
			_, id, err := uris.ParseStatusesPath(replyIRI)
			if err != nil {
				return gtserror.Newf("invalid local status IRI %q: %w", status.InReplyToURI, err)
			}

			// Fetch this status from the database
			localStatus, err := d.state.DB.GetStatusByID(ctx, id)
			if err != nil {
				return gtserror.Newf("error fetching local status %q: %w", id, err)
			}

			// Set the fetched status
			status = localStatus

		} else {
			l.Tracef("following remote status ancestors: %s", status.InReplyToURI)

			// Fetch the remote status found at this IRI
			remoteStatus, _, err := d.getStatusByURI(
				ctx,
				username,
				replyIRI,
			)
			if err != nil {
				return gtserror.Newf("error fetching remote status %q: %w", status.InReplyToURI, err)
			}

			// Set the fetched status
			status = remoteStatus
		}
	}

	return gtserror.Newf("reached %d ancestor iterations for %q", maxIter, ogIRI)
}

func (d *deref) dereferenceStatusDescendants(ctx context.Context, username string, statusIRI *url.URL, parent ap.Statusable) error {
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
				_, statusable, err := d.getStatusByURI(ctx, username, itemIRI)
				if err != nil {
					l.Errorf("error dereferencing remote status %s: %v", itemIRI, err)
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
			// over statuses that have been dereferenced recently.
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

/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package dereferencing

import (
	"context"
	"fmt"
	"net/url"

	"codeberg.org/gruf/go-kv"
	"github.com/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/uris"
)

// maxIter defines how many iterations of descendants or
// ancesters we are willing to follow before returning error.
const maxIter = 1000

// DereferenceThread takes a statusable (something that has withReplies and withInReplyTo),
// and dereferences statusables in the conversation.
//
// This process involves working up and down the chain of replies, and parsing through the collections of IDs
// presented by remote instances as part of their replies collections, and will likely involve making several calls to
// multiple different hosts.
func (d *deref) DereferenceThread(ctx context.Context, username string, statusIRI *url.URL) error {
	l := log.WithFields(kv.Fields{
		{"username", username},
		{"statusIRI", statusIRI},
	}...)

	// Log function start
	l.Trace("beginning")

	// First make sure we have this status in our database
	status, statusable, err := d.GetRemoteStatus(ctx, username, statusIRI, true, false)
	if err != nil {
		return fmt.Errorf("DereferenceThread: error getting initial status with id %s: %s", statusIRI.String(), err)
	}

	// Ensure that ancestors have been fully dereferenced
	if err := d.dereferenceStatusAncestors(ctx, username, status); err != nil {
		l.Errorf("error dereferencing status ancestors: %v", err)
		// we don't return error, we have deref'd as much as we can
	}

	// Ensure that descendants have been fully dereferenced
	if err := d.dereferenceStatusDescendants(ctx, username, statusIRI, statusable); err != nil {
		l.Errorf("error dereferencing status descendants: %v", err)
		// we don't return error, we have deref'd as much as we can
	}

	return nil
}

// dereferenceAncestors has the goal of reaching the oldest ancestor of a given status, and stashing all statuses along the way.
func (d *deref) dereferenceStatusAncestors(ctx context.Context, username string, status *gtsmodel.Status) error {
	// Take ref to original
	ogIRI := status.URI

	// Start log entry with fields
	l := log.WithFields(kv.Fields{
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
			return fmt.Errorf("invalid status InReplyToURI %q: %w", status.InReplyToURI, err)
		}

		if replyIRI.Host == config.GetHost() {
			l.Tracef("following local status ancestors: %s", status.InReplyToURI)

			// This is our status, extract ID from path
			_, id, err := uris.ParseStatusesPath(replyIRI)
			if err != nil {
				return fmt.Errorf("invalid local status IRI %q: %w", status.InReplyToURI, err)
			}

			// Fetch this status from the database
			localStatus, err := d.db.GetStatusByID(ctx, id)
			if err != nil {
				return fmt.Errorf("error fetching local status %q: %w", id, err)
			}

			// Set the fetched status
			status = localStatus

		} else {
			l.Tracef("following remote status ancestors: %s", status.InReplyToURI)

			// Fetch the remote status found at this IRI
			remoteStatus, _, err := d.GetRemoteStatus(ctx, username, replyIRI, false, false)
			if err != nil {
				return fmt.Errorf("error fetching remote status %q: %w", status.InReplyToURI, err)
			}

			// Set the fetched status
			status = remoteStatus
		}
	}

	return fmt.Errorf("reached %d ancestor iterations for %q", maxIter, ogIRI)
}

func (d *deref) dereferenceStatusDescendants(ctx context.Context, username string, statusIRI *url.URL, parent ap.Statusable) error {
	// Take ref to original
	ogIRI := statusIRI

	// Start log entry with fields
	l := log.WithFields(kv.Fields{
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
		iterLen    int
		iterIdx    int
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
			// This is a local status, no looping to do
			if current.statusIRI.Host == config.GetHost() {
				continue stackLoop
			}

			l.Tracef("following remote status descendants: %s", current.statusIRI)

			// Look for an attached status replies (as collection)
			replies := current.statusable.GetActivityStreamsReplies()
			if replies == nil || !replies.IsActivityStreamsCollection() {
				continue stackLoop
			}

			// Get the status replies collection
			collection := replies.GetActivityStreamsCollection()

			// Get the "first" property of the replies collection
			first := collection.GetActivityStreamsFirst()
			if first == nil || !first.IsActivityStreamsCollectionPage() {
				continue stackLoop
			}

			// Set the first activity stream collection page
			current.page = first.GetActivityStreamsCollectionPage()
		}

		for /* page loop */ {
			if current.itemIter == nil {
				// Check this page contains any items...
				items := current.page.GetActivityStreamsItems()
				if current.iterLen = items.Len(); current.iterLen == 0 {
					continue stackLoop
				}

				// Start off the item iterator
				current.itemIter = items.Begin()
				current.iterIdx = -1
			}

		itemLoop:
			for current.iterIdx++; current.iterIdx < current.iterLen; current.iterIdx++ {
				var itemIRI *url.URL

				// Get next item iterator object
				current.itemIter = current.itemIter.Next()

				switch {
				// Item is already an IRI
				case current.itemIter.IsIRI():
					itemIRI = current.itemIter.GetIRI()

				// Item is a note, get the note ID IRI
				case current.itemIter.IsActivityStreamsNote():
					note := current.itemIter.GetActivityStreamsNote()
					if id := note.GetJSONLDId(); id != nil && id.IsIRI() {
						itemIRI = id.GetIRI()
					}
				}

				if itemIRI == nil {
					// Unusable iter object
					continue itemLoop
				}

				if itemIRI.Host == config.GetHost() {
					// This child is one of ours,
					continue itemLoop
				}

				// Dereference the remote status and store in the database
				_, statusable, err := d.GetRemoteStatus(ctx, username, itemIRI, true, false)
				if err != nil {
					l.Errorf("error dereferencing remote status \"%s\": %s", itemIRI, err)
					continue itemLoop
				}

				// Put current and next frame at top of stack
				stack = append(stack, current, &frame{
					statusIRI:  itemIRI,
					statusable: statusable,
				})
			}

			// Item iterator is done
			current.itemIter = nil

			// Get the collection page "next" property
			pageNext := current.page.GetActivityStreamsNext()
			if pageNext == nil || !pageNext.IsIRI() {
				continue stackLoop
			}

			// Get the "next" page property IRI
			pageNextIRI := pageNext.GetIRI()

			// Dereference this next collection page by its IRI
			collectionPage, err := d.DereferenceCollectionPage(ctx, username, pageNextIRI)
			if err != nil {
				l.Errorf("error dereferencing remote collection page \"%s\": %s", pageNextIRI, err)
				continue stackLoop
			}

			// Set the updated collection page
			current.page = collectionPage
		}
	}

	return fmt.Errorf("reached %d descendant iterations for %q", maxIter, ogIRI.String())
}

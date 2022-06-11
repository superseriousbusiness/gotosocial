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

	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/uris"
)

// DereferenceThread takes a statusable (something that has withReplies and withInReplyTo),
// and dereferences statusables in the conversation.
//
// This process involves working up and down the chain of replies, and parsing through the collections of IDs
// presented by remote instances as part of their replies collections, and will likely involve making several calls to
// multiple different hosts.
func (d *deref) DereferenceThread(ctx context.Context, username string, statusIRI *url.URL) error {
	l := logrus.WithFields(logrus.Fields{
		"func":      "DereferenceThread",
		"username":  username,
		"statusIRI": statusIRI.String(),
	})
	l.Trace("entering DereferenceThread")

	// if it's our status we already have everything stashed so we can bail early
	if statusIRI.Host == config.GetHost() {
		l.Trace("iri belongs to us, bailing")
		return nil
	}

	// first make sure we have this status in our db
	_, statusable, err := d.GetRemoteStatus(ctx, username, statusIRI, true, false)
	if err != nil {
		return fmt.Errorf("DereferenceThread: error getting initial status with id %s: %s", statusIRI.String(), err)
	}

	// first iterate up through ancestors, dereferencing if necessary as we go
	if err := d.iterateAncestors(ctx, username, *statusIRI); err != nil {
		return fmt.Errorf("error iterating ancestors of status %s: %s", statusIRI.String(), err)
	}

	// now iterate down through descendants, again dereferencing as we go
	if err := d.iterateDescendants(ctx, username, *statusIRI, statusable); err != nil {
		return fmt.Errorf("error iterating descendants of status %s: %s", statusIRI.String(), err)
	}

	return nil
}

// iterateAncestors has the goal of reaching the oldest ancestor of a given status, and stashing all statuses along the way.
func (d *deref) iterateAncestors(ctx context.Context, username string, statusIRI url.URL) error {
	l := logrus.WithFields(logrus.Fields{
		"func":      "iterateAncestors",
		"username":  username,
		"statusIRI": statusIRI.String(),
	})
	l.Trace("entering iterateAncestors")

	// if it's our status we don't need to dereference anything so we can immediately move up the chain
	if statusIRI.Host == config.GetHost() {
		l.Trace("iri belongs to us, moving up to next ancestor")

		// since this is our status, we know we can extract the id from the status path
		_, id, err := uris.ParseStatusesPath(&statusIRI)
		if err != nil {
			return err
		}

		status, err := d.db.GetStatusByID(ctx, id)
		if err != nil {
			return err
		}

		if status.InReplyToURI == "" {
			// status doesn't reply to anything
			return nil
		}

		nextIRI, err := url.Parse(status.URI)
		if err != nil {
			return err
		}

		return d.iterateAncestors(ctx, username, *nextIRI)
	}

	// If we reach here, we're looking at a remote status
	_, statusable, err := d.GetRemoteStatus(ctx, username, &statusIRI, true, false)
	if err != nil {
		l.Debugf("couldn't get remote status %s: %s; can't iterate any more ancestors", statusIRI.String(), err)
		return nil
	}

	inReplyTo := ap.ExtractInReplyToURI(statusable)
	if inReplyTo == nil || inReplyTo.String() == "" {
		// status doesn't reply to anything
		return nil
	}

	// now move up to the next ancestor
	return d.iterateAncestors(ctx, username, *inReplyTo)
}

func (d *deref) iterateDescendants(ctx context.Context, username string, statusIRI url.URL, statusable ap.Statusable) error {
	l := logrus.WithFields(logrus.Fields{
		"func":      "iterateDescendants",
		"username":  username,
		"statusIRI": statusIRI.String(),
	})
	l.Trace("entering iterateDescendants")

	// if it's our status we already have descendants stashed so we can bail early
	if statusIRI.Host == config.GetHost() {
		l.Trace("iri belongs to us, bailing")
		return nil
	}

	replies := statusable.GetActivityStreamsReplies()
	if replies == nil || !replies.IsActivityStreamsCollection() {
		l.Trace("no replies, bailing")
		return nil
	}

	repliesCollection := replies.GetActivityStreamsCollection()
	if repliesCollection == nil {
		l.Trace("replies collection is nil, bailing")
		return nil
	}

	first := repliesCollection.GetActivityStreamsFirst()
	if first == nil {
		l.Trace("replies collection has no first, bailing")
		return nil
	}

	firstPage := first.GetActivityStreamsCollectionPage()
	if firstPage == nil {
		l.Trace("first has no collection page, bailing")
		return nil
	}

	firstPageNext := firstPage.GetActivityStreamsNext()
	if firstPageNext == nil || !firstPageNext.IsIRI() {
		l.Trace("next is not an iri, bailing")
		return nil
	}

	var foundReplies int
	currentPageIRI := firstPageNext.GetIRI()

pageLoop:
	for {
		l.Tracef("dereferencing page %s", currentPageIRI)
		collectionPage, err := d.DereferenceCollectionPage(ctx, username, currentPageIRI)
		if err != nil {
			l.Debugf("couldn't get remote collection page %s: %s; breaking pageLoop", currentPageIRI, err)
			break pageLoop
		}

		pageItems := collectionPage.GetActivityStreamsItems()
		if pageItems.Len() == 0 {
			// no items on this page, which means we're done
			break pageLoop
		}

		// have a look through items and see what we can find
		for iter := pageItems.Begin(); iter != pageItems.End(); iter = iter.Next() {
			// We're looking for a url to feed to GetRemoteStatus.
			// Each item can be either an IRI, or a Note.
			// If a note, we grab the ID from it and call it, rather than parsing the note.
			var itemURI *url.URL
			switch {
			case iter.IsIRI():
				// iri, easy
				itemURI = iter.GetIRI()
			case iter.IsActivityStreamsNote():
				// note, get the id from it to use as iri
				note := iter.GetActivityStreamsNote()
				noteID := note.GetJSONLDId()
				if noteID != nil && noteID.IsIRI() {
					itemURI = noteID.GetIRI()
				}
			default:
				// if it's not an iri or a note, we don't know how to process it
				continue
			}

			if itemURI.Host == config.GetHost() {
				// skip if the reply is from us -- we already have it then
				continue
			}

			// we can confidently say now that we found something
			foundReplies++

			// get the remote statusable and put it in the db
			_, statusable, err := d.GetRemoteStatus(ctx, username, itemURI, true, false)
			if err == nil {
				// now iterate descendants of *that* status
				if err := d.iterateDescendants(ctx, username, *itemURI, statusable); err != nil {
					continue
				}
			}
		}

		nextPage := collectionPage.GetActivityStreamsNext()
		if nextPage != nil && nextPage.IsIRI() {
			nextPageIRI := nextPage.GetIRI()
			l.Tracef("moving on to next page %s", nextPageIRI)
			currentPageIRI = nextPageIRI
		} else {
			l.Trace("no next page, bailing")
			break pageLoop
		}
	}

	l.Debugf("foundReplies %d", foundReplies)
	return nil
}

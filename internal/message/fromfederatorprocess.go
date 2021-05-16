/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

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

package message

import (
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/transport"
)

func (p *processor) processFromFederator(federatorMsg gtsmodel.FromFederator) error {
	l := p.log.WithFields(logrus.Fields{
		"func":         "processFromFederator",
		"federatorMsg": fmt.Sprintf("%+v", federatorMsg),
	})

	l.Debug("entering function PROCESS FROM FEDERATOR")

	switch federatorMsg.APObjectType {
	case gtsmodel.ActivityStreamsNote:

		incomingStatus, ok := federatorMsg.GTSModel.(*gtsmodel.Status)
		if !ok {
			return errors.New("note was not parseable as *gtsmodel.Status")
		}

		l.Debug("will now derefence incoming status")
		if err := p.dereferenceStatusFields(incomingStatus); err != nil {
			return fmt.Errorf("error dereferencing status from federator: %s", err)
		}
		if err := p.db.UpdateByID(incomingStatus.ID, incomingStatus); err != nil {
			return fmt.Errorf("error updating dereferenced status in the db: %s", err)
		}

		if err := p.notifyStatus(incomingStatus); err != nil {
			return err
		}
	}

	return nil
}

// dereferenceStatusFields fetches all the information we temporarily pinned to an incoming
// federated status, back in the federating db's Create function.
//
// When a status comes in from the federation API, there are certain fields that
// haven't been dereferenced yet, because we needed to provide a snappy synchronous
// response to the caller. By the time it reaches this function though, it's being
// processed asynchronously, so we have all the time in the world to fetch the various
// bits and bobs that are attached to the status, and properly flesh it out, before we
// send the status to any timelines and notify people.
//
// Things to dereference and fetch here:
//
// 1. Media attachments.
// 2. Hashtags.
// 3. Emojis.
// 4. Mentions.
// 5. Posting account.
// 6. Replied-to-status.
//
// SIDE EFFECTS:
// This function will deference all of the above, insert them in the database as necessary,
// and attach them to the status. The status itself will not be added to the database yet,
// that's up the caller to do.
func (p *processor) dereferenceStatusFields(status *gtsmodel.Status) error {
	l := p.log.WithFields(logrus.Fields{
		"func": "dereferenceStatusFields",
		"status": fmt.Sprintf("%+v", status),
	})
	l.Debug("entering function")

	var t transport.Transport
	var err error
	var username string
	// TODO: dereference with a user that's addressed by the status
	t, err = p.federator.GetTransportForUser(username)
	if err != nil {
		return fmt.Errorf("error creating transport: %s", err)
	}

	// 1. Media attachments.
	//
	// At this point we should know:
	// * the media type of the file we're looking for (a.File.ContentType)
	// * the blurhash (a.Blurhash)
	// * the file type (a.Type)
	// * the remote URL (a.RemoteURL)
	// This should be enough to pass along to the media processor.
	attachmentIDs := []string{}
	for _, a := range status.GTSMediaAttachments {
		l.Debug("dereferencing attachment: %+v", a)

		// it might have been processed elsewhere so check first if it's already in the database or not
		maybeAttachment := &gtsmodel.MediaAttachment{}
		err := p.db.GetWhere("remote_url", a.RemoteURL, maybeAttachment)
		if err == nil {
			// we already have it in the db, dereferenced, no need to do it again
			l.Debugf("attachment already exists with id %s", maybeAttachment.ID)
			attachmentIDs = append(attachmentIDs, maybeAttachment.ID)
			continue
		}
		if _, ok := err.(db.ErrNoEntries); !ok {
			// we have a real error
			return fmt.Errorf("error checking db for existence of attachment with remote url %s: %s", a.RemoteURL, err)
		}
		// it just doesn't exist yet so carry on
		l.Debug("attachment doesn't exist yet, calling ProcessRemoteAttachment", a)
		deferencedAttachment, err := p.mediaHandler.ProcessRemoteAttachment(t, a, status.AccountID)
		if err != nil {
			p.log.Errorf("error dereferencing status attachment: %s", err)
			continue
		}
		l.Debugf("dereferenced attachment: %+v", deferencedAttachment)
		deferencedAttachment.StatusID = deferencedAttachment.ID
		if err := p.db.Put(deferencedAttachment); err != nil {
			return fmt.Errorf("error inserting dereferenced attachment with remote url %s: %s", a.RemoteURL, err)
		}
		deferencedAttachment.Description = a.Description
		attachmentIDs = append(attachmentIDs, deferencedAttachment.ID)
	}
	status.Attachments = attachmentIDs

	return nil
}

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
	"net/url"

	"github.com/google/uuid"
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

	switch federatorMsg.APActivityType {
	case gtsmodel.ActivityStreamsCreate:
		// CREATE
		switch federatorMsg.APObjectType {
		case gtsmodel.ActivityStreamsNote:
			// CREATE A STATUS
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
		case gtsmodel.ActivityStreamsProfile:
			// CREATE AN ACCOUNT
			incomingAccount, ok := federatorMsg.GTSModel.(*gtsmodel.Account)
			if !ok {
				return errors.New("profile was not parseable as *gtsmodel.Account")
			}

			l.Debug("will now derefence incoming account")
			if err := p.dereferenceAccountFields(incomingAccount, "", false); err != nil {
				return fmt.Errorf("error dereferencing account from federator: %s", err)
			}
			if err := p.db.UpdateByID(incomingAccount.ID, incomingAccount); err != nil {
				return fmt.Errorf("error updating dereferenced account in the db: %s", err)
			}
		}
	case gtsmodel.ActivityStreamsUpdate:
		// UPDATE
		switch federatorMsg.APObjectType {
		case gtsmodel.ActivityStreamsProfile:
			// UPDATE AN ACCOUNT
			incomingAccount, ok := federatorMsg.GTSModel.(*gtsmodel.Account)
			if !ok {
				return errors.New("profile was not parseable as *gtsmodel.Account")
			}

			l.Debug("will now derefence incoming account")
			if err := p.dereferenceAccountFields(incomingAccount, federatorMsg.ReceivingAccount.Username, true); err != nil {
				return fmt.Errorf("error dereferencing account from federator: %s", err)
			}
			if err := p.db.UpdateByID(incomingAccount.ID, incomingAccount); err != nil {
				return fmt.Errorf("error updating dereferenced account in the db: %s", err)
			}
		}
	case gtsmodel.ActivityStreamsDelete:
		// DELETE
		switch federatorMsg.APObjectType {
		case gtsmodel.ActivityStreamsNote:
			// DELETE A STATUS
			// TODO: handle side effects of status deletion here:
			// 1. delete all media associated with status
			// 2. delete boosts of status
			// 3. etc etc etc
		case gtsmodel.ActivityStreamsProfile:
			// DELETE A PROFILE/ACCOUNT
			// TODO: handle side effects of account deletion here: delete all objects, statuses, media etc associated with account
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
		"func":   "dereferenceStatusFields",
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

	// the status should have an ID by now, but just in case it doesn't let's generate one here
	// because we'll need it further down
	if status.ID == "" {
		status.ID = uuid.NewString()
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
		l.Debugf("dereferencing attachment: %+v", a)

		// it might have been processed elsewhere so check first if it's already in the database or not
		maybeAttachment := &gtsmodel.MediaAttachment{}
		err := p.db.GetWhere([]db.Where{{Key: "remote_url", Value: a.RemoteURL}}, maybeAttachment)
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
		deferencedAttachment.StatusID = status.ID
		if err := p.db.Put(deferencedAttachment); err != nil {
			return fmt.Errorf("error inserting dereferenced attachment with remote url %s: %s", a.RemoteURL, err)
		}
		deferencedAttachment.Description = a.Description
		attachmentIDs = append(attachmentIDs, deferencedAttachment.ID)
	}
	status.Attachments = attachmentIDs

	// 2. Hashtags

	// 3. Emojis

	// 4. Mentions
	// At this point, mentions should have the namestring and mentionedAccountURI set on them.
	//
	// We should dereference any accounts mentioned here which we don't have in our db yet, by their URI.
	mentions := []string{}
	for _, m := range status.GTSMentions {
		uri, err := url.Parse(m.MentionedAccountURI)
		if err != nil {
			l.Debugf("error parsing mentioned account uri %s: %s", m.MentionedAccountURI, err)
			continue
		}

		m.StatusID = status.ID
		m.OriginAccountID = status.GTSAccount.ID
		m.OriginAccountURI = status.GTSAccount.URI

		targetAccount := &gtsmodel.Account{}
		if err := p.db.GetWhere([]db.Where{{Key: "uri", Value: uri.String()}}, targetAccount); err != nil {
			// proper error
			if _, ok := err.(db.ErrNoEntries); !ok {
				return fmt.Errorf("db error checking for account with uri %s", uri.String())
			}

			// we just don't have it yet, so we should go get it....
			accountable, err := p.federator.DereferenceRemoteAccount(username, uri)
			if err != nil {
				// we can't dereference it so just skip it
				l.Debugf("error dereferencing remote account with uri %s: %s", uri.String(), err)
				continue
			}

			targetAccount, err = p.tc.ASRepresentationToAccount(accountable, false)
			if err != nil {
				l.Debugf("error converting remote account with uri %s into gts model: %s", uri.String(), err)
				continue
			}

			if err := p.db.Put(targetAccount); err != nil {
				return fmt.Errorf("db error inserting account with uri %s", uri.String())
			}
		}

		// by this point, we know the targetAccount exists in our database with an ID :)
		m.TargetAccountID = targetAccount.ID
		if err := p.db.Put(m); err != nil {
			return fmt.Errorf("error creating mention: %s", err)
		}
		mentions = append(mentions, m.ID)
	}
	status.Mentions = mentions

	return nil
}

func (p *processor) dereferenceAccountFields(account *gtsmodel.Account, requestingUsername string, refresh bool) error {
	l := p.log.WithFields(logrus.Fields{
		"func":               "dereferenceAccountFields",
		"requestingUsername": requestingUsername,
	})

	t, err := p.federator.GetTransportForUser(requestingUsername)
	if err != nil {
		return fmt.Errorf("error getting transport for user: %s", err)
	}

	// fetch the header and avatar
	if err := p.fetchHeaderAndAviForAccount(account, t, refresh); err != nil {
		// if this doesn't work, just skip it -- we can do it later
		l.Debugf("error fetching header/avi for account: %s", err)
	}

	if err := p.db.UpdateByID(account.ID, account); err != nil {
		return fmt.Errorf("error updating account in database: %s", err)
	}

	return nil
}

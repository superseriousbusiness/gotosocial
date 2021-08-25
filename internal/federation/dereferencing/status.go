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

package dereferencing

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

	"github.com/go-fed/activity/streams"
	"github.com/go-fed/activity/streams/vocab"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
)

// EnrichRemoteStatus takes a status that's already been inserted into the database in a minimal form,
// and populates it with additional fields, media, etc.
//
// EnrichRemoteStatus is mostly useful for calling after a status has been initially created by
// the federatingDB's Create function, but additional dereferencing is needed on it.
func (d *deref) EnrichRemoteStatus(ctx context.Context, username string, status *gtsmodel.Status) (*gtsmodel.Status, error) {
	if err := d.populateStatusFields(ctx, status, username); err != nil {
		return nil, err
	}

	if err := d.db.UpdateByID(ctx, status.ID, status); err != nil {
		return nil, fmt.Errorf("EnrichRemoteStatus: error updating status: %s", err)
	}

	return status, nil
}

// GetRemoteStatus completely dereferences a remote status, converts it to a GtS model status,
// puts it in the database, and returns it to a caller. The boolean indicates whether the status is new
// to us or not. If we haven't seen the status before, bool will be true. If we have seen the status before,
// it will be false.
//
// If refresh is true, then even if we have the status in our database already, it will be dereferenced from its
// remote representation, as will its owner.
//
// If a dereference was performed, then the function also returns the ap.Statusable representation for further processing.
//
// SIDE EFFECTS: remote status will be stored in the database, and the remote status owner will also be stored.
func (d *deref) GetRemoteStatus(ctx context.Context, username string, remoteStatusID *url.URL, refresh bool) (*gtsmodel.Status, ap.Statusable, bool, error) {
	new := true

	// check if we already have the status in our db
	maybeStatus, err := d.db.GetStatusByURI(ctx, remoteStatusID.String())
	if err == nil {
		// we've seen this status before so it's not new
		new = false

		// if we're not being asked to refresh, we can just return the maybeStatus as-is and avoid doing any external calls
		if !refresh {
			return maybeStatus, nil, new, nil
		}
	}

	statusable, err := d.dereferenceStatusable(ctx, username, remoteStatusID)
	if err != nil {
		return nil, statusable, new, fmt.Errorf("GetRemoteStatus: error dereferencing statusable: %s", err)
	}

	accountURI, err := ap.ExtractAttributedTo(statusable)
	if err != nil {
		return nil, statusable, new, fmt.Errorf("GetRemoteStatus: error extracting attributedTo: %s", err)
	}

	// do this so we know we have the remote account of the status in the db
	_, _, err = d.GetRemoteAccount(ctx, username, accountURI, false)
	if err != nil {
		return nil, statusable, new, fmt.Errorf("GetRemoteStatus: couldn't derive status author: %s", err)
	}

	gtsStatus, err := d.typeConverter.ASStatusToStatus(ctx, statusable)
	if err != nil {
		return nil, statusable, new, fmt.Errorf("GetRemoteStatus: error converting statusable to status: %s", err)
	}

	if new {
		ulid, err := id.NewULIDFromTime(gtsStatus.CreatedAt)
		if err != nil {
			return nil, statusable, new, fmt.Errorf("GetRemoteStatus: error generating new id for status: %s", err)
		}
		gtsStatus.ID = ulid

		if err := d.populateStatusFields(ctx, gtsStatus, username); err != nil {
			return nil, statusable, new, fmt.Errorf("GetRemoteStatus: error populating status fields: %s", err)
		}

		if err := d.db.PutStatus(ctx, gtsStatus); err != nil {
			return nil, statusable, new, fmt.Errorf("GetRemoteStatus: error putting new status: %s", err)
		}
	} else {
		gtsStatus.ID = maybeStatus.ID

		if err := d.populateStatusFields(ctx, gtsStatus, username); err != nil {
			return nil, statusable, new, fmt.Errorf("GetRemoteStatus: error populating status fields: %s", err)
		}

		if err := d.db.UpdateByID(ctx, gtsStatus.ID, gtsStatus); err != nil {
			return nil, statusable, new, fmt.Errorf("GetRemoteStatus: error updating status: %s", err)
		}
	}

	return gtsStatus, statusable, new, nil
}

func (d *deref) dereferenceStatusable(ctx context.Context, username string, remoteStatusID *url.URL) (ap.Statusable, error) {
	if blocked, err := d.db.IsDomainBlocked(ctx, remoteStatusID.Host); blocked || err != nil {
		return nil, fmt.Errorf("DereferenceStatusable: domain %s is blocked", remoteStatusID.Host)
	}

	transport, err := d.transportController.NewTransportForUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("DereferenceStatusable: transport err: %s", err)
	}

	b, err := transport.Dereference(context.Background(), remoteStatusID)
	if err != nil {
		return nil, fmt.Errorf("DereferenceStatusable: error deferencing %s: %s", remoteStatusID.String(), err)
	}

	m := make(map[string]interface{})
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, fmt.Errorf("DereferenceStatusable: error unmarshalling bytes into json: %s", err)
	}

	t, err := streams.ToType(context.Background(), m)
	if err != nil {
		return nil, fmt.Errorf("DereferenceStatusable: error resolving json into ap vocab type: %s", err)
	}

	// Article, Document, Image, Video, Note, Page, Event, Place, Mention, Profile
	switch t.GetTypeName() {
	case gtsmodel.ActivityStreamsArticle:
		p, ok := t.(vocab.ActivityStreamsArticle)
		if !ok {
			return nil, errors.New("DereferenceStatusable: error resolving type as ActivityStreamsArticle")
		}
		return p, nil
	case gtsmodel.ActivityStreamsDocument:
		p, ok := t.(vocab.ActivityStreamsDocument)
		if !ok {
			return nil, errors.New("DereferenceStatusable: error resolving type as ActivityStreamsDocument")
		}
		return p, nil
	case gtsmodel.ActivityStreamsImage:
		p, ok := t.(vocab.ActivityStreamsImage)
		if !ok {
			return nil, errors.New("DereferenceStatusable: error resolving type as ActivityStreamsImage")
		}
		return p, nil
	case gtsmodel.ActivityStreamsVideo:
		p, ok := t.(vocab.ActivityStreamsVideo)
		if !ok {
			return nil, errors.New("DereferenceStatusable: error resolving type as ActivityStreamsVideo")
		}
		return p, nil
	case gtsmodel.ActivityStreamsNote:
		p, ok := t.(vocab.ActivityStreamsNote)
		if !ok {
			return nil, errors.New("DereferenceStatusable: error resolving type as ActivityStreamsNote")
		}
		return p, nil
	case gtsmodel.ActivityStreamsPage:
		p, ok := t.(vocab.ActivityStreamsPage)
		if !ok {
			return nil, errors.New("DereferenceStatusable: error resolving type as ActivityStreamsPage")
		}
		return p, nil
	case gtsmodel.ActivityStreamsEvent:
		p, ok := t.(vocab.ActivityStreamsEvent)
		if !ok {
			return nil, errors.New("DereferenceStatusable: error resolving type as ActivityStreamsEvent")
		}
		return p, nil
	case gtsmodel.ActivityStreamsPlace:
		p, ok := t.(vocab.ActivityStreamsPlace)
		if !ok {
			return nil, errors.New("DereferenceStatusable: error resolving type as ActivityStreamsPlace")
		}
		return p, nil
	case gtsmodel.ActivityStreamsProfile:
		p, ok := t.(vocab.ActivityStreamsProfile)
		if !ok {
			return nil, errors.New("DereferenceStatusable: error resolving type as ActivityStreamsProfile")
		}
		return p, nil
	}

	return nil, fmt.Errorf("DereferenceStatusable: type name %s not supported", t.GetTypeName())
}

// populateStatusFields fetches all the information we temporarily pinned to an incoming
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
func (d *deref) populateStatusFields(ctx context.Context, status *gtsmodel.Status, requestingUsername string) error {
	l := d.log.WithFields(logrus.Fields{
		"func":   "dereferenceStatusFields",
		"status": fmt.Sprintf("%+v", status),
	})
	l.Debug("entering function")

	// make sure we have a status URI and that the domain in question isn't blocked
	statusURI, err := url.Parse(status.URI)
	if err != nil {
		return fmt.Errorf("DereferenceStatusFields: couldn't parse status URI %s: %s", status.URI, err)
	}
	if blocked, err := d.db.IsDomainBlocked(ctx, statusURI.Host); blocked || err != nil {
		return fmt.Errorf("DereferenceStatusFields: domain %s is blocked", statusURI.Host)
	}

	// we can continue -- create a new transport here because we'll probably need it
	t, err := d.transportController.NewTransportForUsername(ctx, requestingUsername)
	if err != nil {
		return fmt.Errorf("error creating transport: %s", err)
	}

	// in case the status doesn't have an id yet (ie., it hasn't entered the database yet), then create one
	if status.ID == "" {
		newID, err := id.NewULIDFromTime(status.CreatedAt)
		if err != nil {
			return err
		}
		status.ID = newID
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
	for _, a := range status.Attachments {
		l.Tracef("dereferencing attachment: %+v", a)

		// it might have been processed elsewhere so check first if it's already in the database or not
		maybeAttachment := &gtsmodel.MediaAttachment{}
		err := d.db.GetWhere(ctx, []db.Where{{Key: "remote_url", Value: a.RemoteURL}}, maybeAttachment)
		if err == nil {
			// we already have it in the db, dereferenced, no need to do it again
			l.Tracef("attachment already exists with id %s", maybeAttachment.ID)
			attachmentIDs = append(attachmentIDs, maybeAttachment.ID)
			continue
		}
		if err != db.ErrNoEntries {
			// we have a real error
			return fmt.Errorf("error checking db for existence of attachment with remote url %s: %s", a.RemoteURL, err)
		}
		// it just doesn't exist yet so carry on
		l.Debug("attachment doesn't exist yet, calling ProcessRemoteAttachment", a)
		deferencedAttachment, err := d.mediaHandler.ProcessRemoteAttachment(ctx, t, a, status.AccountID)
		if err != nil {
			l.Errorf("error dereferencing status attachment: %s", err)
			continue
		}
		l.Debugf("dereferenced attachment: %+v", deferencedAttachment)
		deferencedAttachment.StatusID = status.ID
		deferencedAttachment.Description = a.Description
		if err := d.db.Put(ctx, deferencedAttachment); err != nil {
			return fmt.Errorf("error inserting dereferenced attachment with remote url %s: %s", a.RemoteURL, err)
		}
		attachmentIDs = append(attachmentIDs, deferencedAttachment.ID)
	}
	status.AttachmentIDs = attachmentIDs

	// 2. Hashtags

	// 3. Emojis

	// 4. Mentions
	// At this point, mentions should have the namestring and mentionedAccountURI set on them.
	//
	// We should dereference any accounts mentioned here which we don't have in our db yet, by their URI.
	mentionIDs := []string{}
	for _, m := range status.Mentions {
		if m.ID != "" {
			// we've already populated this mention, since it has an ID
			l.Debug("mention already populated")
			continue
		}

		if m.TargetAccountURI == "" {
			// can't do anything with this mention
			l.Debug("target URI not set on mention")
			continue
		}

		targetAccountURI, err := url.Parse(m.TargetAccountURI)
		if err != nil {
			l.Debugf("error parsing mentioned account uri %s: %s", m.TargetAccountURI, err)
			continue
		}

		var targetAccount *gtsmodel.Account
		if a, err := d.db.GetAccountByURL(ctx, targetAccountURI.String()); err == nil {
			targetAccount = a
		} else if a, _, err := d.GetRemoteAccount(ctx, requestingUsername, targetAccountURI, false); err == nil {
			targetAccount = a
		} else {
			// we can't find the target account so bail
			l.Debug("can't retrieve account targeted by mention")
			continue
		}

		mID, err := id.NewRandomULID()
		if err != nil {
			return err
		}

		m = &gtsmodel.Mention{
			ID:               mID,
			StatusID:         status.ID,
			Status:           m.Status,
			CreatedAt:        status.CreatedAt,
			UpdatedAt:        status.UpdatedAt,
			OriginAccountID:  status.Account.ID,
			OriginAccountURI: status.AccountURI,
			OriginAccount:    status.Account,
			TargetAccountID:  targetAccount.ID,
			TargetAccount:    targetAccount,
			NameString:       m.NameString,
			TargetAccountURI: targetAccount.URI,
			TargetAccountURL: targetAccount.URL,
		}

		if err := d.db.Put(ctx, m); err != nil {
			return fmt.Errorf("error creating mention: %s", err)
		}
		mentionIDs = append(mentionIDs, m.ID)
	}
	status.MentionIDs = mentionIDs

	// status has replyToURI but we don't have an ID yet for the status it replies to
	if status.InReplyToURI != "" && status.InReplyToID == "" {
		statusURI, err := url.Parse(status.InReplyToURI)
		if err != nil {
			return err
		}
		if replyToStatus, err := d.db.GetStatusByURI(ctx, status.InReplyToURI); err == nil {
			// we have the status
			status.InReplyToID = replyToStatus.ID
			status.InReplyTo = replyToStatus
			status.InReplyToAccountID = replyToStatus.AccountID
			status.InReplyToAccount = replyToStatus.Account
		} else if replyToStatus, _, _, err := d.GetRemoteStatus(ctx, requestingUsername, statusURI, false); err == nil {
			// we got the status
			status.InReplyToID = replyToStatus.ID
			status.InReplyTo = replyToStatus
			status.InReplyToAccountID = replyToStatus.AccountID
			status.InReplyToAccount = replyToStatus.Account
		}
	}
	return nil
}

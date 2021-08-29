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
	"strings"

	"github.com/go-fed/activity/streams"
	"github.com/go-fed/activity/streams/vocab"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
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
// 5. Replied-to-status.
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

	statusIRI, err := url.Parse(status.URI)
	if err != nil {
		return fmt.Errorf("populateStatusFields: couldn't parse status URI %s: %s", status.URI, err)
	}

	blocked, err := d.db.IsURIBlocked(ctx, statusIRI)
	if err != nil {
		return fmt.Errorf("populateStatusFields: error checking blocked status of %s: %s", statusIRI, err)
	}
	if blocked {
		return fmt.Errorf("populateStatusFields: domain %s is blocked", statusIRI)
	}

	// in case the status doesn't have an id yet (ie., it hasn't entered the database yet), then create one
	if status.ID == "" {
		newID, err := id.NewULIDFromTime(status.CreatedAt)
		if err != nil {
			return fmt.Errorf("populateStatusFields: error creating ulid for status: %s", err)
		}
		status.ID = newID
	}

	// 1. Media attachments.
	if err := d.populateStatusAttachments(ctx, status, requestingUsername); err != nil {
		return fmt.Errorf("populateStatusFields: error populating status attachments: %s", err)
	}

	// 2. Hashtags
	// TODO

	// 3. Emojis
	// TODO

	// 4. Mentions
	if err := d.populateStatusMentions(ctx, status, requestingUsername); err != nil {
		return fmt.Errorf("populateStatusFields: error populating status mentions: %s", err)
	}

	// 5. Replied-to-status.
	if err := d.populateStatusRepliedTo(ctx, status, requestingUsername); err != nil {
		return fmt.Errorf("populateStatusFields: error populating status repliedTo: %s", err)
	}

	return nil
}

func (d *deref) populateStatusMentions(ctx context.Context, status *gtsmodel.Status, requestingUsername string) error {
	l := d.log

	// At this point, mentions should have the namestring and mentionedAccountURI set on them.
	// We can use these to find the accounts.

	mentionIDs := []string{}
	newMentions := []*gtsmodel.Mention{}
	for _, m := range status.Mentions {
		if m.ID != "" {
			// we've already populated this mention, since it has an ID
			l.Debug("populateStatusMentions: mention already populated")
			mentionIDs = append(mentionIDs, m.ID)
			newMentions = append(newMentions, m)
			continue
		}

		if m.TargetAccountURI == "" {
			l.Debug("populateStatusMentions: target URI not set on mention")
			continue
		}

		targetAccountURI, err := url.Parse(m.TargetAccountURI)
		if err != nil {
			l.Debugf("populateStatusMentions: error parsing mentioned account uri %s: %s", m.TargetAccountURI, err)
			continue
		}

		var targetAccount *gtsmodel.Account
		errs := []string{}

		// check if account is in the db already
		if a, err := d.db.GetAccountByURL(ctx, targetAccountURI.String()); err != nil {
			errs = append(errs, err.Error())
		} else {
			l.Debugf("populateStatusMentions: got target account %s with id %s through GetAccountByURL", targetAccountURI, a.ID)
			targetAccount = a
		}

		if targetAccount == nil {
			// we didn't find the account in our database already
			// check if we can get the account remotely (dereference it)
			if a, _, err := d.GetRemoteAccount(ctx, requestingUsername, targetAccountURI, false); err != nil {
				errs = append(errs, err.Error())
			} else {
				l.Debugf("populateStatusMentions: got target account %s with id %s through GetRemoteAccount", targetAccountURI, a.ID)
				targetAccount = a
			}
		}

		if targetAccount == nil {
			l.Debugf("populateStatusMentions: couldn't get target account %s: %s", m.TargetAccountURI, strings.Join(errs, " : "))
			continue
		}

		mID, err := id.NewRandomULID()
		if err != nil {
			return fmt.Errorf("populateStatusMentions: error generating ulid: %s", err)
		}

		newMention := &gtsmodel.Mention{
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

		if err := d.db.Put(ctx, newMention); err != nil {
			return fmt.Errorf("populateStatusMentions: error creating mention: %s", err)
		}

		mentionIDs = append(mentionIDs, newMention.ID)
		newMentions = append(newMentions, newMention)
	}

	status.MentionIDs = mentionIDs
	status.Mentions = newMentions

	return nil
}

func (d *deref) populateStatusAttachments(ctx context.Context, status *gtsmodel.Status, requestingUsername string) error {
	l := d.log

	// At this point we should know:
	// * the media type of the file we're looking for (a.File.ContentType)
	// * the file type (a.Type)
	// * the remote URL (a.RemoteURL)
	// This should be enough to dereference the piece of media.

	attachmentIDs := []string{}
	attachments := []*gtsmodel.MediaAttachment{}

	for _, a := range status.Attachments {

		aURL, err := url.Parse(a.RemoteURL)
		if err != nil {
			l.Errorf("populateStatusAttachments: couldn't parse attachment url %s: %s", a.RemoteURL, err)
			continue
		}

		attachment, err := d.GetRemoteAttachment(ctx, requestingUsername, aURL, status.AccountID, status.ID, a.File.ContentType)
		if err != nil {
			l.Errorf("populateStatusAttachments: couldn't get remote attachment %s: %s", a.RemoteURL, err)
		}

		attachmentIDs = append(attachmentIDs, attachment.ID)
		attachments = append(attachments, attachment)
	}

	status.AttachmentIDs = attachmentIDs
	status.Attachments = attachments

	return nil
}

func (d *deref) populateStatusRepliedTo(ctx context.Context, status *gtsmodel.Status, requestingUsername string) error {
	if status.InReplyToURI != "" && status.InReplyToID == "" {
		statusURI, err := url.Parse(status.InReplyToURI)
		if err != nil {
			return err
		}

		var replyToStatus *gtsmodel.Status
		errs := []string{}

		// see if we have the status in our db already
		if s, err := d.db.GetStatusByURI(ctx, status.InReplyToURI); err != nil {
			errs = append(errs, err.Error())
		} else {
			replyToStatus = s
		}

		if replyToStatus == nil {
			// didn't find the status in our db, try to get it remotely
			if s, _, _, err := d.GetRemoteStatus(ctx, requestingUsername, statusURI, false); err != nil {
				errs = append(errs, err.Error())
			} else {
				replyToStatus = s
			}
		}

		if replyToStatus == nil {
			return fmt.Errorf("populateStatusRepliedTo: couldn't get reply to status with uri %s: %s", statusURI, strings.Join(errs, " : "))
		}

		// we have the status
		status.InReplyToID = replyToStatus.ID
		status.InReplyTo = replyToStatus
		status.InReplyToAccountID = replyToStatus.AccountID
		status.InReplyToAccount = replyToStatus.Account
	}

	return nil
}

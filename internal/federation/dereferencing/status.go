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
func (d *deref) EnrichRemoteStatus(username string, status *gtsmodel.Status) (*gtsmodel.Status, error) {
	if err := d.populateStatusFields(status, username); err != nil {
		return nil, err
	}

	if err := d.db.UpdateByID(status.ID, status); err != nil {
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
func (d *deref) GetRemoteStatus(username string, remoteStatusID *url.URL, refresh bool) (*gtsmodel.Status, ap.Statusable, bool, error) {
	new := true

	// check if we already have the status in our db
	maybeStatus := &gtsmodel.Status{}
	if err := d.db.GetWhere([]db.Where{{Key: "uri", Value: remoteStatusID.String()}}, maybeStatus); err == nil {
		// we've seen this status before so it's not new
		new = false

		// if we're not being asked to refresh, we can just return the maybeStatus as-is and avoid doing any external calls
		if !refresh {
			return maybeStatus, nil, new, nil
		}
	}

	statusable, err := d.dereferenceStatusable(username, remoteStatusID)
	if err != nil {
		return nil, statusable, new, fmt.Errorf("GetRemoteStatus: error dereferencing statusable: %s", err)
	}

	accountURI, err := ap.ExtractAttributedTo(statusable)
	if err != nil {
		return nil, statusable, new, fmt.Errorf("GetRemoteStatus: error extracting attributedTo: %s", err)
	}

	// do this so we know we have the remote account of the status in the db
	_, _, err = d.GetRemoteAccount(username, accountURI, refresh)
	if err != nil {
		return nil, statusable, new, fmt.Errorf("GetRemoteStatus: couldn't derive status author: %s", err)
	}

	gtsStatus, err := d.typeConverter.ASStatusToStatus(statusable)
	if err != nil {
		return nil, statusable, new, fmt.Errorf("GetRemoteStatus: error converting statusable to status: %s", err)
	}

	if new {
		ulid, err := id.NewRandomULID()
		if err != nil {
			return nil, statusable, new, fmt.Errorf("GetRemoteStatus: error generating new id for status: %s", err)
		}
		gtsStatus.ID = ulid

		if err := d.populateStatusFields(gtsStatus, username); err != nil {
			return nil, statusable, new, fmt.Errorf("GetRemoteStatus: error populating status fields: %s", err)
		}

		if err := d.db.Put(gtsStatus); err != nil {
			return nil, statusable, new, fmt.Errorf("GetRemoteStatus: error putting new status: %s", err)
		}
	} else {
		gtsStatus.ID = maybeStatus.ID

		if err := d.populateStatusFields(gtsStatus, username); err != nil {
			return nil, statusable, new, fmt.Errorf("GetRemoteStatus: error populating status fields: %s", err)
		}

		if err := d.db.UpdateByID(gtsStatus.ID, gtsStatus); err != nil {
			return nil, statusable, new, fmt.Errorf("GetRemoteStatus: error updating status: %s", err)
		}
	}

	return gtsStatus, statusable, new, nil
}

func (d *deref) dereferenceStatusable(username string, remoteStatusID *url.URL) (ap.Statusable, error) {
	if blocked, err := d.blockedDomain(remoteStatusID.Host); blocked || err != nil {
		return nil, fmt.Errorf("DereferenceStatusable: domain %s is blocked", remoteStatusID.Host)
	}

	transport, err := d.transportController.NewTransportForUsername(username)
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
func (d *deref) populateStatusFields(status *gtsmodel.Status, requestingUsername string) error {
	l := d.log.WithFields(logrus.Fields{
		"func":   "dereferenceStatusFields",
		"status": fmt.Sprintf("%+v", status),
	})
	l.Debug("entering function")

	statusURI, err := url.Parse(status.URI)
	if err != nil {
		return fmt.Errorf("DereferenceStatusFields: couldn't parse status URI %s: %s", status.URI, err)
	}
	if blocked, err := d.blockedDomain(statusURI.Host); blocked || err != nil {
		return fmt.Errorf("DereferenceStatusFields: domain %s is blocked", statusURI.Host)
	}

	t, err := d.transportController.NewTransportForUsername(requestingUsername)
	if err != nil {
		return fmt.Errorf("error creating transport: %s", err)
	}

	// in case the status doesn't have an id yet (ie., it hasn't entered the database yet)
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
	for _, a := range status.GTSMediaAttachments {
		l.Debugf("dereferencing attachment: %+v", a)

		// it might have been processed elsewhere so check first if it's already in the database or not
		maybeAttachment := &gtsmodel.MediaAttachment{}
		err := d.db.GetWhere([]db.Where{{Key: "remote_url", Value: a.RemoteURL}}, maybeAttachment)
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
		deferencedAttachment, err := d.mediaHandler.ProcessRemoteAttachment(t, a, status.AccountID)
		if err != nil {
			l.Errorf("error dereferencing status attachment: %s", err)
			continue
		}
		l.Debugf("dereferenced attachment: %+v", deferencedAttachment)
		deferencedAttachment.StatusID = status.ID
		deferencedAttachment.Description = a.Description
		if err := d.db.Put(deferencedAttachment); err != nil {
			return fmt.Errorf("error inserting dereferenced attachment with remote url %s: %s", a.RemoteURL, err)
		}
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
		if m.ID == "" {
			mID, err := id.NewRandomULID()
			if err != nil {
				return err
			}
			m.ID = mID
		}

		uri, err := url.Parse(m.MentionedAccountURI)
		if err != nil {
			l.Debugf("error parsing mentioned account uri %s: %s", m.MentionedAccountURI, err)
			continue
		}

		m.StatusID = status.ID
		m.OriginAccountID = status.GTSAuthorAccount.ID
		m.OriginAccountURI = status.GTSAuthorAccount.URI

		targetAccount := &gtsmodel.Account{}
		if err := d.db.GetWhere([]db.Where{{Key: "uri", Value: uri.String()}}, targetAccount); err != nil {
			// proper error
			if _, ok := err.(db.ErrNoEntries); !ok {
				return fmt.Errorf("db error checking for account with uri %s", uri.String())
			}

			// we just don't have it yet, so we should go get it....
			accountable, err := d.dereferenceAccountable(requestingUsername, uri)
			if err != nil {
				// we can't dereference it so just skip it
				l.Debugf("error dereferencing remote account with uri %s: %s", uri.String(), err)
				continue
			}

			targetAccount, err = d.typeConverter.ASRepresentationToAccount(accountable, false)
			if err != nil {
				l.Debugf("error converting remote account with uri %s into gts model: %s", uri.String(), err)
				continue
			}

			targetAccountID, err := id.NewRandomULID()
			if err != nil {
				return err
			}
			targetAccount.ID = targetAccountID

			if err := d.db.Put(targetAccount); err != nil {
				return fmt.Errorf("db error inserting account with uri %s", uri.String())
			}
		}

		// by this point, we know the targetAccount exists in our database with an ID :)
		m.TargetAccountID = targetAccount.ID
		if err := d.db.Put(m); err != nil {
			return fmt.Errorf("error creating mention: %s", err)
		}
		mentions = append(mentions, m.ID)
	}
	status.Mentions = mentions

	// status has replyToURI but we don't have an ID yet for the status it replies to
	if status.InReplyToURI != "" && status.InReplyToID == "" {
		replyToStatus := &gtsmodel.Status{}
		if err := d.db.GetWhere([]db.Where{{Key: "uri", Value: status.InReplyToURI}}, replyToStatus); err == nil {
			// we have the status
			status.InReplyToID = replyToStatus.ID
			status.InReplyToAccountID = replyToStatus.AccountID
		}
	}

	return nil
}

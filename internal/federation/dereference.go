package federation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

	"github.com/go-fed/activity/streams"
	"github.com/go-fed/activity/streams/vocab"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/transport"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
)

func (f *federator) findOldestAncestor(username string, statusIRI url.URL) (*url.URL, error) {
	// start by assuming the status we have is the oldest -- this might change as we work through the chain
	oldestIRI := &statusIRI

searchLoop:
	for {
		// begin by checking if we already have this oldestIRI in our database
		gtsStatus := &gtsmodel.Status{}
		if err := f.db.GetWhere([]db.Where{{Key: "uri", Value: oldestIRI.String()}}, gtsStatus); err == nil {
			// nice, we already have it as a gts status so our life just got easier

			var inReplyToURI string
			if gtsStatus.InReplyToURI != "" {
				// we already have the replyToURI on the current oldest status
				inReplyToURI = gtsStatus.InReplyToURI
			} else if gtsStatus.InReplyToID != "" {
				// we don't have the replyToURI, but we do have a status ID so the status replied to should be in our db already...
				repliedGTSStatus := &gtsmodel.Status{}
				if err := f.db.GetByID(gtsStatus.InReplyToID, repliedGTSStatus); err != nil {
					return nil, fmt.Errorf("database error getting status with id %s: %s", gtsStatus.InReplyToID, err)
				}
				inReplyToURI = repliedGTSStatus.URI
			} else {
				// this status doesn't reply to anything
				break searchLoop
			}

			// set the oldestIRI to the parent we just found, and go to the next iteration
			repliedGTSStatusIRI, err := url.Parse(inReplyToURI)
			if err != nil {
				return nil, fmt.Errorf("error parsing URI %s: %s", inReplyToURI, err)
			}
			oldestIRI = repliedGTSStatusIRI
			continue
		}

		// we don't have the oldestIRI status in our database yet so we'll have to dereference it to see if it replies to something
		statusable, err := f.DereferenceRemoteStatus(username, oldestIRI)
		if err != nil {
			return nil, fmt.Errorf("error dereferencing %s: %s", oldestIRI.String(), err)
		}

		inReplyToURI := typeutils.ExtractInReplyToURI(statusable)
		if inReplyToURI != nil {
			oldestIRI = inReplyToURI
			continue
		}

		// if we reach this point we couldn't find something older than oldestIRI so we're done
		break searchLoop
	}

	return oldestIRI, nil
}

func (f *federator) stashDescendants(username string, statusIRI url.URL) error {
	l := f.log.WithFields(logrus.Fields{
		"func":      "stashDescendants",
		"username":  username,
		"statusIRI": statusIRI.String(),
	})
	l.Debug("entering stashDescendants")

	// if it's our status we already have descendants stashed so we can bail early
	if statusIRI.Host == f.config.Host {
		l.Debug("iri belongs to us, bailing")
		return nil
	}

	// fetch the remote representation of the given status
	statusable, err := f.DereferenceRemoteStatus(username, &statusIRI)
	if err != nil {
		return fmt.Errorf("stashDescendants: error dereferencing status with id %s: %s", statusIRI.String(), err)
	}

	if err := f.FullyDereferenceStatusableAndAccount(username, statusable); err != nil {
		return fmt.Errorf("stashDescendants: error fully dereferencing statusable: %s", err)
	}

	replies := statusable.GetActivityStreamsReplies()
	if replies == nil || !replies.IsActivityStreamsCollection() {
		l.Debug("no replies, bailing")
		return nil
	}

	repliesCollection := replies.GetActivityStreamsCollection()
	if repliesCollection == nil {
		l.Debug("replies collection is nil, bailing")
		return nil
	}

	first := repliesCollection.GetActivityStreamsFirst()
	if first == nil {
		l.Debug("replies collection has no first, bailing")
		return nil
	}

	firstPage := first.GetActivityStreamsCollectionPage()
	if firstPage == nil {
		l.Debug("first has no collection page, bailing")
		return nil
	}

	firstPageNext := firstPage.GetActivityStreamsNext()
	if firstPageNext == nil || !firstPageNext.IsIRI() {
		l.Debug("next is not an iri, bailing")
		return nil
	}

	var foundReplies int
	currentPageIRI := firstPageNext.GetIRI()

pageLoop:
	for {
		l.Debugf("dereferencing page %s", currentPageIRI)
		nextPage, err := f.DereferenceCollectionPage(username, currentPageIRI)
		if err != nil {
			return nil
		}

		nextItems := typeutils.ExtractURLItems(nextPage)
		if len(nextItems) == 0 {
			// no items on this page, which means we're done
			break pageLoop
		}

		for _, i := range nextItems {
			if i.Host == f.config.Host {
				// skip if the reply is from us -- we already have it then
				continue
			}
			foundReplies = foundReplies + 1
			f.stashDescendants(username, *i)
		}

		next := nextPage.GetActivityStreamsNext()
		if next != nil && next.IsIRI() {
			l.Debug("setting next page")
			currentPageIRI = next.GetIRI()
		} else {
			l.Debug("no next page, bailing")
			break pageLoop
		}
	}

	l.Debugf("foundReplies %d", foundReplies)
	return nil
}

func (f *federator) FullyDereferenceStatusableAndAccount(username string, statusable typeutils.Statusable) error {

	idProp := statusable.GetJSONLDId()
	if idProp == nil || !idProp.IsIRI() {
		return errors.New("FullyDereferenceStatusableAndAccount: couldn't extract iri from statusable")
	}

	// make sure we don't already have this status in our db
	if err := f.db.GetWhere([]db.Where{{Key: "uri", Value: idProp.GetIRI().String()}}, &gtsmodel.Status{}); err == nil {
		// we already have it
		return nil
	}

	// make sure we have the author account in the db
	attributedToProp := statusable.GetActivityStreamsAttributedTo()
	for iter := attributedToProp.Begin(); iter != attributedToProp.End(); iter = iter.Next() {
		if !iter.IsIRI() {
			continue
		}

		accountURI := iter.GetIRI()
		if err := f.db.GetWhere([]db.Where{{Key: "uri", Value: accountURI.String()}}, &gtsmodel.Account{}); err == nil {
			// we already have it, nice
			continue
		}

		// we don't have the status author account yet so dereference it
		accountable, err := f.DereferenceRemoteAccount(username, accountURI)
		if err != nil {
			return fmt.Errorf("FullyDereferenceStatusAndAccount: error dereferencing remote account with id %s: %s", accountURI.String(), err)
		}
		account, err := f.typeConverter.ASRepresentationToAccount(accountable, false)
		if err != nil {
			return fmt.Errorf("FullyDereferenceStatusAndAccount: error converting dereferenced account with id %s into account : %s", accountURI.String(), err)
		}

		accountID, err := id.NewRandomULID()
		if err != nil {
			return err
		}
		account.ID = accountID

		if err := f.db.Put(account); err != nil {
			return fmt.Errorf("FullyDereferenceStatusAndAccount: error putting dereferenced account with id %s into database : %s", accountURI.String(), err)
		}

		if err := f.DereferenceAccountFields(account, username, false); err != nil {
			return fmt.Errorf("FullyDereferenceStatusAndAccount: error dereferencing fields on account with id %s : %s", accountURI.String(), err)
		}
	}

	gtsStatus, err := f.typeConverter.ASStatusToStatus(statusable)
	if err != nil {
		return fmt.Errorf("FullyDereferenceStatusAndAccount: error converting statusable: %s", err)
	}

	id, err := id.NewULIDFromTime(gtsStatus.CreatedAt)
	if err != nil {
		return fmt.Errorf("FullyDereferenceStatusAndAccount: error generating id: %s", err)
	}
	gtsStatus.ID = id

	if err := f.db.Put(gtsStatus); err != nil {
		return fmt.Errorf("FullyDereferenceStatusAndAccount: error putting dereferenced status with id %s into the db: %s", gtsStatus.URI, err)
	}

	// now dereference additional fields straight away (we're already async here so we have time)
	if err := f.DereferenceStatusFields(gtsStatus, username); err != nil {
		return fmt.Errorf("FullyDereferenceStatusAndAccount: error dereferencing status fields for status with id %s: %s", gtsStatus.URI, err)
	}

	// update with the newly dereferenced fields
	if err := f.db.UpdateByID(gtsStatus.ID, gtsStatus); err != nil {
		return fmt.Errorf("FullyDereferenceStatusAndAccount: error updating dereferenced status in the db: %s", err)
	}

	return nil
}

// DereferenceRemoteThread takes a statusable (something that has withReplies and withInReplyTo),
// and dereferences statusables in the conversation that are CC or TO public.
//
// This process involves working up and down the chain of replies, and parsing through the collections of IDs
// presented by remote instances as part of their replies collections, and will likely involve making several calls to
// multiple different hosts.
func (f *federator) DereferenceRemoteThread(username string, statusIRI *url.URL) error {
	// first we need to find the oldest ancestor of this thread, or as near as we can get
	oldestIRI, err := f.findOldestAncestor(username, *statusIRI)
	if err != nil {
		return fmt.Errorf("error finding oldest ancestor of status %s: %s", statusIRI.String(), err)
	}

	// now that we have the oldest, we can work our way down from that to find descendants/replies
	return f.stashDescendants(username, *oldestIRI)
}

// DereferenceCollectionPage returns the activitystreams CollectionPage at the specified IRI, or an error if something goes wrong.
func (f *federator) DereferenceCollectionPage(username string, pageIRI *url.URL) (typeutils.CollectionPageable, error) {
	if blocked, err := f.blockedDomain(pageIRI.Host); blocked || err != nil {
		return nil, fmt.Errorf("DereferenceCollectionPage: domain %s is blocked", pageIRI.Host)
	}

	transport, err := f.GetTransportForUser(username)
	if err != nil {
		return nil, fmt.Errorf("transport err: %s", err)
	}

	b, err := transport.Dereference(context.Background(), pageIRI)
	if err != nil {
		return nil, fmt.Errorf("error deferencing %s: %s", pageIRI.String(), err)
	}

	m := make(map[string]interface{})
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, fmt.Errorf("error unmarshalling bytes into json: %s", err)
	}

	t, err := streams.ToType(context.Background(), m)
	if err != nil {
		return nil, fmt.Errorf("error resolving json into ap vocab type: %s", err)
	}

	if t.GetTypeName() != gtsmodel.ActivityStreamsCollectionPage {
		return nil, fmt.Errorf("type name %s not supported", t.GetTypeName())
	}

	p, ok := t.(vocab.ActivityStreamsCollectionPage)
	if !ok {
		return nil, errors.New("error resolving type as activitystreams collection page")
	}

	return p, nil
}

func (f *federator) DereferenceRemoteAccount(username string, remoteAccountID *url.URL) (typeutils.Accountable, error) {
	f.startHandshake(username, remoteAccountID)
	defer f.stopHandshake(username, remoteAccountID)

	if blocked, err := f.blockedDomain(remoteAccountID.Host); blocked || err != nil {
		return nil, fmt.Errorf("DereferenceRemoteAccount: domain %s is blocked", remoteAccountID.Host)
	}

	transport, err := f.GetTransportForUser(username)
	if err != nil {
		return nil, fmt.Errorf("transport err: %s", err)
	}

	b, err := transport.Dereference(context.Background(), remoteAccountID)
	if err != nil {
		return nil, fmt.Errorf("error deferencing %s: %s", remoteAccountID.String(), err)
	}

	m := make(map[string]interface{})
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, fmt.Errorf("error unmarshalling bytes into json: %s", err)
	}

	t, err := streams.ToType(context.Background(), m)
	if err != nil {
		return nil, fmt.Errorf("error resolving json into ap vocab type: %s", err)
	}

	switch t.GetTypeName() {
	case string(gtsmodel.ActivityStreamsPerson):
		p, ok := t.(vocab.ActivityStreamsPerson)
		if !ok {
			return nil, errors.New("error resolving type as activitystreams person")
		}
		return p, nil
	case string(gtsmodel.ActivityStreamsApplication):
		p, ok := t.(vocab.ActivityStreamsApplication)
		if !ok {
			return nil, errors.New("error resolving type as activitystreams application")
		}
		return p, nil
	case string(gtsmodel.ActivityStreamsService):
		p, ok := t.(vocab.ActivityStreamsService)
		if !ok {
			return nil, errors.New("error resolving type as activitystreams service")
		}
		return p, nil
	}

	return nil, fmt.Errorf("type name %s not supported", t.GetTypeName())
}

func (f *federator) DereferenceRemoteStatus(username string, remoteStatusID *url.URL) (typeutils.Statusable, error) {
	if blocked, err := f.blockedDomain(remoteStatusID.Host); blocked || err != nil {
		return nil, fmt.Errorf("DereferenceRemoteStatus: domain %s is blocked", remoteStatusID.Host)
	}

	transport, err := f.GetTransportForUser(username)
	if err != nil {
		return nil, fmt.Errorf("transport err: %s", err)
	}

	b, err := transport.Dereference(context.Background(), remoteStatusID)
	if err != nil {
		return nil, fmt.Errorf("error deferencing %s: %s", remoteStatusID.String(), err)
	}

	m := make(map[string]interface{})
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, fmt.Errorf("error unmarshalling bytes into json: %s", err)
	}

	t, err := streams.ToType(context.Background(), m)
	if err != nil {
		return nil, fmt.Errorf("error resolving json into ap vocab type: %s", err)
	}

	// Article, Document, Image, Video, Note, Page, Event, Place, Mention, Profile
	switch t.GetTypeName() {
	case gtsmodel.ActivityStreamsArticle:
		p, ok := t.(vocab.ActivityStreamsArticle)
		if !ok {
			return nil, errors.New("error resolving type as ActivityStreamsArticle")
		}
		return p, nil
	case gtsmodel.ActivityStreamsDocument:
		p, ok := t.(vocab.ActivityStreamsDocument)
		if !ok {
			return nil, errors.New("error resolving type as ActivityStreamsDocument")
		}
		return p, nil
	case gtsmodel.ActivityStreamsImage:
		p, ok := t.(vocab.ActivityStreamsImage)
		if !ok {
			return nil, errors.New("error resolving type as ActivityStreamsImage")
		}
		return p, nil
	case gtsmodel.ActivityStreamsVideo:
		p, ok := t.(vocab.ActivityStreamsVideo)
		if !ok {
			return nil, errors.New("error resolving type as ActivityStreamsVideo")
		}
		return p, nil
	case gtsmodel.ActivityStreamsNote:
		p, ok := t.(vocab.ActivityStreamsNote)
		if !ok {
			return nil, errors.New("error resolving type as ActivityStreamsNote")
		}
		return p, nil
	case gtsmodel.ActivityStreamsPage:
		p, ok := t.(vocab.ActivityStreamsPage)
		if !ok {
			return nil, errors.New("error resolving type as ActivityStreamsPage")
		}
		return p, nil
	case gtsmodel.ActivityStreamsEvent:
		p, ok := t.(vocab.ActivityStreamsEvent)
		if !ok {
			return nil, errors.New("error resolving type as ActivityStreamsEvent")
		}
		return p, nil
	case gtsmodel.ActivityStreamsPlace:
		p, ok := t.(vocab.ActivityStreamsPlace)
		if !ok {
			return nil, errors.New("error resolving type as ActivityStreamsPlace")
		}
		return p, nil
	case gtsmodel.ActivityStreamsProfile:
		p, ok := t.(vocab.ActivityStreamsProfile)
		if !ok {
			return nil, errors.New("error resolving type as ActivityStreamsProfile")
		}
		return p, nil
	}

	return nil, fmt.Errorf("type name %s not supported", t.GetTypeName())
}

func (f *federator) DereferenceRemoteInstance(username string, remoteInstanceURI *url.URL) (*gtsmodel.Instance, error) {
	if blocked, err := f.blockedDomain(remoteInstanceURI.Host); blocked || err != nil {
		return nil, fmt.Errorf("DereferenceRemoteInstance: domain %s is blocked", remoteInstanceURI.Host)
	}

	transport, err := f.GetTransportForUser(username)
	if err != nil {
		return nil, fmt.Errorf("transport err: %s", err)
	}

	return transport.DereferenceInstance(context.Background(), remoteInstanceURI)
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
func (f *federator) DereferenceStatusFields(status *gtsmodel.Status, requestingUsername string) error {
	l := f.log.WithFields(logrus.Fields{
		"func":   "dereferenceStatusFields",
		"status": fmt.Sprintf("%+v", status),
	})
	l.Debug("entering function")

	statusURI, err := url.Parse(status.URI)
	if err != nil {
		return fmt.Errorf("DereferenceStatusFields: couldn't parse status URI %s: %s", status.URI, err)
	}
	if blocked, err := f.blockedDomain(statusURI.Host); blocked || err != nil {
		return fmt.Errorf("DereferenceStatusFields: domain %s is blocked", statusURI.Host)
	}

	t, err := f.GetTransportForUser(requestingUsername)
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
		err := f.db.GetWhere([]db.Where{{Key: "remote_url", Value: a.RemoteURL}}, maybeAttachment)
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
		deferencedAttachment, err := f.mediaHandler.ProcessRemoteAttachment(t, a, status.AccountID)
		if err != nil {
			l.Errorf("error dereferencing status attachment: %s", err)
			continue
		}
		l.Debugf("dereferenced attachment: %+v", deferencedAttachment)
		deferencedAttachment.StatusID = status.ID
		deferencedAttachment.Description = a.Description
		if err := f.db.Put(deferencedAttachment); err != nil {
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
		if err := f.db.GetWhere([]db.Where{{Key: "uri", Value: uri.String()}}, targetAccount); err != nil {
			// proper error
			if _, ok := err.(db.ErrNoEntries); !ok {
				return fmt.Errorf("db error checking for account with uri %s", uri.String())
			}

			// we just don't have it yet, so we should go get it....
			accountable, err := f.DereferenceRemoteAccount(requestingUsername, uri)
			if err != nil {
				// we can't dereference it so just skip it
				l.Debugf("error dereferencing remote account with uri %s: %s", uri.String(), err)
				continue
			}

			targetAccount, err = f.typeConverter.ASRepresentationToAccount(accountable, false)
			if err != nil {
				l.Debugf("error converting remote account with uri %s into gts model: %s", uri.String(), err)
				continue
			}

			targetAccountID, err := id.NewRandomULID()
			if err != nil {
				return err
			}
			targetAccount.ID = targetAccountID

			if err := f.db.Put(targetAccount); err != nil {
				return fmt.Errorf("db error inserting account with uri %s", uri.String())
			}
		}

		// by this point, we know the targetAccount exists in our database with an ID :)
		m.TargetAccountID = targetAccount.ID
		if err := f.db.Put(m); err != nil {
			return fmt.Errorf("error creating mention: %s", err)
		}
		mentions = append(mentions, m.ID)
	}
	status.Mentions = mentions

	return nil
}

func (f *federator) DereferenceAccountFields(account *gtsmodel.Account, requestingUsername string, refresh bool) error {
	l := f.log.WithFields(logrus.Fields{
		"func":               "dereferenceAccountFields",
		"requestingUsername": requestingUsername,
	})

	accountURI, err := url.Parse(account.URI)
	if err != nil {
		return fmt.Errorf("DereferenceAccountFields: couldn't parse account URI %s: %s", account.URI, err)
	}
	if blocked, err := f.blockedDomain(accountURI.Host); blocked || err != nil {
		return fmt.Errorf("DereferenceAccountFields: domain %s is blocked", accountURI.Host)
	}

	t, err := f.GetTransportForUser(requestingUsername)
	if err != nil {
		return fmt.Errorf("error getting transport for user: %s", err)
	}

	// fetch the header and avatar
	if err := f.fetchHeaderAndAviForAccount(account, t, refresh); err != nil {
		// if this doesn't work, just skip it -- we can do it later
		l.Debugf("error fetching header/avi for account: %s", err)
	}

	if err := f.db.UpdateByID(account.ID, account); err != nil {
		return fmt.Errorf("error updating account in database: %s", err)
	}

	return nil
}

func (f *federator) DereferenceAnnounce(announce *gtsmodel.Status, requestingUsername string) error {
	if announce.GTSBoostedStatus == nil || announce.GTSBoostedStatus.URI == "" {
		// we can't do anything unfortunately
		return errors.New("DereferenceAnnounce: no URI to dereference")
	}

	boostedStatusURI, err := url.Parse(announce.GTSBoostedStatus.URI)
	if err != nil {
		return fmt.Errorf("DereferenceAnnounce: couldn't parse boosted status URI %s: %s", announce.GTSBoostedStatus.URI, err)
	}
	if blocked, err := f.blockedDomain(boostedStatusURI.Host); blocked || err != nil {
		return fmt.Errorf("DereferenceAnnounce: domain %s is blocked", boostedStatusURI.Host)
	}

	// dereference statuses in the thread of the boosted status
	if err := f.DereferenceRemoteThread(requestingUsername, boostedStatusURI); err != nil {
		return fmt.Errorf("DereferenceAnnounce: error dereferencing thread of boosted status: %s", err)
	}

	// check if we already have the boosted status in the database
	boostedStatus := &gtsmodel.Status{}
	err = f.db.GetWhere([]db.Where{{Key: "uri", Value: announce.GTSBoostedStatus.URI}}, boostedStatus)
	if err == nil {
		// nice, we already have it so we don't actually need to dereference it from remote
		announce.Content = boostedStatus.Content
		announce.ContentWarning = boostedStatus.ContentWarning
		announce.ActivityStreamsType = boostedStatus.ActivityStreamsType
		announce.Sensitive = boostedStatus.Sensitive
		announce.Language = boostedStatus.Language
		announce.Text = boostedStatus.Text
		announce.BoostOfID = boostedStatus.ID
		announce.BoostOfAccountID = boostedStatus.AccountID
		announce.Visibility = boostedStatus.Visibility
		announce.VisibilityAdvanced = boostedStatus.VisibilityAdvanced
		announce.GTSBoostedStatus = boostedStatus
		return nil
	}

	// we don't have it so we need to dereference it
	statusable, err := f.DereferenceRemoteStatus(requestingUsername, boostedStatusURI)
	if err != nil {
		return fmt.Errorf("dereferenceAnnounce: error dereferencing remote status with id %s: %s", announce.GTSBoostedStatus.URI, err)
	}

	// make sure we have the author account in the db
	attributedToProp := statusable.GetActivityStreamsAttributedTo()
	for iter := attributedToProp.Begin(); iter != attributedToProp.End(); iter = iter.Next() {
		accountURI := iter.GetIRI()
		if accountURI == nil {
			continue
		}

		if err := f.db.GetWhere([]db.Where{{Key: "uri", Value: accountURI.String()}}, &gtsmodel.Account{}); err == nil {
			// we already have it, fine
			continue
		}

		// we don't have the boosted status author account yet so dereference it
		accountable, err := f.DereferenceRemoteAccount(requestingUsername, accountURI)
		if err != nil {
			return fmt.Errorf("dereferenceAnnounce: error dereferencing remote account with id %s: %s", accountURI.String(), err)
		}
		account, err := f.typeConverter.ASRepresentationToAccount(accountable, false)
		if err != nil {
			return fmt.Errorf("dereferenceAnnounce: error converting dereferenced account with id %s into account : %s", accountURI.String(), err)
		}

		accountID, err := id.NewRandomULID()
		if err != nil {
			return err
		}
		account.ID = accountID

		if err := f.db.Put(account); err != nil {
			return fmt.Errorf("dereferenceAnnounce: error putting dereferenced account with id %s into database : %s", accountURI.String(), err)
		}

		if err := f.DereferenceAccountFields(account, requestingUsername, false); err != nil {
			return fmt.Errorf("dereferenceAnnounce: error dereferencing fields on account with id %s : %s", accountURI.String(), err)
		}
	}

	// now convert the statusable into something we can understand
	boostedStatus, err = f.typeConverter.ASStatusToStatus(statusable)
	if err != nil {
		return fmt.Errorf("dereferenceAnnounce: error converting dereferenced statusable with id %s into status : %s", announce.GTSBoostedStatus.URI, err)
	}

	boostedStatusID, err := id.NewULIDFromTime(boostedStatus.CreatedAt)
	if err != nil {
		return nil
	}
	boostedStatus.ID = boostedStatusID

	if err := f.db.Put(boostedStatus); err != nil {
		return fmt.Errorf("dereferenceAnnounce: error putting dereferenced status with id %s into the db: %s", announce.GTSBoostedStatus.URI, err)
	}

	// now dereference additional fields straight away (we're already async here so we have time)
	if err := f.DereferenceStatusFields(boostedStatus, requestingUsername); err != nil {
		return fmt.Errorf("dereferenceAnnounce: error dereferencing status fields for status with id %s: %s", announce.GTSBoostedStatus.URI, err)
	}

	// update with the newly dereferenced fields
	if err := f.db.UpdateByID(boostedStatus.ID, boostedStatus); err != nil {
		return fmt.Errorf("dereferenceAnnounce: error updating dereferenced status in the db: %s", err)
	}

	// we have everything we need!
	announce.Content = boostedStatus.Content
	announce.ContentWarning = boostedStatus.ContentWarning
	announce.ActivityStreamsType = boostedStatus.ActivityStreamsType
	announce.Sensitive = boostedStatus.Sensitive
	announce.Language = boostedStatus.Language
	announce.Text = boostedStatus.Text
	announce.BoostOfID = boostedStatus.ID
	announce.BoostOfAccountID = boostedStatus.AccountID
	announce.Visibility = boostedStatus.Visibility
	announce.VisibilityAdvanced = boostedStatus.VisibilityAdvanced
	announce.GTSBoostedStatus = boostedStatus
	return nil
}

// fetchHeaderAndAviForAccount fetches the header and avatar for a remote account, using a transport
// on behalf of requestingUsername.
//
// targetAccount's AvatarMediaAttachmentID and HeaderMediaAttachmentID will be updated as necessary.
//
// SIDE EFFECTS: remote header and avatar will be stored in local storage, and the database will be updated
// to reflect the creation of these new attachments.
func (f *federator) fetchHeaderAndAviForAccount(targetAccount *gtsmodel.Account, t transport.Transport, refresh bool) error {
	accountURI, err := url.Parse(targetAccount.URI)
	if err != nil {
		return fmt.Errorf("fetchHeaderAndAviForAccount: couldn't parse account URI %s: %s", targetAccount.URI, err)
	}
	if blocked, err := f.blockedDomain(accountURI.Host); blocked || err != nil {
		return fmt.Errorf("fetchHeaderAndAviForAccount: domain %s is blocked", accountURI.Host)
	}

	if targetAccount.AvatarRemoteURL != "" && (targetAccount.AvatarMediaAttachmentID == "" || refresh) {
		a, err := f.mediaHandler.ProcessRemoteHeaderOrAvatar(t, &gtsmodel.MediaAttachment{
			RemoteURL: targetAccount.AvatarRemoteURL,
			Avatar:    true,
		}, targetAccount.ID)
		if err != nil {
			return fmt.Errorf("error processing avatar for user: %s", err)
		}
		targetAccount.AvatarMediaAttachmentID = a.ID
	}

	if targetAccount.HeaderRemoteURL != "" && (targetAccount.HeaderMediaAttachmentID == "" || refresh) {
		a, err := f.mediaHandler.ProcessRemoteHeaderOrAvatar(t, &gtsmodel.MediaAttachment{
			RemoteURL: targetAccount.HeaderRemoteURL,
			Header:    true,
		}, targetAccount.ID)
		if err != nil {
			return fmt.Errorf("error processing header for user: %s", err)
		}
		targetAccount.HeaderMediaAttachmentID = a.ID
	}
	return nil
}

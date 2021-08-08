package dereferencing

import (
	"fmt"
	"net/url"

	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
)

// DereferenceThread takes a statusable (something that has withReplies and withInReplyTo),
// and dereferences statusables in the conversation.
//
// This process involves working up and down the chain of replies, and parsing through the collections of IDs
// presented by remote instances as part of their replies collections, and will likely involve making several calls to
// multiple different hosts.
func (d *deref) DereferenceThread(username string, statusIRI *url.URL) error {
	l := d.log.WithFields(logrus.Fields{
		"func":      "DereferenceThread",
		"username":  username,
		"statusIRI": statusIRI.String(),
	})
	l.Debug("entering DereferenceThread")

	// if it's our status we already have everything stashed so we can bail early
	if statusIRI.Host == d.config.Host {
		l.Debug("iri belongs to us, bailing")
		return nil
	}

	// first make sure we have this status in our db
	_, _, _, err := d.GetRemoteStatus(username, statusIRI, true)
	if err != nil {
		return fmt.Errorf("DereferenceThread: error getting status with id %s: %s", statusIRI.String(), err)
	}



	// first iterate up through ancestors, dereferencing if necessary as we go
	if err := d.iterateAncestors(username, *statusIRI); err != nil {
		return fmt.Errorf("error iterating ancestors of status %s: %s", statusIRI.String(), err)
	}

	// now iterate down through descendants, again dereferencing as we go
	if err := d.iterateDescendants(username, *statusIRI); err != nil {
		return fmt.Errorf("error iterating descendants of status %s: %s", statusIRI.String(), err)
	}

	return nil
}

func (d *deref) iterateAncestors(username string, statusIRI url.URL) error {
	l := d.log.WithFields(logrus.Fields{
		"func":      "iterateAncestors",
		"username":  username,
		"statusIRI": statusIRI.String(),
	})
	l.Debug("entering iterateAncestors")

	// if it's our status we already have ancestors stashed so we can bail early
	if statusIRI.Host == d.config.Host {
		l.Debug("iri belongs to us, bailing")
		return nil
	}

	// fetch the remote representation of the given status
	statusable, err := d.dereferenceStatusable(username, &statusIRI)
	if err != nil {
		return fmt.Errorf("iterateAncestors: error dereferencing status with id %s: %s", statusIRI.String(), err)
	}

	currentIRI := &statusIRI
searchLoop:
	for {
		// begin by checking if we already have this iri in our database
		gtsStatus := &gtsmodel.Status{}
		if err := d.db.GetWhere([]db.Where{{Key: "uri", Value: currentIRI.String()}}, gtsStatus); err == nil {
			// nice, we already have it as a gts status so our life just got easier

			var inReplyToURI string
			if gtsStatus.InReplyToURI != "" {
				// we already have the replyToURI on the current oldest status
				inReplyToURI = gtsStatus.InReplyToURI
			} else if gtsStatus.InReplyToID != "" {
				// we don't have the replyToURI, but we do have a status ID so the status replied to should be in our db already...
				repliedGTSStatus := &gtsmodel.Status{}
				if err := d.db.GetByID(gtsStatus.InReplyToID, repliedGTSStatus); err != nil {
					return fmt.Errorf("database error getting status with id %s: %s", gtsStatus.InReplyToID, err)
				}
				inReplyToURI = repliedGTSStatus.URI
			} else {
				// this status doesn't reply to anything
				break searchLoop
			}

			// set the oldestIRI to the parent we just found, and go to the next iteration
			repliedGTSStatusIRI, err := url.Parse(inReplyToURI)
			if err != nil {
				return fmt.Errorf("error parsing URI %s: %s", inReplyToURI, err)
			}
			currentIRI = repliedGTSStatusIRI
			continue
		}

		// we don't have currentIRI status in our database yet so we'll have to dereference it to see if it replies to something
		statusable, err := d.dereferenceStatusable(username, currentIRI)
		if err != nil {
			l.Debugf("error dereferencing %s: %s", currentIRI.String(), err)
			break searchLoop
		}

		inReplyToURI := ap.ExtractInReplyToURI(statusable)
		if inReplyToURI != nil {
			currentIRI = inReplyToURI
			continue
		}

		// if we reach this point we couldn't find something older than oldestIRI so we're done
		break searchLoop
	}

	return nil
}

func (d *deref) iterateDescendants(username string, statusIRI url.URL) error {
	l := d.log.WithFields(logrus.Fields{
		"func":      "iterateDescendants",
		"username":  username,
		"statusIRI": statusIRI.String(),
	})
	l.Debug("entering iterateDescendants")

	// if it's our status we already have descendants stashed so we can bail early
	if statusIRI.Host == d.config.Host {
		l.Debug("iri belongs to us, bailing")
		return nil
	}

	_, new, err := d.GetRemoteStatus(username, &statusIRI)
	if err != nil {
		return fmt.Errorf("iterateDescendants: error getting status with id %s: %s", statusIRI.String(), err)
	}

	if !new {
		// since we've already seen this status, we don't need to proceed with further iteration
		return nil
	}

	// fetch the remote representation of the given status
	statusable, err := d.dereferenceStatusable(username, &statusIRI)
	if err != nil {
		return fmt.Errorf("iterateDescendants: error dereferencing status with id %s: %s", statusIRI.String(), err)
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
		nextPage, err := d.DereferenceCollectionPage(username, currentPageIRI)
		if err != nil {
			return nil
		}

		nextItems := typeutils.ExtractURLItems(nextPage)
		if len(nextItems) == 0 {
			// no items on this page, which means we're done
			break pageLoop
		}

		for _, i := range nextItems {
			if i.Host == d.config.Host {
				// skip if the reply is from us -- we already have it then
				continue
			}
			foundReplies = foundReplies + 1
			d.iterateDescendants(username, *i)
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

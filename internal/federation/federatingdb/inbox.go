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

package federatingdb

import (
	"context"
	"fmt"
	"net/url"

	"github.com/go-fed/activity/pub"
	"github.com/go-fed/activity/streams"
	"github.com/go-fed/activity/streams/vocab"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// InboxContains returns true if the OrderedCollection at 'inbox'
// contains the specified 'id'.
//
// The library makes this call only after acquiring a lock first.
func (f *federatingDB) InboxContains(c context.Context, inbox, id *url.URL) (contains bool, err error) {
	l := f.log.WithFields(
		logrus.Fields{
			"func": "InboxContains",
			"id":   id.String(),
		},
	)
	l.Debugf("entering INBOXCONTAINS function with for inbox %s and id %s", inbox.String(), id.String())

	if !util.IsInboxPath(inbox) {
		return false, fmt.Errorf("%s is not an inbox URI", inbox.String())
	}

	activityI := c.Value(util.APActivity)
	if activityI == nil {
		return false, fmt.Errorf("no activity was set for id %s", id.String())
	}
	activity, ok := activityI.(pub.Activity)
	if !ok || activity == nil {
		return false, fmt.Errorf("could not parse contextual activity for id %s", id.String())
	}

	l.Debugf("activity type %s for id %s", activity.GetTypeName(), id.String())

	return false, nil
}

// GetInbox returns the first ordered collection page of the outbox at
// the specified IRI, for prepending new items.
//
// The library makes this call only after acquiring a lock first.
func (f *federatingDB) GetInbox(c context.Context, inboxIRI *url.URL) (inbox vocab.ActivityStreamsOrderedCollectionPage, err error) {
	l := f.log.WithFields(
		logrus.Fields{
			"func": "GetInbox",
		},
	)
	l.Debugf("entering GETINBOX function with inboxIRI %s", inboxIRI.String())
	return streams.NewActivityStreamsOrderedCollectionPage(), nil
}

// SetInbox saves the inbox value given from GetInbox, with new items
// prepended. Note that the new items must not be added as independent
// database entries. Separate calls to Create will do that.
//
// The library makes this call only after acquiring a lock first.
func (f *federatingDB) SetInbox(c context.Context, inbox vocab.ActivityStreamsOrderedCollectionPage) error {
	l := f.log.WithFields(
		logrus.Fields{
			"func": "SetInbox",
		},
	)
	l.Debug("entering SETINBOX function")
	return nil
}

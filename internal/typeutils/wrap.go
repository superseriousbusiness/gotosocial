// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package typeutils

import (
	"net/url"

	"github.com/superseriousbusiness/activity/pub"
	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/uris"
)

// WrapPersonInUpdate ...
func (c *Converter) WrapPersonInUpdate(person vocab.ActivityStreamsPerson, originAccount *gtsmodel.Account) (vocab.ActivityStreamsUpdate, error) {
	update := streams.NewActivityStreamsUpdate()

	// set the actor
	actorURI, err := url.Parse(originAccount.URI)
	if err != nil {
		return nil, gtserror.Newf("error parsing url %s: %w", originAccount.URI, err)
	}
	actorProp := streams.NewActivityStreamsActorProperty()
	actorProp.AppendIRI(actorURI)
	update.SetActivityStreamsActor(actorProp)

	// set the ID
	newID, err := id.NewRandomULID()
	if err != nil {
		return nil, err
	}

	idString := uris.GenerateURIForUpdate(originAccount.Username, newID)
	idURI, err := url.Parse(idString)
	if err != nil {
		return nil, gtserror.Newf("error parsing url %s: %w", idString, err)
	}
	idProp := streams.NewJSONLDIdProperty()
	idProp.SetIRI(idURI)
	update.SetJSONLDId(idProp)

	// set the person as the object here
	objectProp := streams.NewActivityStreamsObjectProperty()
	objectProp.AppendActivityStreamsPerson(person)
	update.SetActivityStreamsObject(objectProp)

	// to should be public
	toURI, err := url.Parse(pub.PublicActivityPubIRI)
	if err != nil {
		return nil, gtserror.Newf("error parsing url %s: %w", pub.PublicActivityPubIRI, err)
	}
	toProp := streams.NewActivityStreamsToProperty()
	toProp.AppendIRI(toURI)
	update.SetActivityStreamsTo(toProp)

	// bcc followers
	followersURI, err := url.Parse(originAccount.FollowersURI)
	if err != nil {
		return nil, gtserror.Newf("error parsing url %s: %w", originAccount.FollowersURI, err)
	}
	bccProp := streams.NewActivityStreamsBccProperty()
	bccProp.AppendIRI(followersURI)
	update.SetActivityStreamsBcc(bccProp)

	return update, nil
}

func WrapStatusableInCreate(status ap.Statusable, iriOnly bool) vocab.ActivityStreamsCreate {
	create := streams.NewActivityStreamsCreate()
	wrapStatusableInActivity(create, status, iriOnly)
	return create
}

func WrapStatusableInUpdate(status ap.Statusable, iriOnly bool) vocab.ActivityStreamsUpdate {
	update := streams.NewActivityStreamsUpdate()
	wrapStatusableInActivity(update, status, iriOnly)
	return update
}

// wrapStatusableInActivity adds the required ap.Statusable data to the given ap.Activityable.
func wrapStatusableInActivity(activity ap.Activityable, status ap.Statusable, iriOnly bool) {
	idIRI := ap.GetJSONLDId(status) // activity ID formatted as {$statusIRI}/activity#{$typeName}
	ap.MustSet(ap.SetJSONLDIdStr, ap.WithJSONLDId(activity), idIRI.String()+"/activity#"+activity.GetTypeName())
	appendStatusableToActivity(activity, status, iriOnly)
	ap.AppendTo(activity, ap.GetTo(status)...)
	ap.AppendCc(activity, ap.GetCc(status)...)
	ap.AppendActorIRIs(activity, ap.GetAttributedTo(status)...)
	ap.SetPublished(activity, ap.GetPublished(status))
}

// appendStatusableToActivity appends a Statusable type to an Activityable, handling case of Question, Note or just IRI type.
func appendStatusableToActivity(activity ap.Activityable, status ap.Statusable, iriOnly bool) {
	// Get existing object property or allocate new.
	objProp := activity.GetActivityStreamsObject()
	if objProp == nil {
		objProp = streams.NewActivityStreamsObjectProperty()
		activity.SetActivityStreamsObject(objProp)
	}

	if iriOnly {
		// Only append status IRI.
		idIRI := ap.GetJSONLDId(status)
		objProp.AppendIRI(idIRI)
	} else if poll, ok := ap.ToPollable(status); ok {
		// Our Pollable implementer is an AS Question type.
		question := poll.(vocab.ActivityStreamsQuestion)
		objProp.AppendActivityStreamsQuestion(question)
	} else {
		// All of our other Statusable types are AS Note.
		note := status.(vocab.ActivityStreamsNote)
		objProp.AppendActivityStreamsNote(note)
	}
}

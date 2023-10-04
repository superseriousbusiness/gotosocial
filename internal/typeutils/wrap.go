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

// WrapNoteInCreate wraps a Statusable with a Create activity.
//
// If objectIRIOnly is set to true, then the function won't put the *entire* note in the Object field of the Create,
// but just the AP URI of the note. This is useful in cases where you want to give a remote server something to dereference,
// and still have control over whether or not they're allowed to actually see the contents.
func (c *Converter) WrapStatusableInCreate(status ap.Statusable, iriOnly bool) (ap.Activityable, error) {
	var activity ap.Activityable

	if pollable, ok := ap.ToPollable(status); ok && !iriOnly {
		// If Statusable is actually sub-type Pollable,
		// i.e. an AS Question, it is already any activity
		// that represents a "create" (logically at least).
		activity = pollable
	} else {
		// Else, allocate new AS create activity.
		create := streams.NewActivityStreamsCreate()

		// Create new object property for the activity
		// to hold the statusable itself, or the IRI only.
		objProp := streams.NewActivityStreamsObjectProperty()
		if iriOnly {
			// Fetch the status IRI and append.
			iri := status.GetJSONLDId().GetIRI()
			objProp.AppendIRI(iri)
		} else {
			// Our regular statuses are always AS Note types.
			asNote := status.(vocab.ActivityStreamsNote)
			objProp.AppendActivityStreamsNote(asNote)
		}

		// Set object property on the activity.
		create.SetActivityStreamsObject(objProp)

		// Set the activity.
		activity = create
	}

	// Copy over remaining required properties from status to activity.
	if err := copyFromStatusToActivity(status, activity); err != nil {
		return nil, err
	}

	return activity, nil
}

// WrapStatusableInUpdate wraps a Statusable with an Update activity.
//
// If objectIRIOnly is set to true, then the function won't put the *entire* note in the Object field of the Create,
// but just the AP URI of the note. This is useful in cases where you want to give a remote server something to dereference,
// and still have control over whether or not they're allowed to actually see the contents.
func (c *Converter) WrapStatusableInUpdate(status ap.Statusable, iriOnly bool) (vocab.ActivityStreamsUpdate, error) {
	// Allocate a new AS update activity.
	update := streams.NewActivityStreamsUpdate()

	// Create new object property and set required status form.
	objProp := streams.NewActivityStreamsObjectProperty()
	if iriOnly {
		// Fetch the status IRI and append.
		iri := status.GetJSONLDId().GetIRI()
		objProp.AppendIRI(iri)
	} else if _, ok := status.(ap.Pollable); ok {
		// Our poll statuses are always AS Question types.
		asQuestion := status.(vocab.ActivityStreamsQuestion)
		objProp.AppendActivityStreamsQuestion(asQuestion)
	} else {
		// Our regular statuses are always AS Note types.
		asNote := status.(vocab.ActivityStreamsNote)
		objProp.AppendActivityStreamsNote(asNote)
	}

	// Set the object property on activity.
	update.SetActivityStreamsObject(objProp)

	// Copy over remaining required properties from status to activity.
	if err := copyFromStatusToActivity(status, update); err != nil {
		return nil, err
	}

	return update, nil
}

// copyFromStatusToActivity copies over requried properties in an activity from a statusable.
func copyFromStatusToActivity(statusable ap.Statusable, activityable ap.Activityable) error {
	// Fetch the IRI of this status.
	iri := statusable.GetJSONLDId().GetIRI()

	// Create new IRI (URL) object for this activity.
	idIRI, err := url.Parse(iri.String() + "/activity")
	if err != nil {
		return gtserror.Newf("invalid activity id: %w", err)
	}

	// Wrap in ID property and add to activity.
	idProp := streams.NewJSONLDIdProperty()
	idProp.SetIRI(idIRI)
	activityable.SetJSONLDId(idProp)

	// Fetch attributed-to URI to use as actor IRI from statusable.
	actorIRI, err := ap.ExtractAttributedToURI(statusable)
	if err != nil {
		return gtserror.Newf("couldn't extract attributedTo: %w", err)
	}

	// Wrap in actor property and add to the activity.
	actorProp := streams.NewActivityStreamsActorProperty()
	actorProp.AppendIRI(actorIRI)
	activityable.SetActivityStreamsActor(actorProp)

	// Extract TO URIs from statusable and add to activity.
	toProp := streams.NewActivityStreamsToProperty()
	for _, to := range ap.ExtractToURIs(statusable) {
		toProp.AppendIRI(to)
	}
	activityable.SetActivityStreamsTo(toProp)

	// Extract CC URIs from statusable and add to activity.
	ccProp := streams.NewActivityStreamsCcProperty()
	for _, cc := range ap.ExtractCcURIs(statusable) {
		ccProp.AppendIRI(cc)
	}
	activityable.SetActivityStreamsCc(ccProp)

	return nil
}

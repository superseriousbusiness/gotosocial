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
	"code.superseriousbusiness.org/activity/streams"
	"code.superseriousbusiness.org/activity/streams/vocab"
	"code.superseriousbusiness.org/gotosocial/internal/ap"
	"code.superseriousbusiness.org/gotosocial/internal/id"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"code.superseriousbusiness.org/gotosocial/internal/uris"
)

// WrapAccountableInUpdate wraps the given accountable
// in an Update activity with the accountable as the object.
//
// The Update will be addressed to Public and bcc followers.
func (c *Converter) WrapAccountableInUpdate(accountable ap.Accountable) (vocab.ActivityStreamsUpdate, error) {
	update := streams.NewActivityStreamsUpdate()

	// Set actor IRI to this accountable's IRI.
	ap.AppendActorIRIs(update, ap.GetJSONLDId(accountable))

	// Set the update ID
	updateURI := uris.GenerateURIForUpdate(ap.ExtractPreferredUsername(accountable), id.NewULID())
	ap.MustSet(ap.SetJSONLDIdStr, ap.WithJSONLDId(update), updateURI)

	// Set the accountable as the object of the update.
	objectProp := streams.NewActivityStreamsObjectProperty()
	switch t := accountable.(type) {
	case vocab.ActivityStreamsPerson:
		objectProp.AppendActivityStreamsPerson(t)
	case vocab.ActivityStreamsService:
		objectProp.AppendActivityStreamsService(t)
	default:
		log.Panicf(nil, "%T was neither person nor service", t)
	}
	update.SetActivityStreamsObject(objectProp)

	// to should be public.
	ap.AppendTo(update, ap.PublicURI())

	// bcc should be followers.
	ap.AppendBcc(update, ap.GetFollowers(accountable))

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

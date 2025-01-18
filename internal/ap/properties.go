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

package ap

import (
	"fmt"
	"net/url"
	"time"

	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
)

// MustGet performs the given 'Get$Property(with) (T, error)' signature function, panicking on error.
// func MustGet[W, T any](fn func(W) (T, error), with W) T {
// 	t, err := fn(with)
// 	if err != nil {
// 		panicfAt(3, "error getting property on %T: %w", with, err)
// 	}
// 	return t
// }

// MustSet performs the given 'Set$Property(with, T) error' signature function, panicking on error.
func MustSet[W, T any](fn func(W, T) error, with W, value T) {
	err := fn(with, value)
	if err != nil {
		panicfAt(3, "error setting property on %T: %w", with, err)
	}
}

// AppendSet performs the given 'Append$Property(with, ...T) error' signature function, panicking on error.
// func MustAppend[W, T any](fn func(W, ...T) error, with W, values ...T) {
// 	err := fn(with, values...)
// 	if err != nil {
// 		panicfAt(3, "error appending properties on %T: %w", with, err)
// 	}
// }

// GetJSONLDId returns the ID of 'with', or nil.
func GetJSONLDId(with WithJSONLDId) *url.URL {
	idProp := with.GetJSONLDId()
	if idProp == nil || !idProp.IsXMLSchemaAnyURI() {
		return nil
	}
	return idProp.Get()
}

// SetJSONLDId sets the given URL to the JSONLD ID of 'with'.
func SetJSONLDId(with WithJSONLDId, id *url.URL) {
	idProp := with.GetJSONLDId()
	if idProp == nil {
		idProp = streams.NewJSONLDIdProperty()
		with.SetJSONLDId(idProp)
	}
	idProp.SetIRI(id)
}

// SetJSONLDIdStr sets the given string to the JSONLDID of 'with'. Returns error
func SetJSONLDIdStr(with WithJSONLDId, id string) error {
	u, err := url.Parse(id)
	if err != nil {
		return fmt.Errorf("error parsing id url: %w", err)
	}
	SetJSONLDId(with, u)
	return nil
}

// GetTo returns the IRIs contained in the To property of 'with'. Panics on entries with missing ID.
func GetTo(with WithTo) []*url.URL {
	toProp := with.GetActivityStreamsTo()
	return getIRIs[vocab.ActivityStreamsToPropertyIterator](toProp)
}

// AppendTo appends the given IRIs to the To property of 'with'.
func AppendTo(with WithTo, to ...*url.URL) {
	appendIRIs(func() Property[vocab.ActivityStreamsToPropertyIterator] {
		toProp := with.GetActivityStreamsTo()
		if toProp == nil {
			toProp = streams.NewActivityStreamsToProperty()
			with.SetActivityStreamsTo(toProp)
		}
		return toProp
	}, to...)
}

// GetCc returns the IRIs contained in the Cc property of 'with'. Panics on entries with missing ID.
func GetCc(with WithCc) []*url.URL {
	ccProp := with.GetActivityStreamsCc()
	return extractIRIs[vocab.ActivityStreamsCcPropertyIterator](ccProp)
}

// AppendCc appends the given IRIs to the Cc property of 'with'.
func AppendCc(with WithCc, cc ...*url.URL) {
	appendIRIs(func() Property[vocab.ActivityStreamsCcPropertyIterator] {
		ccProp := with.GetActivityStreamsCc()
		if ccProp == nil {
			ccProp = streams.NewActivityStreamsCcProperty()
			with.SetActivityStreamsCc(ccProp)
		}
		return ccProp
	}, cc...)
}

// GetBcc returns the IRIs contained in the Bcc property of 'with'. Panics on entries with missing ID.
func GetBcc(with WithBcc) []*url.URL {
	bccProp := with.GetActivityStreamsBcc()
	return extractIRIs[vocab.ActivityStreamsBccPropertyIterator](bccProp)
}

// AppendBcc appends the given IRIs to the Bcc property of 'with'.
func AppendBcc(with WithBcc, bcc ...*url.URL) {
	appendIRIs(func() Property[vocab.ActivityStreamsBccPropertyIterator] {
		bccProp := with.GetActivityStreamsBcc()
		if bccProp == nil {
			bccProp = streams.NewActivityStreamsBccProperty()
			with.SetActivityStreamsBcc(bccProp)
		}
		return bccProp
	}, bcc...)
}

// GetURL returns the IRIs contained in the URL property of 'with'.
func GetURL(with WithURL) []*url.URL {
	urlProp := with.GetActivityStreamsUrl()
	if urlProp == nil || urlProp.Len() == 0 {
		return nil
	}
	urls := make([]*url.URL, 0, urlProp.Len())
	for i := 0; i < urlProp.Len(); i++ {
		at := urlProp.At(i)
		if at.IsXMLSchemaAnyURI() {
			u := at.GetXMLSchemaAnyURI()
			urls = append(urls, u)
		}
	}
	return urls
}

// AppendURL appends the given URLs to the URL property of 'with'.
func AppendURL(with WithURL, url ...*url.URL) {
	if len(url) == 0 {
		return
	}
	urlProp := with.GetActivityStreamsUrl()
	if urlProp == nil {
		urlProp = streams.NewActivityStreamsUrlProperty()
		with.SetActivityStreamsUrl(urlProp)
	}
	for _, u := range url {
		urlProp.AppendXMLSchemaAnyURI(u)
	}
}

// GetActorIRIs returns the IRIs contained in the Actor property of 'with'.
func GetActorIRIs(with WithActor) []*url.URL {
	actorProp := with.GetActivityStreamsActor()
	return extractIRIs[vocab.ActivityStreamsActorPropertyIterator](actorProp)
}

// AppendActorIRIs appends the given IRIs to the Actor property of 'with'.
func AppendActorIRIs(with WithActor, actor ...*url.URL) {
	appendIRIs(func() Property[vocab.ActivityStreamsActorPropertyIterator] {
		actorProp := with.GetActivityStreamsActor()
		if actorProp == nil {
			actorProp = streams.NewActivityStreamsActorProperty()
			with.SetActivityStreamsActor(actorProp)
		}
		return actorProp
	}, actor...)
}

// GetObjectIRIs returns the IRIs contained in the Object property of 'with'.
func GetObjectIRIs(with WithObject) []*url.URL {
	objectProp := with.GetActivityStreamsObject()
	return extractIRIs[vocab.ActivityStreamsObjectPropertyIterator](objectProp)
}

// AppendObjectIRIs appends the given IRIs to the Object property of 'with'.
func AppendObjectIRIs(with WithObject, object ...*url.URL) {
	appendIRIs(func() Property[vocab.ActivityStreamsObjectPropertyIterator] {
		objectProp := with.GetActivityStreamsObject()
		if objectProp == nil {
			objectProp = streams.NewActivityStreamsObjectProperty()
			with.SetActivityStreamsObject(objectProp)
		}
		return objectProp
	}, object...)
}

// GetTargetIRIs returns the IRIs contained in the Target property of 'with'.
func GetTargetIRIs(with WithTarget) []*url.URL {
	targetProp := with.GetActivityStreamsTarget()
	return extractIRIs[vocab.ActivityStreamsTargetPropertyIterator](targetProp)
}

// AppendTargetIRIs appends the given IRIs to the Target property of 'with'.
func AppendTargetIRIs(with WithTarget, target ...*url.URL) {
	appendIRIs(func() Property[vocab.ActivityStreamsTargetPropertyIterator] {
		targetProp := with.GetActivityStreamsTarget()
		if targetProp == nil {
			targetProp = streams.NewActivityStreamsTargetProperty()
			with.SetActivityStreamsTarget(targetProp)
		}
		return targetProp
	}, target...)
}

// GetAttributedTo returns the IRIs contained in the AttributedTo property of 'with'.
func GetAttributedTo(with WithAttributedTo) []*url.URL {
	attribProp := with.GetActivityStreamsAttributedTo()
	return extractIRIs[vocab.ActivityStreamsAttributedToPropertyIterator](attribProp)
}

// AppendAttributedTo appends the given IRIs to the AttributedTo property of 'with'.
func AppendAttributedTo(with WithAttributedTo, attribTo ...*url.URL) {
	appendIRIs(func() Property[vocab.ActivityStreamsAttributedToPropertyIterator] {
		attribProp := with.GetActivityStreamsAttributedTo()
		if attribProp == nil {
			attribProp = streams.NewActivityStreamsAttributedToProperty()
			with.SetActivityStreamsAttributedTo(attribProp)
		}
		return attribProp
	}, attribTo...)
}

// GetInReplyTo returns the IRIs contained in the InReplyTo property of 'with'.
func GetInReplyTo(with WithInReplyTo) []*url.URL {
	replyProp := with.GetActivityStreamsInReplyTo()
	return extractIRIs[vocab.ActivityStreamsInReplyToPropertyIterator](replyProp)
}

// AppendInReplyTo appends the given IRIs to the InReplyTo property of 'with'.
func AppendInReplyTo(with WithInReplyTo, replyTo ...*url.URL) {
	appendIRIs(func() Property[vocab.ActivityStreamsInReplyToPropertyIterator] {
		replyProp := with.GetActivityStreamsInReplyTo()
		if replyProp == nil {
			replyProp = streams.NewActivityStreamsInReplyToProperty()
			with.SetActivityStreamsInReplyTo(replyProp)
		}
		return replyProp
	}, replyTo...)
}

// GetInbox returns the IRI contained in the Inbox property of 'with'.
func GetInbox(with WithInbox) *url.URL {
	inboxProp := with.GetActivityStreamsInbox()
	if inboxProp == nil || !inboxProp.IsIRI() {
		return nil
	}
	return inboxProp.GetIRI()
}

// SetInbox sets the given IRI on the Inbox property of 'with'.
func SetInbox(with WithInbox, inbox *url.URL) {
	inboxProp := with.GetActivityStreamsInbox()
	if inboxProp == nil {
		inboxProp = streams.NewActivityStreamsInboxProperty()
		with.SetActivityStreamsInbox(inboxProp)
	}
	inboxProp.SetIRI(inbox)
}

// GetOutbox returns the IRI contained in the Outbox property of 'with'.
func GetOutbox(with WithOutbox) *url.URL {
	outboxProp := with.GetActivityStreamsOutbox()
	if outboxProp == nil || !outboxProp.IsIRI() {
		return nil
	}
	return outboxProp.GetIRI()
}

// SetOutbox sets the given IRI on the Outbox property of 'with'.
func SetOutbox(with WithOutbox, outbox *url.URL) {
	outboxProp := with.GetActivityStreamsOutbox()
	if outboxProp == nil {
		outboxProp = streams.NewActivityStreamsOutboxProperty()
		with.SetActivityStreamsOutbox(outboxProp)
	}
	outboxProp.SetIRI(outbox)
}

// GetFollowers returns the IRI contained in the Following property of 'with'.
func GetFollowing(with WithFollowing) *url.URL {
	followProp := with.GetActivityStreamsFollowing()
	if followProp == nil || !followProp.IsIRI() {
		return nil
	}
	return followProp.GetIRI()
}

// SetFollowers sets the given IRI on the Following property of 'with'.
func SetFollowing(with WithFollowing, following *url.URL) {
	followProp := with.GetActivityStreamsFollowing()
	if followProp == nil {
		followProp = streams.NewActivityStreamsFollowingProperty()
		with.SetActivityStreamsFollowing(followProp)
	}
	followProp.SetIRI(following)
}

// GetFollowers returns the IRI contained in the Followers property of 'with'.
func GetFollowers(with WithFollowers) *url.URL {
	followProp := with.GetActivityStreamsFollowers()
	if followProp == nil || !followProp.IsIRI() {
		return nil
	}
	return followProp.GetIRI()
}

// SetFollowers sets the given IRI on the Followers property of 'with'.
func SetFollowers(with WithFollowers, followers *url.URL) {
	followProp := with.GetActivityStreamsFollowers()
	if followProp == nil {
		followProp = streams.NewActivityStreamsFollowersProperty()
		with.SetActivityStreamsFollowers(followProp)
	}
	followProp.SetIRI(followers)
}

// GetFeatured returns the IRI contained in the Featured property of 'with'.
func GetFeatured(with WithFeatured) *url.URL {
	featuredProp := with.GetTootFeatured()
	if featuredProp == nil || !featuredProp.IsIRI() {
		return nil
	}
	return featuredProp.GetIRI()
}

// SetFeatured sets the given IRI on the Featured property of 'with'.
func SetFeatured(with WithFeatured, featured *url.URL) {
	featuredProp := with.GetTootFeatured()
	if featuredProp == nil {
		featuredProp = streams.NewTootFeaturedProperty()
		with.SetTootFeatured(featuredProp)
	}
	featuredProp.SetIRI(featured)
}

// GetMovedTo returns the IRI contained in the movedTo property of 'with'.
func GetMovedTo(with WithMovedTo) *url.URL {
	movedToProp := with.GetActivityStreamsMovedTo()
	if movedToProp == nil || !movedToProp.IsIRI() {
		return nil
	}
	return movedToProp.GetIRI()
}

// SetMovedTo sets the given IRI on the movedTo property of 'with'.
func SetMovedTo(with WithMovedTo, movedTo *url.URL) {
	movedToProp := with.GetActivityStreamsMovedTo()
	if movedToProp == nil {
		movedToProp = streams.NewActivityStreamsMovedToProperty()
		with.SetActivityStreamsMovedTo(movedToProp)
	}
	movedToProp.SetIRI(movedTo)
}

// GetAlsoKnownAs returns the IRI contained in the alsoKnownAs property of 'with'.
func GetAlsoKnownAs(with WithAlsoKnownAs) []*url.URL {
	alsoKnownAsProp := with.GetActivityStreamsAlsoKnownAs()
	return getIRIs[vocab.ActivityStreamsAlsoKnownAsPropertyIterator](alsoKnownAsProp)
}

// SetAlsoKnownAs sets the given IRIs on the alsoKnownAs property of 'with'.
func SetAlsoKnownAs(with WithAlsoKnownAs, alsoKnownAs []*url.URL) {
	appendIRIs(func() Property[vocab.ActivityStreamsAlsoKnownAsPropertyIterator] {
		alsoKnownAsProp := with.GetActivityStreamsAlsoKnownAs()
		if alsoKnownAsProp == nil {
			alsoKnownAsProp = streams.NewActivityStreamsAlsoKnownAsProperty()
			with.SetActivityStreamsAlsoKnownAs(alsoKnownAsProp)
		}
		return alsoKnownAsProp
	}, alsoKnownAs...)
}

// GetPublished returns the time contained in the Published property of 'with'.
func GetPublished(with WithPublished) time.Time {
	publishProp := with.GetActivityStreamsPublished()
	if publishProp == nil || !publishProp.IsXMLSchemaDateTime() {
		return time.Time{}
	}
	return publishProp.Get()
}

// SetPublished sets the given time on the Published property of 'with'.
func SetPublished(with WithPublished, published time.Time) {
	publishProp := with.GetActivityStreamsPublished()
	if publishProp == nil {
		publishProp = streams.NewActivityStreamsPublishedProperty()
		with.SetActivityStreamsPublished(publishProp)
	}
	publishProp.Set(published)
}

// GetUpdated returns the time contained in the Updated property of 'with'.
func GetUpdated(with WithUpdated) time.Time {
	updateProp := with.GetActivityStreamsUpdated()
	if updateProp == nil || !updateProp.IsXMLSchemaDateTime() {
		return time.Time{}
	}
	return updateProp.Get()
}

// SetUpdated sets the given time on the Updated property of 'with'.
func SetUpdated(with WithUpdated, updated time.Time) {
	updateProp := with.GetActivityStreamsUpdated()
	if updateProp == nil {
		updateProp = streams.NewActivityStreamsUpdatedProperty()
		with.SetActivityStreamsUpdated(updateProp)
	}
	updateProp.Set(updated)
}

// GetEndTime returns the time contained in the EndTime property of 'with'.
func GetEndTime(with WithEndTime) time.Time {
	endTimeProp := with.GetActivityStreamsEndTime()
	if endTimeProp == nil || !endTimeProp.IsXMLSchemaDateTime() {
		return time.Time{}
	}
	return endTimeProp.Get()
}

// SetEndTime sets the given time on the EndTime property of 'with'.
func SetEndTime(with WithEndTime, end time.Time) {
	endTimeProp := with.GetActivityStreamsEndTime()
	if endTimeProp == nil {
		endTimeProp = streams.NewActivityStreamsEndTimeProperty()
		with.SetActivityStreamsEndTime(endTimeProp)
	}
	endTimeProp.Set(end)
}

// GetEndTime returns the times contained in the Closed property of 'with'.
func GetClosed(with WithClosed) []time.Time {
	closedProp := with.GetActivityStreamsClosed()
	if closedProp == nil || closedProp.Len() == 0 {
		return nil
	}
	closed := make([]time.Time, 0, closedProp.Len())
	for i := 0; i < closedProp.Len(); i++ {
		at := closedProp.At(i)
		if at.IsXMLSchemaDateTime() {
			t := at.GetXMLSchemaDateTime()
			closed = append(closed, t)
		}
	}
	return closed
}

// AppendClosed appends the given times to the Closed property of 'with'.
func AppendClosed(with WithClosed, closed ...time.Time) {
	if len(closed) == 0 {
		return
	}
	closedProp := with.GetActivityStreamsClosed()
	if closedProp == nil {
		closedProp = streams.NewActivityStreamsClosedProperty()
		with.SetActivityStreamsClosed(closedProp)
	}
	for _, closed := range closed {
		closedProp.AppendXMLSchemaDateTime(closed)
	}
}

// GetVotersCount returns the integer contained in the VotersCount property of 'with', if found.
func GetVotersCount(with WithVotersCount) int {
	votersProp := with.GetTootVotersCount()
	if votersProp == nil || !votersProp.IsXMLSchemaNonNegativeInteger() {
		return 0
	}
	return votersProp.Get()
}

// SetVotersCount sets the given count on the VotersCount property of 'with'.
func SetVotersCount(with WithVotersCount, count int) {
	votersProp := with.GetTootVotersCount()
	if votersProp == nil {
		votersProp = streams.NewTootVotersCountProperty()
		with.SetTootVotersCount(votersProp)
	}
	votersProp.Set(count)
}

// GetDiscoverable returns the boolean contained in the Discoverable property of 'with'.
//
// Returns default 'false' if property unusable or not set.
func GetDiscoverable(with WithDiscoverable) bool {
	discoverProp := with.GetTootDiscoverable()
	if discoverProp == nil || !discoverProp.IsXMLSchemaBoolean() {
		return false
	}
	return discoverProp.Get()
}

// SetDiscoverable sets the given boolean on the Discoverable property of 'with'.
func SetDiscoverable(with WithDiscoverable, discoverable bool) {
	discoverProp := with.GetTootDiscoverable()
	if discoverProp == nil {
		discoverProp = streams.NewTootDiscoverableProperty()
		with.SetTootDiscoverable(discoverProp)
	}
	discoverProp.Set(discoverable)
}

// GetManuallyApprovesFollowers returns the boolean contained in the ManuallyApprovesFollowers property of 'with'.
//
// Returns default 'true' if property unusable or not set.
func GetManuallyApprovesFollowers(with WithManuallyApprovesFollowers) bool {
	mafProp := with.GetActivityStreamsManuallyApprovesFollowers()
	if mafProp == nil || !mafProp.IsXMLSchemaBoolean() {
		return true
	}
	return mafProp.Get()
}

// SetManuallyApprovesFollowers sets the given boolean on the ManuallyApprovesFollowers property of 'with'.
func SetManuallyApprovesFollowers(with WithManuallyApprovesFollowers, manuallyApprovesFollowers bool) {
	mafProp := with.GetActivityStreamsManuallyApprovesFollowers()
	if mafProp == nil {
		mafProp = streams.NewActivityStreamsManuallyApprovesFollowersProperty()
		with.SetActivityStreamsManuallyApprovesFollowers(mafProp)
	}
	mafProp.Set(manuallyApprovesFollowers)
}

// GetApprovedBy returns the URL contained in
// the ApprovedBy property of 'with', if set.
func GetApprovedBy(with WithApprovedBy) *url.URL {
	mafProp := with.GetGoToSocialApprovedBy()
	if mafProp == nil || !mafProp.IsIRI() {
		return nil
	}
	return mafProp.Get()
}

// SetApprovedBy sets the given url
// on the ApprovedBy property of 'with'.
func SetApprovedBy(with WithApprovedBy, approvedBy *url.URL) {
	abProp := with.GetGoToSocialApprovedBy()
	if abProp == nil {
		abProp = streams.NewGoToSocialApprovedByProperty()
		with.SetGoToSocialApprovedBy(abProp)
	}
	abProp.Set(approvedBy)
}

// extractIRIs extracts just the AP IRIs from an iterable
// property that may contain types (with IRIs) or just IRIs.
//
// If you know the property contains only IRIs and no types,
// then use getIRIs instead, since it's slightly faster.
func extractIRIs[T TypeOrIRI](prop Property[T]) []*url.URL {
	if prop == nil || prop.Len() == 0 {
		return nil
	}
	ids := make([]*url.URL, 0, prop.Len())
	for i := 0; i < prop.Len(); i++ {
		at := prop.At(i)
		if t := at.GetType(); t != nil {
			id := GetJSONLDId(t)
			if id != nil {
				ids = append(ids, id)
				continue
			}
		}
		if at.IsIRI() {
			id := at.GetIRI()
			if id != nil {
				ids = append(ids, id)
				continue
			}
		}
	}
	return ids
}

// getIRIs gets AP IRIs from an iterable property of IRIs.
//
// Types will be ignored; to extract IRIs from an iterable
// that may contain types too, use extractIRIs.
func getIRIs[T WithIRI](prop Property[T]) []*url.URL {
	if prop == nil || prop.Len() == 0 {
		return nil
	}
	ids := make([]*url.URL, 0, prop.Len())
	for i := 0; i < prop.Len(); i++ {
		at := prop.At(i)
		if at.IsIRI() {
			id := at.GetIRI()
			if id != nil {
				ids = append(ids, id)
				continue
			}
		}
	}
	return ids
}

func appendIRIs[T WithIRI](getProp func() Property[T], iri ...*url.URL) {
	if len(iri) == 0 {
		return
	}
	prop := getProp()
	if prop == nil {
		// check outside loop.
		panic("prop not set")
	}
	for _, iri := range iri {
		prop.AppendIRI(iri)
	}
}

// panicfAt panics with a call to gtserror.NewfAt() with given args (+1 to calldepth).
func panicfAt(calldepth int, msg string, args ...any) {
	panic(gtserror.NewfAt(calldepth+1, msg, args...))
}

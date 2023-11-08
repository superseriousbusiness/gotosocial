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
	if idProp == nil {
		return nil
	}
	id := idProp.Get()
	if id == nil {
		return nil
	}
	return id
}

// SetJSONLDId sets the given URL to the JSONLD ID of 'with'.
func SetJSONLDId(with WithJSONLDId, id *url.URL) {
	idProp := with.GetJSONLDId()
	if idProp == nil {
		idProp = streams.NewJSONLDIdProperty()
	}
	idProp.SetIRI(id)
	with.SetJSONLDId(idProp)
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
	return getIRIs[vocab.ActivityStreamsCcPropertyIterator](ccProp)
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
	return getIRIs[vocab.ActivityStreamsBccPropertyIterator](bccProp)
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

// GetActor returns the IRIs contained in the Actor property of 'with'. Panics on entries with missing ID.
func GetActor(with WithActor) []*url.URL {
	actorProp := with.GetActivityStreamsActor()
	return getIRIs[vocab.ActivityStreamsActorPropertyIterator](actorProp)
}

// AppendActor appends the given IRIs to the Actor property of 'with'.
func AppendActor(with WithActor, actor ...*url.URL) {
	appendIRIs(func() Property[vocab.ActivityStreamsActorPropertyIterator] {
		actorProp := with.GetActivityStreamsActor()
		if actorProp == nil {
			actorProp = streams.NewActivityStreamsActorProperty()
			with.SetActivityStreamsActor(actorProp)
		}
		return actorProp
	}, actor...)
}

// GetAttributedTo returns the IRIs contained in the AttributedTo property of 'with'. Panics on entries with missing ID.
func GetAttributedTo(with WithAttributedTo) []*url.URL {
	attribProp := with.GetActivityStreamsAttributedTo()
	return getIRIs[vocab.ActivityStreamsAttributedToPropertyIterator](attribProp)
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

// GetInReplyTo returns the IRIs contained in the InReplyTo property of 'with'. Panics on entries with missing ID.
func GetInReplyTo(with WithInReplyTo) []*url.URL {
	replyProp := with.GetActivityStreamsInReplyTo()
	return getIRIs[vocab.ActivityStreamsInReplyToPropertyIterator](replyProp)
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

// GetPublished returns the time contained in the Published property of 'with'.
func GetPublished(with WithPublished) time.Time {
	publishProp := with.GetActivityStreamsPublished()
	if publishProp == nil {
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

// GetEndTime returns the time contained in the EndTime property of 'with'.
func GetEndTime(with WithEndTime) time.Time {
	endTimeProp := with.GetActivityStreamsEndTime()
	if endTimeProp == nil {
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
		if t := at.GetXMLSchemaDateTime(); !t.IsZero() {
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
	if votersProp == nil {
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

func getIRIs[T TypeOrIRI](prop Property[T]) []*url.URL {
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

func appendIRIs[T TypeOrIRI](getProp func() Property[T], iri ...*url.URL) {
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

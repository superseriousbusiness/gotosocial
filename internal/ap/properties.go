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
	"net/url"

	"github.com/superseriousbusiness/activity/pub"
	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
)

// GetJSONLDId returns the ID of 'with'. It will ALWAYS be non-nil, or panics.
func GetJSONLDId(with WithJSONLDId) *url.URL {
	idProp := with.GetJSONLDId()
	if idProp == nil {
		panicfAt(3, "%T contains no JSONLD ID property", with)
	}
	id := idProp.Get()
	if id == nil {
		panicfAt(3, "%T ID property contains no url", with)
	}
	return id
}

// SetJSONLDId sets the given string to the JSONLD ID of 'with'. Panics on failed URL parse.
func SetJSONLDId(with WithJSONLDId, id string) {
	idProp := streams.NewJSONLDIdProperty()
	idProp.SetIRI(mustParseURL(id))
	with.SetJSONLDId(idProp)
}

// GetTo returns the IRIs contained in the To property of 'with'. Panics on entries with missing ID.
func GetTo(with WithTo) []*url.URL {
	toProp := with.GetActivityStreamsTo()
	return GetIRIs(toProp)
}

// AppendTo appends the given IRIs to the To property of 'with'.
func AppendTo(with WithTo, to ...*url.URL) {
	AppendIRIs(func() Property[vocab.ActivityStreamsToPropertyIterator] {
		toProp := with.GetActivityStreamsTo()
		if toProp == nil {
			toProp = streams.NewActivityStreamsToProperty()
		}
		return toProp
	}, to...)
}

// GetCc returns the IRIs contained in the Cc property of 'with'. Panics on entries with missing ID.
func GetCc(with WithCc) []*url.URL {
	ccProp := with.GetActivityStreamsCc()
	return GetIRIs(ccProp)
}

// AppendCc appends the given IRIs to the Cc property of 'with'.
func AppendCc(with WithCc, cc ...*url.URL) {
	AppendIRIs(func() Property[vocab.ActivityStreamsCcPropertyIterator] {
		ccProp := with.GetActivityStreamsCc()
		if ccProp == nil {
			ccProp = streams.NewActivityStreamsCcProperty()
		}
		return ccProp
	}, cc...)
}

// GetBcc returns the IRIs contained in the Bcc property of 'with'. Panics on entries with missing ID.
func GetBcc(with WithBcc) []*url.URL {
	bccProp := with.GetActivityStreamsBcc()
	return GetIRIs(bccProp)
}

// AppendBcc appends the given IRIs to the Bcc property of 'with'.
func AppendBcc(with WithBcc, bcc ...*url.URL) {
	AppendIRIs(func() Property[vocab.ActivityStreamsBccPropertyIterator] {
		bccProp := with.GetActivityStreamsBcc()
		if bccProp == nil {
			bccProp = streams.NewActivityStreamsBccProperty()
		}
		return bccProp
	}, bcc...)
}

// GetActor returns the IRIs contained in the Actor property of 'with'. Panics on entries with missing ID.
func GetActor(with WithActor) []*url.URL {
	actorProp := with.GetActivityStreamsActor()
	return GetIRIs(actorProp)
}

// AppendActor appends the given IRIs to the Actor property of 'with'.
func AppendActor(with WithActor, actor ...*url.URL) {
	AppendIRIs(func() Property[vocab.ActivityStreamsActorPropertyIterator] {
		actorProp := with.GetActivityStreamsActor()
		if actorProp == nil {
			actorProp = streams.NewActivityStreamsActorProperty()
		}
		return actorProp
	}, actor...)
}

// GetAttributedTo returns the IRIs contained in the AttributedTo property of 'with'. Panics on entries with missing ID.
func GetAttributedTo(with WithAttributedTo) []*url.URL {
	attribProp := with.GetActivityStreamsAttributedTo()
	return GetIRIs(attribProp)
}

// AppendAttributedTo appends the given IRIs to the AttributedTo property of 'with'.
func AppendAttributedTo(with WithAttributedTo, attribTo ...*url.URL) {
	AppendIRIs(func() Property[vocab.ActivityStreamsAttributedToPropertyIterator] {
		attribProp := with.GetActivityStreamsAttributedTo()
		if attribProp == nil {
			attribProp = streams.NewActivityStreamsAttributedToProperty()
		}
		return attribProp
	}, attribTo...)
}

// GetInReplyTo returns the IRIs contained in the InReplyTo property of 'with'. Panics on entries with missing ID.
func GetInReplyTo(with WithInReplyTo) []*url.URL {
	replyProp := with.GetActivityStreamsInReplyTo()
	return GetIRIs(replyProp)
}

// AppendInReplyTo appends the given IRIs to the InReplyTo property of 'with'.
func AppendInReplyTo(with WithInReplyTo, replyTo ...*url.URL) {
	AppendIRIs(func() Property[vocab.ActivityStreamsInReplyToPropertyIterator] {
		replyProp := with.GetActivityStreamsInReplyTo()
		if replyProp == nil {
			replyProp = streams.NewActivityStreamsInReplyToProperty()
		}
		return replyProp
	}, replyTo...)
}

type Property[T TypeOrIRI] interface {
	Len() int
	At(int) T

	AppendIRI(*url.URL)
	SetIRI(int, *url.URL)
}

func GetIRIs[T TypeOrIRI](prop Property[T]) []*url.URL {
	if prop == nil || prop.Len() == 0 {
		return nil
	}
	ids := make([]*url.URL, prop.Len())
	for i := 0; i < prop.Len(); i++ {
		at := prop.At(i)
		ids[i] = mustToID(at)
	}
	return ids
}

func AppendIRIs[T TypeOrIRI](getProp func() Property[T], iri ...*url.URL) {
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

// mustParseURL calls url.Parse() and panics on error.
func mustParseURL(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panicfAt(3, "error parsing url %s: %w", s, err)
	}
	return u
}

// mustToID calls pub.ToId() and panics on error.
func mustToID(i pub.IdProperty) *url.URL {
	id, err := pub.ToId(i)
	if err != nil {
		panicfAt(3, "error getting id of %T: %w", i, err)
	}
	return id
}

// mustGetID calls pub.GetId() and panics on error.
func mustGetID(t vocab.Type) *url.URL {
	id, err := pub.GetId(t)
	if err != nil {
		panicfAt(3, "error getting id of %T: %w", t, err)
	}
	return id
}

// panicfAt panics with a call to gtserror.NewfAt() with given args (+1 to calldepth).
func panicfAt(calldepth int, msg string, args ...any) {
	panic(gtserror.NewfAt(calldepth+1, msg, args...))
}

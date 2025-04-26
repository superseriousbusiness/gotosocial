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

package ap_test

import (
	"net/url"
	"slices"
	"testing"

	"code.superseriousbusiness.org/activity/pub"
	"code.superseriousbusiness.org/activity/streams"
	"code.superseriousbusiness.org/activity/streams/vocab"
	"code.superseriousbusiness.org/gotosocial/internal/ap"
)

var testIteratorIRIs = [][]string{
	{
		"https://google.com",
		"https://mastodon.social",
		"http://naughty.naughty.website/heres/the/porn",
		"https://god.monarchies.suck?yes=they&really=do",
	},
	{
		// zero length
	},
	{
		"https://superseriousbusiness.org",
		"http://gotosocial.tv/@slothsgonewild",
	},
}

func TestToCollectionIterator(t *testing.T) {
	for _, iris := range testIteratorIRIs {
		testToCollectionIterator(t, toCollection(iris), "", iris)
		testToCollectionIterator(t, toOrderedCollection(iris), "", iris)
	}
	testToCollectionIterator(t, streams.NewActivityStreamsAdd(), "*typeadd.ActivityStreamsAdd(Add) was not Collection-like", nil)
	testToCollectionIterator(t, streams.NewActivityStreamsBlock(), "*typeblock.ActivityStreamsBlock(Block) was not Collection-like", nil)
}

func TestToCollectionPageIterator(t *testing.T) {
	for _, iris := range testIteratorIRIs {
		testToCollectionPageIterator(t, toCollectionPage(iris), "", iris)
		testToCollectionPageIterator(t, toOrderedCollectionPage(iris), "", iris)
	}
	testToCollectionPageIterator(t, streams.NewActivityStreamsAdd(), "*typeadd.ActivityStreamsAdd(Add) was not CollectionPage-like", nil)
	testToCollectionPageIterator(t, streams.NewActivityStreamsBlock(), "*typeblock.ActivityStreamsBlock(Block) was not CollectionPage-like", nil)
}

func testToCollectionIterator(t *testing.T, in vocab.Type, expectErr string, expectIRIs []string) {
	collect, err := ap.ToCollectionIterator(in)
	if !errCheck(err, expectErr) {
		t.Fatalf("did not return expected error: expect=%v receive=%v", expectErr, err)
	}
	iris := gatherFromIterator(collect)
	if !slices.Equal(iris, expectIRIs) {
		t.Fatalf("did not return expected iris: expect=%v receive=%v", expectIRIs, iris)
	}
}

func testToCollectionPageIterator(t *testing.T, in vocab.Type, expectErr string, expectIRIs []string) {
	page, err := ap.ToCollectionPageIterator(in)
	if !errCheck(err, expectErr) {
		t.Fatalf("did not return expected error: expect=%v receive=%v", expectErr, err)
	}
	iris := gatherFromIterator(page)
	if !slices.Equal(iris, expectIRIs) {
		t.Fatalf("did not return expected iris: expect=%v receive=%v", expectIRIs, iris)
	}
}

func toCollection(iris []string) vocab.ActivityStreamsCollection {
	collect := streams.NewActivityStreamsCollection()
	collect.SetActivityStreamsItems(toItems(iris))
	return collect
}

func toOrderedCollection(iris []string) vocab.ActivityStreamsOrderedCollection {
	collect := streams.NewActivityStreamsOrderedCollection()
	collect.SetActivityStreamsOrderedItems(toOrderedItems(iris))
	return collect
}

func toCollectionPage(iris []string) vocab.ActivityStreamsCollectionPage {
	page := streams.NewActivityStreamsCollectionPage()
	page.SetActivityStreamsItems(toItems(iris))
	return page
}

func toOrderedCollectionPage(iris []string) vocab.ActivityStreamsOrderedCollectionPage {
	page := streams.NewActivityStreamsOrderedCollectionPage()
	page.SetActivityStreamsOrderedItems(toOrderedItems(iris))
	return page
}

func toItems(iris []string) vocab.ActivityStreamsItemsProperty {
	items := streams.NewActivityStreamsItemsProperty()
	for _, iri := range iris {
		u, _ := url.Parse(iri)
		items.AppendIRI(u)
	}
	return items
}

func toOrderedItems(iris []string) vocab.ActivityStreamsOrderedItemsProperty {
	items := streams.NewActivityStreamsOrderedItemsProperty()
	for _, iri := range iris {
		u, _ := url.Parse(iri)
		items.AppendIRI(u)
	}
	return items
}

func gatherFromIterator(iter ap.CollectionIterator) []string {
	var iris []string
	if iter == nil {
		return nil
	}
	for item := iter.NextItem(); item != nil; item = iter.NextItem() {
		id, _ := pub.ToId(item)
		if id != nil {
			iris = append(iris, id.String())
		}
	}
	return iris
}

func errCheck(err error, str string) bool {
	if err == nil {
		return str == ""
	}
	return err.Error() == str
}

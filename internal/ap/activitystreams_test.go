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
	"encoding/json"
	"net/url"
	"testing"

	"codeberg.org/superseriousbusiness/activity/streams/vocab"
	"github.com/stretchr/testify/assert"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/paging"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

func TestASCollection(t *testing.T) {
	const (
		proto = "https"
		host  = "zorg.flabormagorg.xyz"
		path  = "/users/itsa_me_mario"

		idURI = proto + "://" + host + path
		total = 10
	)

	// Create JSON string of expected output.
	expect := toJSON(map[string]any{
		"@context":   "https://www.w3.org/ns/activitystreams",
		"type":       "Collection",
		"id":         idURI,
		"first":      idURI + "?limit=40",
		"totalItems": total,
	})

	// Create new collection using builder function.
	c := ap.NewASCollection(ap.CollectionParams{
		ID:    parseURI(idURI),
		First: new(paging.Page),
		Query: url.Values{"limit": []string{"40"}},
		Total: util.Ptr(total),
	})

	// Serialize collection.
	s := toJSON(c)

	// Ensure outputs are equal.
	assert.Equal(t, expect, s)
}

func TestASCollectionTotalOnly(t *testing.T) {
	const (
		proto = "https"
		host  = "zorg.flabormagorg.xyz"
		path  = "/users/itsa_me_mario"

		idURI = proto + "://" + host + path
		total = 10
	)

	// Create JSON string of expected output.
	expect := toJSON(map[string]any{
		"@context":   "https://www.w3.org/ns/activitystreams",
		"type":       "Collection",
		"id":         idURI,
		"totalItems": total,
	})

	// Create new collection using builder function.
	c := ap.NewASCollection(ap.CollectionParams{
		ID:    parseURI(idURI),
		Total: util.Ptr(total),
	})

	// Serialize collection.
	s := toJSON(c)

	// Ensure outputs are equal.
	assert.Equal(t, expect, s)
}

func TestASCollectionPage(t *testing.T) {
	const (
		proto = "https"
		host  = "zorg.flabormagorg.xyz"
		path  = "/users/itsa_me_mario"

		idURI = proto + "://" + host + path
		total = 10

		minID = "minimum"
		maxID = "maximum"
		limit = 40
		count = 2
	)

	// Create the current page.
	currPg := &paging.Page{
		Limit: 40,
	}

	// Create JSON string of expected output.
	expect := toJSON(map[string]any{
		"@context":   "https://www.w3.org/ns/activitystreams",
		"type":       "CollectionPage",
		"id":         currPg.ToLink(proto, host, path, nil),
		"partOf":     idURI,
		"next":       currPg.Next(minID, maxID).ToLink(proto, host, path, nil),
		"prev":       currPg.Prev(minID, maxID).ToLink(proto, host, path, nil),
		"items":      []interface{}{},
		"totalItems": total,
	})

	// Create new collection page using builder function.
	p := ap.NewASCollectionPage(ap.CollectionPageParams{
		CollectionParams: ap.CollectionParams{
			ID:    parseURI(idURI),
			Total: util.Ptr(total),
		},

		Current: currPg,
		Next:    currPg.Next(minID, maxID),
		Prev:    currPg.Prev(minID, maxID),

		Append: func(i int, ipb ap.ItemsPropertyBuilder) {},
		Count:  count,
	})

	// Serialize page.
	s := toJSON(p)

	// Ensure outputs are equal.
	assert.Equal(t, expect, s)
}

func TestASOrderedCollection(t *testing.T) {
	const (
		idURI = "https://zorg.flabormagorg.xyz/users/itsa_me_mario"
		total = 10
	)

	// Create JSON string of expected output.
	expect := toJSON(map[string]any{
		"@context":   "https://www.w3.org/ns/activitystreams",
		"type":       "OrderedCollection",
		"id":         idURI,
		"first":      idURI + "?limit=40",
		"totalItems": total,
	})

	// Create new collection using builder function.
	c := ap.NewASOrderedCollection(ap.CollectionParams{
		ID:    parseURI(idURI),
		First: new(paging.Page),
		Query: url.Values{"limit": []string{"40"}},
		Total: util.Ptr(total),
	})

	// Serialize collection.
	s := toJSON(c)

	// Ensure outputs are equal.
	assert.Equal(t, expect, s)
}

func TestASOrderedCollectionTotalOnly(t *testing.T) {
	const (
		idURI = "https://zorg.flabormagorg.xyz/users/itsa_me_mario"
		total = 10
	)

	// Create JSON string of expected output.
	expect := toJSON(map[string]any{
		"@context":   "https://www.w3.org/ns/activitystreams",
		"type":       "OrderedCollection",
		"id":         idURI,
		"totalItems": total,
	})

	// Create new collection using builder function.
	c := ap.NewASOrderedCollection(ap.CollectionParams{
		ID:    parseURI(idURI),
		Total: util.Ptr(total),
	})

	// Serialize collection.
	s := toJSON(c)

	// Ensure outputs are equal.
	assert.Equal(t, expect, s)
}

func TestASOrderedCollectionNoTotal(t *testing.T) {
	const (
		idURI = "https://zorg.flabormagorg.xyz/users/itsa_me_mario"
	)

	// Create JSON string of expected output.
	expect := toJSON(map[string]any{
		"@context": "https://www.w3.org/ns/activitystreams",
		"type":     "OrderedCollection",
		"id":       idURI,
	})

	// Create new collection using builder function.
	c := ap.NewASOrderedCollection(ap.CollectionParams{
		ID: parseURI(idURI),
	})

	// Serialize collection.
	s := toJSON(c)

	// Ensure outputs are equal.
	assert.Equal(t, expect, s)
}

func TestASOrderedCollectionPage(t *testing.T) {
	const (
		proto = "https"
		host  = "zorg.flabormagorg.xyz"
		path  = "/users/itsa_me_mario"

		idURI = proto + "://" + host + path
		total = 10

		minID = "minimum"
		maxID = "maximum"
		limit = 40
		count = 2
	)

	// Create the current page.
	currPg := &paging.Page{
		Limit: 40,
	}

	// Create JSON string of expected output.
	expect := toJSON(map[string]any{
		"@context":     "https://www.w3.org/ns/activitystreams",
		"type":         "OrderedCollectionPage",
		"id":           currPg.ToLink(proto, host, path, nil),
		"partOf":       idURI,
		"next":         currPg.Next(minID, maxID).ToLink(proto, host, path, nil),
		"prev":         currPg.Prev(minID, maxID).ToLink(proto, host, path, nil),
		"orderedItems": []interface{}{},
		"totalItems":   total,
	})

	// Create new collection page using builder function.
	p := ap.NewASOrderedCollectionPage(ap.CollectionPageParams{
		CollectionParams: ap.CollectionParams{
			ID:    parseURI(idURI),
			Total: util.Ptr(total),
		},

		Current: currPg,
		Next:    currPg.Next(minID, maxID),
		Prev:    currPg.Prev(minID, maxID),

		Append: func(i int, ipb ap.ItemsPropertyBuilder) {},
		Count:  count,
	})

	// Serialize page.
	s := toJSON(p)

	// Ensure outputs are equal.
	assert.Equal(t, expect, s)
}

func parseURI(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	return u
}

// toJSON will return indented JSON serialized form of 'a'.
func toJSON(a any) string {
	v, ok := a.(vocab.Type)
	if ok {
		m, err := ap.Serialize(v)
		if err != nil {
			panic(err)
		}
		a = m
	}
	b, err := json.MarshalIndent(a, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(b)
}

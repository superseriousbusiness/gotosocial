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
	"bytes"
	"context"
	"io"
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/ap"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"github.com/stretchr/testify/suite"
)

type ResolveTestSuite struct {
	APTestSuite
}

func (suite *ResolveTestSuite) TestResolveDocumentAsStatusable() {
	b := []byte(suite.typeToJson(suite.document1))

	statusable, err := ap.ResolveStatusable(
		context.Background(), io.NopCloser(bytes.NewReader(b)),
	)
	suite.NoError(err)
	suite.NotNil(statusable)
}

func (suite *ResolveTestSuite) TestResolveDocumentAsAccountable() {
	b := []byte(suite.typeToJson(suite.document1))

	accountable, err := ap.ResolveAccountable(
		context.Background(), io.NopCloser(bytes.NewReader(b)),
	)
	suite.True(gtserror.IsWrongType(err))
	suite.EqualError(err, "ResolveAccountable: cannot resolve vocab type *typedocument.ActivityStreamsDocument as accountable")
	suite.Nil(accountable)
}

func (suite *ResolveTestSuite) TestResolveHTMLAsAccountable() {
	b := []byte(`<!DOCTYPE html>
	<title>.</title>`)

	accountable, err := ap.ResolveAccountable(
		context.Background(), io.NopCloser(bytes.NewReader(b)),
	)
	suite.True(gtserror.IsWrongType(err))
	suite.EqualError(err, "ResolveAccountable: error decoding into json: invalid character '<' looking for beginning of value")
	suite.Nil(accountable)
}

func (suite *ResolveTestSuite) TestResolveNonAPJSONAsAccountable() {
	b := []byte(`{
  "@context": "definitely a legit context muy lord",
  "type": "definitely an account muy lord",
  "pee pee":"poo poo"
}`)

	accountable, err := ap.ResolveAccountable(
		context.Background(), io.NopCloser(bytes.NewReader(b)),
	)
	suite.True(gtserror.IsWrongType(err))
	suite.EqualError(err, "ResolveAccountable: error resolving json into ap vocab type: activity stream did not match any known types")
	suite.Nil(accountable)
}

func (suite *ResolveTestSuite) TestResolveBandwagonAlbumAsStatusable() {
	b := []byte(`{
  "@context": [
    "https://www.w3.org/ns/activitystreams",
    "https://w3id.org/security/v1",
    {
      "discoverable": "toot:discoverable",
      "indexable": "toot:indexable",
      "toot": "https://joinmastodon.org/ns#"
    },
    "https://funkwhale.audio/ns"
  ],
  "artists": [
    {
      "id": "https://bandwagon.fm/@67a0a0808121f77ed3466870",
      "name": "Luka Prinčič",
      "type": "Artist"
    }
  ],
  "attachment": [
    {
      "mediaType": "image/webp",
      "name": "image",
      "type": "Document",
      "url": "https://bandwagon.fm/67a0a219f050061c8b4ce427/attachments/67a0a21bf050061c8b4ce429"
    }
  ],
  "attributedTo": "https://bandwagon.fm/@67a0a0808121f77ed3466870",
  "content": "... a transgenre mutation, a fluid entity, jagged pop, electro-funk, techno-cabaret, a schlager, and soft alternative, queer to the core, satire and tragedy, sharp and fun indulgence for the dance of bodies and brains, activism and hedonism, which would all like to steal your attention.\r\n\r\nDRAGX̶FUNK is pronounced /dɹæɡɑːfʌŋk/.\r\n\r\n---\r\n\r\n## Buy digital\r\n💳 [Stripe](https://buy.stripe.com/6oE8x52iG1Kq5pKeV3)\r\n\r\n---\r\n\r\n## Buy dl/merch\r\n🎵 [Bandcamp](https://lukaprincic.bandcamp.com/album/dragx-funk)  \r\n\r\n---\r\n\r\n## More:\r\n🌐 [prin.lu](https://prin.lu/music/241205_dragx-funk/)  \r\n👉 [kamizdat.si](https://kamizdat.si/releases/dragx-funk-2/)\r\n",
  "context": "https://bandwagon.fm/67a0a219f050061c8b4ce427",
  "id": "https://bandwagon.fm/67a0a219f050061c8b4ce427",
  "library": "https://bandwagon.fm/67a0a219f050061c8b4ce427/pub/children",
  "license": "CC-BY-NC-SA",
  "name": "DRAGX̶FUNK",
  "published": "2025-03-17T11:40:53Z",
  "to": [
    "https://www.w3.org/ns/activitystreams#Public"
  ],
  "tracks": "https://bandwagon.fm/67a0a219f050061c8b4ce427/pub/children",
  "type": "Album",
  "url": "https://bandwagon.fm/67a0a219f050061c8b4ce427"
}`)

	statusable, err := ap.ResolveStatusable(
		context.Background(), io.NopCloser(bytes.NewReader(b)),
	)
	suite.NoError(err)
	suite.NotNil(statusable)
}

func TestResolveTestSuite(t *testing.T) {
	suite.Run(t, &ResolveTestSuite{})
}

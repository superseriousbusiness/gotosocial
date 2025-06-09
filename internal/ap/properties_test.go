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
	"io"
	"net/url"
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/ap"
	"code.superseriousbusiness.org/gotosocial/testrig"
	"github.com/stretchr/testify/suite"
)

type PropertiesTestSuite struct {
	suite.Suite
}

func (suite *PropertiesTestSuite) TestGetStatusableURL() {
	// Pretty good representation of
	// how a peertube video is federated.
	const peertubeVideo = `{
  "@context": [
    "https://www.w3.org/ns/activitystreams"
  ],
  "to": [
    "https://www.w3.org/ns/activitystreams#Public"
  ],
  "cc": [
    "https://example.org/accounts/someone/followers"
  ],
  "type": "Video",
  "id": "https://example.org/videos/watch/942d51e6-9320-4f40-980b-76bba0652bc2",
  "url": [
    {
      "type": "Link",
      "mediaType": "text/html",
      "href": "https://example.org/w/jifTXYpdLJSU269svW8Jdb"
    },
    {
      "type": "Link",
      "mediaType": "text/html",
      "href": "https://example.org/videos/watch/942d51e6-9320-4f40-980b-76bba0652bc2"
    },
    {
      "type": "Link",
      "mediaType": "application/x-mpegURL",
      "href": "https://example.org/static/streaming-playlists/hls/942d51e6-9320-4f40-980b-76bba0652bc2/3d412b0f-3f2e-4509-9d0f-0142223b1752-master.m3u8",
      "tag": [
        {
          "type": "Infohash",
          "name": "4b5a702f76333963655575616e627a5261426269"
        },
        {
          "type": "Infohash",
          "name": "4f6c5552324a39324a55447036735649586b4875"
        },
        {
          "type": "Infohash",
          "name": "476d4154793667574d6d594c7276523471364732"
        },
        {
          "type": "Link",
          "name": "sha256",
          "mediaType": "application/json",
          "href": "https://example.org/static/streaming-playlists/hls/942d51e6-9320-4f40-980b-76bba0652bc2/0c607a4c-ab78-4bed-aeef-9970abd88e77-segments-sha256.json"
        },
        {
          "type": "Link",
          "mediaType": "video/mp4",
          "href": "https://example.org/static/streaming-playlists/hls/942d51e6-9320-4f40-980b-76bba0652bc2/c6b6c9fb-83da-425c-9ce6-680e00eb9ecb-480-fragmented.mp4",
          "height": 480,
          "width": 854,
          "size": 11260985,
          "fps": 30,
          "attachment": [
            {
              "type": "PropertyValue",
              "name": "ffprobe_codec_type",
              "value": "video"
            },
            {
              "type": "PropertyValue",
              "name": "peertube_format_flag",
              "value": "fragmented"
            }
          ]
        },
        {
          "type": "Link",
          "rel": [
            "metadata",
            "video/mp4"
          ],
          "mediaType": "application/json",
          "href": "https://example.org/api/v1/videos/942d51e6-9320-4f40-980b-76bba0652bc2/metadata/236424",
          "height": 480,
          "width": 854,
          "fps": 30
        },
        {
          "type": "Link",
          "mediaType": "application/x-bittorrent",
          "href": "https://example.org/lazy-static/torrents/fac3fb9c-55a6-4e56-82f5-8de8a3f62d8f-480-hls.torrent",
          "height": 480,
          "width": 854,
          "fps": 30
        },
        {
          "type": "Link",
          "mediaType": "application/x-bittorrent;x-scheme-handler/magnet",
          "href": "magnet:?xs=https%3A%2F%2Fexample.org%2Flazy-static%2Ftorrents%2Ffac3fb9c-55a6-4e56-82f5-8de8a3f62d8f-480-hls.torrent&xt=urn:btih:b5a55918c3a05c2459156b6f34570ea64c69fd5a&dn=Na+proch%C3%A1zce+%E2%99%A5%EF%B8%8F+Walking+with+our+gang+%F0%9F%98%83&tr=https%3A%2F%2Fexample.org%2Ftracker%2Fannounce&tr=wss%3A%2F%2Fexample.org%3A443%2Ftracker%2Fsocket&ws=https%3A%2F%2Fexample.org%2Fstatic%2Fstreaming-playlists%2Fhls%2F942d51e6-9320-4f40-980b-76bba0652bc2%2Fc6b6c9fb-83da-425c-9ce6-680e00eb9ecb-480-fragmented.mp4",
          "height": 480,
          "width": 854,
          "fps": 30
        },
        {
          "type": "Link",
          "mediaType": "video/mp4",
          "href": "https://example.org/static/streaming-playlists/hls/942d51e6-9320-4f40-980b-76bba0652bc2/113ebf59-8e27-42f5-b971-d315f3fec77d-0-fragmented.mp4",
          "height": 0,
          "width": 0,
          "size": 1472647,
          "fps": 0,
          "attachment": [
            {
              "type": "PropertyValue",
              "name": "ffprobe_codec_type",
              "value": "audio"
            },
            {
              "type": "PropertyValue",
              "name": "peertube_format_flag",
              "value": "fragmented"
            }
          ]
        },
        {
          "type": "Link",
          "rel": [
            "metadata",
            "video/mp4"
          ],
          "mediaType": "application/json",
          "href": "https://example.org/api/v1/videos/942d51e6-9320-4f40-980b-76bba0652bc2/metadata/236425",
          "height": 0,
          "width": 0,
          "fps": 0
        },
        {
          "type": "Link",
          "mediaType": "application/x-bittorrent",
          "href": "https://example.org/lazy-static/torrents/babc50e0-4643-4467-bde0-5d837d71fed5-0-hls.torrent",
          "height": 0,
          "width": 0,
          "fps": 0
        },
        {
          "type": "Link",
          "mediaType": "application/x-bittorrent;x-scheme-handler/magnet",
          "href": "magnet:?xs=https%3A%2F%2Fexample.org%2Flazy-static%2Ftorrents%2Fbabc50e0-4643-4467-bde0-5d837d71fed5-0-hls.torrent&xt=urn:btih:395086b81fae8b1b9f7d0a03375c66214acff459&dn=Na+proch%C3%A1zce+%E2%99%A5%EF%B8%8F+Walking+with+our+gang+%F0%9F%98%83&tr=https%3A%2F%2Fexample.org%2Ftracker%2Fannounce&tr=wss%3A%2F%2Fexample.org%3A443%2Ftracker%2Fsocket&ws=https%3A%2F%2Fexample.org%2Fstatic%2Fstreaming-playlists%2Fhls%2F942d51e6-9320-4f40-980b-76bba0652bc2%2F113ebf59-8e27-42f5-b971-d315f3fec77d-0-fragmented.mp4",
          "height": 0,
          "width": 0,
          "fps": 0
        }
      ]
    },
    {
      "type": "Link",
      "name": "tracker-http",
      "rel": [
        "tracker",
        "http"
      ],
      "href": "https://example.org/tracker/announce"
    },
    {
      "type": "Link",
      "name": "tracker-websocket",
      "rel": [
        "tracker",
        "websocket"
      ],
      "href": "wss://example.org:443/tracker/socket"
    }
  ]
}`

	// Mix of plain IRIs and Links,
	// we should be able to parse this.
	//
	// The last one with no href should be ignored.
	const mixedPlainURIsAndLinks = `{
  "@context": [
    "https://www.w3.org/ns/activitystreams"
  ],
  "to": [
    "https://www.w3.org/ns/activitystreams#Public"
  ],
  "cc": [
    "https://example.org/accounts/someone/followers"
  ],
  "type": "Video",
  "id": "https://example.org/videos/watch/942d51e6-9320-4f40-980b-76bba0652bc2",
  "url": [
	"https://example.org/videos/watch/942d51e6-9320-4f40-980b-76bba0652bc2",
    {
      "type": "Link",
      "mediaType": "text/html",
      "href": "https://example.org/w/jifTXYpdLJSU269svW8Jdb"
    },
    {
      "type": "Link",
      "mediaType": "text/html",
      "href": "https://example.org/videos/watch/942d51e6-9320-4f40-980b-76bba0652bc2"
    },
    {
      "type": "Link",
      "mediaType": "text/html"
    }
  ]
}`

	for i, test := range []struct {
		in           string
		expectedURLs []*url.URL
	}{
		{
			in: peertubeVideo,
			expectedURLs: []*url.URL{
				testrig.URLMustParse("https://example.org/w/jifTXYpdLJSU269svW8Jdb"),
				testrig.URLMustParse("https://example.org/videos/watch/942d51e6-9320-4f40-980b-76bba0652bc2"),
				testrig.URLMustParse("https://example.org/static/streaming-playlists/hls/942d51e6-9320-4f40-980b-76bba0652bc2/3d412b0f-3f2e-4509-9d0f-0142223b1752-master.m3u8"),
				testrig.URLMustParse("https://example.org/tracker/announce"),
				testrig.URLMustParse("wss://example.org:443/tracker/socket"),
			},
		},
		{
			in: mixedPlainURIsAndLinks,
			expectedURLs: []*url.URL{
				testrig.URLMustParse("https://example.org/videos/watch/942d51e6-9320-4f40-980b-76bba0652bc2"),
				testrig.URLMustParse("https://example.org/w/jifTXYpdLJSU269svW8Jdb"),
				testrig.URLMustParse("https://example.org/videos/watch/942d51e6-9320-4f40-980b-76bba0652bc2"),
			},
		},
	} {
		// Parse input to statusable.
		statusable, err := ap.ResolveStatusable(
			suite.T().Context(),
			io.NopCloser(bytes.NewBufferString(test.in)),
		)
		if err != nil {
			suite.FailNow(err.Error())
		}

		// Ensure URL fields as expected.
		suite.EqualValues(
			test.expectedURLs,
			ap.GetURL(statusable),
			"mismatch in test case %d", i,
		)
	}
}

func TestPropertiesTestSuite(t *testing.T) {
	suite.Run(t, new(PropertiesTestSuite))
}

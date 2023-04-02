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
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type NormalizeTestSuite struct {
	suite.Suite
}

func (suite *NormalizeTestSuite) GetStatusable() (vocab.ActivityStreamsNote, map[string]interface{}) {
	rawJson := `{
		"@context": [
		  "https://www.w3.org/ns/activitystreams",
		  "https://example.org/schemas/litepub-0.1.jsonld",
		  {
			"@language": "und"
		  }
		],
		"actor": "https://example.org/users/someone",
		"attachment": [],
		"attributedTo": "https://example.org/users/someone",
		"cc": [
		  "https://example.org/users/someone/followers"
		],
		"content": "UPDATE: As of this morning there are now more than 7 million Mastodon users, most from the <a class=\"hashtag\" data-tag=\"twittermigration\" href=\"https://example.org/tag/twittermigration\" rel=\"tag ugc\">#TwitterMigration</a>.<br><br>In fact, 100,000 new accounts have been created since last night.<br><br>Since last night&#39;s spike 8,000-12,000 new accounts are being created every hour.<br><br>Yesterday, I estimated that Mastodon would have 8 million users by the end of the week. That might happen a lot sooner if this trend continues.",
		"context": "https://example.org/contexts/01GX0MSHPER1E0FT022Q209EJZ",
		"conversation": "https://example.org/contexts/01GX0MSHPER1E0FT022Q209EJZ",
		"id": "https://example.org/objects/01GX0MT2PA58JNSMK11MCS65YD",
		"published": "2022-11-18T17:43:58.489995Z",
		"replies": {
		  "items": [
			"https://example.org/objects/01GX0MV12MGEG3WF9SWB5K3KRJ"
		  ],
		  "type": "Collection"
		},
		"repliesCount": 0,
		"sensitive": null,
		"source": "UPDATE: As of this morning there are now more than 7 million Mastodon users, most from the #TwitterMigration.\r\n\r\nIn fact, 100,000 new accounts have been created since last night.\r\n\r\nSince last night's spike 8,000-12,000 new accounts are being created every hour.\r\n\r\nYesterday, I estimated that Mastodon would have 8 million users by the end of the week. That might happen a lot sooner if this trend continues.",
		"summary": "",
		"tag": [
		  {
			"href": "https://example.org/tags/twittermigration",
			"name": "#twittermigration",
			"type": "Hashtag"
		  }
		],
		"to": [
		  "https://www.w3.org/ns/activitystreams#Public"
		],
		"type": "Note"
	  }`

	var rawNote map[string]interface{}
	err := json.Unmarshal([]byte(rawJson), &rawNote)
	if err != nil {
		panic(err)
	}

	t, err := streams.ToType(context.Background(), rawNote)
	if err != nil {
		panic(err)
	}

	return t.(vocab.ActivityStreamsNote), rawNote
}

func (suite *NormalizeTestSuite) TestNormalizeActivityObject() {
	note, rawNote := suite.GetStatusable()
	suite.Equal(`update: As of this morning there are now more than 7 million Mastodon users, most from the <a class="hashtag" data-tag="twittermigration" href="https://example.org/tag/twittermigration" rel="tag ugc">#TwitterMigration%3C/a%3E.%3Cbr%3E%3Cbr%3EIn%20fact,%20100,000%20new%20accounts%20have%20been%20created%20since%20last%20night.%3Cbr%3E%3Cbr%3ESince%20last%20night&%2339;s%20spike%208,000-12,000%20new%20accounts%20are%20being%20created%20every%20hour.%3Cbr%3E%3Cbr%3EYesterday,%20I%20estimated%20that%20Mastodon%20would%20have%208%20million%20users%20by%20the%20end%20of%20the%20week.%20That%20might%20happen%20a%20lot%20sooner%20if%20this%20trend%20continues.`, ap.ExtractContent(note))

	create := testrig.WrapAPNoteInCreate(
		testrig.URLMustParse("https://example.org/create_something"),
		testrig.URLMustParse("https://example.org/users/someone"),
		testrig.TimeMustParse("2022-11-18T17:43:58.489995Z"),
		note,
	)

	ap.NormalizeActivityObject(create, map[string]interface{}{"object": rawNote})
	suite.Equal(`UPDATE: As of this morning there are now more than 7 million Mastodon users, most from the <a class="hashtag" data-tag="twittermigration" href="https://example.org/tag/twittermigration" rel="tag ugc">#TwitterMigration</a>.<br><br>In fact, 100,000 new accounts have been created since last night.<br><br>Since last night&#39;s spike 8,000-12,000 new accounts are being created every hour.<br><br>Yesterday, I estimated that Mastodon would have 8 million users by the end of the week. That might happen a lot sooner if this trend continues.`, ap.ExtractContent(note))
}

func TestNormalizeTestSuite(t *testing.T) {
	suite.Run(t, new(NormalizeTestSuite))
}

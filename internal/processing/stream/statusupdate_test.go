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

package stream_test

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/suite"
	statusfilter "github.com/superseriousbusiness/gotosocial/internal/filter/status"
	"github.com/superseriousbusiness/gotosocial/internal/stream"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
)

type StatusUpdateTestSuite struct {
	StreamTestSuite
}

func (suite *StatusUpdateTestSuite) TestStreamNotification() {
	account := suite.testAccounts["local_account_1"]

	openStream, errWithCode := suite.streamProcessor.Open(context.Background(), account, "user")
	suite.NoError(errWithCode)

	editedStatus := suite.testStatuses["remote_account_1_status_1"]
	apiStatus, err := typeutils.NewConverter(&suite.state).StatusToAPIStatus(context.Background(), editedStatus, account, statusfilter.FilterContextNotifications, nil, nil)
	suite.NoError(err)

	suite.streamProcessor.StatusUpdate(context.Background(), account, apiStatus, stream.TimelineHome)

	msg, ok := openStream.Recv(context.Background())
	suite.True(ok)

	dst := new(bytes.Buffer)
	err = json.Indent(dst, []byte(msg.Payload), "", "  ")
	suite.NoError(err)
	suite.Equal(`{
  "id": "01FVW7JHQFSFK166WWKR8CBA6M",
  "created_at": "2021-09-20T10:40:37.000Z",
  "in_reply_to_id": null,
  "in_reply_to_account_id": null,
  "sensitive": false,
  "spoiler_text": "",
  "visibility": "unlisted",
  "language": "en",
  "uri": "http://fossbros-anonymous.io/users/foss_satan/statuses/01FVW7JHQFSFK166WWKR8CBA6M",
  "url": "http://fossbros-anonymous.io/@foss_satan/statuses/01FVW7JHQFSFK166WWKR8CBA6M",
  "replies_count": 0,
  "reblogs_count": 0,
  "favourites_count": 0,
  "favourited": false,
  "reblogged": false,
  "muted": false,
  "bookmarked": false,
  "pinned": false,
  "content": "dark souls status bot: \"thoughts of dog\"",
  "reblog": null,
  "account": {
    "id": "01F8MH5ZK5VRH73AKHQM6Y9VNX",
    "username": "foss_satan",
    "acct": "foss_satan@fossbros-anonymous.io",
    "display_name": "big gerald",
    "locked": false,
    "discoverable": true,
    "bot": false,
    "created_at": "2021-09-26T10:52:36.000Z",
    "note": "i post about like, i dunno, stuff, or whatever!!!!",
    "url": "http://fossbros-anonymous.io/@foss_satan",
    "avatar": "",
    "avatar_static": "",
    "header": "http://localhost:8080/assets/default_header.webp",
    "header_static": "http://localhost:8080/assets/default_header.webp",
    "header_description": "Flat gray background (default header).",
    "followers_count": 0,
    "following_count": 0,
    "statuses_count": 3,
    "last_status_at": "2021-09-11",
    "emojis": [],
    "fields": []
  },
  "media_attachments": [
    {
      "id": "01FVW7RXPQ8YJHTEXYPE7Q8ZY0",
      "type": "image",
      "url": "http://localhost:8080/fileserver/01F8MH5ZK5VRH73AKHQM6Y9VNX/attachment/original/01FVW7RXPQ8YJHTEXYPE7Q8ZY0.jpg",
      "text_url": "http://localhost:8080/fileserver/01F8MH5ZK5VRH73AKHQM6Y9VNX/attachment/original/01FVW7RXPQ8YJHTEXYPE7Q8ZY0.jpg",
      "preview_url": "http://localhost:8080/fileserver/01F8MH5ZK5VRH73AKHQM6Y9VNX/attachment/small/01FVW7RXPQ8YJHTEXYPE7Q8ZY0.webp",
      "remote_url": "http://fossbros-anonymous.io/attachments/original/13bbc3f8-2b5e-46ea-9531-40b4974d9912.jpg",
      "preview_remote_url": null,
      "meta": {
        "original": {
          "width": 472,
          "height": 291,
          "size": "472x291",
          "aspect": 1.6219932
        },
        "small": {
          "width": 472,
          "height": 291,
          "size": "472x291",
          "aspect": 1.6219932
        },
        "focus": {
          "x": 0,
          "y": 0
        }
      },
      "description": "tweet from thoughts of dog: i drank. all the water. in my bowl. earlier. but just now. i returned. to the same bowl. and it was. full again.. the bowl. is haunted",
      "blurhash": "L3Q9_@4n9E?axW4mD$Mx~q00Di%L"
    }
  ],
  "mentions": [],
  "tags": [],
  "emojis": [],
  "card": null,
  "poll": null,
  "interaction_policy": {
    "can_favourite": {
      "always": [
        "public",
        "me"
      ],
      "with_approval": []
    },
    "can_reply": {
      "always": [
        "public",
        "me"
      ],
      "with_approval": []
    },
    "can_reblog": {
      "always": [
        "public",
        "me"
      ],
      "with_approval": []
    }
  }
}`, dst.String())
	suite.Equal(msg.Event, "status.update")
}

func TestStatusUpdateTestSuite(t *testing.T) {
	suite.Run(t, &StatusUpdateTestSuite{})
}

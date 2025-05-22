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

package spam_test

import (
	"bytes"
	"io"
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/ap"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/id"
	"github.com/stretchr/testify/suite"
)

type StatusableTestSuite struct {
	FilterStandardTestSuite
}

const (
	// Message that mentions 5 people (including receiver),
	// and contains a errant link.
	spam1 = `{
  "@context": [
    "https://www.w3.org/ns/activitystreams",
    {
      "ostatus": "http://ostatus.org#",
      "atomUri": "ostatus:atomUri",
      "inReplyToAtomUri": "ostatus:inReplyToAtomUri",
      "conversation": "ostatus:conversation",
      "sensitive": "as:sensitive",
      "toot": "http://joinmastodon.org/ns#",
      "votersCount": "toot:votersCount"
    }
  ],
  "id": "http://fossbros-anonymous.io/users/foss_satan/statuses/111985188827079562",
  "type": "Note",
  "summary": null,
  "inReplyTo": null,
  "published": "2024-02-24T07:06:14Z",
  "url": "http://fossbros-anonymous.io/@foss_satan/111985188827079562",
  "attributedTo": "http://fossbros-anonymous.io/users/foss_satan",
  "to": [
    "https://www.w3.org/ns/activitystreams#Public"
  ],
  "cc": [
    "http://fossbros-anonymous.io/users/foss_satan/followers",
    "https://example.org/users/9gol6f8zff",
    "https://example.net/users/nityosan",
    "https://a.misskey.instance.com/users/9c06ylkgsx",
    "https://another.misskey.instance.com/users/9eklgce5yk",
    "http://localhost:8080/users/the_mighty_zork"
  ],
  "sensitive": false,
  "atomUri": "http://fossbros-anonymous.io/users/foss_satan/statuses/111985188827079562",
  "inReplyToAtomUri": null,
  "content": "<p><a href=\"https://spammylink.org/\" target=\"_blank\" rel=\"nofollow noopener noreferrer\" translate=\"no\"><span class=\"invisible\">https://</span><span class=\"\">spammylink.org/</span><span class=\"invisible\"></span></a></p><p><span class=\"h-card\" translate=\"no\"><a href=\"https://example.org/@Nao_ya_ia22\" class=\"u-url mention\">@<span>Nao_ya_ia22</span></a></span><br /><span class=\"h-card\" translate=\"no\"><a href=\"https://example.net/@nityosan\" class=\"u-url mention\">@<span>nityosan</span></a></span><br /><span class=\"h-card\" translate=\"no\"><a href=\"https://a.misskey.instance.com/@FIzxive\" class=\"u-url mention\">@<span>FIzxive</span></a></span><br /><span class=\"h-card\" translate=\"no\"><a href=\"https://another.misskey.instance.com/@mendako\" class=\"u-url mention\">@<span>mendako</span></a></span><br /><span class=\"h-card\" translate=\"no\"><a href=\"http://localhost:8080/@the_mighty_zork\" class=\"u-url mention\">@<span>the_mighty_zork</span></a></span></p>",
  "attachment": [],
  "tag": [
    {
      "type": "Mention",
      "href": "https://example.org/users/9gol6f8zff",
      "name": "@Nao_ya_ia22@example.org"
    },
    {
      "type": "Mention",
      "href": "https://example.net/users/nityosan",
      "name": "@nityosan@example.net"
    },
    {
      "type": "Mention",
      "href": "https://a.misskey.instance.com/users/9c06ylkgsx",
      "name": "@FIzxive@a.misskey.instance.com"
    },
    {
      "type": "Mention",
      "href": "https://another.misskey.instance.com/users/9eklgce5yk",
      "name": "@mendako@another.misskey.instance.com"
    },
    {
      "type": "Mention",
      "href": "http://localhost:8080/users/the_mighty_zork",
      "name": "@the_mighty_zork@localhost:8080"
    }
  ],
  "replies": {
    "id": "http://fossbros-anonymous.io/users/foss_satan/statuses/111985188827079562/replies",
    "type": "Collection",
    "first": {
      "type": "CollectionPage",
      "next": "http://fossbros-anonymous.io/users/foss_satan/statuses/111985188827079562/replies?only_other_accounts=true&page=true",
      "partOf": "http://fossbros-anonymous.io/users/foss_satan/statuses/111985188827079562/replies",
      "items": []
    }
  }
}`

	// Message that mentions 4 people (including receiver),
	// and contains a errant link.
	spam2 = `{
  "@context": [
    "https://www.w3.org/ns/activitystreams",
    {
      "ostatus": "http://ostatus.org#",
      "atomUri": "ostatus:atomUri",
      "inReplyToAtomUri": "ostatus:inReplyToAtomUri",
      "conversation": "ostatus:conversation",
      "sensitive": "as:sensitive",
      "toot": "http://joinmastodon.org/ns#",
      "votersCount": "toot:votersCount"
    }
  ],
  "id": "http://fossbros-anonymous.io/users/foss_satan/statuses/111985188827079562",
  "type": "Note",
  "summary": null,
  "inReplyTo": null,
  "published": "2024-02-24T07:06:14Z",
  "url": "http://fossbros-anonymous.io/@foss_satan/111985188827079562",
  "attributedTo": "http://fossbros-anonymous.io/users/foss_satan",
  "to": [
    "https://www.w3.org/ns/activitystreams#Public"
  ],
  "cc": [
    "http://fossbros-anonymous.io/users/foss_satan/followers",
    "https://example.net/users/nityosan",
    "https://a.misskey.instance.com/users/9c06ylkgsx",
    "https://another.misskey.instance.com/users/9eklgce5yk",
    "http://localhost:8080/users/the_mighty_zork"
  ],
  "sensitive": false,
  "atomUri": "http://fossbros-anonymous.io/users/foss_satan/statuses/111985188827079562",
  "inReplyToAtomUri": null,
  "content": "<p><a href=\"https://spammylink.org/\" target=\"_blank\" rel=\"nofollow noopener noreferrer\" translate=\"no\"><span class=\"invisible\">https://</span><span class=\"\">spammylink.org/</span><span class=\"invisible\"></span></a></p><p><span class=\"h-card\" translate=\"no\"><a href=\"https://example.net/@nityosan\" class=\"u-url mention\">@<span>nityosan</span></a></span><br /><span class=\"h-card\" translate=\"no\"><a href=\"https://a.misskey.instance.com/@FIzxive\" class=\"u-url mention\">@<span>FIzxive</span></a></span><br /><span class=\"h-card\" translate=\"no\"><a href=\"https://another.misskey.instance.com/@mendako\" class=\"u-url mention\">@<span>mendako</span></a></span><br /><span class=\"h-card\" translate=\"no\"><a href=\"http://localhost:8080/@the_mighty_zork\" class=\"u-url mention\">@<span>the_mighty_zork</span></a></span></p>",
  "attachment": [],
  "tag": [
    {
      "type": "Mention",
      "href": "https://example.net/users/nityosan",
      "name": "@nityosan@example.net"
    },
    {
      "type": "Mention",
      "href": "https://a.misskey.instance.com/users/9c06ylkgsx",
      "name": "@FIzxive@a.misskey.instance.com"
    },
    {
      "type": "Mention",
      "href": "https://another.misskey.instance.com/users/9eklgce5yk",
      "name": "@mendako@another.misskey.instance.com"
    },
    {
      "type": "Mention",
      "href": "http://localhost:8080/users/the_mighty_zork",
      "name": "@the_mighty_zork@localhost:8080"
    }
  ],
  "replies": {
    "id": "http://fossbros-anonymous.io/users/foss_satan/statuses/111985188827079562/replies",
    "type": "Collection",
    "first": {
      "type": "CollectionPage",
      "next": "http://fossbros-anonymous.io/users/foss_satan/statuses/111985188827079562/replies?only_other_accounts=true&page=true",
      "partOf": "http://fossbros-anonymous.io/users/foss_satan/statuses/111985188827079562/replies",
      "items": []
    }
  }
}`

	// Message that mentions 4 people (including receiver),
	// but contains no errant links.
	spam3 = `{
  "@context": [
    "https://www.w3.org/ns/activitystreams",
    {
      "ostatus": "http://ostatus.org#",
      "atomUri": "ostatus:atomUri",
      "inReplyToAtomUri": "ostatus:inReplyToAtomUri",
      "conversation": "ostatus:conversation",
      "sensitive": "as:sensitive",
      "toot": "http://joinmastodon.org/ns#",
      "votersCount": "toot:votersCount"
    }
  ],
  "id": "http://fossbros-anonymous.io/users/foss_satan/statuses/111985188827079562",
  "type": "Note",
  "summary": null,
  "inReplyTo": null,
  "published": "2024-02-24T07:06:14Z",
  "url": "http://fossbros-anonymous.io/@foss_satan/111985188827079562",
  "attributedTo": "http://fossbros-anonymous.io/users/foss_satan",
  "to": [
    "https://www.w3.org/ns/activitystreams#Public"
  ],
  "cc": [
    "http://fossbros-anonymous.io/users/foss_satan/followers",
    "https://example.net/users/nityosan",
    "https://a.misskey.instance.com/users/9c06ylkgsx",
    "https://another.misskey.instance.com/users/9eklgce5yk",
    "http://localhost:8080/users/the_mighty_zork"
  ],
  "sensitive": false,
  "atomUri": "http://fossbros-anonymous.io/users/foss_satan/statuses/111985188827079562",
  "inReplyToAtomUri": null,
  "content": "<p><span class=\"h-card\" translate=\"no\"><a href=\"https://example.net/@nityosan\" class=\"u-url mention\">@<span>nityosan</span></a></span><br /><span class=\"h-card\" translate=\"no\"><a href=\"https://a.misskey.instance.com/@FIzxive\" class=\"u-url mention\">@<span>FIzxive</span></a></span><br /><span class=\"h-card\" translate=\"no\"><a href=\"https://another.misskey.instance.com/@mendako\" class=\"u-url mention\">@<span>mendako</span></a></span><br /><span class=\"h-card\" translate=\"no\"><a href=\"http://localhost:8080/@the_mighty_zork\" class=\"u-url mention\">@<span>the_mighty_zork</span></a></span></p>",
  "attachment": [],
  "tag": [
    {
      "type": "Mention",
      "href": "https://example.net/users/nityosan",
      "name": "@nityosan@example.net"
    },
    {
      "type": "Mention",
      "href": "https://a.misskey.instance.com/users/9c06ylkgsx",
      "name": "@FIzxive@a.misskey.instance.com"
    },
    {
      "type": "Mention",
      "href": "https://another.misskey.instance.com/users/9eklgce5yk",
      "name": "@mendako@another.misskey.instance.com"
    },
    {
      "type": "Mention",
      "href": "http://localhost:8080/users/the_mighty_zork",
      "name": "@the_mighty_zork@localhost:8080"
    }
  ],
  "replies": {
    "id": "http://fossbros-anonymous.io/users/foss_satan/statuses/111985188827079562/replies",
    "type": "Collection",
    "first": {
      "type": "CollectionPage",
      "next": "http://fossbros-anonymous.io/users/foss_satan/statuses/111985188827079562/replies?only_other_accounts=true&page=true",
      "partOf": "http://fossbros-anonymous.io/users/foss_satan/statuses/111985188827079562/replies",
      "items": []
    }
  }
}`

	// Message that mentions 4 people (including receiver),
	// contains no errant links, but 1 attachment.
	spam4 = `{
  "@context": [
    "https://www.w3.org/ns/activitystreams",
    {
      "ostatus": "http://ostatus.org#",
      "atomUri": "ostatus:atomUri",
      "inReplyToAtomUri": "ostatus:inReplyToAtomUri",
      "conversation": "ostatus:conversation",
      "sensitive": "as:sensitive",
      "toot": "http://joinmastodon.org/ns#",
      "votersCount": "toot:votersCount"
    }
  ],
  "id": "http://fossbros-anonymous.io/users/foss_satan/statuses/111985188827079562",
  "type": "Note",
  "summary": null,
  "inReplyTo": null,
  "published": "2024-02-24T07:06:14Z",
  "url": "http://fossbros-anonymous.io/@foss_satan/111985188827079562",
  "attributedTo": "http://fossbros-anonymous.io/users/foss_satan",
  "to": [
    "https://www.w3.org/ns/activitystreams#Public"
  ],
  "cc": [
    "http://fossbros-anonymous.io/users/foss_satan/followers",
    "https://example.net/users/nityosan",
    "https://a.misskey.instance.com/users/9c06ylkgsx",
    "https://another.misskey.instance.com/users/9eklgce5yk",
    "http://localhost:8080/users/the_mighty_zork"
  ],
  "sensitive": false,
  "atomUri": "http://fossbros-anonymous.io/users/foss_satan/statuses/111985188827079562",
  "inReplyToAtomUri": null,
  "content": "<p><span class=\"h-card\" translate=\"no\"><a href=\"https://example.net/@nityosan\" class=\"u-url mention\">@<span>nityosan</span></a></span><br /><span class=\"h-card\" translate=\"no\"><a href=\"https://a.misskey.instance.com/@FIzxive\" class=\"u-url mention\">@<span>FIzxive</span></a></span><br /><span class=\"h-card\" translate=\"no\"><a href=\"https://another.misskey.instance.com/@mendako\" class=\"u-url mention\">@<span>mendako</span></a></span><br /><span class=\"h-card\" translate=\"no\"><a href=\"http://localhost:8080/@the_mighty_zork\" class=\"u-url mention\">@<span>the_mighty_zork</span></a></span></p>",
  "attachment": [
    {
      "blurhash": "LIIE|gRj00WB-;j[t7j[4nWBj[Rj",
      "mediaType": "image/jpeg",
      "name": "",
      "type": "Document",
      "url": "http://fossbros-anonymous.io/fileserver/01F8MH17FWEB39HZJ76B6VXSKF/attachment/original/01F8MH6NEM8D7527KZAECTCR76.jpg"
    }
  ],
  "tag": [
    {
      "type": "Mention",
      "href": "https://example.net/users/nityosan",
      "name": "@nityosan@example.net"
    },
    {
      "type": "Mention",
      "href": "https://a.misskey.instance.com/users/9c06ylkgsx",
      "name": "@FIzxive@a.misskey.instance.com"
    },
    {
      "type": "Mention",
      "href": "https://another.misskey.instance.com/users/9eklgce5yk",
      "name": "@mendako@another.misskey.instance.com"
    },
    {
      "type": "Mention",
      "href": "http://localhost:8080/users/the_mighty_zork",
      "name": "@the_mighty_zork@localhost:8080"
    }
  ],
  "replies": {
    "id": "http://fossbros-anonymous.io/users/foss_satan/statuses/111985188827079562/replies",
    "type": "Collection",
    "first": {
      "type": "CollectionPage",
      "next": "http://fossbros-anonymous.io/users/foss_satan/statuses/111985188827079562/replies?only_other_accounts=true&page=true",
      "partOf": "http://fossbros-anonymous.io/users/foss_satan/statuses/111985188827079562/replies",
      "items": []
    }
  }
}`

	// Message that mentions 4 people (including receiver),
	// and contains a errant link, and receiver follows
	// another mentioned account.
	spam5 = `{
  "@context": [
    "https://www.w3.org/ns/activitystreams",
    {
      "ostatus": "http://ostatus.org#",
      "atomUri": "ostatus:atomUri",
      "inReplyToAtomUri": "ostatus:inReplyToAtomUri",
      "conversation": "ostatus:conversation",
      "sensitive": "as:sensitive",
      "toot": "http://joinmastodon.org/ns#",
      "votersCount": "toot:votersCount"
    }
  ],
  "id": "http://fossbros-anonymous.io/users/foss_satan/statuses/111985188827079562",
  "type": "Note",
  "summary": null,
  "inReplyTo": null,
  "published": "2024-02-24T07:06:14Z",
  "url": "http://fossbros-anonymous.io/@foss_satan/111985188827079562",
  "attributedTo": "http://fossbros-anonymous.io/users/foss_satan",
  "to": [
    "https://www.w3.org/ns/activitystreams#Public"
  ],
  "cc": [
    "http://fossbros-anonymous.io/users/foss_satan/followers",
    "https://example.net/users/nityosan",
    "https://a.misskey.instance.com/users/9c06ylkgsx",
    "http://localhost:8080/users/admin",
    "http://localhost:8080/users/the_mighty_zork"
  ],
  "sensitive": false,
  "atomUri": "http://fossbros-anonymous.io/users/foss_satan/statuses/111985188827079562",
  "inReplyToAtomUri": null,
  "content": "<p><a href=\"https://spammylink.org/\" target=\"_blank\" rel=\"nofollow noopener noreferrer\" translate=\"no\"><span class=\"invisible\">https://</span><span class=\"\">spammylink.org/</span><span class=\"invisible\"></span></a></p><p><span class=\"h-card\" translate=\"no\"><a href=\"https://example.net/@nityosan\" class=\"u-url mention\">@<span>nityosan</span></a></span><br /><span class=\"h-card\" translate=\"no\"><a href=\"https://a.misskey.instance.com/@FIzxive\" class=\"u-url mention\">@<span>FIzxive</span></a></span><br /><span class=\"h-card\" translate=\"no\"><a href=\"http://localhost:8080/@admin\" class=\"u-url mention\">@<span>admin</span></a></span><br /><span class=\"h-card\" translate=\"no\"><a href=\"http://localhost:8080/@the_mighty_zork\" class=\"u-url mention\">@<span>the_mighty_zork</span></a></span></p>",
  "attachment": [],
  "tag": [
    {
      "type": "Mention",
      "href": "https://example.net/users/nityosan",
      "name": "@nityosan@example.net"
    },
    {
      "type": "Mention",
      "href": "https://a.misskey.instance.com/users/9c06ylkgsx",
      "name": "@FIzxive@a.misskey.instance.com"
    },
    {
      "type": "Mention",
      "href": "http://localhost:8080/users/admin",
      "name": "@admin@localhost:8080"
    },
    {
      "type": "Mention",
      "href": "http://localhost:8080/users/the_mighty_zork",
      "name": "@the_mighty_zork@localhost:8080"
    }
  ],
  "replies": {
    "id": "http://fossbros-anonymous.io/users/foss_satan/statuses/111985188827079562/replies",
    "type": "Collection",
    "first": {
      "type": "CollectionPage",
      "next": "http://fossbros-anonymous.io/users/foss_satan/statuses/111985188827079562/replies?only_other_accounts=true&page=true",
      "partOf": "http://fossbros-anonymous.io/users/foss_satan/statuses/111985188827079562/replies",
      "items": []
    }
  }
}`

	// Message that mentions 3 people, contains a
	// errant link, and receiver follows another
	// mentioned account. However, receiver is not mentioned.
	spam6 = `{
  "@context": [
    "https://www.w3.org/ns/activitystreams",
    {
      "ostatus": "http://ostatus.org#",
      "atomUri": "ostatus:atomUri",
      "inReplyToAtomUri": "ostatus:inReplyToAtomUri",
      "conversation": "ostatus:conversation",
      "sensitive": "as:sensitive",
      "toot": "http://joinmastodon.org/ns#",
      "votersCount": "toot:votersCount"
    }
  ],
  "id": "http://fossbros-anonymous.io/users/foss_satan/statuses/111985188827079562",
  "type": "Note",
  "summary": null,
  "inReplyTo": null,
  "published": "2024-02-24T07:06:14Z",
  "url": "http://fossbros-anonymous.io/@foss_satan/111985188827079562",
  "attributedTo": "http://fossbros-anonymous.io/users/foss_satan",
  "to": [
    "https://www.w3.org/ns/activitystreams#Public"
  ],
  "cc": [
    "http://fossbros-anonymous.io/users/foss_satan/followers",
    "https://example.net/users/nityosan",
    "https://a.misskey.instance.com/users/9c06ylkgsx",
    "http://localhost:8080/users/admin"
  ],
  "sensitive": false,
  "atomUri": "http://fossbros-anonymous.io/users/foss_satan/statuses/111985188827079562",
  "inReplyToAtomUri": null,
  "content": "<p><a href=\"https://spammylink.org/\" target=\"_blank\" rel=\"nofollow noopener noreferrer\" translate=\"no\"><span class=\"invisible\">https://</span><span class=\"\">spammylink.org/</span><span class=\"invisible\"></span></a></p><p><span class=\"h-card\" translate=\"no\"><a href=\"https://example.net/@nityosan\" class=\"u-url mention\">@<span>nityosan</span></a></span><br /><span class=\"h-card\" translate=\"no\"><a href=\"https://a.misskey.instance.com/@FIzxive\" class=\"u-url mention\">@<span>FIzxive</span></a></span><br /><span class=\"h-card\" translate=\"no\"><a href=\"http://localhost:8080/@admin\" class=\"u-url mention\">@<span>admin</span></a></span></p>",
  "attachment": [],
  "tag": [
    {
      "type": "Mention",
      "href": "https://example.net/users/nityosan",
      "name": "@nityosan@example.net"
    },
    {
      "type": "Mention",
      "href": "https://a.misskey.instance.com/users/9c06ylkgsx",
      "name": "@FIzxive@a.misskey.instance.com"
    },
    {
      "type": "Mention",
      "href": "http://localhost:8080/users/admin",
      "name": "@admin@localhost:8080"
    }
  ],
  "replies": {
    "id": "http://fossbros-anonymous.io/users/foss_satan/statuses/111985188827079562/replies",
    "type": "Collection",
    "first": {
      "type": "CollectionPage",
      "next": "http://fossbros-anonymous.io/users/foss_satan/statuses/111985188827079562/replies?only_other_accounts=true&page=true",
      "partOf": "http://fossbros-anonymous.io/users/foss_satan/statuses/111985188827079562/replies",
      "items": []
    }
  }
}`

	// Message that mentions 4 people (including receiver),
	// and hash a hashtag, but contains no errant links.
	spam7 = `{
    "@context": [
      "https://www.w3.org/ns/activitystreams",
      {
        "ostatus": "http://ostatus.org#",
        "atomUri": "ostatus:atomUri",
        "inReplyToAtomUri": "ostatus:inReplyToAtomUri",
        "conversation": "ostatus:conversation",
        "sensitive": "as:sensitive",
        "toot": "http://joinmastodon.org/ns#",
        "votersCount": "toot:votersCount"
      }
    ],
    "id": "http://fossbros-anonymous.io/users/foss_satan/statuses/111985188827079562",
    "type": "Note",
    "summary": null,
    "inReplyTo": null,
    "published": "2024-02-24T07:06:14Z",
    "url": "http://fossbros-anonymous.io/@foss_satan/111985188827079562",
    "attributedTo": "http://fossbros-anonymous.io/users/foss_satan",
    "to": [
      "https://www.w3.org/ns/activitystreams#Public"
    ],
    "cc": [
      "http://fossbros-anonymous.io/users/foss_satan/followers",
      "https://example.net/users/nityosan",
      "https://a.misskey.instance.com/users/9c06ylkgsx",
      "https://another.misskey.instance.com/users/9eklgce5yk",
      "http://localhost:8080/users/the_mighty_zork"
    ],
    "sensitive": false,
    "atomUri": "http://fossbros-anonymous.io/users/foss_satan/statuses/111985188827079562",
    "inReplyToAtomUri": null,
    "content": "<p><a href=\"https://fossbros-anonymous.io/tags/gotosocial\" class=\"mention hashtag\" rel=\"tag\">#<span>gotosocial</span></a> smells<br/><br/><span class=\"h-card\" translate=\"no\"><a href=\"https://example.net/@nityosan\" class=\"u-url mention\">@<span>nityosan</span></a></span><br /><span class=\"h-card\" translate=\"no\"><a href=\"https://a.misskey.instance.com/@FIzxive\" class=\"u-url mention\">@<span>FIzxive</span></a></span><br /><span class=\"h-card\" translate=\"no\"><a href=\"https://another.misskey.instance.com/@mendako\" class=\"u-url mention\">@<span>mendako</span></a></span><br /><span class=\"h-card\" translate=\"no\"><a href=\"http://localhost:8080/@the_mighty_zork\" class=\"u-url mention\">@<span>the_mighty_zork</span></a></span></p>",
    "attachment": [],
    "tag": [
      {
        "type": "Mention",
        "href": "https://example.net/users/nityosan",
        "name": "@nityosan@example.net"
      },
      {
        "type": "Mention",
        "href": "https://a.misskey.instance.com/users/9c06ylkgsx",
        "name": "@FIzxive@a.misskey.instance.com"
      },
      {
        "type": "Mention",
        "href": "https://another.misskey.instance.com/users/9eklgce5yk",
        "name": "@mendako@another.misskey.instance.com"
      },
      {
        "type": "Mention",
        "href": "http://localhost:8080/users/the_mighty_zork",
        "name": "@the_mighty_zork@localhost:8080"
      },
      {
        "type": "Hashtag",
        "href": "https://fossbros-anonymous.io/tags/gotosocial",
        "name": "#gotosocial"
      }  
    ],
    "replies": {
      "id": "http://fossbros-anonymous.io/users/foss_satan/statuses/111985188827079562/replies",
      "type": "Collection",
      "first": {
        "type": "CollectionPage",
        "next": "http://fossbros-anonymous.io/users/foss_satan/statuses/111985188827079562/replies?only_other_accounts=true&page=true",
        "partOf": "http://fossbros-anonymous.io/users/foss_satan/statuses/111985188827079562/replies",
        "items": []
      }
    }
  }`

	// Same as spam7, except message doesn't
	// have a hashtag in the tags array.
	spam8 = `{
    "@context": [
      "https://www.w3.org/ns/activitystreams",
      {
        "ostatus": "http://ostatus.org#",
        "atomUri": "ostatus:atomUri",
        "inReplyToAtomUri": "ostatus:inReplyToAtomUri",
        "conversation": "ostatus:conversation",
        "sensitive": "as:sensitive",
        "toot": "http://joinmastodon.org/ns#",
        "votersCount": "toot:votersCount"
      }
    ],
    "id": "http://fossbros-anonymous.io/users/foss_satan/statuses/111985188827079562",
    "type": "Note",
    "summary": null,
    "inReplyTo": null,
    "published": "2024-02-24T07:06:14Z",
    "url": "http://fossbros-anonymous.io/@foss_satan/111985188827079562",
    "attributedTo": "http://fossbros-anonymous.io/users/foss_satan",
    "to": [
      "https://www.w3.org/ns/activitystreams#Public"
    ],
    "cc": [
      "http://fossbros-anonymous.io/users/foss_satan/followers",
      "https://example.net/users/nityosan",
      "https://a.misskey.instance.com/users/9c06ylkgsx",
      "https://another.misskey.instance.com/users/9eklgce5yk",
      "http://localhost:8080/users/the_mighty_zork"
    ],
    "sensitive": false,
    "atomUri": "http://fossbros-anonymous.io/users/foss_satan/statuses/111985188827079562",
    "inReplyToAtomUri": null,
    "content": "<p><a href=\"https://fossbros-anonymous.io/tags/gotosocial\" class=\"mention hashtag\" rel=\"tag\">#<span>gotosocial</span></a> smells<br/><br/><span class=\"h-card\" translate=\"no\"><a href=\"https://example.net/@nityosan\" class=\"u-url mention\">@<span>nityosan</span></a></span><br /><span class=\"h-card\" translate=\"no\"><a href=\"https://a.misskey.instance.com/@FIzxive\" class=\"u-url mention\">@<span>FIzxive</span></a></span><br /><span class=\"h-card\" translate=\"no\"><a href=\"https://another.misskey.instance.com/@mendako\" class=\"u-url mention\">@<span>mendako</span></a></span><br /><span class=\"h-card\" translate=\"no\"><a href=\"http://localhost:8080/@the_mighty_zork\" class=\"u-url mention\">@<span>the_mighty_zork</span></a></span></p>",
    "attachment": [],
    "tag": [
      {
        "type": "Mention",
        "href": "https://example.net/users/nityosan",
        "name": "@nityosan@example.net"
      },
      {
        "type": "Mention",
        "href": "https://a.misskey.instance.com/users/9c06ylkgsx",
        "name": "@FIzxive@a.misskey.instance.com"
      },
      {
        "type": "Mention",
        "href": "https://another.misskey.instance.com/users/9eklgce5yk",
        "name": "@mendako@another.misskey.instance.com"
      },
      {
        "type": "Mention",
        "href": "http://localhost:8080/users/the_mighty_zork",
        "name": "@the_mighty_zork@localhost:8080"
      }
    ],
    "replies": {
      "id": "http://fossbros-anonymous.io/users/foss_satan/statuses/111985188827079562/replies",
      "type": "Collection",
      "first": {
        "type": "CollectionPage",
        "next": "http://fossbros-anonymous.io/users/foss_satan/statuses/111985188827079562/replies?only_other_accounts=true&page=true",
        "partOf": "http://fossbros-anonymous.io/users/foss_satan/statuses/111985188827079562/replies",
        "items": []
      }
    }
  }`
)

func (suite *StatusableTestSuite) TestStatusableOK() {
	var (
		ctx       = suite.T().Context()
		receiver  = suite.testAccounts["local_account_1"]
		requester = suite.testAccounts["remote_account_1"]
	)

	type testStruct struct {
		message string
		check   func(error)
	}

	for _, test := range []testStruct{
		{
			// SPAM: status mentions 5 or more people
			message: spam1,
			check: func(err error) {
				suite.True(gtserror.IsSpam(err), "expected Spam, got %+v", err)
			},
		},
		{
			// SPAM: receiver doesn't know a mentioned account, and status has attachments or errant links
			message: spam2,
			check: func(err error) {
				suite.True(gtserror.IsSpam(err), "expected Spam, got %+v", err)
			},
		},
		{
			// NOT SPAM: receiver doesn't know a mentioned account, but status has no attachments or errant links
			message: spam3,
			check: func(err error) {
				suite.NoError(err, "expected not spam, got %+v", err)
			},
		},
		{
			// SPAM: receiver doesn't know a mentioned account, and status has attachments or errant links
			message: spam4,
			check: func(err error) {
				suite.True(gtserror.IsSpam(err), "expected Spam, got %+v", err)
			},
		},
		{
			// NOT SPAM: receiver knows a mentioned account
			message: spam5,
			check: func(err error) {
				suite.NoError(err, "expected not spam, got %+v", err)
			},
		},
		{
			// SPAM: receiver does not follow requester, and is not mentioned
			message: spam6,
			check: func(err error) {
				suite.True(gtserror.IsNotRelevant(err), "expected NotRelevant, got %+v", err)
			},
		},
		{
			// NOT SPAM: receiver doesn't know a mentioned account, but status has no attachments or errant links
			message: spam7,
			check: func(err error) {
				suite.NoError(err, "expected not spam, got %+v", err)
			},
		},
		{
			// SPAM: receiver doesn't know a mentioned account, and status has attachments or errant links
			message: spam8,
			check: func(err error) {
				suite.True(gtserror.IsSpam(err), "expected Spam, got %+v", err)
			},
		},
	} {
		rc := io.NopCloser(bytes.NewReader([]byte(test.message)))

		statusable, err := ap.ResolveStatusable(ctx, rc)
		if err != nil {
			suite.FailNow(err.Error())
		}

		err = suite.filter.StatusableOK(ctx, receiver, requester, statusable)
		test.check(err)
	}

	// Put a follow in place from receiver to requester.
	fID := id.NewULID()
	if err := suite.state.DB.PutFollow(ctx, &gtsmodel.Follow{
		ID:              fID,
		URI:             "http://localhost:8080/users/the_mighty_zork/follows/" + fID,
		AccountID:       receiver.ID,
		TargetAccountID: requester.ID,
	}); err != nil {
		suite.FailNow(err.Error())
	}

	// Run all the tests again. They should all
	// be OK since receiver now follows requester.
	for _, test := range []testStruct{
		{
			message: spam1,
			check: func(err error) {
				suite.NoError(err, "expected not spam, got %+v", err)
			},
		},
		{
			message: spam2,
			check: func(err error) {
				suite.NoError(err, "expected not spam, got %+v", err)
			},
		},
		{
			message: spam3,
			check: func(err error) {
				suite.NoError(err, "expected not spam, got %+v", err)
			},
		},
		{
			message: spam4,
			check: func(err error) {
				suite.NoError(err, "expected not spam, got %+v", err)
			},
		},
		{
			message: spam5,
			check: func(err error) {
				suite.NoError(err, "expected not spam, got %+v", err)
			},
		},
		{
			message: spam6,
			check: func(err error) {
				suite.NoError(err, "expected not spam, got %+v", err)
			},
		},
		{
			message: spam7,
			check: func(err error) {
				suite.NoError(err, "expected not spam, got %+v", err)
			},
		},
		{
			message: spam8,
			check: func(err error) {
				suite.NoError(err, "expected not spam, got %+v", err)
			},
		},
	} {
		rc := io.NopCloser(bytes.NewReader([]byte(test.message)))

		statusable, err := ap.ResolveStatusable(ctx, rc)
		if err != nil {
			suite.FailNow(err.Error())
		}

		err = suite.filter.StatusableOK(ctx, receiver, requester, statusable)
		test.check(err)
	}
}

func TestStatusableTestSuite(t *testing.T) {
	suite.Run(t, &StatusableTestSuite{})
}

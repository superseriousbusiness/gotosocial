package ap_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
)

type NormalizeTestSuite struct {
	suite.Suite
}

func (suite *NormalizeTestSuite) GetStatusable() (ap.Statusable, map[string]interface{}) {
	rawJson := `{
		"@context": [
		  "https://www.w3.org/ns/activitystreams",
		  "https://atomicpoet.org/schemas/litepub-0.1.jsonld",
		  {
			"@language": "und"
		  }
		],
		"actor": "https://atomicpoet.org/users/atomicpoet",
		"attachment": [
		  {
			"mediaType": "image/png",
			"name": "Chart showing huge spike in Mastodon sign-ups",
			"type": "Document",
			"url": "https://atomicpoet.org/media/5af783bf7b425400563acb15b3b2994ffa3932df5598bedf7718842df800a29e.png"
		  }
		],
		"attributedTo": "https://atomicpoet.org/users/atomicpoet",
		"cc": [
		  "https://atomicpoet.org/users/atomicpoet/followers"
		],
		"content": "UPDATE: As of this morning there are now more than 7 million Mastodon users, most from the <a class=\"hashtag\" data-tag=\"twittermigration\" href=\"https://atomicpoet.org/tag/twittermigration\" rel=\"tag ugc\">#TwitterMigration</a>.<br><br>In fact, 100,000 new accounts have been created since last night.<br><br>Since last night&#39;s spike 8,000-12,000 new accounts are being created every hour.<br><br>Yesterday, I estimated that Mastodon would have 8 million users by the end of the week. That might happen a lot sooner if this trend continues.",
		"context": "https://atomicpoet.org/contexts/f03aa94b-e71a-4dd1-911c-dd8e1185b3dd",
		"conversation": "https://atomicpoet.org/contexts/f03aa94b-e71a-4dd1-911c-dd8e1185b3dd",
		"id": "https://atomicpoet.org/objects/bd2c6e20-d03d-4c9e-a724-b8738c5df00b",
		"published": "2022-11-18T17:43:58.489995Z",
		"replies": {
		  "items": [
			"https://atomicpoet.org/objects/1bda3590-19c9-496e-bffc-81ee0e5b32e4"
		  ],
		  "type": "Collection"
		},
		"repliesCount": 0,
		"sensitive": null,
		"source": "UPDATE: As of this morning there are now more than 7 million Mastodon users, most from the #TwitterMigration.\r\n\r\nIn fact, 100,000 new accounts have been created since last night.\r\n\r\nSince last night's spike 8,000-12,000 new accounts are being created every hour.\r\n\r\nYesterday, I estimated that Mastodon would have 8 million users by the end of the week. That might happen a lot sooner if this trend continues.",
		"summary": "",
		"tag": [
		  {
			"href": "https://atomicpoet.org/tags/twittermigration",
			"name": "#twittermigration",
			"type": "Hashtag"
		  }
		],
		"to": [
		  "https://www.w3.org/ns/activitystreams#Public"
		],
		"type": "Note"
	  }`

	var jsonAsMap map[string]interface{}
	err := json.Unmarshal([]byte(rawJson), &jsonAsMap)
	if err != nil {
		panic(err)
	}

	t, err := streams.ToType(context.Background(), jsonAsMap)
	if err != nil {
		panic(err)
	}

	return t.(ap.Statusable), jsonAsMap
}

func (suite *NormalizeTestSuite) TestNormalizeStatusable() {
	statusable, rawJson := 
}

func TestNormalizeTestSuite(t *testing.T) {
	suite.Run(t, new(NormalizeTestSuite))
}

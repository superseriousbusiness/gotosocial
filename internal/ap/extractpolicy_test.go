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

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

type ExtractPolicyTestSuite struct {
	APTestSuite
}

func (suite *ExtractPolicyTestSuite) TestExtractPolicy() {
	rawNote := `{
  "@context": [
    "https://gotosocial.org/ns",
    "https://www.w3.org/ns/activitystreams"
  ],
  "content": "hey @f0x and @dumpsterqueer",
  "contentMap": {
    "en": "hey @f0x and @dumpsterqueer",
    "fr": "bonjour @f0x et @dumpsterqueer"
  },
  "interactionPolicy": {
    "canLike": {
      "always": [
        "https://www.w3.org/ns/activitystreams#Public"
      ],
      "approvalRequired": []
    },
    "canReply": {
      "always": [
        "http://localhost:8080/users/the_mighty_zork",
        "http://localhost:8080/users/the_mighty_zork/followers",
        "https://gts.superseriousbusiness.org/users/dumpsterqueer",
        "https://gts.superseriousbusiness.org/users/f0x"
      ],
      "approvalRequired": [
        "https://www.w3.org/ns/activitystreams#Public"
      ]
    },
    "canAnnounce": {
      "always": [
        "http://localhost:8080/users/the_mighty_zork"
      ],
      "approvalRequired": [
        "https://www.w3.org/ns/activitystreams#Public"
      ]
    }
  },
  "tag": [
    {
      "href": "https://gts.superseriousbusiness.org/users/dumpsterqueer",
      "name": "@dumpsterqueer@superseriousbusiness.org",
      "type": "Mention"
    },
    {
      "href": "https://gts.superseriousbusiness.org/users/f0x",
      "name": "@f0x@superseriousbusiness.org",
      "type": "Mention"
    }
  ],
  "type": "Note"
}`

	statusable, err := ap.ResolveStatusable(
		context.Background(),
		io.NopCloser(
			bytes.NewBufferString(rawNote),
		),
	)
	if err != nil {
		suite.FailNow(err.Error())
	}

	policy := ap.ExtractInteractionPolicy(
		statusable,
		// Zork didn't actually create
		// this status but nevermind.
		suite.testAccounts["local_account_1"],
	)

	expectedPolicy := &gtsmodel.InteractionPolicy{
		CanLike: gtsmodel.PolicyRules{
			Always: gtsmodel.PolicyValues{
				gtsmodel.PolicyValuePublic,
			},
			WithApproval: gtsmodel.PolicyValues{},
		},
		CanReply: gtsmodel.PolicyRules{
			Always: gtsmodel.PolicyValues{
				gtsmodel.PolicyValueAuthor,
				gtsmodel.PolicyValueFollowers,
				"https://gts.superseriousbusiness.org/users/dumpsterqueer",
				"https://gts.superseriousbusiness.org/users/f0x",
			},
			WithApproval: gtsmodel.PolicyValues{
				gtsmodel.PolicyValuePublic,
			},
		},
		CanAnnounce: gtsmodel.PolicyRules{
			Always: gtsmodel.PolicyValues{
				gtsmodel.PolicyValueAuthor,
			},
			WithApproval: gtsmodel.PolicyValues{
				gtsmodel.PolicyValuePublic,
			},
		},
	}
	suite.EqualValues(expectedPolicy, policy)
}

func TestExtractPolicyTestSuite(t *testing.T) {
	suite.Run(t, &ExtractPolicyTestSuite{})
}

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
	"fmt"
	"testing"

	"code.superseriousbusiness.org/activity/streams"
	"code.superseriousbusiness.org/gotosocial/internal/ap"
	"github.com/stretchr/testify/suite"
)

type ExtractFocusTestSuite struct {
	APTestSuite
}

func (suite *ExtractFocusTestSuite) TestExtractFocus() {
	ctx := suite.T().Context()

	type test struct {
		data    string
		expectX float32
		expectY float32
	}

	for _, test := range []test{
		{
			// Fine.
			data:    "-0.5, 0.5",
			expectX: -0.5,
			expectY: 0.5,
		},
		{
			// Also fine.
			data:    "1, 1",
			expectX: 1,
			expectY: 1,
		},
		{
			// Out of range.
			data:    "1.5, 1",
			expectX: 0,
			expectY: 0,
		},
		{
			// Too many points.
			data:    "1, 1, 0",
			expectX: 0,
			expectY: 0,
		},
		{
			// Not enough points.
			data:    "1",
			expectX: 0,
			expectY: 0,
		},
	} {
		// Wrap provided test.data
		// in a minimal Attachmentable.
		const fmts = `{
  "@context": [
    "https://www.w3.org/ns/activitystreams",
    {
      "focalPoint": {
        "@container": "@list",
        "@id": "toot:focalPoint"
      },
      "toot": "http://joinmastodon.org/ns#"
    }
  ],
  "focalPoint": [ %s ],
  "type": "Image"
}`

		// Unmarshal test data.
		data := fmt.Sprintf(fmts, test.data)
		m := make(map[string]any)
		if err := json.Unmarshal([]byte(data), &m); err != nil {
			suite.FailNow(err.Error())
		}

		// Convert to type.
		t, err := streams.ToType(ctx, m)
		if err != nil {
			suite.FailNow(err.Error())
		}

		// Convert to attachmentable.
		attachmentable, ok := t.(ap.Attachmentable)
		if !ok {
			suite.FailNow("", "%T was not Attachmentable", t)
		}

		// Check extracted focus.
		focus := ap.ExtractFocus(attachmentable)
		if focus.X != test.expectX || focus.Y != test.expectY {
			suite.Fail("",
				"expected x=%.2f y=%.2f got x=%.2f y=%.2f",
				test.expectX, test.expectY, focus.X, focus.Y,
			)
		}
	}
}

func TestExtractFocusTestSuite(t *testing.T) {
	suite.Run(t, new(ExtractFocusTestSuite))
}

package ap_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"codeberg.org/superseriousbusiness/activity/streams"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
)

type ExtractFocusTestSuite struct {
	APTestSuite
}

func (suite *ExtractFocusTestSuite) TestExtractFocus() {
	ctx := context.Background()

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

package util_test

import (
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/api/util"
)

func TestIsASContentType(t *testing.T) {
	for _, test := range []struct {
		Input  string
		Expect bool
	}{
		{
			Input:  "application/activity+json",
			Expect: true,
		},
		{
			Input:  "application/activity+json; charset=utf-8",
			Expect: true,
		},
		{
			Input:  "application/activity+json;charset=utf-8",
			Expect: true,
		},
		{
			Input:  "application/activity+json ;charset=utf-8",
			Expect: true,
		},
		{
			Input:  "application/activity+json ; charset=utf-8",
			Expect: true,
		},
		{
			Input:  "application/ld+json;profile=https://www.w3.org/ns/activitystreams",
			Expect: true,
		},
		{
			Input:  "application/ld+json;profile=\"https://www.w3.org/ns/activitystreams\"",
			Expect: true,
		},
		{
			Input:  "application/ld+json ;profile=https://www.w3.org/ns/activitystreams",
			Expect: true,
		},
		{
			Input:  "application/ld+json ;profile=\"https://www.w3.org/ns/activitystreams\"",
			Expect: true,
		},
		{
			Input:  "application/ld+json ; profile=https://www.w3.org/ns/activitystreams",
			Expect: true,
		},
		{
			Input:  "application/ld+json ; profile=\"https://www.w3.org/ns/activitystreams\"",
			Expect: true,
		},
		{
			Input:  "application/ld+json; profile=https://www.w3.org/ns/activitystreams",
			Expect: true,
		},
		{
			Input:  "application/ld+json; profile=\"https://www.w3.org/ns/activitystreams\"",
			Expect: true,
		},
		{
			Input:  "application/ld+json",
			Expect: false,
		},
	} {
		if util.ASContentType(test.Input) != test.Expect {
			t.Errorf("did not get expected result %v for input: %s", test.Expect, test.Input)
		}
	}
}

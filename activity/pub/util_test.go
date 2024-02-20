package pub

import (
	"testing"
)

func TestHeaderIsActivityPubMediaType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			"Mastodon Accept Header",
			"application/activity+json, application/ld+json",
			true,
		},
		{
			"Plain Type",
			"application/activity+json",
			true,
		},
		{
			"Missing Profile",
			"application/ld+json",
			false,
		},
		{
			"With Profile",
			"application/ld+json ; profile=https://www.w3.org/ns/activitystreams",
			true,
		},
		{
			"With Quoted Profile",
			"application/ld+json ; profile=\"https://www.w3.org/ns/activitystreams\"",
			true,
		},
		{
			"With Profile (End Space)",
			"application/ld+json; profile=https://www.w3.org/ns/activitystreams",
			true,
		},
		{
			"With Quoted Profile (End Space)",
			"application/ld+json; profile=\"https://www.w3.org/ns/activitystreams\"",
			true,
		},
		{
			"With Profile (Begin Space)",
			"application/ld+json ;profile=https://www.w3.org/ns/activitystreams",
			true,
		},
		{
			"With Quoted Profile (Begin Space)",
			"application/ld+json ;profile=\"https://www.w3.org/ns/activitystreams\"",
			true,
		},
		{
			"With Profile (No Space)",
			"application/ld+json;profile=https://www.w3.org/ns/activitystreams",
			true,
		},
		{
			"With Quoted Profile (No Space)",
			"application/ld+json;profile=\"https://www.w3.org/ns/activitystreams\"",
			true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if actual := headerIsActivityPubMediaType(test.input); actual != test.expected {
				t.Fatalf("expected %v, got %v", test.expected, actual)
			}
		})
	}
}

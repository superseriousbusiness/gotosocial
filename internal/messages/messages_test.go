package messages_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

var fromClientAPICases = []struct {
	msg  messages.FromClientAPI
	data []byte
}{
	{
		msg: messages.FromClientAPI{
			APObjectType:   ap.ObjectNote,
			APActivityType: ap.ActivityCreate,
			GTSModel:       &gtsmodel.Status{ID: "69", Content: "hehe"},
			TargetURI:      "https://gotosocial.org",
			Origin:         &gtsmodel.Account{ID: "654321"},
			Target:         &gtsmodel.Account{ID: "123456"},
		},
		data: toJSON(map[string]any{
			"ap_object_type":   ap.ObjectNote,
			"ap_activity_type": ap.ActivityCreate,
			"gts_model":        json.RawMessage(toJSON(&gtsmodel.Status{ID: "69", Content: "hehe"})),
			"gts_model_type":   "*gtsmodel.Status",
			"target_uri":       "https://gotosocial.org",
			"origin_id":        "654321",
			"target_id":        "123456",
		}),
	},
	{
		msg: messages.FromClientAPI{
			APObjectType:   ap.ObjectProfile,
			APActivityType: ap.ActivityUpdate,
			GTSModel:       &gtsmodel.Account{ID: "420", DisplayName: "Her Fuckin' Maj Queen Liz", Memorial: util.Ptr(true)},
			TargetURI:      "https://uk-queen-is-dead.org",
			Origin:         &gtsmodel.Account{ID: "123456"},
			Target:         &gtsmodel.Account{ID: "654321"},
		},
		data: toJSON(map[string]any{
			"ap_object_type":   ap.ObjectProfile,
			"ap_activity_type": ap.ActivityUpdate,
			"gts_model":        json.RawMessage(toJSON(&gtsmodel.Account{ID: "420", DisplayName: "Her Fuckin' Maj Queen Liz", Memorial: util.Ptr(true)})),
			"gts_model_type":   "*gtsmodel.Account",
			"target_uri":       "https://uk-queen-is-dead.org",
			"origin_id":        "123456",
			"target_id":        "654321",
		}),
	},
}

var fromFediAPICases = []struct {
	msg  messages.FromFediAPI
	data []byte
}{
	// {
	// 	msg:  messages.FromFediAPI{},
	// 	data: toJSON(map[string]any{}),
	// },
}

func TestSerializeFromClientAPI(t *testing.T) {
	for _, test := range fromClientAPICases {
		// Serialize test message to blob.
		data, err := test.msg.Serialize()
		if err != nil {
			t.Fatal(err)
		}

		// Check that serialized JSON data is as expected.
		assert.JSONEq(t, string(test.data), string(data))
	}
}

func TestDeserializeFromClientAPI(t *testing.T) {
	for _, test := range fromClientAPICases {
		var msg messages.FromClientAPI

		// Deserialize test message blob.
		err := msg.Deserialize(test.data)
		if err != nil {
			t.Fatal(err)
		}

		// Check that msg is as expected.
		assert.Equal(t, test.msg, msg)
	}
}

func TestSerializeFromFediAPI(t *testing.T) {
	for _, test := range fromFediAPICases {
		// Serialize test message to blob.
		data, err := test.msg.Serialize()
		if err != nil {
			t.Fatal(err)
		}

		// Check that data is as expected.
		assert.Equal(t, test.data, data)
	}
}

func TestDeserializeFromFediAPI(t *testing.T) {
	for _, test := range fromFediAPICases {
		msg := new(messages.FromFediAPI)

		// Deserialize test message blob.
		err := msg.Deserialize(test.data)
		if err != nil {
			t.Fatal(err)
		}

		// Check that msg is as expected.
		assert.Equal(t, test.msg, msg)
	}

}

func indent(b []byte) string {
	buf := bytes.NewBuffer(nil)
	err := json.Indent(buf, b, "", "    ")
	if err != nil {
		panic(err)
	}
	return buf.String()
}

func toJSON(a any) []byte {
	b, err := json.Marshal(a)
	if err != nil {
		panic(err)
	}
	return b
}

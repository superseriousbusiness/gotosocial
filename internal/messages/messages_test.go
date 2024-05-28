package messages_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
)

var fromClientAPICases = []struct {
	msg  messages.FromClientAPI
	data []byte
}{
	{
		msg:  messages.FromClientAPI{},
		data: toJSON(map[string]any{}),
	},
}

var fromFediAPICases = []struct {
	msg  messages.FromFediAPI
	data []byte
}{
	{
		msg:  messages.FromFediAPI{},
		data: toJSON(map[string]any{}),
	},
}

func TestSerializeFromClientAPI(t *testing.T) {
	for _, test := range fromClientAPICases {
		// Serialize test message to blob.
		data, err := test.msg.Serialize()
		if err != nil {
			t.Fatal(err)
		}

		// Check that data is as expected.
		assert.Equal(t, test.data, data)
	}
}

func TestDeserializeFromClientAPI(t *testing.T) {
	for _, test := range fromClientAPICases {
		msg := new(messages.FromClientAPI)

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

func toJSON(a any) []byte {
	b, err := json.Marshal(a)
	if err != nil {
		panic(err)
	}
	return b
}

package streams

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/go-test/deep"
	"github.com/superseriousbusiness/activity/streams/vocab"
)

// TestTable represents a test entry based on example data from a
// specification.
type TestTable struct {
	// The following are guaranteed to be populated
	name         string
	expectedJSON string
	// The following may be nil
	expectedStruct                vocab.Type
	deserializer                  func(map[string]interface{}) (vocab.Type, error)
	unknown                       func(map[string]interface{}) map[string]interface{}
	skipDeserializationTest       bool
	skipDeserializationTestReason string
}

// Gets the test table for the specification example data.
func GetTestTable() []TestTable {
	return []TestTable{
		{
			name:           "Example 1",
			expectedJSON:   example1,
			expectedStruct: example1Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeObjectActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 2",
			expectedJSON:   example2,
			expectedStruct: example2Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeLinkActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 3",
			expectedJSON:   example3,
			expectedStruct: example3Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeActivityActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 4",
			expectedJSON:   example4,
			expectedStruct: example4Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeTravelActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 5",
			expectedJSON:   example5,
			expectedStruct: example5Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeCollectionActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 6",
			expectedJSON:   example6,
			expectedStruct: example6Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeOrderedCollectionActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 7",
			expectedJSON:   example7,
			expectedStruct: example7Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeCollectionPageActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 8",
			expectedJSON:   example8,
			expectedStruct: example8Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeOrderedCollectionPageActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 9",
			expectedJSON:   example9,
			expectedStruct: example9Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeAcceptActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 10",
			expectedJSON:   example10,
			expectedStruct: example10Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeAcceptActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 11",
			expectedJSON:   example11,
			expectedStruct: example11Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeTentativeAcceptActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 12",
			expectedJSON:   example12,
			expectedStruct: example12Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeAddActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 13",
			expectedJSON:   example13,
			expectedStruct: example13Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeAddActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 14",
			expectedJSON:   example14,
			expectedStruct: example14Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeArriveActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 15",
			expectedJSON:   example15,
			expectedStruct: example15Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeCreateActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 16",
			expectedJSON:   example16,
			expectedStruct: example16Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeDeleteActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 17",
			expectedJSON:   example17,
			expectedStruct: example17Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeFollowActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 18",
			expectedJSON:   example18,
			expectedStruct: example18Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeIgnoreActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 19",
			expectedJSON:   example19,
			expectedStruct: example19Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeJoinActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 20",
			expectedJSON:   example20,
			expectedStruct: example20Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeLeaveActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 21",
			expectedJSON:   example21,
			expectedStruct: example21Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeLeaveActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 22",
			expectedJSON:   example22,
			expectedStruct: example22Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeLikeActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 23",
			expectedJSON:   example23,
			expectedStruct: example23Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeOfferActivityStreams()(m, map[string]string{})
			},
			unknown: example23Unknown,
		},
		{
			name:           "Example 24",
			expectedJSON:   example24,
			expectedStruct: example24Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeInviteActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 25",
			expectedJSON:   example25,
			expectedStruct: example25Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeRejectActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 26",
			expectedJSON:   example26,
			expectedStruct: example26Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeTentativeRejectActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 27",
			expectedJSON:   example27,
			expectedStruct: example27Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeRemoveActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 28",
			expectedJSON:   example28,
			expectedStruct: example28Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeRemoveActivityStreams()(m, map[string]string{})
			},
			unknown: example28Unknown,
		},
		{
			name:           "Example 29",
			expectedJSON:   example29,
			expectedStruct: example29Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeUndoActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 30",
			expectedJSON:   example30,
			expectedStruct: example30Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeUpdateActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 31",
			expectedJSON:   example31,
			expectedStruct: example31Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeViewActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 32",
			expectedJSON:   example32,
			expectedStruct: example32Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeListenActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 33",
			expectedJSON:   example33,
			expectedStruct: example33Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeReadActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 34",
			expectedJSON:   example34,
			expectedStruct: example34Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeMoveActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 35",
			expectedJSON:   example35,
			expectedStruct: example35Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeTravelActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 36",
			expectedJSON:   example36,
			expectedStruct: example36Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeAnnounceActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 37",
			expectedJSON:   example37,
			expectedStruct: example37Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeBlockActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 38",
			expectedJSON:   example38,
			expectedStruct: example38Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeFlagActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 39",
			expectedJSON:   example39,
			expectedStruct: example39Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeDislikeActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 40",
			expectedJSON:   example40,
			expectedStruct: example40Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeQuestionActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 41",
			expectedJSON:   example41,
			expectedStruct: example41Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeQuestionActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 42",
			expectedJSON:   example42,
			expectedStruct: example42Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeApplicationActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 43",
			expectedJSON:   example43,
			expectedStruct: example43Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeGroupActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 44",
			expectedJSON:   example44,
			expectedStruct: example44Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeOrganizationActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 45",
			expectedJSON:   example45,
			expectedStruct: example45Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializePersonActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 46",
			expectedJSON:   example46,
			expectedStruct: example46Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeServiceActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 47",
			expectedJSON:   example47,
			expectedStruct: example47Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeRelationshipActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 48",
			expectedJSON:   example48,
			expectedStruct: example48Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeArticleActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 49",
			expectedJSON:   example49,
			expectedStruct: example49Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeDocumentActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 50",
			expectedJSON:   example50,
			expectedStruct: example50Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeAudioActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 51",
			expectedJSON:   example51,
			expectedStruct: example51Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeImageActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 52",
			expectedJSON:   example52,
			expectedStruct: example52Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeVideoActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 53",
			expectedJSON:   example53,
			expectedStruct: example53Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeNoteActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 54",
			expectedJSON:   example54,
			expectedStruct: example54Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializePageActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 55",
			expectedJSON:   example55,
			expectedStruct: example55Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeEventActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 56",
			expectedJSON:   example56,
			expectedStruct: example56Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializePlaceActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 57",
			expectedJSON:   example57,
			expectedStruct: example57Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializePlaceActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 58",
			expectedJSON:   example58,
			expectedStruct: example58Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeMentionActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 59",
			expectedJSON:   example59,
			expectedStruct: example59Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeProfileActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 60",
			expectedJSON:   example60,
			expectedStruct: example60Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeOrderedCollectionActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:         "Example 61",
			expectedJSON: example61,
			unknown:      example61Unknown,
		},
		{
			name:         "Example 62",
			expectedJSON: example62,
			unknown:      example62Unknown,
		},
		{
			name:           "Example 63",
			expectedJSON:   example63,
			expectedStruct: example63Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeOfferActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 64",
			expectedJSON:   example64,
			expectedStruct: example64Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeOfferActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 65",
			expectedJSON:   example65,
			expectedStruct: example65Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeOfferActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 66",
			expectedJSON:   example66,
			expectedStruct: example66Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeNoteActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 67",
			expectedJSON:   example67,
			expectedStruct: example67Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeImageActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 68",
			expectedJSON:   example68,
			expectedStruct: example68Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeImageActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 69",
			expectedJSON:   example69,
			expectedStruct: example69Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeNoteActivityStreams()(m, map[string]string{})
			},
			unknown: example69Unknown,
		},
		{
			name:           "Example 70",
			expectedJSON:   example70,
			expectedStruct: example70Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeOfferActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 71",
			expectedJSON:   example71,
			expectedStruct: example71Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeOfferActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 72",
			expectedJSON:   example72,
			expectedStruct: example72Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeOfferActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 73",
			expectedJSON:   example73,
			expectedStruct: example73Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeCollectionActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 74",
			expectedJSON:   example74,
			expectedStruct: example74Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeCollectionActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 75",
			expectedJSON:   example75,
			expectedStruct: example75Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeCollectionActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 76",
			expectedJSON:   example76,
			expectedStruct: example76Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeCollectionActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 77",
			expectedJSON:   example77,
			expectedStruct: example77Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeCollectionActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 78",
			expectedJSON:   example78,
			expectedStruct: example78Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeNoteActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 79",
			expectedJSON:   example79,
			expectedStruct: example79Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeNoteActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 80",
			expectedJSON:   example80,
			expectedStruct: example80Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeNoteActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 81",
			expectedJSON:   example81,
			expectedStruct: example81Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeNoteActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 82",
			expectedJSON:   example82,
			expectedStruct: example82Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeNoteActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 83",
			expectedJSON:   example83,
			expectedStruct: example83Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeNoteActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 84",
			expectedJSON:   example84,
			expectedStruct: example84Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeNoteActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 85",
			expectedJSON:   example85,
			expectedStruct: example85Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeListenActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 86",
			expectedJSON:   example86,
			expectedStruct: example86Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeCollectionActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 87",
			expectedJSON:   example87,
			expectedStruct: example87Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeCollectionActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 88",
			expectedJSON:   example88,
			expectedStruct: example88Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializePersonActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 89",
			expectedJSON:   example89,
			expectedStruct: example89Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeCollectionActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 90",
			expectedJSON:   example90,
			expectedStruct: example90Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeOrderedCollectionActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 91",
			expectedJSON:   example91,
			expectedStruct: example91Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeQuestionActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 92",
			expectedJSON:   example92,
			expectedStruct: example92Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeQuestionActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 93",
			expectedJSON:   example93,
			expectedStruct: example93Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeQuestionActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 94",
			expectedJSON:   example94,
			expectedStruct: example94Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeMoveActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 95",
			expectedJSON:   example95,
			expectedStruct: example95Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeCollectionPageActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 96",
			expectedJSON:   example96,
			expectedStruct: example96Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeCollectionPageActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 97",
			expectedJSON:   example97,
			expectedStruct: example97Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeLikeActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 98",
			expectedJSON:   example98,
			expectedStruct: example98Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeLikeActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 99",
			expectedJSON:   example99,
			expectedStruct: example99Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeLikeActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 100",
			expectedJSON:   example100,
			expectedStruct: example100Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeCollectionPageActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 101",
			expectedJSON:   example101,
			expectedStruct: example101Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeCollectionPageActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 102",
			expectedJSON:   example102,
			expectedStruct: example102Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeVideoActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 103",
			expectedJSON:   example103,
			expectedStruct: example103Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeActivityActivityStreams()(m, map[string]string{})
			},
			unknown: example103Unknown,
		},
		{
			name:           "Example 104",
			expectedJSON:   example104,
			expectedStruct: example104Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeNoteActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 105",
			expectedJSON:   example105,
			expectedStruct: example105Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeImageActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 106",
			expectedJSON:   example106,
			expectedStruct: example106Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeOfferActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 107",
			expectedJSON:   example107,
			expectedStruct: example107Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeOfferActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 108",
			expectedJSON:   example108,
			expectedStruct: example108Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeOfferActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 109",
			expectedJSON:   example109,
			expectedStruct: example109Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeDocumentActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 110",
			expectedJSON:   example110,
			expectedStruct: example110Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeDocumentActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 111",
			expectedJSON:   example111,
			expectedStruct: example111Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeDocumentActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 112",
			expectedJSON:   example112,
			expectedStruct: example112Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializePlaceActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 113",
			expectedJSON:   example113,
			expectedStruct: example113Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializePlaceActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 114",
			expectedJSON:   example114,
			expectedStruct: example114Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeNoteActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 115",
			expectedJSON:   example115,
			expectedStruct: example115Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeNoteActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 116",
			expectedJSON:   example116,
			expectedStruct: example116Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeNoteActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 117",
			expectedJSON:   example117,
			expectedStruct: example117Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeNoteActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 118",
			expectedJSON:   example118,
			expectedStruct: example118Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeNoteActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 119",
			expectedJSON:   example119,
			expectedStruct: example119Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeVideoActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 120",
			expectedJSON:   example120,
			expectedStruct: example120Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeLinkActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 121",
			expectedJSON:   example121,
			expectedStruct: example121Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeLinkActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 122",
			expectedJSON:   example122,
			expectedStruct: example122Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeLinkActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 123",
			expectedJSON:   example123,
			expectedStruct: example123Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeCollectionPageActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 124",
			expectedJSON:   example124,
			expectedStruct: example124Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializePlaceActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 125",
			expectedJSON:   example125,
			expectedStruct: example125Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializePlaceActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 126",
			expectedJSON:   example126,
			expectedStruct: example126Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeLinkActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 127",
			expectedJSON:   example127,
			expectedStruct: example127Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeEventActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 128",
			expectedJSON:   example128,
			expectedStruct: example128Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeNoteActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 129",
			expectedJSON:   example129,
			expectedStruct: example129Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeEventActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 130",
			expectedJSON:   example130,
			expectedStruct: example130Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializePlaceActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 131",
			expectedJSON:   example131,
			expectedStruct: example131Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeLinkActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 132",
			expectedJSON:   example132,
			expectedStruct: example132Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeOrderedCollectionPageActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 133",
			expectedJSON:   example133,
			expectedStruct: example133Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeNoteActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 134",
			expectedJSON:   example134,
			expectedStruct: example134Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeNoteActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 135",
			expectedJSON:   example135,
			expectedStruct: example135Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeCollectionActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 136",
			expectedJSON:   example136,
			expectedStruct: example136Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializePlaceActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 137",
			expectedJSON:   example137,
			expectedStruct: example137Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeNoteActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 138",
			expectedJSON:   example138,
			expectedStruct: example138Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeLinkActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 139",
			expectedJSON:   example139,
			expectedStruct: example139Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeRelationshipActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 140",
			expectedJSON:   example140,
			expectedStruct: example140Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeRelationshipActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 141",
			expectedJSON:   example141,
			expectedStruct: example141Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeProfileActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 142",
			expectedJSON:   example142,
			expectedStruct: example142Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeTombstoneActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 143",
			expectedJSON:   example143,
			expectedStruct: example143Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeTombstoneActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 144",
			expectedJSON:   example144,
			expectedStruct: example144Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeCollectionActivityStreams()(m, map[string]string{})
			},
			unknown: example144Unknown,
		},
		{
			name:           "Example 145",
			expectedJSON:   example145,
			expectedStruct: example145Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeCollectionActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 146",
			expectedJSON:   example146,
			expectedStruct: example146Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeCreateActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 147",
			expectedJSON:   example147,
			expectedStruct: example147Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeOfferActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 148",
			expectedJSON:   example148,
			expectedStruct: example148Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeCollectionActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 149",
			expectedJSON:   example149,
			expectedStruct: example149Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializePlaceActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 150",
			expectedJSON:   example150,
			expectedStruct: example150Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializePlaceActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 151",
			expectedJSON:   example151,
			expectedStruct: example151Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeQuestionActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 152",
			expectedJSON:   example152,
			expectedStruct: example152Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeQuestionActivityStreams()(m, map[string]string{})
			},
			unknown: example152Unknown,
		},
		{
			name:         "Example 153",
			expectedJSON: example153,
			unknown:      example153Unknown,
		},
		{
			name:           "Example 154",
			expectedJSON:   example154,
			expectedStruct: example154Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeQuestionActivityStreams()(m, map[string]string{})
			},
			unknown: example154Unknown,
		},
		{
			name:           "Example 155",
			expectedJSON:   example155,
			expectedStruct: example155Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeCollectionActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 156",
			expectedJSON:   example156,
			expectedStruct: example156Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeCollectionActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Example 157",
			expectedJSON:   example157,
			expectedStruct: example157Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeNoteActivityStreams()(m, map[string]string{})
			},
			unknown: example157Unknown,
		},
		{
			name:           "Example 158",
			expectedJSON:   example158,
			expectedStruct: example158Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeNoteActivityStreams()(m, map[string]string{})
			},
			unknown: example158Unknown,
		},
		{
			name:           "Example 159",
			expectedJSON:   example159,
			expectedStruct: example159Type(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeMoveActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Person Example (With Public Key)",
			expectedJSON:   personExampleWithPublicKey,
			expectedStruct: personExampleWithPublicKeyType(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializePersonActivityStreams()(m, map[string]string{})
			},
		},
		{
			name:           "Service w/ Multiple schema:PropertyValue Attachments",
			expectedJSON:   serviceHasAttachmentWithUnknown,
			expectedStruct: serviceHasAttachmentWithUnknownType(),
			deserializer: func(m map[string]interface{}) (vocab.Type, error) {
				return mgr.DeserializeServiceActivityStreams()(m, map[string]string{})
			},
			skipDeserializationTest:       true,
			skipDeserializationTestReason: "If go-fed gets the JSON, it won't match the form of the constructed type.",
		},
	}
}

func TestDeserialization(t *testing.T) {
	deep.CompareUnexportedFields = true
	for _, r := range GetTestTable() {
		if r.skipDeserializationTest {
			t.Logf("Skipping %q: %s", r.name, r.skipDeserializationTestReason)
			continue
		}
		r := r // shadow loop variable
		t.Run(r.name, func(t *testing.T) {
			// Test Deserialize
			m := make(map[string]interface{})
			err := json.Unmarshal([]byte(r.expectedJSON), &m)
			if err != nil {
				t.Errorf("Cannot json.Unmarshal: %v", err)
				return
			}
			if r.deserializer == nil || r.expectedStruct == nil {
				t.Skip("No expected struct or deserializer, skipping deserialization test")
				return
			}
			// Delete the @context -- it will trigger unwanted differences due to the
			// Unknown field.
			delete(m, "@context")
			actual, err := r.deserializer(m)
			if err != nil {
				t.Errorf("Cannot Deserialize: %v", err)
				return
			}
			if diff := reflect.DeepEqual(actual, r.expectedStruct); !diff {
				if r.unknown != nil {
					t.Log("Got expected difference due to unknown additions")
				} else {
					deepDiff := deep.Equal(actual, r.expectedStruct)
					t.Errorf("Deserialize deep equal is false: %v", deepDiff)
				}
			} else if r.unknown != nil {
				t.Error("Expected a difference when there are unknown types")
			}
		})
	}
}

func TestSerialization(t *testing.T) {
	for _, r := range GetTestTable() {
		r := r // shadow loop variable
		t.Run(r.name, func(t *testing.T) {
			m := make(map[string]interface{})
			var err error
			if r.expectedStruct != nil {
				m, err = SerializeForTest(r.expectedStruct)
				if err != nil {
					t.Errorf("Cannot Serialize: %v", err)
					return
				}
			}
			if r.unknown != nil {
				m = r.unknown(m)
			}
			b, err := json.Marshal(m)
			if err != nil {
				t.Errorf("Cannot json.Marshal: %v", err)
				return
			}
			if diff, err := GetJSONDiff(b, []byte(r.expectedJSON)); err == nil && diff != nil {
				t.Error("Serialize JSON equality is false:")
				for _, d := range diff {
					t.Log(d)
				}
			} else if err != nil {
				t.Errorf("GetJSONDiff returned error: %v", err)
			}
		})
	}
}

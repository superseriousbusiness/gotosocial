package pub

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
)

// TestActivityStreamsHandler tests the handler for serving ActivityPub
// requests.
func TestActivityStreamsHandler(t *testing.T) {
	ctx := context.Background()
	setupFn := func(ctl *gomock.Controller) (db *MockDatabase, clock *MockClock, hf HandlerFunc) {
		db = NewMockDatabase(ctl)
		clock = NewMockClock(ctl)
		hf = NewActivityStreamsHandler(db, clock)
		return
	}
	t.Run("IgnoresIfNotActivityPubGetRequest", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		_, _, hf := setupFn(ctl)
		resp := httptest.NewRecorder()
		req := httptest.NewRequest("GET", testNoteId1, nil)
		// Run & Verify
		isAPReq, err := hf(ctx, resp, req)
		assertEqual(t, isAPReq, false)
		assertEqual(t, err, nil)
		assertEqual(t, len(resp.Result().Header), 0)
	})
	t.Run("ReturnsErrorWhenDatabaseFetchReturnsError", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		mockDb, _, hf := setupFn(ctl)
		resp := httptest.NewRecorder()
		req := toAPRequest(httptest.NewRequest("GET", testNoteId1, nil))
		testErr := fmt.Errorf("test error")
		// Mock
		mockDb.EXPECT().Lock(ctx, mustParse(testNoteId1)).Return(func() {}, nil)
		mockDb.EXPECT().Get(ctx, mustParse(testNoteId1)).Return(nil, testErr)
		// Run & Verify
		isAPReq, err := hf(ctx, resp, req)
		assertEqual(t, isAPReq, true)
		assertEqual(t, err, testErr)
		assertEqual(t, len(resp.Result().Header), 0)
	})
	t.Run("ServesTombstoneWithStatusGone", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		mockDb, mockClock, hf := setupFn(ctl)
		resp := httptest.NewRecorder()
		req := toAPRequest(httptest.NewRequest("GET", testNoteId1, nil))
		// Mock
		mockDb.EXPECT().Lock(ctx, mustParse(testNoteId1)).Return(func() {}, nil)
		mockDb.EXPECT().Get(ctx, mustParse(testNoteId1)).Return(testTombstone, nil)
		mockClock.EXPECT().Now().Return(now())
		// Run & Verify
		isAPReq, err := hf(ctx, resp, req)
		assertEqual(t, isAPReq, true)
		assertEqual(t, err, nil)
		assertEqual(t, resp.Code, http.StatusGone)
		respV := resp.Result()
		assertEqual(t, respV.Header.Get(contentTypeHeader), "application/ld+json; profile=\"https://www.w3.org/ns/activitystreams\"")
		assertEqual(t, respV.Header.Get(dateHeader), nowDateHeader())
		assertNotEqual(t, len(respV.Header.Get(digestHeader)), 0)
		b, err := ioutil.ReadAll(respV.Body)
		assertEqual(t, err, nil)
		assertByteEqual(t, b, mustSerializeToBytes(testTombstone))
	})
	t.Run("ServesContentWithStatusOk", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		mockDb, mockClock, hf := setupFn(ctl)
		resp := httptest.NewRecorder()
		req := toAPRequest(httptest.NewRequest("GET", testNoteId1, nil))
		// Mock
		mockDb.EXPECT().Lock(ctx, mustParse(testNoteId1)).Return(func() {}, nil)
		mockDb.EXPECT().Get(ctx, mustParse(testNoteId1)).Return(testMyNote, nil)
		mockClock.EXPECT().Now().Return(now())
		// Run & Verify
		isAPReq, err := hf(ctx, resp, req)
		assertEqual(t, isAPReq, true)
		assertEqual(t, err, nil)
		assertEqual(t, resp.Code, http.StatusOK)
		respV := resp.Result()
		assertEqual(t, respV.Header.Get(contentTypeHeader), "application/ld+json; profile=\"https://www.w3.org/ns/activitystreams\"")
		assertEqual(t, respV.Header.Get(dateHeader), nowDateHeader())
		assertNotEqual(t, len(respV.Header.Get(digestHeader)), 0)
		b, err := ioutil.ReadAll(respV.Body)
		assertEqual(t, err, nil)
		assertByteEqual(t, b, mustSerializeToBytes(testMyNote))
	})
}

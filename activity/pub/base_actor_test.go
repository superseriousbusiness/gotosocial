package pub

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/superseriousbusiness/activity/streams/vocab"
)

// TestBaseActorSocialProtocol tests the Actor returned with NewCustomActor
// and only having the SocialProtocol enabled.
func TestBaseActorSocialProtocol(t *testing.T) {
	// Set up test case
	setupData()
	ctx := context.Background()
	setupFn := func(ctl *gomock.Controller) (delegate *MockDelegateActor, clock *MockClock, a Actor) {
		delegate = NewMockDelegateActor(ctl)
		clock = NewMockClock(ctl)
		a = NewCustomActor(
			delegate,
			/*enableSocialProtocol=*/ true,
			/*enableFederatedProtocol=*/ false,
			clock)
		return
	}
	// Run tests
	t.Run("PostInboxIgnoresNonActivityPubRequest", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		_, _, a := setupFn(ctl)
		resp := httptest.NewRecorder()
		req := toPostInboxRequest(testCreate)
		// Run the test
		handled, err := a.PostInbox(ctx, resp, req)
		// Verify results
		assertEqual(t, err, nil)
		assertEqual(t, handled, false)
		assertEqual(t, len(resp.Result().Header), 0)
	})
	t.Run("PostInboxNotAllowed", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		_, _, a := setupFn(ctl)
		resp := httptest.NewRecorder()
		req := toAPRequest(toPostInboxRequest(testCreate))
		// Run the test
		handled, err := a.PostInbox(ctx, resp, req)
		// Verify results
		assertEqual(t, err, nil)
		assertEqual(t, handled, true)
		assertEqual(t, resp.Code, http.StatusMethodNotAllowed)
	})
	t.Run("GetInboxIgnoresNonActivityPubRequest", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		_, _, a := setupFn(ctl)
		resp := httptest.NewRecorder()
		req := toGetInboxRequest()
		// Run the test
		handled, err := a.GetInbox(ctx, resp, req)
		// Verify results
		assertEqual(t, err, nil)
		assertEqual(t, handled, false)
		assertEqual(t, len(resp.Result().Header), 0)
	})
	t.Run("GetInboxDeniesIfNotAuthenticated", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		delegate, _, a := setupFn(ctl)
		resp := httptest.NewRecorder()
		req := toAPRequest(toGetInboxRequest())
		delegate.EXPECT().AuthenticateGetInbox(ctx, resp, req).DoAndReturn(func(ctx context.Context, resp http.ResponseWriter, req *http.Request) (context.Context, bool, error) {
			resp.WriteHeader(http.StatusForbidden)
			return ctx, false, nil
		})
		// Run the test
		handled, err := a.GetInbox(ctx, resp, req)
		// Verify results
		assertEqual(t, err, nil)
		assertEqual(t, handled, true)
		assertEqual(t, resp.Code, http.StatusForbidden)
	})
	t.Run("GetInboxRespondsWithDataAndHeaders", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		delegate, clock, a := setupFn(ctl)
		resp := httptest.NewRecorder()
		req := toAPRequest(toGetInboxRequest())
		delegate.EXPECT().AuthenticateGetInbox(ctx, resp, req).Return(ctx, true, nil)
		delegate.EXPECT().GetInbox(ctx, req).Return(testOrderedCollectionUniqueElems, nil)
		clock.EXPECT().Now().Return(now())
		// Run the test
		handled, err := a.GetInbox(ctx, resp, req)
		// Verify results
		assertEqual(t, err, nil)
		assertEqual(t, handled, true)
		assertEqual(t, resp.Code, http.StatusOK)
		respV := resp.Result()
		assertEqual(t, respV.Header.Get(contentTypeHeader), "application/ld+json; profile=\"https://www.w3.org/ns/activitystreams\"")
		assertEqual(t, respV.Header.Get(dateHeader), nowDateHeader())
		assertNotEqual(t, len(respV.Header.Get(digestHeader)), 0)
		b, err := ioutil.ReadAll(respV.Body)
		assertEqual(t, err, nil)
		assertByteEqual(t, b, []byte(testOrderedCollectionUniqueElemsString))
	})
	t.Run("GetInboxDeduplicatesData", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		delegate, clock, a := setupFn(ctl)
		resp := httptest.NewRecorder()
		req := toAPRequest(toGetInboxRequest())
		delegate.EXPECT().AuthenticateGetInbox(ctx, resp, req).Return(ctx, true, nil)
		delegate.EXPECT().GetInbox(ctx, req).Return(testOrderedCollectionDupedElems, nil)
		clock.EXPECT().Now().Return(now())
		// Run the test
		_, err := a.GetInbox(ctx, resp, req)
		// Verify results
		assertEqual(t, err, nil)
		respV := resp.Result()
		b, err := ioutil.ReadAll(respV.Body)
		assertEqual(t, err, nil)
		assertByteEqual(t, b, []byte(testOrderedCollectionDedupedElemsString))
	})
	t.Run("PostOutboxIgnoresNonActivityPubRequest", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		_, _, a := setupFn(ctl)
		resp := httptest.NewRecorder()
		req := toPostOutboxRequest(testCreateNoId)
		// Run the test
		handled, err := a.PostOutbox(ctx, resp, req)
		// Verify results
		assertEqual(t, err, nil)
		assertEqual(t, handled, false)
		assertEqual(t, len(resp.Result().Header), 0)
	})
	t.Run("PostOutboxDeniesIfNotAuthenticated", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		delegate, _, a := setupFn(ctl)
		resp := httptest.NewRecorder()
		req := toAPRequest(toPostOutboxRequest(testCreateNoId))
		delegate.EXPECT().AuthenticatePostOutbox(ctx, resp, req).DoAndReturn(func(ctx context.Context, resp http.ResponseWriter, req *http.Request) (context.Context, bool, error) {
			resp.WriteHeader(http.StatusForbidden)
			return ctx, false, nil
		})
		// Run the test
		handled, err := a.PostOutbox(ctx, resp, req)
		// Verify results
		assertEqual(t, err, nil)
		assertEqual(t, handled, true)
		assertEqual(t, resp.Code, http.StatusForbidden)
	})
	t.Run("PostOutboxBadRequestIfUnknownType", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		delegate, _, a := setupFn(ctl)
		resp := httptest.NewRecorder()
		req := toAPRequest(toPostOutboxUnknownRequest())
		delegate.EXPECT().AuthenticatePostOutbox(ctx, resp, req).Return(ctx, true, nil)
		// Run the test
		handled, err := a.PostOutbox(ctx, resp, req)
		// Verify results
		assertEqual(t, err, nil)
		assertEqual(t, handled, true)
		assertEqual(t, resp.Code, http.StatusBadRequest)
	})
	t.Run("PostOutboxRespondsWithDataAndHeaders", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		delegate, _, a := setupFn(ctl)
		resp := httptest.NewRecorder()
		req := toAPRequest(toPostOutboxRequest(testCreateNoId))
		delegate.EXPECT().AuthenticatePostOutbox(ctx, resp, req).Return(ctx, true, nil)
		delegate.EXPECT().PostOutboxRequestBodyHook(ctx, req, toDeserializedForm(testCreateNoId)).Return(ctx, nil)
		delegate.EXPECT().AddNewIDs(ctx, toDeserializedForm(testCreateNoId)).DoAndReturn(func(c context.Context, activity Activity) error {
			withNewId(activity)
			return nil
		})
		delegate.EXPECT().PostOutbox(
			ctx,
			withNewId(toDeserializedForm(testCreateNoId)),
			mustParse(testMyOutboxIRI),
			mustSerialize(testCreateNoId),
		).Return(true, nil)
		// Run the test
		handled, err := a.PostOutbox(ctx, resp, req)
		// Verify results
		assertEqual(t, err, nil)
		assertEqual(t, handled, true)
		assertEqual(t, resp.Code, http.StatusCreated)
		respV := resp.Result()
		assertEqual(t, respV.Header.Get(locationHeader), testNewActivityIRI)
	})
	t.Run("PostOutboxWrapsInCreate", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		delegate, _, a := setupFn(ctl)
		resp := httptest.NewRecorder()
		req := toAPRequest(toPostOutboxRequest(testMyNote))
		delegate.EXPECT().AuthenticatePostOutbox(ctx, resp, req).Return(ctx, true, nil)
		delegate.EXPECT().PostOutboxRequestBodyHook(ctx, req, toDeserializedForm(testMyNote)).Return(ctx, nil)
		delegate.EXPECT().WrapInCreate(ctx, toDeserializedForm(testMyNote), mustParse(testMyOutboxIRI)).DoAndReturn(func(c context.Context, t vocab.Type, u *url.URL) (vocab.ActivityStreamsCreate, error) {
			return wrappedInCreate(t), nil
		})
		delegate.EXPECT().AddNewIDs(ctx, wrappedInCreate(toDeserializedForm(testMyNote))).DoAndReturn(func(c context.Context, activity Activity) error {
			withNewId(activity)
			return nil
		})
		delegate.EXPECT().PostOutbox(
			ctx,
			withNewId(wrappedInCreate(toDeserializedForm(testMyNote))),
			mustParse(testMyOutboxIRI),
			mustSerialize(toDeserializedForm(testMyNote)),
		).Return(true, nil)
		// Run the test
		handled, err := a.PostOutbox(ctx, resp, req)
		// Verify results
		assertEqual(t, err, nil)
		assertEqual(t, handled, true)
		assertEqual(t, resp.Code, http.StatusCreated)
	})
	t.Run("PostOutboxBadRequestForErrObjectRequired", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		delegate, _, a := setupFn(ctl)
		resp := httptest.NewRecorder()
		req := toAPRequest(toPostOutboxRequest(testCreateNoId))
		delegate.EXPECT().AuthenticatePostOutbox(ctx, resp, req).Return(ctx, true, nil)
		delegate.EXPECT().PostOutboxRequestBodyHook(ctx, req, toDeserializedForm(testCreateNoId)).Return(ctx, nil)
		delegate.EXPECT().AddNewIDs(ctx, toDeserializedForm(testCreateNoId)).DoAndReturn(func(c context.Context, activity Activity) error {
			withNewId(activity)
			return nil
		})
		delegate.EXPECT().PostOutbox(
			ctx,
			withNewId(toDeserializedForm(testCreateNoId)),
			mustParse(testMyOutboxIRI),
			mustSerialize(testCreateNoId),
		).Return(true, ErrObjectRequired)
		// Run the test
		handled, err := a.PostOutbox(ctx, resp, req)
		// Verify results
		assertEqual(t, err, nil)
		assertEqual(t, handled, true)
		assertEqual(t, resp.Code, http.StatusBadRequest)
	})
	t.Run("PostOutboxBadRequestForErrTargetRequired", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		delegate, _, a := setupFn(ctl)
		resp := httptest.NewRecorder()
		req := toAPRequest(toPostOutboxRequest(testCreateNoId))
		delegate.EXPECT().AuthenticatePostOutbox(ctx, resp, req).Return(ctx, true, nil)
		delegate.EXPECT().PostOutboxRequestBodyHook(ctx, req, toDeserializedForm(testCreateNoId)).Return(ctx, nil)
		delegate.EXPECT().AddNewIDs(ctx, toDeserializedForm(testCreateNoId)).DoAndReturn(func(c context.Context, activity Activity) error {
			withNewId(activity)
			return nil
		})
		delegate.EXPECT().PostOutbox(
			ctx,
			withNewId(toDeserializedForm(testCreateNoId)),
			mustParse(testMyOutboxIRI),
			mustSerialize(testCreateNoId),
		).Return(true, ErrTargetRequired)
		// Run the test
		handled, err := a.PostOutbox(ctx, resp, req)
		// Verify results
		assertEqual(t, err, nil)
		assertEqual(t, handled, true)
		assertEqual(t, resp.Code, http.StatusBadRequest)
	})
	t.Run("GetOutboxIgnoresNonActivityPubRequest", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		_, _, a := setupFn(ctl)
		resp := httptest.NewRecorder()
		req := toGetOutboxRequest()
		// Run the test
		handled, err := a.GetOutbox(ctx, resp, req)
		// Verify results
		assertEqual(t, err, nil)
		assertEqual(t, handled, false)
		assertEqual(t, len(resp.Result().Header), 0)
	})
	t.Run("GetOutboxDeniesIfNotAuthenticated", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		delegate, _, a := setupFn(ctl)
		resp := httptest.NewRecorder()
		req := toAPRequest(toGetOutboxRequest())
		delegate.EXPECT().AuthenticateGetOutbox(ctx, resp, req).DoAndReturn(func(ctx context.Context, resp http.ResponseWriter, req *http.Request) (context.Context, bool, error) {
			resp.WriteHeader(http.StatusForbidden)
			return ctx, false, nil
		})
		// Run the test
		handled, err := a.GetOutbox(ctx, resp, req)
		// Verify results
		assertEqual(t, err, nil)
		assertEqual(t, handled, true)
		assertEqual(t, resp.Code, http.StatusForbidden)
	})
	t.Run("GetOutboxRespondsWithDataAndHeaders", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		delegate, clock, a := setupFn(ctl)
		resp := httptest.NewRecorder()
		req := toAPRequest(toGetOutboxRequest())
		delegate.EXPECT().AuthenticateGetOutbox(ctx, resp, req).Return(ctx, true, nil)
		delegate.EXPECT().GetOutbox(ctx, req).Return(testOrderedCollectionUniqueElems, nil)
		clock.EXPECT().Now().Return(now())
		// Run the test
		handled, err := a.GetOutbox(ctx, resp, req)
		// Verify results
		assertEqual(t, err, nil)
		assertEqual(t, handled, true)
		assertEqual(t, resp.Code, http.StatusOK)
		respV := resp.Result()
		assertEqual(t, respV.Header.Get(contentTypeHeader), "application/ld+json; profile=\"https://www.w3.org/ns/activitystreams\"")
		assertEqual(t, respV.Header.Get(dateHeader), nowDateHeader())
		assertNotEqual(t, len(respV.Header.Get(digestHeader)), 0)
		b, err := ioutil.ReadAll(respV.Body)
		assertEqual(t, err, nil)
		assertByteEqual(t, b, []byte(testOrderedCollectionUniqueElemsString))
	})
}

// TestBaseActorFederatingProtocol tests the Actor returned with
// NewCustomActor and only having the FederatingProtocol enabled.
func TestBaseActorFederatingProtocol(t *testing.T) {
	// Set up test case
	setupData()
	ctx := context.Background()
	setupFn := func(ctl *gomock.Controller) (delegate *MockDelegateActor, clock *MockClock, a Actor) {
		delegate = NewMockDelegateActor(ctl)
		clock = NewMockClock(ctl)
		a = NewCustomActor(
			delegate,
			/*enableSocialProtocol=*/ false,
			/*enableFederatedProtocol=*/ true,
			clock)
		return
	}
	// Run tests
	t.Run("PostInboxIgnoresNonActivityPubRequest", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		_, _, a := setupFn(ctl)
		resp := httptest.NewRecorder()
		req := toPostInboxRequest(testCreate)
		// Run the test
		handled, err := a.PostInbox(ctx, resp, req)
		// Verify results
		assertEqual(t, err, nil)
		assertEqual(t, handled, false)
		assertEqual(t, len(resp.Result().Header), 0)
	})
	t.Run("PostInboxDeniesIfNotAuthenticated", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		delegate, _, a := setupFn(ctl)
		resp := httptest.NewRecorder()
		req := toAPRequest(toPostInboxRequest(testCreate))
		delegate.EXPECT().AuthenticatePostInbox(ctx, resp, req).DoAndReturn(func(ctx context.Context, resp http.ResponseWriter, req *http.Request) (context.Context, bool, error) {
			resp.WriteHeader(http.StatusForbidden)
			return ctx, false, nil
		})
		// Run the test
		handled, err := a.PostInbox(ctx, resp, req)
		// Verify results
		assertEqual(t, err, nil)
		assertEqual(t, handled, true)
		assertEqual(t, resp.Code, http.StatusForbidden)
	})
	t.Run("PostInboxBadRequestIfUnknownType", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		delegate, _, a := setupFn(ctl)
		resp := httptest.NewRecorder()
		req := toAPRequest(toPostInboxUnknownRequest())
		delegate.EXPECT().AuthenticatePostInbox(ctx, resp, req).Return(ctx, true, nil)
		// Run the test
		handled, err := a.PostInbox(ctx, resp, req)
		// Verify results
		assertEqual(t, err, nil)
		assertEqual(t, handled, true)
		assertEqual(t, resp.Code, http.StatusBadRequest)
	})
	t.Run("PostInboxBadRequestIfActivityHasNoId", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		delegate, _, a := setupFn(ctl)
		resp := httptest.NewRecorder()
		req := toAPRequest(toPostOutboxRequest(testCreateNoId))
		delegate.EXPECT().AuthenticatePostInbox(ctx, resp, req).Return(ctx, true, nil)
		// Run the test
		handled, err := a.PostInbox(ctx, resp, req)
		// Verify results
		assertEqual(t, err, nil)
		assertEqual(t, handled, true)
		assertEqual(t, resp.Code, http.StatusBadRequest)
	})
	t.Run("PostInboxDeniesIfNotAuthorized", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		delegate, _, a := setupFn(ctl)
		resp := httptest.NewRecorder()
		req := toAPRequest(toPostInboxRequest(testCreate))
		delegate.EXPECT().AuthenticatePostInbox(ctx, resp, req).Return(ctx, true, nil)
		delegate.EXPECT().PostInboxRequestBodyHook(ctx, req, toDeserializedForm(testCreate)).Return(ctx, nil)
		delegate.EXPECT().AuthorizePostInbox(ctx, resp, toDeserializedForm(testCreate)).DoAndReturn(func(ctx context.Context, resp http.ResponseWriter, activity Activity) (bool, error) {
			resp.WriteHeader(http.StatusForbidden)
			return false, nil
		})
		// Run the test
		handled, err := a.PostInbox(ctx, resp, req)
		// Verify results
		assertEqual(t, err, nil)
		assertEqual(t, handled, true)
		assertEqual(t, resp.Code, http.StatusForbidden)
	})
	t.Run("PostInboxRespondsWithStatus", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		delegate, _, a := setupFn(ctl)
		resp := httptest.NewRecorder()
		req := toAPRequest(toPostInboxRequest(testCreate))
		delegate.EXPECT().AuthenticatePostInbox(ctx, resp, req).Return(ctx, true, nil)
		delegate.EXPECT().PostInboxRequestBodyHook(ctx, req, toDeserializedForm(testCreate)).Return(ctx, nil)
		delegate.EXPECT().AuthorizePostInbox(ctx, resp, toDeserializedForm(testCreate)).Return(true, nil)
		delegate.EXPECT().PostInbox(ctx, mustParse(testMyInboxIRI), toDeserializedForm(testCreate)).Return(nil)
		delegate.EXPECT().InboxForwarding(ctx, mustParse(testMyInboxIRI), toDeserializedForm(testCreate)).Return(nil)
		// Run the test
		handled, err := a.PostInbox(ctx, resp, req)
		// Verify results
		assertEqual(t, err, nil)
		assertEqual(t, handled, true)
		assertEqual(t, resp.Code, http.StatusOK)
	})
	t.Run("PostInboxBadRequestForErrObjectRequired", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		delegate, _, a := setupFn(ctl)
		resp := httptest.NewRecorder()
		req := toAPRequest(toPostInboxRequest(testCreate))
		delegate.EXPECT().AuthenticatePostInbox(ctx, resp, req).Return(ctx, true, nil)
		delegate.EXPECT().PostInboxRequestBodyHook(ctx, req, toDeserializedForm(testCreate)).Return(ctx, nil)
		delegate.EXPECT().AuthorizePostInbox(ctx, resp, toDeserializedForm(testCreate)).Return(true, nil)
		delegate.EXPECT().PostInbox(ctx, mustParse(testMyInboxIRI), toDeserializedForm(testCreate)).Return(ErrObjectRequired)
		// Run the test
		handled, err := a.PostInbox(ctx, resp, req)
		// Verify results
		assertEqual(t, err, nil)
		assertEqual(t, handled, true)
		assertEqual(t, resp.Code, http.StatusBadRequest)
	})
	t.Run("PostInboxBadRequestForErrTargetRequired", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		delegate, _, a := setupFn(ctl)
		resp := httptest.NewRecorder()
		req := toAPRequest(toPostInboxRequest(testCreate))
		delegate.EXPECT().AuthenticatePostInbox(ctx, resp, req).Return(ctx, true, nil)
		delegate.EXPECT().PostInboxRequestBodyHook(ctx, req, toDeserializedForm(testCreate)).Return(ctx, nil)
		delegate.EXPECT().AuthorizePostInbox(ctx, resp, toDeserializedForm(testCreate)).Return(true, nil)
		delegate.EXPECT().PostInbox(ctx, mustParse(testMyInboxIRI), toDeserializedForm(testCreate)).Return(ErrTargetRequired)
		// Run the test
		handled, err := a.PostInbox(ctx, resp, req)
		// Verify results
		assertEqual(t, err, nil)
		assertEqual(t, handled, true)
		assertEqual(t, resp.Code, http.StatusBadRequest)
	})
	t.Run("GetInboxIgnoresNonActivityPubRequest", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		_, _, a := setupFn(ctl)
		resp := httptest.NewRecorder()
		req := toGetInboxRequest()
		// Run the test
		handled, err := a.GetInbox(ctx, resp, req)
		// Verify results
		assertEqual(t, err, nil)
		assertEqual(t, handled, false)
		assertEqual(t, len(resp.Result().Header), 0)
	})
	t.Run("GetInboxDeniesIfNotAuthenticated", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		delegate, _, a := setupFn(ctl)
		resp := httptest.NewRecorder()
		req := toAPRequest(toGetInboxRequest())
		delegate.EXPECT().AuthenticateGetInbox(ctx, resp, req).DoAndReturn(func(ctx context.Context, resp http.ResponseWriter, req *http.Request) (context.Context, bool, error) {
			resp.WriteHeader(http.StatusForbidden)
			return ctx, false, nil
		})
		// Run the test
		handled, err := a.GetInbox(ctx, resp, req)
		// Verify results
		assertEqual(t, err, nil)
		assertEqual(t, handled, true)
		assertEqual(t, resp.Code, http.StatusForbidden)
	})
	t.Run("GetInboxRespondsWithDataAndHeaders", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		delegate, clock, a := setupFn(ctl)
		resp := httptest.NewRecorder()
		req := toAPRequest(toGetInboxRequest())
		delegate.EXPECT().AuthenticateGetInbox(ctx, resp, req).Return(ctx, true, nil)
		delegate.EXPECT().GetInbox(ctx, req).Return(testOrderedCollectionUniqueElems, nil)
		clock.EXPECT().Now().Return(now())
		// Run the test
		handled, err := a.GetInbox(ctx, resp, req)
		// Verify results
		assertEqual(t, err, nil)
		assertEqual(t, handled, true)
		assertEqual(t, resp.Code, http.StatusOK)
		respV := resp.Result()
		assertEqual(t, respV.Header.Get(contentTypeHeader), "application/ld+json; profile=\"https://www.w3.org/ns/activitystreams\"")
		assertEqual(t, respV.Header.Get(dateHeader), nowDateHeader())
		assertNotEqual(t, len(respV.Header.Get(digestHeader)), 0)
		b, err := ioutil.ReadAll(respV.Body)
		assertEqual(t, err, nil)
		assertByteEqual(t, b, []byte(testOrderedCollectionUniqueElemsString))
	})
	t.Run("GetInboxDeduplicatesData", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		delegate, clock, a := setupFn(ctl)
		resp := httptest.NewRecorder()
		req := toAPRequest(toGetInboxRequest())
		delegate.EXPECT().AuthenticateGetInbox(ctx, resp, req).Return(ctx, true, nil)
		delegate.EXPECT().GetInbox(ctx, req).Return(testOrderedCollectionDupedElems, nil)
		clock.EXPECT().Now().Return(now())
		// Run the test
		_, err := a.GetInbox(ctx, resp, req)
		// Verify results
		assertEqual(t, err, nil)
		respV := resp.Result()
		b, err := ioutil.ReadAll(respV.Body)
		assertEqual(t, err, nil)
		assertByteEqual(t, b, []byte(testOrderedCollectionDedupedElemsString))
	})
	t.Run("PostOutboxIgnoresNonActivityPubRequest", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		_, _, a := setupFn(ctl)
		resp := httptest.NewRecorder()
		req := toPostOutboxRequest(testCreateNoId)
		// Run the test
		handled, err := a.PostOutbox(ctx, resp, req)
		// Verify results
		assertEqual(t, err, nil)
		assertEqual(t, handled, false)
		assertEqual(t, len(resp.Result().Header), 0)
	})
	t.Run("PostOutboxNotAllowed", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		_, _, a := setupFn(ctl)
		resp := httptest.NewRecorder()
		req := toAPRequest(toPostOutboxRequest(testCreateNoId))
		// Run the test
		handled, err := a.PostOutbox(ctx, resp, req)
		// Verify results
		assertEqual(t, err, nil)
		assertEqual(t, handled, true)
		assertEqual(t, resp.Code, http.StatusMethodNotAllowed)
	})
	t.Run("GetOutboxIgnoresNonActivityPubRequest", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		_, _, a := setupFn(ctl)
		resp := httptest.NewRecorder()
		req := toGetOutboxRequest()
		// Run the test
		handled, err := a.GetOutbox(ctx, resp, req)
		// Verify results
		assertEqual(t, err, nil)
		assertEqual(t, handled, false)
		assertEqual(t, len(resp.Result().Header), 0)
	})
	t.Run("GetOutboxDeniesIfNotAuthenticated", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		delegate, _, a := setupFn(ctl)
		resp := httptest.NewRecorder()
		req := toAPRequest(toGetOutboxRequest())
		delegate.EXPECT().AuthenticateGetOutbox(ctx, resp, req).DoAndReturn(func(ctx context.Context, resp http.ResponseWriter, req *http.Request) (context.Context, bool, error) {
			resp.WriteHeader(http.StatusForbidden)
			return ctx, false, nil
		})
		// Run the test
		handled, err := a.GetOutbox(ctx, resp, req)
		// Verify results
		assertEqual(t, err, nil)
		assertEqual(t, handled, true)
		assertEqual(t, resp.Code, http.StatusForbidden)
	})
	t.Run("GetOutboxRespondsWithDataAndHeaders", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		delegate, clock, a := setupFn(ctl)
		resp := httptest.NewRecorder()
		req := toAPRequest(toGetOutboxRequest())
		delegate.EXPECT().AuthenticateGetOutbox(ctx, resp, req).Return(ctx, true, nil)
		delegate.EXPECT().GetOutbox(ctx, req).Return(testOrderedCollectionUniqueElems, nil)
		clock.EXPECT().Now().Return(now())
		// Run the test
		handled, err := a.GetOutbox(ctx, resp, req)
		// Verify results
		assertEqual(t, err, nil)
		assertEqual(t, handled, true)
		assertEqual(t, resp.Code, http.StatusOK)
		respV := resp.Result()
		assertEqual(t, respV.Header.Get(contentTypeHeader), "application/ld+json; profile=\"https://www.w3.org/ns/activitystreams\"")
		assertEqual(t, respV.Header.Get(dateHeader), nowDateHeader())
		assertNotEqual(t, len(respV.Header.Get(digestHeader)), 0)
		b, err := ioutil.ReadAll(respV.Body)
		assertEqual(t, err, nil)
		assertByteEqual(t, b, []byte(testOrderedCollectionUniqueElemsString))
	})
}

// TestBaseActor tests the Actor returned with NewCustomActor and having both
// the SocialProtocol and FederatingProtocol enabled.
func TestBaseActor(t *testing.T) {
	// Set up test case
	setupData()
	ctx := context.Background()
	setupFn := func(ctl *gomock.Controller) (delegate *MockDelegateActor, clock *MockClock, a Actor) {
		delegate = NewMockDelegateActor(ctl)
		clock = NewMockClock(ctl)
		a = NewCustomActor(
			delegate,
			/*enableSocialProtocol=*/ true,
			/*enableFederatedProtocol=*/ true,
			clock)
		return
	}
	// Run tests
	t.Run("PostInboxRespondsWithStatus", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		delegate, _, a := setupFn(ctl)
		resp := httptest.NewRecorder()
		req := toAPRequest(toPostInboxRequest(testCreate))
		delegate.EXPECT().AuthenticatePostInbox(ctx, resp, req).Return(ctx, true, nil)
		delegate.EXPECT().PostInboxRequestBodyHook(ctx, req, toDeserializedForm(testCreate)).Return(ctx, nil)
		delegate.EXPECT().AuthorizePostInbox(ctx, resp, toDeserializedForm(testCreate)).Return(true, nil)
		delegate.EXPECT().PostInbox(ctx, mustParse(testMyInboxIRI), toDeserializedForm(testCreate)).Return(nil)
		delegate.EXPECT().InboxForwarding(ctx, mustParse(testMyInboxIRI), toDeserializedForm(testCreate)).Return(nil)
		// Run the test
		handled, err := a.PostInbox(ctx, resp, req)
		// Verify results
		assertEqual(t, err, nil)
		assertEqual(t, handled, true)
		assertEqual(t, resp.Code, http.StatusOK)
	})
	t.Run("PostOutboxRespondsWithDataAndHeaders", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		delegate, _, a := setupFn(ctl)
		resp := httptest.NewRecorder()
		req := toAPRequest(toPostOutboxRequest(testCreateNoId))
		delegate.EXPECT().AuthenticatePostOutbox(ctx, resp, req).Return(ctx, true, nil)
		delegate.EXPECT().PostOutboxRequestBodyHook(ctx, req, toDeserializedForm(testCreateNoId)).Return(ctx, nil)
		delegate.EXPECT().AddNewIDs(ctx, toDeserializedForm(testCreateNoId)).DoAndReturn(func(c context.Context, activity Activity) error {
			withNewId(activity)
			return nil
		})
		delegate.EXPECT().PostOutbox(
			ctx,
			withNewId(toDeserializedForm(testCreateNoId)),
			mustParse(testMyOutboxIRI),
			mustSerialize(testCreateNoId),
		).Return(false, nil)
		// Run the test
		handled, err := a.PostOutbox(ctx, resp, req)
		// Verify results
		assertEqual(t, err, nil)
		assertEqual(t, handled, true)
		assertEqual(t, resp.Code, http.StatusCreated)
		respV := resp.Result()
		assertEqual(t, respV.Header.Get(locationHeader), testNewActivityIRI)
	})
	t.Run("PostOutboxFederates", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		delegate, _, a := setupFn(ctl)
		resp := httptest.NewRecorder()
		req := toAPRequest(toPostOutboxRequest(testCreateNoId))
		delegate.EXPECT().AuthenticatePostOutbox(ctx, resp, req).Return(ctx, true, nil)
		delegate.EXPECT().PostOutboxRequestBodyHook(ctx, req, toDeserializedForm(testCreateNoId)).Return(ctx, nil)
		delegate.EXPECT().AddNewIDs(ctx, toDeserializedForm(testCreateNoId)).DoAndReturn(func(c context.Context, activity Activity) error {
			withNewId(activity)
			return nil
		})
		delegate.EXPECT().PostOutbox(
			ctx,
			withNewId(toDeserializedForm(testCreateNoId)),
			mustParse(testMyOutboxIRI),
			mustSerialize(testCreateNoId),
		).Return(true, nil)
		delegate.EXPECT().Deliver(ctx, mustParse(testMyOutboxIRI), withNewId(toDeserializedForm(testCreateNoId))).Return(nil)
		// Run the test
		handled, err := a.PostOutbox(ctx, resp, req)
		// Verify results
		assertEqual(t, err, nil)
		assertEqual(t, handled, true)
		assertEqual(t, resp.Code, http.StatusCreated)
		respV := resp.Result()
		assertEqual(t, respV.Header.Get(locationHeader), testNewActivityIRI)
	})
}

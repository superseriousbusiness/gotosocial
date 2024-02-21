package pub

import (
	"context"
	"fmt"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/activity/streams/vocab"
)

// TestPassThroughMethods tests the methods that pass-through to other
// dependency-injected types.
func TestPassThroughMethods(t *testing.T) {
	ctx := context.Background()
	resp := httptest.NewRecorder()
	setupFn := func(ctl *gomock.Controller) (c *MockCommonBehavior, fp *MockFederatingProtocol, sp *MockSocialProtocol, db *MockDatabase, cl *MockClock, a DelegateActor) {
		setupData()
		c = NewMockCommonBehavior(ctl)
		fp = NewMockFederatingProtocol(ctl)
		sp = NewMockSocialProtocol(ctl)
		db = NewMockDatabase(ctl)
		cl = NewMockClock(ctl)
		a = &SideEffectActor{
			common: c,
			s2s:    fp,
			c2s:    sp,
			db:     db,
			clock:  cl,
		}
		return
	}
	// Run tests
	t.Run("AuthenticatePostInbox", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		_, fp, _, _, _, a := setupFn(ctl)
		req := toAPRequest(toPostInboxRequest(testCreate))
		fp.EXPECT().AuthenticatePostInbox(ctx, resp, req).Return(ctx, true, testErr)
		// Run
		_, b, err := a.AuthenticatePostInbox(ctx, resp, req)
		// Verify
		assertEqual(t, b, true)
		assertEqual(t, err, testErr)
	})
	t.Run("AuthenticateGetInbox", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		c, _, _, _, _, a := setupFn(ctl)
		req := toAPRequest(toGetInboxRequest())
		c.EXPECT().AuthenticateGetInbox(ctx, resp, req).Return(ctx, true, testErr)
		// Run
		_, b, err := a.AuthenticateGetInbox(ctx, resp, req)
		// Verify
		assertEqual(t, b, true)
		assertEqual(t, err, testErr)
	})
	t.Run("AuthenticatePostOutbox", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		_, _, sp, _, _, a := setupFn(ctl)
		req := toAPRequest(toPostOutboxRequest(testCreate))
		sp.EXPECT().AuthenticatePostOutbox(ctx, resp, req).Return(ctx, true, testErr)
		// Run
		_, b, err := a.AuthenticatePostOutbox(ctx, resp, req)
		// Verify
		assertEqual(t, b, true)
		assertEqual(t, err, testErr)
	})
	t.Run("AuthenticateGetOutbox", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		c, _, _, _, _, a := setupFn(ctl)
		req := toAPRequest(toGetOutboxRequest())
		c.EXPECT().AuthenticateGetOutbox(ctx, resp, req).Return(ctx, true, testErr)
		// Run
		_, b, err := a.AuthenticateGetOutbox(ctx, resp, req)
		// Verify
		assertEqual(t, b, true)
		assertEqual(t, err, testErr)
	})
	t.Run("GetOutbox", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		c, _, _, _, _, a := setupFn(ctl)
		req := toAPRequest(toGetOutboxRequest())
		c.EXPECT().GetOutbox(ctx, req).Return(testOrderedCollectionUniqueElems, testErr)
		// Run
		p, err := a.GetOutbox(ctx, req)
		// Verify
		assertEqual(t, p, testOrderedCollectionUniqueElems)
		assertEqual(t, err, testErr)
	})
	t.Run("GetInbox", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		_, fp, _, _, _, a := setupFn(ctl)
		req := toAPRequest(toGetInboxRequest())
		fp.EXPECT().GetInbox(ctx, req).Return(testOrderedCollectionUniqueElems, testErr)
		// Run
		p, err := a.GetInbox(ctx, req)
		// Verify
		assertEqual(t, p, testOrderedCollectionUniqueElems)
		assertEqual(t, err, testErr)
	})
}

// TestAuthorizePostInbox tests the Authorization for a federated message, which
// is only based on blocks.
func TestAuthorizePostInbox(t *testing.T) {
	ctx := context.Background()
	resp := httptest.NewRecorder()
	setupFn := func(ctl *gomock.Controller) (c *MockCommonBehavior, fp *MockFederatingProtocol, sp *MockSocialProtocol, db *MockDatabase, cl *MockClock, a DelegateActor) {
		setupData()
		c = NewMockCommonBehavior(ctl)
		fp = NewMockFederatingProtocol(ctl)
		sp = NewMockSocialProtocol(ctl)
		db = NewMockDatabase(ctl)
		cl = NewMockClock(ctl)
		a = &SideEffectActor{
			common: c,
			s2s:    fp,
			c2s:    sp,
			db:     db,
			clock:  cl,
		}
		return
	}
	// Run tests
	t.Run("ActorAuthorized", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		_, fp, _, _, _, a := setupFn(ctl)
		fp.EXPECT().Blocked(ctx, []*url.URL{mustParse(testFederatedActorIRI)}).Return(false, nil)
		// Run
		b, err := a.AuthorizePostInbox(ctx, resp, testCreate)
		// Verify
		assertEqual(t, b, true)
		assertEqual(t, err, nil)
	})
	t.Run("ActorNotAuthorized", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		_, fp, _, _, _, a := setupFn(ctl)
		fp.EXPECT().Blocked(ctx, []*url.URL{mustParse(testFederatedActorIRI)}).Return(true, nil)
		// Run
		b, err := a.AuthorizePostInbox(ctx, resp, testCreate)
		// Verify
		assertEqual(t, b, false)
		assertEqual(t, err, nil)
	})
	t.Run("AllActorsAuthorized", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		_, fp, _, _, _, a := setupFn(ctl)
		fp.EXPECT().Blocked(ctx, []*url.URL{mustParse(testFederatedActorIRI), mustParse(testFederatedActorIRI2)}).Return(false, nil)
		// Run
		b, err := a.AuthorizePostInbox(ctx, resp, testCreate2)
		// Verify
		assertEqual(t, b, true)
		assertEqual(t, err, nil)
	})
	t.Run("OneActorNotAuthorized", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		_, fp, _, _, _, a := setupFn(ctl)
		fp.EXPECT().Blocked(ctx, []*url.URL{mustParse(testFederatedActorIRI), mustParse(testFederatedActorIRI2)}).Return(true, nil)
		// Run
		b, err := a.AuthorizePostInbox(ctx, resp, testCreate2)
		// Verify
		assertEqual(t, b, false)
		assertEqual(t, err, nil)
	})
}

// TestPostInbox ensures that the main application side effects of receiving a
// federated message occur.
func TestPostInbox(t *testing.T) {
	ctx := context.Background()
	setupFn := func(ctl *gomock.Controller) (c *MockCommonBehavior, fp *MockFederatingProtocol, sp *MockSocialProtocol, db *MockDatabase, cl *MockClock, a DelegateActor) {
		setupData()
		c = NewMockCommonBehavior(ctl)
		fp = NewMockFederatingProtocol(ctl)
		sp = NewMockSocialProtocol(ctl)
		db = NewMockDatabase(ctl)
		cl = NewMockClock(ctl)
		a = &SideEffectActor{
			common: c,
			s2s:    fp,
			c2s:    sp,
			db:     db,
			clock:  cl,
		}
		return
	}
	// Run tests
	t.Run("AddsToEmptyInbox", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		_, fp, _, db, _, a := setupFn(ctl)
		inboxIRI := mustParse(testMyInboxIRI)
		gomock.InOrder(
			db.EXPECT().Lock(ctx, inboxIRI).Return(func() {}, nil),
			db.EXPECT().InboxContains(ctx, inboxIRI, mustParse(testFederatedActivityIRI)).Return(false, nil),
			db.EXPECT().GetInbox(ctx, inboxIRI).Return(testEmptyOrderedCollection, nil),
			db.EXPECT().SetInbox(ctx, testOrderedCollectionWithFederatedId).Return(nil),
		)
		fp.EXPECT().FederatingCallbacks(ctx).Return(FederatingWrappedCallbacks{}, nil, nil)
		fp.EXPECT().DefaultCallback(ctx, testListen).Return(nil)
		// Run
		err := a.PostInbox(ctx, inboxIRI, testListen)
		// Verify
		assertEqual(t, err, nil)
	})
	t.Run("DoesNotAddToInboxNorDoSideEffectsIfDuplicate", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		_, _, _, db, _, a := setupFn(ctl)
		inboxIRI := mustParse(testMyInboxIRI)
		gomock.InOrder(
			db.EXPECT().Lock(ctx, inboxIRI).Return(func() {}, nil),
			db.EXPECT().InboxContains(ctx, inboxIRI, mustParse(testFederatedActivityIRI)).Return(true, nil),
		)
		// Run
		err := a.PostInbox(ctx, inboxIRI, testListen)
		// Verify
		assertEqual(t, err, nil)
	})
	t.Run("AddsToInbox", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		_, fp, _, db, _, a := setupFn(ctl)
		inboxIRI := mustParse(testMyInboxIRI)
		gomock.InOrder(
			db.EXPECT().Lock(ctx, inboxIRI).Return(func() {}, nil),
			db.EXPECT().InboxContains(ctx, inboxIRI, mustParse(testFederatedActivityIRI)).Return(false, nil),
			db.EXPECT().GetInbox(ctx, inboxIRI).Return(testOrderedCollectionWithFederatedId2, nil),
			db.EXPECT().SetInbox(ctx, testOrderedCollectionWithBothFederatedIds).Return(nil),
		)
		fp.EXPECT().FederatingCallbacks(ctx).Return(FederatingWrappedCallbacks{}, nil, nil)
		fp.EXPECT().DefaultCallback(ctx, testListen).Return(nil)
		// Run
		err := a.PostInbox(ctx, inboxIRI, testListen)
		// Verify
		assertEqual(t, err, nil)
	})
	t.Run("ResolvesToCustomFunction", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		_, fp, _, db, _, a := setupFn(ctl)
		inboxIRI := mustParse(testMyInboxIRI)
		gomock.InOrder(
			db.EXPECT().Lock(ctx, inboxIRI).Return(func() {}, nil),
			db.EXPECT().InboxContains(ctx, inboxIRI, mustParse(testFederatedActivityIRI)).Return(false, nil),
			db.EXPECT().GetInbox(ctx, inboxIRI).Return(testEmptyOrderedCollection, nil),
			db.EXPECT().SetInbox(ctx, testOrderedCollectionWithFederatedId).Return(nil),
		)
		pass := false
		fp.EXPECT().FederatingCallbacks(ctx).Return(FederatingWrappedCallbacks{}, []interface{}{
			func(c context.Context, a vocab.ActivityStreamsListen) error {
				pass = true
				return nil
			},
		}, nil)
		// Run
		err := a.PostInbox(ctx, inboxIRI, testListen)
		// Verify
		assertEqual(t, err, nil)
		assertEqual(t, pass, true)
	})
	t.Run("ResolvesToOverriddenFunction", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		_, fp, _, db, _, a := setupFn(ctl)
		inboxIRI := mustParse(testMyInboxIRI)
		gomock.InOrder(
			db.EXPECT().Lock(ctx, inboxIRI).Return(func() {}, nil),
			db.EXPECT().InboxContains(ctx, inboxIRI, mustParse(testFederatedActivityIRI)).Return(false, nil),
			db.EXPECT().GetInbox(ctx, inboxIRI).Return(testEmptyOrderedCollection, nil),
			db.EXPECT().SetInbox(ctx, testOrderedCollectionWithFederatedId).Return(nil),
		)
		pass := false
		fp.EXPECT().FederatingCallbacks(ctx).Return(FederatingWrappedCallbacks{}, []interface{}{
			func(c context.Context, a vocab.ActivityStreamsCreate) error {
				pass = true
				return nil
			},
		}, nil)
		// Run
		err := a.PostInbox(ctx, inboxIRI, testCreate)
		// Verify
		assertEqual(t, err, nil)
		assertEqual(t, pass, true)
	})
	t.Run("ResolvesToDefaultFunction", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		_, fp, _, db, _, a := setupFn(ctl)
		inboxIRI := mustParse(testMyInboxIRI)
		gomock.InOrder(
			db.EXPECT().Lock(ctx, inboxIRI).Return(func() {}, nil),
			db.EXPECT().InboxContains(ctx, inboxIRI, mustParse(testFederatedActivityIRI)).Return(false, nil),
			db.EXPECT().GetInbox(ctx, inboxIRI).Return(testEmptyOrderedCollection, nil),
			db.EXPECT().SetInbox(ctx, testOrderedCollectionWithFederatedId).Return(nil),
		)
		pass := false
		fp.EXPECT().FederatingCallbacks(ctx).Return(FederatingWrappedCallbacks{
			Create: func(c context.Context, a vocab.ActivityStreamsCreate) error {
				pass = true
				return nil
			},
		}, nil, nil)
		db.EXPECT().Lock(ctx, mustParse(testNoteId1)).Return(func() {}, nil)
		db.EXPECT().Create(ctx, testFederatedNote)
		// Run
		err := a.PostInbox(ctx, inboxIRI, testCreate)
		// Verify
		assertEqual(t, err, nil)
		assertEqual(t, pass, true)
	})
}

// TestInboxForwarding ensures that the inbox forwarding logic is correct.
func TestInboxForwarding(t *testing.T) {
	ctx := context.Background()
	setupFn := func(ctl *gomock.Controller) (c *MockCommonBehavior, fp *MockFederatingProtocol, sp *MockSocialProtocol, db *MockDatabase, cl *MockClock, a DelegateActor) {
		setupData()
		c = NewMockCommonBehavior(ctl)
		fp = NewMockFederatingProtocol(ctl)
		sp = NewMockSocialProtocol(ctl)
		db = NewMockDatabase(ctl)
		cl = NewMockClock(ctl)
		a = &SideEffectActor{
			common: c,
			s2s:    fp,
			c2s:    sp,
			db:     db,
			clock:  cl,
		}
		return
	}
	t.Run("DoesNotForwardIfAlreadyExists", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		_, _, _, db, _, a := setupFn(ctl)
		gomock.InOrder(
			db.EXPECT().Lock(ctx, mustParse(testFederatedActivityIRI)).Return(func() {}, nil),
			db.EXPECT().Exists(ctx, mustParse(testFederatedActivityIRI)).Return(true, nil),
		)
		// Run
		err := a.InboxForwarding(ctx, mustParse(testMyInboxIRI), testListen)
		// Verify
		assertEqual(t, err, nil)
	})
	t.Run("DoesNotForwardIfToCollectionNotOwned", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		_, _, _, db, _, a := setupFn(ctl)
		input := addToIds(testListen)
		gomock.InOrder(
			db.EXPECT().Lock(ctx, mustParse(testFederatedActivityIRI)).Return(func() {}, nil),
			db.EXPECT().Exists(ctx, mustParse(testFederatedActivityIRI)).Return(false, nil),
			db.EXPECT().Create(ctx, input).Return(nil),
			db.EXPECT().Lock(ctx, mustParse(testToIRI)).Return(func() {}, nil),
			db.EXPECT().Owns(ctx, mustParse(testToIRI)).Return(false, nil),
			db.EXPECT().Lock(ctx, mustParse(testToIRI2)).Return(func() {}, nil),
			db.EXPECT().Owns(ctx, mustParse(testToIRI2)).Return(false, nil),
		)
		// Run
		err := a.InboxForwarding(ctx, mustParse(testMyInboxIRI), input)
		// Verify
		assertEqual(t, err, nil)
	})
	t.Run("DoesNotForwardIfCcCollectionNotOwned", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		_, _, _, db, _, a := setupFn(ctl)
		input := mustAddCcIds(testListen)
		gomock.InOrder(
			db.EXPECT().Lock(ctx, mustParse(testFederatedActivityIRI)).Return(func() {}, nil),
			db.EXPECT().Exists(ctx, mustParse(testFederatedActivityIRI)).Return(false, nil),
			db.EXPECT().Create(ctx, input).Return(nil),
			db.EXPECT().Lock(ctx, mustParse(testCcIRI)).Return(func() {}, nil),
			db.EXPECT().Owns(ctx, mustParse(testCcIRI)).Return(false, nil),
			db.EXPECT().Lock(ctx, mustParse(testCcIRI2)).Return(func() {}, nil),
			db.EXPECT().Owns(ctx, mustParse(testCcIRI2)).Return(false, nil),
		)
		// Run
		err := a.InboxForwarding(ctx, mustParse(testMyInboxIRI), input)
		// Verify
		assertEqual(t, err, nil)
	})
	t.Run("DoesNotForwardIfAudienceCollectionNotOwned", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		_, _, _, db, _, a := setupFn(ctl)
		input := mustAddAudienceIds(testListen)
		gomock.InOrder(
			db.EXPECT().Lock(ctx, mustParse(testFederatedActivityIRI)).Return(func() {}, nil),
			db.EXPECT().Exists(ctx, mustParse(testFederatedActivityIRI)).Return(false, nil),
			db.EXPECT().Create(ctx, input).Return(nil),
			db.EXPECT().Lock(ctx, mustParse(testAudienceIRI)).Return(func() {}, nil),
			db.EXPECT().Owns(ctx, mustParse(testAudienceIRI)).Return(false, nil),
			db.EXPECT().Lock(ctx, mustParse(testAudienceIRI2)).Return(func() {}, nil),
			db.EXPECT().Owns(ctx, mustParse(testAudienceIRI2)).Return(false, nil),
		)
		// Run
		err := a.InboxForwarding(ctx, mustParse(testMyInboxIRI), input)
		// Verify
		assertEqual(t, err, nil)
	})
	t.Run("DoesNotForwardIfToOwnedButNotCollection", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		_, _, _, db, _, a := setupFn(ctl)
		input := addToIds(testListen)
		gomock.InOrder(
			db.EXPECT().Lock(ctx, mustParse(testFederatedActivityIRI)).Return(func() {}, nil),
			db.EXPECT().Exists(ctx, mustParse(testFederatedActivityIRI)).Return(false, nil),
			db.EXPECT().Create(ctx, input).Return(nil),
			db.EXPECT().Lock(ctx, mustParse(testToIRI)).Return(func() {}, nil),
			db.EXPECT().Owns(ctx, mustParse(testToIRI)).Return(true, nil),
			db.EXPECT().Lock(ctx, mustParse(testToIRI2)).Return(func() {}, nil),
			db.EXPECT().Owns(ctx, mustParse(testToIRI2)).Return(true, nil),
			db.EXPECT().Lock(ctx, mustParse(testToIRI)).Return(func() {}, nil),
			db.EXPECT().Get(ctx, mustParse(testToIRI)).Return(testPerson, nil),
			db.EXPECT().Lock(ctx, mustParse(testToIRI2)).Return(func() {}, nil),
			db.EXPECT().Get(ctx, mustParse(testToIRI2)).Return(testService, nil),
		)
		// Run
		err := a.InboxForwarding(ctx, mustParse(testMyInboxIRI), input)
		// Verify
		assertEqual(t, err, nil)
	})
	t.Run("DoesNotForwardIfCcOwnedButNotCollection", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		_, _, _, db, _, a := setupFn(ctl)
		input := mustAddCcIds(testListen)
		gomock.InOrder(
			db.EXPECT().Lock(ctx, mustParse(testFederatedActivityIRI)).Return(func() {}, nil),
			db.EXPECT().Exists(ctx, mustParse(testFederatedActivityIRI)).Return(false, nil),
			db.EXPECT().Create(ctx, input).Return(nil),
			db.EXPECT().Lock(ctx, mustParse(testCcIRI)).Return(func() {}, nil),
			db.EXPECT().Owns(ctx, mustParse(testCcIRI)).Return(true, nil),
			db.EXPECT().Lock(ctx, mustParse(testCcIRI2)).Return(func() {}, nil),
			db.EXPECT().Owns(ctx, mustParse(testCcIRI2)).Return(true, nil),
			db.EXPECT().Lock(ctx, mustParse(testCcIRI)).Return(func() {}, nil),
			db.EXPECT().Get(ctx, mustParse(testCcIRI)).Return(testPerson, nil),
			db.EXPECT().Lock(ctx, mustParse(testCcIRI2)).Return(func() {}, nil),
			db.EXPECT().Get(ctx, mustParse(testCcIRI2)).Return(testService, nil),
		)
		// Run
		err := a.InboxForwarding(ctx, mustParse(testMyInboxIRI), input)
		// Verify
		assertEqual(t, err, nil)
	})
	t.Run("DoesNotForwardIfAudienceOwnedButNotCollection", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		_, _, _, db, _, a := setupFn(ctl)
		input := mustAddAudienceIds(testListen)
		gomock.InOrder(
			db.EXPECT().Lock(ctx, mustParse(testFederatedActivityIRI)).Return(func() {}, nil),
			db.EXPECT().Exists(ctx, mustParse(testFederatedActivityIRI)).Return(false, nil),
			db.EXPECT().Create(ctx, input).Return(nil),
			db.EXPECT().Lock(ctx, mustParse(testAudienceIRI)).Return(func() {}, nil),
			db.EXPECT().Owns(ctx, mustParse(testAudienceIRI)).Return(true, nil),
			db.EXPECT().Lock(ctx, mustParse(testAudienceIRI2)).Return(func() {}, nil),
			db.EXPECT().Owns(ctx, mustParse(testAudienceIRI2)).Return(true, nil),
			db.EXPECT().Lock(ctx, mustParse(testAudienceIRI)).Return(func() {}, nil),
			db.EXPECT().Get(ctx, mustParse(testAudienceIRI)).Return(testPerson, nil),
			db.EXPECT().Lock(ctx, mustParse(testAudienceIRI2)).Return(func() {}, nil),
			db.EXPECT().Get(ctx, mustParse(testAudienceIRI2)).Return(testService, nil),
		)
		// Run
		err := a.InboxForwarding(ctx, mustParse(testMyInboxIRI), input)
		// Verify
		assertEqual(t, err, nil)
	})
	t.Run("DoesNotForwardIfNoChainIsOwned", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		cm, fp, _, db, _, a := setupFn(ctl)
		input := mustAddTagIds(
			mustAddAudienceIds(testListen))
		mockTPortTag := NewMockTransport(ctl)
		mockTPortTag2 := NewMockTransport(ctl)
		gomock.InOrder(
			db.EXPECT().Lock(ctx, mustParse(testFederatedActivityIRI)).Return(func() {}, nil),
			db.EXPECT().Exists(ctx, mustParse(testFederatedActivityIRI)).Return(false, nil),
			db.EXPECT().Create(ctx, input).Return(nil),
			db.EXPECT().Lock(ctx, mustParse(testAudienceIRI)).Return(func() {}, nil),
			db.EXPECT().Owns(ctx, mustParse(testAudienceIRI)).Return(true, nil),
			db.EXPECT().Lock(ctx, mustParse(testAudienceIRI2)).Return(func() {}, nil),
			db.EXPECT().Owns(ctx, mustParse(testAudienceIRI2)).Return(true, nil),
			db.EXPECT().Lock(ctx, mustParse(testAudienceIRI)).Return(func() {}, nil),
			db.EXPECT().Get(ctx, mustParse(testAudienceIRI)).Return(testOrderedCollectionOfActors, nil),
			db.EXPECT().Lock(ctx, mustParse(testAudienceIRI2)).Return(func() {}, nil),
			db.EXPECT().Get(ctx, mustParse(testAudienceIRI2)).Return(testCollectionOfActors, nil),
			fp.EXPECT().MaxInboxForwardingRecursionDepth(ctx).Return(0),
			// hasInboxForwardingValues
			db.EXPECT().Lock(ctx, mustParse(testTagIRI)).Return(func() {}, nil),
			db.EXPECT().Owns(ctx, mustParse(testTagIRI)).Return(false, nil),
			db.EXPECT().Lock(ctx, mustParse(testTagIRI2)).Return(func() {}, nil),
			db.EXPECT().Owns(ctx, mustParse(testTagIRI2)).Return(false, nil),
			db.EXPECT().Lock(ctx, mustParse(testNoteId1)).Return(func() {}, nil),
			db.EXPECT().Owns(ctx, mustParse(testNoteId1)).Return(false, nil),
			cm.EXPECT().NewTransport(ctx, mustParse(testMyInboxIRI), goFedUserAgent()).Return(mockTPortTag, nil),
			mockTPortTag.EXPECT().Dereference(ctx, mustParse(testTagIRI)).Return(
				mustWrapInGETResponse(mustParse(testTagIRI), newObjectWithId(testTagIRI)), nil),
			cm.EXPECT().NewTransport(ctx, mustParse(testMyInboxIRI), goFedUserAgent()).Return(mockTPortTag2, nil),
			mockTPortTag2.EXPECT().Dereference(ctx, mustParse(testTagIRI2)).Return(
				mustWrapInGETResponse(mustParse(testTagIRI2), newObjectWithId(testTagIRI2)), nil),
		)
		// Run
		err := a.InboxForwarding(ctx, mustParse(testMyInboxIRI), input)
		// Verify
		assertEqual(t, err, nil)
	})
	t.Run("ForwardsToRecipients", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		cm, fp, _, db, _, a := setupFn(ctl)
		input := mustAddTagIds(
			mustAddAudienceIds(testListen))
		tPort := NewMockTransport(ctl)
		gomock.InOrder(
			db.EXPECT().Lock(ctx, mustParse(testFederatedActivityIRI)).Return(func() {}, nil),
			db.EXPECT().Exists(ctx, mustParse(testFederatedActivityIRI)).Return(false, nil),
			db.EXPECT().Create(ctx, input).Return(nil),
			db.EXPECT().Lock(ctx, mustParse(testAudienceIRI)).Return(func() {}, nil),
			db.EXPECT().Owns(ctx, mustParse(testAudienceIRI)).Return(true, nil),
			db.EXPECT().Lock(ctx, mustParse(testAudienceIRI2)).Return(func() {}, nil),
			db.EXPECT().Owns(ctx, mustParse(testAudienceIRI2)).Return(true, nil),
			db.EXPECT().Lock(ctx, mustParse(testAudienceIRI)).Return(func() {}, nil),
			db.EXPECT().Get(ctx, mustParse(testAudienceIRI)).Return(testOrderedCollectionOfActors, nil),
			db.EXPECT().Lock(ctx, mustParse(testAudienceIRI2)).Return(func() {}, nil),
			db.EXPECT().Get(ctx, mustParse(testAudienceIRI2)).Return(testCollectionOfActors, nil),
			fp.EXPECT().MaxInboxForwardingRecursionDepth(ctx).Return(0),
			// hasInboxForwardingValues
			db.EXPECT().Lock(ctx, mustParse(testTagIRI)).Return(func() {}, nil),
			db.EXPECT().Owns(ctx, mustParse(testTagIRI)).Return(true, nil),
			// after hasInboxForwardingValues
			fp.EXPECT().FilterForwarding(
				ctx,
				[]*url.URL{
					mustParse(testAudienceIRI),
					mustParse(testAudienceIRI2),
				},
				input,
			).Return(
				[]*url.URL{
					mustParse(testAudienceIRI),
				},
				nil,
			),
			// deliverToRecipients
			cm.EXPECT().NewTransport(ctx, mustParse(testMyInboxIRI), goFedUserAgent()).Return(tPort, nil),
			tPort.EXPECT().BatchDeliver(
				ctx,
				mustSerializeToBytes(input),
				[]*url.URL{
					mustParse(testFederatedActorIRI3),
					mustParse(testFederatedActorIRI4),
				},
			),
		)
		// Run
		err := a.InboxForwarding(ctx, mustParse(testMyInboxIRI), input)
		// Verify
		assertEqual(t, err, nil)
	})
	t.Run("ForwardsToRecipientsIfChainIsNested", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		cm, fp, _, db, _, a := setupFn(ctl)
		input := mustAddAudienceIds(testNestedInReplyTo)
		tPort := NewMockTransport(ctl)
		gomock.InOrder(
			db.EXPECT().Lock(ctx, mustParse(testFederatedActivityIRI)).Return(func() {}, nil),
			db.EXPECT().Exists(ctx, mustParse(testFederatedActivityIRI)).Return(false, nil),
			db.EXPECT().Create(ctx, input).Return(nil),
			db.EXPECT().Lock(ctx, mustParse(testAudienceIRI)).Return(func() {}, nil),
			db.EXPECT().Owns(ctx, mustParse(testAudienceIRI)).Return(true, nil),
			db.EXPECT().Lock(ctx, mustParse(testAudienceIRI2)).Return(func() {}, nil),
			db.EXPECT().Owns(ctx, mustParse(testAudienceIRI2)).Return(true, nil),
			db.EXPECT().Lock(ctx, mustParse(testAudienceIRI)).Return(func() {}, nil),
			db.EXPECT().Get(ctx, mustParse(testAudienceIRI)).Return(testOrderedCollectionOfActors, nil),
			db.EXPECT().Lock(ctx, mustParse(testAudienceIRI2)).Return(func() {}, nil),
			db.EXPECT().Get(ctx, mustParse(testAudienceIRI2)).Return(testCollectionOfActors, nil),
			fp.EXPECT().MaxInboxForwardingRecursionDepth(ctx).Return(0),
			// hasInboxForwardingValues
			db.EXPECT().Lock(ctx, mustParse(testNoteId1)).Return(func() {}, nil),
			db.EXPECT().Owns(ctx, mustParse(testNoteId1)).Return(false, nil),
			db.EXPECT().Lock(ctx, mustParse(inReplyToIRI)).Return(func() {}, nil),
			db.EXPECT().Owns(ctx, mustParse(inReplyToIRI)).Return(true, nil),
			// after hasInboxForwardingValues
			fp.EXPECT().FilterForwarding(
				ctx,
				[]*url.URL{
					mustParse(testAudienceIRI),
					mustParse(testAudienceIRI2),
				},
				input,
			).Return(
				[]*url.URL{
					mustParse(testAudienceIRI),
				},
				nil,
			),
			// deliverToRecipients
			cm.EXPECT().NewTransport(ctx, mustParse(testMyInboxIRI), goFedUserAgent()).Return(tPort, nil),
			tPort.EXPECT().BatchDeliver(
				ctx,
				mustSerializeToBytes(input),
				[]*url.URL{
					mustParse(testFederatedActorIRI3),
					mustParse(testFederatedActorIRI4),
				},
			),
		)
		// Run
		err := a.InboxForwarding(ctx, mustParse(testMyInboxIRI), input)
		// Verify
		assertEqual(t, err, nil)
	})
	t.Run("ForwardsToRecipientsAfterDereferencing", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		cm, fp, _, db, _, a := setupFn(ctl)
		input := mustAddTagIds(
			mustAddAudienceIds(testListen))
		tagTPort := NewMockTransport(ctl)
		tagTPort2 := NewMockTransport(ctl)
		tPort := NewMockTransport(ctl)
		gomock.InOrder(
			db.EXPECT().Lock(ctx, mustParse(testFederatedActivityIRI)).Return(func() {}, nil),
			db.EXPECT().Exists(ctx, mustParse(testFederatedActivityIRI)).Return(false, nil),
			db.EXPECT().Create(ctx, input).Return(nil),
			db.EXPECT().Lock(ctx, mustParse(testAudienceIRI)).Return(func() {}, nil),
			db.EXPECT().Owns(ctx, mustParse(testAudienceIRI)).Return(true, nil),
			db.EXPECT().Lock(ctx, mustParse(testAudienceIRI2)).Return(func() {}, nil),
			db.EXPECT().Owns(ctx, mustParse(testAudienceIRI2)).Return(true, nil),
			db.EXPECT().Lock(ctx, mustParse(testAudienceIRI)).Return(func() {}, nil),
			db.EXPECT().Get(ctx, mustParse(testAudienceIRI)).Return(testOrderedCollectionOfActors, nil),
			db.EXPECT().Lock(ctx, mustParse(testAudienceIRI2)).Return(func() {}, nil),
			db.EXPECT().Get(ctx, mustParse(testAudienceIRI2)).Return(testCollectionOfActors, nil),
			fp.EXPECT().MaxInboxForwardingRecursionDepth(ctx).Return(0),
			// hasInboxForwardingValues
			db.EXPECT().Lock(ctx, mustParse(testTagIRI)).Return(func() {}, nil),
			db.EXPECT().Owns(ctx, mustParse(testTagIRI)).Return(false, nil),
			db.EXPECT().Lock(ctx, mustParse(testTagIRI2)).Return(func() {}, nil),
			db.EXPECT().Owns(ctx, mustParse(testTagIRI2)).Return(false, nil),
			db.EXPECT().Lock(ctx, mustParse(testNoteId1)).Return(func() {}, nil),
			db.EXPECT().Owns(ctx, mustParse(testNoteId1)).Return(false, nil),
			cm.EXPECT().NewTransport(ctx, mustParse(testMyInboxIRI), goFedUserAgent()).Return(tagTPort, nil),
			tagTPort.EXPECT().Dereference(ctx, mustParse(testTagIRI)).Return(mustSerializeToBytes(mustAddInReplyToIds(newActivityWithId(testTagIRI))), nil),
			cm.EXPECT().NewTransport(ctx, mustParse(testMyInboxIRI), goFedUserAgent()).Return(tagTPort2, nil),
			tagTPort2.EXPECT().Dereference(ctx, mustParse(testTagIRI2)).Return(mustSerializeToBytes(newActivityWithId(testTagIRI2)), nil),
			db.EXPECT().Lock(ctx, mustParse(inReplyToIRI)).Return(func() {}, nil),
			db.EXPECT().Owns(ctx, mustParse(inReplyToIRI)).Return(true, nil),
			// after hasInboxForwardingValues
			fp.EXPECT().FilterForwarding(
				ctx,
				[]*url.URL{
					mustParse(testAudienceIRI),
					mustParse(testAudienceIRI2),
				},
				input,
			).Return(
				[]*url.URL{
					mustParse(testAudienceIRI),
				},
				nil,
			),
			// deliverToRecipients
			cm.EXPECT().NewTransport(ctx, mustParse(testMyInboxIRI), goFedUserAgent()).Return(tPort, nil),
			tPort.EXPECT().BatchDeliver(
				ctx,
				mustSerializeToBytes(input),
				[]*url.URL{
					mustParse(testFederatedActorIRI3),
					mustParse(testFederatedActorIRI4),
				},
			),
		)
		// Run
		err := a.InboxForwarding(ctx, mustParse(testMyInboxIRI), input)
		// Verify
		assertEqual(t, err, nil)
	})
	t.Run("DoesNotForwardIfChainIsNestedTooDeep", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		_, fp, _, db, _, a := setupFn(ctl)
		input := mustAddAudienceIds(testNestedInReplyTo)
		gomock.InOrder(
			db.EXPECT().Lock(ctx, mustParse(testFederatedActivityIRI)).Return(func() {}, nil),
			db.EXPECT().Exists(ctx, mustParse(testFederatedActivityIRI)).Return(false, nil),
			db.EXPECT().Create(ctx, input).Return(nil),
			db.EXPECT().Lock(ctx, mustParse(testAudienceIRI)).Return(func() {}, nil),
			db.EXPECT().Owns(ctx, mustParse(testAudienceIRI)).Return(true, nil),
			db.EXPECT().Lock(ctx, mustParse(testAudienceIRI2)).Return(func() {}, nil),
			db.EXPECT().Owns(ctx, mustParse(testAudienceIRI2)).Return(true, nil),
			db.EXPECT().Lock(ctx, mustParse(testAudienceIRI)).Return(func() {}, nil),
			db.EXPECT().Get(ctx, mustParse(testAudienceIRI)).Return(testOrderedCollectionOfActors, nil),
			db.EXPECT().Lock(ctx, mustParse(testAudienceIRI2)).Return(func() {}, nil),
			db.EXPECT().Get(ctx, mustParse(testAudienceIRI2)).Return(testCollectionOfActors, nil),
			fp.EXPECT().MaxInboxForwardingRecursionDepth(ctx).Return(1),
			// hasInboxForwardingValues
			db.EXPECT().Lock(ctx, mustParse(testNoteId1)).Return(func() {}, nil),
			db.EXPECT().Owns(ctx, mustParse(testNoteId1)).Return(false, nil),
		)
		// Run
		err := a.InboxForwarding(ctx, mustParse(testMyInboxIRI), input)
		// Verify
		assertEqual(t, err, nil)
	})
}

// TestPostOutbox ensures that the main application side effects of receiving a
// social protocol message occur.
func TestPostOutbox(t *testing.T) {
	ctx := context.Background()
	setupFn := func(ctl *gomock.Controller) (c *MockCommonBehavior, fp *MockFederatingProtocol, sp *MockSocialProtocol, db *MockDatabase, cl *MockClock, a DelegateActor) {
		setupData()
		c = NewMockCommonBehavior(ctl)
		fp = NewMockFederatingProtocol(ctl)
		sp = NewMockSocialProtocol(ctl)
		db = NewMockDatabase(ctl)
		cl = NewMockClock(ctl)
		a = &SideEffectActor{
			common: c,
			s2s:    fp,
			c2s:    sp,
			db:     db,
			clock:  cl,
		}
		return
	}
	t.Run("AddsToEmptyOutbox", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		_, _, sp, db, _, a := setupFn(ctl)
		outboxIRI := mustParse(testMyOutboxIRI)
		gomock.InOrder(
			db.EXPECT().Lock(ctx, mustParse(testNewActivityIRI)).Return(func() {}, nil),
			db.EXPECT().Create(ctx, testMyListen),
			db.EXPECT().Lock(ctx, outboxIRI).Return(func() {}, nil),
			db.EXPECT().GetOutbox(ctx, outboxIRI).Return(testEmptyOrderedCollection, nil),
			db.EXPECT().SetOutbox(ctx, testOrderedCollectionWithNewId).Return(nil),
		)
		sp.EXPECT().SocialCallbacks(ctx).Return(SocialWrappedCallbacks{}, nil, nil)
		sp.EXPECT().DefaultCallback(ctx, testMyListen).Return(nil)
		// Run
		deliverable, err := a.PostOutbox(ctx, testMyListen, outboxIRI, mustSerialize(testMyListen))
		// Verify
		assertEqual(t, err, nil)
		assertEqual(t, deliverable, true)
	})
	t.Run("AddsToOutbox", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		_, _, sp, db, _, a := setupFn(ctl)
		outboxIRI := mustParse(testMyOutboxIRI)
		gomock.InOrder(
			db.EXPECT().Lock(ctx, mustParse(testNewActivityIRI)).Return(func() {}, nil),
			db.EXPECT().Create(ctx, testMyListen),
			db.EXPECT().Lock(ctx, outboxIRI).Return(func() {}, nil),
			db.EXPECT().GetOutbox(ctx, outboxIRI).Return(testOrderedCollectionWithNewId2, nil),
			db.EXPECT().SetOutbox(ctx, testOrderedCollectionWithBothNewIds).Return(nil),
		)
		sp.EXPECT().SocialCallbacks(ctx).Return(SocialWrappedCallbacks{}, nil, nil)
		sp.EXPECT().DefaultCallback(ctx, testMyListen).Return(nil)
		// Run
		deliverable, err := a.PostOutbox(ctx, testMyListen, outboxIRI, mustSerialize(testMyListen))
		// Verify
		assertEqual(t, err, nil)
		assertEqual(t, deliverable, true)
	})
	t.Run("ResolvesToCustomFunction", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		_, _, sp, db, _, a := setupFn(ctl)
		outboxIRI := mustParse(testMyOutboxIRI)
		gomock.InOrder(
			db.EXPECT().Lock(ctx, mustParse(testNewActivityIRI)).Return(func() {}, nil),
			db.EXPECT().Create(ctx, testMyListen),
			db.EXPECT().Lock(ctx, outboxIRI).Return(func() {}, nil),
			db.EXPECT().GetOutbox(ctx, outboxIRI).Return(testEmptyOrderedCollection, nil),
			db.EXPECT().SetOutbox(ctx, testOrderedCollectionWithNewId).Return(nil),
		)
		pass := false
		sp.EXPECT().SocialCallbacks(ctx).Return(SocialWrappedCallbacks{}, []interface{}{
			func(c context.Context, a vocab.ActivityStreamsListen) error {
				pass = true
				return nil
			},
		}, nil)
		// Run
		deliverable, err := a.PostOutbox(ctx, testMyListen, outboxIRI, mustSerialize(testMyListen))
		// Verify
		assertEqual(t, err, nil)
		assertEqual(t, deliverable, true)
		assertEqual(t, pass, true)
	})
	t.Run("ResolvesToOverriddenFunction", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		_, _, sp, db, _, a := setupFn(ctl)
		outboxIRI := mustParse(testMyOutboxIRI)
		gomock.InOrder(
			db.EXPECT().Lock(ctx, mustParse(testNewActivityIRI)).Return(func() {}, nil),
			db.EXPECT().Create(ctx, testMyCreate),
			db.EXPECT().Lock(ctx, outboxIRI).Return(func() {}, nil),
			db.EXPECT().GetOutbox(ctx, outboxIRI).Return(testEmptyOrderedCollection, nil),
			db.EXPECT().SetOutbox(ctx, testOrderedCollectionWithNewId).Return(nil),
		)
		pass := false
		sp.EXPECT().SocialCallbacks(ctx).Return(SocialWrappedCallbacks{}, []interface{}{
			func(c context.Context, a vocab.ActivityStreamsCreate) error {
				pass = true
				return nil
			},
		}, nil)
		// Run
		deliverable, err := a.PostOutbox(ctx, testMyCreate, outboxIRI, mustSerialize(testMyCreate))
		// Verify
		assertEqual(t, err, nil)
		assertEqual(t, deliverable, true)
		assertEqual(t, pass, true)
	})
	t.Run("ResolvesToDefaultFunction", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		_, _, sp, db, _, a := setupFn(ctl)
		outboxIRI := mustParse(testMyInboxIRI)
		gomock.InOrder(
			db.EXPECT().Lock(ctx, mustParse(testNewActivityIRI)).Return(func() {}, nil),
			db.EXPECT().Create(ctx, testMyCreate),
			db.EXPECT().Lock(ctx, outboxIRI).Return(func() {}, nil),
			db.EXPECT().GetOutbox(ctx, outboxIRI).Return(testEmptyOrderedCollection, nil),
			db.EXPECT().SetOutbox(ctx, testOrderedCollectionWithNewId).Return(nil),
		)
		pass := false
		sp.EXPECT().SocialCallbacks(ctx).Return(SocialWrappedCallbacks{
			Create: func(c context.Context, a vocab.ActivityStreamsCreate) error {
				pass = true
				return nil
			},
		}, nil, nil)
		db.EXPECT().Lock(ctx, mustParse(testNoteId1)).Return(func() {}, nil)
		db.EXPECT().Create(ctx, testMyNote)
		// Run
		deliverable, err := a.PostOutbox(ctx, testMyCreate, outboxIRI, mustSerialize(testMyCreate))
		// Verify
		// Verify
		assertEqual(t, err, nil)
		assertEqual(t, deliverable, true)
		assertEqual(t, pass, true)
	})
}

// TestAddNewIDs ensures that new 'id' properties are set on an activity and all
// of its 'object' property values if it is a Create activity.
func TestAddNewIDs(t *testing.T) {
	ctx := context.Background()
	setupFn := func(ctl *gomock.Controller) (c *MockCommonBehavior, fp *MockFederatingProtocol, sp *MockSocialProtocol, db *MockDatabase, cl *MockClock, a DelegateActor) {
		setupData()
		c = NewMockCommonBehavior(ctl)
		fp = NewMockFederatingProtocol(ctl)
		sp = NewMockSocialProtocol(ctl)
		db = NewMockDatabase(ctl)
		cl = NewMockClock(ctl)
		a = &SideEffectActor{
			common: c,
			s2s:    fp,
			c2s:    sp,
			db:     db,
			clock:  cl,
		}
		return
	}
	t.Run("AddsIdToActivityWithoutId", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		_, _, _, db, _, a := setupFn(ctl)
		db.EXPECT().NewID(ctx, testMyListenNoId).Return(mustParse(testNewActivityIRI2), nil)
		// Run
		err := a.AddNewIDs(ctx, testMyListenNoId)
		// Verify
		assertEqual(t, err, nil)
		resultId := testMyListenNoId.GetJSONLDId()
		assertNotEqual(t, resultId, nil)
		assertEqual(t, resultId.Get().String(), mustParse(testNewActivityIRI2).String())
	})
	t.Run("AddsIdToActivityWithId", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		_, _, _, db, _, a := setupFn(ctl)
		db.EXPECT().NewID(ctx, testMyListen).Return(mustParse(testNewActivityIRI2), nil)
		// Run
		err := a.AddNewIDs(ctx, testMyListen)
		// Verify
		assertEqual(t, err, nil)
		resultId := testMyListen.GetJSONLDId()
		assertNotEqual(t, resultId, nil)
		assertEqual(t, resultId.Get().String(), mustParse(testNewActivityIRI2).String())
	})
	t.Run("AddsIdsToObjectsIfCreateActivity", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		_, _, _, db, _, a := setupFn(ctl)
		db.EXPECT().NewID(ctx, testMyCreate).Return(mustParse(testNewActivityIRI2), nil)
		db.EXPECT().NewID(ctx, testMyNote).Return(mustParse(testNewActivityIRI3), nil)
		// Run
		err := a.AddNewIDs(ctx, testMyCreate)
		// Verify
		assertEqual(t, err, nil)
		op := testMyCreate.GetActivityStreamsObject()
		assertNotEqual(t, op, nil)
		assertEqual(t, op.Len(), 1)
		n := op.At(0).GetActivityStreamsNote()
		assertNotEqual(t, n, nil)
		noteId := n.GetJSONLDId()
		assertNotEqual(t, noteId, nil)
		assertEqual(t, noteId.Get().String(), mustParse(testNewActivityIRI3).String())
	})
	t.Run("DoesNotAddIdsToObjectsIfNotCreateActivity", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		_, _, _, db, _, a := setupFn(ctl)
		db.EXPECT().NewID(ctx, testMyListenNoId).Return(mustParse(testNewActivityIRI2), nil)
		// Run
		err := a.AddNewIDs(ctx, testMyListenNoId)
		// Verify
		assertEqual(t, err, nil)
		op := testMyListenNoId.GetActivityStreamsObject()
		assertNotEqual(t, op, nil)
		assertEqual(t, op.Len(), 1)
		n := op.At(0).GetActivityStreamsNote()
		assertNotEqual(t, n, nil)
		noteId := n.GetJSONLDId()
		assertEqual(t, noteId, nil)
	})
}

// TestDeliver ensures federated delivery of an activity happens correctly to
// the ActivityPub specification.
func TestDeliver(t *testing.T) {
	baseActivityFn := func() vocab.ActivityStreamsCreate {
		act := streams.NewActivityStreamsCreate()
		id := streams.NewJSONLDIdProperty()
		id.Set(mustParse(testNewActivityIRI))
		act.SetJSONLDId(id)
		op := streams.NewActivityStreamsObjectProperty()
		note := streams.NewActivityStreamsNote()
		op.AppendActivityStreamsNote(note)
		act.SetActivityStreamsObject(op)
		return act
	}
	ctx := context.Background()
	setupFn := func(ctl *gomock.Controller) (c *MockCommonBehavior, fp *MockFederatingProtocol, sp *MockSocialProtocol, db *MockDatabase, cl *MockClock, a DelegateActor) {
		setupData()
		c = NewMockCommonBehavior(ctl)
		fp = NewMockFederatingProtocol(ctl)
		sp = NewMockSocialProtocol(ctl)
		db = NewMockDatabase(ctl)
		cl = NewMockClock(ctl)
		a = &SideEffectActor{
			common: c,
			s2s:    fp,
			c2s:    sp,
			db:     db,
			clock:  cl,
		}
		return
	}
	t.Run("SendToRecipientsInTo", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		c, mockFp, _, mockDb, _, a := setupFn(ctl)
		mockTp := NewMockTransport(ctl)
		act := baseActivityFn()
		to := streams.NewActivityStreamsToProperty()
		to.AppendIRI(mustParse(testFederatedActorIRI))
		to.AppendIRI(mustParse(testFederatedActorIRI2))
		act.SetActivityStreamsTo(to)
		expectRecip := []*url.URL{
			mustParse(testFederatedInboxIRI),
			mustParse(testFederatedInboxIRI2),
		}
		// Mock
		c.EXPECT().NewTransport(ctx, mustParse(testMyOutboxIRI), goFedUserAgent()).Return(
			mockTp, nil)
		mockFp.EXPECT().MaxDeliveryRecursionDepth(ctx).Return(1)
		mockTp.EXPECT().Dereference(ctx, mustParse(testFederatedActorIRI)).Return(
			mustSerializeToBytes(testFederatedPerson1), nil)
		mockTp.EXPECT().Dereference(ctx, mustParse(testFederatedActorIRI2)).Return(
			mustSerializeToBytes(testFederatedPerson2), nil)
		mockDb.EXPECT().Lock(ctx, mustParse(testMyOutboxIRI)).Return(func() {}, nil)
		mockDb.EXPECT().ActorForOutbox(ctx, mustParse(testMyOutboxIRI)).Return(
			mustParse(testPersonIRI), nil)
		mockDb.EXPECT().Lock(ctx, mustParse(testPersonIRI)).Return(func() {}, nil)
		mockDb.EXPECT().Get(ctx, mustParse(testPersonIRI)).Return(
			testMyPerson, nil)
		c.EXPECT().NewTransport(ctx, mustParse(testMyOutboxIRI), goFedUserAgent()).Return(
			mockTp, nil)
		mockTp.EXPECT().BatchDeliver(ctx, mustSerializeToBytes(act), expectRecip)
		// Run & Verify
		err := a.Deliver(ctx, mustParse(testMyOutboxIRI), act)
		assertEqual(t, err, nil)
	})
	t.Run("SendToRecipientsInBto", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		c, mockFp, _, mockDb, _, a := setupFn(ctl)
		mockTp := NewMockTransport(ctl)
		act := baseActivityFn()
		bto := streams.NewActivityStreamsBtoProperty()
		bto.AppendIRI(mustParse(testFederatedActorIRI))
		bto.AppendIRI(mustParse(testFederatedActorIRI2))
		act.SetActivityStreamsBto(bto)
		expectAct := baseActivityFn() // Ensure Bto is stripped
		expectRecip := []*url.URL{
			mustParse(testFederatedInboxIRI),
			mustParse(testFederatedInboxIRI2),
		}
		// Mock
		c.EXPECT().NewTransport(ctx, mustParse(testMyOutboxIRI), goFedUserAgent()).Return(
			mockTp, nil)
		mockFp.EXPECT().MaxDeliveryRecursionDepth(ctx).Return(1)
		mockTp.EXPECT().Dereference(ctx, mustParse(testFederatedActorIRI)).Return(
			mustSerializeToBytes(testFederatedPerson1), nil)
		mockTp.EXPECT().Dereference(ctx, mustParse(testFederatedActorIRI2)).Return(
			mustSerializeToBytes(testFederatedPerson2), nil)
		mockDb.EXPECT().Lock(ctx, mustParse(testMyOutboxIRI)).Return(func() {}, nil)
		mockDb.EXPECT().ActorForOutbox(ctx, mustParse(testMyOutboxIRI)).Return(
			mustParse(testPersonIRI), nil)
		mockDb.EXPECT().Lock(ctx, mustParse(testPersonIRI)).Return(func() {}, nil)
		mockDb.EXPECT().Get(ctx, mustParse(testPersonIRI)).Return(
			testMyPerson, nil)
		c.EXPECT().NewTransport(ctx, mustParse(testMyOutboxIRI), goFedUserAgent()).Return(
			mockTp, nil)
		mockTp.EXPECT().BatchDeliver(ctx, mustSerializeToBytes(expectAct), expectRecip)
		// Run & Verify
		err := a.Deliver(ctx, mustParse(testMyOutboxIRI), act)
		assertEqual(t, err, nil)
	})
	t.Run("SendToRecipientsInCc", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		c, mockFp, _, mockDb, _, a := setupFn(ctl)
		mockTp := NewMockTransport(ctl)
		act := baseActivityFn()
		cc := streams.NewActivityStreamsCcProperty()
		cc.AppendIRI(mustParse(testFederatedActorIRI))
		cc.AppendIRI(mustParse(testFederatedActorIRI2))
		act.SetActivityStreamsCc(cc)
		expectRecip := []*url.URL{
			mustParse(testFederatedInboxIRI),
			mustParse(testFederatedInboxIRI2),
		}
		// Mock
		c.EXPECT().NewTransport(ctx, mustParse(testMyOutboxIRI), goFedUserAgent()).Return(
			mockTp, nil)
		mockFp.EXPECT().MaxDeliveryRecursionDepth(ctx).Return(1)
		mockTp.EXPECT().Dereference(ctx, mustParse(testFederatedActorIRI)).Return(
			mustSerializeToBytes(testFederatedPerson1), nil)
		mockTp.EXPECT().Dereference(ctx, mustParse(testFederatedActorIRI2)).Return(
			mustSerializeToBytes(testFederatedPerson2), nil)
		mockDb.EXPECT().Lock(ctx, mustParse(testMyOutboxIRI)).Return(func() {}, nil)
		mockDb.EXPECT().ActorForOutbox(ctx, mustParse(testMyOutboxIRI)).Return(
			mustParse(testPersonIRI), nil)
		mockDb.EXPECT().Lock(ctx, mustParse(testPersonIRI)).Return(func() {}, nil)
		mockDb.EXPECT().Get(ctx, mustParse(testPersonIRI)).Return(
			testMyPerson, nil)
		c.EXPECT().NewTransport(ctx, mustParse(testMyOutboxIRI), goFedUserAgent()).Return(
			mockTp, nil)
		mockTp.EXPECT().BatchDeliver(ctx, mustSerializeToBytes(act), expectRecip)
		// Run & Verify
		err := a.Deliver(ctx, mustParse(testMyOutboxIRI), act)
		assertEqual(t, err, nil)
	})
	t.Run("SendToRecipientsInBcc", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		c, mockFp, _, mockDb, _, a := setupFn(ctl)
		mockTp := NewMockTransport(ctl)
		act := baseActivityFn()
		bcc := streams.NewActivityStreamsBccProperty()
		bcc.AppendIRI(mustParse(testFederatedActorIRI))
		bcc.AppendIRI(mustParse(testFederatedActorIRI2))
		act.SetActivityStreamsBcc(bcc)
		expectAct := baseActivityFn() // Ensure Bcc is stripped
		expectRecip := []*url.URL{
			mustParse(testFederatedInboxIRI),
			mustParse(testFederatedInboxIRI2),
		}
		// Mock
		c.EXPECT().NewTransport(ctx, mustParse(testMyOutboxIRI), goFedUserAgent()).Return(
			mockTp, nil)
		mockFp.EXPECT().MaxDeliveryRecursionDepth(ctx).Return(1)
		mockTp.EXPECT().Dereference(ctx, mustParse(testFederatedActorIRI)).Return(
			mustSerializeToBytes(testFederatedPerson1), nil)
		mockTp.EXPECT().Dereference(ctx, mustParse(testFederatedActorIRI2)).Return(
			mustSerializeToBytes(testFederatedPerson2), nil)
		mockDb.EXPECT().Lock(ctx, mustParse(testMyOutboxIRI)).Return(func() {}, nil)
		mockDb.EXPECT().ActorForOutbox(ctx, mustParse(testMyOutboxIRI)).Return(
			mustParse(testPersonIRI), nil)
		mockDb.EXPECT().Lock(ctx, mustParse(testPersonIRI)).Return(func() {}, nil)
		mockDb.EXPECT().Get(ctx, mustParse(testPersonIRI)).Return(
			testMyPerson, nil)
		c.EXPECT().NewTransport(ctx, mustParse(testMyOutboxIRI), goFedUserAgent()).Return(
			mockTp, nil)
		mockTp.EXPECT().BatchDeliver(ctx, mustSerializeToBytes(expectAct), expectRecip)
		// Run & Verify
		err := a.Deliver(ctx, mustParse(testMyOutboxIRI), act)
		assertEqual(t, err, nil)
	})
	t.Run("SendToRecipientsInAudience", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		c, mockFp, _, mockDb, _, a := setupFn(ctl)
		mockTp := NewMockTransport(ctl)
		act := baseActivityFn()
		aud := streams.NewActivityStreamsAudienceProperty()
		aud.AppendIRI(mustParse(testFederatedActorIRI))
		aud.AppendIRI(mustParse(testFederatedActorIRI2))
		act.SetActivityStreamsAudience(aud)
		expectRecip := []*url.URL{
			mustParse(testFederatedInboxIRI),
			mustParse(testFederatedInboxIRI2),
		}
		// Mock
		c.EXPECT().NewTransport(ctx, mustParse(testMyOutboxIRI), goFedUserAgent()).Return(
			mockTp, nil)
		mockFp.EXPECT().MaxDeliveryRecursionDepth(ctx).Return(1)
		mockTp.EXPECT().Dereference(ctx, mustParse(testFederatedActorIRI)).Return(
			mustSerializeToBytes(testFederatedPerson1), nil)
		mockTp.EXPECT().Dereference(ctx, mustParse(testFederatedActorIRI2)).Return(
			mustSerializeToBytes(testFederatedPerson2), nil)
		mockDb.EXPECT().Lock(ctx, mustParse(testMyOutboxIRI)).Return(func() {}, nil)
		mockDb.EXPECT().ActorForOutbox(ctx, mustParse(testMyOutboxIRI)).Return(
			mustParse(testPersonIRI), nil)
		mockDb.EXPECT().Lock(ctx, mustParse(testPersonIRI)).Return(func() {}, nil)
		mockDb.EXPECT().Get(ctx, mustParse(testPersonIRI)).Return(
			testMyPerson, nil)
		c.EXPECT().NewTransport(ctx, mustParse(testMyOutboxIRI), goFedUserAgent()).Return(
			mockTp, nil)
		mockTp.EXPECT().BatchDeliver(ctx, mustSerializeToBytes(act), expectRecip)
		// Run & Verify
		err := a.Deliver(ctx, mustParse(testMyOutboxIRI), act)
		assertEqual(t, err, nil)
	})
	t.Run("DoesNotSendToPublicIRI", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		c, mockFp, _, mockDb, _, a := setupFn(ctl)
		mockTp := NewMockTransport(ctl)
		act := baseActivityFn()
		to := streams.NewActivityStreamsToProperty()
		to.AppendIRI(mustParse(testFederatedActorIRI))
		to.AppendIRI(mustParse(testFederatedActorIRI2))
		to.AppendIRI(mustParse(PublicActivityPubIRI))
		act.SetActivityStreamsTo(to)
		expectRecip := []*url.URL{
			mustParse(testFederatedInboxIRI),
			mustParse(testFederatedInboxIRI2),
		}
		// Mock
		c.EXPECT().NewTransport(ctx, mustParse(testMyOutboxIRI), goFedUserAgent()).Return(
			mockTp, nil)
		mockFp.EXPECT().MaxDeliveryRecursionDepth(ctx).Return(1)
		mockTp.EXPECT().Dereference(ctx, mustParse(testFederatedActorIRI)).Return(
			mustSerializeToBytes(testFederatedPerson1), nil)
		mockTp.EXPECT().Dereference(ctx, mustParse(testFederatedActorIRI2)).Return(
			mustSerializeToBytes(testFederatedPerson2), nil)
		mockDb.EXPECT().Lock(ctx, mustParse(testMyOutboxIRI)).Return(func() {}, nil)
		mockDb.EXPECT().ActorForOutbox(ctx, mustParse(testMyOutboxIRI)).Return(
			mustParse(testPersonIRI), nil)
		mockDb.EXPECT().Lock(ctx, mustParse(testPersonIRI)).Return(func() {}, nil)
		mockDb.EXPECT().Get(ctx, mustParse(testPersonIRI)).Return(
			testMyPerson, nil)
		c.EXPECT().NewTransport(ctx, mustParse(testMyOutboxIRI), goFedUserAgent()).Return(
			mockTp, nil)
		mockTp.EXPECT().BatchDeliver(ctx, mustSerializeToBytes(act), expectRecip)
		// Run & Verify
		err := a.Deliver(ctx, mustParse(testMyOutboxIRI), act)
		assertEqual(t, err, nil)
	})
	t.Run("RecursivelyResolveCollectionActors", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		c, mockFp, _, mockDb, _, a := setupFn(ctl)
		mockTp := NewMockTransport(ctl)
		act := baseActivityFn()
		to := streams.NewActivityStreamsToProperty()
		to.AppendIRI(mustParse(testAudienceIRI))
		act.SetActivityStreamsTo(to)
		expectRecip := []*url.URL{
			mustParse(testFederatedInboxIRI),
			mustParse(testFederatedInboxIRI2),
		}
		// Mock
		c.EXPECT().NewTransport(ctx, mustParse(testMyOutboxIRI), goFedUserAgent()).Return(
			mockTp, nil)
		mockFp.EXPECT().MaxDeliveryRecursionDepth(ctx).Return(2)
		mockTp.EXPECT().Dereference(ctx, mustParse(testAudienceIRI)).Return(
			mustSerializeToBytes(testCollectionOfActors), nil)
		mockTp.EXPECT().Dereference(ctx, mustParse(testFederatedActorIRI)).Return(
			mustSerializeToBytes(testFederatedPerson1), nil)
		mockTp.EXPECT().Dereference(ctx, mustParse(testFederatedActorIRI2)).Return(
			mustSerializeToBytes(testFederatedPerson2), nil)
		mockDb.EXPECT().Lock(ctx, mustParse(testMyOutboxIRI)).Return(func() {}, nil)
		mockDb.EXPECT().ActorForOutbox(ctx, mustParse(testMyOutboxIRI)).Return(
			mustParse(testPersonIRI), nil)
		mockDb.EXPECT().Lock(ctx, mustParse(testPersonIRI)).Return(func() {}, nil)
		mockDb.EXPECT().Get(ctx, mustParse(testPersonIRI)).Return(
			testMyPerson, nil)
		c.EXPECT().NewTransport(ctx, mustParse(testMyOutboxIRI), goFedUserAgent()).Return(
			mockTp, nil)
		mockTp.EXPECT().BatchDeliver(ctx, mustSerializeToBytes(act), expectRecip)
		// Run & Verify
		err := a.Deliver(ctx, mustParse(testMyOutboxIRI), act)
		assertEqual(t, err, nil)
	})
	t.Run("RecursivelyResolveOrderedCollectionActors", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		c, mockFp, _, mockDb, _, a := setupFn(ctl)
		mockTp := NewMockTransport(ctl)
		act := baseActivityFn()
		to := streams.NewActivityStreamsToProperty()
		to.AppendIRI(mustParse(testAudienceIRI))
		act.SetActivityStreamsTo(to)
		expectRecip := []*url.URL{
			mustParse(testFederatedInboxIRI),
			mustParse(testFederatedInboxIRI2),
		}
		// Mock
		c.EXPECT().NewTransport(ctx, mustParse(testMyOutboxIRI), goFedUserAgent()).Return(
			mockTp, nil)
		mockFp.EXPECT().MaxDeliveryRecursionDepth(ctx).Return(2)
		mockTp.EXPECT().Dereference(ctx, mustParse(testAudienceIRI)).Return(
			mustSerializeToBytes(testOrderedCollectionOfActors), nil)
		mockTp.EXPECT().Dereference(ctx, mustParse(testFederatedActorIRI3)).Return(
			mustSerializeToBytes(testFederatedPerson1), nil)
		mockTp.EXPECT().Dereference(ctx, mustParse(testFederatedActorIRI4)).Return(
			mustSerializeToBytes(testFederatedPerson2), nil)
		mockDb.EXPECT().Lock(ctx, mustParse(testMyOutboxIRI)).Return(func() {}, nil)
		mockDb.EXPECT().ActorForOutbox(ctx, mustParse(testMyOutboxIRI)).Return(
			mustParse(testPersonIRI), nil)
		mockDb.EXPECT().Lock(ctx, mustParse(testPersonIRI)).Return(func() {}, nil)
		mockDb.EXPECT().Get(ctx, mustParse(testPersonIRI)).Return(
			testMyPerson, nil)
		c.EXPECT().NewTransport(ctx, mustParse(testMyOutboxIRI), goFedUserAgent()).Return(
			mockTp, nil)
		mockTp.EXPECT().BatchDeliver(ctx, mustSerializeToBytes(act), expectRecip)
		// Run & Verify
		err := a.Deliver(ctx, mustParse(testMyOutboxIRI), act)
		assertEqual(t, err, nil)
	})
	t.Run("DoesNotRecursivelyResolveCollectionActorsIfExceedingMaxDepth", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		c, mockFp, _, mockDb, _, a := setupFn(ctl)
		mockTp := NewMockTransport(ctl)
		act := baseActivityFn()
		to := streams.NewActivityStreamsToProperty()
		to.AppendIRI(mustParse(testAudienceIRI))
		act.SetActivityStreamsTo(to)
		// Mock
		c.EXPECT().NewTransport(ctx, mustParse(testMyOutboxIRI), goFedUserAgent()).Return(
			mockTp, nil)
		mockFp.EXPECT().MaxDeliveryRecursionDepth(ctx).Return(1)
		mockTp.EXPECT().Dereference(ctx, mustParse(testAudienceIRI)).Return(
			mustSerializeToBytes(testCollectionOfActors), nil)
		mockDb.EXPECT().Lock(ctx, mustParse(testMyOutboxIRI)).Return(func() {}, nil)
		mockDb.EXPECT().ActorForOutbox(ctx, mustParse(testMyOutboxIRI)).Return(
			mustParse(testPersonIRI), nil)
		mockDb.EXPECT().Lock(ctx, mustParse(testPersonIRI)).Return(func() {}, nil)
		mockDb.EXPECT().Get(ctx, mustParse(testPersonIRI)).Return(
			testMyPerson, nil)
		c.EXPECT().NewTransport(ctx, mustParse(testMyOutboxIRI), goFedUserAgent()).Return(
			mockTp, nil)
		mockTp.EXPECT().BatchDeliver(ctx, mustSerializeToBytes(act), nil)
		// Run & Verify
		err := a.Deliver(ctx, mustParse(testMyOutboxIRI), act)
		assertEqual(t, err, nil)
	})
	t.Run("DedupesRecipients", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		c, mockFp, _, mockDb, _, a := setupFn(ctl)
		mockTp := NewMockTransport(ctl)
		act := baseActivityFn()
		to := streams.NewActivityStreamsToProperty()
		to.AppendIRI(mustParse(testFederatedActorIRI))
		to.AppendIRI(mustParse(testFederatedActorIRI2))
		act.SetActivityStreamsTo(to)
		bto := streams.NewActivityStreamsBtoProperty()
		bto.AppendIRI(mustParse(testFederatedActorIRI))
		bto.AppendIRI(mustParse(testFederatedActorIRI2))
		act.SetActivityStreamsBto(bto)
		cc := streams.NewActivityStreamsCcProperty()
		cc.AppendIRI(mustParse(testFederatedActorIRI))
		cc.AppendIRI(mustParse(testFederatedActorIRI2))
		act.SetActivityStreamsCc(cc)
		bcc := streams.NewActivityStreamsBccProperty()
		bcc.AppendIRI(mustParse(testFederatedActorIRI))
		bcc.AppendIRI(mustParse(testFederatedActorIRI2))
		act.SetActivityStreamsBcc(bcc)
		expectAct := baseActivityFn() // Ensure Bcc & Bto are stripped
		expectAct.SetActivityStreamsTo(to)
		expectAct.SetActivityStreamsCc(cc)
		expectRecip := []*url.URL{
			mustParse(testFederatedInboxIRI),
			mustParse(testFederatedInboxIRI2),
		}
		// Mock
		c.EXPECT().NewTransport(ctx, mustParse(testMyOutboxIRI), goFedUserAgent()).Return(
			mockTp, nil)
		mockFp.EXPECT().MaxDeliveryRecursionDepth(ctx).Return(1)
		mockTp.EXPECT().Dereference(ctx, mustParse(testFederatedActorIRI)).Return(
			mustSerializeToBytes(testFederatedPerson1), nil).Times(4)
		mockTp.EXPECT().Dereference(ctx, mustParse(testFederatedActorIRI2)).Return(
			mustSerializeToBytes(testFederatedPerson2), nil).Times(4)
		mockDb.EXPECT().Lock(ctx, mustParse(testMyOutboxIRI)).Return(func() {}, nil)
		mockDb.EXPECT().ActorForOutbox(ctx, mustParse(testMyOutboxIRI)).Return(
			mustParse(testPersonIRI), nil)
		mockDb.EXPECT().Lock(ctx, mustParse(testPersonIRI)).Return(func() {}, nil)
		mockDb.EXPECT().Get(ctx, mustParse(testPersonIRI)).Return(
			testMyPerson, nil)
		c.EXPECT().NewTransport(ctx, mustParse(testMyOutboxIRI), goFedUserAgent()).Return(
			mockTp, nil)
		mockTp.EXPECT().BatchDeliver(ctx, mustSerializeToBytes(expectAct), expectRecip)
		// Run & Verify
		err := a.Deliver(ctx, mustParse(testMyOutboxIRI), act)
		assertEqual(t, err, nil)
	})
	t.Run("StripsBtoOnObject", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		c, mockFp, _, mockDb, _, a := setupFn(ctl)
		mockTp := NewMockTransport(ctl)
		act := baseActivityFn()
		bto := streams.NewActivityStreamsBtoProperty()
		bto.AppendIRI(mustParse(testFederatedActorIRI))
		bto.AppendIRI(mustParse(testFederatedActorIRI2))
		act.SetActivityStreamsBto(bto)
		act.GetActivityStreamsObject().At(0).GetActivityStreamsNote().SetActivityStreamsBto(bto)
		expectAct := baseActivityFn() // Ensure Bto is stripped
		expectRecip := []*url.URL{
			mustParse(testFederatedInboxIRI),
			mustParse(testFederatedInboxIRI2),
		}
		// Mock
		c.EXPECT().NewTransport(ctx, mustParse(testMyOutboxIRI), goFedUserAgent()).Return(
			mockTp, nil)
		mockFp.EXPECT().MaxDeliveryRecursionDepth(ctx).Return(1)
		mockTp.EXPECT().Dereference(ctx, mustParse(testFederatedActorIRI)).Return(
			mustSerializeToBytes(testFederatedPerson1), nil)
		mockTp.EXPECT().Dereference(ctx, mustParse(testFederatedActorIRI2)).Return(
			mustSerializeToBytes(testFederatedPerson2), nil)
		mockDb.EXPECT().Lock(ctx, mustParse(testMyOutboxIRI)).Return(func() {}, nil)
		mockDb.EXPECT().ActorForOutbox(ctx, mustParse(testMyOutboxIRI)).Return(
			mustParse(testPersonIRI), nil)
		mockDb.EXPECT().Lock(ctx, mustParse(testPersonIRI)).Return(func() {}, nil)
		mockDb.EXPECT().Get(ctx, mustParse(testPersonIRI)).Return(
			testMyPerson, nil)
		c.EXPECT().NewTransport(ctx, mustParse(testMyOutboxIRI), goFedUserAgent()).Return(
			mockTp, nil)
		mockTp.EXPECT().BatchDeliver(ctx, mustSerializeToBytes(expectAct), expectRecip)
		// Run & Verify
		err := a.Deliver(ctx, mustParse(testMyOutboxIRI), act)
		assertEqual(t, err, nil)
	})
	t.Run("StripsBccOnObject", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		c, mockFp, _, mockDb, _, a := setupFn(ctl)
		mockTp := NewMockTransport(ctl)
		act := baseActivityFn()
		bcc := streams.NewActivityStreamsBccProperty()
		bcc.AppendIRI(mustParse(testFederatedActorIRI))
		bcc.AppendIRI(mustParse(testFederatedActorIRI2))
		act.SetActivityStreamsBcc(bcc)
		act.GetActivityStreamsObject().At(0).GetActivityStreamsNote().SetActivityStreamsBcc(bcc)
		expectAct := baseActivityFn() // Ensure Bto is stripped
		expectRecip := []*url.URL{
			mustParse(testFederatedInboxIRI),
			mustParse(testFederatedInboxIRI2),
		}
		// Mock
		c.EXPECT().NewTransport(ctx, mustParse(testMyOutboxIRI), goFedUserAgent()).Return(
			mockTp, nil)
		mockFp.EXPECT().MaxDeliveryRecursionDepth(ctx).Return(1)
		mockTp.EXPECT().Dereference(ctx, mustParse(testFederatedActorIRI)).Return(
			mustSerializeToBytes(testFederatedPerson1), nil)
		mockTp.EXPECT().Dereference(ctx, mustParse(testFederatedActorIRI2)).Return(
			mustSerializeToBytes(testFederatedPerson2), nil)
		mockDb.EXPECT().Lock(ctx, mustParse(testMyOutboxIRI)).Return(func() {}, nil)
		mockDb.EXPECT().ActorForOutbox(ctx, mustParse(testMyOutboxIRI)).Return(
			mustParse(testPersonIRI), nil)
		mockDb.EXPECT().Lock(ctx, mustParse(testPersonIRI)).Return(func() {}, nil)
		mockDb.EXPECT().Get(ctx, mustParse(testPersonIRI)).Return(
			testMyPerson, nil)
		c.EXPECT().NewTransport(ctx, mustParse(testMyOutboxIRI), goFedUserAgent()).Return(
			mockTp, nil)
		mockTp.EXPECT().BatchDeliver(ctx, mustSerializeToBytes(expectAct), expectRecip)
		// Run & Verify
		err := a.Deliver(ctx, mustParse(testMyOutboxIRI), act)
		assertEqual(t, err, nil)
	})
	t.Run("DoesNotReturnErrorIfDereferenceRecipientFails", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		c, mockFp, _, mockDb, _, a := setupFn(ctl)
		mockTp := NewMockTransport(ctl)
		act := baseActivityFn()
		to := streams.NewActivityStreamsToProperty()
		to.AppendIRI(mustParse(testFederatedActorIRI))
		to.AppendIRI(mustParse(testFederatedActorIRI2))
		act.SetActivityStreamsTo(to)
		expectRecip := []*url.URL{
			mustParse(testFederatedInboxIRI2),
		}
		// Mock
		c.EXPECT().NewTransport(ctx, mustParse(testMyOutboxIRI), goFedUserAgent()).Return(
			mockTp, nil)
		mockFp.EXPECT().MaxDeliveryRecursionDepth(ctx).Return(1)
		mockTp.EXPECT().Dereference(ctx, mustParse(testFederatedActorIRI)).Return(
			[]byte{}, fmt.Errorf("test error"))
		mockTp.EXPECT().Dereference(ctx, mustParse(testFederatedActorIRI2)).Return(
			mustSerializeToBytes(testFederatedPerson2), nil)
		mockDb.EXPECT().Lock(ctx, mustParse(testMyOutboxIRI)).Return(func() {}, nil)
		mockDb.EXPECT().ActorForOutbox(ctx, mustParse(testMyOutboxIRI)).Return(
			mustParse(testPersonIRI), nil)
		mockDb.EXPECT().Lock(ctx, mustParse(testPersonIRI)).Return(func() {}, nil)
		mockDb.EXPECT().Get(ctx, mustParse(testPersonIRI)).Return(
			testMyPerson, nil)
		c.EXPECT().NewTransport(ctx, mustParse(testMyOutboxIRI), goFedUserAgent()).Return(
			mockTp, nil)
		mockTp.EXPECT().BatchDeliver(ctx, mustSerializeToBytes(act), expectRecip)
		// Run & Verify
		err := a.Deliver(ctx, mustParse(testMyOutboxIRI), act)
		assertEqual(t, err, nil)
	})
	t.Run("ReturnsErrorIfBatchDeliverFails", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		c, mockFp, _, mockDb, _, a := setupFn(ctl)
		mockTp := NewMockTransport(ctl)
		act := baseActivityFn()
		to := streams.NewActivityStreamsToProperty()
		to.AppendIRI(mustParse(testFederatedActorIRI))
		to.AppendIRI(mustParse(testFederatedActorIRI2))
		act.SetActivityStreamsTo(to)
		expectRecip := []*url.URL{
			mustParse(testFederatedInboxIRI),
			mustParse(testFederatedInboxIRI2),
		}
		expectErr := fmt.Errorf("test error")
		// Mock
		c.EXPECT().NewTransport(ctx, mustParse(testMyOutboxIRI), goFedUserAgent()).Return(
			mockTp, nil)
		mockFp.EXPECT().MaxDeliveryRecursionDepth(ctx).Return(1)
		mockTp.EXPECT().Dereference(ctx, mustParse(testFederatedActorIRI)).Return(
			mustSerializeToBytes(testFederatedPerson1), nil)
		mockTp.EXPECT().Dereference(ctx, mustParse(testFederatedActorIRI2)).Return(
			mustSerializeToBytes(testFederatedPerson2), nil)
		mockDb.EXPECT().Lock(ctx, mustParse(testMyOutboxIRI)).Return(func() {}, nil)
		mockDb.EXPECT().ActorForOutbox(ctx, mustParse(testMyOutboxIRI)).Return(
			mustParse(testPersonIRI), nil)
		mockDb.EXPECT().Lock(ctx, mustParse(testPersonIRI)).Return(func() {}, nil)
		mockDb.EXPECT().Get(ctx, mustParse(testPersonIRI)).Return(
			testMyPerson, nil)
		c.EXPECT().NewTransport(ctx, mustParse(testMyOutboxIRI), goFedUserAgent()).Return(
			mockTp, nil)
		mockTp.EXPECT().BatchDeliver(ctx, mustSerializeToBytes(act), expectRecip).Return(
			expectErr)
		// Run & Verify
		err := a.Deliver(ctx, mustParse(testMyOutboxIRI), act)
		assertEqual(t, err, expectErr)
	})
}

// TestWrapInCreate ensures an object received by the Social Protocol is
// properly wrapped in a Create Activity.
func TestWrapInCreate(t *testing.T) {
	baseNoteFn := func() (vocab.ActivityStreamsNote, vocab.ActivityStreamsCreate) {
		n := streams.NewActivityStreamsNote()
		id := streams.NewJSONLDIdProperty()
		id.Set(mustParse(testNoteId1))
		n.SetJSONLDId(id)
		cr := streams.NewActivityStreamsCreate()
		op := streams.NewActivityStreamsObjectProperty()
		op.AppendActivityStreamsNote(n)
		cr.SetActivityStreamsObject(op)
		actorProp := streams.NewActivityStreamsActorProperty()
		actorProp.AppendIRI(mustParse(testPersonIRI))
		cr.SetActivityStreamsActor(actorProp)
		return n, cr
	}
	ctx := context.Background()
	setupFn := func(ctl *gomock.Controller) (c *MockCommonBehavior, fp *MockFederatingProtocol, sp *MockSocialProtocol, db *MockDatabase, cl *MockClock, a DelegateActor) {
		setupData()
		c = NewMockCommonBehavior(ctl)
		fp = NewMockFederatingProtocol(ctl)
		sp = NewMockSocialProtocol(ctl)
		db = NewMockDatabase(ctl)
		cl = NewMockClock(ctl)
		a = &SideEffectActor{
			common: c,
			s2s:    fp,
			c2s:    sp,
			db:     db,
			clock:  cl,
		}
		return
	}
	t.Run("CreateHasObjectAndActor", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		_, _, _, mockDb, _, a := setupFn(ctl)
		n, expect := baseNoteFn()
		// Mock
		mockDb.EXPECT().Lock(ctx, mustParse(testMyOutboxIRI)).Return(func() {}, nil)
		mockDb.EXPECT().ActorForOutbox(ctx, mustParse(testMyOutboxIRI)).Return(
			mustParse(testPersonIRI), nil)
		// Run & Verify
		got, err := a.WrapInCreate(ctx, n, mustParse(testMyOutboxIRI))
		assertEqual(t, err, nil)
		assertByteEqual(t, mustSerializeToBytes(got), mustSerializeToBytes(expect))
	})
	t.Run("CreateHasTo", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		_, _, _, mockDb, _, a := setupFn(ctl)
		n, expect := baseNoteFn()
		to := streams.NewActivityStreamsToProperty()
		to.AppendIRI(mustParse(testFederatedActorIRI))
		to.AppendIRI(mustParse(testFederatedActorIRI2))
		n.SetActivityStreamsTo(to)
		expect.SetActivityStreamsTo(to)
		// Mock
		mockDb.EXPECT().Lock(ctx, mustParse(testMyOutboxIRI)).Return(func() {}, nil)
		mockDb.EXPECT().ActorForOutbox(ctx, mustParse(testMyOutboxIRI)).Return(
			mustParse(testPersonIRI), nil)
		// Run & Verify
		got, err := a.WrapInCreate(ctx, n, mustParse(testMyOutboxIRI))
		assertEqual(t, err, nil)
		assertByteEqual(t, mustSerializeToBytes(got), mustSerializeToBytes(expect))
	})
	t.Run("CreateHasCc", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		_, _, _, mockDb, _, a := setupFn(ctl)
		n, expect := baseNoteFn()
		cc := streams.NewActivityStreamsCcProperty()
		cc.AppendIRI(mustParse(testFederatedActorIRI))
		cc.AppendIRI(mustParse(testFederatedActorIRI2))
		n.SetActivityStreamsCc(cc)
		expect.SetActivityStreamsCc(cc)
		// Mock
		mockDb.EXPECT().Lock(ctx, mustParse(testMyOutboxIRI)).Return(func() {}, nil)
		mockDb.EXPECT().ActorForOutbox(ctx, mustParse(testMyOutboxIRI)).Return(
			mustParse(testPersonIRI), nil)
		// Run & Verify
		got, err := a.WrapInCreate(ctx, n, mustParse(testMyOutboxIRI))
		assertEqual(t, err, nil)
		assertByteEqual(t, mustSerializeToBytes(got), mustSerializeToBytes(expect))
	})
	t.Run("CreateHasBto", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		_, _, _, mockDb, _, a := setupFn(ctl)
		n, expect := baseNoteFn()
		bto := streams.NewActivityStreamsBtoProperty()
		bto.AppendIRI(mustParse(testFederatedActorIRI))
		bto.AppendIRI(mustParse(testFederatedActorIRI2))
		n.SetActivityStreamsBto(bto)
		expect.SetActivityStreamsBto(bto)
		// Mock
		mockDb.EXPECT().Lock(ctx, mustParse(testMyOutboxIRI)).Return(func() {}, nil)
		mockDb.EXPECT().ActorForOutbox(ctx, mustParse(testMyOutboxIRI)).Return(
			mustParse(testPersonIRI), nil)
		// Run & Verify
		got, err := a.WrapInCreate(ctx, n, mustParse(testMyOutboxIRI))
		assertEqual(t, err, nil)
		assertByteEqual(t, mustSerializeToBytes(got), mustSerializeToBytes(expect))
	})
	t.Run("CreateHasBcc", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		_, _, _, mockDb, _, a := setupFn(ctl)
		n, expect := baseNoteFn()
		bcc := streams.NewActivityStreamsBccProperty()
		bcc.AppendIRI(mustParse(testFederatedActorIRI))
		bcc.AppendIRI(mustParse(testFederatedActorIRI2))
		n.SetActivityStreamsBcc(bcc)
		expect.SetActivityStreamsBcc(bcc)
		// Mock
		mockDb.EXPECT().Lock(ctx, mustParse(testMyOutboxIRI)).Return(func() {}, nil)
		mockDb.EXPECT().ActorForOutbox(ctx, mustParse(testMyOutboxIRI)).Return(
			mustParse(testPersonIRI), nil)
		// Run & Verify
		got, err := a.WrapInCreate(ctx, n, mustParse(testMyOutboxIRI))
		assertEqual(t, err, nil)
		assertByteEqual(t, mustSerializeToBytes(got), mustSerializeToBytes(expect))
	})
	t.Run("CreateHasAudience", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		_, _, _, mockDb, _, a := setupFn(ctl)
		n, expect := baseNoteFn()
		aud := streams.NewActivityStreamsAudienceProperty()
		aud.AppendIRI(mustParse(testFederatedActorIRI))
		aud.AppendIRI(mustParse(testFederatedActorIRI2))
		n.SetActivityStreamsAudience(aud)
		expect.SetActivityStreamsAudience(aud)
		// Mock
		mockDb.EXPECT().Lock(ctx, mustParse(testMyOutboxIRI)).Return(func() {}, nil)
		mockDb.EXPECT().ActorForOutbox(ctx, mustParse(testMyOutboxIRI)).Return(
			mustParse(testPersonIRI), nil)
		// Run & Verify
		got, err := a.WrapInCreate(ctx, n, mustParse(testMyOutboxIRI))
		assertEqual(t, err, nil)
		assertByteEqual(t, mustSerializeToBytes(got), mustSerializeToBytes(expect))
	})
	t.Run("CreateHasPublished", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		_, _, _, mockDb, _, a := setupFn(ctl)
		n, expect := baseNoteFn()
		pub := streams.NewActivityStreamsPublishedProperty()
		pub.Set(time.Now())
		n.SetActivityStreamsPublished(pub)
		expect.SetActivityStreamsPublished(pub)
		// Mock
		mockDb.EXPECT().Lock(ctx, mustParse(testMyOutboxIRI)).Return(func() {}, nil)
		mockDb.EXPECT().ActorForOutbox(ctx, mustParse(testMyOutboxIRI)).Return(
			mustParse(testPersonIRI), nil)
		// Run & Verify
		got, err := a.WrapInCreate(ctx, n, mustParse(testMyOutboxIRI))
		assertEqual(t, err, nil)
		assertByteEqual(t, mustSerializeToBytes(got), mustSerializeToBytes(expect))
	})
}

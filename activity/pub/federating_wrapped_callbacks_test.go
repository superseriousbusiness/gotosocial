package pub

import (
	"context"
	"net/url"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/activity/streams/vocab"
)

// TestFederatedCallbacks tests the overriding functionality.
func TestFederatedCallbacks(t *testing.T) {
	t.Run("ReturnsOtherCallback", func(t *testing.T) {
		ok := false
		o := func(context.Context, vocab.ActivityStreamsListen) error {
			ok = true
			return nil
		}
		var w FederatingWrappedCallbacks
		for _, f := range w.callbacks([]interface{}{o}) {
			if fn, ok := f.(func(context.Context, vocab.ActivityStreamsListen) error); ok {
				fn(nil, nil)
			}
		}
		if !ok {
			t.Fatalf("could not find extra function")
		}
	})
	t.Run("OverridesCreate", func(t *testing.T) {
		ok := false
		o := func(context.Context, vocab.ActivityStreamsCreate) error {
			ok = true
			return nil
		}
		var w FederatingWrappedCallbacks
		for _, f := range w.callbacks([]interface{}{o}) {
			if fn, ok := f.(func(context.Context, vocab.ActivityStreamsCreate) error); ok {
				fn(nil, nil)
			}
		}
		if !ok {
			t.Fatalf("could not find overridden function")
		}
	})
	t.Run("OverridesUpdate", func(t *testing.T) {
		ok := false
		o := func(context.Context, vocab.ActivityStreamsUpdate) error {
			ok = true
			return nil
		}
		var w FederatingWrappedCallbacks
		for _, f := range w.callbacks([]interface{}{o}) {
			if fn, ok := f.(func(context.Context, vocab.ActivityStreamsUpdate) error); ok {
				fn(nil, nil)
			}
		}
		if !ok {
			t.Fatalf("could not find overridden function")
		}
	})
	t.Run("OverridesDelete", func(t *testing.T) {
		ok := false
		o := func(context.Context, vocab.ActivityStreamsDelete) error {
			ok = true
			return nil
		}
		var w FederatingWrappedCallbacks
		for _, f := range w.callbacks([]interface{}{o}) {
			if fn, ok := f.(func(context.Context, vocab.ActivityStreamsDelete) error); ok {
				fn(nil, nil)
			}
		}
		if !ok {
			t.Fatalf("could not find overridden function")
		}
	})
	t.Run("OverridesFollow", func(t *testing.T) {
		ok := false
		o := func(context.Context, vocab.ActivityStreamsFollow) error {
			ok = true
			return nil
		}
		var w FederatingWrappedCallbacks
		for _, f := range w.callbacks([]interface{}{o}) {
			if fn, ok := f.(func(context.Context, vocab.ActivityStreamsFollow) error); ok {
				fn(nil, nil)
			}
		}
		if !ok {
			t.Fatalf("could not find overridden function")
		}
	})
	t.Run("OverridesAccept", func(t *testing.T) {
		ok := false
		o := func(context.Context, vocab.ActivityStreamsAccept) error {
			ok = true
			return nil
		}
		var w FederatingWrappedCallbacks
		for _, f := range w.callbacks([]interface{}{o}) {
			if fn, ok := f.(func(context.Context, vocab.ActivityStreamsAccept) error); ok {
				fn(nil, nil)
			}
		}
		if !ok {
			t.Fatalf("could not find overridden function")
		}
	})
	t.Run("OverridesReject", func(t *testing.T) {
		ok := false
		o := func(context.Context, vocab.ActivityStreamsReject) error {
			ok = true
			return nil
		}
		var w FederatingWrappedCallbacks
		for _, f := range w.callbacks([]interface{}{o}) {
			if fn, ok := f.(func(context.Context, vocab.ActivityStreamsReject) error); ok {
				fn(nil, nil)
			}
		}
		if !ok {
			t.Fatalf("could not find overridden function")
		}
	})
	t.Run("OverridesAdd", func(t *testing.T) {
		ok := false
		o := func(context.Context, vocab.ActivityStreamsAdd) error {
			ok = true
			return nil
		}
		var w FederatingWrappedCallbacks
		for _, f := range w.callbacks([]interface{}{o}) {
			if fn, ok := f.(func(context.Context, vocab.ActivityStreamsAdd) error); ok {
				fn(nil, nil)
			}
		}
		if !ok {
			t.Fatalf("could not find overridden function")
		}
	})
	t.Run("OverridesRemove", func(t *testing.T) {
		ok := false
		o := func(context.Context, vocab.ActivityStreamsRemove) error {
			ok = true
			return nil
		}
		var w FederatingWrappedCallbacks
		for _, f := range w.callbacks([]interface{}{o}) {
			if fn, ok := f.(func(context.Context, vocab.ActivityStreamsRemove) error); ok {
				fn(nil, nil)
			}
		}
		if !ok {
			t.Fatalf("could not find overridden function")
		}
	})
	t.Run("OverridesLike", func(t *testing.T) {
		ok := false
		o := func(context.Context, vocab.ActivityStreamsLike) error {
			ok = true
			return nil
		}
		var w FederatingWrappedCallbacks
		for _, f := range w.callbacks([]interface{}{o}) {
			if fn, ok := f.(func(context.Context, vocab.ActivityStreamsLike) error); ok {
				fn(nil, nil)
			}
		}
		if !ok {
			t.Fatalf("could not find overridden function")
		}
	})
	t.Run("OverridesAnnounce", func(t *testing.T) {
		ok := false
		o := func(context.Context, vocab.ActivityStreamsAnnounce) error {
			ok = true
			return nil
		}
		var w FederatingWrappedCallbacks
		for _, f := range w.callbacks([]interface{}{o}) {
			if fn, ok := f.(func(context.Context, vocab.ActivityStreamsAnnounce) error); ok {
				fn(nil, nil)
			}
		}
		if !ok {
			t.Fatalf("could not find overridden function")
		}
	})
	t.Run("OverridesUndo", func(t *testing.T) {
		ok := false
		o := func(context.Context, vocab.ActivityStreamsUndo) error {
			ok = true
			return nil
		}
		var w FederatingWrappedCallbacks
		for _, f := range w.callbacks([]interface{}{o}) {
			if fn, ok := f.(func(context.Context, vocab.ActivityStreamsUndo) error); ok {
				fn(nil, nil)
			}
		}
		if !ok {
			t.Fatalf("could not find overridden function")
		}
	})
	t.Run("OverridesBlock", func(t *testing.T) {
		ok := false
		o := func(context.Context, vocab.ActivityStreamsBlock) error {
			ok = true
			return nil
		}
		var w FederatingWrappedCallbacks
		for _, f := range w.callbacks([]interface{}{o}) {
			if fn, ok := f.(func(context.Context, vocab.ActivityStreamsBlock) error); ok {
				fn(nil, nil)
			}
		}
		if !ok {
			t.Fatalf("could not find overridden function")
		}
	})
}

func TestFederatedCreate(t *testing.T) {
	newCreateFn := func() vocab.ActivityStreamsCreate {
		c := streams.NewActivityStreamsCreate()
		id := streams.NewJSONLDIdProperty()
		id.Set(mustParse(testFederatedActivityIRI))
		c.SetJSONLDId(id)
		actor := streams.NewActivityStreamsActorProperty()
		actor.AppendIRI(mustParse(testFederatedActorIRI))
		c.SetActivityStreamsActor(actor)
		op := streams.NewActivityStreamsObjectProperty()
		op.AppendActivityStreamsNote(testFederatedNote)
		c.SetActivityStreamsObject(op)
		return c
	}
	ctx := context.Background()
	setupFn := func(ctl *gomock.Controller) (w FederatingWrappedCallbacks, mockDB *MockDatabase, mockTp *MockTransport) {
		mockDB = NewMockDatabase(ctl)
		mockTp = NewMockTransport(ctl)
		w.db = mockDB
		w.newTransport = func(c context.Context, a *url.URL, s string) (Transport, error) {
			return mockTp, nil
		}
		return
	}
	t.Run("ErrorIfNoObject", func(t *testing.T) {
		c := newCreateFn()
		c.SetActivityStreamsObject(nil)
		var w FederatingWrappedCallbacks
		err := w.create(ctx, c)
		if err == nil {
			t.Fatalf("expected error, got none")
		}
	})
	t.Run("ErrorIfObjectLengthZero", func(t *testing.T) {
		c := newCreateFn()
		c.GetActivityStreamsObject().Remove(0)
		var w FederatingWrappedCallbacks
		err := w.create(ctx, c)
		if err == nil {
			t.Fatalf("expected error, got none")
		}
	})
	t.Run("CreatesFederatedObject", func(t *testing.T) {
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		w, mockDB, _ := setupFn(ctl)
		mockDB.EXPECT().Lock(ctx, mustParse(testNoteId1)).Return(func() {}, nil)
		mockDB.EXPECT().Create(ctx, testFederatedNote)
		c := newCreateFn()
		err := w.create(ctx, c)
		if err != nil {
			t.Fatalf("got error %s", err)
		}
	})
	t.Run("CreatesAllFederatedObjects", func(t *testing.T) {
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		w, mockDB, _ := setupFn(ctl)
		mockDB.EXPECT().Lock(ctx, mustParse(testNoteId1)).Return(func() {}, nil)
		mockDB.EXPECT().Create(ctx, testFederatedNote)
		mockDB.EXPECT().Lock(ctx, mustParse(testNoteId2)).Return(func() {}, nil)
		mockDB.EXPECT().Create(ctx, testFederatedNote2)
		c := newCreateFn()
		c.GetActivityStreamsObject().AppendActivityStreamsNote(testFederatedNote2)
		err := w.create(ctx, c)
		if err != nil {
			t.Fatalf("got error %s", err)
		}
	})
	t.Run("DereferencesIRIObject", func(t *testing.T) {
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		w, mockDB, mockTp := setupFn(ctl)
		mockDB.EXPECT().Lock(ctx, mustParse(testNoteId1)).Return(func() {}, nil)
		mockDB.EXPECT().Create(ctx, toDeserializedForm(testFederatedNote))
		mockTp.EXPECT().Dereference(ctx, mustParse(testNoteId1)).Return(
			mustWrapInGETResponse(mustParse(testNoteId1), testFederatedNote), nil)
		c := newCreateFn()
		op := streams.NewActivityStreamsObjectProperty()
		op.AppendIRI(mustParse(testNoteId1))
		c.SetActivityStreamsObject(op)
		err := w.create(ctx, c)
		if err != nil {
			t.Fatalf("got error %s", err)
		}
	})
	t.Run("CallsCustomCallback", func(t *testing.T) {
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		w, mockDB, _ := setupFn(ctl)
		mockDB.EXPECT().Lock(ctx, mustParse(testNoteId1)).Return(func() {}, nil)
		mockDB.EXPECT().Create(ctx, testFederatedNote)
		c := newCreateFn()
		var gotc context.Context
		var got vocab.ActivityStreamsCreate
		w.Create = func(ctx context.Context, v vocab.ActivityStreamsCreate) error {
			gotc = ctx
			got = v
			return nil
		}
		err := w.create(ctx, c)
		if err != nil {
			t.Fatalf("got error %s", err)
		}
		assertEqual(t, ctx, gotc)
		assertEqual(t, c, got)
	})
}

func TestFederatedUpdate(t *testing.T) {
	newUpdateFn := func() vocab.ActivityStreamsUpdate {
		u := streams.NewActivityStreamsUpdate()
		id := streams.NewJSONLDIdProperty()
		id.Set(mustParse(testNewActivityIRI))
		u.SetJSONLDId(id)
		actor := streams.NewActivityStreamsActorProperty()
		actor.AppendIRI(mustParse(testFederatedActorIRI))
		u.SetActivityStreamsActor(actor)
		op := streams.NewActivityStreamsObjectProperty()
		op.AppendActivityStreamsNote(testFederatedNote)
		u.SetActivityStreamsObject(op)
		return u
	}
	ctx := context.Background()
	setupFn := func(ctl *gomock.Controller) (w FederatingWrappedCallbacks, mockDB *MockDatabase) {
		mockDB = NewMockDatabase(ctl)
		w.db = mockDB
		return
	}
	t.Run("ErrorIfNoObject", func(t *testing.T) {
		u := newUpdateFn()
		u.SetActivityStreamsObject(nil)
		var w FederatingWrappedCallbacks
		err := w.update(ctx, u)
		if err == nil {
			t.Fatalf("expected error, got none")
		}
	})
	t.Run("ErrorIfObjectLengthZero", func(t *testing.T) {
		u := newUpdateFn()
		u.GetActivityStreamsObject().Remove(0)
		var w FederatingWrappedCallbacks
		err := w.update(ctx, u)
		if err == nil {
			t.Fatalf("expected error, got none")
		}
	})
	t.Run("ErrorIfOriginMismatchesObject", func(t *testing.T) {
		u := newUpdateFn()
		id := streams.NewJSONLDIdProperty()
		id.Set(mustParse(testFederatedActivityIRI))
		u.SetJSONLDId(id)
		var w FederatingWrappedCallbacks
		err := w.update(ctx, u)
		if err == nil {
			t.Fatalf("expected error, got none")
		}
	})
	t.Run("UpdatesFederatedObject", func(t *testing.T) {
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		w, mockDB := setupFn(ctl)
		mockDB.EXPECT().Lock(ctx, mustParse(testNoteId1)).Return(func() {}, nil)
		mockDB.EXPECT().Update(ctx, testFederatedNote)
		u := newUpdateFn()
		err := w.update(ctx, u)
		if err != nil {
			t.Fatalf("got error %s", err)
		}
	})
	t.Run("UpdatesAllFederatedObjects", func(t *testing.T) {
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		w, mockDB := setupFn(ctl)
		mockDB.EXPECT().Lock(ctx, mustParse(testNoteId1)).Return(func() {}, nil)
		mockDB.EXPECT().Update(ctx, testFederatedNote)
		mockDB.EXPECT().Lock(ctx, mustParse(testNoteId2)).Return(func() {}, nil)
		mockDB.EXPECT().Update(ctx, testFederatedNote2)
		u := newUpdateFn()
		u.GetActivityStreamsObject().AppendActivityStreamsNote(testFederatedNote2)
		err := w.update(ctx, u)
		if err != nil {
			t.Fatalf("got error %s", err)
		}
	})
	t.Run("ErrorIfObjectIsIRI", func(t *testing.T) {
		u := newUpdateFn()
		op := streams.NewActivityStreamsObjectProperty()
		op.AppendIRI(mustParse(testNoteId1))
		u.SetActivityStreamsObject(op)
		var w FederatingWrappedCallbacks
		err := w.update(ctx, u)
		if err == nil {
			t.Fatalf("expected error, got none")
		}
	})
	t.Run("CallsCustomCallback", func(t *testing.T) {
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		w, mockDB := setupFn(ctl)
		mockDB.EXPECT().Lock(ctx, mustParse(testNoteId1)).Return(func() {}, nil)
		mockDB.EXPECT().Update(ctx, testFederatedNote)
		u := newUpdateFn()
		var gotc context.Context
		var got vocab.ActivityStreamsUpdate
		w.Update = func(ctx context.Context, v vocab.ActivityStreamsUpdate) error {
			gotc = ctx
			got = v
			return nil
		}
		err := w.update(ctx, u)
		if err != nil {
			t.Fatalf("got error %s", err)
		}
		assertEqual(t, ctx, gotc)
		assertEqual(t, u, got)
	})
}

func TestFederatedDelete(t *testing.T) {
	newDeleteFn := func() vocab.ActivityStreamsDelete {
		d := streams.NewActivityStreamsDelete()
		id := streams.NewJSONLDIdProperty()
		id.Set(mustParse(testNewActivityIRI))
		d.SetJSONLDId(id)
		actor := streams.NewActivityStreamsActorProperty()
		actor.AppendIRI(mustParse(testFederatedActorIRI))
		d.SetActivityStreamsActor(actor)
		op := streams.NewActivityStreamsObjectProperty()
		op.AppendIRI(mustParse(testNoteId1))
		d.SetActivityStreamsObject(op)
		return d
	}
	ctx := context.Background()
	setupFn := func(ctl *gomock.Controller) (w FederatingWrappedCallbacks, mockDB *MockDatabase) {
		mockDB = NewMockDatabase(ctl)
		w.db = mockDB
		return
	}
	t.Run("ErrorIfNoObject", func(t *testing.T) {
		d := newDeleteFn()
		d.SetActivityStreamsObject(nil)
		var w FederatingWrappedCallbacks
		err := w.deleteFn(ctx, d)
		if err == nil {
			t.Fatalf("expected error, got none")
		}
	})
	t.Run("ErrorIfObjectLengthZero", func(t *testing.T) {
		d := newDeleteFn()
		d.GetActivityStreamsObject().Remove(0)
		var w FederatingWrappedCallbacks
		err := w.deleteFn(ctx, d)
		if err == nil {
			t.Fatalf("expected error, got none")
		}
	})
	t.Run("ErrorIfOriginMismatchesObject", func(t *testing.T) {
		d := newDeleteFn()
		id := streams.NewJSONLDIdProperty()
		id.Set(mustParse(testFederatedActivityIRI))
		d.SetJSONLDId(id)
		var w FederatingWrappedCallbacks
		err := w.deleteFn(ctx, d)
		if err == nil {
			t.Fatalf("expected error, got none")
		}
	})
	t.Run("DeletesFederatedObject", func(t *testing.T) {
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		w, mockDB := setupFn(ctl)
		mockDB.EXPECT().Lock(ctx, mustParse(testNoteId1)).Return(func() {}, nil)
		mockDB.EXPECT().Delete(ctx, mustParse(testNoteId1))
		d := newDeleteFn()
		err := w.deleteFn(ctx, d)
		if err != nil {
			t.Fatalf("got error %s", err)
		}
	})
	t.Run("DeletesAllFederatedObjects", func(t *testing.T) {
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		w, mockDB := setupFn(ctl)
		mockDB.EXPECT().Lock(ctx, mustParse(testNoteId1)).Return(func() {}, nil)
		mockDB.EXPECT().Delete(ctx, mustParse(testNoteId1))
		mockDB.EXPECT().Lock(ctx, mustParse(testNoteId2)).Return(func() {}, nil)
		mockDB.EXPECT().Delete(ctx, mustParse(testNoteId2))
		d := newDeleteFn()
		d.GetActivityStreamsObject().AppendIRI(mustParse(testNoteId2))
		err := w.deleteFn(ctx, d)
		if err != nil {
			t.Fatalf("got error %s", err)
		}
	})
	t.Run("CallsCustomCallback", func(t *testing.T) {
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		w, mockDB := setupFn(ctl)
		mockDB.EXPECT().Lock(ctx, mustParse(testNoteId1)).Return(func() {}, nil)
		mockDB.EXPECT().Delete(ctx, mustParse(testNoteId1))
		d := newDeleteFn()
		var gotc context.Context
		var got vocab.ActivityStreamsDelete
		w.Delete = func(ctx context.Context, v vocab.ActivityStreamsDelete) error {
			gotc = ctx
			got = v
			return nil
		}
		err := w.deleteFn(ctx, d)
		if err != nil {
			t.Fatalf("got error %s", err)
		}
		assertEqual(t, ctx, gotc)
		assertEqual(t, d, got)
	})
}

func TestFederatedFollow(t *testing.T) {
	newFollowFn := func() vocab.ActivityStreamsFollow {
		f := streams.NewActivityStreamsFollow()
		id := streams.NewJSONLDIdProperty()
		id.Set(mustParse(testNewActivityIRI))
		f.SetJSONLDId(id)
		actor := streams.NewActivityStreamsActorProperty()
		actor.AppendIRI(mustParse(testFederatedActorIRI))
		f.SetActivityStreamsActor(actor)
		op := streams.NewActivityStreamsObjectProperty()
		op.AppendIRI(mustParse(testFederatedActorIRI2))
		f.SetActivityStreamsObject(op)
		return f
	}
	ctx := context.Background()
	setupFn := func(ctl *gomock.Controller) (w FederatingWrappedCallbacks, mockDB *MockDatabase) {
		mockDB = NewMockDatabase(ctl)
		w.db = mockDB
		w.inboxIRI = mustParse(testMyInboxIRI)
		return
	}
	t.Run("ErrorIfNoObject", func(t *testing.T) {
		f := newFollowFn()
		f.SetActivityStreamsObject(nil)
		var w FederatingWrappedCallbacks
		err := w.follow(ctx, f)
		if err == nil {
			t.Fatalf("expected error, got none")
		}
	})
	t.Run("ErrorIfObjectLengthZero", func(t *testing.T) {
		f := newFollowFn()
		f.GetActivityStreamsObject().Remove(0)
		var w FederatingWrappedCallbacks
		err := w.follow(ctx, f)
		if err == nil {
			t.Fatalf("expected error, got none")
		}
	})
	t.Run("OnFollowNothingDoesNothing", func(t *testing.T) {
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		w, mockDB := setupFn(ctl)
		w.OnFollow = OnFollowDoNothing
		mockDB.EXPECT().Lock(ctx, mustParse(testMyInboxIRI)).Return(func() {}, nil)
		mockDB.EXPECT().ActorForInbox(ctx, mustParse(testMyInboxIRI)).Return(
			mustParse(testFederatedActorIRI2), nil)
		f := newFollowFn()
		err := w.follow(ctx, f)
		if err != nil {
			t.Fatalf("got error %s", err)
		}
	})
	t.Run("OnFollowAutomaticallyAcceptUpdatesFollowers", func(t *testing.T) {
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		w, mockDB := setupFn(ctl)
		w.OnFollow = OnFollowAutomaticallyAccept
		w.addNewIds = func(c context.Context, activity Activity) error {
			return nil
		}
		w.deliver = func(c context.Context, outboxIRI *url.URL, activity Activity) error {
			return nil
		}
		followers := streams.NewActivityStreamsCollection()
		expectFollowers := streams.NewActivityStreamsCollection()
		expectItems := streams.NewActivityStreamsItemsProperty()
		expectItems.AppendIRI(mustParse(testFederatedActorIRI))
		expectFollowers.SetActivityStreamsItems(expectItems)
		mockDB.EXPECT().Lock(ctx, mustParse(testMyInboxIRI)).Return(func() {}, nil)
		mockDB.EXPECT().ActorForInbox(ctx, mustParse(testMyInboxIRI)).Return(
			mustParse(testFederatedActorIRI2), nil)
		mockDB.EXPECT().Lock(ctx, mustParse(testFederatedActorIRI2)).Return(func() {}, nil)
		mockDB.EXPECT().Followers(ctx, mustParse(testFederatedActorIRI2)).Return(
			followers, nil)
		mockDB.EXPECT().Update(ctx, expectFollowers)
		mockDB.EXPECT().Lock(ctx, mustParse(testMyInboxIRI)).Return(func() {}, nil)
		mockDB.EXPECT().OutboxForInbox(ctx, mustParse(testMyInboxIRI)).Return(
			mustParse(testMyOutboxIRI), nil)
		f := newFollowFn()
		err := w.follow(ctx, f)
		if err != nil {
			t.Fatalf("got error %s", err)
		}
	})
	t.Run("OnFollowAutomaticallyAcceptDelivers", func(t *testing.T) {
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		w, mockDB := setupFn(ctl)
		w.OnFollow = OnFollowAutomaticallyAccept
		w.addNewIds = func(c context.Context, activity Activity) error {
			return nil
		}
		w.deliver = func(c context.Context, outboxIRI *url.URL, activity Activity) error {
			if !streams.IsOrExtendsActivityStreamsAccept(activity) {
				t.Fatalf("expected Accept, got %T", activity)
			}
			return nil
		}
		followers := streams.NewActivityStreamsCollection()
		mockDB.EXPECT().Lock(ctx, mustParse(testMyInboxIRI)).Return(func() {}, nil)
		mockDB.EXPECT().ActorForInbox(ctx, mustParse(testMyInboxIRI)).Return(
			mustParse(testFederatedActorIRI2), nil)
		mockDB.EXPECT().Lock(ctx, mustParse(testFederatedActorIRI2)).Return(func() {}, nil)
		mockDB.EXPECT().Followers(ctx, mustParse(testFederatedActorIRI2)).Return(
			followers, nil)
		mockDB.EXPECT().Update(ctx, gomock.Any())
		mockDB.EXPECT().Lock(ctx, mustParse(testMyInboxIRI)).Return(func() {}, nil)
		mockDB.EXPECT().OutboxForInbox(ctx, mustParse(testMyInboxIRI)).Return(
			mustParse(testMyOutboxIRI), nil)
		f := newFollowFn()
		err := w.follow(ctx, f)
		if err != nil {
			t.Fatalf("got error %s", err)
		}
	})
	t.Run("OnFollowAutomaticallyRejectDelivers", func(t *testing.T) {
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		w, mockDB := setupFn(ctl)
		w.OnFollow = OnFollowAutomaticallyReject
		w.addNewIds = func(c context.Context, activity Activity) error {
			return nil
		}
		w.deliver = func(c context.Context, outboxIRI *url.URL, activity Activity) error {
			if !streams.IsOrExtendsActivityStreamsReject(activity) {
				t.Fatalf("expected Reject, got %T", activity)
			}
			return nil
		}
		mockDB.EXPECT().Lock(ctx, mustParse(testMyInboxIRI)).Return(func() {}, nil)
		mockDB.EXPECT().ActorForInbox(ctx, mustParse(testMyInboxIRI)).Return(
			mustParse(testFederatedActorIRI2), nil)
		mockDB.EXPECT().Lock(ctx, mustParse(testMyInboxIRI)).Return(func() {}, nil)
		mockDB.EXPECT().OutboxForInbox(ctx, mustParse(testMyInboxIRI)).Return(
			mustParse(testMyOutboxIRI), nil)
		f := newFollowFn()
		err := w.follow(ctx, f)
		if err != nil {
			t.Fatalf("got error %s", err)
		}
	})
	t.Run("CallsCustomCallback", func(t *testing.T) {
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		w, mockDB := setupFn(ctl)
		w.OnFollow = OnFollowDoNothing
		mockDB.EXPECT().Lock(ctx, mustParse(testMyInboxIRI)).Return(func() {}, nil)
		mockDB.EXPECT().ActorForInbox(ctx, mustParse(testMyInboxIRI)).Return(
			mustParse(testFederatedActorIRI2), nil)
		f := newFollowFn()
		var gotc context.Context
		var got vocab.ActivityStreamsFollow
		w.Follow = func(ctx context.Context, v vocab.ActivityStreamsFollow) error {
			gotc = ctx
			got = v
			return nil
		}
		err := w.follow(ctx, f)
		if err != nil {
			t.Fatalf("got error %s", err)
		}
		assertEqual(t, ctx, gotc)
		assertEqual(t, f, got)
	})
}

func TestFederatedAccept(t *testing.T) {
	newAcceptFn := func() vocab.ActivityStreamsAccept {
		c := streams.NewActivityStreamsAccept()
		id := streams.NewJSONLDIdProperty()
		id.Set(mustParse(testFederatedActivityIRI2))
		c.SetJSONLDId(id)
		actor := streams.NewActivityStreamsActorProperty()
		actor.AppendIRI(mustParse(testFederatedActorIRI))
		c.SetActivityStreamsActor(actor)
		op := streams.NewActivityStreamsObjectProperty()
		op.AppendActivityStreamsFollow(testFollow)
		c.SetActivityStreamsObject(op)
		return c
	}
	ctx := context.Background()
	setupFn := func(ctl *gomock.Controller) (w FederatingWrappedCallbacks, mockDB *MockDatabase, mockTp *MockTransport) {
		mockDB = NewMockDatabase(ctl)
		mockTp = NewMockTransport(ctl)
		w.inboxIRI = mustParse(testMyInboxIRI)
		w.db = mockDB
		w.newTransport = func(c context.Context, a *url.URL, s string) (Transport, error) {
			return mockTp, nil
		}
		return
	}
	t.Run("DoesNothingIfNoObjects", func(t *testing.T) {
		a := newAcceptFn()
		a.SetActivityStreamsObject(nil)
		var w FederatingWrappedCallbacks
		err := w.accept(ctx, a)
		if err != nil {
			t.Fatalf("got error %s", err)
		}
	})
	t.Run("DereferencesObjectIRI", func(t *testing.T) {
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		w, mockDB, mockTp := setupFn(ctl)
		followers := streams.NewActivityStreamsCollection()
		expectFollowers := streams.NewActivityStreamsCollection()
		expectItems := streams.NewActivityStreamsItemsProperty()
		expectItems.AppendIRI(mustParse(testFederatedActorIRI))
		expectFollowers.SetActivityStreamsItems(expectItems)
		mockDB.EXPECT().Lock(ctx, mustParse(testMyInboxIRI)).Return(func() {}, nil)
		mockDB.EXPECT().ActorForInbox(ctx, mustParse(testMyInboxIRI)).Return(
			mustParse(testFederatedActorIRI2), nil)
		mockTp.EXPECT().Dereference(ctx, mustParse(testFederatedActivityIRI)).Return(
			mustWrapInGETResponse(mustParse(testFederatedActivityIRI), testFollow), nil)
		mockDB.EXPECT().Lock(ctx, mustParse(testFederatedActivityIRI)).Return(func() {}, nil)
		mockDB.EXPECT().Get(ctx, mustParse(testFederatedActivityIRI)).Return(
			testFollow, nil)
		mockDB.EXPECT().Lock(ctx, mustParse(testFederatedActorIRI2)).Return(func() {}, nil)
		mockDB.EXPECT().Following(ctx, mustParse(testFederatedActorIRI2)).Return(
			followers, nil)
		mockDB.EXPECT().Update(ctx, expectFollowers)
		a := newAcceptFn()
		op := streams.NewActivityStreamsObjectProperty()
		op.AppendIRI(mustParse(testFederatedActivityIRI))
		a.SetActivityStreamsObject(op)
		err := w.accept(ctx, a)
		if err != nil {
			t.Fatalf("got error %s", err)
		}
	})
	t.Run("IgnoresNonFollowObjects", func(t *testing.T) {
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		w, mockDB, _ := setupFn(ctl)
		mockDB.EXPECT().Lock(ctx, mustParse(testMyInboxIRI)).Return(func() {}, nil)
		mockDB.EXPECT().ActorForInbox(ctx, mustParse(testMyInboxIRI)).Return(
			mustParse(testFederatedActorIRI2), nil)
		a := newAcceptFn()
		op := streams.NewActivityStreamsObjectProperty()
		op.AppendActivityStreamsListen(testListen)
		a.SetActivityStreamsObject(op)
		err := w.accept(ctx, a)
		if err != nil {
			t.Fatalf("got error %s", err)
		}
	})
	t.Run("IgnoresFollowObjectsNotContainingMe", func(t *testing.T) {
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		w, mockDB, _ := setupFn(ctl)
		mockDB.EXPECT().Lock(ctx, mustParse(testMyInboxIRI)).Return(func() {}, nil)
		mockDB.EXPECT().ActorForInbox(ctx, mustParse(testMyInboxIRI)).Return(
			mustParse(testFederatedActorIRI3), nil)
		a := newAcceptFn()
		err := w.accept(ctx, a)
		if err != nil {
			t.Fatalf("got error %s", err)
		}
	})
	t.Run("ErrorIfPeerLiedAboutOurFollowId", func(t *testing.T) {
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		w, mockDB, _ := setupFn(ctl)
		mockDB.EXPECT().Lock(ctx, mustParse(testMyInboxIRI)).Return(func() {}, nil)
		mockDB.EXPECT().ActorForInbox(ctx, mustParse(testMyInboxIRI)).Return(
			mustParse(testFederatedActorIRI2), nil)
		mockDB.EXPECT().Lock(ctx, mustParse(testFederatedActivityIRI)).Return(func() {}, nil)
		mockDB.EXPECT().Get(ctx, mustParse(testFederatedActivityIRI)).Return(
			testListen, nil)
		a := newAcceptFn()
		err := w.accept(ctx, a)
		if err == nil {
			t.Fatalf("expected error, got none")
		}
	})
	t.Run("UpdatesFollowingCollection", func(t *testing.T) {
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		w, mockDB, _ := setupFn(ctl)
		followers := streams.NewActivityStreamsCollection()
		expectFollowers := streams.NewActivityStreamsCollection()
		expectItems := streams.NewActivityStreamsItemsProperty()
		expectItems.AppendIRI(mustParse(testFederatedActorIRI))
		expectFollowers.SetActivityStreamsItems(expectItems)
		mockDB.EXPECT().Lock(ctx, mustParse(testMyInboxIRI)).Return(func() {}, nil)
		mockDB.EXPECT().ActorForInbox(ctx, mustParse(testMyInboxIRI)).Return(
			mustParse(testFederatedActorIRI2), nil)
		mockDB.EXPECT().Lock(ctx, mustParse(testFederatedActivityIRI)).Return(func() {}, nil)
		mockDB.EXPECT().Get(ctx, mustParse(testFederatedActivityIRI)).Return(
			testFollow, nil)
		mockDB.EXPECT().Lock(ctx, mustParse(testFederatedActorIRI2)).Return(func() {}, nil)
		mockDB.EXPECT().Following(ctx, mustParse(testFederatedActorIRI2)).Return(
			followers, nil)
		mockDB.EXPECT().Update(ctx, expectFollowers)
		a := newAcceptFn()
		err := w.accept(ctx, a)
		if err != nil {
			t.Fatalf("got error %s", err)
		}
	})
	t.Run("CallsCustomCallback", func(t *testing.T) {
		a := newAcceptFn()
		a.SetActivityStreamsObject(nil)
		var w FederatingWrappedCallbacks
		var gotc context.Context
		var got vocab.ActivityStreamsAccept
		w.Accept = func(ctx context.Context, v vocab.ActivityStreamsAccept) error {
			gotc = ctx
			got = v
			return nil
		}
		err := w.accept(ctx, a)
		if err != nil {
			t.Fatalf("got error %s", err)
		}
		assertEqual(t, ctx, gotc)
		assertEqual(t, a, got)
	})
}

func TestFederatedReject(t *testing.T) {
	ctx := context.Background()
	t.Run("CallsCustomCallback", func(t *testing.T) {
		r := streams.NewActivityStreamsReject()
		var w FederatingWrappedCallbacks
		var gotc context.Context
		var got vocab.ActivityStreamsReject
		w.Reject = func(ctx context.Context, v vocab.ActivityStreamsReject) error {
			gotc = ctx
			got = v
			return nil
		}
		err := w.reject(ctx, r)
		if err != nil {
			t.Fatalf("got error %s", err)
		}
		assertEqual(t, ctx, gotc)
		assertEqual(t, r, got)
	})
}

func TestFederatedAdd(t *testing.T) {
	newAddFn := func() vocab.ActivityStreamsAdd {
		a := streams.NewActivityStreamsAdd()
		id := streams.NewJSONLDIdProperty()
		id.Set(mustParse(testFederatedActivityIRI))
		a.SetJSONLDId(id)
		actor := streams.NewActivityStreamsActorProperty()
		actor.AppendIRI(mustParse(testFederatedActorIRI))
		a.SetActivityStreamsActor(actor)
		op := streams.NewActivityStreamsObjectProperty()
		op.AppendActivityStreamsNote(testFederatedNote)
		a.SetActivityStreamsObject(op)
		tp := streams.NewActivityStreamsTargetProperty()
		tp.AppendIRI(mustParse(testAudienceIRI))
		a.SetActivityStreamsTarget(tp)
		return a
	}
	ctx := context.Background()
	setupFn := func(ctl *gomock.Controller) (w FederatingWrappedCallbacks, mockDB *MockDatabase) {
		mockDB = NewMockDatabase(ctl)
		w.db = mockDB
		return
	}
	t.Run("ErrorIfNoObject", func(t *testing.T) {
		a := newAddFn()
		a.SetActivityStreamsObject(nil)
		var w FederatingWrappedCallbacks
		err := w.add(ctx, a)
		if err == nil {
			t.Fatalf("expected error, got none")
		}
	})
	t.Run("ErrorIfObjectLengthZero", func(t *testing.T) {
		a := newAddFn()
		a.GetActivityStreamsObject().Remove(0)
		var w FederatingWrappedCallbacks
		err := w.add(ctx, a)
		if err == nil {
			t.Fatalf("expected error, got none")
		}
	})
	t.Run("ErrorIfNoTarget", func(t *testing.T) {
		a := newAddFn()
		a.SetActivityStreamsTarget(nil)
		var w FederatingWrappedCallbacks
		err := w.add(ctx, a)
		if err == nil {
			t.Fatalf("expected error, got none")
		}
	})
	t.Run("ErrorIfTargetLengthZero", func(t *testing.T) {
		a := newAddFn()
		a.GetActivityStreamsTarget().Remove(0)
		var w FederatingWrappedCallbacks
		err := w.add(ctx, a)
		if err == nil {
			t.Fatalf("expected error, got none")
		}
	})
	t.Run("AddsAllObjectIdsToCollection", func(t *testing.T) {
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		w, mockDB := setupFn(ctl)
		col1 := streams.NewActivityStreamsCollection()
		expectCol1 := streams.NewActivityStreamsCollection()
		items1 := streams.NewActivityStreamsItemsProperty()
		items1.AppendIRI(mustParse(testNoteId1))
		items1.AppendIRI(mustParse(testNoteId2))
		expectCol1.SetActivityStreamsItems(items1)
		mockDB.EXPECT().Lock(ctx, mustParse(testAudienceIRI)).Return(func() {}, nil)
		mockDB.EXPECT().Owns(ctx, mustParse(testAudienceIRI)).Return(
			true, nil)
		mockDB.EXPECT().Get(ctx, mustParse(testAudienceIRI)).Return(
			col1, nil)
		mockDB.EXPECT().Update(ctx, expectCol1).Return(nil)
		a := newAddFn()
		a.GetActivityStreamsObject().AppendActivityStreamsNote(testFederatedNote2)
		err := w.add(ctx, a)
		if err != nil {
			t.Fatalf("got error %s", err)
		}
	})
	t.Run("AddsAllObjectIdsToOrderedCollection", func(t *testing.T) {
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		w, mockDB := setupFn(ctl)
		col1 := streams.NewActivityStreamsOrderedCollection()
		expectCol1 := streams.NewActivityStreamsOrderedCollection()
		items1 := streams.NewActivityStreamsOrderedItemsProperty()
		items1.AppendIRI(mustParse(testNoteId1))
		items1.AppendIRI(mustParse(testNoteId2))
		expectCol1.SetActivityStreamsOrderedItems(items1)
		mockDB.EXPECT().Lock(ctx, mustParse(testAudienceIRI)).Return(func() {}, nil)
		mockDB.EXPECT().Owns(ctx, mustParse(testAudienceIRI)).Return(
			true, nil)
		mockDB.EXPECT().Get(ctx, mustParse(testAudienceIRI)).Return(
			col1, nil)
		mockDB.EXPECT().Update(ctx, expectCol1).Return(nil)
		a := newAddFn()
		a.GetActivityStreamsObject().AppendActivityStreamsNote(testFederatedNote2)
		err := w.add(ctx, a)
		if err != nil {
			t.Fatalf("got error %s", err)
		}
	})
	t.Run("AddsAllObjectIdsToEachTarget", func(t *testing.T) {
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		w, mockDB := setupFn(ctl)
		col1 := streams.NewActivityStreamsCollection()
		expectCol1 := streams.NewActivityStreamsCollection()
		items1 := streams.NewActivityStreamsItemsProperty()
		items1.AppendIRI(mustParse(testNoteId1))
		expectCol1.SetActivityStreamsItems(items1)
		col2 := streams.NewActivityStreamsCollection()
		expectCol2 := streams.NewActivityStreamsCollection()
		items2 := streams.NewActivityStreamsItemsProperty()
		items2.AppendIRI(mustParse(testNoteId1))
		expectCol2.SetActivityStreamsItems(items2)
		mockDB.EXPECT().Lock(ctx, mustParse(testAudienceIRI)).Return(func() {}, nil)
		mockDB.EXPECT().Owns(ctx, mustParse(testAudienceIRI)).Return(
			true, nil)
		mockDB.EXPECT().Get(ctx, mustParse(testAudienceIRI)).Return(
			col1, nil)
		mockDB.EXPECT().Update(ctx, expectCol1).Return(nil)
		mockDB.EXPECT().Lock(ctx, mustParse(testAudienceIRI2)).Return(func() {}, nil)
		mockDB.EXPECT().Owns(ctx, mustParse(testAudienceIRI2)).Return(
			true, nil)
		mockDB.EXPECT().Get(ctx, mustParse(testAudienceIRI2)).Return(
			col2, nil)
		mockDB.EXPECT().Update(ctx, expectCol2).Return(nil)
		a := newAddFn()
		a.GetActivityStreamsTarget().AppendIRI(mustParse(testAudienceIRI2))
		err := w.add(ctx, a)
		if err != nil {
			t.Fatalf("got error %s", err)
		}
	})
	t.Run("ReturnsErrorIfTargetIsNotCollection", func(t *testing.T) {
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		w, mockDB := setupFn(ctl)
		notCol := streams.NewActivityStreamsNote()
		mockDB.EXPECT().Lock(ctx, mustParse(testAudienceIRI)).Return(func() {}, nil)
		mockDB.EXPECT().Owns(ctx, mustParse(testAudienceIRI)).Return(
			true, nil)
		mockDB.EXPECT().Get(ctx, mustParse(testAudienceIRI)).Return(
			notCol, nil)
		a := newAddFn()
		err := w.add(ctx, a)
		if err == nil {
			t.Fatalf("expected error, got none")
		}
	})
	t.Run("CallsCustomCallback", func(t *testing.T) {
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		w, mockDB := setupFn(ctl)
		col1 := streams.NewActivityStreamsCollection()
		mockDB.EXPECT().Lock(ctx, mustParse(testAudienceIRI)).Return(func() {}, nil)
		mockDB.EXPECT().Owns(ctx, mustParse(testAudienceIRI)).Return(
			true, nil)
		mockDB.EXPECT().Get(ctx, mustParse(testAudienceIRI)).Return(
			col1, nil)
		mockDB.EXPECT().Update(ctx, gomock.Any()).Return(nil)
		var gotc context.Context
		var got vocab.ActivityStreamsAdd
		w.Add = func(ctx context.Context, v vocab.ActivityStreamsAdd) error {
			gotc = ctx
			got = v
			return nil
		}
		a := newAddFn()
		err := w.add(ctx, a)
		if err != nil {
			t.Fatalf("got error %s", err)
		}
		assertEqual(t, ctx, gotc)
		assertEqual(t, a, got)
	})
}

func TestFederatedRemove(t *testing.T) {
	newRemoveFn := func() vocab.ActivityStreamsRemove {
		r := streams.NewActivityStreamsRemove()
		id := streams.NewJSONLDIdProperty()
		id.Set(mustParse(testFederatedActivityIRI))
		r.SetJSONLDId(id)
		actor := streams.NewActivityStreamsActorProperty()
		actor.AppendIRI(mustParse(testFederatedActorIRI))
		r.SetActivityStreamsActor(actor)
		op := streams.NewActivityStreamsObjectProperty()
		op.AppendActivityStreamsNote(testFederatedNote)
		r.SetActivityStreamsObject(op)
		tp := streams.NewActivityStreamsTargetProperty()
		tp.AppendIRI(mustParse(testAudienceIRI))
		r.SetActivityStreamsTarget(tp)
		return r
	}
	ctx := context.Background()
	setupFn := func(ctl *gomock.Controller) (w FederatingWrappedCallbacks, mockDB *MockDatabase) {
		mockDB = NewMockDatabase(ctl)
		w.db = mockDB
		return
	}
	t.Run("ErrorIfNoObject", func(t *testing.T) {
		r := newRemoveFn()
		r.SetActivityStreamsObject(nil)
		var w FederatingWrappedCallbacks
		err := w.remove(ctx, r)
		if err == nil {
			t.Fatalf("expected error, got none")
		}
	})
	t.Run("ErrorIfObjectLengthZero", func(t *testing.T) {
		r := newRemoveFn()
		r.GetActivityStreamsObject().Remove(0)
		var w FederatingWrappedCallbacks
		err := w.remove(ctx, r)
		if err == nil {
			t.Fatalf("expected error, got none")
		}
	})
	t.Run("ErrorIfNoTarget", func(t *testing.T) {
		r := newRemoveFn()
		r.SetActivityStreamsTarget(nil)
		var w FederatingWrappedCallbacks
		err := w.remove(ctx, r)
		if err == nil {
			t.Fatalf("expected error, got none")
		}
	})
	t.Run("ErrorIfTargetLengthZero", func(t *testing.T) {
		r := newRemoveFn()
		r.GetActivityStreamsTarget().Remove(0)
		var w FederatingWrappedCallbacks
		err := w.remove(ctx, r)
		if err == nil {
			t.Fatalf("expected error, got none")
		}
	})
	t.Run("RemovesAllObjectIdsFromCollection", func(t *testing.T) {
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		w, mockDB := setupFn(ctl)
		col1 := streams.NewActivityStreamsCollection()
		items := streams.NewActivityStreamsItemsProperty()
		items.AppendIRI(mustParse(testAudienceIRI))
		items.AppendIRI(mustParse(testAudienceIRI2))
		items.AppendIRI(mustParse(testNoteId1))
		items.AppendIRI(mustParse(testNoteId2))
		col1.SetActivityStreamsItems(items)
		expectCol1 := streams.NewActivityStreamsCollection()
		items1 := streams.NewActivityStreamsItemsProperty()
		items1.AppendIRI(mustParse(testAudienceIRI))
		items1.AppendIRI(mustParse(testAudienceIRI2))
		expectCol1.SetActivityStreamsItems(items1)
		mockDB.EXPECT().Lock(ctx, mustParse(testAudienceIRI)).Return(func() {}, nil)
		mockDB.EXPECT().Owns(ctx, mustParse(testAudienceIRI)).Return(
			true, nil)
		mockDB.EXPECT().Get(ctx, mustParse(testAudienceIRI)).Return(
			col1, nil)
		mockDB.EXPECT().Update(ctx, expectCol1).Return(nil)
		r := newRemoveFn()
		r.GetActivityStreamsObject().AppendActivityStreamsNote(testFederatedNote2)
		err := w.remove(ctx, r)
		if err != nil {
			t.Fatalf("got error %s", err)
		}
	})
	t.Run("RemovesAllObjectIdsFromOrderedCollection", func(t *testing.T) {
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		w, mockDB := setupFn(ctl)
		col1 := streams.NewActivityStreamsOrderedCollection()
		items := streams.NewActivityStreamsOrderedItemsProperty()
		items.AppendIRI(mustParse(testAudienceIRI))
		items.AppendIRI(mustParse(testAudienceIRI2))
		items.AppendIRI(mustParse(testNoteId1))
		items.AppendIRI(mustParse(testNoteId2))
		col1.SetActivityStreamsOrderedItems(items)
		expectCol1 := streams.NewActivityStreamsOrderedCollection()
		items1 := streams.NewActivityStreamsOrderedItemsProperty()
		items1.AppendIRI(mustParse(testAudienceIRI))
		items1.AppendIRI(mustParse(testAudienceIRI2))
		expectCol1.SetActivityStreamsOrderedItems(items1)
		mockDB.EXPECT().Lock(ctx, mustParse(testAudienceIRI)).Return(func() {}, nil)
		mockDB.EXPECT().Owns(ctx, mustParse(testAudienceIRI)).Return(
			true, nil)
		mockDB.EXPECT().Get(ctx, mustParse(testAudienceIRI)).Return(
			col1, nil)
		mockDB.EXPECT().Update(ctx, expectCol1).Return(nil)
		r := newRemoveFn()
		r.GetActivityStreamsObject().AppendActivityStreamsNote(testFederatedNote2)
		err := w.remove(ctx, r)
		if err != nil {
			t.Fatalf("got error %s", err)
		}
	})
	t.Run("RemovesAllObjectIdsFromEachTarget", func(t *testing.T) {
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		w, mockDB := setupFn(ctl)
		col1 := streams.NewActivityStreamsCollection()
		items := streams.NewActivityStreamsItemsProperty()
		items.AppendIRI(mustParse(testFederatedActorIRI3))
		items.AppendIRI(mustParse(testFederatedActorIRI4))
		items.AppendIRI(mustParse(testNoteId1))
		col1.SetActivityStreamsItems(items)
		expectCol1 := streams.NewActivityStreamsCollection()
		items1 := streams.NewActivityStreamsItemsProperty()
		items1.AppendIRI(mustParse(testFederatedActorIRI3))
		items1.AppendIRI(mustParse(testFederatedActorIRI4))
		expectCol1.SetActivityStreamsItems(items1)
		col2 := streams.NewActivityStreamsCollection()
		items0 := streams.NewActivityStreamsItemsProperty()
		items0.AppendIRI(mustParse(testFederatedActorIRI))
		items0.AppendIRI(mustParse(testNoteId1))
		items0.AppendIRI(mustParse(testFederatedActorIRI2))
		col2.SetActivityStreamsItems(items0)
		expectCol2 := streams.NewActivityStreamsCollection()
		items2 := streams.NewActivityStreamsItemsProperty()
		items2.AppendIRI(mustParse(testFederatedActorIRI))
		items2.AppendIRI(mustParse(testFederatedActorIRI2))
		expectCol2.SetActivityStreamsItems(items2)
		mockDB.EXPECT().Lock(ctx, mustParse(testAudienceIRI)).Return(func() {}, nil)
		mockDB.EXPECT().Owns(ctx, mustParse(testAudienceIRI)).Return(
			true, nil)
		mockDB.EXPECT().Get(ctx, mustParse(testAudienceIRI)).Return(
			col1, nil)
		mockDB.EXPECT().Update(ctx, expectCol1).Return(nil)
		mockDB.EXPECT().Lock(ctx, mustParse(testAudienceIRI2)).Return(func() {}, nil)
		mockDB.EXPECT().Owns(ctx, mustParse(testAudienceIRI2)).Return(
			true, nil)
		mockDB.EXPECT().Get(ctx, mustParse(testAudienceIRI2)).Return(
			col2, nil)
		mockDB.EXPECT().Update(ctx, expectCol2).Return(nil)
		r := newRemoveFn()
		r.GetActivityStreamsTarget().AppendIRI(mustParse(testAudienceIRI2))
		err := w.remove(ctx, r)
		if err != nil {
			t.Fatalf("got error %s", err)
		}
	})
	t.Run("ReturnsErrorIfTargetIsNotCollection", func(t *testing.T) {
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		w, mockDB := setupFn(ctl)
		notCol := streams.NewActivityStreamsNote()
		mockDB.EXPECT().Lock(ctx, mustParse(testAudienceIRI)).Return(func() {}, nil)
		mockDB.EXPECT().Owns(ctx, mustParse(testAudienceIRI)).Return(
			true, nil)
		mockDB.EXPECT().Get(ctx, mustParse(testAudienceIRI)).Return(
			notCol, nil)
		r := newRemoveFn()
		err := w.remove(ctx, r)
		if err == nil {
			t.Fatalf("expected error, got none")
		}
	})
	t.Run("CallsCustomCallback", func(t *testing.T) {
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		w, mockDB := setupFn(ctl)
		col1 := streams.NewActivityStreamsCollection()
		mockDB.EXPECT().Lock(ctx, mustParse(testAudienceIRI)).Return(func() {}, nil)
		mockDB.EXPECT().Owns(ctx, mustParse(testAudienceIRI)).Return(
			true, nil)
		mockDB.EXPECT().Get(ctx, mustParse(testAudienceIRI)).Return(
			col1, nil)
		mockDB.EXPECT().Update(ctx, gomock.Any()).Return(nil)
		var gotc context.Context
		var got vocab.ActivityStreamsRemove
		w.Remove = func(ctx context.Context, v vocab.ActivityStreamsRemove) error {
			gotc = ctx
			got = v
			return nil
		}
		r := newRemoveFn()
		err := w.remove(ctx, r)
		if err != nil {
			t.Fatalf("got error %s", err)
		}
		assertEqual(t, ctx, gotc)
		assertEqual(t, r, got)
	})
}

func TestFederatedLike(t *testing.T) {
	newLikeFn := func() vocab.ActivityStreamsLike {
		l := streams.NewActivityStreamsLike()
		id := streams.NewJSONLDIdProperty()
		id.Set(mustParse(testFederatedActivityIRI))
		l.SetJSONLDId(id)
		actor := streams.NewActivityStreamsActorProperty()
		actor.AppendIRI(mustParse(testFederatedActorIRI))
		l.SetActivityStreamsActor(actor)
		op := streams.NewActivityStreamsObjectProperty()
		op.AppendActivityStreamsNote(testFederatedNote)
		l.SetActivityStreamsObject(op)
		return l
	}
	ctx := context.Background()
	setupFn := func(ctl *gomock.Controller) (w FederatingWrappedCallbacks, mockDB *MockDatabase) {
		mockDB = NewMockDatabase(ctl)
		w.db = mockDB
		return
	}
	t.Run("ErrorIfNoObject", func(t *testing.T) {
		l := newLikeFn()
		l.SetActivityStreamsObject(nil)
		var w FederatingWrappedCallbacks
		err := w.like(ctx, l)
		if err == nil {
			t.Fatalf("expected error, got none")
		}
	})
	t.Run("ErrorIfObjectLengthZero", func(t *testing.T) {
		l := newLikeFn()
		l.GetActivityStreamsObject().Remove(0)
		var w FederatingWrappedCallbacks
		err := w.like(ctx, l)
		if err == nil {
			t.Fatalf("expected error, got none")
		}
	})
	t.Run("SkipsUnownedObjects", func(t *testing.T) {
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		w, mockDB := setupFn(ctl)
		mockDB.EXPECT().Lock(ctx, mustParse(testNoteId1)).Return(func() {}, nil)
		mockDB.EXPECT().Owns(ctx, mustParse(testNoteId1)).Return(false, nil)
		l := newLikeFn()
		err := w.like(ctx, l)
		if err != nil {
			t.Fatalf("got error %s", err)
		}
	})
	t.Run("AddsToNewLikesCollection", func(t *testing.T) {
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		w, mockDB := setupFn(ctl)
		note := streams.NewActivityStreamsNote()
		expectNote := streams.NewActivityStreamsNote()
		expectLikes := streams.NewActivityStreamsLikesProperty()
		expectCol := streams.NewActivityStreamsCollection()
		expectItems := streams.NewActivityStreamsItemsProperty()
		expectItems.AppendIRI(mustParse(testFederatedActivityIRI))
		expectCol.SetActivityStreamsItems(expectItems)
		expectLikes.SetActivityStreamsCollection(expectCol)
		expectNote.SetActivityStreamsLikes(expectLikes)
		mockDB.EXPECT().Lock(ctx, mustParse(testNoteId1)).Return(func() {}, nil)
		mockDB.EXPECT().Owns(ctx, mustParse(testNoteId1)).Return(true, nil)
		mockDB.EXPECT().Get(ctx, mustParse(testNoteId1)).Return(
			note, nil)
		mockDB.EXPECT().Update(ctx, expectNote).Return(nil)
		l := newLikeFn()
		err := w.like(ctx, l)
		if err != nil {
			t.Fatalf("got error %s", err)
		}
	})
	t.Run("AddsToExistingLikesCollection", func(t *testing.T) {
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		w, mockDB := setupFn(ctl)
		note := streams.NewActivityStreamsNote()
		likes := streams.NewActivityStreamsLikesProperty()
		col := streams.NewActivityStreamsCollection()
		items := streams.NewActivityStreamsItemsProperty()
		items.AppendIRI(mustParse(testFederatedActivityIRI2))
		col.SetActivityStreamsItems(items)
		likes.SetActivityStreamsCollection(col)
		note.SetActivityStreamsLikes(likes)
		expectNote := streams.NewActivityStreamsNote()
		expectLikes := streams.NewActivityStreamsLikesProperty()
		expectCol := streams.NewActivityStreamsCollection()
		expectItems := streams.NewActivityStreamsItemsProperty()
		expectItems.AppendIRI(mustParse(testFederatedActivityIRI))
		expectItems.AppendIRI(mustParse(testFederatedActivityIRI2))
		expectCol.SetActivityStreamsItems(expectItems)
		expectLikes.SetActivityStreamsCollection(expectCol)
		expectNote.SetActivityStreamsLikes(expectLikes)
		mockDB.EXPECT().Lock(ctx, mustParse(testNoteId1)).Return(func() {}, nil)
		mockDB.EXPECT().Owns(ctx, mustParse(testNoteId1)).Return(true, nil)
		mockDB.EXPECT().Get(ctx, mustParse(testNoteId1)).Return(
			note, nil)
		mockDB.EXPECT().Update(ctx, expectNote).Return(nil)
		l := newLikeFn()
		err := w.like(ctx, l)
		if err != nil {
			t.Fatalf("got error %s", err)
		}
	})
	t.Run("AddsToExistingLikesOrderedCollection", func(t *testing.T) {
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		w, mockDB := setupFn(ctl)
		note := streams.NewActivityStreamsNote()
		likes := streams.NewActivityStreamsLikesProperty()
		col := streams.NewActivityStreamsOrderedCollection()
		items := streams.NewActivityStreamsOrderedItemsProperty()
		items.AppendIRI(mustParse(testFederatedActivityIRI2))
		col.SetActivityStreamsOrderedItems(items)
		likes.SetActivityStreamsOrderedCollection(col)
		note.SetActivityStreamsLikes(likes)
		expectNote := streams.NewActivityStreamsNote()
		expectLikes := streams.NewActivityStreamsLikesProperty()
		expectCol := streams.NewActivityStreamsOrderedCollection()
		expectItems := streams.NewActivityStreamsOrderedItemsProperty()
		expectItems.AppendIRI(mustParse(testFederatedActivityIRI))
		expectItems.AppendIRI(mustParse(testFederatedActivityIRI2))
		expectCol.SetActivityStreamsOrderedItems(expectItems)
		expectLikes.SetActivityStreamsOrderedCollection(expectCol)
		expectNote.SetActivityStreamsLikes(expectLikes)
		mockDB.EXPECT().Lock(ctx, mustParse(testNoteId1)).Return(func() {}, nil)
		mockDB.EXPECT().Owns(ctx, mustParse(testNoteId1)).Return(true, nil)
		mockDB.EXPECT().Get(ctx, mustParse(testNoteId1)).Return(
			note, nil)
		mockDB.EXPECT().Update(ctx, expectNote).Return(nil)
		l := newLikeFn()
		err := w.like(ctx, l)
		if err != nil {
			t.Fatalf("got error %s", err)
		}
	})
	t.Run("CallsCustomCallback", func(t *testing.T) {
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		w, mockDB := setupFn(ctl)
		note := streams.NewActivityStreamsNote()
		expectNote := streams.NewActivityStreamsNote()
		expectLikes := streams.NewActivityStreamsLikesProperty()
		expectCol := streams.NewActivityStreamsCollection()
		expectItems := streams.NewActivityStreamsItemsProperty()
		expectItems.AppendIRI(mustParse(testFederatedActivityIRI))
		expectCol.SetActivityStreamsItems(expectItems)
		expectLikes.SetActivityStreamsCollection(expectCol)
		expectNote.SetActivityStreamsLikes(expectLikes)
		mockDB.EXPECT().Lock(ctx, mustParse(testNoteId1)).Return(func() {}, nil)
		mockDB.EXPECT().Owns(ctx, mustParse(testNoteId1)).Return(true, nil)
		mockDB.EXPECT().Get(ctx, mustParse(testNoteId1)).Return(
			note, nil)
		mockDB.EXPECT().Update(ctx, expectNote).Return(nil)
		var gotc context.Context
		var got vocab.ActivityStreamsLike
		w.Like = func(ctx context.Context, v vocab.ActivityStreamsLike) error {
			gotc = ctx
			got = v
			return nil
		}
		l := newLikeFn()
		err := w.like(ctx, l)
		if err != nil {
			t.Fatalf("got error %s", err)
		}
		assertEqual(t, ctx, gotc)
		assertEqual(t, l, got)
	})
}

func TestFederatedAnnounce(t *testing.T) {
	newAnnounceFn := func() vocab.ActivityStreamsAnnounce {
		a := streams.NewActivityStreamsAnnounce()
		id := streams.NewJSONLDIdProperty()
		id.Set(mustParse(testFederatedActivityIRI))
		a.SetJSONLDId(id)
		actor := streams.NewActivityStreamsActorProperty()
		actor.AppendIRI(mustParse(testFederatedActorIRI))
		a.SetActivityStreamsActor(actor)
		op := streams.NewActivityStreamsObjectProperty()
		op.AppendActivityStreamsNote(testFederatedNote)
		a.SetActivityStreamsObject(op)
		return a
	}
	ctx := context.Background()
	setupFn := func(ctl *gomock.Controller) (w FederatingWrappedCallbacks, mockDB *MockDatabase) {
		mockDB = NewMockDatabase(ctl)
		w.db = mockDB
		return
	}
	t.Run("DoesNothingWhenNoObjects", func(t *testing.T) {
		a := newAnnounceFn()
		a.SetActivityStreamsObject(nil)
		var w FederatingWrappedCallbacks
		err := w.announce(ctx, a)
		if err != nil {
			t.Fatalf("got error %s", err)
		}
	})
	t.Run("SkipsUnownedObjects", func(t *testing.T) {
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		w, mockDB := setupFn(ctl)
		mockDB.EXPECT().Lock(ctx, mustParse(testNoteId1)).Return(func() {}, nil)
		mockDB.EXPECT().Owns(ctx, mustParse(testNoteId1)).Return(
			false, nil)
		a := newAnnounceFn()
		err := w.announce(ctx, a)
		if err != nil {
			t.Fatalf("got error %s", err)
		}
	})
	t.Run("AddsToNewSharesCollection", func(t *testing.T) {
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		w, mockDB := setupFn(ctl)
		note := streams.NewActivityStreamsNote()
		expectNote := streams.NewActivityStreamsNote()
		expectShares := streams.NewActivityStreamsSharesProperty()
		expectCol := streams.NewActivityStreamsCollection()
		expectItems := streams.NewActivityStreamsItemsProperty()
		expectItems.AppendIRI(mustParse(testFederatedActivityIRI))
		expectCol.SetActivityStreamsItems(expectItems)
		expectShares.SetActivityStreamsCollection(expectCol)
		expectNote.SetActivityStreamsShares(expectShares)
		mockDB.EXPECT().Lock(ctx, mustParse(testNoteId1)).Return(func() {}, nil)
		mockDB.EXPECT().Owns(ctx, mustParse(testNoteId1)).Return(
			true, nil)
		mockDB.EXPECT().Get(ctx, mustParse(testNoteId1)).Return(
			note, nil)
		mockDB.EXPECT().Update(ctx, expectNote).Return(nil)
		a := newAnnounceFn()
		err := w.announce(ctx, a)
		if err != nil {
			t.Fatalf("got error %s", err)
		}
	})
	t.Run("AddsToExistingSharesCollection", func(t *testing.T) {
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		w, mockDB := setupFn(ctl)
		note := streams.NewActivityStreamsNote()
		shares := streams.NewActivityStreamsSharesProperty()
		col := streams.NewActivityStreamsCollection()
		items := streams.NewActivityStreamsItemsProperty()
		items.AppendIRI(mustParse(testFederatedActivityIRI2))
		col.SetActivityStreamsItems(items)
		shares.SetActivityStreamsCollection(col)
		note.SetActivityStreamsShares(shares)
		expectNote := streams.NewActivityStreamsNote()
		expectShares := streams.NewActivityStreamsSharesProperty()
		expectCol := streams.NewActivityStreamsCollection()
		expectItems := streams.NewActivityStreamsItemsProperty()
		expectItems.AppendIRI(mustParse(testFederatedActivityIRI))
		expectItems.AppendIRI(mustParse(testFederatedActivityIRI2))
		expectCol.SetActivityStreamsItems(expectItems)
		expectShares.SetActivityStreamsCollection(expectCol)
		expectNote.SetActivityStreamsShares(expectShares)
		mockDB.EXPECT().Lock(ctx, mustParse(testNoteId1)).Return(func() {}, nil)
		mockDB.EXPECT().Owns(ctx, mustParse(testNoteId1)).Return(
			true, nil)
		mockDB.EXPECT().Get(ctx, mustParse(testNoteId1)).Return(
			note, nil)
		mockDB.EXPECT().Update(ctx, expectNote).Return(nil)
		a := newAnnounceFn()
		err := w.announce(ctx, a)
		if err != nil {
			t.Fatalf("got error %s", err)
		}
	})
	t.Run("AddsToExistingSharesOrderedCollection", func(t *testing.T) {
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		w, mockDB := setupFn(ctl)
		note := streams.NewActivityStreamsNote()
		shares := streams.NewActivityStreamsSharesProperty()
		col := streams.NewActivityStreamsOrderedCollection()
		items := streams.NewActivityStreamsOrderedItemsProperty()
		items.AppendIRI(mustParse(testFederatedActivityIRI2))
		col.SetActivityStreamsOrderedItems(items)
		shares.SetActivityStreamsOrderedCollection(col)
		note.SetActivityStreamsShares(shares)
		expectNote := streams.NewActivityStreamsNote()
		expectShares := streams.NewActivityStreamsSharesProperty()
		expectCol := streams.NewActivityStreamsOrderedCollection()
		expectItems := streams.NewActivityStreamsOrderedItemsProperty()
		expectItems.AppendIRI(mustParse(testFederatedActivityIRI))
		expectItems.AppendIRI(mustParse(testFederatedActivityIRI2))
		expectCol.SetActivityStreamsOrderedItems(expectItems)
		expectShares.SetActivityStreamsOrderedCollection(expectCol)
		expectNote.SetActivityStreamsShares(expectShares)
		mockDB.EXPECT().Lock(ctx, mustParse(testNoteId1)).Return(func() {}, nil)
		mockDB.EXPECT().Owns(ctx, mustParse(testNoteId1)).Return(
			true, nil)
		mockDB.EXPECT().Get(ctx, mustParse(testNoteId1)).Return(
			note, nil)
		mockDB.EXPECT().Update(ctx, expectNote).Return(nil)
		a := newAnnounceFn()
		err := w.announce(ctx, a)
		if err != nil {
			t.Fatalf("got error %s", err)
		}
	})
	t.Run("CallsCustomCallback", func(t *testing.T) {
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		w, mockDB := setupFn(ctl)
		note := streams.NewActivityStreamsNote()
		mockDB.EXPECT().Lock(ctx, mustParse(testNoteId1)).Return(func() {}, nil)
		mockDB.EXPECT().Owns(ctx, mustParse(testNoteId1)).Return(
			true, nil)
		mockDB.EXPECT().Get(ctx, mustParse(testNoteId1)).Return(
			note, nil)
		mockDB.EXPECT().Update(ctx, gomock.Any()).Return(nil)
		var gotc context.Context
		var got vocab.ActivityStreamsAnnounce
		w.Announce = func(ctx context.Context, v vocab.ActivityStreamsAnnounce) error {
			gotc = ctx
			got = v
			return nil
		}
		a := newAnnounceFn()
		err := w.announce(ctx, a)
		if err != nil {
			t.Fatalf("got error %s", err)
		}
		assertEqual(t, ctx, gotc)
		assertEqual(t, a, got)
	})
}

func TestFederatedUndo(t *testing.T) {
	newUndoFn := func() vocab.ActivityStreamsUndo {
		u := streams.NewActivityStreamsUndo()
		id := streams.NewJSONLDIdProperty()
		id.Set(mustParse(testFederatedActivityIRI2))
		u.SetJSONLDId(id)
		actor := streams.NewActivityStreamsActorProperty()
		actor.AppendIRI(mustParse(testFederatedActorIRI))
		u.SetActivityStreamsActor(actor)
		op := streams.NewActivityStreamsObjectProperty()
		op.AppendActivityStreamsListen(testListen)
		u.SetActivityStreamsObject(op)
		return u
	}
	ctx := context.Background()
	setupFn := func(ctl *gomock.Controller) (w FederatingWrappedCallbacks, mockTp *MockTransport) {
		mockTp = NewMockTransport(ctl)
		w.inboxIRI = mustParse(testMyInboxIRI)
		w.newTransport = func(c context.Context, a *url.URL, s string) (Transport, error) {
			return mockTp, nil
		}
		return
	}
	t.Run("ErrorIfNoObject", func(t *testing.T) {
		u := newUndoFn()
		u.SetActivityStreamsObject(nil)
		var w FederatingWrappedCallbacks
		err := w.undo(ctx, u)
		if err == nil {
			t.Fatalf("expected error, got none")
		}
	})
	t.Run("ErrorIfObjectLengthZero", func(t *testing.T) {
		u := newUndoFn()
		u.GetActivityStreamsObject().Remove(0)
		var w FederatingWrappedCallbacks
		err := w.undo(ctx, u)
		if err == nil {
			t.Fatalf("expected error, got none")
		}
	})
	t.Run("ErrorIfActorMismatch", func(t *testing.T) {
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		w, mockTp := setupFn(ctl)
		mockTp.EXPECT().Dereference(ctx, mustParse(testFederatedActivityIRI)).Return(
			mustWrapInGETResponse(mustParse(testFederatedActivityIRI), testListen), nil)
		u := newUndoFn()
		actor := streams.NewActivityStreamsActorProperty()
		actor.AppendIRI(mustParse(testFederatedActorIRI2))
		u.SetActivityStreamsActor(actor)
		err := w.undo(ctx, u)
		if err == nil {
			t.Fatalf("expected error, got none")
		}
	})
	t.Run("ErrorIfActorMismatchWhenDereferencingIRI", func(t *testing.T) {
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		w, mockTp := setupFn(ctl)
		mockTp.EXPECT().Dereference(ctx, mustParse(testFederatedActivityIRI)).Return(
			mustWrapInGETResponse(mustParse(testFederatedActivityIRI), testFollow), nil)
		u := newUndoFn()
		op := streams.NewActivityStreamsObjectProperty()
		op.AppendIRI(mustParse(testFederatedActivityIRI))
		u.SetActivityStreamsObject(op)
		err := w.undo(ctx, u)
		if err == nil {
			t.Fatalf("expected error, got none")
		}
	})
	t.Run("DereferencesWhenUndoValue", func(t *testing.T) {
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		w, mockTp := setupFn(ctl)
		mockTp.EXPECT().Dereference(ctx, mustParse(testFederatedActivityIRI)).Return(
			mustWrapInGETResponse(mustParse(testFederatedActivityIRI), testListen), nil)
		u := newUndoFn()
		err := w.undo(ctx, u)
		if err != nil {
			t.Fatalf("got error %s", err)
		}
	})
	t.Run("DereferencesWhenUndoIRI", func(t *testing.T) {
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		w, mockTp := setupFn(ctl)
		mockTp.EXPECT().Dereference(ctx, mustParse(testFederatedActivityIRI)).Return(
			mustWrapInGETResponse(mustParse(testFederatedActivityIRI), testListen), nil)
		u := newUndoFn()
		op := streams.NewActivityStreamsObjectProperty()
		op.AppendIRI(mustParse(testFederatedActivityIRI))
		u.SetActivityStreamsObject(op)
		err := w.undo(ctx, u)
		if err != nil {
			t.Fatalf("got error %s", err)
		}
	})
	t.Run("CallsCustomCallback", func(t *testing.T) {
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		w, mockTp := setupFn(ctl)
		mockTp.EXPECT().Dereference(ctx, mustParse(testFederatedActivityIRI)).Return(
			mustWrapInGETResponse(mustParse(testFederatedActivityIRI), testListen), nil)
		var gotc context.Context
		var got vocab.ActivityStreamsUndo
		w.Undo = func(ctx context.Context, v vocab.ActivityStreamsUndo) error {
			gotc = ctx
			got = v
			return nil
		}
		u := newUndoFn()
		err := w.undo(ctx, u)
		if err != nil {
			t.Fatalf("got error %s", err)
		}
		assertEqual(t, ctx, gotc)
		assertEqual(t, u, got)
	})
}

func TestFederatedBlock(t *testing.T) {
	newBlockFn := func() vocab.ActivityStreamsBlock {
		b := streams.NewActivityStreamsBlock()
		id := streams.NewJSONLDIdProperty()
		id.Set(mustParse(testFederatedActivityIRI))
		b.SetJSONLDId(id)
		actor := streams.NewActivityStreamsActorProperty()
		actor.AppendIRI(mustParse(testFederatedActorIRI))
		b.SetActivityStreamsActor(actor)
		op := streams.NewActivityStreamsObjectProperty()
		op.AppendIRI(mustParse(testFederatedActorIRI2))
		b.SetActivityStreamsObject(op)
		return b
	}
	ctx := context.Background()
	t.Run("ErrorIfNoObject", func(t *testing.T) {
		b := newBlockFn()
		b.SetActivityStreamsObject(nil)
		var w FederatingWrappedCallbacks
		err := w.block(ctx, b)
		if err == nil {
			t.Fatalf("expected error, got none")
		}
	})
	t.Run("ErrorIfObjectLengthZero", func(t *testing.T) {
		b := newBlockFn()
		b.GetActivityStreamsObject().Remove(0)
		var w FederatingWrappedCallbacks
		err := w.block(ctx, b)
		if err == nil {
			t.Fatalf("expected error, got none")
		}
	})
	t.Run("CallsCustomCallback", func(t *testing.T) {
		var w FederatingWrappedCallbacks
		var gotc context.Context
		var got vocab.ActivityStreamsBlock
		w.Block = func(ctx context.Context, v vocab.ActivityStreamsBlock) error {
			gotc = ctx
			got = v
			return nil
		}
		b := newBlockFn()
		err := w.block(ctx, b)
		if err != nil {
			t.Fatalf("got error %s", err)
		}
		assertEqual(t, ctx, gotc)
		assertEqual(t, b, got)
	})
}

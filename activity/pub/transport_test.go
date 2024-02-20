package pub

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/golang/mock/gomock"
)

const (
	testAppAgent = "testApp"
	testPubKeyId = "myPubKeyId"
)

var (
	testPrivKey    = []byte("some private key")
	testRespBody   = []byte("test resp body")
	httpSigSetupFn = func(ctl *gomock.Controller) (t *HttpSigTransport, c *MockClock, hc *MockHttpClient, gs, ps *MockSigner) {
		c = NewMockClock(ctl)
		hc = NewMockHttpClient(ctl)
		gs = NewMockSigner(ctl)
		ps = NewMockSigner(ctl)
		t = NewHttpSigTransport(
			hc,
			testAppAgent,
			c,
			gs,
			ps,
			testPubKeyId,
			testPrivKey)
		return
	}
)

func TestHttpSigTransportDereference(t *testing.T) {
	ctx := context.Background()
	t.Run("ReturnsErrorWhenHTTPStatusError", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		tp, c, hc, gs, _ := httpSigSetupFn(ctl)
		resp := &http.Response{}
		testErr := fmt.Errorf("test error")
		// Mock
		c.EXPECT().Now().Return(now())
		gs.EXPECT().SignRequest(testPrivKey, testPubKeyId, gomock.Any(), nil)
		hc.EXPECT().Do(gomock.Any()).Return(resp, testErr)
		// Run & Verify
		resp, err := tp.Dereference(ctx, mustParse(testNoteId1))
		assertEqual(t, resp, nil)
		assertEqual(t, err, testErr)
	})
	t.Run("Dereferences", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		tp, c, hc, gs, _ := httpSigSetupFn(ctl)
		expectReq, err := http.NewRequest("GET", testNoteId1, nil)
		assertEqual(t, err, nil)
		expectReq = expectReq.WithContext(ctx)
		expectReq.Header.Add(acceptHeader, acceptHeaderValue)
		expectReq.Header.Add("Accept-Charset", "utf-8")
		expectReq.Header.Add("Date", nowDateHeader())
		expectReq.Header.Add("User-Agent", fmt.Sprintf("%s %s", testAppAgent, goFedUserAgent()))
		respR := httptest.NewRecorder()
		respR.Write(testRespBody)
		resp := respR.Result()
		// Mock
		c.EXPECT().Now().Return(now())
		gs.EXPECT().SignRequest(testPrivKey, testPubKeyId, expectReq, nil)
		hc.EXPECT().Do(expectReq).Return(resp, nil)
		// Run & Verify
		resp, err = tp.Dereference(ctx, mustParse(testNoteId1))
		assertEqual(t, err, nil)
		assertEqual(t, resp.StatusCode, http.StatusOK)
		b, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		assertByteEqual(t, b, testRespBody)
		assertEqual(t, err, nil)
	})
}

func TestHttpSigTransportDeliver(t *testing.T) {
	ctx := context.Background()
	t.Run("ReturnsErrorWhenHTTPStatusError", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		tp, c, hc, _, ps := httpSigSetupFn(ctl)
		resp := &http.Response{}
		testErr := fmt.Errorf("test error")
		// Mock
		c.EXPECT().Now().Return(now())
		ps.EXPECT().SignRequest(testPrivKey, testPubKeyId, gomock.Any(), gomock.Any())
		hc.EXPECT().Do(gomock.Any()).Return(resp, testErr)
		// Run & Verify
		err := tp.Deliver(ctx, testRespBody, mustParse(testNoteId1))
		assertEqual(t, err, testErr)
	})
	t.Run("Delivers", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		tp, c, hc, _, ps := httpSigSetupFn(ctl)
		// gomock cannot handle http.NewRequest w/ Body differences.
		respR := httptest.NewRecorder()
		respR.WriteHeader(http.StatusOK)
		resp := respR.Result()
		// Mock
		c.EXPECT().Now().Return(now())
		ps.EXPECT().SignRequest(testPrivKey, testPubKeyId, gomock.Any(), testRespBody)
		hc.EXPECT().Do(gomock.Any()).Return(resp, nil)
		// Run & Verify
		err := tp.Deliver(ctx, testRespBody, mustParse(testFederatedActorIRI))
		assertEqual(t, err, nil)
	})
}

func TestHttpSigTransportBatchDeliver(t *testing.T) {
	ctx := context.Background()
	t.Run("BatchDelivers", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		tp, c, hc, _, ps := httpSigSetupFn(ctl)
		// gomock cannot handle http.NewRequest w/ Body differences.
		respR := httptest.NewRecorder()
		respR.WriteHeader(http.StatusOK)
		resp := respR.Result()
		// Mock
		c.EXPECT().Now().Return(now()).Times(2)
		ps.EXPECT().SignRequest(testPrivKey, testPubKeyId, gomock.Any(), testRespBody).Times(2)
		hc.EXPECT().Do(gomock.Any()).Return(resp, nil).Times(2)
		// Run & Verify
		err := tp.BatchDeliver(ctx, testRespBody, []*url.URL{mustParse(testFederatedActorIRI), mustParse(testFederatedActorIRI2)})
		assertEqual(t, err, nil)
	})
	t.Run("ReturnsErrorWhenOneErrors", func(t *testing.T) {
		// Setup
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		tp, c, hc, _, ps := httpSigSetupFn(ctl)
		// gomock cannot handle http.NewRequest w/ Body differences.
		respR := httptest.NewRecorder()
		respR.WriteHeader(http.StatusOK)
		resp := respR.Result()
		errResp := &http.Response{}
		testErr := fmt.Errorf("test error")
		// Mock
		c.EXPECT().Now().Return(now()).Times(2)
		ps.EXPECT().SignRequest(testPrivKey, testPubKeyId, gomock.Any(), testRespBody).Times(2)
		first := hc.EXPECT().Do(gomock.Any()).Return(resp, nil)
		hc.EXPECT().Do(gomock.Any()).Return(errResp, testErr).After(first)
		// Run & Verify
		err := tp.BatchDeliver(ctx, testRespBody, []*url.URL{mustParse(testFederatedActorIRI), mustParse(testFederatedActorIRI2)})
		assertNotEqual(t, err, nil)

	})
}

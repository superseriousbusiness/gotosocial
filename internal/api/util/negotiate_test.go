package util

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

type testMIMES []string

func (tm testMIMES) String(t *testing.T) string {
	t.Helper()

	res := tm.StringS(t)
	return strings.Join(res, ",")
}

func (tm testMIMES) StringS(t *testing.T) []string {
	t.Helper()

	res := make([]string, 0, len(tm))
	for _, m := range tm {
		res = append(res, string(m))
	}
	return res
}

func TestNegotiateFormat(t *testing.T) {
	tests := []struct {
		incoming []string
		offered  testMIMES
		format   string
	}{
		{incoming: testMIMES{AppJSON}.StringS(t), offered: testMIMES{AppJRDJSON, AppJSON}, format: "application/json"},
		{incoming: testMIMES{AppJRDJSON}.StringS(t), offered: testMIMES{AppJRDJSON, AppJSON}, format: "application/jrd+json"},
		{incoming: testMIMES{AppJRDJSON, AppJSON}.StringS(t), offered: testMIMES{AppJRDJSON}, format: "application/jrd+json"},
		{incoming: testMIMES{AppJRDJSON, AppJSON}.StringS(t), offered: testMIMES{AppJSON}, format: "application/json"},
		{incoming: testMIMES{"text/html,application/xhtml+xml,application/xml;q=0.9;q=0.8"}.StringS(t), offered: testMIMES{AppJSON, AppXML}, format: "application/xml"},
		{incoming: testMIMES{"text/html,application/xhtml+xml,application/xml;q=0.9;q=0.8"}.StringS(t), offered: testMIMES{TextHTML, AppXML}, format: "text/html"},
	}

	for _, tt := range tests {
		name := "incoming:" + strings.Join(tt.incoming, ",") + " offered:" + tt.offered.String(t)
		t.Run(name, func(t *testing.T) {
			tt := tt
			t.Parallel()

			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = &http.Request{
				Header: make(http.Header),
			}
			for _, header := range tt.incoming {
				c.Request.Header.Add("accept", header)
			}

			format := NegotiateFormat(c, tt.offered.StringS(t)...)
			if tt.format != format {
				t.Fatalf("expected format: '%s', got format: '%s'", tt.format, format)
			}
		})
	}
}

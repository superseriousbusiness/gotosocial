package gtserror_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/log"
)

func TestResponseError(t *testing.T) {
	testResponseError(t, http.Response{
		Body: toBody(`{"error": "user not found"}`),
		Request: &http.Request{
			Method: "GET",
			URL:    toURL("https://google.com/users/sundar"),
		},
		Status: "404 Not Found",
	})
	testResponseError(t, http.Response{
		Body: toBody("Unauthorized"),
		Request: &http.Request{
			Method: "POST",
			URL:    toURL("https://google.com/inbox"),
		},
		Status: "401 Unauthorized",
	})
	testResponseError(t, http.Response{
		Body: toBody(""),
		Request: &http.Request{
			Method: "GET",
			URL:    toURL("https://google.com/users/sundar"),
		},
		Status: "404 Not Found",
	})
}

func testResponseError(t *testing.T, rsp http.Response) {
	var body string
	if rsp.Body == http.NoBody {
		body = "<empty>"
	} else {
		var b []byte
		rsp.Body, b = copyBody(rsp.Body)
		trunc := len(b)
		if trunc > 256 {
			trunc = 256
		}
		body = string(b[:trunc])
	}
	expect := fmt.Sprintf(
		"%s%s request to %s failed: status=\"%s\" body=\"%s\"",
		func() string {
			if gtserror.Caller {
				return strings.Split(log.Caller(3), ".")[1] + ": "
			}
			return ""
		}(),
		rsp.Request.Method,
		rsp.Request.URL.String(),
		rsp.Status,
		body,
	)
	err := gtserror.NewFromResponse(&rsp)
	if str := err.Error(); str != expect {
		t.Errorf("unexpected error string: recv=%q expct=%q", str, expect)
	}
}

func toURL(u string) *url.URL {
	url, err := url.Parse(u)
	if err != nil {
		panic(err)
	}
	return url
}

func toBody(s string) io.ReadCloser {
	if s == "" {
		return http.NoBody
	}
	r := strings.NewReader(s)
	return io.NopCloser(r)
}

func copyBody(rc io.ReadCloser) (io.ReadCloser, []byte) {
	b, err := io.ReadAll(rc)
	if err != nil {
		panic(err)
	}
	r := bytes.NewReader(b)
	return io.NopCloser(r), b
}

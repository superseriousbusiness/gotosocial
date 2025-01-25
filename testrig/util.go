// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package testrig

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/url"
	"os"
	"path"
	"time"

	"codeberg.org/gruf/go-byteutil"
	"codeberg.org/gruf/go-kv/format"
	"github.com/superseriousbusiness/gotosocial/internal/filter/visibility"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	tlprocessor "github.com/superseriousbusiness/gotosocial/internal/processing/timeline"
	"github.com/superseriousbusiness/gotosocial/internal/processing/workers"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/timeline"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
)

// Starts workers on the provided state using noop processing functions.
// Useful when you *don't* want to trigger side effects in a test.
func StartNoopWorkers(state *state.State) {
	state.Workers.Client.Process = func(ctx context.Context, msg *messages.FromClientAPI) error { return nil }
	state.Workers.Federator.Process = func(ctx context.Context, msg *messages.FromFediAPI) error { return nil }

	state.Workers.Client.Init(messages.ClientMsgIndices())
	state.Workers.Federator.Init(messages.FederatorMsgIndices())
	state.Workers.Delivery.Init(nil)

	// Specifically do NOT start the workers
	// as caller may require queue contents.
	// (i.e. don't want workers pulling)
	// _ = state.Workers.Client.Start(1)
	// _ = state.Workers.Federator.Start(1)
	// _ = state.Workers.Dereference.Start(1)
	// _ = state.Workers.Media.Start(1)
	//
	// (except for the scheduler, that's fine)
	_ = state.Workers.Scheduler.Start()
}

// Starts workers on the provided state using processing functions from the given
// workers processor. Useful when you *do* want to trigger side effects in a test.
func StartWorkers(state *state.State, processor *workers.Processor) {
	state.Workers.Client.Process = func(ctx context.Context, msg *messages.FromClientAPI) error {
		log.Debugf(ctx, "Workers{}.Client{}.Process(%s)", dump(msg))
		return processor.ProcessFromClientAPI(ctx, msg)
	}

	state.Workers.Federator.Process = func(ctx context.Context, msg *messages.FromFediAPI) error {
		log.Debugf(ctx, "Workers{}.Federator{}.Process(%s)", dump(msg))
		return processor.ProcessFromFediAPI(ctx, msg)
	}

	state.Workers.Client.Init(messages.ClientMsgIndices())
	state.Workers.Federator.Init(messages.FederatorMsgIndices())
	state.Workers.Delivery.Init(nil)

	_ = state.Workers.Scheduler.Start()
	state.Workers.Client.Start(1)
	state.Workers.Federator.Start(1)
	state.Workers.Dereference.Start(1)
	state.Workers.Processing.Start(1)
	state.Workers.WebPush.Start(1)
}

func StopWorkers(state *state.State) {
	_ = state.Workers.Scheduler.Stop()
	state.Workers.Client.Stop()
	state.Workers.Federator.Stop()
	state.Workers.Dereference.Stop()
	state.Workers.Processing.Stop()
	state.Workers.WebPush.Stop()
}

func StartTimelines(state *state.State, visFilter *visibility.Filter, converter *typeutils.Converter) {
	state.Timelines.Home = timeline.NewManager(
		tlprocessor.HomeTimelineGrab(state),
		tlprocessor.HomeTimelineFilter(state, visFilter),
		tlprocessor.HomeTimelineStatusPrepare(state, converter),
		tlprocessor.SkipInsert(),
	)
	if err := state.Timelines.Home.Start(); err != nil {
		panic(fmt.Sprintf("error starting home timeline: %s", err))
	}

	state.Timelines.List = timeline.NewManager(
		tlprocessor.ListTimelineGrab(state),
		tlprocessor.ListTimelineFilter(state, visFilter),
		tlprocessor.ListTimelineStatusPrepare(state, converter),
		tlprocessor.SkipInsert(),
	)
	if err := state.Timelines.List.Start(); err != nil {
		panic(fmt.Sprintf("error starting list timeline: %s", err))
	}
}

// EqualRequestURIs checks whether inputs have equal request URIs,
// handling cases of url.URL{}, *url.URL{}, string, *string.
func EqualRequestURIs(u1, u2 any) bool {
	var uri1, uri2 string

	requestURI := func(in string) (string, error) {
		u, err := url.Parse(in)
		if err != nil {
			return "", err
		}
		return u.RequestURI(), nil
	}

	switch u1 := u1.(type) {
	case url.URL:
		uri1 = u1.RequestURI()
	case *url.URL:
		uri1 = u1.RequestURI()
	case *string:
		var err error
		uri1, err = requestURI(*u1)
		if err != nil {
			return false
		}
	case string:
		var err error
		uri1, err = requestURI(u1)
		if err != nil {
			return false
		}
	default:
		panic("unsupported type")
	}

	switch u2 := u2.(type) {
	case url.URL:
		uri2 = u2.RequestURI()
	case *url.URL:
		uri2 = u2.RequestURI()
	case *string:
		var err error
		uri2, err = requestURI(*u2)
		if err != nil {
			return false
		}
	case string:
		var err error
		uri2, err = requestURI(u2)
		if err != nil {
			return false
		}
	default:
		panic("unsupported type")
	}

	return uri1 == uri2
}

type DataF func() (
	fieldName string,
	fileName string,
	rc io.ReadCloser,
	err error,
)

// CreateMultipartFormData is a handy function for creating a multipart form bytes buffer with data.
//
// If data function is not nil, it should return the fieldName for the data in the form (eg., "data"),
// the fileName (eg., "data.csv"), a readcloser for getting the data, or an error if something goes wrong.
//
// The extraFields param can be used to add extra FormFields to the request, as necessary.
//
// Data function can be nil if only FormFields and string values are required.
//
// The returned bytes.Buffer b can be used like so:
//
//	httptest.NewRequest(http.MethodPost, "https://example.org/whateverpath", bytes.NewReader(b.Bytes()))
//
// The returned *multipart.Writer w can be used to set the content type of the request, like so:
//
//	req.Header.Set("Content-Type", w.FormDataContentType())
func CreateMultipartFormData(
	dataF DataF,
	extraFields map[string][]string,
) (bytes.Buffer, *multipart.Writer, error) {
	var (
		b bytes.Buffer
		w = multipart.NewWriter(&b)
	)

	if dataF != nil {
		fieldName, fileName, rc, err := dataF()
		if err != nil {
			return b, nil, err
		}
		defer rc.Close()

		fw, err := w.CreateFormFile(fieldName, fileName)
		if err != nil {
			return b, nil, err
		}

		if _, err = io.Copy(fw, rc); err != nil {
			return b, nil, err
		}
	}

	for k, vs := range extraFields {
		for _, v := range vs {
			if err := w.WriteField(k, v); err != nil {
				return b, nil, err
			}
		}
	}

	if err := w.Close(); err != nil {
		return b, nil, err
	}

	return b, w, nil
}

// FileToDataF is a convenience function for opening a
// file at the given filePath, and packaging it into a
// DataF for use in CreateMultipartFormData.
func FileToDataF(fieldName string, filePath string) DataF {
	return func() (string, string, io.ReadCloser, error) {
		file, err := os.Open(filePath)
		if err != nil {
			return "", "", nil, err
		}

		return fieldName, path.Base(filePath), file, nil
	}
}

// StringToDataF is a convenience function for wrapping the
// given data into a DataF for use in CreateMultipartFormData.
func StringToDataF(fieldName string, fileName string, data string) DataF {
	return func() (string, string, io.ReadCloser, error) {
		rc := io.NopCloser(bytes.NewBufferString(data))
		return fieldName, fileName, rc, nil
	}
}

// URLMustParse tries to parse the given URL and panics if it can't.
// Should only be used in tests.
func URLMustParse(stringURL string) *url.URL {
	u, err := url.Parse(stringURL)
	if err != nil {
		panic(err)
	}
	return u
}

// TimeMustParse tries to parse the given time as RFC3339, and panics if it can't.
// Should only be used in tests.
func TimeMustParse(timeString string) time.Time {
	t, err := time.Parse(time.RFC3339, timeString)
	if err != nil {
		panic(err)
	}
	return t
}

// WaitFor calls condition every 200ms, returning true
// when condition() returns true, or false after 5s.
//
// It's useful for when you're waiting for something to
// happen, but you don't know exactly how long it will take,
// and you want to fail if the thing doesn't happen within 5s.
func WaitFor(condition func() bool) bool {
	tick := time.NewTicker(200 * time.Millisecond)
	defer tick.Stop()

	timeout := time.NewTimer(5 * time.Second)
	defer timeout.Stop()

	for {
		select {
		case <-tick.C:
			if condition() {
				return true
			}
		case <-timeout.C:
			return false
		}
	}
}

// dump returns debug output of 'v'.
func dump(v any) string {
	var buf byteutil.Buffer
	format.Append(&buf, v)
	return buf.String()
}

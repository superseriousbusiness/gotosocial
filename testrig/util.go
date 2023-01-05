/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package testrig

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/url"
	"os"
	"time"
)

// CreateMultipartFormData is a handy function for taking a fieldname and a filename, and creating a multipart form bytes buffer
// with the file contents set in the given fieldname. The extraFields param can be used to add extra FormFields to the request, as necessary.
// The returned bytes.Buffer b can be used like so:
//
//	httptest.NewRequest(http.MethodPost, "https://example.org/whateverpath", bytes.NewReader(b.Bytes()))
//
// The returned *multipart.Writer w can be used to set the content type of the request, like so:
//
//	req.Header.Set("Content-Type", w.FormDataContentType())
func CreateMultipartFormData(fieldName string, fileName string, extraFields map[string]string) (bytes.Buffer, *multipart.Writer, error) {
	var b bytes.Buffer

	w := multipart.NewWriter(&b)
	var fw io.Writer

	if fileName != "" {
		file, err := os.Open(fileName)
		if err != nil {
			return b, nil, err
		}
		if fw, err = w.CreateFormFile(fieldName, file.Name()); err != nil {
			return b, nil, err
		}
		if _, err = io.Copy(fw, file); err != nil {
			return b, nil, err
		}
	}

	for k, v := range extraFields {
		f, err := w.CreateFormField(k)
		if err != nil {
			return b, nil, err
		}
		if _, err := io.Copy(f, bytes.NewBufferString(v)); err != nil {
			return b, nil, err
		}
	}

	if err := w.Close(); err != nil {
		return b, nil, err
	}
	return b, w, nil
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

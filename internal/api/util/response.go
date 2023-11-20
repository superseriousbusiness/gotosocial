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

package util

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
	"strconv"
	"sync"

	"codeberg.org/gruf/go-byteutil"
	"codeberg.org/gruf/go-fastcopy"
	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

var (
	// Pre-preared response body data.
	StatusOKJSON = mustJSON(map[string]string{
		"status": http.StatusText(http.StatusOK),
	})
	StatusAcceptedJSON = mustJSON(map[string]string{
		"status": http.StatusText(http.StatusAccepted),
	})
	StatusInternalServerErrorJSON = mustJSON(map[string]string{
		"status": http.StatusText(http.StatusInternalServerError),
	})

	// write buffer pool.
	bufPool = sync.Pool{}
)

// JSON calls EncodeJSONResponse() using gin.Context{}, with content-type = AppJSON,
// This function handles the case of JSON unmarshal errors and pools read buffers.
func JSON(c *gin.Context, code int, data any) {
	EncodeJSONResponse(c.Writer, c.Request, code, AppJSON, data)
}

// JSON calls EncodeJSONResponse() using gin.Context{}, with given content-type.
// This function handles the case of JSON unmarshal errors and pools read buffers.
func JSONType(c *gin.Context, code int, contentType string, data any) {
	EncodeJSONResponse(c.Writer, c.Request, code, contentType, data)
}

// Data calls WriteResponseBytes() using gin.Context{}, with given content-type.
func Data(c *gin.Context, code int, contentType string, data []byte) {
	WriteResponseBytes(c.Writer, c.Request, code, contentType, data)
}

// WriteResponse buffered streams 'data' as HTTP response
// to ResponseWriter with given status code content-type.
func WriteResponse(
	rw http.ResponseWriter,
	r *http.Request,
	statusCode int,
	contentType string,
	data io.Reader,
	length int64,
) {
	if length < 0 {
		// The worst-case scenario, length is not known so we need to
		// read the entire thing into memory to know length & respond.
		writeResponseUnknownLength(rw, r, statusCode, contentType, data)
		return
	}

	// The best-case scenario, stream content of known length.
	rw.Header().Set("Content-Type", contentType)
	rw.Header().Set("Content-Length", strconv.FormatInt(length, 10))
	rw.WriteHeader(statusCode)
	if _, err := fastcopy.Copy(rw, data); err != nil {
		log.Errorf(r.Context(), "%v", err)
	}
}

// WriteResponseBytes is functionally similar to
// WriteResponse except that it takes prepared bytes.
func WriteResponseBytes(
	rw http.ResponseWriter,
	r *http.Request,
	statusCode int,
	contentType string,
	data []byte,
) {
	rw.Header().Set("Content-Type", contentType)
	rw.Header().Set("Content-Length", strconv.Itoa(len(data)))
	rw.WriteHeader(statusCode)
	if _, err := rw.Write(data); err != nil {
		log.Errorf(r.Context(), "%v", err)
	}
}

// EncodeJSONResponse encodes 'data' as JSON HTTP response
// to ResponseWriter with given status code, content-type.
func EncodeJSONResponse(
	rw http.ResponseWriter,
	r *http.Request,
	statusCode int,
	contentType string,
	data any,
) {
	// Acquire buffer from pool.
	buf := bufPool.Get().(*byteutil.Buffer)

	if buf == nil {
		// Alloc new buf if needed.
		buf = new(byteutil.Buffer)
		buf.B = make([]byte, 0, 4096)
	}

	// Wrap buffer in JSON encoder.
	enc := json.NewEncoder(buf)

	// Encode JSON data into byte buffer.
	if err := enc.Encode(data); err == nil {

		// Respond with the now-known
		// size byte slice within buf.
		WriteResponseBytes(rw, r,
			statusCode,
			contentType,
			buf.B,
		)
	} else {

		// any error returned here is unrecoverable,
		// set Internal Server Error JSON response.
		WriteResponseBytes(rw, r,
			http.StatusInternalServerError,
			AppJSON,
			StatusInternalServerErrorJSON,
		)
	}

	if cap(buf.B) >= int(^uint16(0)) {
		// drop buffers of large size.
		return
	}

	// Release to pool.
	bufPool.Put(buf)
}

// EncodeJSONResponse encodes 'data' as XML HTTP response
// to ResponseWriter with given status code, content-type.
func EncodeXMLResponse(
	rw http.ResponseWriter,
	r *http.Request,
	statusCode int,
	contentType string,
	data any,
) {
	// Acquire buffer from pool.
	buf := bufPool.Get().(*byteutil.Buffer)

	if buf == nil {
		// Alloc new buf if needed.
		buf = new(byteutil.Buffer)
		buf.B = make([]byte, 0, 4096)
	}

	// Write XML header to buf.
	buf.WriteString(xml.Header)

	// Wrap buffer in XML encoder.
	enc := xml.NewEncoder(buf)

	// Encode JSON data into byte buffer.
	if err := enc.Encode(data); err == nil {

		// Respond with the now-known
		// size byte slice within buf.
		WriteResponseBytes(rw, r,
			statusCode,
			contentType,
			buf.B,
		)
	} else {

		// any error returned here is unrecoverable,
		// set Internal Server Error JSON response.
		WriteResponseBytes(rw, r,
			http.StatusInternalServerError,
			AppJSON,
			StatusInternalServerErrorJSON,
		)
	}

	if cap(buf.B) >= int(^uint16(0)) {
		// drop buffers of large size.
		return
	}

	// Release to pool.
	bufPool.Put(buf)
}

// writeResponseUnknownLength handles reading data of unknown legnth
// efficiently into memory, and passing on to WriteResponseBytes().
func writeResponseUnknownLength(
	rw http.ResponseWriter,
	r *http.Request,
	statusCode int,
	contentType string,
	data io.Reader,
) {
	// Acquire buffer from pool.
	buf := bufPool.Get().(*byteutil.Buffer)

	if buf == nil {
		// Alloc new buf if needed.
		buf = new(byteutil.Buffer)
		buf.B = make([]byte, 0, 4096)
	}

	// Read content into buffer.
	err := writeTo(buf, data)

	if err == nil {

		// Respond with the now-known
		// size byte slice within buf.
		WriteResponseBytes(rw, r,
			statusCode,
			contentType,
			buf.B,
		)
	} else {

		// any error returned here is unrecoverable,
		// set Internal Server Error JSON response.
		WriteResponseBytes(rw, r,
			http.StatusInternalServerError,
			AppJSON,
			StatusInternalServerErrorJSON,
		)
	}

	if cap(buf.B) >= int(^uint16(0)) {
		// drop buffers of large size.
		return
	}

	// Release to pool.
	bufPool.Put(buf)
}

// writeTo reads data from io.Reader into given byte buffer.
func writeTo(w *byteutil.Buffer, r io.Reader) error {
	if w == nil { // nil check outside loop.
		panic("nil buffer")
	}
	for {
		n, err := r.Read(w.B[len(w.B):cap(w.B)])
		w.B = w.B[:len(w.B)+n]
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return err
		}
		if len(w.B) == cap(w.B) {
			// Increase cap (let append pick).
			w.B = append(w.B, 0)[:len(w.B)]
		}
	}
}

// mustJSON converts data to JSON, else panicking.
func mustJSON(data any) []byte {
	b, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	return b
}

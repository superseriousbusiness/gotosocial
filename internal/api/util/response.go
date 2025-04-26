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
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
	"strconv"
	"sync"

	"code.superseriousbusiness.org/gotosocial/internal/log"
	"codeberg.org/gruf/go-byteutil"
	"codeberg.org/gruf/go-fastcopy"
	"github.com/gin-gonic/gin"
)

var (
	// Pre-preared response body data.
	StatusOKJSON = mustJSON(map[string]string{
		"status": http.StatusText(http.StatusOK),
	})
	StatusAcceptedJSON = mustJSON(map[string]string{
		"status": http.StatusText(http.StatusAccepted),
	})
	StatusForbiddenJSON = mustJSON(map[string]string{
		"status": http.StatusText(http.StatusForbidden),
	})
	StatusInternalServerErrorJSON = mustJSON(map[string]string{
		"status": http.StatusText(http.StatusInternalServerError),
	})
	ErrorCapacityExceeded = mustJSON(map[string]string{
		"error": "server capacity exceeded",
	})
	ErrorRateLimited = mustJSON(map[string]string{
		"error": "rate limit reached",
	})
	EmptyJSONObject = json.RawMessage(`{}`)
	EmptyJSONArray  = json.RawMessage(`[]`)

	// write buffer pool.
	bufPool sync.Pool
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
		log.Errorf(r.Context(), "error streaming: %v", err)
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
	if _, err := rw.Write(data); err != nil && err != io.EOF {
		log.Errorf(r.Context(), "error writing: %v", err)
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
	// Acquire buffer.
	buf := getBuf()

	// Wrap buffer in JSON encoder.
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)

	// Encode JSON data into byte buffer.
	if err := enc.Encode(data); err == nil {

		// Drop new-line added by encoder.
		if buf.B[len(buf.B)-1] == '\n' {
			buf.B = buf.B[:len(buf.B)-1]
		}

		// Respond with the now-known
		// size byte slice within buf.
		WriteResponseBytes(rw, r,
			statusCode,
			contentType,
			buf.B,
		)
	} else {
		// This will always be a JSON error, we
		// can't really add any more useful context.
		log.Error(r.Context(), err)

		// Any error returned here is unrecoverable,
		// set Internal Server Error JSON response.
		WriteResponseBytes(rw, r,
			http.StatusInternalServerError,
			AppJSON,
			StatusInternalServerErrorJSON,
		)
	}

	// Release.
	putBuf(buf)
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
	// Acquire buffer.
	buf := getBuf()

	// Write XML header string to buf.
	buf.B = append(buf.B, xml.Header...)

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
		// This will always be an XML error, we
		// can't really add any more useful context.
		log.Error(r.Context(), err)

		// Any error returned here is unrecoverable,
		// set Internal Server Error JSON response.
		WriteResponseBytes(rw, r,
			http.StatusInternalServerError,
			AppJSON,
			StatusInternalServerErrorJSON,
		)
	}

	// Release.
	putBuf(buf)
}

// EncodeCSVResponse encodes 'records' as CSV HTTP response
// to ResponseWriter with given status code, using CSV content-type.
func EncodeCSVResponse(
	rw http.ResponseWriter,
	r *http.Request,
	statusCode int,
	records [][]string,
) {
	// Acquire buffer.
	buf := getBuf()

	// Wrap buffer in CSV writer.
	csvWriter := csv.NewWriter(buf)

	// Write all the records to the buffer.
	if err := csvWriter.WriteAll(records); err == nil {
		// Respond with the now-known
		// size byte slice within buf.
		WriteResponseBytes(rw, r,
			statusCode,
			TextCSV,
			buf.B,
		)
	} else {
		// This will always be an csv error, we
		// can't really add any more useful context.
		log.Error(r.Context(), err)

		// Any error returned here is unrecoverable,
		// set Internal Server Error JSON response.
		WriteResponseBytes(rw, r,
			http.StatusInternalServerError,
			AppJSON,
			StatusInternalServerErrorJSON,
		)
	}

	// Release.
	putBuf(buf)
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
	// Acquire buffer.
	buf := getBuf()

	// Read content into buffer.
	_, err := buf.ReadFrom(data)

	if err == nil {

		// Respond with the now-known
		// size byte slice within buf.
		WriteResponseBytes(rw, r,
			statusCode,
			contentType,
			buf.B,
		)
	} else {
		// This will always be a reader error (non EOF),
		// but that doesn't mean the writer is closed yet!
		log.Errorf(r.Context(), "error reading: %v", err)

		// Any error returned here is unrecoverable,
		// set Internal Server Error JSON response.
		WriteResponseBytes(rw, r,
			http.StatusInternalServerError,
			AppJSON,
			StatusInternalServerErrorJSON,
		)
	}

	// Release.
	putBuf(buf)
}

func getBuf() *byteutil.Buffer {
	// acquire buffer from pool.
	buf, _ := bufPool.Get().(*byteutil.Buffer)

	if buf == nil {
		// alloc new buf if needed.
		buf = new(byteutil.Buffer)
		buf.B = make([]byte, 0, 4096)
	}

	return buf
}

func putBuf(buf *byteutil.Buffer) {
	if cap(buf.B) >= int(^uint16(0)) {
		// drop buffers of large size.
		return
	}

	// ensure empty.
	buf.Reset()

	// release to pool.
	bufPool.Put(buf)
}

// mustJSON converts data to JSON, else panicking.
func mustJSON(data any) []byte {
	b, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	return b
}

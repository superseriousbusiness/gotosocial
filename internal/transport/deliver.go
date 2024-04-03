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

package transport

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"codeberg.org/gruf/go-byteutil"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/queue"
)

func (t *transport) BatchDeliver(ctx context.Context, obj map[string]interface{}, recipients []*url.URL) error {
	var (
		// accumulated prepared reqs.
		reqs []*queue.HTTPRequest

		// accumulated preparation errs.
		errs gtserror.MultiError

		// Get current instance host info.
		domain = config.GetAccountDomain()
		host   = config.GetHost()
	)

	// Marshal object as JSON.
	b, err := json.Marshal(obj)
	if err != nil {
		return gtserror.Newf("error marshaling json: %w", err)
	}

	// Extract object ID.
	id := getObjectID(obj)

	for _, to := range recipients {
		// Skip delivery to recipient if it is "us".
		if to.Host == host || to.Host == domain {
			continue
		}

		// Prepare new http client request.
		req, err := t.prepare(ctx, id, b, to)
		if err != nil {
			errs.Append(err)
			continue
		}

		// Append to request queue.
		reqs = append(reqs, req)
	}

	// Push the request list to HTTP client worker queue.
	t.controller.state.Queues.HTTPRequest.Push(reqs...)

	// Return combined err.
	return errs.Combine()
}

func (t *transport) Deliver(ctx context.Context, obj map[string]interface{}, to *url.URL) error {
	// if 'to' host is our own, skip as we don't need to deliver to ourselves...
	if to.Host == config.GetHost() || to.Host == config.GetAccountDomain() {
		return nil
	}

	// Marshal object as JSON.
	b, err := json.Marshal(obj)
	if err != nil {
		return gtserror.Newf("error marshaling json: %w", err)
	}

	// Extract object ID.
	id := getObjectID(obj)

	// Prepare new http client request.
	req, err := t.prepare(ctx, id, b, to)
	if err != nil {
		return err
	}

	// Push the request to HTTP client worker queue.
	t.controller.state.Queues.HTTPRequest.Push(req)

	return nil
}

// prepare will prepare a POST http.Request{}
// to recipient at 'to', wrapping in a queued
// request object with signing function.
func (t *transport) prepare(
	ctx context.Context,
	objectID string,
	data []byte,
	to *url.URL,
) (
	*queue.HTTPRequest,
	error,
) {
	url := to.String()

	// Use rewindable reader for body.
	var body byteutil.ReadNopCloser
	body.Reset(data)

	// Prepare POST signer.
	sign := t.signPOST(data)

	// Update to-be-used request context with signing details.
	ctx = gtscontext.SetOutgoingPublicKeyID(ctx, t.pubKeyID)
	ctx = gtscontext.SetHTTPClientSignFunc(ctx, sign)

	req, err := http.NewRequestWithContext(ctx, "POST", url, &body)
	if err != nil {
		return nil, gtserror.Newf("error preparing request: %w", err)
	}

	req.Header.Add("Content-Type", string(apiutil.AppActivityLDJSON))
	req.Header.Add("Accept-Charset", "utf-8")

	return &queue.HTTPRequest{
		ObjectID: objectID,
		Request:  req,
	}, nil
}

// getObjectID extracts an object ID from 'serialized' ActivityPub object map.
func getObjectID(obj map[string]interface{}) string {
	switch t := obj["object"].(type) {
	case string:
		return t
	case map[string]interface{}:
		id, _ := t["id"].(string)
		return id
	default:
		return ""
	}
}

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
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/gtscontext"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/httpclient"
	"code.superseriousbusiness.org/gotosocial/internal/transport/delivery"
)

func (t *transport) BatchDeliver(ctx context.Context, obj map[string]interface{}, recipients []*url.URL) error {
	var (
		// accumulated delivery reqs.
		reqs []*delivery.Delivery

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

	// Extract object IDs.
	actID := getActorID(obj)
	objID := getObjectID(obj)
	tgtID := getTargetID(obj)

	for _, to := range recipients {
		// Skip delivery to recipient if it is "us".
		if to.Host == host || to.Host == domain {
			continue
		}

		// Prepare http client request.
		req, err := t.prepare(ctx,
			actID,
			objID,
			tgtID,
			b,
			to,
		)
		if err != nil {
			errs.Append(err)
			continue
		}

		// Append to request queue.
		reqs = append(reqs, req)
	}

	// Push prepared request list to the delivery queue.
	t.controller.state.Workers.Delivery.Queue.Push(reqs...)

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

	// Prepare http client request.
	req, err := t.prepare(ctx,
		getActorID(obj),
		getObjectID(obj),
		getTargetID(obj),
		b,
		to,
	)
	if err != nil {
		return err
	}

	// Push prepared request to the delivery queue.
	t.controller.state.Workers.Delivery.Queue.Push(req)

	return nil
}

// prepare will prepare a POST http.Request{}
// to recipient at 'to', wrapping in a queued
// request object with signing function.
func (t *transport) prepare(
	ctx context.Context,
	actorID string,
	objectID string,
	targetID string,
	data []byte,
	to *url.URL,
) (
	*delivery.Delivery,
	error,
) {
	// Prepare POST signer.
	sign := t.signPOST(data)

	// Use *bytes.Reader for request body,
	// as NewRequest() automatically will
	// set .GetBody and content-length.
	// (this handles necessary rewinding).
	body := bytes.NewReader(data)

	// Update to-be-used request context with signing details.
	ctx = gtscontext.SetOutgoingPublicKeyID(ctx, t.pubKeyID)
	ctx = gtscontext.SetHTTPClientSignFunc(ctx, sign)

	// Prepare a new request with data body directed at URL.
	r, err := http.NewRequestWithContext(ctx, "POST", to.String(), body)
	if err != nil {
		return nil, gtserror.Newf("error preparing request: %w", err)
	}

	// Set our predefined controller user-agent.
	r.Header.Set("User-Agent", t.controller.userAgent)

	// Set the standard ActivityPub content-type + charset headers.
	r.Header.Add("Content-Type", string(apiutil.AppActivityLDJSON))
	r.Header.Add("Accept-Charset", "utf-8")

	// Validate the request before queueing for delivery.
	if err := httpclient.ValidateRequest(r); err != nil {
		return nil, err
	}

	return &delivery.Delivery{
		ActorID:  actorID,
		ObjectID: objectID,
		TargetID: targetID,
		Request:  httpclient.WrapRequest(r),
	}, nil
}

func (t *transport) SignDelivery(dlv *delivery.Delivery) error {
	if dlv.Request.GetBody == nil {
		return gtserror.New("delivery request body not rewindable")
	}

	// Fetch a fresh copy of request body.
	rBody, err := dlv.Request.GetBody()
	if err != nil {
		return gtserror.Newf("error getting request body: %w", err)
	}

	// Read body data into memory.
	data, err := io.ReadAll(rBody)

	// Done with body.
	_ = rBody.Close()

	if err != nil {
		return gtserror.Newf("error reading request body: %w", err)
	}

	// Get signing function for POST data.
	// (note that delivery is ALWAYS POST).
	sign := t.signPOST(data)

	// Extract delivery context.
	ctx := dlv.Request.Context()

	// Update delivery request context with signing details.
	ctx = gtscontext.SetOutgoingPublicKeyID(ctx, t.pubKeyID)
	ctx = gtscontext.SetHTTPClientSignFunc(ctx, sign)
	dlv.Request.Request = dlv.Request.Request.WithContext(ctx)

	return nil
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

// getActorID extracts an actor ID from 'serialized' ActivityPub object map.
func getActorID(obj map[string]interface{}) string {
	switch t := obj["actor"].(type) {
	case string:
		return t
	case map[string]interface{}:
		id, _ := t["id"].(string)
		return id
	default:
		return ""
	}
}

// getTargetID extracts a target ID from 'serialized' ActivityPub object map.
func getTargetID(obj map[string]interface{}) string {
	switch t := obj["target"].(type) {
	case string:
		return t
	case map[string]interface{}:
		id, _ := t["id"].(string)
		return id
	default:
		return ""
	}
}

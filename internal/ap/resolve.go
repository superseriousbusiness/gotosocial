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

package ap

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"code.superseriousbusiness.org/activity/pub"
	"code.superseriousbusiness.org/activity/streams"
	"code.superseriousbusiness.org/activity/streams/vocab"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
)

// ResolveActivity is a util function for pulling a pub.Activity type out of an incoming request body,
// returning the resolved activity type, error and whether to accept activity (false = transient i.e. ignore).
func ResolveIncomingActivity(r *http.Request) (pub.Activity, bool, gtserror.WithCode) {
	// Get "raw" map
	// destination.
	raw := getMap()
	// Release.
	defer putMap(raw)

	// Decode data as JSON into 'raw' map
	// and get the resolved AS vocab.Type.
	// (this handles close of request body).
	t, err := decodeType(r.Context(), r.Body, raw)

	if err != nil {
		// NOTE: if the error here was due to the response body
		// ending early, the connection will have broken so it
		// doesn't matter if we try to return 400 or 500, the
		// error is mainly for our logging. tl;dr there's not a
		// huge need to differentiate between those error types.

		if !streams.IsUnmatchedErr(err) {
			err := gtserror.Newf("error matching json to type: %w", err)
			return nil, false, gtserror.NewErrorInternalError(err)
		}

		const text = "body json not resolvable as ActivityStreams type"
		return nil, false, gtserror.NewErrorBadRequest(errors.New(text), text)
	}

	// Ensure this is an Activity type.
	activity, ok := t.(pub.Activity)
	if !ok {
		text := fmt.Sprintf("cannot resolve vocab type %T as pub.Activity", t)
		return nil, false, gtserror.NewErrorBadRequest(errors.New(text), text)
	}

	if activity.GetJSONLDId() == nil {
		// missing ID indicates a transient ID as per:
		//
		// all objects distributed by the ActivityPub protocol MUST have unique global identifiers,
		// unless they are intentionally transient (short lived activities that are not intended to
		// be able to be looked up, such as some kinds of chat messages or game notifications).
		return nil, false, nil
	}

	// Normalize any Statusable, Accountable, Pollable fields found.
	// (see: https://codeberg.org/superseriousbusiness/gotosocial/issues/1661)
	NormalizeIncomingActivity(activity, raw)

	return activity, true, nil
}

// ResolveStatusable tries to resolve the response data as an ActivityPub
// Statusable representation. It will then perform normalization on the Statusable.
//
// Works for: Article, Document, Image, Video, Note, Page, Event, Place, Profile, Question.
func ResolveStatusable(ctx context.Context, body io.ReadCloser) (Statusable, error) {
	// Get "raw" map
	// destination.
	raw := getMap()
	// Release.
	defer putMap(raw)

	// Decode data as JSON into 'raw' map
	// and get the resolved AS vocab.Type.
	// (this handles close of given body).
	t, err := decodeType(ctx, body, raw)
	if err != nil {
		return nil, gtserror.SetWrongType(err)
	}

	// Attempt to cast as Statusable.
	statusable, ok := ToStatusable(t)
	if !ok {
		err := gtserror.Newf("cannot resolve vocab type %T as statusable", t)
		return nil, gtserror.SetWrongType(err)
	}

	if pollable, ok := ToPollable(statusable); ok {
		// Question requires extra normalization, and
		// fortunately directly implements Statusable.
		NormalizeIncomingPollOptions(pollable, raw)
		statusable = pollable
	}

	NormalizeIncomingContent(statusable, raw)
	NormalizeIncomingAttachments(statusable, raw)
	NormalizeIncomingSummary(statusable, raw)
	NormalizeIncomingName(statusable, raw)

	return statusable, nil
}

// ResolveAccountable tries to resolve the given reader into an ActivityPub
// Accountable representation. It will then perform normalization on the Accountable.
//
// Works for: Application, Group, Organization, Person, Service
func ResolveAccountable(ctx context.Context, body io.ReadCloser) (Accountable, error) {
	// Get "raw" map
	// destination.
	raw := getMap()
	// Release.
	defer putMap(raw)

	// Decode data as JSON into 'raw' map
	// and get the resolved AS vocab.Type.
	// (this handles close of given body).
	t, err := decodeType(ctx, body, raw)
	if err != nil {
		return nil, gtserror.SetWrongType(err)
	}

	// Attempt to cast as Statusable.
	accountable, ok := ToAccountable(t)
	if !ok {
		err := gtserror.Newf("cannot resolve vocab type %T as accountable", t)
		return nil, gtserror.SetWrongType(err)
	}

	NormalizeIncomingSummary(accountable, raw)

	return accountable, nil
}

// ResolveCollection tries to resolve the given reader into an ActivityPub Collection-like
// representation, then wrapping as abstracted iterator. Works for: Collection, OrderedCollection.
func ResolveCollection(ctx context.Context, body io.ReadCloser) (CollectionIterator, error) {
	// Get "raw" map
	// destination.
	raw := getMap()
	// Release.
	defer putMap(raw)

	// Decode data as JSON into 'raw' map
	// and get the resolved AS vocab.Type.
	// (this handles close of given body).
	t, err := decodeType(ctx, body, raw)
	if err != nil {
		return nil, gtserror.SetWrongType(err)
	}

	// Cast as as Collection-like.
	return ToCollectionIterator(t)
}

// ResolveCollectionPage tries to resolve the given reader into an ActivityPub CollectionPage-like
// representation, then wrapping as abstracted iterator. Works for: CollectionPage, OrderedCollectionPage.
func ResolveCollectionPage(ctx context.Context, body io.ReadCloser) (CollectionPageIterator, error) {
	// Get "raw" map
	// destination.
	raw := getMap()
	// Release.
	defer putMap(raw)

	// Decode data as JSON into 'raw' map
	// and get the resolved AS vocab.Type.
	// (this handles close of given body).
	t, err := decodeType(ctx, body, raw)
	if err != nil {
		return nil, gtserror.SetWrongType(err)
	}

	// Cast as as CollectionPage-like.
	return ToCollectionPageIterator(t)
}

// emptydest is an empty JSON decode
// destination useful for "noop" decodes
// to check underlying reader is empty.
var emptydest = &struct{}{}

// decodeType is the package-internal version of DecodeType.
//
// The given map pointer will also be populated with
// the 'raw' JSON data, for further processing.
func decodeType(
	ctx context.Context,
	body io.ReadCloser,
	raw map[string]any,
) (vocab.Type, error) {

	// Wrap body in JSON decoder.
	//
	// We do this instead of using json.Unmarshal()
	// so we can take advantage of the decoder's streamed
	// check of input data as valid JSON. This means that
	// in the cases of garbage input, or even just fallback
	// HTML responses that were incorrectly content-type'd,
	// we can error-out as soon as possible.
	dec := json.NewDecoder(body)

	// Unmarshal JSON source data into "raw" map.
	if err := dec.Decode(&raw); err != nil {
		_ = body.Close() // ensure closed.
		return nil, gtserror.NewfAt(3, "error decoding into json: %w", err)
	}

	// Perform a secondary decode just to ensure we drained the
	// entirety of the data source. Error indicates either extra
	// trailing garbage, or multiple JSON values (invalid data).
	if err := dec.Decode(emptydest); err != io.EOF {
		_ = body.Close() // ensure closed.
		return nil, gtserror.NewfAt(3, "data remaining after json")
	}

	// Done with body.
	_ = body.Close()

	// Resolve an ActivityStreams type.
	t, err := streams.ToType(ctx, raw)
	if err != nil {
		return nil, gtserror.NewfAt(3, "error resolving json into ap vocab type: %w", err)
	}

	return t, nil
}

// DecodeType tries to read and parse the data
// at provided io.ReadCloser as a JSON ActivityPub
// type, failing if not parseable as JSON or not
// resolveable as one of our known AS types.
//
// NOTE: this function handles closing
// given body when it is finished with.
func DecodeType(
	ctx context.Context,
	body io.ReadCloser,
) (vocab.Type, error) {
	// Get "raw" map
	// destination.
	raw := getMap()
	// Release.
	defer putMap(raw)

	return decodeType(ctx, body, raw)
}

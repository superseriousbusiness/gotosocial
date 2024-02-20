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

	"github.com/superseriousbusiness/activity/pub"
	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
)

// decodeBody tries to read and parse the data
// at provided io.Reader as a JSON ActivityPub
// type, failing if not parseable as JSON or not
// resolveable as one of our known AS types.
//
// The given map pointer will also be populated with
// the 'raw' JSON data, for further processing.
func decodeBody(
	ctx context.Context,
	body io.Reader,
	raw map[string]any,
) (vocab.Type, error) {
	// Unmarshal the raw JSON bytes into a "raw" map.
	// This will fail if the input is not parseable
	// as JSON; eg., a remote has returned HTML as a
	// fallback response to an ActivityPub JSON request.
	if err := json.NewDecoder(body).Decode(&raw); err != nil {
		return nil, gtserror.NewfAt(3, "error decoding into json: %w", err)
	}

	// Resolve an ActivityStreams type.
	t, err := streams.ToType(ctx, raw)
	if err != nil {
		return nil, gtserror.NewfAt(3, "error resolving json into ap vocab type: %w", err)
	}

	return t, nil
}

// ResolveActivity is a util function for pulling a pub.Activity type out of an incoming request body,
// returning the resolved activity type, error and whether to accept activity (false = transient i.e. ignore).
func ResolveIncomingActivity(r *http.Request) (pub.Activity, bool, gtserror.WithCode) {
	// Get "raw" map
	// destination.
	raw := getMap()

	// Tidy up when done.
	defer r.Body.Close()

	// Decode data as JSON into 'raw' map
	// and get the resolved AS vocab.Type.
	t, err := decodeBody(r.Context(), r.Body, raw)

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
	// (see: https://github.com/superseriousbusiness/gotosocial/issues/1661)
	NormalizeIncomingActivity(activity, raw)

	// Release.
	putMap(raw)

	return activity, true, nil
}

// ResolveStatusable tries to resolve the response data as an ActivityPub
// Statusable representation. It will then perform normalization on the Statusable.
//
// Works for: Article, Document, Image, Video, Note, Page, Event, Place, Profile, Question.
func ResolveStatusable(ctx context.Context, data io.Reader) (Statusable, error) {
	// Get "raw" map
	// destination.
	raw := getMap()

	// Decode data as JSON into 'raw' map
	// and get the resolved AS vocab.Type.
	t, err := decodeBody(ctx, data, raw)
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

	// Release.
	putMap(raw)

	return statusable, nil
}

// ResolveAccountable tries to resolve the given reader into an ActivityPub
// Accountable representation. It will then perform normalization on the Accountable.
//
// Works for: Application, Group, Organization, Person, Service
func ResolveAccountable(ctx context.Context, data io.Reader) (Accountable, error) {
	// Get "raw" map
	// destination.
	raw := getMap()

	// Decode data as JSON into 'raw' map
	// and get the resolved AS vocab.Type.
	t, err := decodeBody(ctx, data, raw)
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

	// Release.
	putMap(raw)

	return accountable, nil
}

// ResolveCollectionPage tries to resolve the given reader into an ActivityPub CollectionPage-like
// representation, then wrapping as abstracted iterator. Works for: CollectionPage, OrderedCollectionPage.
func ResolveCollectionPage(ctx context.Context, data io.Reader) (CollectionPageIterator, error) {
	// Get "raw" map
	// destination.
	raw := getMap()

	// Decode data as JSON into 'raw' map
	// and get the resolved AS vocab.Type.
	t, err := decodeBody(ctx, data, raw)
	if err != nil {
		return nil, gtserror.SetWrongType(err)
	}

	// Cast as as CollectionPage-like.
	return ToCollectionPageIterator(t)
}

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
	"fmt"

	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/activity/streams/vocab"
)

// ResolveStatusable tries to resolve the given bytes into an ActivityPub Statusable representation.
// It will then perform normalization on the Statusable by calling NormalizeStatusable, so that
// callers don't need to bother doing extra steps.
//
// Works for: Article, Document, Image, Video, Note, Page, Event, Place, Profile
func ResolveStatusable(ctx context.Context, b []byte) (Statusable, error) {
	rawStatusable := make(map[string]interface{})
	if err := json.Unmarshal(b, &rawStatusable); err != nil {
		return nil, fmt.Errorf("ResolveStatusable: error unmarshalling bytes into json: %s", err)
	}

	t, err := streams.ToType(ctx, rawStatusable)
	if err != nil {
		return nil, fmt.Errorf("ResolveStatusable: error resolving json into ap vocab type: %s", err)
	}

	var (
		statusable Statusable
		ok         bool
	)

	switch t.GetTypeName() {
	case ObjectArticle:
		statusable, ok = t.(vocab.ActivityStreamsArticle)
	case ObjectDocument:
		statusable, ok = t.(vocab.ActivityStreamsDocument)
	case ObjectImage:
		statusable, ok = t.(vocab.ActivityStreamsImage)
	case ObjectVideo:
		statusable, ok = t.(vocab.ActivityStreamsVideo)
	case ObjectNote:
		statusable, ok = t.(vocab.ActivityStreamsNote)
	case ObjectPage:
		statusable, ok = t.(vocab.ActivityStreamsPage)
	case ObjectEvent:
		statusable, ok = t.(vocab.ActivityStreamsEvent)
	case ObjectPlace:
		statusable, ok = t.(vocab.ActivityStreamsPlace)
	case ObjectProfile:
		statusable, ok = t.(vocab.ActivityStreamsProfile)
	}

	if !ok || statusable == nil {
		return nil, fmt.Errorf("ResolveStatusable: could not resolve %T to Statusable", t)
	}

	NormalizeStatusableContent(statusable, rawStatusable)
	return statusable, nil
}

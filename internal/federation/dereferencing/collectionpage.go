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

package dereferencing

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
)

// DereferenceCollectionPage returns the activitystreams CollectionPage at the specified IRI, or an error if something goes wrong.
func (d *deref) DereferenceCollectionPage(ctx context.Context, username string, pageIRI *url.URL) (ap.CollectionPageable, error) {
	if blocked, err := d.db.IsDomainBlocked(ctx, pageIRI.Host); blocked || err != nil {
		return nil, fmt.Errorf("DereferenceCollectionPage: domain %s is blocked", pageIRI.Host)
	}

	transport, err := d.transportController.NewTransportForUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("DereferenceCollectionPage: error creating transport: %s", err)
	}

	b, err := transport.Dereference(ctx, pageIRI)
	if err != nil {
		return nil, fmt.Errorf("DereferenceCollectionPage: error deferencing %s: %s", pageIRI.String(), err)
	}

	m := make(map[string]interface{})
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, fmt.Errorf("DereferenceCollectionPage: error unmarshalling bytes into json: %s", err)
	}

	t, err := streams.ToType(ctx, m)
	if err != nil {
		return nil, fmt.Errorf("DereferenceCollectionPage: error resolving json into ap vocab type: %s", err)
	}

	if t.GetTypeName() != ap.ObjectCollectionPage {
		return nil, fmt.Errorf("DereferenceCollectionPage: type name %s not supported", t.GetTypeName())
	}

	p, ok := t.(vocab.ActivityStreamsCollectionPage)
	if !ok {
		return nil, errors.New("DereferenceCollectionPage: error resolving type as activitystreams collection page")
	}

	return p, nil
}

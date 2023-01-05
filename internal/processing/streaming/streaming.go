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

package streaming

import (
	"context"
	"sync"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/stream"
)

// Processor wraps a bunch of functions for processing streaming.
type Processor interface {
	// AuthorizeStreamingRequest returns an oauth2 token info in response to an access token query from the streaming API
	AuthorizeStreamingRequest(ctx context.Context, accessToken string) (*gtsmodel.Account, gtserror.WithCode)
	// OpenStreamForAccount returns a new Stream for the given account, which will contain a channel for passing messages back to the caller.
	OpenStreamForAccount(ctx context.Context, account *gtsmodel.Account, timeline string) (*stream.Stream, gtserror.WithCode)
	// StreamUpdateToAccount streams the given update to any open, appropriate streams belonging to the given account.
	StreamUpdateToAccount(s *apimodel.Status, account *gtsmodel.Account, timeline string) error
	// StreamNotificationToAccount streams the given notification to any open, appropriate streams belonging to the given account.
	StreamNotificationToAccount(n *apimodel.Notification, account *gtsmodel.Account) error
	// StreamDelete streams the delete of the given statusID to *ALL* open streams.
	StreamDelete(statusID string) error
}

type processor struct {
	db          db.DB
	oauthServer oauth.Server
	streamMap   *sync.Map
}

// New returns a new status processor.
func New(db db.DB, oauthServer oauth.Server) Processor {
	return &processor{
		db:          db,
		oauthServer: oauthServer,
		streamMap:   &sync.Map{},
	}
}

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

package account

import (
	"github.com/superseriousbusiness/gotosocial/internal/concurrency"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/federation"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/text"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/internal/visibility"
)

// AccountProcessor wraps functionality for updating, creating, and deleting accounts in response to API requests.
//
// It also contains logic for actions towards accounts such as following, blocking, seeing follows, etc.
type AccountProcessor struct { //nolint:revive
	tc           typeutils.TypeConverter
	mediaManager media.Manager
	clientWorker *concurrency.WorkerPool[messages.FromClientAPI]
	oauthServer  oauth.Server
	filter       visibility.Filter
	formatter    text.Formatter
	db           db.DB
	federator    federation.Federator
	parseMention gtsmodel.ParseMentionFunc
}

// New returns a new account processor.
func New(
	db db.DB,
	tc typeutils.TypeConverter,
	mediaManager media.Manager,
	oauthServer oauth.Server,
	clientWorker *concurrency.WorkerPool[messages.FromClientAPI],
	federator federation.Federator,
	parseMention gtsmodel.ParseMentionFunc,
) AccountProcessor {
	return AccountProcessor{
		tc:           tc,
		mediaManager: mediaManager,
		clientWorker: clientWorker,
		oauthServer:  oauthServer,
		filter:       visibility.NewFilter(db),
		formatter:    text.NewFormatter(db),
		db:           db,
		federator:    federator,
		parseMention: parseMention,
	}
}

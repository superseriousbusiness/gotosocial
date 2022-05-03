/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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

package admin

import (
	"context"
	"mime/multipart"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/internal/worker"
)

// Processor wraps a bunch of functions for processing admin actions.
type Processor interface {
	DomainBlockCreate(ctx context.Context, account *gtsmodel.Account, domain string, obfuscate bool, publicComment string, privateComment string, subscriptionID string) (*apimodel.DomainBlock, gtserror.WithCode)
	DomainBlocksImport(ctx context.Context, account *gtsmodel.Account, domains *multipart.FileHeader) ([]*apimodel.DomainBlock, gtserror.WithCode)
	DomainBlocksGet(ctx context.Context, account *gtsmodel.Account, export bool) ([]*apimodel.DomainBlock, gtserror.WithCode)
	DomainBlockGet(ctx context.Context, account *gtsmodel.Account, id string, export bool) (*apimodel.DomainBlock, gtserror.WithCode)
	DomainBlockDelete(ctx context.Context, account *gtsmodel.Account, id string) (*apimodel.DomainBlock, gtserror.WithCode)
	AccountAction(ctx context.Context, account *gtsmodel.Account, form *apimodel.AdminAccountActionRequest) gtserror.WithCode
	EmojiCreate(ctx context.Context, account *gtsmodel.Account, user *gtsmodel.User, form *apimodel.EmojiCreateRequest) (*apimodel.Emoji, gtserror.WithCode)
	MediaRemotePrune(ctx context.Context, mediaRemoteCacheDays int) gtserror.WithCode
}

type processor struct {
	tc           typeutils.TypeConverter
	mediaManager media.Manager
	clientWorker *worker.Worker[messages.FromClientAPI]
	db           db.DB
}

// New returns a new admin processor.
func New(db db.DB, tc typeutils.TypeConverter, mediaManager media.Manager, clientWorker *worker.Worker[messages.FromClientAPI]) Processor {
	return &processor{
		tc:           tc,
		mediaManager: mediaManager,
		clientWorker: clientWorker,
		db:           db,
	}
}

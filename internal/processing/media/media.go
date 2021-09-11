/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

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

package media

import (
	"context"

	"git.iim.gay/grufwub/go-store/kv"
	"github.com/sirupsen/logrus"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
)

// Processor wraps a bunch of functions for processing media actions.
type Processor interface {
	// Create creates a new media attachment belonging to the given account, using the request form.
	Create(ctx context.Context, account *gtsmodel.Account, form *apimodel.AttachmentRequest) (*apimodel.Attachment, error)
	// Delete deletes the media attachment with the given ID, including all files pertaining to that attachment.
	Delete(ctx context.Context, mediaAttachmentID string) gtserror.WithCode
	GetFile(ctx context.Context, account *gtsmodel.Account, form *apimodel.GetContentRequestForm) (*apimodel.Content, error)
	GetMedia(ctx context.Context, account *gtsmodel.Account, mediaAttachmentID string) (*apimodel.Attachment, gtserror.WithCode)
	Update(ctx context.Context, account *gtsmodel.Account, mediaAttachmentID string, form *apimodel.AttachmentUpdateRequest) (*apimodel.Attachment, gtserror.WithCode)
}

type processor struct {
	tc           typeutils.TypeConverter
	config       *config.Config
	mediaHandler media.Handler
	storage      *kv.KVStore
	db           db.DB
	log          *logrus.Logger
}

// New returns a new media processor.
func New(db db.DB, tc typeutils.TypeConverter, mediaHandler media.Handler, storage *kv.KVStore, config *config.Config, log *logrus.Logger) Processor {
	return &processor{
		tc:           tc,
		config:       config,
		mediaHandler: mediaHandler,
		storage:      storage,
		db:           db,
		log:          log,
	}
}

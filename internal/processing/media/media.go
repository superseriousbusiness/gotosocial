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

package media

import (
	"context"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/storage"
	"github.com/superseriousbusiness/gotosocial/internal/transport"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
)

// Processor wraps a bunch of functions for processing media actions.
type Processor interface {
	// Create creates a new media attachment belonging to the given account, using the request form.
	Create(ctx context.Context, account *gtsmodel.Account, form *apimodel.AttachmentRequest) (*apimodel.Attachment, gtserror.WithCode)
	// Delete deletes the media attachment with the given ID, including all files pertaining to that attachment.
	Delete(ctx context.Context, mediaAttachmentID string) gtserror.WithCode
	// Unattach unattaches the media attachment with the given ID from any statuses it was attached to, making it available
	// for reattachment again.
	Unattach(ctx context.Context, account *gtsmodel.Account, mediaAttachmentID string) (*apimodel.Attachment, gtserror.WithCode)
	// GetFile retrieves a file from storage and streams it back to the caller via an io.reader embedded in *apimodel.Content.
	GetFile(ctx context.Context, account *gtsmodel.Account, form *apimodel.GetContentRequestForm) (*apimodel.Content, gtserror.WithCode)
	GetCustomEmojis(ctx context.Context) ([]*apimodel.Emoji, gtserror.WithCode)
	GetMedia(ctx context.Context, account *gtsmodel.Account, mediaAttachmentID string) (*apimodel.Attachment, gtserror.WithCode)
	Update(ctx context.Context, account *gtsmodel.Account, mediaAttachmentID string, form *apimodel.AttachmentUpdateRequest) (*apimodel.Attachment, gtserror.WithCode)
}

type processor struct {
	tc                  typeutils.TypeConverter
	mediaManager        media.Manager
	transportController transport.Controller
	storage             *storage.Driver
	db                  db.DB
}

// New returns a new media processor.
func New(db db.DB, tc typeutils.TypeConverter, mediaManager media.Manager, transportController transport.Controller, storage *storage.Driver) Processor {
	return &processor{
		tc:                  tc,
		mediaManager:        mediaManager,
		transportController: transportController,
		storage:             storage,
		db:                  db,
	}
}

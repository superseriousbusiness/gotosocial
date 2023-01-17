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

package media

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/transport"
	"github.com/superseriousbusiness/gotosocial/internal/uris"
)

func (p *processor) GetFile(ctx context.Context, requestingAccount *gtsmodel.Account, form *apimodel.GetContentRequestForm) (*apimodel.Content, gtserror.WithCode) {
	// parse the form fields
	mediaSize, err := media.ParseMediaSize(form.MediaSize)
	if err != nil {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("media size %s not valid", form.MediaSize))
	}

	mediaType, err := media.ParseMediaType(form.MediaType)
	if err != nil {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("media type %s not valid", form.MediaType))
	}

	spl := strings.Split(form.FileName, ".")
	if len(spl) != 2 || spl[0] == "" || spl[1] == "" {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("file name %s not parseable", form.FileName))
	}
	wantedMediaID := spl[0]
	owningAccountID := form.AccountID

	// get the account that owns the media and make sure it's not suspended
	owningAccount, err := p.db.GetAccountByID(ctx, owningAccountID)
	if err != nil {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("account with id %s could not be selected from the db: %s", owningAccountID, err))
	}
	if !owningAccount.SuspendedAt.IsZero() {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("account with id %s is suspended", owningAccountID))
	}

	// make sure the requesting account and the media account don't block each other
	if requestingAccount != nil {
		blocked, err := p.db.IsBlocked(ctx, requestingAccount.ID, owningAccountID, true)
		if err != nil {
			return nil, gtserror.NewErrorNotFound(fmt.Errorf("block status could not be established between accounts %s and %s: %s", owningAccountID, requestingAccount.ID, err))
		}
		if blocked {
			return nil, gtserror.NewErrorNotFound(fmt.Errorf("block exists between accounts %s and %s", owningAccountID, requestingAccount.ID))
		}
	}

	// the way we store emojis is a little different from the way we store other attachments,
	// so we need to take different steps depending on the media type being requested
	switch mediaType {
	case media.TypeEmoji:
		return p.getEmojiContent(ctx, wantedMediaID, owningAccountID, mediaSize)
	case media.TypeAttachment, media.TypeHeader, media.TypeAvatar:
		return p.getAttachmentContent(ctx, requestingAccount, wantedMediaID, owningAccountID, mediaSize)
	default:
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("media type %s not recognized", mediaType))
	}
}

func (p *processor) getAttachmentContent(ctx context.Context, requestingAccount *gtsmodel.Account, wantedMediaID string, owningAccountID string, mediaSize media.Size) (*apimodel.Content, gtserror.WithCode) {
	// retrieve attachment from the database and do basic checks on it
	a, err := p.db.GetAttachmentByID(ctx, wantedMediaID)
	if err != nil {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("attachment %s could not be taken from the db: %s", wantedMediaID, err))
	}

	if a.AccountID != owningAccountID {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("attachment %s is not owned by %s", wantedMediaID, owningAccountID))
	}

	if !*a.Cached {
		// if we don't have it cached, then we can assume two things:
		// 1. this is remote media, since local media should never be uncached
		// 2. we need to fetch it again using a transport and the media manager
		remoteMediaIRI, err := url.Parse(a.RemoteURL)
		if err != nil {
			return nil, gtserror.NewErrorNotFound(fmt.Errorf("error parsing remote media iri %s: %s", a.RemoteURL, err))
		}

		// use an empty string as requestingUsername to use the instance account, unless the request for this
		// media has been http signed, then use the requesting account to make the request to remote server
		var requestingUsername string
		if requestingAccount != nil {
			requestingUsername = requestingAccount.Username
		}

		// Pour one out for tobi's original streamed recache
		// (streaming data both to the client and storage).
		// Gone and forever missed <3
		//
		// [
		//   the reason it was removed was because a slow
		//   client connection could hold open a storage
		//   recache operation, and so holding open a media
		//   worker worker.
		// ]

		dataFn := func(innerCtx context.Context) (io.ReadCloser, int64, error) {
			t, err := p.transportController.NewTransportForUsername(innerCtx, requestingUsername)
			if err != nil {
				return nil, 0, err
			}
			return t.DereferenceMedia(transport.WithFastfail(innerCtx), remoteMediaIRI)
		}

		// Start recaching this media with the prepared data function.
		processingMedia, err := p.mediaManager.RecacheMedia(ctx, dataFn, nil, wantedMediaID)
		if err != nil {
			return nil, gtserror.NewErrorNotFound(fmt.Errorf("error recaching media: %s", err))
		}

		// Load attachment and block until complete
		a, err = processingMedia.LoadAttachment(ctx)
		if err != nil {
			return nil, gtserror.NewErrorNotFound(fmt.Errorf("error loading recached attachment: %s", err))
		}
	}

	var (
		storagePath       string
		attachmentContent = &apimodel.Content{
			ContentUpdated: a.UpdatedAt,
		}
	)

	// get file information from the attachment depending on the requested media size
	switch mediaSize {
	case media.SizeOriginal:
		attachmentContent.ContentType = a.File.ContentType
		attachmentContent.ContentLength = int64(a.File.FileSize)
		storagePath = a.File.Path
	case media.SizeSmall:
		attachmentContent.ContentType = a.Thumbnail.ContentType
		attachmentContent.ContentLength = int64(a.Thumbnail.FileSize)
		storagePath = a.Thumbnail.Path
	default:
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("media size %s not recognized for attachment", mediaSize))
	}

	// ... so now we can safely return it
	return p.retrieveFromStorage(ctx, storagePath, attachmentContent)
}

func (p *processor) getEmojiContent(ctx context.Context, fileName string, owningAccountID string, emojiSize media.Size) (*apimodel.Content, gtserror.WithCode) {
	emojiContent := &apimodel.Content{}
	var storagePath string

	// reconstruct the static emoji image url -- reason
	// for using the static URL rather than full size url
	// is that static emojis are always encoded as png,
	// so this is more reliable than using full size url
	imageStaticURL := uris.GenerateURIForAttachment(owningAccountID, string(media.TypeEmoji), string(media.SizeStatic), fileName, "png")

	e, err := p.db.GetEmojiByStaticURL(ctx, imageStaticURL)
	if err != nil {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("emoji %s could not be taken from the db: %s", fileName, err))
	}

	if *e.Disabled {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("emoji %s has been disabled", fileName))
	}

	switch emojiSize {
	case media.SizeOriginal:
		emojiContent.ContentType = e.ImageContentType
		emojiContent.ContentLength = int64(e.ImageFileSize)
		storagePath = e.ImagePath
	case media.SizeStatic:
		emojiContent.ContentType = e.ImageStaticContentType
		emojiContent.ContentLength = int64(e.ImageStaticFileSize)
		storagePath = e.ImageStaticPath
	default:
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("media size %s not recognized for emoji", emojiSize))
	}

	return p.retrieveFromStorage(ctx, storagePath, emojiContent)
}

func (p *processor) retrieveFromStorage(ctx context.Context, storagePath string, content *apimodel.Content) (*apimodel.Content, gtserror.WithCode) {
	if url := p.storage.URL(ctx, storagePath); url != nil {
		content.URL = url
		return content, nil
	}
	reader, err := p.storage.GetStream(ctx, storagePath)
	if err != nil {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("error retrieving from storage: %s", err))
	}

	content.Content = reader
	return content, nil
}

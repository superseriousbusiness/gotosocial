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
	"fmt"
	"io"
	"net/url"
	"strings"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/iotools"
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
	attachmentContent := &apimodel.Content{}
	var storagePath string

	// retrieve attachment from the database and do basic checks on it
	a, err := p.db.GetAttachmentByID(ctx, wantedMediaID)
	if err != nil {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("attachment %s could not be taken from the db: %s", wantedMediaID, err))
	}

	if a.AccountID != owningAccountID {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("attachment %s is not owned by %s", wantedMediaID, owningAccountID))
	}

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

	// if we have the media cached on our server already, we can now simply return it from storage
	if *a.Cached {
		return p.retrieveFromStorage(ctx, storagePath, attachmentContent)
	}

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

	var data media.DataFunc

	if mediaSize == media.SizeSmall {
		// if it's the thumbnail that's requested then the user will have to wait a bit while we process the
		// large version and derive a thumbnail from it, so use the normal recaching procedure: fetch the media,
		// process it, then return the thumbnail data
		data = func(innerCtx context.Context) (io.ReadCloser, int64, error) {
			t, err := p.transportController.NewTransportForUsername(innerCtx, requestingUsername)
			if err != nil {
				return nil, 0, err
			}
			return t.DereferenceMedia(transport.WithFastfail(innerCtx), remoteMediaIRI)
		}
	} else {
		// if it's the full-sized version being requested, we can cheat a bit by streaming data to the user as
		// it's retrieved from the remote server, using tee; this saves the user from having to wait while
		// we process the media on our side
		//
		// this looks a bit like this:
		//
		//                http fetch                       pipe
		// remote server ------------> data function ----------------> api caller
		//                                   |
		//                                   | tee
		//                                   |
		//                                   ▼
		//                            instance storage

		// This pipe will connect the caller to the in-process media retrieval...
		pipeReader, pipeWriter := io.Pipe()

		// Wrap the output pipe to silence any errors during the actual media
		// streaming process. We catch the error later but they must be silenced
		// during stream to prevent interruptions to storage of the actual media.
		silencedWriter := iotools.SilenceWriter(pipeWriter)

		// Pass the reader side of the pipe to the caller to slurp from.
		attachmentContent.Content = pipeReader

		// Create a data function which injects the writer end of the pipe
		// into the data retrieval process. If something goes wrong while
		// doing the data retrieval, we hang up the underlying pipeReader
		// to indicate to the caller that no data is available. It's up to
		// the caller of this processor function to handle that gracefully.
		data = func(innerCtx context.Context) (io.ReadCloser, int64, error) {
			t, err := p.transportController.NewTransportForUsername(innerCtx, requestingUsername)
			if err != nil {
				// propagate the transport error to read end of pipe.
				_ = pipeWriter.CloseWithError(fmt.Errorf("error getting transport for user: %w", err))
				return nil, 0, err
			}

			readCloser, fileSize, err := t.DereferenceMedia(transport.WithFastfail(innerCtx), remoteMediaIRI)
			if err != nil {
				// propagate the dereference error to read end of pipe.
				_ = pipeWriter.CloseWithError(fmt.Errorf("error dereferencing media: %w", err))
				return nil, 0, err
			}

			// Make a TeeReader so that everything read from the readCloser,
			// aka the remote instance, will also be written into the pipe.
			teeReader := io.TeeReader(readCloser, silencedWriter)

			// Wrap teereader to implement original readcloser's close,
			// and also ensuring that we close the pipe from write end.
			return iotools.ReadFnCloser(teeReader, func() error {
				defer func() {
					// We use the error (if any) encountered by the
					// silenced writer to close connection to make sure it
					// gets propagated to the attachment.Content reader.
					_ = pipeWriter.CloseWithError(silencedWriter.Error())
				}()

				return readCloser.Close()
			}), fileSize, nil
		}
	}

	// put the media recached in the queue
	processingMedia, err := p.mediaManager.RecacheMedia(ctx, data, nil, wantedMediaID)
	if err != nil {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("error recaching media: %s", err))
	}

	// if it's the thumbnail, stream the processed thumbnail from storage, after waiting for processing to finish
	if mediaSize == media.SizeSmall {
		// below function call blocks until all processing on the attachment has finished...
		if _, err := processingMedia.LoadAttachment(ctx); err != nil {
			return nil, gtserror.NewErrorNotFound(fmt.Errorf("error loading recached attachment: %s", err))
		}
		// ... so now we can safely return it
		return p.retrieveFromStorage(ctx, storagePath, attachmentContent)
	}

	return attachmentContent, nil
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

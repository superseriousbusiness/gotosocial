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
	"bufio"
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/media"
)

func (p *processor) GetFile(ctx context.Context, account *gtsmodel.Account, form *apimodel.GetContentRequestForm) (*apimodel.Content, gtserror.WithCode) {
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
	expectedAccountID := form.AccountID

	// get the account that owns the media and make sure it's not suspended
	acct, err := p.db.GetAccountByID(ctx, expectedAccountID)
	if err != nil {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("account with id %s could not be selected from the db: %s", expectedAccountID, err))
	}
	if !acct.SuspendedAt.IsZero() {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("account with id %s is suspended", expectedAccountID))
	}

	// make sure the requesting account and the media account don't block each other
	if account != nil {
		blocked, err := p.db.IsBlocked(ctx, account.ID, expectedAccountID, true)
		if err != nil {
			return nil, gtserror.NewErrorNotFound(fmt.Errorf("block status could not be established between accounts %s and %s: %s", expectedAccountID, account.ID, err))
		}
		if blocked {
			return nil, gtserror.NewErrorNotFound(fmt.Errorf("block exists between accounts %s and %s", expectedAccountID, account.ID))
		}
	}

	// the way we store emojis is a little different from the way we store other attachments,
	// so we need to take different steps depending on the media type being requested
	switch mediaType {
	case media.TypeEmoji:
		return p.getEmojiContent(ctx, wantedMediaID, mediaSize)
	case media.TypeAttachment, media.TypeHeader, media.TypeAvatar:
		return p.getAttachmentContent(ctx, account, wantedMediaID, expectedAccountID, mediaSize)
	default:
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("media type %s not recognized", mediaType))
	}
}

func (p *processor) getAttachmentContent(ctx context.Context, requestingAccount *gtsmodel.Account, wantedMediaID string, expectedAccountID string, mediaSize media.Size) (*apimodel.Content, gtserror.WithCode) {
	attachmentContent := &apimodel.Content{}
	var storagePath string

	// retrieve attachment from the database and do basic checks on it
	a, err := p.db.GetAttachmentByID(ctx, wantedMediaID)
	if err != nil {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("attachment %s could not be taken from the db: %s", wantedMediaID, err))
	}

	if a.AccountID != expectedAccountID {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("attachment %s is not owned by %s", wantedMediaID, expectedAccountID))
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
	if a.Cached {
		return p.streamFromStorage(storagePath, attachmentContent)
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
	var postDataCallback media.PostDataCallbackFunc

	if mediaSize == media.SizeSmall {
		// if it's the thumbnail that's requested then the user will have to wait a bit while we process the
		// large version and derive a thumbnail from it, so use the normal recaching procedure: fetch the media,
		// process it, then return the thumbnail data
		data = func(innerCtx context.Context) (io.Reader, int, error) {
			transport, err := p.transportController.NewTransportForUsername(innerCtx, requestingUsername)
			if err != nil {
				return nil, 0, err
			}
			return transport.DereferenceMedia(innerCtx, remoteMediaIRI)
		}
	} else {
		// if it's the full-sized version being requested, we can cheat a bit by streaming data to the user as
		// it's retrieved from the remote server, using tee; this saves the user from having to wait while
		// we process the media on our side
		//
		// this looks a bit like this:
		//
		//                http fetch                   buffered pipe
		// remote server ------------> data function ----------------> api caller
		//                                   |
		//                                   | tee
		//                                   |
		//                                   ▼
		//                            instance storage

		// Buffer each end of the pipe, so that if the caller drops the connection during the flow, the tee
		// reader can continue without having to worry about tee-ing into a closed or blocked pipe.
		pipeReader, pipeWriter := io.Pipe()
		bufferedWriter := bufio.NewWriterSize(pipeWriter, int(attachmentContent.ContentLength))
		bufferedReader := bufio.NewReaderSize(pipeReader, int(attachmentContent.ContentLength))

		// the caller will read from the buffered reader, so it doesn't matter if they drop out without reading everything
		attachmentContent.Content = bufferedReader

		data = func(innerCtx context.Context) (io.Reader, int, error) {
			transport, err := p.transportController.NewTransportForUsername(innerCtx, requestingUsername)
			if err != nil {
				return nil, 0, err
			}

			readCloser, fileSize, err := transport.DereferenceMedia(innerCtx, remoteMediaIRI)
			if err != nil {
				return nil, 0, err
			}

			// Make a TeeReader so that everything read from the readCloser by the media manager will be written into the bufferedWriter.
			// We wrap this in a teeReadCloser which implements io.ReadCloser, so that whoever uses the teeReader can close the readCloser
			// when they're done with it.
			trc := teeReadCloser{
				teeReader: io.TeeReader(readCloser, bufferedWriter),
				close:     readCloser.Close,
			}

			return trc, fileSize, nil
		}

		// close the pipewriter after data has been piped into it, so the reader on the other side doesn't block;
		// we don't need to close the reader here because that's the caller's responsibility
		postDataCallback = func(innerCtx context.Context) error {
			// flush the buffered writer into the buffer of the reader...
			if err := bufferedWriter.Flush(); err != nil {
				return err
			}

			// and close the underlying pipe writer
			if err := pipeWriter.Close(); err != nil {
				return err
			}

			return nil
		}
	}

	// put the media recached in the queue
	processingMedia, err := p.mediaManager.RecacheMedia(ctx, data, postDataCallback, wantedMediaID)
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
		return p.streamFromStorage(storagePath, attachmentContent)
	}

	return attachmentContent, nil
}

func (p *processor) getEmojiContent(ctx context.Context, wantedEmojiID string, emojiSize media.Size) (*apimodel.Content, gtserror.WithCode) {
	emojiContent := &apimodel.Content{}
	var storagePath string

	e := &gtsmodel.Emoji{}
	if err := p.db.GetByID(ctx, wantedEmojiID, e); err != nil {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("emoji %s could not be taken from the db: %s", wantedEmojiID, err))
	}

	if e.Disabled {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("emoji %s has been disabled", wantedEmojiID))
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

	return p.streamFromStorage(storagePath, emojiContent)
}

func (p *processor) streamFromStorage(storagePath string, content *apimodel.Content) (*apimodel.Content, gtserror.WithCode) {
	reader, err := p.storage.GetStream(storagePath)
	if err != nil {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("error retrieving from storage: %s", err))
	}

	content.Content = reader
	return content, nil
}

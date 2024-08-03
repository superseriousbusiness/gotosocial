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

package media

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/regexes"
	"github.com/superseriousbusiness/gotosocial/internal/storage"
	"github.com/superseriousbusiness/gotosocial/internal/uris"
)

// GetFile retrieves a file from storage and streams it back
// to the caller via an io.reader embedded in *apimodel.Content.
func (p *Processor) GetFile(
	ctx context.Context,
	requester *gtsmodel.Account,
	form *apimodel.GetContentRequestForm,
) (*apimodel.Content, gtserror.WithCode) {
	// Parse media size (small, static, original).
	mediaSize, err := parseSize(form.MediaSize)
	if err != nil {
		err := gtserror.Newf("media size %s not valid", form.MediaSize)
		return nil, gtserror.NewErrorNotFound(err)
	}

	// Parse media type (emoji, header, avatar, attachment).
	mediaType, err := parseType(form.MediaType)
	if err != nil {
		err := gtserror.Newf("media type %s not valid", form.MediaType)
		return nil, gtserror.NewErrorNotFound(err)
	}

	// Parse media ID from file name.
	mediaID, _, err := parseFileName(form.FileName)
	if err != nil {
		err := gtserror.Newf("media file name %s not valid", form.FileName)
		return nil, gtserror.NewErrorNotFound(err)
	}

	// Get the account that owns the media
	// and make sure it's not suspended.
	acctID := form.AccountID
	acct, err := p.state.DB.GetAccountByID(ctx, acctID)
	if err != nil {
		err := gtserror.Newf("db error getting account %s: %w", acctID, err)
		return nil, gtserror.NewErrorNotFound(err)
	}

	if acct.IsSuspended() {
		err := gtserror.Newf("account %s is suspended", acctID)
		return nil, gtserror.NewErrorNotFound(err)
	}

	// If requester was authenticated, ensure media
	// owner and requester don't block each other.
	if requester != nil {
		blocked, err := p.state.DB.IsEitherBlocked(ctx, requester.ID, acctID)
		if err != nil {
			err := gtserror.Newf("db error checking block between %s and %s: %w", acctID, requester.ID, err)
			return nil, gtserror.NewErrorNotFound(err)
		}

		if blocked {
			err := gtserror.Newf("block exists between %s and %s", acctID, requester.ID)
			return nil, gtserror.NewErrorNotFound(err)
		}
	}

	// The way we store emojis is a bit different
	// from the way we store other attachments,
	// so we need to take different steps depending
	// on the media type being requested.
	switch mediaType {

	case media.TypeEmoji:
		return p.getEmojiContent(ctx,
			acctID,
			mediaSize,
			mediaID,
		)

	case media.TypeAttachment, media.TypeHeader, media.TypeAvatar:
		return p.getAttachmentContent(ctx,
			requester,
			acctID,
			mediaSize,
			mediaID,
		)

	default:
		err := gtserror.Newf("media type %s not recognized", mediaType)
		return nil, gtserror.NewErrorNotFound(err)
	}
}

func (p *Processor) getAttachmentContent(
	ctx context.Context,
	requester *gtsmodel.Account,
	acctID string,
	sizeStr media.Size,
	mediaID string,
) (
	*apimodel.Content,
	gtserror.WithCode,
) {
	// Get attachment with given ID from the database.
	attach, err := p.state.DB.GetAttachmentByID(ctx, mediaID)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error getting attachment %s: %w", mediaID, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if attach == nil {
		const text = "media not found"
		return nil, gtserror.NewErrorNotFound(errors.New(text), text)
	}

	// Ensure the account
	// actually owns the media.
	if attach.AccountID != acctID {
		const text = "media was not owned by passed account id"
		return nil, gtserror.NewErrorNotFound(errors.New(text) /* no help text! */)
	}

	// Unknown file types indicate no *locally*
	// stored data we can serve. Handle separately.
	if attach.Type == gtsmodel.FileTypeUnknown {
		return handleUnknown(attach)
	}

	// If requester was provided, use their username
	// to create a transport to potentially re-fetch
	// the media. Else falls back to instance account.
	var requestUser string
	if requester != nil {
		requestUser = requester.Username
	}

	// Ensure that stored media is cached.
	// (this handles local media / recaches).
	attach, err = p.federator.RefreshMedia(
		ctx,
		requestUser,
		attach,
		media.AdditionalMediaInfo{},
		false,
	)
	if err != nil {
		err := gtserror.Newf("error recaching media: %w", err)
		return nil, gtserror.NewErrorNotFound(err)
	}

	// Start preparing API content model.
	apiContent := &apimodel.Content{
		ContentUpdated: attach.UpdatedAt,
	}

	// Retrieve appropriate
	// size file from storage.
	switch sizeStr {

	case media.SizeOriginal:
		apiContent.ContentType = attach.File.ContentType
		apiContent.ContentLength = int64(attach.File.FileSize)
		return p.getContent(ctx,
			attach.File.Path,
			apiContent,
		)

	case media.SizeSmall:
		apiContent.ContentType = attach.Thumbnail.ContentType
		apiContent.ContentLength = int64(attach.Thumbnail.FileSize)
		return p.getContent(ctx,
			attach.Thumbnail.Path,
			apiContent,
		)

	default:
		const text = "invalid media attachment size"
		return nil, gtserror.NewErrorBadRequest(errors.New(text), text)
	}
}

func (p *Processor) getEmojiContent(
	ctx context.Context,
	acctID string,
	sizeStr media.Size,
	emojiID string,
) (
	*apimodel.Content,
	gtserror.WithCode,
) {
	// Reconstruct static emoji image URL to search for it.
	// As refreshed emojis use a newly generated path ID to
	// differentiate them (cache-wise) from the original.
	staticURL := uris.URIForAttachment(
		acctID,
		string(media.TypeEmoji),
		string(media.SizeStatic),
		emojiID,
		"png",
	)

	// Search for emoji with given static URL in the database.
	emoji, err := p.state.DB.GetEmojiByStaticURL(ctx, staticURL)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("error fetching emoji from database: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if emoji == nil {
		const text = "emoji not found"
		return nil, gtserror.NewErrorNotFound(errors.New(text), text)
	}

	if *emoji.Disabled {
		const text = "emoji has been disabled"
		return nil, gtserror.NewErrorNotFound(errors.New(text), text)
	}

	// Ensure that stored emoji is cached.
	// (this handles local emoji / recaches).
	emoji, err = p.federator.RecacheEmoji(
		ctx,
		emoji,
	)
	if err != nil {
		err := gtserror.Newf("error recaching emoji: %w", err)
		return nil, gtserror.NewErrorNotFound(err)
	}

	// Start preparing API content model.
	apiContent := &apimodel.Content{}

	// Retrieve appropriate
	// size file from storage.
	switch sizeStr {

	case media.SizeOriginal:
		apiContent.ContentType = emoji.ImageContentType
		apiContent.ContentLength = int64(emoji.ImageFileSize)
		return p.getContent(ctx,
			emoji.ImagePath,
			apiContent,
		)

	case media.SizeStatic:
		apiContent.ContentType = emoji.ImageStaticContentType
		apiContent.ContentLength = int64(emoji.ImageStaticFileSize)
		return p.getContent(ctx,
			emoji.ImageStaticPath,
			apiContent,
		)

	default:
		const text = "invalid media attachment size"
		return nil, gtserror.NewErrorBadRequest(errors.New(text), text)
	}
}

// getContent performs the final file fetching of
// stored content at path in storage. This is
// populated in the apimodel.Content{} and returned.
// (note: this also handles un-proxied S3 storage).
func (p *Processor) getContent(
	ctx context.Context,
	path string,
	content *apimodel.Content,
) (
	*apimodel.Content,
	gtserror.WithCode,
) {
	// If running on S3 storage with proxying disabled then
	// just fetch pre-signed URL instead of the content.
	if url := p.state.Storage.URL(ctx, path); url != nil {
		content.URL = url
		return content, nil
	}

	// Fetch file stream for the stored media at path.
	rc, err := p.state.Storage.GetStream(ctx, path)
	if err != nil && !storage.IsNotFound(err) {
		err := gtserror.Newf("error getting file %s from storage: %w", path, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Ensure found.
	if rc == nil {
		err := gtserror.Newf("file not found at %s", path)
		const text = "file not found"
		return nil, gtserror.NewErrorNotFound(err, text)
	}

	// Return with stream.
	content.Content = rc
	return content, nil
}

// handles serving Content for "unknown" file
// type, ie., a file we couldn't cache (this time).
func handleUnknown(
	attach *gtsmodel.MediaAttachment,
) (*apimodel.Content, gtserror.WithCode) {
	if attach.RemoteURL == "" {
		err := gtserror.Newf("empty remote url for %s", attach.ID)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Parse media remote URL to valid URL object.
	remoteURL, err := url.Parse(attach.RemoteURL)
	if err != nil {
		err := gtserror.Newf("invalid remote url for %s: %w", attach.ID, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if remoteURL == nil {
		err := gtserror.Newf("nil remote url for %s", attach.ID)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Just forward the request to the remote URL,
	// since this is a type we couldn't process.
	url := &storage.PresignedURL{
		URL: remoteURL,

		// We might manage to cache the media
		// at some point, so set a low-ish expiry.
		Expiry: time.Now().Add(2 * time.Hour),
	}

	return &apimodel.Content{URL: url}, nil
}

func parseType(s string) (media.Type, error) {
	switch s {
	case string(media.TypeAttachment):
		return media.TypeAttachment, nil
	case string(media.TypeHeader):
		return media.TypeHeader, nil
	case string(media.TypeAvatar):
		return media.TypeAvatar, nil
	case string(media.TypeEmoji):
		return media.TypeEmoji, nil
	}
	return "", fmt.Errorf("%s not a recognized media.Type", s)
}

func parseSize(s string) (media.Size, error) {
	switch s {
	case string(media.SizeSmall):
		return media.SizeSmall, nil
	case string(media.SizeOriginal):
		return media.SizeOriginal, nil
	case string(media.SizeStatic):
		return media.SizeStatic, nil
	}
	return "", fmt.Errorf("%s not a recognized media.Size", s)
}

// Extract the mediaID and file extension from
// a string like "01J3CTH8CZ6ATDNMG6CPRC36XE.gif"
func parseFileName(s string) (string, string, error) {
	spl := strings.Split(s, ".")
	if len(spl) != 2 || spl[0] == "" || spl[1] == "" {
		return "", "", errors.New("file name not splittable on '.'")
	}

	var (
		mediaID  = spl[0]
		mediaExt = spl[1]
	)

	if !regexes.ULID.MatchString(mediaID) {
		return "", "", fmt.Errorf("%s not a valid ULID", mediaID)
	}

	return mediaID, mediaExt, nil
}

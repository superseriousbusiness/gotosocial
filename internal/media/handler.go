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
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"codeberg.org/gruf/go-store/kv"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/transport"
	"github.com/superseriousbusiness/gotosocial/internal/uris"
)

// EmojiMaxBytes is the maximum permitted bytes of an emoji upload (50kb)
const EmojiMaxBytes = 51200

type Size string

const (
	SizeSmall    Size = "small"    // SizeSmall is the key for small/thumbnail versions of media
	SizeOriginal Size = "original" // SizeOriginal is the key for original/fullsize versions of media and emoji
	SizeStatic   Size = "static"   // SizeStatic is the key for static (non-animated) versions of emoji
)

type Type string

const (
	TypeAttachment Type = "attachment" // TypeAttachment is the key for media attachments
	TypeHeader     Type = "header"     // TypeHeader is the key for profile header requests
	TypeAvatar     Type = "avatar"     // TypeAvatar is the key for profile avatar requests
	TypeEmoji      Type = "emoji"      // TypeEmoji is the key for emoji type requests
)

// Handler provides an interface for parsing, storing, and retrieving media objects like photos, videos, and gifs.
type Handler interface {
	// ProcessHeaderOrAvatar takes a new header image for an account, checks it out, removes exif data from it,
	// puts it in whatever storage backend we're using, sets the relevant fields in the database for the new image,
	// and then returns information to the caller about the new header.
	ProcessHeaderOrAvatar(ctx context.Context, attachment []byte, accountID string, mediaType Type, remoteURL string) (*gtsmodel.MediaAttachment, error)

	// ProcessLocalAttachment takes a new attachment and the requesting account, checks it out, removes exif data from it,
	// puts it in whatever storage backend we're using, sets the relevant fields in the database for the new media,
	// and then returns information to the caller about the attachment. It's the caller's responsibility to put the returned struct
	// in the database.
	ProcessAttachment(ctx context.Context, attachmentBytes []byte, minAttachment *gtsmodel.MediaAttachment) (*gtsmodel.MediaAttachment, error)

	// ProcessLocalEmoji takes a new emoji and a shortcode, cleans it up, puts it in storage, and creates a new
	// *gts.Emoji for it, then returns it to the caller. It's the caller's responsibility to put the returned struct
	// in the database.
	ProcessLocalEmoji(ctx context.Context, emojiBytes []byte, shortcode string) (*gtsmodel.Emoji, error)

	ProcessRemoteHeaderOrAvatar(ctx context.Context, t transport.Transport, currentAttachment *gtsmodel.MediaAttachment, accountID string) (*gtsmodel.MediaAttachment, error)
}

type mediaHandler struct {
	db      db.DB
	storage *kv.KVStore
}

// New returns a new handler with the given db and storage
func New(database db.DB, storage *kv.KVStore) Handler {
	return &mediaHandler{
		db:      database,
		storage: storage,
	}
}

/*
	INTERFACE FUNCTIONS
*/

// ProcessHeaderOrAvatar takes a new header image for an account, checks it out, removes exif data from it,
// puts it in whatever storage backend we're using, sets the relevant fields in the database for the new image,
// and then returns information to the caller about the new header.
func (mh *mediaHandler) ProcessHeaderOrAvatar(ctx context.Context, attachment []byte, accountID string, mediaType Type, remoteURL string) (*gtsmodel.MediaAttachment, error) {
	l := logrus.WithField("func", "SetHeaderForAccountID")

	if mediaType != TypeHeader && mediaType != TypeAvatar {
		return nil, errors.New("header or avatar not selected")
	}

	// make sure we have a type we can handle
	contentType, err := parseContentType(attachment)
	if err != nil {
		return nil, err
	}
	if !SupportedImageType(contentType) {
		return nil, fmt.Errorf("%s is not an accepted image type", contentType)
	}

	if len(attachment) == 0 {
		return nil, fmt.Errorf("passed reader was of size 0")
	}
	l.Tracef("read %d bytes of file", len(attachment))

	// process it
	ma, err := mh.processHeaderOrAvi(attachment, contentType, mediaType, accountID, remoteURL)
	if err != nil {
		return nil, fmt.Errorf("error processing %s: %s", mediaType, err)
	}

	// set it in the database
	if err := mh.db.SetAccountHeaderOrAvatar(ctx, ma, accountID); err != nil {
		return nil, fmt.Errorf("error putting %s in database: %s", mediaType, err)
	}

	return ma, nil
}

// ProcessAttachment takes a new attachment and the owning account, checks it out, removes exif data from it,
// puts it in whatever storage backend we're using, sets the relevant fields in the database for the new media,
// and then returns information to the caller about the attachment.
func (mh *mediaHandler) ProcessAttachment(ctx context.Context, attachmentBytes []byte, minAttachment *gtsmodel.MediaAttachment) (*gtsmodel.MediaAttachment, error) {
	contentType, err := parseContentType(attachmentBytes)
	if err != nil {
		return nil, err
	}

	minAttachment.File.ContentType = contentType

	mainType := strings.Split(contentType, "/")[0]
	switch mainType {
	// case MIMEVideo:
	// 	if !SupportedVideoType(contentType) {
	// 		return nil, fmt.Errorf("video type %s not supported", contentType)
	// 	}
	// 	if len(attachment) == 0 {
	// 		return nil, errors.New("video was of size 0")
	// 	}
	// 	return mh.processVideoAttachment(attachment, accountID, contentType, remoteURL)
	case MIMEImage:
		if !SupportedImageType(contentType) {
			return nil, fmt.Errorf("image type %s not supported", contentType)
		}
		if len(attachmentBytes) == 0 {
			return nil, errors.New("image was of size 0")
		}
		return mh.processImageAttachment(attachmentBytes, minAttachment)
	default:
		break
	}
	return nil, fmt.Errorf("content type %s not (yet) supported", contentType)
}

// ProcessLocalEmoji takes a new emoji and a shortcode, cleans it up, puts it in storage, and creates a new
// *gts.Emoji for it, then returns it to the caller. It's the caller's responsibility to put the returned struct
// in the database.
func (mh *mediaHandler) ProcessLocalEmoji(ctx context.Context, emojiBytes []byte, shortcode string) (*gtsmodel.Emoji, error) {
	var clean []byte
	var err error
	var original *imageAndMeta
	var static *imageAndMeta

	// check content type of the submitted emoji and make sure it's supported by us
	contentType, err := parseContentType(emojiBytes)
	if err != nil {
		return nil, err
	}
	if !supportedEmojiType(contentType) {
		return nil, fmt.Errorf("content type %s not supported for emojis", contentType)
	}

	if len(emojiBytes) == 0 {
		return nil, errors.New("emoji was of size 0")
	}
	if len(emojiBytes) > EmojiMaxBytes {
		return nil, fmt.Errorf("emoji size %d bytes exceeded max emoji size of %d bytes", len(emojiBytes), EmojiMaxBytes)
	}

	// clean any exif data from png but leave gifs alone
	switch contentType {
	case MIMEPng:
		if clean, err = purgeExif(emojiBytes); err != nil {
			return nil, fmt.Errorf("error cleaning exif data: %s", err)
		}
	case MIMEGif:
		clean = emojiBytes
	default:
		return nil, errors.New("media type unrecognized")
	}

	// unlike with other attachments we don't need to derive anything here because we don't care about the width/height etc
	original = &imageAndMeta{
		image: clean,
	}

	static, err = deriveStaticEmoji(clean, contentType)
	if err != nil {
		return nil, fmt.Errorf("error deriving static emoji: %s", err)
	}

	// since emoji aren't 'owned' by an account, but we still want to use the same pattern for serving them through the filserver,
	// (ie., fileserver/ACCOUNT_ID/etc etc) we need to fetch the INSTANCE ACCOUNT from the database. That is, the account that's created
	// with the same username as the instance hostname, which doesn't belong to any particular user.
	instanceAccount, err := mh.db.GetInstanceAccount(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("error fetching instance account: %s", err)
	}

	// the file extension (either png or gif)
	extension := strings.Split(contentType, "/")[1]

	// generate a ulid for the new emoji
	newEmojiID, err := id.NewRandomULID()
	if err != nil {
		return nil, err
	}

	// activitypub uri for the emoji -- unrelated to actually serving the image
	// will be something like https://example.org/emoji/01FPSVBK3H8N7V8XK6KGSQ86EC
	emojiURI := uris.GenerateURIForEmoji(newEmojiID)

	// serve url and storage path for the original emoji -- can be png or gif
	emojiURL := uris.GenerateURIForAttachment(instanceAccount.ID, string(TypeEmoji), string(SizeOriginal), newEmojiID, extension)
	emojiPath := fmt.Sprintf("%s/%s/%s/%s.%s", instanceAccount.ID, TypeEmoji, SizeOriginal, newEmojiID, extension)

	// serve url and storage path for the static version -- will always be png
	emojiStaticURL := uris.GenerateURIForAttachment(instanceAccount.ID, string(TypeEmoji), string(SizeStatic), newEmojiID, "png")
	emojiStaticPath := fmt.Sprintf("%s/%s/%s/%s.png", instanceAccount.ID, TypeEmoji, SizeStatic, newEmojiID)

	// Store the original emoji
	if err := mh.storage.Put(emojiPath, original.image); err != nil {
		return nil, fmt.Errorf("storage error: %s", err)
	}

	// Store the static emoji
	if err := mh.storage.Put(emojiStaticPath, static.image); err != nil {
		return nil, fmt.Errorf("storage error: %s", err)
	}

	// and finally return the new emoji data to the caller -- it's up to them what to do with it
	e := &gtsmodel.Emoji{
		ID:                     newEmojiID,
		Shortcode:              shortcode,
		Domain:                 "", // empty because this is a local emoji
		CreatedAt:              time.Now(),
		UpdatedAt:              time.Now(),
		ImageRemoteURL:         "", // empty because this is a local emoji
		ImageStaticRemoteURL:   "", // empty because this is a local emoji
		ImageURL:               emojiURL,
		ImageStaticURL:         emojiStaticURL,
		ImagePath:              emojiPath,
		ImageStaticPath:        emojiStaticPath,
		ImageContentType:       contentType,
		ImageStaticContentType: MIMEPng, // static version will always be a png
		ImageFileSize:          len(original.image),
		ImageStaticFileSize:    len(static.image),
		ImageUpdatedAt:         time.Now(),
		Disabled:               false,
		URI:                    emojiURI,
		VisibleInPicker:        true,
		CategoryID:             "", // empty because this is a new emoji -- no category yet
	}
	return e, nil
}

func (mh *mediaHandler) ProcessRemoteHeaderOrAvatar(ctx context.Context, t transport.Transport, currentAttachment *gtsmodel.MediaAttachment, accountID string) (*gtsmodel.MediaAttachment, error) {
	if !currentAttachment.Header && !currentAttachment.Avatar {
		return nil, errors.New("provided attachment was set to neither header nor avatar")
	}

	if currentAttachment.Header && currentAttachment.Avatar {
		return nil, errors.New("provided attachment was set to both header and avatar")
	}

	var headerOrAvi Type
	if currentAttachment.Header {
		headerOrAvi = TypeHeader
	} else if currentAttachment.Avatar {
		headerOrAvi = TypeAvatar
	}

	if currentAttachment.RemoteURL == "" {
		return nil, errors.New("no remote URL on media attachment to dereference")
	}
	remoteIRI, err := url.Parse(currentAttachment.RemoteURL)
	if err != nil {
		return nil, fmt.Errorf("error parsing attachment url %s: %s", currentAttachment.RemoteURL, err)
	}

	// for content type, we assume we don't know what to expect...
	expectedContentType := "*/*"
	if currentAttachment.File.ContentType != "" {
		// ... and then narrow it down if we do
		expectedContentType = currentAttachment.File.ContentType
	}

	attachmentBytes, err := t.DereferenceMedia(ctx, remoteIRI, expectedContentType)
	if err != nil {
		return nil, fmt.Errorf("dereferencing remote media with url %s: %s", remoteIRI.String(), err)
	}

	return mh.ProcessHeaderOrAvatar(ctx, attachmentBytes, accountID, headerOrAvi, currentAttachment.RemoteURL)
}

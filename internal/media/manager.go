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
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"codeberg.org/gruf/go-store/kv"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/transport"
	"github.com/superseriousbusiness/gotosocial/internal/uris"
)

// ProcessCallback is triggered by the media manager when an attachment has finished undergoing
// image processing (generation of a blurhash, thumbnail etc) but hasn't yet been inserted into
// the database. It is provided to allow callers to a) access the processed media attachment and b)
// make any last-minute changes to the media attachment before it enters the database.
type ProcessCallback func(*gtsmodel.MediaAttachment) *gtsmodel.MediaAttachment

// defaultCB will be used when a nil ProcessCallback is passed to one of the manager's interface functions.
// It just returns the processed media attachment with no additional changes.
var defaultCB ProcessCallback = func(a *gtsmodel.MediaAttachment) *gtsmodel.MediaAttachment {
	return a
}

// Manager provides an interface for managing media: parsing, storing, and retrieving media objects like photos, videos, and gifs.
type Manager interface {
	ProcessAttachment(ctx context.Context, data []byte, accountID string, cb ProcessCallback) (*gtsmodel.MediaAttachment, error)
}

type manager struct {
	db      db.DB
	storage *kv.KVStore
}

// New returns a media manager with the given db and underlying storage.
func New(database db.DB, storage *kv.KVStore) Manager {
	return &manager{
		db:      database,
		storage: storage,
	}
}

/*
	INTERFACE FUNCTIONS
*/

func (m *manager) ProcessAttachment(ctx context.Context, data []byte, accountID string, cb ProcessCallback) (*gtsmodel.MediaAttachment, error) {
	contentType, err := parseContentType(data)
	if err != nil {
		return nil, err
	}

	mainType := strings.Split(contentType, "/")[0]
	switch mainType {
	case mimeImage:
		if !supportedImage(contentType) {
			return nil, fmt.Errorf("image type %s not supported", contentType)
		}
		if len(data) == 0 {
			return nil, errors.New("image was of size 0")
		}
		return m.processImage(attachmentBytes, minAttachment)
	default:
		return nil, fmt.Errorf("content type %s not (yet) supported", contentType)
	}
}

// ProcessHeaderOrAvatar takes a new header image for an account, checks it out, removes exif data from it,
// puts it in whatever storage backend we're using, sets the relevant fields in the database for the new image,
// and then returns information to the caller about the new header.
func (m *manager) ProcessHeader(ctx context.Context, data []byte, accountID string, cb ProcessCallback) (*gtsmodel.MediaAttachment, error) {

	// make sure we have a type we can handle
	contentType, err := parseContentType(data)
	if err != nil {
		return nil, err
	}

	if !supportedImage(contentType) {
		return nil, fmt.Errorf("%s is not an accepted image type", contentType)
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("passed reader was of size 0")
	}

	// process it
	ma, err := m.processHeaderOrAvi(attachment, contentType, mediaType, accountID, remoteURL)
	if err != nil {
		return nil, fmt.Errorf("error processing %s: %s", mediaType, err)
	}

	// set it in the database
	if err := m.db.SetAccountHeaderOrAvatar(ctx, ma, accountID); err != nil {
		return nil, fmt.Errorf("error putting %s in database: %s", mediaType, err)
	}

	return ma, nil
}

// ProcessLocalEmoji takes a new emoji and a shortcode, cleans it up, puts it in storage, and creates a new
// *gts.Emoji for it, then returns it to the caller. It's the caller's responsibility to put the returned struct
// in the database.
func (m *manager) ProcessLocalEmoji(ctx context.Context, emojiBytes []byte, shortcode string) (*gtsmodel.Emoji, error) {
	var clean []byte
	var err error
	var original *imageAndMeta
	var static *imageAndMeta

	// check content type of the submitted emoji and make sure it's supported by us
	contentType, err := parseContentType(emojiBytes)
	if err != nil {
		return nil, err
	}
	if !supportedEmoji(contentType) {
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
	case mimePng:
		if clean, err = purgeExif(emojiBytes); err != nil {
			return nil, fmt.Errorf("error cleaning exif data: %s", err)
		}
	case mimeGif:
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
	instanceAccount, err := m.db.GetInstanceAccount(ctx, "")
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
	if err := m.storage.Put(emojiPath, original.image); err != nil {
		return nil, fmt.Errorf("storage error: %s", err)
	}

	// Store the static emoji
	if err := m.storage.Put(emojiStaticPath, static.image); err != nil {
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
		ImageStaticContentType: mimePng, // static version will always be a png
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

func (m *manager) ProcessRemoteHeaderOrAvatar(ctx context.Context, t transport.Transport, currentAttachment *gtsmodel.MediaAttachment, accountID string) (*gtsmodel.MediaAttachment, error) {
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

	return m.ProcessHeaderOrAvatar(ctx, attachmentBytes, accountID, headerOrAvi, currentAttachment.RemoteURL)
}

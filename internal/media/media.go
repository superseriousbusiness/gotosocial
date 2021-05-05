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
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/storage"
)

// MediaSize describes the *size* of a piece of media
type MediaSize string

// MediaType describes the *type* of a piece of media
type MediaType string

const (
	// Small is the key for small/thumbnail versions of media
	Small MediaSize = "small"
	// Original is the key for original/fullsize versions of media and emoji
	Original MediaSize = "original"
	// Static is the key for static (non-animated) versions of emoji
	Static MediaSize = "static"

	// Attachment is the key for media attachments
	Attachment MediaType = "attachment"
	// Header is the key for profile header requests
	Header MediaType = "header"
	// Avatar is the key for profile avatar requests
	Avatar MediaType = "avatar"
	// Emoji is the key for emoji type requests
	Emoji MediaType = "emoji"

	// EmojiMaxBytes is the maximum permitted bytes of an emoji upload (50kb)
	EmojiMaxBytes = 51200
)

// Handler provides an interface for parsing, storing, and retrieving media objects like photos, videos, and gifs.
type Handler interface {
	// ProcessHeaderOrAvatar takes a new header image for an account, checks it out, removes exif data from it,
	// puts it in whatever storage backend we're using, sets the relevant fields in the database for the new image,
	// and then returns information to the caller about the new header.
	ProcessHeaderOrAvatar(img []byte, accountID string, mediaType MediaType) (*gtsmodel.MediaAttachment, error)

	// ProcessLocalAttachment takes a new attachment and the requesting account, checks it out, removes exif data from it,
	// puts it in whatever storage backend we're using, sets the relevant fields in the database for the new media,
	// and then returns information to the caller about the attachment.
	ProcessLocalAttachment(attachment []byte, accountID string) (*gtsmodel.MediaAttachment, error)

	// ProcessLocalEmoji takes a new emoji and a shortcode, cleans it up, puts it in storage, and creates a new
	// *gts.Emoji for it, then returns it to the caller. It's the caller's responsibility to put the returned struct
	// in the database.
	ProcessLocalEmoji(emojiBytes []byte, shortcode string) (*gtsmodel.Emoji, error)
}

type mediaHandler struct {
	config  *config.Config
	db      db.DB
	storage storage.Storage
	log     *logrus.Logger
}

// New returns a new handler with the given config, db, storage, and logger
func New(config *config.Config, database db.DB, storage storage.Storage, log *logrus.Logger) Handler {
	return &mediaHandler{
		config:  config,
		db:      database,
		storage: storage,
		log:     log,
	}
}

/*
	INTERFACE FUNCTIONS
*/

// ProcessHeaderOrAvatar takes a new header image for an account, checks it out, removes exif data from it,
// puts it in whatever storage backend we're using, sets the relevant fields in the database for the new image,
// and then returns information to the caller about the new header.
func (mh *mediaHandler) ProcessHeaderOrAvatar(attachment []byte, accountID string, mediaType MediaType) (*gtsmodel.MediaAttachment, error) {
	l := mh.log.WithField("func", "SetHeaderForAccountID")

	if mediaType != Header && mediaType != Avatar {
		return nil, errors.New("header or avatar not selected")
	}

	// make sure we have a type we can handle
	contentType, err := parseContentType(attachment)
	if err != nil {
		return nil, err
	}
	if !supportedImageType(contentType) {
		return nil, fmt.Errorf("%s is not an accepted image type", contentType)
	}

	if len(attachment) == 0 {
		return nil, fmt.Errorf("passed reader was of size 0")
	}
	l.Tracef("read %d bytes of file", len(attachment))

	// process it
	ma, err := mh.processHeaderOrAvi(attachment, contentType, mediaType, accountID)
	if err != nil {
		return nil, fmt.Errorf("error processing %s: %s", mediaType, err)
	}

	// set it in the database
	if err := mh.db.SetHeaderOrAvatarForAccountID(ma, accountID); err != nil {
		return nil, fmt.Errorf("error putting %s in database: %s", mediaType, err)
	}

	return ma, nil
}

// ProcessLocalAttachment takes a new attachment and the requesting account, checks it out, removes exif data from it,
// puts it in whatever storage backend we're using, sets the relevant fields in the database for the new media,
// and then returns information to the caller about the attachment.
func (mh *mediaHandler) ProcessLocalAttachment(attachment []byte, accountID string) (*gtsmodel.MediaAttachment, error) {
	contentType, err := parseContentType(attachment)
	if err != nil {
		return nil, err
	}
	mainType := strings.Split(contentType, "/")[0]
	switch mainType {
	case "video":
		if !supportedVideoType(contentType) {
			return nil, fmt.Errorf("video type %s not supported", contentType)
		}
		if len(attachment) == 0 {
			return nil, errors.New("video was of size 0")
		}
		if len(attachment) > mh.config.MediaConfig.MaxVideoSize {
			return nil, fmt.Errorf("video size %d bytes exceeded max video size of %d bytes", len(attachment), mh.config.MediaConfig.MaxVideoSize)
		}
		return mh.processVideoAttachment(attachment, accountID, contentType)
	case "image":
		if !supportedImageType(contentType) {
			return nil, fmt.Errorf("image type %s not supported", contentType)
		}
		if len(attachment) == 0 {
			return nil, errors.New("image was of size 0")
		}
		if len(attachment) > mh.config.MediaConfig.MaxImageSize {
			return nil, fmt.Errorf("image size %d bytes exceeded max image size of %d bytes", len(attachment), mh.config.MediaConfig.MaxImageSize)
		}
		return mh.processImageAttachment(attachment, accountID, contentType)
	default:
		break
	}
	return nil, fmt.Errorf("content type %s not (yet) supported", contentType)
}

// ProcessLocalEmoji takes a new emoji and a shortcode, cleans it up, puts it in storage, and creates a new
// *gts.Emoji for it, then returns it to the caller. It's the caller's responsibility to put the returned struct
// in the database.
func (mh *mediaHandler) ProcessLocalEmoji(emojiBytes []byte, shortcode string) (*gtsmodel.Emoji, error) {
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

	// clean any exif data from image/png type but leave gifs alone
	switch contentType {
	case "image/png":
		if clean, err = purgeExif(emojiBytes); err != nil {
			return nil, fmt.Errorf("error cleaning exif data: %s", err)
		}
	case "image/gif":
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
	instanceAccount := &gtsmodel.Account{}
	if err := mh.db.GetLocalAccountByUsername(mh.config.Host, instanceAccount); err != nil {
		return nil, fmt.Errorf("error fetching instance account: %s", err)
	}

	// the file extension (either png or gif)
	extension := strings.Split(contentType, "/")[1]

	// create the urls and storage paths
	URLbase := fmt.Sprintf("%s://%s%s", mh.config.StorageConfig.ServeProtocol, mh.config.StorageConfig.ServeHost, mh.config.StorageConfig.ServeBasePath)

	// generate a uuid for the new emoji -- normally we could let the database do this for us,
	// but we need it below so we should create it here instead.
	newEmojiID := uuid.NewString()

	// webfinger uri for the emoji -- unrelated to actually serving the image
	// will be something like https://example.org/emoji/70a7f3d7-7e35-4098-8ce3-9b5e8203bb9c
	emojiURI := fmt.Sprintf("%s://%s/%s/%s", mh.config.Protocol, mh.config.Host, Emoji, newEmojiID)

	// serve url and storage path for the original emoji -- can be png or gif
	emojiURL := fmt.Sprintf("%s/%s/%s/%s/%s.%s", URLbase, instanceAccount.ID, Emoji, Original, newEmojiID, extension)
	emojiPath := fmt.Sprintf("%s/%s/%s/%s/%s.%s", mh.config.StorageConfig.BasePath, instanceAccount.ID, Emoji, Original, newEmojiID, extension)

	// serve url and storage path for the static version -- will always be png
	emojiStaticURL := fmt.Sprintf("%s/%s/%s/%s/%s.png", URLbase, instanceAccount.ID, Emoji, Static, newEmojiID)
	emojiStaticPath := fmt.Sprintf("%s/%s/%s/%s/%s.png", mh.config.StorageConfig.BasePath, instanceAccount.ID, Emoji, Static, newEmojiID)

	// store the original
	if err := mh.storage.StoreFileAt(emojiPath, original.image); err != nil {
		return nil, fmt.Errorf("storage error: %s", err)
	}

	// store the static
	if err := mh.storage.StoreFileAt(emojiStaticPath, static.image); err != nil {
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
		ImageStaticContentType: "image/png", // static version will always be a png
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

/*
	HELPER FUNCTIONS
*/

func (mh *mediaHandler) processVideoAttachment(data []byte, accountID string, contentType string) (*gtsmodel.MediaAttachment, error) {
	return nil, nil
}

func (mh *mediaHandler) processImageAttachment(data []byte, accountID string, contentType string) (*gtsmodel.MediaAttachment, error) {
	var clean []byte
	var err error
	var original *imageAndMeta
	var small *imageAndMeta

	switch contentType {
	case "image/jpeg", "image/png":
		if clean, err = purgeExif(data); err != nil {
			return nil, fmt.Errorf("error cleaning exif data: %s", err)
		}
		original, err = deriveImage(clean, contentType)
		if err != nil {
			return nil, fmt.Errorf("error parsing image: %s", err)
		}
	case "image/gif":
		clean = data
		original, err = deriveGif(clean, contentType)
		if err != nil {
			return nil, fmt.Errorf("error parsing gif: %s", err)
		}
	default:
		return nil, errors.New("media type unrecognized")
	}

	small, err = deriveThumbnail(clean, contentType, 256, 256)
	if err != nil {
		return nil, fmt.Errorf("error deriving thumbnail: %s", err)
	}

	// now put it in storage, take a new uuid for the name of the file so we don't store any unnecessary info about it
	extension := strings.Split(contentType, "/")[1]
	newMediaID := uuid.NewString()

	URLbase := fmt.Sprintf("%s://%s%s", mh.config.StorageConfig.ServeProtocol, mh.config.StorageConfig.ServeHost, mh.config.StorageConfig.ServeBasePath)
	originalURL := fmt.Sprintf("%s/%s/attachment/original/%s.%s", URLbase, accountID, newMediaID, extension)
	smallURL := fmt.Sprintf("%s/%s/attachment/small/%s.jpeg", URLbase, accountID, newMediaID) // all thumbnails/smalls are encoded as jpeg

	// we store the original...
	originalPath := fmt.Sprintf("%s/%s/%s/%s/%s.%s", mh.config.StorageConfig.BasePath, accountID, Attachment, Original, newMediaID, extension)
	if err := mh.storage.StoreFileAt(originalPath, original.image); err != nil {
		return nil, fmt.Errorf("storage error: %s", err)
	}

	// and a thumbnail...
	smallPath := fmt.Sprintf("%s/%s/%s/%s/%s.jpeg", mh.config.StorageConfig.BasePath, accountID, Attachment, Small, newMediaID) // all thumbnails/smalls are encoded as jpeg
	if err := mh.storage.StoreFileAt(smallPath, small.image); err != nil {
		return nil, fmt.Errorf("storage error: %s", err)
	}

	ma := &gtsmodel.MediaAttachment{
		ID:        newMediaID,
		StatusID:  "",
		URL:       originalURL,
		RemoteURL: "",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Type:      gtsmodel.FileTypeImage,
		FileMeta: gtsmodel.FileMeta{
			Original: gtsmodel.Original{
				Width:  original.width,
				Height: original.height,
				Size:   original.size,
				Aspect: original.aspect,
			},
			Small: gtsmodel.Small{
				Width:  small.width,
				Height: small.height,
				Size:   small.size,
				Aspect: small.aspect,
			},
		},
		AccountID:         accountID,
		Description:       "",
		ScheduledStatusID: "",
		Blurhash:          original.blurhash,
		Processing:        2,
		File: gtsmodel.File{
			Path:        originalPath,
			ContentType: contentType,
			FileSize:    len(original.image),
			UpdatedAt:   time.Now(),
		},
		Thumbnail: gtsmodel.Thumbnail{
			Path:        smallPath,
			ContentType: "image/jpeg", // all thumbnails/smalls are encoded as jpeg
			FileSize:    len(small.image),
			UpdatedAt:   time.Now(),
			URL:         smallURL,
			RemoteURL:   "",
		},
		Avatar: false,
		Header: false,
	}

	return ma, nil

}

func (mh *mediaHandler) processHeaderOrAvi(imageBytes []byte, contentType string, mediaType MediaType, accountID string) (*gtsmodel.MediaAttachment, error) {
	var isHeader bool
	var isAvatar bool

	switch mediaType {
	case Header:
		isHeader = true
	case Avatar:
		isAvatar = true
	default:
		return nil, errors.New("header or avatar not selected")
	}

	var clean []byte
	var err error

	switch contentType {
	case "image/jpeg":
		if clean, err = purgeExif(imageBytes); err != nil {
			return nil, fmt.Errorf("error cleaning exif data: %s", err)
		}
	case "image/png":
		if clean, err = purgeExif(imageBytes); err != nil {
			return nil, fmt.Errorf("error cleaning exif data: %s", err)
		}
	case "image/gif":
		clean = imageBytes
	default:
		return nil, errors.New("media type unrecognized")
	}

	original, err := deriveImage(clean, contentType)
	if err != nil {
		return nil, fmt.Errorf("error parsing image: %s", err)
	}

	small, err := deriveThumbnail(clean, contentType, 256, 256)
	if err != nil {
		return nil, fmt.Errorf("error deriving thumbnail: %s", err)
	}

	// now put it in storage, take a new uuid for the name of the file so we don't store any unnecessary info about it
	extension := strings.Split(contentType, "/")[1]
	newMediaID := uuid.NewString()

	URLbase := fmt.Sprintf("%s://%s%s", mh.config.StorageConfig.ServeProtocol, mh.config.StorageConfig.ServeHost, mh.config.StorageConfig.ServeBasePath)
	originalURL := fmt.Sprintf("%s/%s/%s/original/%s.%s", URLbase, accountID, mediaType, newMediaID, extension)
	smallURL := fmt.Sprintf("%s/%s/%s/small/%s.%s", URLbase, accountID, mediaType, newMediaID, extension)

	// we store the original...
	originalPath := fmt.Sprintf("%s/%s/%s/%s/%s.%s", mh.config.StorageConfig.BasePath, accountID, mediaType, Original, newMediaID, extension)
	if err := mh.storage.StoreFileAt(originalPath, original.image); err != nil {
		return nil, fmt.Errorf("storage error: %s", err)
	}

	// and a thumbnail...
	smallPath := fmt.Sprintf("%s/%s/%s/%s/%s.%s", mh.config.StorageConfig.BasePath, accountID, mediaType, Small, newMediaID, extension)
	if err := mh.storage.StoreFileAt(smallPath, small.image); err != nil {
		return nil, fmt.Errorf("storage error: %s", err)
	}

	ma := &gtsmodel.MediaAttachment{
		ID:        newMediaID,
		StatusID:  "",
		URL:       originalURL,
		RemoteURL: "",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Type:      gtsmodel.FileTypeImage,
		FileMeta: gtsmodel.FileMeta{
			Original: gtsmodel.Original{
				Width:  original.width,
				Height: original.height,
				Size:   original.size,
				Aspect: original.aspect,
			},
			Small: gtsmodel.Small{
				Width:  small.width,
				Height: small.height,
				Size:   small.size,
				Aspect: small.aspect,
			},
		},
		AccountID:         accountID,
		Description:       "",
		ScheduledStatusID: "",
		Blurhash:          original.blurhash,
		Processing:        2,
		File: gtsmodel.File{
			Path:        originalPath,
			ContentType: contentType,
			FileSize:    len(original.image),
			UpdatedAt:   time.Now(),
		},
		Thumbnail: gtsmodel.Thumbnail{
			Path:        smallPath,
			ContentType: contentType,
			FileSize:    len(small.image),
			UpdatedAt:   time.Now(),
			URL:         smallURL,
			RemoteURL:   "",
		},
		Avatar: isAvatar,
		Header: isHeader,
	}

	return ma, nil
}

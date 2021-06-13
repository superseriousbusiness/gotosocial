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

package processing

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/transport"
)

/*
	HELPER FUNCTIONS
*/

// TODO: try to combine the below two functions because this is a lot of code repetition.

// updateAccountAvatar does the dirty work of checking the avatar part of an account update form,
// parsing and checking the image, and doing the necessary updates in the database for this to become
// the account's new avatar image.
func (p *processor) updateAccountAvatar(avatar *multipart.FileHeader, accountID string) (*gtsmodel.MediaAttachment, error) {
	var err error
	if int(avatar.Size) > p.config.MediaConfig.MaxImageSize {
		err = fmt.Errorf("avatar with size %d exceeded max image size of %d bytes", avatar.Size, p.config.MediaConfig.MaxImageSize)
		return nil, err
	}
	f, err := avatar.Open()
	if err != nil {
		return nil, fmt.Errorf("could not read provided avatar: %s", err)
	}

	// extract the bytes
	buf := new(bytes.Buffer)
	size, err := io.Copy(buf, f)
	if err != nil {
		return nil, fmt.Errorf("could not read provided avatar: %s", err)
	}
	if size == 0 {
		return nil, errors.New("could not read provided avatar: size 0 bytes")
	}

	// do the setting
	avatarInfo, err := p.mediaHandler.ProcessHeaderOrAvatar(buf.Bytes(), accountID, media.Avatar, "")
	if err != nil {
		return nil, fmt.Errorf("error processing avatar: %s", err)
	}

	return avatarInfo, f.Close()
}

// updateAccountHeader does the dirty work of checking the header part of an account update form,
// parsing and checking the image, and doing the necessary updates in the database for this to become
// the account's new header image.
func (p *processor) updateAccountHeader(header *multipart.FileHeader, accountID string) (*gtsmodel.MediaAttachment, error) {
	var err error
	if int(header.Size) > p.config.MediaConfig.MaxImageSize {
		err = fmt.Errorf("header with size %d exceeded max image size of %d bytes", header.Size, p.config.MediaConfig.MaxImageSize)
		return nil, err
	}
	f, err := header.Open()
	if err != nil {
		return nil, fmt.Errorf("could not read provided header: %s", err)
	}

	// extract the bytes
	buf := new(bytes.Buffer)
	size, err := io.Copy(buf, f)
	if err != nil {
		return nil, fmt.Errorf("could not read provided header: %s", err)
	}
	if size == 0 {
		return nil, errors.New("could not read provided header: size 0 bytes")
	}

	// do the setting
	headerInfo, err := p.mediaHandler.ProcessHeaderOrAvatar(buf.Bytes(), accountID, media.Header, "")
	if err != nil {
		return nil, fmt.Errorf("error processing header: %s", err)
	}

	return headerInfo, f.Close()
}

// fetchHeaderAndAviForAccount fetches the header and avatar for a remote account, using a transport
// on behalf of requestingUsername.
//
// targetAccount's AvatarMediaAttachmentID and HeaderMediaAttachmentID will be updated as necessary.
//
// SIDE EFFECTS: remote header and avatar will be stored in local storage, and the database will be updated
// to reflect the creation of these new attachments.
func (p *processor) fetchHeaderAndAviForAccount(targetAccount *gtsmodel.Account, t transport.Transport, refresh bool) error {
	if targetAccount.AvatarRemoteURL != "" && (targetAccount.AvatarMediaAttachmentID == "" || refresh) {
		a, err := p.mediaHandler.ProcessRemoteHeaderOrAvatar(t, &gtsmodel.MediaAttachment{
			RemoteURL: targetAccount.AvatarRemoteURL,
			Avatar:    true,
		}, targetAccount.ID)
		if err != nil {
			return fmt.Errorf("error processing avatar for user: %s", err)
		}
		targetAccount.AvatarMediaAttachmentID = a.ID
	}

	if targetAccount.HeaderRemoteURL != "" && (targetAccount.HeaderMediaAttachmentID == "" || refresh) {
		a, err := p.mediaHandler.ProcessRemoteHeaderOrAvatar(t, &gtsmodel.MediaAttachment{
			RemoteURL: targetAccount.HeaderRemoteURL,
			Header:    true,
		}, targetAccount.ID)
		if err != nil {
			return fmt.Errorf("error processing header for user: %s", err)
		}
		targetAccount.HeaderMediaAttachmentID = a.ID
	}
	return nil
}

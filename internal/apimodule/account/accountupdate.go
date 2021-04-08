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

package account

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/db/gtsmodel"
	mastotypes "github.com/superseriousbusiness/gotosocial/internal/mastotypes/mastomodel"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// accountUpdateCredentialsPATCHHandler allows a user to modify their account/profile settings.
// It should be served as a PATCH at /api/v1/accounts/update_credentials
//
// TODO: this can be optimized massively by building up a picture of what we want the new account
// details to be, and then inserting it all in the database at once. As it is, we do queries one-by-one
// which is not gonna make the database very happy when lots of requests are going through.
// This way it would also be safer because the update won't happen until *all* the fields are validated.
// Otherwise we risk doing a partial update and that's gonna cause probllleeemmmsss.
func (m *accountModule) accountUpdateCredentialsPATCHHandler(c *gin.Context) {
	l := m.log.WithField("func", "accountUpdateCredentialsPATCHHandler")
	authed, err := oauth.MustAuth(c, true, false, false, true)
	if err != nil {
		l.Debugf("couldn't auth: %s", err)
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	l.Tracef("retrieved account %+v", authed.Account.ID)

	l.Trace("parsing request form")
	form := &mastotypes.UpdateCredentialsRequest{}
	if err := c.ShouldBind(form); err != nil || form == nil {
		l.Debugf("could not parse form from request: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// if everything on the form is nil, then nothing has been set and we shouldn't continue
	if form.Discoverable == nil && form.Bot == nil && form.DisplayName == nil && form.Note == nil && form.Avatar == nil && form.Header == nil && form.Locked == nil && form.Source == nil && form.FieldsAttributes == nil {
		l.Debugf("could not parse form from request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty form submitted"})
		return
	}

	if form.Discoverable != nil {
		if err := m.db.UpdateOneByID(authed.Account.ID, "discoverable", *form.Discoverable, &gtsmodel.Account{}); err != nil {
			l.Debugf("error updating discoverable: %s", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	if form.Bot != nil {
		if err := m.db.UpdateOneByID(authed.Account.ID, "bot", *form.Bot, &gtsmodel.Account{}); err != nil {
			l.Debugf("error updating bot: %s", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	if form.DisplayName != nil {
		if err := util.ValidateDisplayName(*form.DisplayName); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := m.db.UpdateOneByID(authed.Account.ID, "display_name", *form.DisplayName, &gtsmodel.Account{}); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	if form.Note != nil {
		if err := util.ValidateNote(*form.Note); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := m.db.UpdateOneByID(authed.Account.ID, "note", *form.Note, &gtsmodel.Account{}); err != nil {
			l.Debugf("error updating note: %s", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	if form.Avatar != nil && form.Avatar.Size != 0 {
		avatarInfo, err := m.UpdateAccountAvatar(form.Avatar, authed.Account.ID)
		if err != nil {
			l.Debugf("could not update avatar for account %s: %s", authed.Account.ID, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		l.Tracef("new avatar info for account %s is %+v", authed.Account.ID, avatarInfo)
	}

	if form.Header != nil && form.Header.Size != 0 {
		headerInfo, err := m.UpdateAccountHeader(form.Header, authed.Account.ID)
		if err != nil {
			l.Debugf("could not update header for account %s: %s", authed.Account.ID, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		l.Tracef("new header info for account %s is %+v", authed.Account.ID, headerInfo)
	}

	if form.Locked != nil {
		if err := m.db.UpdateOneByID(authed.Account.ID, "locked", *form.Locked, &gtsmodel.Account{}); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	if form.Source != nil {
		if form.Source.Language != nil {
			if err := util.ValidateLanguage(*form.Source.Language); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			if err := m.db.UpdateOneByID(authed.Account.ID, "language", *form.Source.Language, &gtsmodel.Account{}); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}

		if form.Source.Sensitive != nil {
			if err := m.db.UpdateOneByID(authed.Account.ID, "locked", *form.Locked, &gtsmodel.Account{}); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}

		if form.Source.Privacy != nil {
			if err := util.ValidatePrivacy(*form.Source.Privacy); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			if err := m.db.UpdateOneByID(authed.Account.ID, "privacy", *form.Source.Privacy, &gtsmodel.Account{}); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}
	}

	// if form.FieldsAttributes != nil {
	// 	// TODO: parse fields attributes nicely and update
	// }

	// fetch the account with all updated values set
	updatedAccount := &gtsmodel.Account{}
	if err := m.db.GetByID(authed.Account.ID, updatedAccount); err != nil {
		l.Debugf("could not fetch updated account %s: %s", authed.Account.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	acctSensitive, err := m.mastoConverter.AccountToMastoSensitive(updatedAccount)
	if err != nil {
		l.Tracef("could not convert account into mastosensitive account: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	l.Tracef("conversion successful, returning OK and mastosensitive account %+v", acctSensitive)
	c.JSON(http.StatusOK, acctSensitive)
}

/*
	HELPER FUNCTIONS
*/

// TODO: try to combine the below two functions because this is a lot of code repetition.

// UpdateAccountAvatar does the dirty work of checking the avatar part of an account update form,
// parsing and checking the image, and doing the necessary updates in the database for this to become
// the account's new avatar image.
func (m *accountModule) UpdateAccountAvatar(avatar *multipart.FileHeader, accountID string) (*gtsmodel.MediaAttachment, error) {
	var err error
	if int(avatar.Size) > m.config.MediaConfig.MaxImageSize {
		err = fmt.Errorf("avatar with size %d exceeded max image size of %d bytes", avatar.Size, m.config.MediaConfig.MaxImageSize)
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
	avatarInfo, err := m.mediaHandler.SetHeaderOrAvatarForAccountID(buf.Bytes(), accountID, "avatar")
	if err != nil {
		return nil, fmt.Errorf("error processing avatar: %s", err)
	}

	return avatarInfo, f.Close()
}

// UpdateAccountHeader does the dirty work of checking the header part of an account update form,
// parsing and checking the image, and doing the necessary updates in the database for this to become
// the account's new header image.
func (m *accountModule) UpdateAccountHeader(header *multipart.FileHeader, accountID string) (*gtsmodel.MediaAttachment, error) {
	var err error
	if int(header.Size) > m.config.MediaConfig.MaxImageSize {
		err = fmt.Errorf("header with size %d exceeded max image size of %d bytes", header.Size, m.config.MediaConfig.MaxImageSize)
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
	headerInfo, err := m.mediaHandler.SetHeaderOrAvatarForAccountID(buf.Bytes(), accountID, "header")
	if err != nil {
		return nil, fmt.Errorf("error processing header: %s", err)
	}

	return headerInfo, f.Close()
}

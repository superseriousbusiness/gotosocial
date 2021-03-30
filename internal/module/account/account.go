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
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/db/model"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/module"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/router"
	"github.com/superseriousbusiness/gotosocial/internal/util"
	"github.com/superseriousbusiness/gotosocial/pkg/mastotypes"
	"github.com/superseriousbusiness/oauth2/v4"
)

const (
	idKey                 = "id"
	basePath              = "/api/v1/accounts"
	basePathWithID        = basePath + "/:" + idKey
	verifyPath            = basePath + "/verify_credentials"
	updateCredentialsPath = basePath + "/update_credentials"
)

type accountModule struct {
	config       *config.Config
	db           db.DB
	oauthServer  oauth.Server
	mediaHandler media.MediaHandler
	log          *logrus.Logger
}

// New returns a new account module
func New(config *config.Config, db db.DB, oauthServer oauth.Server, mediaHandler media.MediaHandler, log *logrus.Logger) module.ClientAPIModule {
	return &accountModule{
		config:       config,
		db:           db,
		oauthServer:  oauthServer,
		mediaHandler: mediaHandler,
		log:          log,
	}
}

// Route attaches all routes from this module to the given router
func (m *accountModule) Route(r router.Router) error {
	r.AttachHandler(http.MethodPost, basePath, m.accountCreatePOSTHandler)
	r.AttachHandler(http.MethodGet, verifyPath, m.accountVerifyGETHandler)
	r.AttachHandler(http.MethodPatch, updateCredentialsPath, m.accountUpdateCredentialsPATCHHandler)
	return nil
}

// accountCreatePOSTHandler handles create account requests, validates them,
// and puts them in the database if they're valid.
// It should be served as a POST at /api/v1/accounts
func (m *accountModule) accountCreatePOSTHandler(c *gin.Context) {
	l := m.log.WithField("func", "accountCreatePOSTHandler")
	authed, err := oauth.MustAuth(c, true, true, false, false)
	if err != nil {
		l.Debugf("couldn't auth: %s", err)
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	l.Trace("parsing request form")
	form := &mastotypes.AccountCreateRequest{}
	if err := c.ShouldBind(form); err != nil || form == nil {
		l.Debugf("could not parse form from request: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing one or more required form values"})
		return
	}

	l.Tracef("validating form %+v", form)
	if err := validateCreateAccount(form, m.config.AccountsConfig, m.db); err != nil {
		l.Debugf("error validating form: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	clientIP := c.ClientIP()
	l.Tracef("attempting to parse client ip address %s", clientIP)
	signUpIP := net.ParseIP(clientIP)
	if signUpIP == nil {
		l.Debugf("error validating sign up ip address %s", clientIP)
		c.JSON(http.StatusBadRequest, gin.H{"error": "ip address could not be parsed from request"})
		return
	}

	ti, err := m.accountCreate(form, signUpIP, authed.Token, authed.Application)
	if err != nil {
		l.Errorf("internal server error while creating new account: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ti)
}

// accountVerifyGETHandler serves a user's account details to them IF they reached this
// handler while in possession of a valid token, according to the oauth middleware.
// It should be served as a GET at /api/v1/accounts/verify_credentials
func (m *accountModule) accountVerifyGETHandler(c *gin.Context) {
	l := m.log.WithField("func", "accountVerifyGETHandler")
	authed, err := oauth.MustAuth(c, true, false, false, true)
	if err != nil {
		l.Debugf("couldn't auth: %s", err)
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	l.Tracef("retrieved account %+v, converting to mastosensitive...", authed.Account.ID)
	acctSensitive, err := m.db.AccountToMastoSensitive(authed.Account)
	if err != nil {
		l.Tracef("could not convert account into mastosensitive account: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	l.Tracef("conversion successful, returning OK and mastosensitive account %+v", acctSensitive)
	c.JSON(http.StatusOK, acctSensitive)
}

// accountUpdateCredentialsPATCHHandler allows a user to modify their account/profile settings.
// It should be served as a PATCH at /api/v1/accounts/update_credentials
//
// TODO: this can be optimized massively by building up a picture of what we want the new account
// details to be, and then inserting it all in the database at once. As it is, we do queries one-by-one
// which is not gonna make the database very happy when lots of requests are going through.
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
		if err := m.db.UpdateOneByID(authed.Account.ID, "discoverable", *form.Discoverable, &model.Account{}); err != nil {
			l.Debugf("error updating discoverable: %s", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	if form.Bot != nil {
		if err := m.db.UpdateOneByID(authed.Account.ID, "bot", *form.Bot, &model.Account{}); err != nil {
			l.Debugf("error updating bot: %s", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	if form.DisplayName != nil {
		if err := m.db.UpdateOneByID(authed.Account.ID, "display_name", *form.DisplayName, &model.Account{}); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	if form.Note != nil {
		if err := m.db.UpdateOneByID(authed.Account.ID, "note", *form.Note, &model.Account{}); err != nil {
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
		if err := m.db.UpdateOneByID(authed.Account.ID, "locked", *form.Locked, &model.Account{}); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"": err.Error()})
			return
		}
	}

	if form.Source != nil {

	}

	if form.FieldsAttributes != nil {

	}

	// fetch the account with all updated values set
	updatedAccount := &model.Account{}
	if err := m.db.GetByID(authed.Account.ID, updatedAccount); err != nil {
		l.Debugf("could not fetch updated account %s: %s", authed.Account.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	acctSensitive, err := m.db.AccountToMastoSensitive(updatedAccount)
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
func (m *accountModule) UpdateAccountAvatar(avatar *multipart.FileHeader, accountID string) (*model.MediaAttachment, error) {
	var err error
	if avatar.Size > m.config.MediaConfig.MaxImageSize {
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
func (m *accountModule) UpdateAccountHeader(header *multipart.FileHeader, accountID string) (*model.MediaAttachment, error) {
	var err error
	if header.Size > m.config.MediaConfig.MaxImageSize {
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

// accountCreate does the dirty work of making an account and user in the database.
// It then returns a token to the caller, for use with the new account, as per the
// spec here: https://docs.joinmastodon.org/methods/accounts/
func (m *accountModule) accountCreate(form *mastotypes.AccountCreateRequest, signUpIP net.IP, token oauth2.TokenInfo, app *model.Application) (*mastotypes.Token, error) {
	l := m.log.WithField("func", "accountCreate")

	// don't store a reason if we don't require one
	reason := form.Reason
	if !m.config.AccountsConfig.ReasonRequired {
		reason = ""
	}

	l.Trace("creating new username and account")
	user, err := m.db.NewSignup(form.Username, reason, m.config.AccountsConfig.RequireApproval, form.Email, form.Password, signUpIP, form.Locale, app.ID)
	if err != nil {
		return nil, fmt.Errorf("error creating new signup in the database: %s", err)
	}

	l.Tracef("generating a token for user %s with account %s and application %s", user.ID, user.AccountID, app.ID)
	ti, err := m.oauthServer.GenerateUserAccessToken(token, app.ClientSecret, user.ID)
	if err != nil {
		return nil, fmt.Errorf("error creating new access token for user %s: %s", user.ID, err)
	}

	return &mastotypes.Token{
		AccessToken: ti.GetCode(),
		TokenType:   "Bearer",
		Scope:       ti.GetScope(),
		CreatedAt:   ti.GetCodeCreateAt().Unix(),
	}, nil
}

// validateCreateAccount checks through all the necessary prerequisites for creating a new account,
// according to the provided account create request. If the account isn't eligible, an error will be returned.
func validateCreateAccount(form *mastotypes.AccountCreateRequest, c *config.AccountsConfig, database db.DB) error {
	if !c.OpenRegistration {
		return errors.New("registration is not open for this server")
	}

	if err := util.ValidateSignUpUsername(form.Username); err != nil {
		return err
	}

	if err := util.ValidateEmail(form.Email); err != nil {
		return err
	}

	if err := util.ValidateSignUpPassword(form.Password); err != nil {
		return err
	}

	if !form.Agreement {
		return errors.New("agreement to terms and conditions not given")
	}

	if err := util.ValidateLanguage(form.Locale); err != nil {
		return err
	}

	if err := util.ValidateSignUpReason(form.Reason, c.ReasonRequired); err != nil {
		return err
	}

	if err := database.IsEmailAvailable(form.Email); err != nil {
		return err
	}

	if err := database.IsUsernameAvailable(form.Username); err != nil {
		return err
	}

	return nil
}

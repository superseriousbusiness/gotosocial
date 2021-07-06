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
	"mime/multipart"

	"github.com/sirupsen/logrus"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/federation"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/internal/visibility"
	"github.com/superseriousbusiness/oauth2/v4"
)

// Processor wraps a bunch of functions for processing account actions.
type Processor interface {
	// Create processes the given form for creating a new account, returning an oauth token for that account if successful.
	Create(applicationToken oauth2.TokenInfo, application *gtsmodel.Application, form *apimodel.AccountCreateRequest) (*apimodel.Token, error)
	// Delete deletes an account, and all of that account's statuses, media, follows, notifications, etc etc etc.
	// The origin passed here should be either the ID of the account doing the delete (can be itself), or the ID of a domain block.
	Delete(account *gtsmodel.Account, origin string) error
	// Get processes the given request for account information.
	Get(requestingAccount *gtsmodel.Account, targetAccountID string) (*apimodel.Account, error)
	// Update processes the update of an account with the given form
	Update(account *gtsmodel.Account, form *apimodel.UpdateCredentialsRequest) (*apimodel.Account, error)
	// StatusesGet fetches a number of statuses (in time descending order) from the given account, filtered by visibility for
	// the account given in authed.
	StatusesGet(requestingAccount *gtsmodel.Account, targetAccountID string, limit int, excludeReplies bool, maxID string, pinned bool, mediaOnly bool) ([]apimodel.Status, gtserror.WithCode)
	// FollowersGet fetches a list of the target account's followers.
	FollowersGet(requestingAccount *gtsmodel.Account, targetAccountID string) ([]apimodel.Account, gtserror.WithCode)
	// FollowingGet fetches a list of the accounts that target account is following.
	FollowingGet(requestingAccount *gtsmodel.Account, targetAccountID string) ([]apimodel.Account, gtserror.WithCode)
	// RelationshipGet returns a relationship model describing the relationship of the targetAccount to the Authed account.
	RelationshipGet(requestingAccount *gtsmodel.Account, targetAccountID string) (*apimodel.Relationship, gtserror.WithCode)
	// FollowCreate handles a follow request to an account, either remote or local.
	FollowCreate(requestingAccount *gtsmodel.Account, form *apimodel.AccountFollowRequest) (*apimodel.Relationship, gtserror.WithCode)
	// FollowRemove handles the removal of a follow/follow request to an account, either remote or local.
	FollowRemove(requestingAccount *gtsmodel.Account, targetAccountID string) (*apimodel.Relationship, gtserror.WithCode)
	// UpdateHeader does the dirty work of checking the header part of an account update form,
	// parsing and checking the image, and doing the necessary updates in the database for this to become
	// the account's new header image.
	UpdateAvatar(avatar *multipart.FileHeader, accountID string) (*gtsmodel.MediaAttachment, error)
	// UpdateAvatar does the dirty work of checking the avatar part of an account update form,
	// parsing and checking the image, and doing the necessary updates in the database for this to become
	// the account's new avatar image.
	UpdateHeader(header *multipart.FileHeader, accountID string) (*gtsmodel.MediaAttachment, error)
}

type processor struct {
	tc            typeutils.TypeConverter
	config        *config.Config
	mediaHandler  media.Handler
	fromClientAPI chan gtsmodel.FromClientAPI
	oauthServer   oauth.Server
	filter        visibility.Filter
	db            db.DB
	federator     federation.Federator
	log           *logrus.Logger
}

// New returns a new account processor.
func New(db db.DB, tc typeutils.TypeConverter, mediaHandler media.Handler, oauthServer oauth.Server, fromClientAPI chan gtsmodel.FromClientAPI, federator federation.Federator, config *config.Config, log *logrus.Logger) Processor {
	return &processor{
		tc:            tc,
		config:        config,
		mediaHandler:  mediaHandler,
		fromClientAPI: fromClientAPI,
		oauthServer:   oauthServer,
		filter:        visibility.NewFilter(db, log),
		db:            db,
		federator:     federator,
		log:           log,
	}
}

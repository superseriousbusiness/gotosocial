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

package account

import (
	"context"
	"mime/multipart"
	"time"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/concurrency"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/federation"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/text"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/internal/visibility"
	"github.com/superseriousbusiness/oauth2/v4"
)

// Processor wraps a bunch of functions for processing account actions.
type Processor interface {
	// Create processes the given form for creating a new account, returning an oauth token for that account if successful.
	Create(ctx context.Context, applicationToken oauth2.TokenInfo, application *gtsmodel.Application, form *apimodel.AccountCreateRequest) (*apimodel.Token, gtserror.WithCode)
	// Delete deletes an account, and all of that account's statuses, media, follows, notifications, etc etc etc.
	// The origin passed here should be either the ID of the account doing the delete (can be itself), or the ID of a domain block.
	Delete(ctx context.Context, account *gtsmodel.Account, origin string) gtserror.WithCode
	// DeleteLocal is like delete, but specifically for deletion of local accounts rather than federated ones.
	// Unlike Delete, it will propagate the deletion out across the federating API to other instances.
	DeleteLocal(ctx context.Context, account *gtsmodel.Account, form *apimodel.AccountDeleteRequest) gtserror.WithCode
	// Get processes the given request for account information.
	Get(ctx context.Context, requestingAccount *gtsmodel.Account, targetAccountID string) (*apimodel.Account, gtserror.WithCode)
	// GetLocalByUsername processes the given request for account information targeting a local account by username.
	GetLocalByUsername(ctx context.Context, requestingAccount *gtsmodel.Account, username string) (*apimodel.Account, gtserror.WithCode)
	// GetCustomCSSForUsername returns custom css for the given local username.
	GetCustomCSSForUsername(ctx context.Context, username string) (string, gtserror.WithCode)
	// GetRSSFeedForUsername returns RSS feed for the given local username.
	GetRSSFeedForUsername(ctx context.Context, username string) (func() (string, gtserror.WithCode), time.Time, gtserror.WithCode)
	// Update processes the update of an account with the given form
	Update(ctx context.Context, account *gtsmodel.Account, form *apimodel.UpdateCredentialsRequest) (*apimodel.Account, gtserror.WithCode)
	// StatusesGet fetches a number of statuses (in time descending order) from the given account, filtered by visibility for
	// the account given in authed.
	StatusesGet(ctx context.Context, requestingAccount *gtsmodel.Account, targetAccountID string, limit int, excludeReplies bool, excludeReblogs bool, maxID string, minID string, pinned bool, mediaOnly bool, publicOnly bool) (*apimodel.PageableResponse, gtserror.WithCode)
	// WebStatusesGet fetches a number of statuses (in descending order) from the given account. It selects only
	// statuses which are suitable for showing on the public web profile of an account.
	WebStatusesGet(ctx context.Context, targetAccountID string, maxID string) (*apimodel.PageableResponse, gtserror.WithCode)
	// StatusesGet fetches a number of statuses (in time descending order) from the given account, filtered by visibility for
	// the account given in authed.
	BookmarksGet(ctx context.Context, requestingAccount *gtsmodel.Account, limit int, maxID string, minID string) (*apimodel.PageableResponse, gtserror.WithCode)
	// FollowersGet fetches a list of the target account's followers.
	FollowersGet(ctx context.Context, requestingAccount *gtsmodel.Account, targetAccountID string) ([]apimodel.Account, gtserror.WithCode)
	// FollowingGet fetches a list of the accounts that target account is following.
	FollowingGet(ctx context.Context, requestingAccount *gtsmodel.Account, targetAccountID string) ([]apimodel.Account, gtserror.WithCode)
	// RelationshipGet returns a relationship model describing the relationship of the targetAccount to the Authed account.
	RelationshipGet(ctx context.Context, requestingAccount *gtsmodel.Account, targetAccountID string) (*apimodel.Relationship, gtserror.WithCode)
	// FollowCreate handles a follow request to an account, either remote or local.
	FollowCreate(ctx context.Context, requestingAccount *gtsmodel.Account, form *apimodel.AccountFollowRequest) (*apimodel.Relationship, gtserror.WithCode)
	// FollowRemove handles the removal of a follow/follow request to an account, either remote or local.
	FollowRemove(ctx context.Context, requestingAccount *gtsmodel.Account, targetAccountID string) (*apimodel.Relationship, gtserror.WithCode)
	// BlockCreate handles the creation of a block from requestingAccount to targetAccountID, either remote or local.
	BlockCreate(ctx context.Context, requestingAccount *gtsmodel.Account, targetAccountID string) (*apimodel.Relationship, gtserror.WithCode)
	// BlockRemove handles the removal of a block from requestingAccount to targetAccountID, either remote or local.
	BlockRemove(ctx context.Context, requestingAccount *gtsmodel.Account, targetAccountID string) (*apimodel.Relationship, gtserror.WithCode)
	// UpdateAvatar does the dirty work of checking the avatar part of an account update form,
	// parsing and checking the image, and doing the necessary updates in the database for this to become
	// the account's new avatar image.
	UpdateAvatar(ctx context.Context, avatar *multipart.FileHeader, description *string, accountID string) (*gtsmodel.MediaAttachment, error)
	// UpdateHeader does the dirty work of checking the header part of an account update form,
	// parsing and checking the image, and doing the necessary updates in the database for this to become
	// the account's new header image.
	UpdateHeader(ctx context.Context, header *multipart.FileHeader, description *string, accountID string) (*gtsmodel.MediaAttachment, error)
}

type processor struct {
	tc           typeutils.TypeConverter
	mediaManager media.Manager
	clientWorker *concurrency.WorkerPool[messages.FromClientAPI]
	oauthServer  oauth.Server
	filter       visibility.Filter
	formatter    text.Formatter
	db           db.DB
	federator    federation.Federator
	parseMention gtsmodel.ParseMentionFunc
}

// New returns a new account processor.
func New(db db.DB, tc typeutils.TypeConverter, mediaManager media.Manager, oauthServer oauth.Server, clientWorker *concurrency.WorkerPool[messages.FromClientAPI], federator federation.Federator, parseMention gtsmodel.ParseMentionFunc) Processor {
	return &processor{
		tc:           tc,
		mediaManager: mediaManager,
		clientWorker: clientWorker,
		oauthServer:  oauthServer,
		filter:       visibility.NewFilter(db),
		formatter:    text.NewFormatter(db),
		db:           db,
		federator:    federator,
		parseMention: parseMention,
	}
}

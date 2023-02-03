/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

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

package status

import (
	"context"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/concurrency"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	"github.com/superseriousbusiness/gotosocial/internal/text"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/internal/visibility"
)

// Processor wraps a bunch of functions for processing statuses.
type Processor interface {
	// Create processes the given form to create a new status, returning the api model representation of that status if it's OK.
	Create(ctx context.Context, account *gtsmodel.Account, application *gtsmodel.Application, form *apimodel.AdvancedStatusCreateForm) (*apimodel.Status, gtserror.WithCode)
	// Delete processes the delete of a given status, returning the deleted status if the delete goes through.
	Delete(ctx context.Context, account *gtsmodel.Account, targetStatusID string) (*apimodel.Status, gtserror.WithCode)
	// Fave processes the faving of a given status, returning the updated status if the fave goes through.
	Fave(ctx context.Context, account *gtsmodel.Account, targetStatusID string) (*apimodel.Status, gtserror.WithCode)
	// Boost processes the boost/reblog of a given status, returning the newly-created boost if all is well.
	Boost(ctx context.Context, account *gtsmodel.Account, application *gtsmodel.Application, targetStatusID string) (*apimodel.Status, gtserror.WithCode)
	// Unboost processes the unboost/unreblog of a given status, returning the status if all is well.
	Unboost(ctx context.Context, account *gtsmodel.Account, application *gtsmodel.Application, targetStatusID string) (*apimodel.Status, gtserror.WithCode)
	// BoostedBy returns a slice of accounts that have boosted the given status, filtered according to privacy settings.
	BoostedBy(ctx context.Context, account *gtsmodel.Account, targetStatusID string) ([]*apimodel.Account, gtserror.WithCode)
	// FavedBy returns a slice of accounts that have liked the given status, filtered according to privacy settings.
	FavedBy(ctx context.Context, account *gtsmodel.Account, targetStatusID string) ([]*apimodel.Account, gtserror.WithCode)
	// Get gets the given status, taking account of privacy settings and blocks etc.
	Get(ctx context.Context, account *gtsmodel.Account, targetStatusID string) (*apimodel.Status, gtserror.WithCode)
	// Unfave processes the unfaving of a given status, returning the updated status if the fave goes through.
	Unfave(ctx context.Context, account *gtsmodel.Account, targetStatusID string) (*apimodel.Status, gtserror.WithCode)
	// Context returns the context (previous and following posts) from the given status ID
	Context(ctx context.Context, account *gtsmodel.Account, targetStatusID string) (*apimodel.Context, gtserror.WithCode)
	// Bookmarks a status
	Bookmark(ctx context.Context, account *gtsmodel.Account, targetStatusID string) (*apimodel.Status, gtserror.WithCode)
	// Removes a bookmark for a status
	Unbookmark(ctx context.Context, account *gtsmodel.Account, targetStatusID string) (*apimodel.Status, gtserror.WithCode)

	/*
		PROCESSING UTILS
	*/

	ProcessVisibility(ctx context.Context, form *apimodel.AdvancedStatusCreateForm, accountDefaultVis gtsmodel.Visibility, status *gtsmodel.Status) error
	ProcessReplyToID(ctx context.Context, form *apimodel.AdvancedStatusCreateForm, thisAccountID string, status *gtsmodel.Status) gtserror.WithCode
	ProcessMediaIDs(ctx context.Context, form *apimodel.AdvancedStatusCreateForm, thisAccountID string, status *gtsmodel.Status) gtserror.WithCode
	ProcessLanguage(ctx context.Context, form *apimodel.AdvancedStatusCreateForm, accountDefaultLanguage string, status *gtsmodel.Status) error
	ProcessContent(ctx context.Context, form *apimodel.AdvancedStatusCreateForm, accountID string, status *gtsmodel.Status) error
}

type processor struct {
	tc           typeutils.TypeConverter
	db           db.DB
	filter       visibility.Filter
	formatter    text.Formatter
	clientWorker *concurrency.WorkerPool[messages.FromClientAPI]
	parseMention gtsmodel.ParseMentionFunc
}

// New returns a new status processor.
func New(db db.DB, tc typeutils.TypeConverter, clientWorker *concurrency.WorkerPool[messages.FromClientAPI], parseMention gtsmodel.ParseMentionFunc) Processor {
	return &processor{
		tc:           tc,
		db:           db,
		filter:       visibility.NewFilter(db),
		formatter:    text.NewFormatter(db),
		clientWorker: clientWorker,
		parseMention: parseMention,
	}
}

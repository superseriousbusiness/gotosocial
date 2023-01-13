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

package report

import (
	"context"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/concurrency"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
)

type Processor interface {
	ReportsGet(ctx context.Context, account *gtsmodel.Account, resolved *bool, targetAccountID string, maxID string, sinceID string, minID string, limit int) (*apimodel.PageableResponse, gtserror.WithCode)
	ReportGet(ctx context.Context, account *gtsmodel.Account, id string) (*apimodel.Report, gtserror.WithCode)
	Create(ctx context.Context, account *gtsmodel.Account, form *apimodel.ReportCreateRequest) (*apimodel.Report, gtserror.WithCode)
}

type processor struct {
	db           db.DB
	tc           typeutils.TypeConverter
	clientWorker *concurrency.WorkerPool[messages.FromClientAPI]
}

func New(db db.DB, tc typeutils.TypeConverter, clientWorker *concurrency.WorkerPool[messages.FromClientAPI]) Processor {
	return &processor{
		tc:           tc,
		db:           db,
		clientWorker: clientWorker,
	}
}

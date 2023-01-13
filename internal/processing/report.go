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

package processing

import (
	"context"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

func (p *processor) ReportsGet(ctx context.Context, authed *oauth.Auth, resolved *bool, targetAccountID string, maxID string, sinceID string, minID string, limit int) (*apimodel.PageableResponse, gtserror.WithCode) {
	return p.reportProcessor.ReportsGet(ctx, authed.Account, resolved, targetAccountID, maxID, sinceID, minID, limit)
}

func (p *processor) ReportGet(ctx context.Context, authed *oauth.Auth, id string) (*apimodel.Report, gtserror.WithCode) {
	return p.reportProcessor.ReportGet(ctx, authed.Account, id)
}

func (p *processor) ReportCreate(ctx context.Context, authed *oauth.Auth, form *apimodel.ReportCreateRequest) (*apimodel.Report, gtserror.WithCode) {
	return p.reportProcessor.Create(ctx, authed.Account, form)
}

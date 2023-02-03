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
	"fmt"
	"time"

	"github.com/superseriousbusiness/gotosocial/internal/ap"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	"github.com/superseriousbusiness/gotosocial/internal/text"
	"github.com/superseriousbusiness/gotosocial/internal/uris"
)

func (p *processor) Create(ctx context.Context, account *gtsmodel.Account, application *gtsmodel.Application, form *apimodel.AdvancedStatusCreateForm) (*apimodel.Status, gtserror.WithCode) {
	accountURIs := uris.GenerateURIsForAccount(account.Username)
	thisStatusID := id.NewULID()
	local := true
	sensitive := form.Sensitive

	newStatus := &gtsmodel.Status{
		ID:                       thisStatusID,
		URI:                      accountURIs.StatusesURI + "/" + thisStatusID,
		URL:                      accountURIs.StatusesURL + "/" + thisStatusID,
		CreatedAt:                time.Now(),
		UpdatedAt:                time.Now(),
		Local:                    &local,
		AccountID:                account.ID,
		AccountURI:               account.URI,
		ContentWarning:           text.SanitizePlaintext(form.SpoilerText),
		ActivityStreamsType:      ap.ObjectNote,
		Sensitive:                &sensitive,
		Language:                 form.Language,
		CreatedWithApplicationID: application.ID,
		Text:                     form.Status,
	}

	if errWithCode := p.ProcessReplyToID(ctx, form, account.ID, newStatus); errWithCode != nil {
		return nil, errWithCode
	}

	if errWithCode := p.ProcessMediaIDs(ctx, form, account.ID, newStatus); errWithCode != nil {
		return nil, errWithCode
	}

	if err := p.ProcessVisibility(ctx, form, account.Privacy, newStatus); err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	if err := p.ProcessLanguage(ctx, form, account.Language, newStatus); err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	if err := p.ProcessContent(ctx, form, account.ID, newStatus); err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	// put the new status in the database
	if err := p.db.PutStatus(ctx, newStatus); err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	// send it back to the processor for async processing
	p.clientWorker.Queue(messages.FromClientAPI{
		APObjectType:   ap.ObjectNote,
		APActivityType: ap.ActivityCreate,
		GTSModel:       newStatus,
		OriginAccount:  account,
	})

	// return the frontend representation of the new status to the submitter
	apiStatus, err := p.tc.StatusToAPIStatus(ctx, newStatus, account)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("error converting status %s to frontend representation: %s", newStatus.ID, err))
	}

	return apiStatus, nil
}

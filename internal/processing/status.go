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
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

func (p *processor) StatusCreate(authed *oauth.Auth, form *apimodel.AdvancedStatusCreateForm) (*apimodel.Status, error) {
	return p.statusProcessor.Create(authed.Account, authed.Application, form)
}

func (p *processor) StatusDelete(authed *oauth.Auth, targetStatusID string) (*apimodel.Status, error) {
	return p.statusProcessor.Delete(authed.Account, targetStatusID)
}

func (p *processor) StatusFave(authed *oauth.Auth, targetStatusID string) (*apimodel.Status, error) {
	return p.statusProcessor.Fave(authed.Account, targetStatusID)
}

func (p *processor) StatusBoost(authed *oauth.Auth, targetStatusID string) (*apimodel.Status, gtserror.WithCode) {
	return p.statusProcessor.Boost(authed.Account, authed.Application, targetStatusID)
}

func (p *processor) StatusUnboost(authed *oauth.Auth, targetStatusID string) (*apimodel.Status, gtserror.WithCode) {
	return p.statusProcessor.Unboost(authed.Account, authed.Application, targetStatusID)
}

func (p *processor) StatusBoostedBy(authed *oauth.Auth, targetStatusID string) ([]*apimodel.Account, gtserror.WithCode) {
	return p.statusProcessor.BoostedBy(authed.Account, targetStatusID)
}

func (p *processor) StatusFavedBy(authed *oauth.Auth, targetStatusID string) ([]*apimodel.Account, error) {
	return p.statusProcessor.FavedBy(authed.Account, targetStatusID)
}

func (p *processor) StatusGet(authed *oauth.Auth, targetStatusID string) (*apimodel.Status, error) {
	return p.statusProcessor.Get(authed.Account, targetStatusID)
}

func (p *processor) StatusUnfave(authed *oauth.Auth, targetStatusID string) (*apimodel.Status, error) {
	return p.statusProcessor.Unfave(authed.Account, targetStatusID)
}

func (p *processor) StatusGetContext(authed *oauth.Auth, targetStatusID string) (*apimodel.Context, gtserror.WithCode) {
	return p.statusProcessor.Context(authed.Account, targetStatusID)
}

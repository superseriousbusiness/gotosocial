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
	"fmt"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

func (p *processor) InstanceGet(domain string) (*apimodel.Instance, gtserror.WithCode) {
	i := &gtsmodel.Instance{}
	if err := p.db.GetWhere([]db.Where{{Key: "domain", Value: domain}}, i); err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("db error fetching instance %s: %s", p.config.Host, err))
	}

	ai, err := p.tc.InstanceToMasto(i)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("error converting instance to api representation: %s", err))
	}

	return ai, nil
}

func (p *processor) InstancePatch(form *apimodel.InstanceSettingsUpdateRequest) (*apimodel.Instance, gtserror.WithCode) {
	// fetch the instance entry from the db for processing
	i := &gtsmodel.Instance{}
	if err := p.db.GetWhere([]db.Where{{Key: "domain", Value: p.config.Host}}, i); err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("db error fetching instance %s: %s", p.config.Host, err))
	}

	// fetch the instance account from the db for processing
	ia := &gtsmodel.Account{}
	if err := p.db.GetLocalAccountByUsername(p.config.Host, ia); err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("db error fetching instance account %s: %s", p.config.Host, err))
	}

	// validate & update site title if it's set on the form
	if form.Title != nil {
		if err := util.ValidateSiteTitle(*form.Title); err != nil {
			return nil, gtserror.NewErrorBadRequest(err, fmt.Sprintf("site title invalid: %s", err))
		}
		i.Title = util.RemoveHTML(*form.Title) // don't allow html in site title
	}

	// validate & update site contact account if it's set on the form
	if form.ContactUsername != nil {
		// make sure the account with the given username exists in the db
		contactAccount := &gtsmodel.Account{}
		if err := p.db.GetLocalAccountByUsername(*form.ContactUsername, contactAccount); err != nil {
			return nil, gtserror.NewErrorBadRequest(err, fmt.Sprintf("account with username %s not retrievable", *form.ContactUsername))
		}
		// make sure it has a user associated with it
		contactUser := &gtsmodel.User{}
		if err := p.db.GetWhere([]db.Where{{Key: "account_id", Value: contactAccount.ID}}, contactUser); err != nil {
			return nil, gtserror.NewErrorBadRequest(err, fmt.Sprintf("user for account with username %s not retrievable", *form.ContactUsername))
		}
		// suspended accounts cannot be contact accounts
		if !contactAccount.SuspendedAt.IsZero() {
			err := fmt.Errorf("selected contact account %s is suspended", contactAccount.Username)
			return nil, gtserror.NewErrorBadRequest(err, err.Error())
		}
		// unconfirmed or unapproved users cannot be contacts
		if contactUser.ConfirmedAt.IsZero() {
			err := fmt.Errorf("user of selected contact account %s is not confirmed", contactAccount.Username)
			return nil, gtserror.NewErrorBadRequest(err, err.Error())
		}
		if !contactUser.Approved {
			err := fmt.Errorf("user of selected contact account %s is not approved", contactAccount.Username)
			return nil, gtserror.NewErrorBadRequest(err, err.Error())
		}
		// contact account user must be admin or moderator otherwise what's the point of contacting them
		if !contactUser.Admin && !contactUser.Moderator {
			err := fmt.Errorf("user of selected contact account %s is neither admin nor moderator", contactAccount.Username)
			return nil, gtserror.NewErrorBadRequest(err, err.Error())
		}
		i.ContactAccountID = contactAccount.ID
	}

	// validate & update site contact email if it's set on the form
	if form.ContactEmail != nil {
		if err := util.ValidateEmail(*form.ContactEmail); err != nil {
			return nil, gtserror.NewErrorBadRequest(err, err.Error())
		}
		i.ContactEmail = *form.ContactEmail
	}

	// validate & update site short description if it's set on the form
	if form.ShortDescription != nil {
		if err := util.ValidateSiteShortDescription(*form.ShortDescription); err != nil {
			return nil, gtserror.NewErrorBadRequest(err, err.Error())
		}
		i.ShortDescription = util.SanitizeHTML(*form.ShortDescription) // html is OK in site description, but we should sanitize it
	}

	// validate & update site description if it's set on the form
	if form.Description != nil {
		if err := util.ValidateSiteDescription(*form.Description); err != nil {
			return nil, gtserror.NewErrorBadRequest(err, err.Error())
		}
		i.Description = util.SanitizeHTML(*form.Description) // html is OK in site description, but we should sanitize it
	}

	// validate & update site terms if it's set on the form
	if form.Terms != nil {
		if err := util.ValidateSiteTerms(*form.Terms); err != nil {
			return nil, gtserror.NewErrorBadRequest(err, err.Error())
		}
		i.Terms = util.SanitizeHTML(*form.Terms) // html is OK in site terms, but we should sanitize it
	}

	// process avatar if provided
	if form.Avatar != nil && form.Avatar.Size != 0 {
		_, err := p.accountProcessor.UpdateAvatar(form.Avatar, ia.ID)
		if err != nil {
			return nil, gtserror.NewErrorBadRequest(err, "error processing avatar")
		}
	}

	// process header if provided
	if form.Header != nil && form.Header.Size != 0 {
		_, err := p.accountProcessor.UpdateHeader(form.Header, ia.ID)
		if err != nil {
			return nil, gtserror.NewErrorBadRequest(err, "error processing header")
		}
	}

	if err := p.db.UpdateByID(i.ID, i); err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("db error updating instance %s: %s", p.config.Host, err))
	}

	ai, err := p.tc.InstanceToMasto(i)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("error converting instance to api representation: %s", err))
	}

	return ai, nil
}

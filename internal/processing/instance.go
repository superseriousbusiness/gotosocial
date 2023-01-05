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
	"fmt"
	"sort"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/text"
	"github.com/superseriousbusiness/gotosocial/internal/util"
	"github.com/superseriousbusiness/gotosocial/internal/validate"
)

func (p *processor) InstanceGet(ctx context.Context, domain string) (*apimodel.Instance, gtserror.WithCode) {
	i := &gtsmodel.Instance{}
	if err := p.db.GetWhere(ctx, []db.Where{{Key: "domain", Value: domain}}, i); err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("db error fetching instance %s: %s", domain, err))
	}

	ai, err := p.tc.InstanceToAPIInstance(ctx, i)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("error converting instance to api representation: %s", err))
	}

	return ai, nil
}

func (p *processor) InstancePeersGet(ctx context.Context, authed *oauth.Auth, includeSuspended bool, includeOpen bool, flat bool) (interface{}, gtserror.WithCode) {
	domains := []*apimodel.Domain{}

	if includeOpen {
		if !config.GetInstanceExposePeers() && (authed.Account == nil || authed.User == nil) {
			err := fmt.Errorf("peers open query requires an authenticated account/user")
			return nil, gtserror.NewErrorUnauthorized(err, err.Error())
		}

		instances, err := p.db.GetInstancePeers(ctx, false)
		if err != nil && err != db.ErrNoEntries {
			err = fmt.Errorf("error selecting instance peers: %s", err)
			return nil, gtserror.NewErrorInternalError(err)
		}

		for _, i := range instances {
			domain := &apimodel.Domain{Domain: i.Domain}
			domains = append(domains, domain)
		}
	}

	if includeSuspended {
		if !config.GetInstanceExposeSuspended() && (authed.Account == nil || authed.User == nil) {
			err := fmt.Errorf("peers suspended query requires an authenticated account/user")
			return nil, gtserror.NewErrorUnauthorized(err, err.Error())
		}

		domainBlocks := []*gtsmodel.DomainBlock{}
		if err := p.db.GetAll(ctx, &domainBlocks); err != nil && err != db.ErrNoEntries {
			return nil, gtserror.NewErrorInternalError(err)
		}

		for _, d := range domainBlocks {
			if *d.Obfuscate {
				d.Domain = obfuscate(d.Domain)
			}

			domain := &apimodel.Domain{
				Domain:        d.Domain,
				SuspendedAt:   util.FormatISO8601(d.CreatedAt),
				PublicComment: d.PublicComment,
			}
			domains = append(domains, domain)
		}
	}

	sort.Slice(domains, func(i, j int) bool {
		return domains[i].Domain < domains[j].Domain
	})

	if flat {
		flattened := []string{}
		for _, d := range domains {
			flattened = append(flattened, d.Domain)
		}
		return flattened, nil
	}

	return domains, nil
}

func (p *processor) InstancePatch(ctx context.Context, form *apimodel.InstanceSettingsUpdateRequest) (*apimodel.Instance, gtserror.WithCode) {
	// fetch the instance entry from the db for processing
	i := &gtsmodel.Instance{}
	host := config.GetHost()
	if err := p.db.GetWhere(ctx, []db.Where{{Key: "domain", Value: host}}, i); err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("db error fetching instance %s: %s", host, err))
	}

	// fetch the instance account from the db for processing
	ia, err := p.db.GetInstanceAccount(ctx, "")
	if err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("db error fetching instance account %s: %s", host, err))
	}

	updatingColumns := []string{}

	// validate & update site title if it's set on the form
	if form.Title != nil {
		if err := validate.SiteTitle(*form.Title); err != nil {
			return nil, gtserror.NewErrorBadRequest(err, fmt.Sprintf("site title invalid: %s", err))
		}
		updatingColumns = append(updatingColumns, "title")
		i.Title = text.SanitizePlaintext(*form.Title) // don't allow html in site title
	}

	// validate & update site contact account if it's set on the form
	if form.ContactUsername != nil {
		// make sure the account with the given username exists in the db
		contactAccount, err := p.db.GetAccountByUsernameDomain(ctx, *form.ContactUsername, "")
		if err != nil {
			return nil, gtserror.NewErrorBadRequest(err, fmt.Sprintf("account with username %s not retrievable", *form.ContactUsername))
		}
		// make sure it has a user associated with it
		contactUser, err := p.db.GetUserByAccountID(ctx, contactAccount.ID)
		if err != nil {
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
		if !*contactUser.Approved {
			err := fmt.Errorf("user of selected contact account %s is not approved", contactAccount.Username)
			return nil, gtserror.NewErrorBadRequest(err, err.Error())
		}
		// contact account user must be admin or moderator otherwise what's the point of contacting them
		if !*contactUser.Admin && !*contactUser.Moderator {
			err := fmt.Errorf("user of selected contact account %s is neither admin nor moderator", contactAccount.Username)
			return nil, gtserror.NewErrorBadRequest(err, err.Error())
		}
		updatingColumns = append(updatingColumns, "contact_account_id")
		i.ContactAccountID = contactAccount.ID
	}

	// validate & update site contact email if it's set on the form
	if form.ContactEmail != nil {
		contactEmail := *form.ContactEmail
		if contactEmail != "" {
			if err := validate.Email(contactEmail); err != nil {
				return nil, gtserror.NewErrorBadRequest(err, err.Error())
			}
		}
		updatingColumns = append(updatingColumns, "contact_email")
		i.ContactEmail = contactEmail
	}

	// validate & update site short description if it's set on the form
	if form.ShortDescription != nil {
		if err := validate.SiteShortDescription(*form.ShortDescription); err != nil {
			return nil, gtserror.NewErrorBadRequest(err, err.Error())
		}
		updatingColumns = append(updatingColumns, "short_description")
		i.ShortDescription = text.SanitizeHTML(*form.ShortDescription) // html is OK in site description, but we should sanitize it
	}

	// validate & update site description if it's set on the form
	if form.Description != nil {
		if err := validate.SiteDescription(*form.Description); err != nil {
			return nil, gtserror.NewErrorBadRequest(err, err.Error())
		}
		updatingColumns = append(updatingColumns, "description")
		i.Description = text.SanitizeHTML(*form.Description) // html is OK in site description, but we should sanitize it
	}

	// validate & update site terms if it's set on the form
	if form.Terms != nil {
		if err := validate.SiteTerms(*form.Terms); err != nil {
			return nil, gtserror.NewErrorBadRequest(err, err.Error())
		}
		updatingColumns = append(updatingColumns, "terms")
		i.Terms = text.SanitizeHTML(*form.Terms) // html is OK in site terms, but we should sanitize it
	}

	var updateInstanceAccount bool

	// process instance avatar if provided
	if form.Avatar != nil && form.Avatar.Size != 0 {
		avatarInfo, err := p.accountProcessor.UpdateAvatar(ctx, form.Avatar, form.AvatarDescription, ia.ID)
		if err != nil {
			return nil, gtserror.NewErrorBadRequest(err, "error processing avatar")
		}
		ia.AvatarMediaAttachmentID = avatarInfo.ID
		ia.AvatarMediaAttachment = avatarInfo
		updateInstanceAccount = true
	}

	// process instance header if provided
	if form.Header != nil && form.Header.Size != 0 {
		headerInfo, err := p.accountProcessor.UpdateHeader(ctx, form.Header, nil, ia.ID)
		if err != nil {
			return nil, gtserror.NewErrorBadRequest(err, "error processing header")
		}
		ia.HeaderMediaAttachmentID = headerInfo.ID
		ia.HeaderMediaAttachment = headerInfo
		updateInstanceAccount = true
	}

	if updateInstanceAccount {
		// if either avatar or header is updated, we need
		// to update the instance account that stores them
		if err := p.db.UpdateAccount(ctx, ia); err != nil {
			return nil, gtserror.NewErrorInternalError(fmt.Errorf("db error updating instance account: %s", err))
		}
	}

	if len(updatingColumns) != 0 {
		if err := p.db.UpdateByID(ctx, i, i.ID, updatingColumns...); err != nil {
			return nil, gtserror.NewErrorInternalError(fmt.Errorf("db error updating instance %s: %s", host, err))
		}
	}

	ai, err := p.tc.InstanceToAPIInstance(ctx, i)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("error converting instance to api representation: %s", err))
	}

	return ai, nil
}

func obfuscate(domain string) string {
	obfuscated := make([]rune, len(domain))
	for i, r := range domain {
		if i%3 == 1 || i%5 == 1 {
			obfuscated[i] = '*'
		} else {
			obfuscated[i] = r
		}
	}
	return string(obfuscated)
}

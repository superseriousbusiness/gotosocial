// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package processing

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"code.superseriousbusiness.org/gotosocial/internal/text"
	"code.superseriousbusiness.org/gotosocial/internal/typeutils"
	"code.superseriousbusiness.org/gotosocial/internal/util"
	"code.superseriousbusiness.org/gotosocial/internal/util/xslices"
	"code.superseriousbusiness.org/gotosocial/internal/validate"
)

func (p *Processor) InstanceGetV1(ctx context.Context) (*apimodel.InstanceV1, gtserror.WithCode) {
	i, err := p.getThisInstance(ctx)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("db error fetching instance: %s", err))
	}

	ai, err := p.converter.InstanceToAPIV1Instance(ctx, i)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("error converting instance to api representation: %s", err))
	}

	return ai, nil
}

func (p *Processor) InstanceGetV2(ctx context.Context) (*apimodel.InstanceV2, gtserror.WithCode) {
	i, err := p.getThisInstance(ctx)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("db error fetching instance: %s", err))
	}

	ai, err := p.converter.InstanceToAPIV2Instance(ctx, i)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("error converting instance to api representation: %s", err))
	}

	return ai, nil
}

func (p *Processor) InstancePeersGet(
	ctx context.Context,
	includeBlocked bool,
	includeAllowed bool,
	includeOpen bool,
	flatten bool,
	includeSeverity bool,
) (any, gtserror.WithCode) {
	var (
		domainPerms []gtsmodel.DomainPermission
		apiDomains  []*apimodel.Domain
	)

	if includeBlocked {
		blocks, err := p.state.DB.GetDomainBlocks(ctx)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			err := gtserror.Newf("db error getting domain blocks: %w", err)
			return nil, gtserror.NewErrorInternalError(err)
		}

		for _, block := range blocks {
			domainPerms = append(domainPerms, block)
		}

	} else if includeAllowed {
		allows, err := p.state.DB.GetDomainAllows(ctx)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			err := gtserror.Newf("db error getting domain allows: %w", err)
			return nil, gtserror.NewErrorInternalError(err)
		}

		for _, allow := range allows {
			domainPerms = append(domainPerms, allow)
		}
	}

	for _, domainPerm := range domainPerms {
		// Domain may be in Punycode,
		// de-punify it just in case.
		domain := domainPerm.GetDomain()
		depunied, err := util.DePunify(domain)
		if err != nil {
			log.Errorf(ctx, "couldn't depunify domain %s: %v", domain, err)
			continue
		}

		if util.PtrOrZero(domainPerm.GetObfuscate()) {
			// Obfuscate the de-punified version.
			depunied = obfuscate(depunied)
		}

		apiDomain := &apimodel.Domain{
			Domain:  depunied,
			Comment: util.Ptr(domainPerm.GetPublicComment()),
		}

		if domainPerm.GetType() == gtsmodel.DomainPermissionBlock {
			const severity = "suspend"
			apiDomain.Severity = severity
			suspendedAt := domainPerm.GetCreatedAt()
			apiDomain.SuspendedAt = util.FormatISO8601(suspendedAt)
		}

		apiDomains = append(apiDomains, apiDomain)
	}

	if includeOpen {
		instances, err := p.state.DB.GetInstancePeers(ctx, false)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			err = gtserror.Newf("db error getting instance peers: %w", err)
			return nil, gtserror.NewErrorInternalError(err)
		}

		for _, instance := range instances {
			// Domain may be in Punycode,
			// de-punify it just in case.
			domain := instance.Domain
			depunied, err := util.DePunify(domain)
			if err != nil {
				log.Errorf(ctx, "couldn't depunify domain %s: %v", domain, err)
				continue
			}

			apiDomains = append(
				apiDomains,
				&apimodel.Domain{
					Domain: depunied,
				},
			)
		}
	}

	// Sort a-z.
	slices.SortFunc(
		apiDomains,
		func(a, b *apimodel.Domain) int {
			return strings.Compare(a.Domain, b.Domain)
		},
	)

	// Deduplicate.
	apiDomains = xslices.DeduplicateFunc(
		apiDomains,
		func(v *apimodel.Domain) string {
			return v.Domain
		},
	)

	if flatten {
		// Return just the domains.
		return xslices.Gather(
			[]string{},
			apiDomains,
			func(v *apimodel.Domain) string {
				return v.Domain
			},
		), nil
	}

	return apiDomains, nil
}

func (p *Processor) InstanceGetRules(ctx context.Context) ([]apimodel.InstanceRule, gtserror.WithCode) {
	i, err := p.getThisInstance(ctx)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("db error fetching instance: %s", err))
	}

	return typeutils.InstanceRulesToAPIRules(i.Rules), nil
}

func (p *Processor) InstancePatch(ctx context.Context, form *apimodel.InstanceSettingsUpdateRequest) (*apimodel.InstanceV1, gtserror.WithCode) {
	// Fetch this instance from the db for processing.
	instance, err := p.getThisInstance(ctx)
	if err != nil {
		err = fmt.Errorf("db error fetching instance: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Fetch this instance account from the db for processing.
	instanceAcc, err := p.state.DB.GetInstanceAccount(ctx, "")
	if err != nil {
		err = fmt.Errorf("db error fetching instance account: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Columns to update
	// in the database.
	var columns []string

	// Validate & update site
	// title if set on the form.
	if form.Title != nil {
		title := *form.Title
		if err := validate.SiteTitle(title); err != nil {
			return nil, gtserror.NewErrorBadRequest(err, err.Error())
		}

		// Don't allow html in site title.
		instance.Title = text.StripHTMLFromText(title)
		columns = append(columns, "title")
	}

	// Validate & update site contact
	// account if set on the form.
	//
	// Empty username unsets contact.
	if form.ContactUsername != nil {
		contactAccountID, err := p.contactAccountIDForUsername(ctx, *form.ContactUsername)
		if err != nil {
			return nil, gtserror.NewErrorBadRequest(err, err.Error())
		}

		columns = append(columns, "contact_account_id")
		instance.ContactAccountID = contactAccountID
	}

	// Validate & update contact
	// email if set on the form.
	//
	// Empty email unsets contact.
	if form.ContactEmail != nil {
		contactEmail := *form.ContactEmail
		if contactEmail != "" {
			if err := validate.Email(contactEmail); err != nil {
				return nil, gtserror.NewErrorBadRequest(err, err.Error())
			}
		}

		columns = append(columns, "contact_email")
		instance.ContactEmail = contactEmail
	}

	// Validate & update site short
	// description if set on the form.
	if form.ShortDescription != nil {
		shortDescription := *form.ShortDescription
		if err := validate.SiteShortDescription(shortDescription); err != nil {
			return nil, gtserror.NewErrorBadRequest(err, err.Error())
		}

		// Parse description as Markdown, keep
		// the raw version for later editing.
		instance.ShortDescriptionText = shortDescription
		instance.ShortDescription = p.formatter.FromMarkdown(ctx, p.parseMentionFunc, instanceAcc.ID, "", shortDescription).HTML
		columns = append(columns, []string{"short_description", "short_description_text"}...)
	}

	// validate & update site description if it's set on the form
	if form.Description != nil {
		description := *form.Description
		if err := validate.SiteDescription(description); err != nil {
			return nil, gtserror.NewErrorBadRequest(err, err.Error())
		}

		// Parse description as Markdown, keep
		// the raw version for later editing.
		instance.DescriptionText = description
		instance.Description = p.formatter.FromMarkdown(ctx, p.parseMentionFunc, instanceAcc.ID, "", description).HTML
		columns = append(columns, []string{"description", "description_text"}...)
	}

	// validate & update site custom css if it's set on the form
	if form.CustomCSS != nil {
		customCSS := *form.CustomCSS
		if err := validate.InstanceCustomCSS(customCSS); err != nil {
			return nil, gtserror.NewErrorBadRequest(err, err.Error())
		}

		instance.CustomCSS = text.StripHTMLFromText(customCSS)
		columns = append(columns, []string{"custom_css"}...)
	}

	// Validate & update site
	// terms if set on the form.
	if form.Terms != nil {
		terms := *form.Terms
		if err := validate.SiteTerms(terms); err != nil {
			return nil, gtserror.NewErrorBadRequest(err, err.Error())
		}

		// Parse terms as Markdown, keep
		// the raw version for later editing.
		instance.TermsText = terms
		instance.Terms = p.formatter.FromMarkdown(ctx, p.parseMentionFunc, "", "", terms).HTML
		columns = append(columns, []string{"terms", "terms_text"}...)
	}

	var updateInstanceAccount bool

	if form.Avatar != nil && form.Avatar.Size != 0 {
		// Process instance avatar image + description.
		avatarInfo, errWithCode := p.account.UpdateAvatar(ctx,
			instanceAcc,
			form.Avatar,
			form.AvatarDescription,
		)
		if errWithCode != nil {
			return nil, errWithCode
		}
		instanceAcc.AvatarMediaAttachmentID = avatarInfo.ID
		instanceAcc.AvatarMediaAttachment = avatarInfo
		updateInstanceAccount = true
	} else if form.AvatarDescription != nil && instanceAcc.AvatarMediaAttachment != nil {
		// Process just the description for the existing avatar.
		instanceAcc.AvatarMediaAttachment.Description = *form.AvatarDescription
		if err := p.state.DB.UpdateAttachment(ctx, instanceAcc.AvatarMediaAttachment, "description"); err != nil {
			err = fmt.Errorf("db error updating instance avatar description: %w", err)
			return nil, gtserror.NewErrorInternalError(err)
		}
	}

	if form.Header != nil && form.Header.Size != 0 {
		// process instance header image
		headerInfo, errWithCode := p.account.UpdateHeader(ctx,
			instanceAcc,
			form.Header,
			nil,
		)
		if errWithCode != nil {
			return nil, errWithCode
		}
		instanceAcc.HeaderMediaAttachmentID = headerInfo.ID
		instanceAcc.HeaderMediaAttachment = headerInfo
		updateInstanceAccount = true
	}

	if updateInstanceAccount {
		// If either avatar or header is updated, we need
		// to update the instance account that stores them.
		if err := p.state.DB.UpdateAccount(ctx, instanceAcc); err != nil {
			err = fmt.Errorf("db error updating instance account: %w", err)
			return nil, gtserror.NewErrorInternalError(err, err.Error())
		}
	}

	if len(columns) != 0 {
		if err := p.state.DB.UpdateInstance(ctx, instance, columns...); err != nil {
			err = fmt.Errorf("db error updating instance: %w", err)
			return nil, gtserror.NewErrorInternalError(err, err.Error())
		}
	}

	return p.InstanceGetV1(ctx)
}

func (p *Processor) getThisInstance(ctx context.Context) (*gtsmodel.Instance, error) {
	instance, err := p.state.DB.GetInstance(ctx, config.GetHost())
	if err != nil {
		return nil, err
	}

	return instance, nil
}

func (p *Processor) contactAccountIDForUsername(ctx context.Context, username string) (string, error) {
	if username == "" {
		// Easy: unset
		// contact account.
		return "", nil
	}

	// Make sure local account with the given username exists in the db.
	contactAccount, err := p.state.DB.GetAccountByUsernameDomain(ctx, username, "")
	if err != nil {
		err = fmt.Errorf("db error getting selected contact account with username %s: %w", username, err)
		return "", err
	}

	// Make sure account corresponds to a user.
	contactUser, err := p.state.DB.GetUserByAccountID(ctx, contactAccount.ID)
	if err != nil {
		err = fmt.Errorf("db error getting user for selected contact account %s: %w", username, err)
		return "", err
	}

	// Ensure account/user is:
	//
	// - confirmed and approved
	// - not suspended
	// - an admin or a moderator
	if contactUser.ConfirmedAt.IsZero() {
		err := fmt.Errorf("user of selected contact account %s is not confirmed", contactAccount.Username)
		return "", err
	}

	if !*contactUser.Approved {
		err := fmt.Errorf("user of selected contact account %s is not approved", contactAccount.Username)
		return "", err
	}

	if !contactAccount.SuspendedAt.IsZero() {
		err := fmt.Errorf("selected contact account %s is suspended", contactAccount.Username)
		return "", err
	}

	if !*contactUser.Admin && !*contactUser.Moderator {
		err := fmt.Errorf("user of selected contact account %s is neither admin nor moderator", contactAccount.Username)
		return "", err
	}

	// All good!
	return contactAccount.ID, nil
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

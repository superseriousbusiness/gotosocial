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
	"fmt"
	"sort"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/text"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/internal/util"
	"github.com/superseriousbusiness/gotosocial/internal/validate"
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

func (p *Processor) InstancePeersGet(ctx context.Context, includeSuspended bool, includeOpen bool, flat bool) (interface{}, gtserror.WithCode) {
	domains := []*apimodel.Domain{}

	if includeOpen {
		instances, err := p.state.DB.GetInstancePeers(ctx, false)
		if err != nil && err != db.ErrNoEntries {
			err = fmt.Errorf("error selecting instance peers: %s", err)
			return nil, gtserror.NewErrorInternalError(err)
		}

		for _, i := range instances {
			// Domain may be in Punycode,
			// de-punify it just in case.
			d, err := util.DePunify(i.Domain)
			if err != nil {
				log.Errorf(ctx, "couldn't depunify domain %s: %s", i.Domain, err)
				continue
			}

			domains = append(domains, &apimodel.Domain{Domain: d})
		}
	}

	if includeSuspended {
		domainBlocks := []*gtsmodel.DomainBlock{}
		if err := p.state.DB.GetAll(ctx, &domainBlocks); err != nil && err != db.ErrNoEntries {
			return nil, gtserror.NewErrorInternalError(err)
		}

		for _, domainBlock := range domainBlocks {
			// Domain may be in Punycode,
			// de-punify it just in case.
			d, err := util.DePunify(domainBlock.Domain)
			if err != nil {
				log.Errorf(ctx, "couldn't depunify domain %s: %s", domainBlock.Domain, err)
				continue
			}

			if *domainBlock.Obfuscate {
				// Obfuscate the de-punified version.
				d = obfuscate(d)
			}

			domains = append(domains, &apimodel.Domain{
				Domain:        d,
				SuspendedAt:   util.FormatISO8601(domainBlock.CreatedAt),
				PublicComment: domainBlock.PublicComment,
			})
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
		instance.Title = text.SanitizeToPlaintext(title)
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

		instance.CustomCSS = text.SanitizeToPlaintext(customCSS)
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

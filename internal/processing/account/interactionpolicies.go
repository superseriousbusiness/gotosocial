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

package account

import (
	"cmp"
	"context"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
)

func (p *Processor) DefaultInteractionPoliciesGet(
	ctx context.Context,
	requester *gtsmodel.Account,
) (*apimodel.DefaultPolicies, gtserror.WithCode) {
	// Ensure account settings populated.
	if err := p.populateAccountSettings(ctx, requester); err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Take set "direct" policy
	// or global default.
	direct := cmp.Or(
		requester.Settings.InteractionPolicyDirect,
		gtsmodel.DefaultInteractionPolicyDirect(),
	)

	directAPI, err := p.converter.InteractionPolicyToAPIInteractionPolicy(ctx, direct, nil, nil)
	if err != nil {
		err := gtserror.Newf("error converting interaction policy direct: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Take set "private" policy
	// or global default.
	private := cmp.Or(
		requester.Settings.InteractionPolicyFollowersOnly,
		gtsmodel.DefaultInteractionPolicyFollowersOnly(),
	)

	privateAPI, err := p.converter.InteractionPolicyToAPIInteractionPolicy(ctx, private, nil, nil)
	if err != nil {
		err := gtserror.Newf("error converting interaction policy private: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Take set "unlisted" policy
	// or global default.
	unlisted := cmp.Or(
		requester.Settings.InteractionPolicyUnlocked,
		gtsmodel.DefaultInteractionPolicyUnlocked(),
	)

	unlistedAPI, err := p.converter.InteractionPolicyToAPIInteractionPolicy(ctx, unlisted, nil, nil)
	if err != nil {
		err := gtserror.Newf("error converting interaction policy unlisted: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Take set "public" policy
	// or global default.
	public := cmp.Or(
		requester.Settings.InteractionPolicyPublic,
		gtsmodel.DefaultInteractionPolicyPublic(),
	)

	publicAPI, err := p.converter.InteractionPolicyToAPIInteractionPolicy(ctx, public, nil, nil)
	if err != nil {
		err := gtserror.Newf("error converting interaction policy public: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return &apimodel.DefaultPolicies{
		Direct:   *directAPI,
		Private:  *privateAPI,
		Unlisted: *unlistedAPI,
		Public:   *publicAPI,
	}, nil
}

func (p *Processor) DefaultInteractionPoliciesUpdate(
	ctx context.Context,
	requester *gtsmodel.Account,
	form *apimodel.UpdateInteractionPoliciesRequest,
) (*apimodel.DefaultPolicies, gtserror.WithCode) {
	// Lock on this account as we're modifying its Settings.
	unlock := p.state.ProcessingLocks.Lock(requester.URI)
	defer unlock()

	// Ensure account settings populated.
	if err := p.populateAccountSettings(ctx, requester); err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	if form.Direct == nil {
		// Unset/return to global default.
		requester.Settings.InteractionPolicyDirect = nil
	} else {
		policy, err := typeutils.APIInteractionPolicyToInteractionPolicy(
			form.Direct,
			apimodel.VisibilityDirect,
		)
		if err != nil {
			return nil, gtserror.NewErrorUnprocessableEntity(err, err.Error())
		}

		// Set new default policy.
		requester.Settings.InteractionPolicyDirect = policy
	}

	if form.Private == nil {
		// Unset/return to global default.
		requester.Settings.InteractionPolicyFollowersOnly = nil
	} else {
		policy, err := typeutils.APIInteractionPolicyToInteractionPolicy(
			form.Private,
			apimodel.VisibilityPrivate,
		)
		if err != nil {
			return nil, gtserror.NewErrorUnprocessableEntity(err, err.Error())
		}

		// Set new default policy.
		requester.Settings.InteractionPolicyFollowersOnly = policy
	}

	if form.Unlisted == nil {
		// Unset/return to global default.
		requester.Settings.InteractionPolicyUnlocked = nil
	} else {
		policy, err := typeutils.APIInteractionPolicyToInteractionPolicy(
			form.Unlisted,
			apimodel.VisibilityUnlisted,
		)
		if err != nil {
			return nil, gtserror.NewErrorUnprocessableEntity(err, err.Error())
		}

		// Set new default policy.
		requester.Settings.InteractionPolicyUnlocked = policy
	}

	if form.Public == nil {
		// Unset/return to global default.
		requester.Settings.InteractionPolicyPublic = nil
	} else {
		policy, err := typeutils.APIInteractionPolicyToInteractionPolicy(
			form.Public,
			apimodel.VisibilityPublic,
		)
		if err != nil {
			return nil, gtserror.NewErrorUnprocessableEntity(err, err.Error())
		}

		// Set new default policy.
		requester.Settings.InteractionPolicyPublic = policy
	}

	if err := p.state.DB.UpdateAccountSettings(ctx, requester.Settings); err != nil {
		err := gtserror.Newf("db error updating setttings: %w", err)
		return nil, gtserror.NewErrorInternalError(err, err.Error())
	}

	return p.DefaultInteractionPoliciesGet(ctx, requester)
}

// populateAccountSettings just ensures that
// Settings is populated on the given account.
func (p *Processor) populateAccountSettings(
	ctx context.Context,
	acct *gtsmodel.Account,
) error {
	if acct.Settings != nil {
		// Already populated.
		return nil
	}

	// Not populated,
	// get from db.
	var err error
	acct.Settings, err = p.state.DB.GetAccountSettings(ctx, acct.ID)
	if err != nil {
		return gtserror.Newf(
			"db error getting settings for account %s: %w",
			acct.ID, err,
		)
	}

	return nil
}

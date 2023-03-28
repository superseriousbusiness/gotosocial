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
	"context"
	"errors"
	"fmt"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

// FollowersGet fetches a list of the target account's followers.
func (p *Processor) FollowersGet(ctx context.Context, requestingAccount *gtsmodel.Account, targetAccountID string) ([]apimodel.Account, gtserror.WithCode) {
	if blocked, err := p.state.DB.IsEitherBlocked(ctx, requestingAccount.ID, targetAccountID); err != nil {
		err = fmt.Errorf("FollowersGet: db error checking block: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	} else if blocked {
		err = errors.New("FollowersGet: block exists between accounts")
		return nil, gtserror.NewErrorNotFound(err)
	}

	follows, err := p.state.DB.GetAccountFollowers(ctx, targetAccountID)
	if err != nil {
		if !errors.Is(err, db.ErrNoEntries) {
			err = fmt.Errorf("FollowersGet: db error getting followers: %w", err)
			return nil, gtserror.NewErrorInternalError(err)
		}
		return []apimodel.Account{}, nil
	}

	return p.accountsFromFollows(ctx, follows, requestingAccount.ID)
}

// FollowingGet fetches a list of the accounts that target account is following.
func (p *Processor) FollowingGet(ctx context.Context, requestingAccount *gtsmodel.Account, targetAccountID string) ([]apimodel.Account, gtserror.WithCode) {
	if blocked, err := p.state.DB.IsEitherBlocked(ctx, requestingAccount.ID, targetAccountID); err != nil {
		err = fmt.Errorf("FollowingGet: db error checking block: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	} else if blocked {
		err = errors.New("FollowingGet: block exists between accounts")
		return nil, gtserror.NewErrorNotFound(err)
	}

	follows, err := p.state.DB.GetAccountFollows(ctx, targetAccountID)
	if err != nil {
		if !errors.Is(err, db.ErrNoEntries) {
			err = fmt.Errorf("FollowingGet: db error getting followers: %w", err)
			return nil, gtserror.NewErrorInternalError(err)
		}
		return []apimodel.Account{}, nil
	}

	return p.targetAccountsFromFollows(ctx, follows, requestingAccount.ID)
}

// RelationshipGet returns a relationship model describing the relationship of the targetAccount to the Authed account.
func (p *Processor) RelationshipGet(ctx context.Context, requestingAccount *gtsmodel.Account, targetAccountID string) (*apimodel.Relationship, gtserror.WithCode) {
	if requestingAccount == nil {
		return nil, gtserror.NewErrorForbidden(errors.New("not authed"))
	}

	gtsR, err := p.state.DB.GetRelationship(ctx, requestingAccount.ID, targetAccountID)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("error getting relationship: %s", err))
	}

	r, err := p.tc.RelationshipToAPIRelationship(ctx, gtsR)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("error converting relationship: %s", err))
	}

	return r, nil
}

func (p *Processor) accountsFromFollows(ctx context.Context, follows []*gtsmodel.Follow, requestingAccountID string) ([]apimodel.Account, gtserror.WithCode) {
	accounts := make([]apimodel.Account, 0, len(follows))
	for _, follow := range follows {
		if follow.Account == nil {
			// No account set for some reason; just skip.
			log.WithContext(ctx).WithField("follow", follow).Warn("follow had no associated account")
			continue
		}

		if blocked, err := p.state.DB.IsEitherBlocked(ctx, requestingAccountID, follow.AccountID); err != nil {
			err = fmt.Errorf("accountsFromFollows: db error checking block: %w", err)
			return nil, gtserror.NewErrorInternalError(err)
		} else if blocked {
			continue
		}

		account, err := p.tc.AccountToAPIAccountPublic(ctx, follow.Account)
		if err != nil {
			err = fmt.Errorf("accountsFromFollows: error converting account to api account: %w", err)
			return nil, gtserror.NewErrorInternalError(err)
		}

		accounts = append(accounts, *account)
	}
	return accounts, nil
}

func (p *Processor) targetAccountsFromFollows(ctx context.Context, follows []*gtsmodel.Follow, requestingAccountID string) ([]apimodel.Account, gtserror.WithCode) {
	accounts := make([]apimodel.Account, 0, len(follows))
	for _, follow := range follows {
		if follow.TargetAccount == nil {
			// No account set for some reason; just skip.
			log.WithContext(ctx).WithField("follow", follow).Warn("follow had no associated target account")
			continue
		}

		if blocked, err := p.state.DB.IsEitherBlocked(ctx, requestingAccountID, follow.TargetAccountID); err != nil {
			err = fmt.Errorf("targetAccountsFromFollows: db error checking block: %w", err)
			return nil, gtserror.NewErrorInternalError(err)
		} else if blocked {
			continue
		}

		account, err := p.tc.AccountToAPIAccountPublic(ctx, follow.TargetAccount)
		if err != nil {
			err = fmt.Errorf("targetAccountsFromFollows: error converting account to api account: %w", err)
			return nil, gtserror.NewErrorInternalError(err)
		}

		accounts = append(accounts, *account)
	}
	return accounts, nil
}

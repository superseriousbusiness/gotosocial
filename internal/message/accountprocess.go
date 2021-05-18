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

package message

import (
	"errors"
	"fmt"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// accountCreate does the dirty work of making an account and user in the database.
// It then returns a token to the caller, for use with the new account, as per the
// spec here: https://docs.joinmastodon.org/methods/accounts/
func (p *processor) AccountCreate(authed *oauth.Auth, form *apimodel.AccountCreateRequest) (*apimodel.Token, error) {
	l := p.log.WithField("func", "accountCreate")

	if err := p.db.IsEmailAvailable(form.Email); err != nil {
		return nil, err
	}

	if err := p.db.IsUsernameAvailable(form.Username); err != nil {
		return nil, err
	}

	// don't store a reason if we don't require one
	reason := form.Reason
	if !p.config.AccountsConfig.ReasonRequired {
		reason = ""
	}

	l.Trace("creating new username and account")
	user, err := p.db.NewSignup(form.Username, reason, p.config.AccountsConfig.RequireApproval, form.Email, form.Password, form.IP, form.Locale, authed.Application.ID)
	if err != nil {
		return nil, fmt.Errorf("error creating new signup in the database: %s", err)
	}

	l.Tracef("generating a token for user %s with account %s and application %s", user.ID, user.AccountID, authed.Application.ID)
	accessToken, err := p.oauthServer.GenerateUserAccessToken(authed.Token, authed.Application.ClientSecret, user.ID)
	if err != nil {
		return nil, fmt.Errorf("error creating new access token for user %s: %s", user.ID, err)
	}

	return &apimodel.Token{
		AccessToken: accessToken.GetAccess(),
		TokenType:   "Bearer",
		Scope:       accessToken.GetScope(),
		CreatedAt:   accessToken.GetAccessCreateAt().Unix(),
	}, nil
}

func (p *processor) AccountGet(authed *oauth.Auth, targetAccountID string) (*apimodel.Account, error) {
	targetAccount := &gtsmodel.Account{}
	if err := p.db.GetByID(targetAccountID, targetAccount); err != nil {
		if _, ok := err.(db.ErrNoEntries); ok {
			return nil, errors.New("account not found")
		}
		return nil, fmt.Errorf("db error: %s", err)
	}

	// lazily dereference things on the account if it hasn't been done yet
	var requestingUsername string
	if authed.Account != nil {
		requestingUsername = authed.Account.Username
	}
	if err := p.dereferenceAccountFields(targetAccount, requestingUsername); err != nil {
		p.log.WithField("func", "AccountGet").Debugf("dereferencing account: %s", err)
	}

	var mastoAccount *apimodel.Account
	var err error
	if authed.Account != nil && targetAccount.ID == authed.Account.ID {
		mastoAccount, err = p.tc.AccountToMastoSensitive(targetAccount)
	} else {
		mastoAccount, err = p.tc.AccountToMastoPublic(targetAccount)
	}
	if err != nil {
		return nil, fmt.Errorf("error converting account: %s", err)
	}
	return mastoAccount, nil
}

func (p *processor) AccountUpdate(authed *oauth.Auth, form *apimodel.UpdateCredentialsRequest) (*apimodel.Account, error) {
	l := p.log.WithField("func", "AccountUpdate")

	if form.Discoverable != nil {
		if err := p.db.UpdateOneByID(authed.Account.ID, "discoverable", *form.Discoverable, &gtsmodel.Account{}); err != nil {
			return nil, fmt.Errorf("error updating discoverable: %s", err)
		}
	}

	if form.Bot != nil {
		if err := p.db.UpdateOneByID(authed.Account.ID, "bot", *form.Bot, &gtsmodel.Account{}); err != nil {
			return nil, fmt.Errorf("error updating bot: %s", err)
		}
	}

	if form.DisplayName != nil {
		if err := util.ValidateDisplayName(*form.DisplayName); err != nil {
			return nil, err
		}
		if err := p.db.UpdateOneByID(authed.Account.ID, "display_name", *form.DisplayName, &gtsmodel.Account{}); err != nil {
			return nil, err
		}
	}

	if form.Note != nil {
		if err := util.ValidateNote(*form.Note); err != nil {
			return nil, err
		}
		if err := p.db.UpdateOneByID(authed.Account.ID, "note", *form.Note, &gtsmodel.Account{}); err != nil {
			return nil, err
		}
	}

	if form.Avatar != nil && form.Avatar.Size != 0 {
		avatarInfo, err := p.updateAccountAvatar(form.Avatar, authed.Account.ID)
		if err != nil {
			return nil, err
		}
		l.Tracef("new avatar info for account %s is %+v", authed.Account.ID, avatarInfo)
	}

	if form.Header != nil && form.Header.Size != 0 {
		headerInfo, err := p.updateAccountHeader(form.Header, authed.Account.ID)
		if err != nil {
			return nil, err
		}
		l.Tracef("new header info for account %s is %+v", authed.Account.ID, headerInfo)
	}

	if form.Locked != nil {
		if err := p.db.UpdateOneByID(authed.Account.ID, "locked", *form.Locked, &gtsmodel.Account{}); err != nil {
			return nil, err
		}
	}

	if form.Source != nil {
		if form.Source.Language != nil {
			if err := util.ValidateLanguage(*form.Source.Language); err != nil {
				return nil, err
			}
			if err := p.db.UpdateOneByID(authed.Account.ID, "language", *form.Source.Language, &gtsmodel.Account{}); err != nil {
				return nil, err
			}
		}

		if form.Source.Sensitive != nil {
			if err := p.db.UpdateOneByID(authed.Account.ID, "locked", *form.Locked, &gtsmodel.Account{}); err != nil {
				return nil, err
			}
		}

		if form.Source.Privacy != nil {
			if err := util.ValidatePrivacy(*form.Source.Privacy); err != nil {
				return nil, err
			}
			if err := p.db.UpdateOneByID(authed.Account.ID, "privacy", *form.Source.Privacy, &gtsmodel.Account{}); err != nil {
				return nil, err
			}
		}
	}

	// fetch the account with all updated values set
	updatedAccount := &gtsmodel.Account{}
	if err := p.db.GetByID(authed.Account.ID, updatedAccount); err != nil {
		return nil, fmt.Errorf("could not fetch updated account %s: %s", authed.Account.ID, err)
	}

	acctSensitive, err := p.tc.AccountToMastoSensitive(updatedAccount)
	if err != nil {
		return nil, fmt.Errorf("could not convert account into mastosensitive account: %s", err)
	}
	return acctSensitive, nil
}

func (p *processor) AccountStatusesGet(authed *oauth.Auth, targetAccountID string, limit int, excludeReplies bool, maxID string, pinned bool, mediaOnly bool) ([]apimodel.Status, ErrorWithCode) {
	targetAccount := &gtsmodel.Account{}
	if err := p.db.GetByID(targetAccountID, targetAccount); err != nil {
		if _, ok := err.(db.ErrNoEntries); ok {
			return nil, NewErrorNotFound(fmt.Errorf("no entry found for account id %s", targetAccountID))
		}
		return nil, NewErrorInternalError(err)
	}

	statuses := []gtsmodel.Status{}
	apiStatuses := []apimodel.Status{}
	if err := p.db.GetStatusesByTimeDescending(targetAccountID, &statuses, limit, excludeReplies, maxID, pinned, mediaOnly); err != nil {
		if _, ok := err.(db.ErrNoEntries); ok {
			return apiStatuses, nil
		}
		return nil, NewErrorInternalError(err)
	}

	for _, s := range statuses {
		relevantAccounts, err := p.db.PullRelevantAccountsFromStatus(&s)
		if err != nil {
			return nil, NewErrorInternalError(fmt.Errorf("error getting relevant statuses: %s", err))
		}

		visible, err := p.db.StatusVisible(&s, targetAccount, authed.Account, relevantAccounts)
		if err != nil {
			return nil, NewErrorInternalError(fmt.Errorf("error checking status visibility: %s", err))
		}
		if !visible {
			continue
		}

		var boostedStatus *gtsmodel.Status
		if s.BoostOfID != "" {
			bs := &gtsmodel.Status{}
			if err := p.db.GetByID(s.BoostOfID, bs); err != nil {
				return nil, NewErrorInternalError(fmt.Errorf("error getting boosted status: %s", err))
			}
			boostedRelevantAccounts, err := p.db.PullRelevantAccountsFromStatus(bs)
			if err != nil {
				return nil, NewErrorInternalError(fmt.Errorf("error getting relevant accounts from boosted status: %s", err))
			}

			boostedVisible, err := p.db.StatusVisible(bs, relevantAccounts.BoostedAccount, authed.Account, boostedRelevantAccounts)
			if err != nil {
				return nil, NewErrorInternalError(fmt.Errorf("error checking boosted status visibility: %s", err))
			}

			if boostedVisible {
				boostedStatus = bs
			}
		}

		apiStatus, err := p.tc.StatusToMasto(&s, targetAccount, authed.Account, relevantAccounts.BoostedAccount, relevantAccounts.ReplyToAccount, boostedStatus)
		if err != nil {
			return nil, NewErrorInternalError(fmt.Errorf("error converting status to masto: %s", err))
		}

		apiStatuses = append(apiStatuses, *apiStatus)
	}

	return apiStatuses, nil
}

func (p *processor) AccountFollowersGet(authed *oauth.Auth, targetAccountID string) ([]apimodel.Account, ErrorWithCode) {
	blocked, err := p.db.Blocked(authed.Account.ID, targetAccountID)
	if err != nil {
		return nil, NewErrorInternalError(err)
	}

	if blocked {
		return nil, NewErrorNotFound(fmt.Errorf("block exists between accounts"))
	}

	followers := []gtsmodel.Follow{}
	accounts := []apimodel.Account{}
	if err := p.db.GetFollowersByAccountID(targetAccountID, &followers); err != nil {
		if _, ok := err.(db.ErrNoEntries); ok {
			return accounts, nil
		}
		return nil, NewErrorInternalError(err)
	}

	for _, f := range followers {
		blocked, err := p.db.Blocked(authed.Account.ID, f.AccountID)
		if err != nil {
			return nil, NewErrorInternalError(err)
		}
		if blocked {
			continue
		}

		a := &gtsmodel.Account{}
		if err := p.db.GetByID(f.AccountID, a); err != nil {
			if _, ok := err.(db.ErrNoEntries); ok {
				continue
			}
			return nil, NewErrorInternalError(err)
		}

		// derefence account fields in case we haven't done it already
		if err := p.dereferenceAccountFields(a, authed.Account.Username); err != nil {
			// don't bail if we can't fetch them, we'll try another time
			p.log.WithField("func", "AccountFollowersGet").Debugf("error dereferencing account fields: %s", err)
		}

		account, err := p.tc.AccountToMastoPublic(a)
		if err != nil {
			return nil, NewErrorInternalError(err)
		}
		accounts = append(accounts, *account)
	}
	return accounts, nil
}

func (p *processor) AccountRelationshipGet(authed *oauth.Auth, targetAccountID string) (*apimodel.Relationship, ErrorWithCode) {
	if authed == nil || authed.Account == nil {
		return nil, NewErrorForbidden(errors.New("not authed"))
	}

	gtsR, err := p.db.GetRelationship(authed.Account.ID, targetAccountID)
	if err != nil {
		return nil, NewErrorInternalError(fmt.Errorf("error getting relationship: %s", err))
	}

	r, err := p.tc.RelationshipToMasto(gtsR)
	if err != nil {
		return nil, NewErrorInternalError(fmt.Errorf("error converting relationship: %s", err))
	}

	return r, nil
}

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

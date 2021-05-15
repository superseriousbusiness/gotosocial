package message

import (
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

func (p *processor) FollowRequestsGet(auth *oauth.Auth) ([]apimodel.Account, ErrorWithCode) {
	frs := []gtsmodel.FollowRequest{}
	if err := p.db.GetFollowRequestsForAccountID(auth.Account.ID, &frs); err != nil {
		if _, ok := err.(db.ErrNoEntries); !ok {
			return nil, NewErrorInternalError(err)
		}
	}

	accts := []apimodel.Account{}
	for _, fr := range frs {
		acct := &gtsmodel.Account{}
		if err := p.db.GetByID(fr.AccountID, acct); err != nil {
			return nil, NewErrorInternalError(err)
		}
		mastoAcct, err := p.tc.AccountToMastoPublic(acct)
		if err != nil {
			return nil, NewErrorInternalError(err)
		}
		accts = append(accts, *mastoAcct)
	}
	return accts, nil
}

func (p *processor) FollowRequestAccept(auth *oauth.Auth, accountID string) ErrorWithCode {
	if err := p.db.AcceptFollowRequest(accountID, auth.Account.ID); err != nil {
		return NewErrorNotFound(err)
	}
	return nil
}

func (p *processor) FollowRequestDeny(auth *oauth.Auth) ErrorWithCode {
	return nil
}

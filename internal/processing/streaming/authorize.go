package streaming

import (
	"context"
	"fmt"

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (p *processor) AuthorizeStreamingRequest(ctx context.Context, accessToken string) (*gtsmodel.Account, error) {
	ti, err := p.oauthServer.LoadAccessToken(context.Background(), accessToken)
	if err != nil {
		return nil, fmt.Errorf("AuthorizeStreamingRequest: error loading access token: %s", err)
	}

	uid := ti.GetUserID()
	if uid == "" {
		return nil, fmt.Errorf("AuthorizeStreamingRequest: no userid in token")
	}

	// fetch user's and account for this user id
	user := &gtsmodel.User{}
	if err := p.db.GetByID(ctx, uid, user); err != nil || user == nil {
		return nil, fmt.Errorf("AuthorizeStreamingRequest: no user found for validated uid %s", uid)
	}

	acct := &gtsmodel.Account{}
	if err := p.db.GetByID(ctx, user.AccountID, acct); err != nil || acct == nil {
		return nil, fmt.Errorf("AuthorizeStreamingRequest: no account retrieved for user with id %s", uid)
	}

	return acct, nil
}

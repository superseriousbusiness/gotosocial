package message

import (
	"fmt"
	"net/http"

	"github.com/go-fed/activity/streams"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (p *processor) GetAPUser(requestedUsername string, request *http.Request) (interface{}, ErrorWithCode) {
	// get the account the request is referring to
	requestedAccount := &gtsmodel.Account{}
	if err := p.db.GetLocalAccountByUsername(requestedUsername, requestedAccount); err != nil {
		return nil, NewErrorNotFound(fmt.Errorf("database error getting account with username %s: %s", requestedUsername, err))
	}

	// authenticate the request
	requestingAccountURI, err := p.federator.AuthenticateFederatedRequest(requestedUsername, request)
	if err != nil {
		return nil, NewErrorNotAuthorized(err)
	}

	requestingAccount := &gtsmodel.Account{}
	err = p.db.GetWhere("uri", requestingAccountURI.String(), requestingAccount)
	if err != nil {
		if _, ok := err.(db.ErrNoEntries); !ok {
			// we don't have an entry for this account yet
			// what we do now should depend on our chosen federation method
			// for now though, we'll just dereference it
			// TODO: slow-fed
			requestingPerson, err := p.federator.DereferenceRemoteAccount(requestedUsername, requestingAccountURI)
			if err != nil {
				return nil, NewErrorInternalError(err)
			}
			requestedAccount, err = p.tc.ASPersonToAccount(requestingPerson)
			if err != nil {
				return nil, NewErrorInternalError(err)
			}
			if err := p.db.Put(requestingAccount); err != nil {
				return nil, NewErrorInternalError(err)
			}
		} else {
			// something has actually gone wrong
			return nil, NewErrorInternalError(err)
		}
	}

	blocked, err := p.db.Blocked(requestedAccount.ID, requestingAccount.ID)
	if err != nil {
		return nil, NewErrorInternalError(err)
	}

	if blocked {
		return nil, NewErrorNotAuthorized(fmt.Errorf("block exists between accounts %s and %s", requestedAccount.ID, requestingAccount.ID))
	}

	requestedPerson, err := p.tc.AccountToAS(requestedAccount)
	if err != nil {
		return nil, NewErrorInternalError(err)
	}

	data, err := streams.Serialize(requestedPerson)
	if err != nil {
		return nil, NewErrorInternalError(err)
	}

	return data, nil
}

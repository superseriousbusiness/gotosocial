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

package processing

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/go-fed/activity/streams"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// authenticateAndDereferenceFediRequest authenticates the HTTP signature of an incoming federation request, using the given
// username to perform the validation. It will *also* dereference the originator of the request and return it as a gtsmodel account
// for further processing. NOTE that this function will have the side effect of putting the dereferenced account into the database,
// and passing it into the processor through a channel for further asynchronous processing.
func (p *processor) authenticateAndDereferenceFediRequest(username string, r *http.Request) (*gtsmodel.Account, error) {

	// first authenticate
	requestingAccountURI, err := p.federator.AuthenticateFederatedRequest(username, r)
	if err != nil {
		return nil, fmt.Errorf("couldn't authenticate request for username %s: %s", username, err)
	}

	// OK now we can do the dereferencing part
	// we might already have an entry for this account so check that first
	requestingAccount := &gtsmodel.Account{}

	err = p.db.GetWhere([]db.Where{{Key: "uri", Value: requestingAccountURI.String()}}, requestingAccount)
	if err == nil {
		// we do have it yay, return it
		return requestingAccount, nil
	}

	if _, ok := err.(db.ErrNoEntries); !ok {
		// something has actually gone wrong so bail
		return nil, fmt.Errorf("database error getting account with uri %s: %s", requestingAccountURI.String(), err)
	}

	// we just don't have an entry for this account yet
	// what we do now should depend on our chosen federation method
	// for now though, we'll just dereference it
	// TODO: slow-fed
	requestingPerson, err := p.federator.DereferenceRemoteAccount(username, requestingAccountURI)
	if err != nil {
		return nil, fmt.Errorf("couldn't dereference %s: %s", requestingAccountURI.String(), err)
	}

	// convert it to our internal account representation
	requestingAccount, err = p.tc.ASRepresentationToAccount(requestingPerson, false)
	if err != nil {
		return nil, fmt.Errorf("couldn't convert dereferenced uri %s to gtsmodel account: %s", requestingAccountURI.String(), err)
	}

	// shove it in the database for later
	if err := p.db.Put(requestingAccount); err != nil {
		return nil, fmt.Errorf("database error inserting account with uri %s: %s", requestingAccountURI.String(), err)
	}

	// put it in our channel to queue it for async processing
	p.FromFederator() <- gtsmodel.FromFederator{
		APObjectType:   gtsmodel.ActivityStreamsProfile,
		APActivityType: gtsmodel.ActivityStreamsCreate,
		GTSModel:       requestingAccount,
	}

	return requestingAccount, nil
}

func (p *processor) GetFediUser(requestedUsername string, request *http.Request) (interface{}, ErrorWithCode) {
	// get the account the request is referring to
	requestedAccount := &gtsmodel.Account{}
	if err := p.db.GetLocalAccountByUsername(requestedUsername, requestedAccount); err != nil {
		return nil, NewErrorNotFound(fmt.Errorf("database error getting account with username %s: %s", requestedUsername, err))
	}

	// authenticate the request
	requestingAccount, err := p.authenticateAndDereferenceFediRequest(requestedUsername, request)
	if err != nil {
		return nil, NewErrorNotAuthorized(err)
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

func (p *processor) GetFediFollowers(requestedUsername string, request *http.Request) (interface{}, ErrorWithCode) {
	// get the account the request is referring to
	requestedAccount := &gtsmodel.Account{}
	if err := p.db.GetLocalAccountByUsername(requestedUsername, requestedAccount); err != nil {
		return nil, NewErrorNotFound(fmt.Errorf("database error getting account with username %s: %s", requestedUsername, err))
	}

	// authenticate the request
	requestingAccount, err := p.authenticateAndDereferenceFediRequest(requestedUsername, request)
	if err != nil {
		return nil, NewErrorNotAuthorized(err)
	}

	blocked, err := p.db.Blocked(requestedAccount.ID, requestingAccount.ID)
	if err != nil {
		return nil, NewErrorInternalError(err)
	}

	if blocked {
		return nil, NewErrorNotAuthorized(fmt.Errorf("block exists between accounts %s and %s", requestedAccount.ID, requestingAccount.ID))
	}

	requestedAccountURI, err := url.Parse(requestedAccount.URI)
	if err != nil {
		return nil, NewErrorInternalError(fmt.Errorf("error parsing url %s: %s", requestedAccount.URI, err))
	}

	requestedFollowers, err := p.federator.FederatingDB().Followers(context.Background(), requestedAccountURI)
	if err != nil {
		return nil, NewErrorInternalError(fmt.Errorf("error fetching followers for uri %s: %s", requestedAccountURI.String(), err))
	}

	data, err := streams.Serialize(requestedFollowers)
	if err != nil {
		return nil, NewErrorInternalError(err)
	}

	return data, nil
}

func (p *processor) GetFediFollowing(requestedUsername string, request *http.Request) (interface{}, ErrorWithCode) {
	// get the account the request is referring to
	requestedAccount := &gtsmodel.Account{}
	if err := p.db.GetLocalAccountByUsername(requestedUsername, requestedAccount); err != nil {
		return nil, NewErrorNotFound(fmt.Errorf("database error getting account with username %s: %s", requestedUsername, err))
	}

	// authenticate the request
	requestingAccount, err := p.authenticateAndDereferenceFediRequest(requestedUsername, request)
	if err != nil {
		return nil, NewErrorNotAuthorized(err)
	}

	blocked, err := p.db.Blocked(requestedAccount.ID, requestingAccount.ID)
	if err != nil {
		return nil, NewErrorInternalError(err)
	}

	if blocked {
		return nil, NewErrorNotAuthorized(fmt.Errorf("block exists between accounts %s and %s", requestedAccount.ID, requestingAccount.ID))
	}

	requestedAccountURI, err := url.Parse(requestedAccount.URI)
	if err != nil {
		return nil, NewErrorInternalError(fmt.Errorf("error parsing url %s: %s", requestedAccount.URI, err))
	}

	requestedFollowing, err := p.federator.FederatingDB().Following(context.Background(), requestedAccountURI)
	if err != nil {
		return nil, NewErrorInternalError(fmt.Errorf("error fetching following for uri %s: %s", requestedAccountURI.String(), err))
	}

	data, err := streams.Serialize(requestedFollowing)
	if err != nil {
		return nil, NewErrorInternalError(err)
	}

	return data, nil
}

func (p *processor) GetFediStatus(requestedUsername string, requestedStatusID string, request *http.Request) (interface{}, ErrorWithCode) {
	// get the account the request is referring to
	requestedAccount := &gtsmodel.Account{}
	if err := p.db.GetLocalAccountByUsername(requestedUsername, requestedAccount); err != nil {
		return nil, NewErrorNotFound(fmt.Errorf("database error getting account with username %s: %s", requestedUsername, err))
	}

	// authenticate the request
	requestingAccount, err := p.authenticateAndDereferenceFediRequest(requestedUsername, request)
	if err != nil {
		return nil, NewErrorNotAuthorized(err)
	}

	blocked, err := p.db.Blocked(requestedAccount.ID, requestingAccount.ID)
	if err != nil {
		return nil, NewErrorInternalError(err)
	}

	if blocked {
		return nil, NewErrorNotAuthorized(fmt.Errorf("block exists between accounts %s and %s", requestedAccount.ID, requestingAccount.ID))
	}

	s := &gtsmodel.Status{}
	if err := p.db.GetWhere([]db.Where{
		{Key: "id", Value: requestedStatusID},
		{Key: "account_id", Value: requestedAccount.ID},
	}, s); err != nil {
		return nil, NewErrorNotFound(fmt.Errorf("database error getting status with id %s and account id %s: %s", requestedStatusID, requestedAccount.ID, err))
	}

	asStatus, err := p.tc.StatusToAS(s)
	if err != nil {
		return nil, NewErrorInternalError(err)
	}

	data, err := streams.Serialize(asStatus)
	if err != nil {
		return nil, NewErrorInternalError(err)
	}

	return data, nil
}

func (p *processor) GetWebfingerAccount(requestedUsername string, request *http.Request) (*apimodel.WebfingerAccountResponse, ErrorWithCode) {
	// get the account the request is referring to
	requestedAccount := &gtsmodel.Account{}
	if err := p.db.GetLocalAccountByUsername(requestedUsername, requestedAccount); err != nil {
		return nil, NewErrorNotFound(fmt.Errorf("database error getting account with username %s: %s", requestedUsername, err))
	}

	// return the webfinger representation
	return &apimodel.WebfingerAccountResponse{
		Subject: fmt.Sprintf("acct:%s@%s", requestedAccount.Username, p.config.Host),
		Aliases: []string{
			requestedAccount.URI,
			requestedAccount.URL,
		},
		Links: []apimodel.WebfingerLink{
			{
				Rel:  "http://webfinger.net/rel/profile-page",
				Type: "text/html",
				Href: requestedAccount.URL,
			},
			{
				Rel:  "self",
				Type: "application/activity+json",
				Href: requestedAccount.URI,
			},
		},
	}, nil
}

func (p *processor) InboxPost(ctx context.Context, w http.ResponseWriter, r *http.Request) (bool, error) {
	contextWithChannel := context.WithValue(ctx, util.APFromFederatorChanKey, p.fromFederator)
	posted, err := p.federator.FederatingActor().PostInbox(contextWithChannel, w, r)
	return posted, err
}

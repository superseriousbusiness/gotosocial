/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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
	"errors"
	"fmt"
	"net/url"
	"strings"

	"codeberg.org/gruf/go-kv"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/federation/dereferencing"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/transport"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

func (p *processor) SearchGet(ctx context.Context, authed *oauth.Auth, search *apimodel.SearchQuery) (*apimodel.SearchResult, gtserror.WithCode) {
	l := log.WithFields(kv.Fields{
		{"query", search.Query},
	}...)

	// tidy up the query and make sure it wasn't just spaces
	query := strings.TrimSpace(search.Query)
	if query == "" {
		err := errors.New("search query was empty string after trimming space")
		return nil, gtserror.NewErrorBadRequest(err, err.Error())
	}

	searchResult := &apimodel.SearchResult{
		Accounts: []apimodel.Account{},
		Statuses: []apimodel.Status{},
		Hashtags: []apimodel.Tag{},
	}

	// currently the search will only ever return one result,
	// so return nothing if the offset is greater than 0
	if search.Offset > 0 {
		return searchResult, nil
	}

	foundAccounts := []*gtsmodel.Account{}
	foundStatuses := []*gtsmodel.Status{}

	var foundOne bool

	/*
		SEARCH BY MENTION
		check if the query is something like @whatever_username@example.org -- this means it's likely a remote account
	*/
	maybeNamestring := query
	if maybeNamestring[0] != '@' {
		maybeNamestring = "@" + maybeNamestring
	}

	if username, domain, err := util.ExtractNamestringParts(maybeNamestring); err == nil {
		l.Debugf("search term %s is a mention, looking it up...", maybeNamestring)
		foundAccount, err := p.searchAccountByMention(ctx, authed, username, domain, search.Resolve)
		if err != nil {
			l.Debugf("error looking up account %s: %s", maybeNamestring, err)
		} else {
			foundAccounts = append(foundAccounts, foundAccount)
			foundOne = true
			l.Debugf("got an account for %s by searching by mention", maybeNamestring)
		}
	}

	/*
		SEARCH BY URI
		check if the query is a URI with a recognizable scheme and dereference it
	*/
	if !foundOne {
		if uri, err := url.Parse(query); err == nil {
			if uri.Scheme == "https" || uri.Scheme == "http" {
				l.Debugf("search term %s is a uri, looking it up...", uri)

				// don't attempt to resolve (ie., dereference) local accounts/statuses
				resolve := search.Resolve
				if uri.Host == config.GetHost() || uri.Host == config.GetAccountDomain() {
					resolve = false
				}

				// check if it's a status or an account
				foundStatus, err := p.searchStatusByURI(ctx, authed, uri, resolve)
				if err != nil {
					log.Debugf("error searching status by uri %s: %s", uri, err)
				} else {
					foundStatuses = append(foundStatuses, foundStatus)
					foundOne = true
					l.Debug("got a status by searching by URI")
				}

				if !foundOne {
					foundAccount, err := p.searchAccountByURI(ctx, authed, uri, resolve)
					if err != nil {
						log.Debugf("error searching account by uri %s: %s", uri, err)
					} else {
						foundAccounts = append(foundAccounts, foundAccount)
						foundOne = true
						l.Debug("got an account by searching by URI")
					}
				}
			}
		}
	}

	if !foundOne {
		// we got nothing, we can return early
		log.Debugf("found nothing for query %s, returning", query)
		return searchResult, nil
	}

	/*
		FROM HERE ON we have our search results, it's just a matter of filtering them according to what this user is allowed to see,
		and then converting them into our frontend format.
	*/
	for _, foundAccount := range foundAccounts {
		// make sure there's no block in either direction between the account and the requester
		if blocked, err := p.db.IsBlocked(ctx, authed.Account.ID, foundAccount.ID, true); err == nil && !blocked {
			// all good, convert it and add it to the results
			if apiAcct, err := p.tc.AccountToAPIAccountPublic(ctx, foundAccount); err == nil && apiAcct != nil {
				searchResult.Accounts = append(searchResult.Accounts, *apiAcct)
			}
		}
	}

	for _, foundStatus := range foundStatuses {
		if visible, err := p.filter.StatusVisible(ctx, foundStatus, authed.Account); !visible || err != nil {
			continue
		}

		apiStatus, err := p.tc.StatusToAPIStatus(ctx, foundStatus, authed.Account)
		if err != nil {
			continue
		}

		searchResult.Statuses = append(searchResult.Statuses, *apiStatus)
	}

	return searchResult, nil
}

func (p *processor) searchStatusByURI(ctx context.Context, authed *oauth.Auth, uri *url.URL, resolve bool) (*gtsmodel.Status, error) {
	status, statusable, err := p.federator.GetStatus(ctx, authed.Account.Username, uri, true, true)
	if err != nil {
		return nil, fmt.Errorf("searchStatusByURI: error fetching status %s: %v", uri, err)
	}

	if !*status.Local && statusable != nil {
		// Attempt to dereference the status thread while we are here
		p.federator.DereferenceRemoteThread(transport.WithFastfail(ctx), authed.Account.Username, uri, status, statusable)
	}

	return status, nil
}

func (p *processor) searchAccountByURI(ctx context.Context, authed *oauth.Auth, uri *url.URL, resolve bool) (*gtsmodel.Account, error) {
	account, err := p.federator.GetAccount(transport.WithFastfail(ctx), dereferencing.GetAccountParams{
		RequestingUsername: authed.Account.Username,
		RemoteAccountID:    uri,
		Blocking:           true,
		SkipResolve:        !resolve,
	})

	if err != nil {
		return nil, fmt.Errorf("searchAccountByURI: error calling %s, and resolve is false", uri)
	}

	return account, nil
}

func (p *processor) searchAccountByMention(ctx context.Context, authed *oauth.Auth, username string, domain string, resolve bool) (*gtsmodel.Account, error) {
	// if it's a local account we can skip a whole bunch of stuff
	if domain == config.GetHost() || domain == config.GetAccountDomain() || domain == "" {
		maybeAcct, err := p.db.GetAccountByUsernameDomain(ctx, username, "")
		if err != nil {
			return nil, fmt.Errorf("searchAccountByMention: error getting local account by username: %s", err)
		}
		return maybeAcct, nil
	}

	// we don't have it yet, try to find it remotely
	return p.federator.GetAccount(transport.WithFastfail(ctx), dereferencing.GetAccountParams{
		RequestingUsername:    authed.Account.Username,
		RemoteAccountUsername: username,
		RemoteAccountHost:     domain,
		Blocking:              true,
		SkipResolve:           !resolve,
	})
}

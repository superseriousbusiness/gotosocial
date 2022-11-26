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
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/federation/dereferencing"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/transport"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// Implementation note: in this function, we tend to log errors
// at debug level rather than return them. This is because the
// search has a sort of fallthrough logic: if we can't get a result
// with x search, we should try with y search rather than returning.
//
// If we get to the end and still haven't found anything, even then
// we shouldn't return an error, just return an empty search result.
//
// The only exception to this is when we get a malformed query, in
// which case we return a bad request error so the user knows they
// did something funky.
func (p *processor) SearchGet(ctx context.Context, authed *oauth.Auth, search *apimodel.SearchQuery) (*apimodel.SearchResult, gtserror.WithCode) {
	// tidy up the query and make sure it wasn't just spaces
	query := strings.TrimSpace(search.Query)
	if query == "" {
		err := errors.New("search query was empty string after trimming space")
		return nil, gtserror.NewErrorBadRequest(err, err.Error())
	}

	l := log.WithFields(kv.Fields{{"query", query}}...)

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
		l.Trace("search term is a mention, looking it up...")
		foundAccount, err := p.searchAccountByMention(ctx, authed, username, domain, search.Resolve)
		if err != nil {
			// if we got a namestring, and we can't get an account for it, we should
			// return already, because there's no point going any further + trying to
			// parse this as a URL or whatever
			l.Debugf("error looking up account: %s", err)
			return searchResult, nil
		}

		foundAccounts = append(foundAccounts, foundAccount)
		foundOne = true
		l.Trace("got an account by searching by mention")
	}

	/*
		SEARCH BY URI
		check if the query is a URI with a recognizable scheme and dereference it
	*/
	if !foundOne {
		if uri, err := url.Parse(query); err == nil {
			if uri.Scheme == "https" || uri.Scheme == "http" {
				l.Trace("search term is a uri, looking it up...")

				// don't attempt to resolve (ie., dereference) local accounts/statuses
				resolve := search.Resolve
				if uri.Host == config.GetHost() || uri.Host == config.GetAccountDomain() {
					resolve = false
				}

				// check if it's a status...
				foundStatus, err := p.searchStatusByURI(ctx, authed, uri, resolve)
				if err != nil {
					if !errors.Is(err, db.ErrNoEntries) {
						l.Debugf("error searching status by uri: %s", err)
					}
				} else {
					foundStatuses = append(foundStatuses, foundStatus)
					foundOne = true
					l.Trace("got a status by searching by URI")
				}

				// ... or an account
				if !foundOne {
					foundAccount, err := p.searchAccountByURI(ctx, authed, uri, resolve)
					if err != nil {
						if !errors.Is(err, db.ErrNoEntries) {
							l.Debugf("error searching account by uri %s: %s", uri, err)
						}
					} else {
						foundAccounts = append(foundAccounts, foundAccount)
						foundOne = true
						l.Trace("got an account by searching by URI")
					}
				}
			}
		}
	}

	if !foundOne {
		// we got nothing, we can return early
		l.Trace("found nothing, returning")
		return searchResult, nil
	}

	/*
		FROM HERE ON we have our search results, it's just a matter of filtering them according to what this user is allowed to see,
		and then converting them into our frontend format.
	*/
	for _, foundAccount := range foundAccounts {
		// make sure there's no block in either direction between the account and the requester
		blocked, err := p.db.IsBlocked(ctx, authed.Account.ID, foundAccount.ID, true)
		if err != nil {
			l.Debugf("error checking block between %s and %s: %s", authed.Account.ID, foundAccount.ID, err)
			continue
		}

		if blocked {
			l.Tracef("block exists between %s and %s, skipping this result", authed.Account.ID, foundAccount.ID)
			continue
		}

		apiAcct, err := p.tc.AccountToAPIAccountPublic(ctx, foundAccount)
		if err != nil {
			l.Debugf("error converting account %s to api account: %s", foundAccount.ID, err)
			continue
		}

		searchResult.Accounts = append(searchResult.Accounts, *apiAcct)
	}

	for _, foundStatus := range foundStatuses {
		// make sure each found status is visible to the requester
		visible, err := p.filter.StatusVisible(ctx, foundStatus, authed.Account)
		if err != nil {
			l.Debugf("error checking visibility of status %s for account %s: %s", foundStatus.ID, authed.Account.ID, err)
			continue
		}

		if !visible {
			l.Tracef("status %s is not visible to account %s, skipping this result", foundStatus.ID, authed.Account.ID)
			continue
		}

		apiStatus, err := p.tc.StatusToAPIStatus(ctx, foundStatus, authed.Account)
		if err != nil {
			l.Debugf("error converting status %s to api status: %s", foundStatus.ID, err)
			continue
		}

		searchResult.Statuses = append(searchResult.Statuses, *apiStatus)
	}

	return searchResult, nil
}

func (p *processor) searchStatusByURI(ctx context.Context, authed *oauth.Auth, uri *url.URL, resolve bool) (*gtsmodel.Status, error) {
	status, statusable, err := p.federator.GetStatus(transport.WithFastfail(ctx), authed.Account.Username, uri, true, true)
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

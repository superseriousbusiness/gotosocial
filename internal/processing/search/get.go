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

package search

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"net/url"
	"strings"

	"codeberg.org/gruf/go-kv"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/federation/dereferencing"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

var noResults = &apimodel.SearchResult{
	Accounts: []*apimodel.Account{},
	Statuses: []*apimodel.Status{},
	Hashtags: []*apimodel.Tag{},
}

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
func (p *Processor) Get(ctx context.Context, account *gtsmodel.Account, searchQuery *apimodel.SearchQuery) (*apimodel.SearchResult, gtserror.WithCode) {
	// Normalize query.
	query := strings.TrimSpace(searchQuery.Query)
	if query == "" {
		err := errors.New("search query was empty string after trimming space")
		return nil, gtserror.NewErrorBadRequest(err, err.Error())
	}

	log.
		WithContext(ctx).
		WithFields(kv.Fields{
			{"query", query},
		}...).
		Debugf("beginning search with query %s", query)

	// Currently the search will only ever return one result,
	// so return nothing if the offset is greater than 0.
	if searchQuery.Offset > 0 {
		return noResults, nil
	}

	var (
		foundStatuses = []*gtsmodel.Status{}
		foundAccounts = []*gtsmodel.Account{}
		appendStatus  = func(foundStatus *gtsmodel.Status) { foundStatuses = append(foundStatuses, foundStatus) }
		appendAccount = func(foundAccount *gtsmodel.Account) { foundAccounts = append(foundAccounts, foundAccount) }
		keepLooking   bool
		err           error
	)

	// Check if the query is something like '@whatever_user' or '@whatever_user@somewhere.com'.
	keepLooking, err = p.searchByNamestring(ctx, account, query, searchQuery.Resolve, appendAccount)
	if err != nil {
		err = fmt.Errorf("error searching by namestring: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if !keepLooking {
		// Return whatever we have.
		return p.packageSearchResponse(
			ctx,
			account,
			foundAccounts,
			foundStatuses,
		)
	}

	// Check if the query is a URI with a recognizable scheme and use it to look for accounts or statuses.
	keepLooking, err = p.searchByURI(ctx, account, query, searchQuery.Resolve, appendAccount, appendStatus)
	if err != nil {
		err = fmt.Errorf("error searching by URI: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if !keepLooking {
		// Return whatever we have.
		return p.packageSearchResponse(
			ctx,
			account,
			foundAccounts,
			foundStatuses,
		)
	}

	// Search for accounts and statuses using the query as arbitrary text.
	keepLooking, err = p.searchByText(ctx, account, query, appendAccount, appendStatus)
	if err != nil {
		err = fmt.Errorf("error searching by text: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if !keepLooking {
		// Return whatever we have.
		return p.packageSearchResponse(
			ctx,
			account,
			foundAccounts,
			foundStatuses,
		)
	}

	return noResults, nil
}

func (p *Processor) searchByNamestring(
	ctx context.Context,
	requestingAccount *gtsmodel.Account,
	query string,
	resolve bool,
	appendAccount func(*gtsmodel.Account),
) (bool, error) {
	if !strings.Contains(query, "@") {
		// If there's no '@' in the query at all, we know
		// for sure that it can't be a namestring. Caller
		// should keep looking with another search method.
		return true, nil
	}

	if query[0] != '@' {
		// There's an '@' somewhere in the query, but it's
		// not the first character, so likely this is a
		// slightly malformed namestring which looks like
		// an email address (eg., someone@example.org).
		if _, err := mail.ParseAddress(query); err != nil {
			// If we can't parse this is as an email address,
			// there's not much we can do. No need to return
			// error though; caller should just keep looking
			// with another search method.
			return true, nil //nolint:nilerr
		}

		// Normalize query by just prepending '@'; we'll
		// end up with something like @someone@example.org
		query = "@" + query
	}

	// See if we ended up with something that looks
	// like @test_user or @test_user@example.org.
	username, domain, err := util.ExtractNamestringParts(query)
	if err != nil {
		// No need to return error; just not a namestring
		// we can search with. Caller should keep looking
		// with another search method.
		return true, nil //nolint:nilerr
	}

	// Domain may be empty, which is fine, but
	// don't search for an empty domain block.
	if domain != "" {
		blocked, err := p.state.DB.IsDomainBlocked(ctx, domain)
		if err != nil {
			err = fmt.Errorf("error checking domain block: %w", err)
			return false, gtserror.NewErrorInternalError(err)
		}

		if blocked {
			// Don't search for blocked domains.
			// Caller should stop looking.
			return false, nil
		}
	}

	// Check if username + domain points to an account.
	foundAccount, err := p.searchAccountByUsernameDomain(ctx, requestingAccount, username, domain, resolve)
	if err != nil {
		// Check for semi-expected error types.
		// On one of these, we can continue.
		var (
			errNotRetrievable = new(*dereferencing.ErrNotRetrievable) // Item can't be dereferenced.
			errWrongType      = new(*ap.ErrWrongType)                 // Item was dereferenced, but wasn't an account.
		)

		if !errors.As(err, errNotRetrievable) && !errors.As(err, errWrongType) {
			err = fmt.Errorf("error looking up %s as account: %w", query, err)
			return false, gtserror.NewErrorInternalError(err)
		}
	} else {
		appendAccount(foundAccount) // Hit!
	}

	// Regardless of whether we have a hit, return false
	// to indicate caller should stop looking; namestrings
	// are a pretty specific format so it's unlikely the
	// caller was looking for something other than an account.
	return false, nil
}

// searchAccountByUsernameDomain looks for one account with the given
// username and domain. If domain is empty, or equal to our domain,
//
// Will return either a hit, an ErrNotRetrievable, an ErrWrongType,
// or a real error that the caller should handle.
func (p *Processor) searchAccountByUsernameDomain(
	ctx context.Context,
	requestingAccount *gtsmodel.Account,
	username string,
	domain string,
	resolve bool,
) (*gtsmodel.Account, error) {
	if resolve {
		// We're allowed to resolve, leave the
		// rest up to the dereferencer functions.
		account, _, err := p.federator.GetAccountByUsernameDomain(
			gtscontext.SetFastFail(ctx),
			requestingAccount.Username,
			username, domain,
		)

		return account, err
	}

	// We're not allowed to resolve; search database only.
	var usernameDomain string
	if domain == "" || domain == config.GetHost() || domain == config.GetAccountDomain() {
		// Local lookup, normalize domain.
		domain = ""
		usernameDomain = username
	} else {
		// Remote lookup.
		usernameDomain = username + "@" + domain
	}

	// Search the database for existing account with USERNAME@DOMAIN
	account, err := p.state.DB.GetAccountByUsernameDomain(ctx, username, domain)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err = fmt.Errorf("error checking database for account %s: %w", usernameDomain, err)
		return nil, err
	}

	if account != nil {
		// We got a hit! No need to continue.
		return account, nil
	}

	err = fmt.Errorf("account %s could not be retrieved locally and we cannot resolve", usernameDomain)
	return nil, dereferencing.NewErrNotRetrievable(err)
}

// searchByURI looks for account(s) or a status with the given URI
// set as either its URL or ActivityPub URI. If it gets hits, it
// will call the provided append functions to return results.
//
// The boolean return value indicates to the caller whether the
// search should continue (true) or stop (false). False will be
// returned in cases where a hit has been found, the domain of the
// searched URI is blocked, or an unrecoverable error has occurred.
func (p *Processor) searchByURI(
	ctx context.Context,
	requestingAccount *gtsmodel.Account,
	query string,
	resolve bool,
	appendAccount func(*gtsmodel.Account),
	appendStatus func(*gtsmodel.Status),
) (bool, error) {
	uri, err := url.Parse(query)
	if err != nil {
		// No need to return error; just not a URI
		// we can search with. Caller should keep
		// looking with another search method.
		return true, nil //nolint:nilerr
	}

	if !(uri.Scheme == "https" || uri.Scheme == "http") {
		// This might just be a weirdly-parsed URI,
		// since Go's url package tends to be a bit
		// trigger-happy when deciding things are URIs.
		// Indicate caller should keep looking.
		return true, nil
	}

	blocked, err := p.state.DB.IsURIBlocked(ctx, uri)
	if err != nil {
		err = fmt.Errorf("error checking domain block: %w", err)
		return false, gtserror.NewErrorInternalError(err)
	}

	if blocked {
		// Don't search for blocked domains.
		// Caller should stop looking.
		return false, nil
	}

	// Check if URI points to an account.
	foundAccount, err := p.searchAccountByURI(ctx, requestingAccount, uri, resolve)
	if err != nil {
		// Check for semi-expected error types.
		// On one of these, we can continue.
		var (
			errNotRetrievable = new(*dereferencing.ErrNotRetrievable) // Item can't be dereferenced.
			errWrongType      = new(*ap.ErrWrongType)                 // Item was dereferenced, but wasn't an account.
		)

		if !errors.As(err, errNotRetrievable) && !errors.As(err, errWrongType) {
			err = fmt.Errorf("error looking up %s as account: %w", uri, err)
			return false, gtserror.NewErrorInternalError(err)
		}
	} else {
		// Hit; return false to indicate caller should
		// stop looking, since it's extremely unlikely
		// a status and an account will have the same URL.
		appendAccount(foundAccount)
		return false, nil
	}

	// Check if URI points to a status.
	foundStatus, err := p.searchStatusByURI(ctx, requestingAccount, uri, resolve)
	if err != nil {
		// Check for semi-expected error types.
		// On one of these, we can continue.
		var (
			errNotRetrievable = new(*dereferencing.ErrNotRetrievable) // Item can't be dereferenced.
			errWrongType      = new(*ap.ErrWrongType)                 // Item was dereferenced, but wasn't a status.
		)

		if !errors.As(err, errNotRetrievable) && !errors.As(err, errWrongType) {
			err = fmt.Errorf("error looking up %s as status: %w", uri, err)
			return false, gtserror.NewErrorInternalError(err)
		}
	} else {
		// Hit; return false to indicate caller should
		// stop looking, since it's extremely unlikely
		// a status and an account will have the same URL.
		appendStatus(foundStatus)
		return false, nil
	}

	// No errors, but no hits either; since this
	// was a URI, caller should stop looking.
	return false, nil
}

// searchAccountByURI looks for one account with the given URI.
// Will return either a hit, an ErrNotRetrievable, an ErrWrongType,
// or a real error that the caller should handle.
func (p *Processor) searchAccountByURI(
	ctx context.Context,
	requestingAccount *gtsmodel.Account,
	uri *url.URL,
	resolve bool,
) (*gtsmodel.Account, error) {
	if resolve {
		// We're allowed to resolve, leave the
		// rest up to the dereferencer functions.
		account, _, err := p.federator.GetAccountByURI(
			gtscontext.SetFastFail(ctx),
			requestingAccount.Username,
			uri,
		)

		return account, err
	}

	// We're not allowed to resolve; search database only.
	uriStr := uri.String()

	// Search by ActivityPub URI.
	account, err := p.state.DB.GetAccountByURI(ctx, uriStr)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err = fmt.Errorf("error checking database for account using URI %s: %w", uriStr, err)
		return nil, err
	}

	if account != nil {
		// We got a hit! No need to continue.
		return account, nil
	}

	// No hit yet. Fallback to try by URL.
	account, err = p.state.DB.GetAccountByURL(ctx, uriStr)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err = fmt.Errorf("error checking database for account using URL %s: %w", uriStr, err)
		return nil, err
	}

	if account != nil {
		// We got a hit! No need to continue.
		return account, nil
	}

	err = fmt.Errorf("account %s could not be retrieved locally and we cannot resolve", uriStr)
	return nil, dereferencing.NewErrNotRetrievable(err)
}

// searchStatusByURI looks for one status with the given URI.
// Will return either a hit, an ErrNotRetrievable, an ErrWrongType,
// or a real error that the caller should handle.
func (p *Processor) searchStatusByURI(
	ctx context.Context,
	requestingAccount *gtsmodel.Account,
	uri *url.URL,
	resolve bool,
) (*gtsmodel.Status, error) {
	if resolve {
		// We're allowed to resolve, leave the
		// rest up to the dereferencer functions.
		status, _, err := p.federator.GetStatusByURI(
			gtscontext.SetFastFail(ctx),
			requestingAccount.Username,
			uri,
		)

		return status, err
	}

	// We're not allowed to resolve; search database only.
	uriStr := uri.String()

	// Search by ActivityPub URI.
	status, err := p.state.DB.GetStatusByURI(ctx, uriStr)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err = fmt.Errorf("error checking database for status using URI %s: %w", uriStr, err)
		return nil, err
	}

	if status != nil {
		// We got a hit! No need to continue.
		return status, nil
	}

	// No hit yet. Fallback to try by URL.
	status, err = p.state.DB.GetStatusByURL(ctx, uriStr)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err = fmt.Errorf("error checking database for status using URL %s: %w", uriStr, err)
		return nil, err
	}

	if status != nil {
		// We got a hit! No need to continue.
		return status, nil
	}

	err = fmt.Errorf("status %s could not be retrieved locally and we cannot resolve", uriStr)
	return nil, dereferencing.NewErrNotRetrievable(err)
}

func (p *Processor) searchByText(
	ctx context.Context,
	requestingAccount *gtsmodel.Account,
	query string,
	appendAccount func(*gtsmodel.Account),
	appendStatus func(*gtsmodel.Status),
) (bool, error) {
	// Search for accounts using the given text.
	return true, nil
}

func (p *Processor) packageSearchResponse(
	ctx context.Context,
	requestingAccount *gtsmodel.Account,
	accounts []*gtsmodel.Account,
	statuses []*gtsmodel.Status,
) (*apimodel.SearchResult, gtserror.WithCode) {
	result := &apimodel.SearchResult{
		Accounts: make([]*apimodel.Account, 0, len(accounts)),
		Statuses: make([]*apimodel.Status, 0, len(statuses)),
	}

	for _, account := range accounts {
		// Ensure requester can see result account.
		visible, err := p.filter.AccountVisible(ctx, requestingAccount, account)
		if err != nil {
			err = fmt.Errorf("error checking visibility of account %s for account %s: %w", account.ID, requestingAccount.ID, err)
			return nil, gtserror.NewErrorInternalError(err)
		}

		if !visible {
			log.Debugf(ctx, "account %s is not visible to account %s, skipping this result", account.ID, requestingAccount.ID)
			continue
		}

		apiAccount, err := p.tc.AccountToAPIAccountPublic(ctx, account)
		if err != nil {
			log.Debugf(ctx, "skipping account %s because it couldn't be converted to its api representation: %s", account.ID, err)
			continue
		}

		result.Accounts = append(result.Accounts, apiAccount)
	}

	for _, status := range statuses {
		// Ensure requester can see result status.
		visible, err := p.filter.StatusVisible(ctx, requestingAccount, status)
		if err != nil {
			err = fmt.Errorf("error checking visibility of status %s for account %s: %w", status.ID, requestingAccount.ID, err)
			return nil, gtserror.NewErrorInternalError(err)
		}

		if !visible {
			log.Debugf(ctx, "status %s is not visible to account %s, skipping this result", status.ID, requestingAccount.ID)
			continue
		}

		apiStatus, err := p.tc.StatusToAPIStatus(ctx, status, requestingAccount)
		if err != nil {
			log.Debugf(ctx, "skipping status %s because it couldn't be converted to its api representation: %s", status.ID, err)
			continue
		}

		result.Statuses = append(result.Statuses, apiStatus)
	}

	return result, nil
}

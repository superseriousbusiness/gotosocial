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
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

const (
	queryTypeAny      = ""
	queryTypeAccounts = "accounts"
	queryTypeStatuses = "statuses"
	queryTypeHashtags = "hashtags"
)

// Get performs a search for accounts and/or statuses using the
// provided request parameters.
//
// Implementation note: in this function, we try to only return
// an error to the caller they've submitted a bad request, or when
// a serious error has occurred. This is because the search has a
// sort of fallthrough logic: if we can't get a result with one
// type of search, we should proceed with y search rather than
// returning an early error.
//
// If we get to the end and still haven't found anything, even
// then we shouldn't return an error, just return an empty result.
func (p *Processor) Get(
	ctx context.Context,
	account *gtsmodel.Account,
	req *apimodel.SearchRequest,
) (*apimodel.SearchResult, gtserror.WithCode) {
	var (
		maxID     = req.MaxID
		minID     = req.MinID
		limit     = req.Limit
		offset    = req.Offset
		query     = strings.TrimSpace(req.Query)                      // Trim trailing/leading whitespace.
		queryType = strings.TrimSpace(strings.ToLower(req.QueryType)) // Trim trailing/leading whitespace; convert to lowercase.
		resolve   = req.Resolve
		following = req.Following
	)

	// Validate query.
	if query == "" {
		err := errors.New("search query was empty string after trimming space")
		return nil, gtserror.NewErrorBadRequest(err, err.Error())
	}

	// Validate query type.
	switch queryType {
	case queryTypeAny, queryTypeAccounts, queryTypeStatuses, queryTypeHashtags:
		// No problem.
	default:
		err := fmt.Errorf(
			"search query type %s was not recognized, valid options are ['%s', '%s', '%s', '%s']",
			queryType, queryTypeAny, queryTypeAccounts, queryTypeStatuses, queryTypeHashtags,
		)
		return nil, gtserror.NewErrorBadRequest(err, err.Error())
	}

	log.
		WithContext(ctx).
		WithFields(kv.Fields{
			{"maxID", maxID},
			{"minID", minID},
			{"limit", limit},
			{"offset", offset},
			{"query", query},
			{"queryType", queryType},
			{"resolve", resolve},
			{"following", following},
		}...).
		Debugf("beginning search")

	// todo: Currently we don't support offset for paging;
	// a caller can page using maxID or minID, but if they
	// supply an offset greater than 0, return nothing as
	// though there were no additional results.
	if req.Offset > 0 {
		return p.packageSearchResult(ctx, account, nil, nil)
	}

	var (
		foundStatuses = make([]*gtsmodel.Status, 0, limit)
		foundAccounts = make([]*gtsmodel.Account, 0, limit)
		appendStatus  = func(foundStatus *gtsmodel.Status) { foundStatuses = append(foundStatuses, foundStatus) }
		appendAccount = func(foundAccount *gtsmodel.Account) { foundAccounts = append(foundAccounts, foundAccount) }
		keepLooking   bool
		err           error
	)

	// Only try to search by namestring if search type includes
	// accounts, since this is all namestring search can return.
	if includeAccounts(queryType) {
		// Copy query to avoid altering original.
		queryC := query

		// If query looks vaguely like an email address, ie. it doesn't
		// start with '@' but it has '@' in it somewhere, it's probably
		// a poorly-formed namestring. Be generous and correct for this.
		if strings.Contains(queryC, "@") && queryC[0] != '@' {
			if _, err := mail.ParseAddress(queryC); err == nil {
				// Yep, really does look like
				// an email address! Be nice.
				queryC = "@" + queryC
			}
		}

		// Search using what may or may not be a namestring.
		keepLooking, err = p.accountsByNamestring(
			ctx,
			account,
			maxID,
			minID,
			limit,
			offset,
			queryC,
			resolve,
			following,
			appendAccount,
		)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			err = gtserror.Newf("error searching by namestring: %w", err)
			return nil, gtserror.NewErrorInternalError(err)
		}

		if !keepLooking {
			// Return whatever we have.
			return p.packageSearchResult(
				ctx,
				account,
				foundAccounts,
				foundStatuses,
			)
		}
	}

	// Check if the query is a URI with a recognizable
	// scheme and use it to look for accounts or statuses.
	keepLooking, err = p.byURI(
		ctx,
		account,
		query,
		queryType,
		resolve,
		appendAccount,
		appendStatus,
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err = gtserror.Newf("error searching by URI: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if !keepLooking {
		// Return whatever we have.
		return p.packageSearchResult(
			ctx,
			account,
			foundAccounts,
			foundStatuses,
		)
	}

	// As a last resort, search for accounts and
	// statuses using the query as arbitrary text.
	if err := p.byText(
		ctx,
		account,
		maxID,
		minID,
		limit,
		offset,
		query,
		queryType,
		following,
		appendAccount,
		appendStatus,
	); err != nil && !errors.Is(err, db.ErrNoEntries) {
		err = gtserror.Newf("error searching by text: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Return whatever we ended
	// up with (could be nothing).
	return p.packageSearchResult(
		ctx,
		account,
		foundAccounts,
		foundStatuses,
	)
}

// accountsByNamestring searches for accounts using the
// provided namestring query. If domain is not set in
// the namestring, it may return more than one result
// by doing a text search in the database for accounts
// matching the query. Otherwise, it tries to return an
// exact match.
func (p *Processor) accountsByNamestring(
	ctx context.Context,
	requestingAccount *gtsmodel.Account,
	maxID string,
	minID string,
	limit int,
	offset int,
	query string,
	resolve bool,
	following bool,
	appendAccount func(*gtsmodel.Account),
) (bool, error) {
	// See if we have something that looks like a namestring.
	username, domain, err := util.ExtractNamestringParts(query)
	if err != nil {
		// No need to return error; just not a namestring
		// we can search with. Caller should keep looking
		// with another search method.
		return true, nil //nolint:nilerr
	}

	if domain == "" {
		// No error, but no domain set. That means the query
		// looked like '@someone' which is not an exact search.
		// Try to search for any accounts that match the query
		// string, and let the caller know they should stop.
		return false, p.accountsByText(
			ctx,
			requestingAccount.ID,
			maxID,
			minID,
			limit,
			offset,
			// OK to assume username is set now. Use
			// it instead of query to omit leading '@'.
			username,
			following,
			appendAccount,
		)
	}

	// No error, and domain and username were both set.
	// Caller is likely trying to search for an exact
	// match, from either a remote instance or local.
	foundAccount, err := p.accountByUsernameDomain(
		ctx,
		requestingAccount,
		username,
		domain,
		resolve,
	)
	if err != nil {
		// Check for semi-expected error types.
		// On one of these, we can continue.
		if !gtserror.Unretrievable(err) && !gtserror.WrongType(err) {
			err = gtserror.Newf("error looking up %s as account: %w", query, err)
			return false, gtserror.NewErrorInternalError(err)
		}
	} else {
		appendAccount(foundAccount)
	}

	// Regardless of whether we have a hit at this point,
	// return false to indicate caller should stop looking;
	// namestrings are a very specific format so it's unlikely
	// the caller was looking for something other than an account.
	return false, nil
}

// accountByUsernameDomain looks for one account with the given
// username and domain. If domain is empty, or equal to our domain,
// search will be confined to local accounts.
//
// Will return either a hit, an ErrNotRetrievable, an ErrWrongType,
// or a real error that the caller should handle.
func (p *Processor) accountByUsernameDomain(
	ctx context.Context,
	requestingAccount *gtsmodel.Account,
	username string,
	domain string,
	resolve bool,
) (*gtsmodel.Account, error) {
	var usernameDomain string
	if domain == "" || domain == config.GetHost() || domain == config.GetAccountDomain() {
		// Local lookup, normalize domain.
		domain = ""
		usernameDomain = username
	} else {
		// Remote lookup.
		usernameDomain = username + "@" + domain

		// Ensure domain not blocked.
		blocked, err := p.state.DB.IsDomainBlocked(ctx, domain)
		if err != nil {
			err = gtserror.Newf("error checking domain block: %w", err)
			return nil, gtserror.NewErrorInternalError(err)
		} else if blocked {
			// Don't search on blocked domain.
			err = gtserror.New("domain blocked")
			return nil, gtserror.SetUnretrievable(err)
		}
	}

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

	// We're not allowed to resolve. Search the database
	// for existing account with given username + domain.
	account, err := p.state.DB.GetAccountByUsernameDomain(ctx, username, domain)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err = gtserror.Newf("error checking database for account %s: %w", usernameDomain, err)
		return nil, err
	}

	if account != nil {
		// We got a hit! No need to continue.
		return account, nil
	}

	err = fmt.Errorf("account %s could not be retrieved locally and we cannot resolve", usernameDomain)
	return nil, gtserror.SetUnretrievable(err)
}

// byURI looks for account(s) or a status with the given URI
// set as either its URL or ActivityPub URI. If it gets hits, it
// will call the provided append functions to return results.
//
// The boolean return value indicates to the caller whether the
// search should continue (true) or stop (false). False will be
// returned in cases where a hit has been found, the domain of the
// searched URI is blocked, or an unrecoverable error has occurred.
func (p *Processor) byURI(
	ctx context.Context,
	requestingAccount *gtsmodel.Account,
	query string,
	queryType string,
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
		err = gtserror.Newf("error checking domain block: %w", err)
		return false, gtserror.NewErrorInternalError(err)
	}

	if blocked {
		// Don't search for blocked domains.
		// Caller should stop looking.
		return false, nil
	}

	if includeAccounts(queryType) {
		// Check if URI points to an account.
		foundAccount, err := p.accountByURI(ctx, requestingAccount, uri, resolve)
		if err != nil {
			// Check for semi-expected error types.
			// On one of these, we can continue.
			if !gtserror.Unretrievable(err) && !gtserror.WrongType(err) {
				err = gtserror.Newf("error looking up %s as account: %w", uri, err)
				return false, gtserror.NewErrorInternalError(err)
			}
		} else {
			// Hit; return false to indicate caller should
			// stop looking, since it's extremely unlikely
			// a status and an account will have the same URL.
			appendAccount(foundAccount)
			return false, nil
		}
	}

	if includeStatuses(queryType) {
		// Check if URI points to a status.
		foundStatus, err := p.statusByURI(ctx, requestingAccount, uri, resolve)
		if err != nil {
			// Check for semi-expected error types.
			// On one of these, we can continue.
			if !gtserror.Unretrievable(err) && !gtserror.WrongType(err) {
				err = gtserror.Newf("error looking up %s as status: %w", uri, err)
				return false, gtserror.NewErrorInternalError(err)
			}
		} else {
			// Hit; return false to indicate caller should
			// stop looking, since it's extremely unlikely
			// a status and an account will have the same URL.
			appendStatus(foundStatus)
			return false, nil
		}
	}

	// No errors, but no hits either; since this
	// was a URI, caller should stop looking.
	return false, nil
}

// accountByURI looks for one account with the given URI.
// If resolve is false, it will only look in the database.
// If resolve is true, it will try to resolve the account
// from remote using the URI, if necessary.
//
// Will return either a hit, ErrNotRetrievable, ErrWrongType,
// or a real error that the caller should handle.
func (p *Processor) accountByURI(
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
	uriStr := uri.String() // stringify uri just once

	// Search by ActivityPub URI.
	account, err := p.state.DB.GetAccountByURI(ctx, uriStr)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err = gtserror.Newf("error checking database for account using URI %s: %w", uriStr, err)
		return nil, err
	}

	if account != nil {
		// We got a hit! No need to continue.
		return account, nil
	}

	// No hit yet. Fallback to try by URL.
	account, err = p.state.DB.GetAccountByURL(ctx, uriStr)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err = gtserror.Newf("error checking database for account using URL %s: %w", uriStr, err)
		return nil, err
	}

	if account != nil {
		// We got a hit! No need to continue.
		return account, nil
	}

	err = fmt.Errorf("account %s could not be retrieved locally and we cannot resolve", uriStr)
	return nil, gtserror.SetUnretrievable(err)
}

// statusByURI looks for one status with the given URI.
// If resolve is false, it will only look in the database.
// If resolve is true, it will try to resolve the status
// from remote using the URI, if necessary.
//
// Will return either a hit, ErrNotRetrievable, ErrWrongType,
// or a real error that the caller should handle.
func (p *Processor) statusByURI(
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
	uriStr := uri.String() // stringify uri just once

	// Search by ActivityPub URI.
	status, err := p.state.DB.GetStatusByURI(ctx, uriStr)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err = gtserror.Newf("error checking database for status using URI %s: %w", uriStr, err)
		return nil, err
	}

	if status != nil {
		// We got a hit! No need to continue.
		return status, nil
	}

	// No hit yet. Fallback to try by URL.
	status, err = p.state.DB.GetStatusByURL(ctx, uriStr)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err = gtserror.Newf("error checking database for status using URL %s: %w", uriStr, err)
		return nil, err
	}

	if status != nil {
		// We got a hit! No need to continue.
		return status, nil
	}

	err = fmt.Errorf("status %s could not be retrieved locally and we cannot resolve", uriStr)
	return nil, gtserror.SetUnretrievable(err)
}

// byText searches in the database for accounts and/or
// statuses containing the given query string, using
// the provided parameters.
//
// If queryType is any (empty string), both accounts
// and statuses will be searched, else only the given
// queryType of item will be returned.
func (p *Processor) byText(
	ctx context.Context,
	requestingAccount *gtsmodel.Account,
	maxID string,
	minID string,
	limit int,
	offset int,
	query string,
	queryType string,
	following bool,
	appendAccount func(*gtsmodel.Account),
	appendStatus func(*gtsmodel.Status),
) error {
	if queryType == queryTypeAny {
		// If search type is any, ignore maxID and minID
		// parameters, since we can't use them to page
		// on both accounts and statuses simultaneously.
		maxID = ""
		minID = ""
	}

	if includeAccounts(queryType) {
		// Search for accounts using the given text.
		if err := p.accountsByText(ctx,
			requestingAccount.ID,
			maxID,
			minID,
			limit,
			offset,
			query,
			following,
			appendAccount,
		); err != nil {
			return err
		}
	}

	if includeStatuses(queryType) {
		// Search for statuses using the given text.
		if err := p.statusesByText(ctx,
			requestingAccount.ID,
			maxID,
			minID,
			limit,
			offset,
			query,
			appendStatus,
		); err != nil {
			return err
		}
	}

	return nil
}

// accountsByText searches in the database for limit
// number of accounts using the given query text.
func (p *Processor) accountsByText(
	ctx context.Context,
	requestingAccountID string,
	maxID string,
	minID string,
	limit int,
	offset int,
	query string,
	following bool,
	appendAccount func(*gtsmodel.Account),
) error {
	accounts, err := p.state.DB.SearchForAccounts(
		ctx,
		requestingAccountID,
		query, maxID, minID, limit, following, offset)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return gtserror.Newf("error checking database for accounts using text %s: %w", query, err)
	}

	for _, account := range accounts {
		appendAccount(account)
	}

	return nil
}

// statusesByText searches in the database for limit
// number of statuses using the given query text.
func (p *Processor) statusesByText(
	ctx context.Context,
	requestingAccountID string,
	maxID string,
	minID string,
	limit int,
	offset int,
	query string,
	appendStatus func(*gtsmodel.Status),
) error {
	statuses, err := p.state.DB.SearchForStatuses(
		ctx,
		requestingAccountID,
		query, maxID, minID, limit, offset)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return gtserror.Newf("error checking database for statuses using text %s: %w", query, err)
	}

	for _, status := range statuses {
		appendStatus(status)
	}

	return nil
}

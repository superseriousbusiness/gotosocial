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
	"github.com/superseriousbusiness/gotosocial/internal/text"
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

		// Include instance accounts in the first
		// parts of this search. This will be
		// changed to 'false' when doing text
		// search in the database in the latter
		// parts of this function.
		includeInstanceAccounts = true

		// Assume caller doesn't want to see
		// blocked accounts. This will change
		// to 'true' if caller is searching
		// for a specific account.
		includeBlockedAccounts = false
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
		return p.packageSearchResult(
			ctx,
			account,
			nil, nil, nil, // No results.
			req.APIv1,
			includeInstanceAccounts,
			includeBlockedAccounts,
		)
	}

	var (
		foundStatuses = make([]*gtsmodel.Status, 0, limit)
		foundAccounts = make([]*gtsmodel.Account, 0, limit)
		foundTags     = make([]*gtsmodel.Tag, 0, limit)
		appendStatus  = func(s *gtsmodel.Status) { foundStatuses = append(foundStatuses, s) }
		appendAccount = func(a *gtsmodel.Account) { foundAccounts = append(foundAccounts, a) }
		appendTag     = func(t *gtsmodel.Tag) { foundTags = append(foundTags, t) }
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

		// See if we have something that looks like a namestring.
		username, domain, err := util.ExtractNamestringParts(queryC)
		if err == nil {
			// We managed to parse query as a namestring.
			// If domain was set, this is a very specific
			// search for a particular account, so show
			// that account to the caller even if they
			// have it blocked. They might be looking
			// for it to unblock it again!
			domainSet := (domain != "")
			includeBlockedAccounts = domainSet

			err = p.accountsByNamestring(
				ctx,
				account,
				maxID,
				minID,
				limit,
				offset,
				username,
				domain,
				resolve,
				following,
				appendAccount,
			)
			if err != nil && !errors.Is(err, db.ErrNoEntries) {
				err = gtserror.Newf("error searching by namestring: %w", err)
				return nil, gtserror.NewErrorInternalError(err)
			}

			// If domain was set, we know this is
			// a full namestring, and not a url or
			// just a username, so we should stop
			// looking now and just return what we
			// have (if anything). Otherwise we'll
			// let the search keep going.
			if domainSet {
				return p.packageSearchResult(
					ctx,
					account,
					foundAccounts,
					foundStatuses,
					foundTags,
					req.APIv1,
					includeInstanceAccounts,
					includeBlockedAccounts,
				)
			}
		}
	}

	// Check if we're searching by a known URI scheme.
	// (This might just be a weirdly-parsed URI,
	// since Go's url package tends to be a bit
	// trigger-happy when deciding things are URIs).
	uri, err := url.Parse(query)
	if err == nil && (uri.Scheme == "https" || uri.Scheme == "http") {
		// URI is pretty specific so we can safely assume
		// caller wants to include blocked accounts too.
		includeBlockedAccounts = true

		if err := p.byURI(
			ctx,
			account,
			uri,
			queryType,
			resolve,
			appendAccount,
			appendStatus,
		); err != nil && !errors.Is(err, db.ErrNoEntries) {
			err = gtserror.Newf("error searching by URI: %w", err)
			return nil, gtserror.NewErrorInternalError(err)
		}

		// This was a URI, so at this point just return
		// whatever we have. You can't search hashtags by
		// URI, and shouldn't do full-text with a URI either.
		return p.packageSearchResult(
			ctx,
			account,
			foundAccounts,
			foundStatuses,
			foundTags,
			req.APIv1,
			includeInstanceAccounts,
			includeBlockedAccounts,
		)
	}

	// If query looks like a hashtag (ie., starts
	// with '#'), then search for tags.
	//
	// Since '#' is a very unique prefix and isn't
	// shared among account or status searches, we
	// can save a bit of time by searching for this
	// now, and bailing quickly if we get no results,
	// or we're not allowed to include hashtags in
	// search results.
	//
	// We know that none of the subsequent searches
	// would show any good results either, and those
	// searches are *much* more expensive.
	keepLooking, err := p.hashtag(
		ctx,
		maxID,
		minID,
		limit,
		offset,
		query,
		queryType,
		appendTag,
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err = gtserror.Newf("error searching for hashtag: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if !keepLooking {
		// Return whatever we have.
		return p.packageSearchResult(
			ctx,
			account,
			foundAccounts,
			foundStatuses,
			foundTags,
			req.APIv1,
			includeInstanceAccounts,
			includeBlockedAccounts,
		)
	}

	// As a last resort, search for accounts and
	// statuses using the query as arbitrary text.
	//
	// At this point we no longer want to include
	// instance accounts in the results, since searching
	// for something like 'mastodon', for example, will
	// include a million instance/service accounts that
	// have 'mastodon' in the domain, and therefore in
	// the username, making the search results useless.
	includeInstanceAccounts = false
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
		foundTags,
		req.APIv1,
		includeInstanceAccounts,
		includeBlockedAccounts,
	)
}

// accountsByNamestring searches for accounts using the
// provided username and domain. If domain is not set,
// it may return more than one result by doing a text
// search in the database for accounts matching the query.
// Otherwise, it tries to return an exact match.
func (p *Processor) accountsByNamestring(
	ctx context.Context,
	requestingAccount *gtsmodel.Account,
	maxID string,
	minID string,
	limit int,
	offset int,
	username string,
	domain string,
	resolve bool,
	following bool,
	appendAccount func(*gtsmodel.Account),
) error {
	if domain == "" {
		// No error, but no domain set. That means the query
		// looked like '@someone' which is not an exact search.
		// Try to search for any accounts that match the query
		// string, and let the caller know they should stop.
		return p.accountsByText(
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
			err = gtserror.Newf("error looking up @%s@%s as account: %w", username, domain, err)
			return gtserror.NewErrorInternalError(err)
		}
	} else {
		appendAccount(foundAccount)
	}

	return nil
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
	uri *url.URL,
	queryType string,
	resolve bool,
	appendAccount func(*gtsmodel.Account),
	appendStatus func(*gtsmodel.Status),
) error {
	blocked, err := p.state.DB.IsURIBlocked(ctx, uri)
	if err != nil {
		err = gtserror.Newf("error checking domain block: %w", err)
		return gtserror.NewErrorInternalError(err)
	}

	if blocked {
		// Don't search for
		// blocked domains.
		return nil
	}

	if includeAccounts(queryType) {
		// Check if URI points to an account.
		foundAccount, err := p.accountByURI(ctx, requestingAccount, uri, resolve)
		if err != nil {
			// Check for semi-expected error types.
			// On one of these, we can continue.
			if !gtserror.Unretrievable(err) && !gtserror.WrongType(err) {
				err = gtserror.Newf("error looking up %s as account: %w", uri, err)
				return gtserror.NewErrorInternalError(err)
			}
		} else {
			// Hit! Return early since it's extremely unlikely
			// a status and an account will have the same URL.
			appendAccount(foundAccount)
			return nil
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
				return gtserror.NewErrorInternalError(err)
			}
		} else {
			// Hit! Return early since it's extremely unlikely
			// a status and an account will have the same URL.
			appendStatus(foundStatus)
			return nil
		}
	}

	// No errors, but no hits
	// either; that's fine.
	return nil
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

func (p *Processor) hashtag(
	ctx context.Context,
	maxID string,
	minID string,
	limit int,
	offset int,
	query string,
	queryType string,
	appendTag func(*gtsmodel.Tag),
) (bool, error) {
	if query[0] != '#' {
		// Query doesn't look like a hashtag,
		// but if we're being instructed to
		// look explicitly *only* for hashtags,
		// let's be generous and assume caller
		// just left out the hash prefix.

		if queryType != queryTypeHashtags {
			// Nope, search isn't explicitly
			// for hashtags, keep looking.
			return true, nil
		}

		// Search is explicitly for
		// tags, let this one through.
	} else if !includeHashtags(queryType) {
		// Query looks like a hashtag,
		// but we're not meant to include
		// hashtags in the results.
		//
		// Indicate to caller they should
		// stop looking, since they're not
		// going to get results for this by
		// looking in any other way.
		return false, nil
	}

	// Query looks like a hashtag, and we're allowed
	// to search for hashtags.
	//
	// Ensure this is a valid tag for our instance.
	normalized, ok := text.NormalizeHashtag(query)
	if !ok {
		// Couldn't normalize/not a
		// valid hashtag after all.
		// Caller should stop looking.
		return false, nil
	}

	// Search for tags starting with the normalized string.
	tags, err := p.state.DB.SearchForTags(
		ctx,
		normalized,
		maxID,
		minID,
		limit,
		offset,
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf(
			"error checking database for tags using text %s: %w",
			normalized, err,
		)
		return false, err
	}

	// Return whatever we got.
	for _, tag := range tags {
		appendTag(tag)
	}

	return false, nil
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

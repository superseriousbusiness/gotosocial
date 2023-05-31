package search

import (
	"context"
	"errors"
	"strings"

	"codeberg.org/gruf/go-kv"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/federation/dereferencing"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

func (p *Processor) SearchAccounts(
	ctx context.Context,
	requestingAccount *gtsmodel.Account,
	query string,
	limit int,
	offset int,
	resolve bool,
	following bool,
) ([]*apimodel.Account, gtserror.WithCode) {
	var (
		foundAccounts = make([]*gtsmodel.Account, 0, limit)
		appendAccount = func(foundAccount *gtsmodel.Account) { foundAccounts = append(foundAccounts, foundAccount) }
	)

	// Validate query.
	query = strings.TrimSpace(query)
	if query == "" {
		err := gtserror.New("search query was empty string after trimming space")
		return nil, gtserror.NewErrorBadRequest(err, err.Error())
	}

	log.
		WithContext(ctx).
		WithFields(kv.Fields{
			{"limit", limit},
			{"offset", offset},
			{"query", query},
			{"resolve", resolve},
			{"following", following},
		}...).
		Debugf("beginning search")

	// todo: Currently we don't support offset for paging;
	// if they supply an offset greater than 0, return nothing
	// as though there were no additional results.
	if offset > 0 {
		return p.packageAccounts(ctx, requestingAccount, foundAccounts)
	}

	// See if the caller supplied a namestring. If they did,
	// we can either look it up locally or try to resolve it.
	username, domain, _ := util.ExtractNamestringParts(query)
	switch {

	case domain != "":
		// Namestring with a defined domain; ensure not blocked.
		blocked, err := p.state.DB.IsDomainBlocked(ctx, domain)
		if err != nil {
			err = gtserror.Newf("error checking domain block: %w", err)
			return nil, gtserror.NewErrorInternalError(err)
		}

		if blocked {
			// Don't search for blocked domains.
			return p.packageAccounts(ctx, requestingAccount, foundAccounts)
		}

		// If domain was defined, username must be
		// as well, so we can safely fall through.
		fallthrough

	case username != "":
		// Check if username (+ domain) points to
		// an account; resolve if we're allowed.
		foundAccount, err := p.searchAccountByUsernameDomain(
			ctx,
			requestingAccount,
			username,
			domain,
			resolve,
		)
		if err != nil {
			// Check for semi-expected error types.
			// On one of these, we can continue.
			var (
				errNotRetrievable = new(*dereferencing.ErrNotRetrievable) // Item can't be dereferenced.
				errWrongType      = new(*ap.ErrWrongType)                 // Item was dereferenced, but wasn't an account.
			)

			if !errors.As(err, errNotRetrievable) && !errors.As(err, errWrongType) {
				err = gtserror.Newf("error looking up %s as account: %w", query, err)
				return nil, gtserror.NewErrorInternalError(err)
			}
		} else {
			appendAccount(foundAccount)
		}

	default:
		// Not a namestring. Do a text
		// search for accounts instead.
		if err := p.searchByText(
			ctx,
			requestingAccount,
			id.Highest,
			id.Lowest,
			limit,
			offset,
			query,
			queryTypeAccounts,
			following,
			appendAccount,
			nil,
		); err != nil {
			err = gtserror.Newf("error searching by text: %w", err)
			return nil, gtserror.NewErrorInternalError(err)
		}
	}

	return p.packageAccounts(ctx, requestingAccount, foundAccounts)
}

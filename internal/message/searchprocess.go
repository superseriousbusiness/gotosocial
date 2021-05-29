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

package message

import (
	"errors"
	"net/url"
	"strings"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

func (p *processor) SearchGet(authed *oauth.Auth, searchQuery *apimodel.SearchQuery) (*apimodel.SearchResult, ErrorWithCode) {
	results := &apimodel.SearchResult{
		Accounts: []apimodel.Account{},
		Statuses: []apimodel.Status{},
		Hashtags: []apimodel.Tag{},
	}
	foundAccounts := []*gtsmodel.Account{}
	foundStatuses := []*gtsmodel.Status{}
	// foundHashtags := []*gtsmodel.Tag{}

	// convert the query to lowercase and trim leading/trailing spaces
	query := strings.ToLower(strings.TrimSpace(searchQuery.Query))

	// check if the query is a URI and just do a lookup for that, straight up
	if uri, err := url.Parse(query); err == nil {
		// 1. check if it's a status
		foundStatus, err := p.searchStatusByURI(authed, uri, searchQuery.Resolve)
		if err != nil {
			return nil, NewErrorInternalError(err)
		}
		if foundStatus != nil {
			foundStatuses = append(foundStatuses, foundStatus)
		}

		// 2. check if it's an account
		foundAccount, err := p.searchAccountByURI(authed, uri, searchQuery.Resolve)
		if err != nil {
			return nil, NewErrorInternalError(err)
		}
		if foundAccount != nil {
			foundAccounts = append(foundAccounts, foundAccount)
		}
	}

	// check if the query is something like @whatever_username@example.org -- this means it's a remote account
	if util.IsMention(searchQuery.Query) {
		foundAccount, err := p.searchAccountByMention(authed, searchQuery.Query, searchQuery.Resolve)
		if err != nil {
			return nil, NewErrorInternalError(err)
		}
		if foundAccount != nil {
			foundAccounts = append(foundAccounts, foundAccount)
		}
	}

	/*
		FROM HERE ON we have our search results, it's just a matter of filtering them according to what this user is allowed to see,
		and then converting them into our frontend format.
	*/
	for _, foundAccount := range foundAccounts {
		// make sure there's no block in either direction between the account and the requester
		if blocked, err := p.db.Blocked(authed.Account.ID, foundAccount.ID); err == nil && !blocked {
			// all good, convert it and add it to the results
			acctMasto, err := p.tc.AccountToMastoPublic(foundAccount)
			if err != nil {
				return nil, NewErrorInternalError(err)
			}
			results.Accounts = append(results.Accounts, *acctMasto)
		}
	}

	for _, foundStatus := range foundStatuses {
		statusOwner := &gtsmodel.Account{}
		if err := p.db.GetByID(foundStatus.AccountID, statusOwner); err != nil {
			continue
		}

		relevantAccounts, err := p.db.PullRelevantAccountsFromStatus(foundStatus)
		if err != nil {
			continue
		}
		if visible, err := p.db.StatusVisible(foundStatus, statusOwner, authed.Account, relevantAccounts); !visible || err != nil {
			continue
		}

		statusMasto, err := p.tc.StatusToMasto(foundStatus, statusOwner, authed.Account, relevantAccounts.BoostedAccount, relevantAccounts.ReplyToAccount, nil)
		if err != nil {
			continue
		}

		results.Statuses = append(results.Statuses, *statusMasto)
	}

	return results, nil
}

func (p *processor) searchStatusByURI(authed *oauth.Auth, uri *url.URL, resolve bool) (foundStatus *gtsmodel.Status, err error) {
	// 1. check if it's a status
	maybeStatus := &gtsmodel.Status{}
	if err = p.db.GetWhere([]db.Where{{Key: "uri", Value: uri.String(), CaseInsensitive: true}}, maybeStatus); err == nil {
		// we have it and it's a status
		foundStatus = maybeStatus
		return
	} else if err = p.db.GetWhere([]db.Where{{Key: "url", Value: uri.String(), CaseInsensitive: true}}, maybeStatus); err == nil {
		// we have it and it's a status
		foundStatus = maybeStatus
		return
	}

	// we don't have it locally so dereference it if we're allowed to
	if resolve {
		statusable, err := p.federator.DereferenceRemoteStatus(authed.Account.Username, uri)
		if err == nil {
			// it IS a status!

			// extract the status owner's IRI from the statusable
			var statusOwnerURI *url.URL
			statusAttributedTo := statusable.GetActivityStreamsAttributedTo()
			for i := statusAttributedTo.Begin(); i != statusAttributedTo.End(); i = i.Next() {
				if i.IsIRI() {
					statusOwnerURI = i.GetIRI()
					break
				}
			}
			if statusOwnerURI == nil {
				return nil, NewErrorInternalError(errors.New("couldn't extract ownerAccountURI from statusable"))
			}

			// make sure the status owner exists in the db by searching for it
			_, err := p.searchAccountByURI(authed, statusOwnerURI, resolve)
			if err != nil {
				return nil, err
			}

			// we have the status owner, we have the dereferenced status, so now we should finish dereferencing the status properly

			// first turn it into a gtsmodel.Status
			status, err := p.tc.ASStatusToStatus(statusable)
			if err != nil {
				return nil, NewErrorInternalError(err)
			}

			// put it in the DB so it gets a UUID
			if err := p.db.Put(status); err != nil {
				return nil, NewErrorInternalError(err)
			}

			// properly dereference everything in the status (media attachments etc)
			if err := p.dereferenceStatusFields(status, authed.Account.Username); err != nil {
				return nil, NewErrorInternalError(err)
			}

			// update with the nicely dereferenced status
			if err := p.db.UpdateByID(status.ID, status); err != nil {
				return nil, NewErrorInternalError(err)
			}

			foundStatus = status
		}
	}
	return
}

func (p *processor) searchAccountByURI(authed *oauth.Auth, uri *url.URL, resolve bool) (foundAccount *gtsmodel.Account, err error) {
	maybeAccount := &gtsmodel.Account{}
	if err = p.db.GetWhere([]db.Where{{Key: "uri", Value: uri.String(), CaseInsensitive: true}}, maybeAccount); err == nil {
		// we have it and it's an account
		foundAccount = maybeAccount
		return
	} else if err = p.db.GetWhere([]db.Where{{Key: "url", Value: uri.String(), CaseInsensitive: true}}, maybeAccount); err == nil {
		// we have it and it's an account
		foundAccount = maybeAccount
		return
	}
	if resolve {
		// we don't have it locally so try and dereference it
		accountable, err := p.federator.DereferenceRemoteAccount(authed.Account.Username, uri)
		if err == nil {
			// it IS an account!
			account, err := p.tc.ASRepresentationToAccount(accountable, false)
			if err != nil {
				return nil, NewErrorInternalError(err)
			}

			if err := p.db.Put(account); err != nil {
				return nil, NewErrorInternalError(err)
			}

			if err := p.dereferenceAccountFields(account, authed.Account.Username, false); err != nil {
				return nil, NewErrorInternalError(err)
			}

			foundAccount = account
		}
	}
	return
}

func (p *processor) searchAccountByMention(authed *oauth.Auth, mention string, resolve bool) (foundAccount *gtsmodel.Account, err error) {
	// query is for a remote account
	username, domain, err := util.ExtractMentionParts(mention)
	if err != nil {
		return nil, NewErrorBadRequest(err)
	}

	// if it's a local account we can skip a whole bunch of stuff
	maybeAcct := &gtsmodel.Account{}
	if domain == p.config.Host {
		if err = p.db.GetLocalAccountByUsername(username, maybeAcct); err != nil {
			return
		}
		foundAccount = maybeAcct
		return
	}

	// it's not a local account so first we'll check if it's in the database already...
	where := []db.Where{
		{Key: "username", Value: username, CaseInsensitive: true},
		{Key: "domain", Value: domain, CaseInsensitive: true},
	}
	err = p.db.GetWhere(where, maybeAcct)
	if err == nil {
		// we've got it stored locally already!
		foundAccount = maybeAcct
		return
	}

	if _, ok := err.(db.ErrNoEntries); !ok {
		// if it's  not errNoEntries there's been a real database error so bail at this point
		return nil, NewErrorInternalError(err)
	}

	// we got a db.ErrNoEntries, so we just don't have the account locally stored -- check if we can dereference it
	if resolve {
		// we're allowed to resolve it so let's try

		// first we need to webfinger the remote account to convert the username and domain into the activitypub URI for the account
		acctURI, err := p.federator.FingerRemoteAccount(authed.Account.Username, username, domain)
		if err != nil {
			// something went wrong doing the webfinger lookup so we can't process the request
			return nil, NewErrorInternalError(err)
		}

		// dereference the account based on the URI we retrieved from the webfinger lookup
		accountable, err := p.federator.DereferenceRemoteAccount(authed.Account.Username, acctURI)
		if err != nil {
			// something went wrong doing the dereferencing so we can't process the request
			return nil, NewErrorInternalError(err)
		}

		// convert the dereferenced account to the gts model of that account
		foundAccount, err = p.tc.ASRepresentationToAccount(accountable, false)
		if err != nil {
			// something went wrong doing the conversion to a gtsmodel.Account so we can't process the request
			return nil, NewErrorInternalError(err)
		}

		// put this new account in our database
		if err := p.db.Put(foundAccount); err != nil {
			return nil, NewErrorInternalError(err)
		}

		// properly dereference all the fields on the account immediately
		if err := p.dereferenceAccountFields(foundAccount, authed.Account.Username, true); err != nil {
			return nil, NewErrorInternalError(err)
		}
	}

	return
}

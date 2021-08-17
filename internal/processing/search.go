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
	"fmt"
	"net/url"
	"strings"

	"github.com/sirupsen/logrus"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

func (p *processor) SearchGet(authed *oauth.Auth, searchQuery *apimodel.SearchQuery) (*apimodel.SearchResult, gtserror.WithCode) {
	l := p.log.WithFields(logrus.Fields{
		"func":  "SearchGet",
		"query": searchQuery.Query,
	})

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

	var foundOne bool
	// check if the query is something like @whatever_username@example.org -- this means it's a remote account
	if !foundOne && util.IsMention(searchQuery.Query) {
		l.Debug("search term is a mention, looking it up...")
		foundAccount, err := p.searchAccountByMention(authed, searchQuery.Query, searchQuery.Resolve)
		if err == nil && foundAccount != nil {
			foundAccounts = append(foundAccounts, foundAccount)
			foundOne = true
			l.Debug("got an account by searching by mention")
		}
	}

	// check if the query is a URI and just do a lookup for that, straight up
	if uri, err := url.Parse(query); err == nil && !foundOne {
		// 1. check if it's a status
		if foundStatus, err := p.searchStatusByURI(authed, uri, searchQuery.Resolve); err == nil && foundStatus != nil {
			foundStatuses = append(foundStatuses, foundStatus)
			foundOne = true
			l.Debug("got a status by searching by URI")
		}

		// 2. check if it's an account
		if foundAccount, err := p.searchAccountByURI(authed, uri, searchQuery.Resolve); err == nil && foundAccount != nil {
			foundAccounts = append(foundAccounts, foundAccount)
			foundOne = true
			l.Debug("got an account by searching by URI")
		}
	}

	if !foundOne {
		// we haven't found anything yet so search for text now
		l.Debug("nothing found by mention or by URI, will fall back to searching by text now")
	}

	/*
		FROM HERE ON we have our search results, it's just a matter of filtering them according to what this user is allowed to see,
		and then converting them into our frontend format.
	*/
	for _, foundAccount := range foundAccounts {
		// make sure there's no block in either direction between the account and the requester
		if blocked, err := p.db.Blocked(authed.Account.ID, foundAccount.ID, true); err == nil && !blocked {
			// all good, convert it and add it to the results
			if acctMasto, err := p.tc.AccountToMastoPublic(foundAccount); err == nil && acctMasto != nil {
				results.Accounts = append(results.Accounts, *acctMasto)
			}
		}
	}

	for _, foundStatus := range foundStatuses {
		statusOwner := &gtsmodel.Account{}
		if err := p.db.GetByID(foundStatus.AccountID, statusOwner); err != nil {
			continue
		}

		if visible, err := p.filter.StatusVisible(foundStatus, authed.Account); !visible || err != nil {
			continue
		}

		statusMasto, err := p.tc.StatusToMasto(foundStatus, authed.Account)
		if err != nil {
			continue
		}

		results.Statuses = append(results.Statuses, *statusMasto)
	}

	return results, nil
}

func (p *processor) searchStatusByURI(authed *oauth.Auth, uri *url.URL, resolve bool) (*gtsmodel.Status, error) {
	l := p.log.WithFields(logrus.Fields{
		"func":    "searchStatusByURI",
		"uri":     uri.String(),
		"resolve": resolve,
	})

	maybeStatus := &gtsmodel.Status{}
	if err := p.db.GetWhere([]db.Where{{Key: "uri", Value: uri.String(), CaseInsensitive: true}}, maybeStatus); err == nil {
		// we have it and it's a status
		return maybeStatus, nil
	} else if err := p.db.GetWhere([]db.Where{{Key: "url", Value: uri.String(), CaseInsensitive: true}}, maybeStatus); err == nil {
		// we have it and it's a status
		return maybeStatus, nil
	}

	// we don't have it locally so dereference it if we're allowed to
	if resolve {
		status, _, _, err := p.federator.GetRemoteStatus(authed.Account.Username, uri, true)
		if err == nil {
			if err := p.federator.DereferenceRemoteThread(authed.Account.Username, uri); err != nil {
				// try to deref the thread while we're here
				l.Debugf("searchStatusByURI: error dereferencing remote thread: %s", err)
			}
			return status, nil
		}
	}
	return nil, nil
}

func (p *processor) searchAccountByURI(authed *oauth.Auth, uri *url.URL, resolve bool) (*gtsmodel.Account, error) {
	maybeAccount := &gtsmodel.Account{}
	if err := p.db.GetWhere([]db.Where{{Key: "uri", Value: uri.String(), CaseInsensitive: true}}, maybeAccount); err == nil {
		// we have it and it's an account
		return maybeAccount, nil
	} else if err = p.db.GetWhere([]db.Where{{Key: "url", Value: uri.String(), CaseInsensitive: true}}, maybeAccount); err == nil {
		// we have it and it's an account
		return maybeAccount, nil
	}
	if resolve {
		// we don't have it locally so try and dereference it
		account, _, err := p.federator.GetRemoteAccount(authed.Account.Username, uri, true)
		if err != nil {
			return nil, fmt.Errorf("searchAccountByURI: error dereferencing account with uri %s: %s", uri.String(), err)
		}
		return account, nil
	}
	return nil, nil
}

func (p *processor) searchAccountByMention(authed *oauth.Auth, mention string, resolve bool) (*gtsmodel.Account, error) {
	// query is for a remote account
	username, domain, err := util.ExtractMentionParts(mention)
	if err != nil {
		return nil, fmt.Errorf("searchAccountByMention: error extracting mention parts: %s", err)
	}

	// if it's a local account we can skip a whole bunch of stuff
	maybeAcct := &gtsmodel.Account{}
	if domain == p.config.Host {
		if err = p.db.GetLocalAccountByUsername(username, maybeAcct); err != nil {
			return nil, fmt.Errorf("searchAccountByMention: error getting local account by username: %s", err)
		}
		return maybeAcct, nil
	}

	// it's not a local account so first we'll check if it's in the database already...
	where := []db.Where{
		{Key: "username", Value: username, CaseInsensitive: true},
		{Key: "domain", Value: domain, CaseInsensitive: true},
	}
	err = p.db.GetWhere(where, maybeAcct)
	if err == nil {
		// we've got it stored locally already!
		return maybeAcct, nil
	}

	if err != db.ErrNoEntries {
		// if it's  not errNoEntries there's been a real database error so bail at this point
		return nil, fmt.Errorf("searchAccountByMention: database error: %s", err)
	}

	// we got a db.ErrNoEntries, so we just don't have the account locally stored -- check if we can dereference it
	if resolve {
		// we're allowed to resolve it so let's try

		// first we need to webfinger the remote account to convert the username and domain into the activitypub URI for the account
		acctURI, err := p.federator.FingerRemoteAccount(authed.Account.Username, username, domain)
		if err != nil {
			// something went wrong doing the webfinger lookup so we can't process the request
			return nil, fmt.Errorf("searchAccountByMention: error fingering remote account with username %s and domain %s: %s", username, domain, err)
		}

		// we don't have it locally so try and dereference it
		account, _, err := p.federator.GetRemoteAccount(authed.Account.Username, acctURI, true)
		if err != nil {
			return nil, fmt.Errorf("searchAccountByMention: error dereferencing account with uri %s: %s", acctURI.String(), err)
		}
		return account, nil
	}

	return nil, nil
}

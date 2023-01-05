/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

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
*/package trans

import (
	"context"
	"fmt"
	"os"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	transmodel "github.com/superseriousbusiness/gotosocial/internal/trans/model"
)

func (e *exporter) exportAccounts(ctx context.Context, where []db.Where, file *os.File) ([]*transmodel.Account, error) {
	// select using the 'where' we've been provided
	accounts := []*transmodel.Account{}
	if err := e.db.GetWhere(ctx, where, &accounts); err != nil {
		return nil, fmt.Errorf("exportAccounts: error selecting accounts: %s", err)
	}

	// write any accounts found to file
	for _, a := range accounts {
		if err := e.accountEncode(ctx, file, a); err != nil {
			return nil, fmt.Errorf("exportAccounts: error encoding account: %s", err)
		}
	}

	return accounts, nil
}

func (e *exporter) exportBlocks(ctx context.Context, accounts []*transmodel.Account, file *os.File) ([]*transmodel.Block, error) {
	blocksUnique := make(map[string]*transmodel.Block)

	// for each account we want to export both where it's blocking and where it's blocked
	for _, a := range accounts {
		// 1. export blocks owned by given account
		whereBlocking := []db.Where{{Key: "account_id", Value: a.ID}}
		blocking := []*transmodel.Block{}
		if err := e.db.GetWhere(ctx, whereBlocking, &blocking); err != nil {
			return nil, fmt.Errorf("exportBlocks: error selecting blocks owned by account %s: %s", a.ID, err)
		}
		for _, b := range blocking {
			b.Type = transmodel.TransBlock
			if err := e.simpleEncode(ctx, file, b, b.ID); err != nil {
				return nil, fmt.Errorf("exportBlocks: error encoding block owned by account %s: %s", a.ID, err)
			}
			blocksUnique[b.ID] = b
		}

		// 2. export blocks that target given account
		whereBlocked := []db.Where{{Key: "target_account_id", Value: a.ID}}
		blocked := []*transmodel.Block{}
		if err := e.db.GetWhere(ctx, whereBlocked, &blocked); err != nil {
			return nil, fmt.Errorf("exportBlocks: error selecting blocks targeting account %s: %s", a.ID, err)
		}
		for _, b := range blocked {
			b.Type = transmodel.TransBlock
			if err := e.simpleEncode(ctx, file, b, b.ID); err != nil {
				return nil, fmt.Errorf("exportBlocks: error encoding block targeting account %s: %s", a.ID, err)
			}
			blocksUnique[b.ID] = b
		}
	}

	// now return all the blocks we found
	blocks := []*transmodel.Block{}
	for _, b := range blocksUnique {
		blocks = append(blocks, b)
	}

	return blocks, nil
}

func (e *exporter) exportDomainBlocks(ctx context.Context, file *os.File) ([]*transmodel.DomainBlock, error) {
	domainBlocks := []*transmodel.DomainBlock{}

	if err := e.db.GetAll(ctx, &domainBlocks); err != nil {
		return nil, fmt.Errorf("exportBlocks: error selecting domain blocks: %s", err)
	}

	for _, b := range domainBlocks {
		b.Type = transmodel.TransDomainBlock
		if err := e.simpleEncode(ctx, file, b, b.ID); err != nil {
			return nil, fmt.Errorf("exportBlocks: error encoding domain block: %s", err)
		}
	}

	return domainBlocks, nil
}

func (e *exporter) exportFollows(ctx context.Context, accounts []*transmodel.Account, file *os.File) ([]*transmodel.Follow, error) {
	followsUnique := make(map[string]*transmodel.Follow)

	// for each account we want to export both where it's following and where it's followed
	for _, a := range accounts {
		// 1. export follows owned by given account
		whereFollowing := []db.Where{{Key: "account_id", Value: a.ID}}
		following := []*transmodel.Follow{}
		if err := e.db.GetWhere(ctx, whereFollowing, &following); err != nil {
			return nil, fmt.Errorf("exportFollows: error selecting follows owned by account %s: %s", a.ID, err)
		}
		for _, follow := range following {
			follow.Type = transmodel.TransFollow
			if err := e.simpleEncode(ctx, file, follow, follow.ID); err != nil {
				return nil, fmt.Errorf("exportFollows: error encoding follow owned by account %s: %s", a.ID, err)
			}
			followsUnique[follow.ID] = follow
		}

		// 2. export follows that target given account
		whereFollowed := []db.Where{{Key: "target_account_id", Value: a.ID}}
		followed := []*transmodel.Follow{}
		if err := e.db.GetWhere(ctx, whereFollowed, &followed); err != nil {
			return nil, fmt.Errorf("exportFollows: error selecting follows targeting account %s: %s", a.ID, err)
		}
		for _, follow := range followed {
			follow.Type = transmodel.TransFollow
			if err := e.simpleEncode(ctx, file, follow, follow.ID); err != nil {
				return nil, fmt.Errorf("exportFollows: error encoding follow targeting account %s: %s", a.ID, err)
			}
			followsUnique[follow.ID] = follow
		}
	}

	// now return all the follows we found
	follows := []*transmodel.Follow{}
	for _, follow := range followsUnique {
		follows = append(follows, follow)
	}

	return follows, nil
}

func (e *exporter) exportFollowRequests(ctx context.Context, accounts []*transmodel.Account, file *os.File) ([]*transmodel.FollowRequest, error) {
	frsUnique := make(map[string]*transmodel.FollowRequest)

	// for each account we want to export both where it's following and where it's followed
	for _, a := range accounts {
		// 1. export follow requests owned by given account
		whereRequesting := []db.Where{{Key: "account_id", Value: a.ID}}
		requesting := []*transmodel.FollowRequest{}
		if err := e.db.GetWhere(ctx, whereRequesting, &requesting); err != nil {
			return nil, fmt.Errorf("exportFollowRequests: error selecting follow requests owned by account %s: %s", a.ID, err)
		}
		for _, fr := range requesting {
			fr.Type = transmodel.TransFollowRequest
			if err := e.simpleEncode(ctx, file, fr, fr.ID); err != nil {
				return nil, fmt.Errorf("exportFollowRequests: error encoding follow request owned by account %s: %s", a.ID, err)
			}
			frsUnique[fr.ID] = fr
		}

		// 2. export follow requests that target given account
		whereRequested := []db.Where{{Key: "target_account_id", Value: a.ID}}
		requested := []*transmodel.FollowRequest{}
		if err := e.db.GetWhere(ctx, whereRequested, &requested); err != nil {
			return nil, fmt.Errorf("exportFollowRequests: error selecting follow requests targeting account %s: %s", a.ID, err)
		}
		for _, fr := range requested {
			fr.Type = transmodel.TransFollowRequest
			if err := e.simpleEncode(ctx, file, fr, fr.ID); err != nil {
				return nil, fmt.Errorf("exportFollowRequests: error encoding follow request targeting account %s: %s", a.ID, err)
			}
			frsUnique[fr.ID] = fr
		}
	}

	// now return all the followRequests we found
	followRequests := []*transmodel.FollowRequest{}
	for _, fr := range frsUnique {
		followRequests = append(followRequests, fr)
	}

	return followRequests, nil
}

func (e *exporter) exportInstances(ctx context.Context, file *os.File) ([]*transmodel.Instance, error) {
	instances := []*transmodel.Instance{}

	if err := e.db.GetAll(ctx, &instances); err != nil {
		return nil, fmt.Errorf("exportInstances: error selecting instance: %s", err)
	}

	for _, u := range instances {
		u.Type = transmodel.TransInstance
		if err := e.simpleEncode(ctx, file, u, u.ID); err != nil {
			return nil, fmt.Errorf("exportInstances: error encoding instance: %s", err)
		}
	}

	return instances, nil
}

func (e *exporter) exportUsers(ctx context.Context, file *os.File) ([]*transmodel.User, error) {
	users := []*transmodel.User{}

	if err := e.db.GetAll(ctx, &users); err != nil {
		return nil, fmt.Errorf("exportUsers: error selecting users: %s", err)
	}

	for _, u := range users {
		u.Type = transmodel.TransUser
		if err := e.simpleEncode(ctx, file, u, u.ID); err != nil {
			return nil, fmt.Errorf("exportUsers: error encoding user: %s", err)
		}
	}

	return users, nil
}

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

package account

import (
	"context"
	"errors"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
)

// ExportStats returns the requester's export stats,
// ie., the counts of items that can be exported.
func (p *Processor) ExportStats(
	ctx context.Context,
	requester *gtsmodel.Account,
) (*apimodel.AccountExportStats, gtserror.WithCode) {
	exportStats, err := p.converter.AccountToExportStats(ctx, requester)
	if err != nil {
		err = gtserror.Newf("db error getting export stats: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return exportStats, nil
}

// ExportFollowing returns a CSV file of
// accounts that the requester follows.
func (p *Processor) ExportFollowing(
	ctx context.Context,
	requester *gtsmodel.Account,
) ([][]string, gtserror.WithCode) {
	// Fetch accounts followed by requester,
	// using a nil page to get everything.
	following, err := p.state.DB.GetAccountFollows(ctx, requester.ID, nil)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err = gtserror.Newf("db error getting follows: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Convert accounts to CSV-compatible
	// records, with appropriate column headers.
	records, err := p.converter.FollowingToCSV(ctx, following)
	if err != nil {
		err = gtserror.Newf("error converting follows to records: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return records, nil
}

// ExportFollowers returns a CSV file of
// accounts that follow the requester.
func (p *Processor) ExportFollowers(
	ctx context.Context,
	requester *gtsmodel.Account,
) ([][]string, gtserror.WithCode) {
	// Fetch accounts following requester,
	// using a nil page to get everything.
	followers, err := p.state.DB.GetAccountFollowers(ctx, requester.ID, nil)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err = gtserror.Newf("db error getting followers: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Convert accounts to CSV-compatible
	// records, with appropriate column headers.
	records, err := p.converter.FollowersToCSV(ctx, followers)
	if err != nil {
		err = gtserror.Newf("error converting followers to records: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return records, nil
}

// ExportLists returns a CSV file of
// lists created by the requester.
func (p *Processor) ExportLists(
	ctx context.Context,
	requester *gtsmodel.Account,
) ([][]string, gtserror.WithCode) {
	lists, err := p.state.DB.GetListsByAccountID(ctx, requester.ID)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err = gtserror.Newf("db error getting lists: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Convert lists to CSV-compatible records.
	records, err := p.converter.ListsToCSV(ctx, lists)
	if err != nil {
		err = gtserror.Newf("error converting lists to records: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return records, nil
}

// ExportBlocks returns a CSV file of
// account blocks created by the requester.
func (p *Processor) ExportBlocks(
	ctx context.Context,
	requester *gtsmodel.Account,
) ([][]string, gtserror.WithCode) {
	blocks, err := p.state.DB.GetAccountBlocking(ctx, requester.ID, nil)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err = gtserror.Newf("db error getting blocks: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Convert blocks to CSV-compatible records.
	records, err := p.converter.BlocksToCSV(ctx, blocks)
	if err != nil {
		err = gtserror.Newf("error converting blocks to records: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return records, nil
}

// ExportMutes returns a CSV file of
// account mutes created by the requester.
func (p *Processor) ExportMutes(
	ctx context.Context,
	requester *gtsmodel.Account,
) ([][]string, gtserror.WithCode) {
	mutes, err := p.state.DB.GetAccountMutes(ctx, requester.ID, nil)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err = gtserror.Newf("db error getting mutes: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Convert mutes to CSV-compatible records.
	records, err := p.converter.MutesToCSV(ctx, mutes)
	if err != nil {
		err = gtserror.Newf("error converting mutes to records: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return records, nil
}

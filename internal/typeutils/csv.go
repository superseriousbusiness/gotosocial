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

package typeutils

import (
	"context"
	"strconv"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (c *Converter) AccountToExportStats(
	ctx context.Context,
	a *gtsmodel.Account,
) (*apimodel.AccountExportStats, error) {
	// Ensure account stats populated.
	if a.Stats == nil {
		if err := c.state.DB.PopulateAccountStats(ctx, a); err != nil {
			return nil, gtserror.Newf(
				"error getting stats for account %s: %w",
				a.ID, err,
			)
		}
	}

	listsCount, err := c.state.DB.CountListsForAccountID(ctx, a.ID)
	if err != nil {
		return nil, gtserror.Newf(
			"error counting lists for account %s: %w",
			a.ID, err,
		)
	}

	blockingCount, err := c.state.DB.CountAccountBlocks(ctx, a.ID)
	if err != nil {
		return nil, gtserror.Newf(
			"error counting lists for account %s: %w",
			a.ID, err,
		)
	}

	mutingCount, err := c.state.DB.CountAccountMutes(ctx, a.ID)
	if err != nil {
		return nil, gtserror.Newf(
			"error counting lists for account %s: %w",
			a.ID, err,
		)
	}

	return &apimodel.AccountExportStats{
		FollowersCount: *a.Stats.FollowersCount,
		FollowingCount: *a.Stats.FollowingCount,
		StatusesCount:  *a.Stats.StatusesCount,
		ListsCount:     listsCount,
		BlocksCount:    blockingCount,
		MutesCount:     mutingCount,
	}, nil
}

// FollowingToCSV converts a slice of follows into
// a slice of CSV-compatible Following records.
func (c *Converter) FollowingToCSV(
	ctx context.Context,
	following []*gtsmodel.Follow,
) ([][]string, error) {
	// Records should be length of
	// input + 1 so we can add headers.
	records := make([][]string, 1, len(following)+1)

	// Add headers at the
	// top of records.
	records[0] = []string{
		"Account address",
		"Show boosts",
	}

	// We need to know our own domain for this.
	// Try account domain, fall back to host.
	thisDomain := config.GetAccountDomain()
	if thisDomain == "" {
		thisDomain = config.GetHost()
	}

	// For each item, add a record.
	for _, follow := range following {
		if follow.TargetAccount == nil {
			// Retrieve target account.
			var err error
			follow.TargetAccount, err = c.state.DB.GetAccountByID(
				// Barebones is fine here.
				gtscontext.SetBarebones(ctx),
				follow.TargetAccountID,
			)
			if err != nil {
				return nil, gtserror.Newf(
					"db error getting target account for follow %s: %w",
					follow.ID, err,
				)
			}
		}

		domain := follow.TargetAccount.Domain
		if domain == "" {
			// Local account,
			// use our domain.
			domain = thisDomain
		}

		records = append(records, []string{
			// Account address: eg., someone@example.org
			// -- NOTE: without the leading '@'!
			follow.TargetAccount.Username + "@" + domain,
			// Show boosts: eg., true
			strconv.FormatBool(*follow.ShowReblogs),
		})
	}

	return records, nil
}

// FollowersToCSV converts a slice of follows into
// a slice of CSV-compatible Followers records.
func (c *Converter) FollowersToCSV(
	ctx context.Context,
	followers []*gtsmodel.Follow,
) ([][]string, error) {
	// Records should be length of
	// input + 1 so we can add headers.
	records := make([][]string, 1, len(followers)+1)

	// Add header at the
	// top of records.
	records[0] = []string{
		"Account address",
	}

	// We need to know our own domain for this.
	// Try account domain, fall back to host.
	thisDomain := config.GetAccountDomain()
	if thisDomain == "" {
		thisDomain = config.GetHost()
	}

	// For each item, add a record.
	for _, follow := range followers {
		if follow.Account == nil {
			// Retrieve account.
			var err error
			follow.Account, err = c.state.DB.GetAccountByID(
				// Barebones is fine here.
				gtscontext.SetBarebones(ctx),
				follow.AccountID,
			)
			if err != nil {
				return nil, gtserror.Newf(
					"db error getting account for follow %s: %w",
					follow.ID, err,
				)
			}
		}

		domain := follow.Account.Domain
		if domain == "" {
			// Local account,
			// use our domain.
			domain = thisDomain
		}

		records = append(records, []string{
			// Account address: eg., someone@example.org
			// -- NOTE: without the leading '@'!
			follow.Account.Username + "@" + domain,
		})
	}

	return records, nil
}

// FollowersToCSV converts a slice of follows into
// a slice of CSV-compatible Followers records.
func (c *Converter) ListsToCSV(
	ctx context.Context,
	lists []*gtsmodel.List,
) ([][]string, error) {
	// We need to know our own domain for this.
	// Try account domain, fall back to host.
	thisDomain := config.GetAccountDomain()
	if thisDomain == "" {
		thisDomain = config.GetHost()
	}

	// NOTE: Mastodon-compatible lists
	// CSV doesn't use column headers.
	records := make([][]string, 0)

	// For each item, add a record.
	for _, list := range lists {
		for _, entry := range list.ListEntries {
			if entry.Follow == nil {
				// Retrieve follow.
				var err error
				entry.Follow, err = c.state.DB.GetFollowByID(
					ctx,
					entry.FollowID,
				)
				if err != nil {
					return nil, gtserror.Newf(
						"db error getting follow for list entry %s: %w",
						entry.ID, err,
					)
				}
			}

			if entry.Follow.TargetAccount == nil {
				// Retrieve account.
				var err error
				entry.Follow.TargetAccount, err = c.state.DB.GetAccountByID(
					// Barebones is fine here.
					gtscontext.SetBarebones(ctx),
					entry.Follow.TargetAccountID,
				)
				if err != nil {
					return nil, gtserror.Newf(
						"db error getting target account for list entry %s: %w",
						entry.ID, err,
					)
				}
			}

			var (
				username = entry.Follow.TargetAccount.Username
				domain   = entry.Follow.TargetAccount.Domain
			)

			if domain == "" {
				// Local account,
				// use our domain.
				domain = thisDomain
			}

			records = append(records, []string{
				// List title: eg., Very cool list
				list.Title,
				// Account address: eg., someone@example.org
				// -- NOTE: without the leading '@'!
				username + "@" + domain,
			})
		}

	}

	return records, nil
}

// BlocksToCSV converts a slice of blocks into
// a slice of CSV-compatible blocks records.
func (c *Converter) BlocksToCSV(
	ctx context.Context,
	blocks []*gtsmodel.Block,
) ([][]string, error) {
	// We need to know our own domain for this.
	// Try account domain, fall back to host.
	thisDomain := config.GetAccountDomain()
	if thisDomain == "" {
		thisDomain = config.GetHost()
	}

	// NOTE: Mastodon-compatible blocks
	// CSV doesn't use column headers.
	records := make([][]string, 0, len(blocks))

	// For each item, add a record.
	for _, block := range blocks {
		if block.TargetAccount == nil {
			// Retrieve target account.
			var err error
			block.TargetAccount, err = c.state.DB.GetAccountByID(
				// Barebones is fine here.
				gtscontext.SetBarebones(ctx),
				block.TargetAccountID,
			)
			if err != nil {
				return nil, gtserror.Newf(
					"db error getting target account for block %s: %w",
					block.ID, err,
				)
			}
		}

		domain := block.TargetAccount.Domain
		if domain == "" {
			// Local account,
			// use our domain.
			domain = thisDomain
		}

		records = append(records, []string{
			// Account address: eg., someone@example.org
			// -- NOTE: without the leading '@'!
			block.TargetAccount.Username + "@" + domain,
		})
	}

	return records, nil
}

// MutesToCSV converts a slice of mutes into
// a slice of CSV-compatible mute records.
func (c *Converter) MutesToCSV(
	ctx context.Context,
	mutes []*gtsmodel.UserMute,
) ([][]string, error) {
	// Records should be length of
	// input + 1 so we can add headers.
	records := make([][]string, 1, len(mutes)+1)

	// Add headers at the
	// top of records.
	records[0] = []string{
		"Account address",
		"Hide notifications",
	}

	// We need to know our own domain for this.
	// Try account domain, fall back to host.
	thisDomain := config.GetAccountDomain()
	if thisDomain == "" {
		thisDomain = config.GetHost()
	}

	// For each item, add a record.
	for _, mute := range mutes {
		if mute.TargetAccount == nil {
			// Retrieve target account.
			var err error
			mute.TargetAccount, err = c.state.DB.GetAccountByID(
				// Barebones is fine here.
				gtscontext.SetBarebones(ctx),
				mute.TargetAccountID,
			)
			if err != nil {
				return nil, gtserror.Newf(
					"db error getting target account for mute %s: %w",
					mute.ID, err,
				)
			}
		}

		domain := mute.TargetAccount.Domain
		if domain == "" {
			// Local account,
			// use our domain.
			domain = thisDomain
		}

		records = append(records, []string{
			// Account address: eg., someone@example.org
			// -- NOTE: without the leading '@'!
			mute.TargetAccount.Username + "@" + domain,
			// Hide notifications: eg., true
			strconv.FormatBool(*mute.Notifications),
		})
	}

	return records, nil
}

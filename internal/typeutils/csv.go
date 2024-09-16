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
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

func (c *Converter) AccountToExportStats(
	ctx context.Context,
	a *gtsmodel.Account,
) (*apimodel.AccountExportStats, error) {
	// Ensure account stats populated.
	if err := c.state.DB.PopulateAccountStats(ctx, a); err != nil {
		return nil, gtserror.Newf(
			"error getting stats for account %s: %w",
			a.ID, err,
		)
	}

	listsCount, err := c.state.DB.CountListsByAccountID(ctx, a.ID)
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
		"Notify on new posts",
		"Languages",
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
			// Notify on new posts, eg., true
			strconv.FormatBool(*follow.Notify),
			// Languages: compat only, leave blank.
			"",
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

		// Get all follows contained with this list.
		follows, err := c.state.DB.GetFollowsInList(ctx,
			list.ID,
			nil,
		)
		if err != nil {
			err := gtserror.Newf("db error getting follows for list: %w", err)
			return nil, err
		}

		// Append each follow as CSV record.
		for _, follow := range follows {
			var (
				// Extract username / domain from target.
				username = follow.TargetAccount.Username
				domain   = follow.TargetAccount.Domain
			)

			if domain == "" {
				// Local account,
				// use our domain.
				domain = thisDomain
			}

			records = append(records, []string{
				// List title: e.g.
				// Very cool list
				list.Title,

				// Account address: e.g.,
				// someone@example.org
				// NOTE: without the leading '@'!
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

// CSVToFollowing converts a slice of CSV records
// to a slice of barebones *gtsmodel.Follow's,
// ready for further processing.
//
// Only TargetAccount.Username, TargetAccount.Domain,
// and ShowReblogs will be set on each Follow.
func (c *Converter) CSVToFollowing(
	ctx context.Context,
	records [][]string,
) ([]*gtsmodel.Follow, error) {
	// We need to know our own domain for this.
	// Try account domain, fall back to host.
	var (
		thisHost          = config.GetHost()
		thisAccountDomain = config.GetAccountDomain()
		follows           = make([]*gtsmodel.Follow, 0, len(records))
	)

	for _, record := range records {
		recordLen := len(record)

		// Older versions of this Masto CSV
		// schema may not include "Show boosts",
		// "Notify on new posts", or "Languages",
		// so be lenient here in what we accept.
		if recordLen == 0 ||
			recordLen > 4 {
			// Badly formatted,
			// skip this one.
			continue
		}

		// "Account address"
		namestring := record[0]
		if namestring == "" {
			// Badly formatted,
			// skip this one.
			continue
		}

		if namestring == "Account address" {
			// CSV header row,
			// skip this one.
			continue
		}

		// Prepend with "@"
		// if not included.
		if namestring[0] != '@' {
			namestring = "@" + namestring
		}

		username, domain, err := util.ExtractNamestringParts(namestring)
		if err != nil {
			// Badly formatted,
			// skip this one.
			continue
		}

		if domain == thisHost || domain == thisAccountDomain {
			// Clear the domain,
			// since it's ours.
			domain = ""
		}

		// "Show boosts"
		var showReblogs *bool
		if recordLen > 1 {
			b, err := strconv.ParseBool(record[1])
			if err != nil {
				// Badly formatted,
				// skip this one.
				continue
			}
			showReblogs = &b
		}

		// "Notify on new posts"
		var notify *bool
		if recordLen > 2 {
			b, err := strconv.ParseBool(record[2])
			if err != nil {
				// Badly formatted,
				// skip this one.
				continue
			}
			notify = &b
		}

		// TODO: "Languages"
		//
		// Ignore this for now as we
		// don't do anything with it.

		// Looks good, whack it in the slice.
		follows = append(follows, &gtsmodel.Follow{
			TargetAccount: &gtsmodel.Account{
				Username: username,
				Domain:   domain,
			},
			ShowReblogs: showReblogs,
			Notify:      notify,
		})
	}

	return follows, nil
}

// CSVToBlocks converts a slice of CSV records
// to a slice of barebones *gtsmodel.Block's,
// ready for further processing.
//
// Only TargetAccount.Username and TargetAccount.Domain
// will be set on each Block.
func (c *Converter) CSVToBlocks(
	ctx context.Context,
	records [][]string,
) ([]*gtsmodel.Block, error) {
	// We need to know our own domain for this.
	// Try account domain, fall back to host.
	var (
		thisHost          = config.GetHost()
		thisAccountDomain = config.GetAccountDomain()
		blocks            = make([]*gtsmodel.Block, 0, len(records))
	)

	for _, record := range records {
		if len(record) != 1 {
			// Badly formatted,
			// skip this one.
			continue
		}

		namestring := record[0]
		if namestring == "" {
			// Badly formatted,
			// skip this one.
			continue
		}

		// Prepend with "@"
		// if not included.
		if namestring[0] != '@' {
			namestring = "@" + namestring
		}

		username, domain, err := util.ExtractNamestringParts(namestring)
		if err != nil {
			// Badly formatted,
			// skip this one.
			continue
		}

		if domain == thisHost || domain == thisAccountDomain {
			// Clear the domain,
			// since it's ours.
			domain = ""
		}

		// Looks good, whack it in the slice.
		blocks = append(blocks, &gtsmodel.Block{
			TargetAccount: &gtsmodel.Account{
				Username: username,
				Domain:   domain,
			},
		})
	}

	return blocks, nil
}

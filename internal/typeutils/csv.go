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
	"cmp"
	"context"
	"slices"
	"strconv"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/util"
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

	blockingCount, err := c.state.DB.CountAccountBlocking(ctx, a.ID)
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
//
// Each follow should be populated.
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

	// Pre-sort the follows
	// by domain and username.
	slices.SortFunc(
		following,
		func(a *gtsmodel.Follow, b *gtsmodel.Follow) int {
			aStr := a.TargetAccount.Domain + "/" + a.TargetAccount.Username
			bStr := b.TargetAccount.Domain + "/" + b.TargetAccount.Username
			return cmp.Compare(aStr, bStr)
		},
	)

	// For each item, add a record.
	for _, follow := range following {
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
//
// Each follow should be populated.
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

	// Pre-sort the follows
	// by domain and username.
	slices.SortFunc(
		followers,
		func(a *gtsmodel.Follow, b *gtsmodel.Follow) int {
			aStr := a.Account.Domain + "/" + a.Account.Username
			bStr := b.Account.Domain + "/" + b.Account.Username
			return cmp.Compare(aStr, bStr)
		},
	)

	// For each item, add a record.
	for _, follow := range followers {
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

	// Pre-sort the lists
	// alphabetically.
	slices.SortFunc(
		lists,
		func(a *gtsmodel.List, b *gtsmodel.List) int {
			return cmp.Compare(a.Title, b.Title)
		},
	)

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

		// Pre-sort the follows
		// by domain and username.
		slices.SortFunc(
			follows,
			func(a *gtsmodel.Follow, b *gtsmodel.Follow) int {
				aStr := a.TargetAccount.Domain + "/" + a.TargetAccount.Username
				bStr := b.TargetAccount.Domain + "/" + b.TargetAccount.Username
				return cmp.Compare(aStr, bStr)
			},
		)

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
//
// Each block should be populated.
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

	// Pre-sort the blocks
	// by domain and username.
	slices.SortFunc(
		blocks,
		func(a *gtsmodel.Block, b *gtsmodel.Block) int {
			aStr := a.TargetAccount.Domain + "/" + a.TargetAccount.Username
			bStr := b.TargetAccount.Domain + "/" + b.TargetAccount.Username
			return cmp.Compare(aStr, bStr)
		},
	)

	// For each item, add a record.
	for _, block := range blocks {
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
//
// Each mute should be populated.
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

	// Pre-sort the mutes
	// by domain and username.
	slices.SortFunc(
		mutes,
		func(a *gtsmodel.UserMute, b *gtsmodel.UserMute) int {
			aStr := a.TargetAccount.Domain + "/" + a.TargetAccount.Username
			bStr := b.TargetAccount.Domain + "/" + b.TargetAccount.Username
			return cmp.Compare(aStr, bStr)
		},
	)

	// For each item, add a record.
	for _, mute := range mutes {
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

// CSVToMutes converts a slice of CSV records
// to a slice of barebones *gtsmodel.UserMute's,
// ready for further processing.
//
// Only TargetAccount.Username, TargetAccount.Domain,
// and Notifications will be set on each mute.
//
// The CSV format does not hold expiration data, so
// all imported mutes will be permanent, possibly
// overwriting existing temporary mutes.
func (c *Converter) CSVToMutes(
	ctx context.Context,
	records [][]string,
) ([]*gtsmodel.UserMute, error) {
	// We need to know our own domain for this.
	// Try account domain, fall back to host.
	var (
		thisHost          = config.GetHost()
		thisAccountDomain = config.GetAccountDomain()
		mutes             = make([]*gtsmodel.UserMute, 0, len(records)-1)
	)

	for _, record := range records {
		recordLen := len(record)

		// Older versions of this Masto CSV
		// schema may not include "Hide notifications",
		// so be lenient here in what we accept.
		if recordLen == 0 ||
			recordLen > 2 {
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

		// "Hide notifications"
		var hideNotifications *bool
		if recordLen > 1 {
			b, err := strconv.ParseBool(record[1])
			if err != nil {
				// Badly formatted,
				// skip this one.
				continue
			}
			hideNotifications = &b
		}

		// Looks good, whack it in the slice.
		mutes = append(mutes, &gtsmodel.UserMute{
			TargetAccount: &gtsmodel.Account{
				Username: username,
				Domain:   domain,
			},
			Notifications: hideNotifications,
		})
	}

	return mutes, nil
}

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
	"encoding/csv"
	"errors"
	"fmt"
	"mime/multipart"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/log"
)

func (p *Processor) ImportData(
	ctx context.Context,
	requester *gtsmodel.Account,
	data *multipart.FileHeader,
	importType string,
	overwrite bool,
) gtserror.WithCode {
	switch importType {

	case "following":
		return p.importFollowing(
			ctx,
			requester,
			data,
			overwrite,
		)

	case "blocks":
		return p.importBlocks(
			ctx,
			requester,
			data,
			overwrite,
		)

	case "mutes":
		return p.importMutes(
			ctx,
			requester,
			data,
			overwrite,
		)

	default:
		const text = "import type not yet supported"
		return gtserror.NewErrorUnprocessableEntity(errors.New(text), text)
	}
}

func (p *Processor) importFollowing(
	ctx context.Context,
	requester *gtsmodel.Account,
	followingData *multipart.FileHeader,
	overwrite bool,
) gtserror.WithCode {
	file, err := followingData.Open()
	if err != nil {
		err := fmt.Errorf("error opening following data file: %w", err)
		return gtserror.NewErrorBadRequest(err, err.Error())
	}
	defer file.Close()

	// Parse records out of the file.
	records, err := csv.NewReader(file).ReadAll()
	if err != nil {
		err := fmt.Errorf("error reading following data file: %w", err)
		return gtserror.NewErrorBadRequest(err, err.Error())
	}

	// Convert the records into a slice of barebones follows.
	//
	// Only TargetAccount.Username, TargetAccount.Domain,
	// and ShowReblogs will be set on each Follow.
	follows, err := p.converter.CSVToFollowing(ctx, records)
	if err != nil {
		err := fmt.Errorf("error converting records to follows: %w", err)
		return gtserror.NewErrorBadRequest(err, err.Error())
	}

	// Do remaining processing of this import asynchronously.
	f := importFollowingAsyncF(p, requester, follows, overwrite)
	p.state.Workers.Processing.Queue.Push(f)

	return nil
}

func importFollowingAsyncF(
	p *Processor,
	requester *gtsmodel.Account,
	follows []*gtsmodel.Follow,
	overwrite bool,
) func(context.Context) {
	return func(ctx context.Context) {
		// Map used to store wanted
		// follow targets (if overwriting).
		var wantedFollows map[string]struct{}

		if overwrite {
			// If we're overwriting, we need to get current
			// follow(-req)s owned by requester *before*
			// making any changes, so that we can remove
			// unwanted follows after we've created new ones.
			prevFollows, err := p.state.DB.GetAccountFollows(ctx, requester.ID, nil)
			if err != nil {
				log.Errorf(ctx, "db error getting following: %v", err)
				return
			}

			prevFollowReqs, err := p.state.DB.GetAccountFollowRequesting(ctx, requester.ID, nil)
			if err != nil {
				log.Errorf(ctx, "db error getting follow requesting: %v", err)
				return
			}

			// Initialize new follows map.
			wantedFollows = make(map[string]struct{}, len(follows))

			// Once we've created (or tried to create)
			// the required follows, go through previous
			// follow(-request)s and remove unwanted ones.
			defer func() {

				// AccountIDs to unfollow.
				toRemove := []string{}

				// Check previous follows.
				for _, prev := range prevFollows {
					username := prev.TargetAccount.Username
					domain := prev.TargetAccount.Domain

					_, wanted := wantedFollows[username+"@"+domain]
					if !wanted {
						toRemove = append(toRemove, prev.TargetAccountID)
					}
				}

				// Now any pending follow requests.
				for _, prev := range prevFollowReqs {
					username := prev.TargetAccount.Username
					domain := prev.TargetAccount.Domain

					_, wanted := wantedFollows[username+"@"+domain]
					if !wanted {
						toRemove = append(toRemove, prev.TargetAccountID)
					}
				}

				// Remove each discovered
				// unwanted follow.
				for _, accountID := range toRemove {
					if _, errWithCode := p.FollowRemove(
						ctx,
						requester,
						accountID,
					); errWithCode != nil {
						log.Errorf(ctx, "could not unfollow account: %v", errWithCode.Unwrap())
						continue
					}
				}
			}()
		}

		// Go through the follows parsed from CSV
		// file, and create / update each one.
		for _, follow := range follows {
			var (
				// Username of the target.
				username = follow.TargetAccount.Username

				// Domain of the target.
				// Empty for our domain.
				domain = follow.TargetAccount.Domain

				// Show reblogs on
				// the new follow.
				showReblogs = follow.ShowReblogs

				// Notify when new
				// follow posts.
				notify = follow.Notify
			)

			if overwrite {
				// We'll be overwriting, so store
				// this new follow in our handy map.
				wantedFollows[username+"@"+domain] = struct{}{}
			}

			// Get the target account, dereferencing it if necessary.
			targetAcct, _, err := p.federator.Dereferencer.GetAccountByUsernameDomain(
				ctx,
				requester.Username,
				username,
				domain,
			)
			if err != nil {
				log.Errorf(ctx, "could not retrieve account: %v", err)
				continue
			}

			// Use the processor's FollowCreate function
			// to create or update the follow. This takes
			// account of existing follows, and also sends
			// the follow to the FromClientAPI processor.
			if _, errWithCode := p.FollowCreate(
				ctx,
				requester,
				&apimodel.AccountFollowRequest{
					ID:      targetAcct.ID,
					Reblogs: showReblogs,
					Notify:  notify,
				},
			); errWithCode != nil {
				log.Errorf(ctx, "could not follow account: %v", errWithCode.Unwrap())
				continue
			}
		}
	}
}

func (p *Processor) importBlocks(
	ctx context.Context,
	requester *gtsmodel.Account,
	blocksData *multipart.FileHeader,
	overwrite bool,
) gtserror.WithCode {
	file, err := blocksData.Open()
	if err != nil {
		err := fmt.Errorf("error opening blocks data file: %w", err)
		return gtserror.NewErrorBadRequest(err, err.Error())
	}
	defer file.Close()

	// Parse records out of the file.
	records, err := csv.NewReader(file).ReadAll()
	if err != nil {
		err := fmt.Errorf("error reading blocks data file: %w", err)
		return gtserror.NewErrorBadRequest(err, err.Error())
	}

	// Convert the records into a slice of barebones blocks.
	//
	// Only TargetAccount.Username and TargetAccount.Domain,
	// will be set on each Block.
	blocks, err := p.converter.CSVToBlocks(ctx, records)
	if err != nil {
		err := fmt.Errorf("error converting records to blocks: %w", err)
		return gtserror.NewErrorBadRequest(err, err.Error())
	}

	// Do remaining processing of this import asynchronously.
	f := importBlocksAsyncF(p, requester, blocks, overwrite)
	p.state.Workers.Processing.Queue.Push(f)

	return nil
}

func importBlocksAsyncF(
	p *Processor,
	requester *gtsmodel.Account,
	blocks []*gtsmodel.Block,
	overwrite bool,
) func(context.Context) {
	return func(ctx context.Context) {
		// Map used to store wanted
		// block targets (if overwriting).
		var wantedBlocks map[string]struct{}

		if overwrite {
			// If we're overwriting, we need to get current
			// blocks owned by requester *before* making any
			// changes, so that we can remove unwanted blocks
			// after we've created new ones.
			var (
				prevBlocks []*gtsmodel.Block
				err        error
			)

			prevBlocks, err = p.state.DB.GetAccountBlocks(ctx, requester.ID, nil)
			if err != nil {
				log.Errorf(ctx, "db error getting blocks: %v", err)
				return
			}

			// Initialize new blocks map.
			wantedBlocks = make(map[string]struct{}, len(blocks))

			// Once we've created (or tried to create)
			// the required blocks, go through previous
			// blocks and remove unwanted ones.
			defer func() {
				for _, prev := range prevBlocks {
					username := prev.TargetAccount.Username
					domain := prev.TargetAccount.Domain

					_, wanted := wantedBlocks[username+"@"+domain]
					if wanted {
						// Leave this
						// one alone.
						continue
					}

					if _, errWithCode := p.BlockRemove(
						ctx,
						requester,
						prev.TargetAccountID,
					); errWithCode != nil {
						log.Errorf(ctx, "could not unblock account: %v", errWithCode.Unwrap())
						continue
					}
				}
			}()
		}

		// Go through the blocks parsed from CSV
		// file, and create / update each one.
		for _, block := range blocks {
			var (
				// Username of the target.
				username = block.TargetAccount.Username

				// Domain of the target.
				// Empty for our domain.
				domain = block.TargetAccount.Domain
			)

			if overwrite {
				// We'll be overwriting, so store
				// this new block in our handy map.
				wantedBlocks[username+"@"+domain] = struct{}{}
			}

			// Get the target account, dereferencing it if necessary.
			targetAcct, _, err := p.federator.Dereferencer.GetAccountByUsernameDomain(
				ctx,
				// Provide empty request user to use the
				// instance account to deref the account.
				//
				// It's pointless to make lots of calls
				// to a remote from an account that's about
				// to block that account.
				"",
				username,
				domain,
			)
			if err != nil {
				log.Errorf(ctx, "could not retrieve account: %v", err)
				continue
			}

			// Use the processor's BlockCreate function
			// to create or update the block. This takes
			// account of existing blocks, and also sends
			// the block to the FromClientAPI processor.
			if _, errWithCode := p.BlockCreate(
				ctx,
				requester,
				targetAcct.ID,
			); errWithCode != nil {
				log.Errorf(ctx, "could not block account: %v", errWithCode.Unwrap())
				continue
			}
		}
	}
}

func (p *Processor) importMutes(
	ctx context.Context,
	requester *gtsmodel.Account,
	mutesData *multipart.FileHeader,
	overwrite bool,
) gtserror.WithCode {
	file, err := mutesData.Open()
	if err != nil {
		err := fmt.Errorf("error opening mutes data file: %w", err)
		return gtserror.NewErrorBadRequest(err, err.Error())
	}
	defer file.Close()

	// Parse records out of the file.
	records, err := csv.NewReader(file).ReadAll()
	if err != nil {
		err := fmt.Errorf("error reading mutes data file: %w", err)
		return gtserror.NewErrorBadRequest(err, err.Error())
	}

	// Convert the records into a slice of barebones mutes.
	//
	// Only TargetAccount.Username, TargetAccount.Domain,
	// and Notifications will be set on each mute.
	mutes, err := p.converter.CSVToMutes(ctx, records)
	if err != nil {
		err := fmt.Errorf("error converting records to mutes: %w", err)
		return gtserror.NewErrorBadRequest(err, err.Error())
	}

	// Do remaining processing of this import asynchronously.
	f := importMutesAsyncF(p, requester, mutes, overwrite)
	p.state.Workers.Processing.Queue.Push(f)

	return nil
}

func importMutesAsyncF(
	p *Processor,
	requester *gtsmodel.Account,
	mutes []*gtsmodel.UserMute,
	overwrite bool,
) func(context.Context) {
	return func(ctx context.Context) {
		// Map used to store wanted
		// mute targets (if overwriting).
		var wantedMutes map[string]struct{}

		if overwrite {
			// If we're overwriting, we need to get current
			// mutes owned by requester *before* making any
			// changes, so that we can remove unwanted mutes
			// after we've created new ones.
			var (
				prevMutes []*gtsmodel.UserMute
				err       error
			)

			prevMutes, err = p.state.DB.GetAccountMutes(ctx, requester.ID, nil)
			if err != nil {
				log.Errorf(ctx, "db error getting mutes: %v", err)
				return
			}

			// Initialize new mutes map.
			wantedMutes = make(map[string]struct{}, len(mutes))

			// Once we've created (or tried to create)
			// the required mutes, go through previous
			// mutes and remove unwanted ones.
			defer func() {
				for _, prev := range prevMutes {
					username := prev.TargetAccount.Username
					domain := prev.TargetAccount.Domain

					_, wanted := wantedMutes[username+"@"+domain]
					if wanted {
						// Leave this
						// one alone.
						continue
					}

					if _, errWithCode := p.MuteRemove(
						ctx,
						requester,
						prev.TargetAccountID,
					); errWithCode != nil {
						log.Errorf(ctx, "could not unmute account: %v", errWithCode.Unwrap())
						continue
					}
				}
			}()
		}

		// Go through the mutes parsed from CSV
		// file, and create / update each one.
		for _, mute := range mutes {
			var (
				// Username of the target.
				username = mute.TargetAccount.Username

				// Domain of the target.
				// Empty for our domain.
				domain = mute.TargetAccount.Domain
			)

			if overwrite {
				// We'll be overwriting, so store
				// this new mute in our handy map.
				wantedMutes[username+"@"+domain] = struct{}{}
			}

			// Get the target account, dereferencing it if necessary.
			targetAcct, _, err := p.federator.Dereferencer.GetAccountByUsernameDomain(
				ctx,
				// Provide empty request user to use the
				// instance account to deref the account.
				//
				// It's pointless to make lots of calls
				// to a remote from an account that's about
				// to mute that account.
				"",
				username,
				domain,
			)
			if err != nil {
				log.Errorf(ctx, "could not retrieve account: %v", err)
				continue
			}

			// Use the processor's MuteCreate function
			// to create or update the mute. This takes
			// account of existing mutes, and also sends
			// the mute to the FromClientAPI processor.
			if _, errWithCode := p.MuteCreate(
				ctx,
				requester,
				targetAcct.ID,
				&apimodel.UserMuteCreateUpdateRequest{Notifications: mute.Notifications},
			); errWithCode != nil {
				log.Errorf(ctx, "could not mute account: %v", errWithCode.Unwrap())
				continue
			}
		}
	}
}

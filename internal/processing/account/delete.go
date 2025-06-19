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
	"net"
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/ap"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtscontext"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"code.superseriousbusiness.org/gotosocial/internal/messages"
	"code.superseriousbusiness.org/gotosocial/internal/util"
	"codeberg.org/gruf/go-kv"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const deleteSelectLimit = 100

// Delete deletes an account, and all of that account's statuses, media, follows, notifications, etc etc etc.
// The origin passed here should be either the ID of the account doing the delete (can be itself), or the ID of a domain block.
//
// This delete function handles the case of both local and remote accounts, and processes side
// effects synchronously to not clog worker queues with potentially tens-of-thousands of requests.
func (p *Processor) Delete(ctx context.Context, account *gtsmodel.Account, origin string) error {

	// Prepare a new log entry for account delete.
	log := log.WithContext(ctx).WithFields(kv.Fields{
		{"uri", account.URI},
		{"origin", origin},
	}...)

	var err error

	// Log operation start / stop.
	log.Info("start account delete")
	defer func() {
		if err != nil {
			log.Errorf("fatal error during account delete: %v", err)
		} else {
			log.Info("finished account delete")
		}
	}()

	// Delete statuses *before* anything else as for local
	// accounts we need to federate out deletes, which relies
	// on follows for addressing the appropriate accounts.
	p.deleteAccountStatuses(ctx, &log, account)

	// Now delete relationships to / from account.
	p.deleteAccountRelations(ctx, &log, account)

	// Now delete any notifications to / from account.
	p.deleteAccountNotifications(ctx, &log, account)

	// Delete other peripheral objects ownable /
	// manageable by any local / remote account.
	p.deleteAccountPeripheral(ctx, &log, account)

	if account.IsLocal() {
		// We delete tokens, applications and clients for
		// account as one of the last stages during deletion,
		// as other database models rely on these.
		if err = p.deleteUserAndTokensForAccount(ctx, &log, account); err != nil {
			return err
		}
	}

	// To prevent the account being created again,
	// (which would cause horrible federation shenanigans),
	// the account will be stubbed out to an unusable state
	// with no identifying info remaining, but NOT deleted.
	columns := stubbifyAccount(account, origin)
	if err = p.state.DB.UpdateAccount(ctx, account, columns...); err != nil {
		return gtserror.Newf("error stubbing out account: %v", err)
	}

	return nil
}

func (p *Processor) deleteUserAndTokensForAccount(
	ctx context.Context,
	log *log.Entry,
	account *gtsmodel.Account,
) error {

	// Fetch the associated user for account, on fail return
	// early as all other parts of this func rely on this user.
	user, err := p.state.DB.GetUserByAccountID(ctx, account.ID)
	if err != nil {
		return gtserror.Newf("error getting account user: %v", err)
	}

	// Get list of applications managed by deleting user.
	apps, err := p.state.DB.GetApplicationsManagedByUserID(ctx,
		user.ID,
		nil, // i.e. all
	)
	if err != nil {
		log.Errorf("error getting user applications: %v", err)
	}

	// Delete each app and any tokens it had created
	// (not necessarily owned by deleted account).
	for _, app := range apps {
		if err := p.state.DB.DeleteTokensByClientID(ctx, app.ClientID); err != nil {
			log.Errorf("error deleting application token: %v", err)
		}
		if err := p.state.DB.DeleteApplicationByID(ctx, app.ID); err != nil {
			log.Errorf("error deleting user application: %v", err)
		}
	}

	// Get any remaining access tokens owned by user.
	tokens, err := p.state.DB.GetAccessTokens(ctx,
		user.ID,
		nil, // i.e. all
	)
	if err != nil {
		log.Errorf("error getting user access tokens: %v", err)
	}

	// Delete user access tokens.
	for _, token := range tokens {
		if err := p.state.DB.DeleteTokenByID(ctx, token.ID); err != nil {
			log.Errorf("error deleting user access token: %v", err)
		}
	}

	// Delete any web push subscriptions created by this local user account.
	if err := p.state.DB.DeleteWebPushSubscriptionsByAccountID(ctx, account.ID); err != nil {
		log.Errorf("error deleting account web push subscriptions: %v", err)
	}

	// To prevent the user being created again,
	// the user will be stubbed out to an unusable state
	// with no identifying info remaining, but NOT deleted.
	columns := stubbifyUser(user)
	if err := p.state.DB.UpdateUser(ctx, user, columns...); err != nil {
		return gtserror.Newf("error stubbing out user: %w", err)
	}

	return nil
}

func (p *Processor) deleteAccountRelations(
	ctx context.Context,
	log *log.Entry,
	account *gtsmodel.Account,
) {
	// Get a list of the follows targeting this account.
	followedBy, err := p.state.DB.GetAccountFollowers(ctx,
		account.ID,
		nil, // i.e. all
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		log.Errorf("error getting account followed-bys: %v", err)
	}

	// Delete these follows from database.
	for _, follow := range followedBy {
		if err := p.state.DB.DeleteFollowByID(ctx, follow.ID); err != nil {
			log.Errorf("error deleting account followed-by %s: %v", follow.URI, err)
		}
	}

	// Get a list of the follow requests targeting this account.
	followRequestedBy, err := p.state.DB.GetAccountFollowRequests(ctx,
		account.ID,
		nil, // i.e. all
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		log.Errorf("error getting account follow-requested-bys: %v", err)
	}

	// Delete these follow requests from database.
	for _, followReq := range followRequestedBy {
		if err := p.state.DB.DeleteFollowRequestByID(ctx, followReq.ID); err != nil {
			log.Errorf("error deleting account follow-requested-by %s: %v", followReq.URI, err)
		}
	}

	// Get a list of the blocks targeting this account.
	blockedBy, err := p.state.DB.GetAccountBlockedBy(ctx,
		account.ID,
		nil, // i.e. all
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		log.Errorf("error getting account blocked-bys: %v", err)
	}

	// Delete these blocks from database.
	for _, block := range blockedBy {
		if err := p.state.DB.DeleteBlockByID(ctx, block.ID); err != nil {
			log.Errorf("error deleting account blocked-by %s: %v", block.URI, err)
		}
	}

	// Get the follows originating from this account.
	following, err := p.state.DB.GetAccountFollows(ctx,
		account.ID,
		nil, // i.e. all
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		log.Errorf("error getting account follows: %v", err)
	}

	// Delete these follows from database.
	for _, follow := range following {
		if err := p.state.DB.DeleteFollowByID(ctx, follow.ID); err != nil {
			log.Errorf("error deleting account followed %s: %v", follow.URI, err)
		}
	}

	// Get a list of the follow requests originating from this account.
	followRequesting, err := p.state.DB.GetAccountFollowRequesting(ctx,
		account.ID,
		nil, // i.e. all
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		log.Errorf("error getting account follow-requests: %v", err)
	}

	// Delete these follow requests from database.
	for _, followReq := range followRequesting {
		if err := p.state.DB.DeleteFollowRequestByID(ctx, followReq.ID); err != nil {
			log.Errorf("error deleting account follow-request %s: %v", followReq.URI, err)
		}
	}

	// Get the blocks originating from this account.
	blocking, err := p.state.DB.GetAccountBlocking(ctx,
		account.ID,
		nil, // i.e. all
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		log.Errorf("error getting account blocks: %v", err)
	}

	// Delete these blocks from database.
	for _, block := range blocking {
		if err := p.state.DB.DeleteBlockByID(ctx, block.ID); err != nil {
			log.Errorf("error deleting account block %s: %v", block.URI, err)
		}
	}

	// Delete all mutes targetting / originating from account.
	if err := p.state.DB.DeleteAccountMutes(ctx, account.ID); // nocollapse
	err != nil && !errors.Is(err, db.ErrNoEntries) {
		log.Errorf("error deleting mutes to / from account: %v", err)
	}

	if account.IsLocal() {
		// Process side-effects for deleting
		// of account follows from local user.
		for _, follow := range following {
			p.processSideEffect(ctx, log,
				ap.ActivityUndo,
				ap.ActivityFollow,
				follow,
				account,
				follow.TargetAccount,
			)
		}

		// Process side-effects for deleting of account follow requests
		// from local user. Though handled as though UNDO of a follow.
		for _, followReq := range followRequesting {
			p.processSideEffect(ctx, log,
				ap.ActivityUndo,
				ap.ActivityFollow,
				&gtsmodel.Follow{
					ID:              followReq.ID,
					URI:             followReq.URI,
					AccountID:       followReq.AccountID,
					Account:         followReq.Account,
					TargetAccountID: followReq.TargetAccountID,
					TargetAccount:   followReq.TargetAccount,
					ShowReblogs:     new(bool),
					Notify:          new(bool),
				},
				account,
				followReq.TargetAccount,
			)
		}

		// Process side-effects for deleting
		// of account blocks from local user.
		for _, block := range blocking {
			p.processSideEffect(ctx, log,
				ap.ActivityUndo,
				ap.ActivityBlock,
				block,
				account,
				block.TargetAccount,
			)
		}
	}
}

func (p *Processor) deleteAccountStatuses(
	ctx context.Context,
	log *log.Entry,
	account *gtsmodel.Account,
) {
	var maxID string

	for {
		// Page through deleting account's statuses.
		statuses, err := p.state.DB.GetAccountStatuses(
			gtscontext.SetBarebones(ctx),
			account.ID,
			deleteSelectLimit,
			false,
			false,
			maxID,
			"",
			false,
			false,
		)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			log.Errorf("error getting account statuses: %v", err)
			return
		}

		if len(statuses) == 0 {
			// reached
			// the end.
			break
		}

		// Update next maxID from last status.
		maxID = statuses[len(statuses)-1].ID

		for _, status := range statuses {
			// Ensure account is set.
			status.Account = account

			// Look for any boosts of this status in DB.
			boosts, err := p.state.DB.GetStatusBoosts(
				gtscontext.SetBarebones(ctx),
				status.ID,
			)
			if err != nil && !errors.Is(err, db.ErrNoEntries) {
				log.Errorf("error getting status boosts for %s: %v", status.URI, err)
				continue
			}

			// Prepare to Undo each boost.
			for _, boost := range boosts {

				// Fetch the owning account of this boost.
				boost.Account, err = p.state.DB.GetAccountByID(
					gtscontext.SetBarebones(ctx),
					boost.AccountID,
				)
				if err != nil {
					log.Errorf("error getting owner %s of status boost %s: %v",
						boost.AccountURI, boost.URI, err)
					continue
				}

				// Process boost undo event.
				p.processSideEffect(ctx, log,
					ap.ActivityUndo,
					ap.ActivityAnnounce,
					boost,
					account,
					account,
				)
			}

			// Process status delete event.
			p.processSideEffect(ctx, log,
				ap.ActivityDelete,
				ap.ObjectNote,
				status,
				account,
				account,
			)
		}
	}
}

func (p *Processor) deleteAccountNotifications(
	ctx context.Context,
	log *log.Entry,
	account *gtsmodel.Account,
) {
	if account.IsLocal() {
		// Delete all types of notifications targeting this local account.
		if err := p.state.DB.DeleteNotifications(ctx, nil, account.ID, ""); // nocollapse
		err != nil && !errors.Is(err, db.ErrNoEntries) {
			log.Errorf("error deleting notifications targeting account: %v", err)
		}
	}

	// Delete all types of notifications originating from this account.
	if err := p.state.DB.DeleteNotifications(ctx, nil, "", account.ID); // nocollapse
	err != nil && !errors.Is(err, db.ErrNoEntries) {
		log.Errorf("error deleting notifications originating from account: %v", err)
	}
}

func (p *Processor) deleteAccountPeripheral(
	ctx context.Context,
	log *log.Entry,
	account *gtsmodel.Account,
) {
	if account.IsLocal() {
		// Delete all bookmarks owned by given account, only for local.
		if err := p.state.DB.DeleteStatusBookmarks(ctx, account.ID, ""); // nocollapse
		err != nil && !errors.Is(err, db.ErrNoEntries) {
			log.Errorf("error deleting bookmarks by account: %v", err)
		}

		// Delete all faves owned by given account, only for local.
		if err := p.state.DB.DeleteStatusFaves(ctx, account.ID, ""); // nocollapse
		err != nil && !errors.Is(err, db.ErrNoEntries) {
			log.Errorf("error deleting faves by account: %v", err)
		}

		// Delete all conversations owned by given account, only for local.
		//
		// *Participated* conversations will be retained, leaving up to *their* owners.
		if err := p.state.DB.DeleteConversationsByOwnerAccountID(ctx, account.ID); // nocollapse
		err != nil && !errors.Is(err, db.ErrNoEntries) {
			log.Errorf("error deleting conversations by account: %v", err)
		}

		// Delete all followed tags owned by given account, only for local.
		if err := p.state.DB.DeleteFollowedTagsByAccountID(ctx, account.ID); // nocollapse
		err != nil && !errors.Is(err, db.ErrNoEntries) {
			log.Errorf("error deleting followed tags by account: %v", err)
		}

		// Delete stats model stored for given account, only for local.
		if err := p.state.DB.DeleteAccountStats(ctx, account.ID); err != nil {
			log.Errorf("error deleting stats for account: %v", err)
		}
	}

	// Delete all bookmarks targeting given account, local and remote.
	if err := p.state.DB.DeleteStatusBookmarks(ctx, "", account.ID); // nocollapse
	err != nil && !errors.Is(err, db.ErrNoEntries) {
		log.Errorf("error deleting bookmarks targeting account: %v", err)
	}

	// Delete all faves targeting given account, local and remote.
	if err := p.state.DB.DeleteStatusFaves(ctx, "", account.ID); // nocollapse
	err != nil && !errors.Is(err, db.ErrNoEntries) {
		log.Errorf("error deleting faves targeting account: %v", err)
	}

	// Delete all poll votes owned by given account, local and remote.
	if err := p.state.DB.DeletePollVotesByAccountID(ctx, account.ID); // nocollapse
	err != nil && !errors.Is(err, db.ErrNoEntries) {
		log.Errorf("error deleting poll votes by account: %v", err)
	}

	// Delete all interaction requests from given account, local and remote.
	if err := p.state.DB.DeleteInteractionRequestsByInteractingAccountID(ctx, account.ID); // nocollapse
	err != nil && !errors.Is(err, db.ErrNoEntries) {
		log.Errorf("error deleting interaction requests by account: %v", err)
	}
}

// processSideEffect will process the given side effect details,
// with appropriate worker depending on if origin is local / remote.
func (p *Processor) processSideEffect(
	ctx context.Context,
	log *log.Entry,
	activityType string,
	objectType string,
	gtsModel any,
	origin *gtsmodel.Account,
	target *gtsmodel.Account,
) {
	if origin.IsLocal() {
		// Process side-effect through our client API as this is a local account.
		if err := p.state.Workers.Client.Process(ctx, &messages.FromClientAPI{
			APActivityType: activityType,
			APObjectType:   objectType,
			GTSModel:       gtsModel,
			Origin:         origin,
			Target:         target,
		}); err != nil {
			log.Errorf("error processing %s of %s during local account %s delete: %v", activityType, objectType, origin.ID, err)
		}
	} else {
		// Process side-effect through our fedi API as this is a remote account.
		if err := p.state.Workers.Federator.Process(ctx, &messages.FromFediAPI{
			APActivityType: activityType,
			APObjectType:   objectType,
			GTSModel:       gtsModel,
			Requesting:     origin,
			Receiving:      target,
		}); err != nil {
			log.Errorf("error processing %s of %s during local account %s delete: %v", activityType, objectType, origin.ID, err)
		}
	}
}

// stubbifyAccount renders the given account as a stub,
// removing most information from it and marking it as
// suspended.
//
// The origin parameter refers to the origin of the
// suspension action; should be an account ID or domain
// block ID.
//
// For caller's convenience, this function returns the db
// names of all columns that are updated by it.
func stubbifyAccount(account *gtsmodel.Account, origin string) []string {
	var (
		now   = time.Now()
		never = time.Time{}
	)

	account.FetchedAt = never
	account.AvatarMediaAttachmentID = ""
	account.AvatarRemoteURL = ""
	account.HeaderMediaAttachmentID = ""
	account.HeaderRemoteURL = ""
	account.DisplayName = ""
	account.EmojiIDs = nil
	account.Emojis = nil
	account.Fields = nil
	account.Note = ""
	account.NoteRaw = ""
	account.MemorializedAt = never
	account.AlsoKnownAsURIs = nil
	account.MovedToURI = ""
	account.Discoverable = util.Ptr(false)
	account.SuspendedAt = now
	account.SuspensionOrigin = origin

	return []string{
		"fetched_at",
		"avatar_media_attachment_id",
		"avatar_remote_url",
		"header_media_attachment_id",
		"header_remote_url",
		"display_name",
		"emojis",
		"fields",
		"note",
		"note_raw",
		"memorialized_at",
		"also_known_as_uris",
		"moved_to_uri",
		"discoverable",
		"suspended_at",
		"suspension_origin",
	}
}

// stubbifyUser renders the given user as a stub,
// removing sensitive information like IP addresses
// and sign-in times, but keeping email addresses to
// prevent the same email address from creating another
// account on this instance.
//
// `encrypted_password` is set to the bcrypt hash of a
// random uuid, so if the action is reversed, the user
// will have to reset their password via email.
//
// For caller's convenience, this function returns the db
// names of all columns that are updated by it.
func stubbifyUser(user *gtsmodel.User) []string {
	uuid, err := uuid.New().MarshalBinary()
	if err != nil {
		// this should never happen,
		// it indicates /dev/random
		// is misbehaving.
		panic(err)
	}

	dummyPassword, err := bcrypt.GenerateFromPassword(uuid, bcrypt.DefaultCost)
	if err != nil {
		// this should never happen,
		// it indicates /dev/random
		// is misbehaving.
		panic(err)
	}

	never := time.Time{}

	user.EncryptedPassword = string(dummyPassword)
	user.SignUpIP = net.IPv4zero
	user.Locale = ""
	user.CreatedByApplicationID = ""
	user.LastEmailedAt = never
	user.ConfirmationToken = ""
	user.ConfirmationSentAt = never
	user.ResetPasswordToken = ""
	user.ResetPasswordSentAt = never

	return []string{
		"encrypted_password",
		"sign_up_ip",
		"locale",
		"created_by_application_id",
		"last_emailed_at",
		"confirmation_token",
		"confirmation_sent_at",
		"reset_password_token",
		"reset_password_sent_at",
	}
}

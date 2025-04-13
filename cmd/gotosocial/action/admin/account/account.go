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
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/superseriousbusiness/gotosocial/cmd/gotosocial/action"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db/bundb"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/util"
	"github.com/superseriousbusiness/gotosocial/internal/validate"
	"golang.org/x/crypto/bcrypt"
)

func initState(ctx context.Context) (*state.State, error) {
	var state state.State
	state.Caches.Init()
	if err := state.Caches.Start(); err != nil {
		return nil, fmt.Errorf("error starting caches: %w", err)
	}

	// Only set state DB connection.
	// Don't need Actions or Workers for this (yet).
	dbConn, err := bundb.NewBunDBService(ctx, &state)
	if err != nil {
		return nil, fmt.Errorf("error creating dbConn: %w", err)
	}
	state.DB = dbConn

	return &state, nil
}

func stopState(state *state.State) error {
	err := state.DB.Close()
	state.Caches.Stop()
	return err
}

// Create creates a new account and user
// in the database using the provided flags.
var Create action.GTSAction = func(ctx context.Context) error {
	state, err := initState(ctx)
	if err != nil {
		return err
	}

	defer func() {
		// Ensure state gets stopped on return.
		if err := stopState(state); err != nil {
			log.Error(ctx, err)
		}
	}()

	username := config.GetAdminAccountUsername()
	if err := validate.Username(username); err != nil {
		return err
	}

	usernameAvailable, err := state.DB.IsUsernameAvailable(ctx, username)
	if err != nil {
		return err
	}

	if !usernameAvailable {
		return fmt.Errorf("username %s is already in use", username)
	}

	email := config.GetAdminAccountEmail()
	if err := validate.Email(email); err != nil {
		return err
	}

	emailAvailable, err := state.DB.IsEmailAvailable(ctx, email)
	if err != nil {
		return err
	}

	if !emailAvailable {
		return fmt.Errorf("email address %s is already in use", email)
	}

	password := config.GetAdminAccountPassword()
	if err := validate.Password(password); err != nil {
		return err
	}

	_, err = state.DB.NewSignup(ctx, gtsmodel.NewSignup{
		Username:      username,
		Email:         email,
		Password:      password,
		EmailVerified: true, // Assume cli user wants email marked as verified already.
		PreApproved:   true, // Assume cli user wants account marked as approved already.
	})
	return err
}

// List returns all existing local accounts.
var List action.GTSAction = func(ctx context.Context) error {
	state, err := initState(ctx)
	if err != nil {
		return err
	}

	users, err := state.DB.GetAllUsers(ctx)
	if err != nil {
		return err
	}

	fmtBool := func(b *bool) string {
		if b == nil {
			return "unknown"
		}
		if *b {
			return "yes"
		}
		return "no"
	}

	fmtDate := func(t time.Time) string {
		if t.Equal(time.Time{}) {
			return "no"
		}
		return "yes"
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	fmt.Fprintln(w, "user\taccount\tapproved\tadmin\tmoderator\tsuspended\tconfirmed")
	for _, u := range users {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n", u.Account.Username, u.AccountID, fmtBool(u.Approved), fmtBool(u.Admin), fmtBool(u.Moderator), fmtDate(u.Account.SuspendedAt), fmtDate(u.ConfirmedAt))
	}
	return w.Flush()
}

// Confirm sets a user to Approved, sets Email to the current
// UnconfirmedEmail value, and sets ConfirmedAt to now.
var Confirm action.GTSAction = func(ctx context.Context) error {
	state, err := initState(ctx)
	if err != nil {
		return err
	}

	defer func() {
		// Ensure state gets stopped on return.
		if err := stopState(state); err != nil {
			log.Error(ctx, err)
		}
	}()

	username := config.GetAdminAccountUsername()
	if err := validate.Username(username); err != nil {
		return err
	}

	account, err := state.DB.GetAccountByUsernameDomain(ctx, username, "")
	if err != nil {
		return err
	}

	user, err := state.DB.GetUserByAccountID(ctx, account.ID)
	if err != nil {
		return err
	}

	user.Approved = func() *bool { a := true; return &a }()
	user.Email = user.UnconfirmedEmail
	user.ConfirmedAt = time.Now()
	user.SignUpIP = nil
	return state.DB.UpdateUser(
		ctx, user,
		"approved",
		"email",
		"confirmed_at",
		"sign_up_ip",
	)
}

// Promote sets admin + moderator flags on a user to true.
var Promote action.GTSAction = func(ctx context.Context) error {
	state, err := initState(ctx)
	if err != nil {
		return err
	}

	defer func() {
		// Ensure state gets stopped on return.
		if err := stopState(state); err != nil {
			log.Error(ctx, err)
		}
	}()

	username := config.GetAdminAccountUsername()
	if err := validate.Username(username); err != nil {
		return err
	}

	account, err := state.DB.GetAccountByUsernameDomain(ctx, username, "")
	if err != nil {
		return err
	}

	user, err := state.DB.GetUserByAccountID(ctx, account.ID)
	if err != nil {
		return err
	}

	user.Admin = func() *bool { a := true; return &a }()
	user.Moderator = func() *bool { a := true; return &a }()
	return state.DB.UpdateUser(
		ctx, user,
		"admin", "moderator",
	)
}

// Demote sets admin + moderator flags on a user to false.
var Demote action.GTSAction = func(ctx context.Context) error {
	state, err := initState(ctx)
	if err != nil {
		return err
	}

	defer func() {
		// Ensure state gets stopped on return.
		if err := stopState(state); err != nil {
			log.Error(ctx, err)
		}
	}()

	username := config.GetAdminAccountUsername()
	if err := validate.Username(username); err != nil {
		return err
	}

	a, err := state.DB.GetAccountByUsernameDomain(ctx, username, "")
	if err != nil {
		return err
	}

	user, err := state.DB.GetUserByAccountID(ctx, a.ID)
	if err != nil {
		return err
	}

	user.Admin = func() *bool { a := false; return &a }()
	user.Moderator = func() *bool { a := false; return &a }()
	return state.DB.UpdateUser(
		ctx, user,
		"admin", "moderator",
	)
}

// Disable sets Disabled to true on a user.
var Disable action.GTSAction = func(ctx context.Context) error {
	state, err := initState(ctx)
	if err != nil {
		return err
	}

	defer func() {
		// Ensure state gets stopped on return.
		if err := stopState(state); err != nil {
			log.Error(ctx, err)
		}
	}()

	username := config.GetAdminAccountUsername()
	if err := validate.Username(username); err != nil {
		return err
	}

	account, err := state.DB.GetAccountByUsernameDomain(ctx, username, "")
	if err != nil {
		return err
	}

	user, err := state.DB.GetUserByAccountID(ctx, account.ID)
	if err != nil {
		return err
	}

	user.Disabled = util.Ptr(true)
	return state.DB.UpdateUser(
		ctx, user,
		"disabled",
	)
}

// Enable sets Disabled to false on a user.
var Enable action.GTSAction = func(ctx context.Context) error {
	state, err := initState(ctx)
	if err != nil {
		return err
	}

	defer func() {
		// Ensure state gets stopped on return.
		if err := stopState(state); err != nil {
			log.Error(ctx, err)
		}
	}()

	username := config.GetAdminAccountUsername()
	if err := validate.Username(username); err != nil {
		return err
	}

	account, err := state.DB.GetAccountByUsernameDomain(ctx, username, "")
	if err != nil {
		return err
	}

	user, err := state.DB.GetUserByAccountID(ctx, account.ID)
	if err != nil {
		return err
	}

	user.Disabled = util.Ptr(false)
	return state.DB.UpdateUser(
		ctx, user,
		"disabled",
	)
}

// Password sets the password of target account.
var Password action.GTSAction = func(ctx context.Context) error {
	state, err := initState(ctx)
	if err != nil {
		return err
	}

	defer func() {
		// Ensure state gets stopped on return.
		if err := stopState(state); err != nil {
			log.Error(ctx, err)
		}
	}()

	username := config.GetAdminAccountUsername()
	if err := validate.Username(username); err != nil {
		return err
	}

	password := config.GetAdminAccountPassword()
	if err := validate.Password(password); err != nil {
		return err
	}

	account, err := state.DB.GetAccountByUsernameDomain(ctx, username, "")
	if err != nil {
		return err
	}

	user, err := state.DB.GetUserByAccountID(ctx, account.ID)
	if err != nil {
		return err
	}

	encryptedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("error hashing password: %s", err)
	}

	user.EncryptedPassword = string(encryptedPassword)
	log.Info(ctx, "Updating password; you must restart the server to use the new password.")
	return state.DB.UpdateUser(
		ctx, user,
		"encrypted_password",
	)
}

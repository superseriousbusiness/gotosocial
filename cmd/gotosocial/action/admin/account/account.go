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
*/

package account

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/superseriousbusiness/gotosocial/cmd/gotosocial/action"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db/bundb"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/validate"
	"golang.org/x/crypto/bcrypt"
)

// Create creates a new account in the database using the provided flags.
var Create action.GTSAction = func(ctx context.Context) error {
	var state state.State
	state.Caches.Init()

	dbConn, err := bundb.NewBunDBService(ctx, &state)
	if err != nil {
		return fmt.Errorf("error creating dbservice: %s", err)
	}

	// Set the state DB connection
	state.DB = dbConn

	username := config.GetAdminAccountUsername()
	if username == "" {
		return errors.New("no username set")
	}
	if err := validate.Username(username); err != nil {
		return err
	}

	usernameAvailable, err := dbConn.IsUsernameAvailable(ctx, username)
	if err != nil {
		return err
	}
	if !usernameAvailable {
		return fmt.Errorf("username %s is already in use", username)
	}

	email := config.GetAdminAccountEmail()
	if email == "" {
		return errors.New("no email set")
	}
	if err := validate.Email(email); err != nil {
		return err
	}

	emailAvailable, err := dbConn.IsEmailAvailable(ctx, email)
	if err != nil {
		return err
	}
	if !emailAvailable {
		return fmt.Errorf("email address %s is already in use", email)
	}

	password := config.GetAdminAccountPassword()
	if password == "" {
		return errors.New("no password set")
	}
	if err := validate.NewPassword(password); err != nil {
		return err
	}

	_, err = dbConn.NewSignup(ctx, username, "", false, email, password, nil, "", "", true, "", false)
	if err != nil {
		return err
	}

	return dbConn.Stop(ctx)
}

// Confirm sets a user to Approved, sets Email to the current UnconfirmedEmail value, and sets ConfirmedAt to now.
var Confirm action.GTSAction = func(ctx context.Context) error {
	var state state.State
	state.Caches.Init()

	dbConn, err := bundb.NewBunDBService(ctx, &state)
	if err != nil {
		return fmt.Errorf("error creating dbservice: %s", err)
	}

	// Set the state DB connection
	state.DB = dbConn

	username := config.GetAdminAccountUsername()
	if username == "" {
		return errors.New("no username set")
	}
	if err := validate.Username(username); err != nil {
		return err
	}

	a, err := dbConn.GetAccountByUsernameDomain(ctx, username, "")
	if err != nil {
		return err
	}

	u, err := dbConn.GetUserByAccountID(ctx, a.ID)
	if err != nil {
		return err
	}

	updatingColumns := []string{"approved", "email", "confirmed_at"}
	approved := true
	u.Approved = &approved
	u.Email = u.UnconfirmedEmail
	u.ConfirmedAt = time.Now()
	if err := dbConn.UpdateUser(ctx, u, updatingColumns...); err != nil {
		return err
	}

	return dbConn.Stop(ctx)
}

// Promote sets a user to admin.
var Promote action.GTSAction = func(ctx context.Context) error {
	var state state.State
	state.Caches.Init()

	dbConn, err := bundb.NewBunDBService(ctx, &state)
	if err != nil {
		return fmt.Errorf("error creating dbservice: %s", err)
	}

	// Set the state DB connection
	state.DB = dbConn

	username := config.GetAdminAccountUsername()
	if username == "" {
		return errors.New("no username set")
	}
	if err := validate.Username(username); err != nil {
		return err
	}

	a, err := dbConn.GetAccountByUsernameDomain(ctx, username, "")
	if err != nil {
		return err
	}

	u, err := dbConn.GetUserByAccountID(ctx, a.ID)
	if err != nil {
		return err
	}

	admin := true
	u.Admin = &admin
	if err := dbConn.UpdateUser(ctx, u, "admin"); err != nil {
		return err
	}

	return dbConn.Stop(ctx)
}

// Demote sets admin on a user to false.
var Demote action.GTSAction = func(ctx context.Context) error {
	var state state.State
	state.Caches.Init()

	dbConn, err := bundb.NewBunDBService(ctx, &state)
	if err != nil {
		return fmt.Errorf("error creating dbservice: %s", err)
	}

	// Set the state DB connection
	state.DB = dbConn

	username := config.GetAdminAccountUsername()
	if username == "" {
		return errors.New("no username set")
	}
	if err := validate.Username(username); err != nil {
		return err
	}

	a, err := dbConn.GetAccountByUsernameDomain(ctx, username, "")
	if err != nil {
		return err
	}

	u, err := dbConn.GetUserByAccountID(ctx, a.ID)
	if err != nil {
		return err
	}

	admin := false
	u.Admin = &admin
	if err := dbConn.UpdateUser(ctx, u, "admin"); err != nil {
		return err
	}

	return dbConn.Stop(ctx)
}

// Disable sets Disabled to true on a user.
var Disable action.GTSAction = func(ctx context.Context) error {
	var state state.State
	state.Caches.Init()

	dbConn, err := bundb.NewBunDBService(ctx, &state)
	if err != nil {
		return fmt.Errorf("error creating dbservice: %s", err)
	}

	// Set the state DB connection
	state.DB = dbConn

	username := config.GetAdminAccountUsername()
	if username == "" {
		return errors.New("no username set")
	}
	if err := validate.Username(username); err != nil {
		return err
	}

	a, err := dbConn.GetAccountByUsernameDomain(ctx, username, "")
	if err != nil {
		return err
	}

	u, err := dbConn.GetUserByAccountID(ctx, a.ID)
	if err != nil {
		return err
	}

	disabled := true
	u.Disabled = &disabled
	if err := dbConn.UpdateUser(ctx, u, "disabled"); err != nil {
		return err
	}

	return dbConn.Stop(ctx)
}

// Password sets the password of target account.
var Password action.GTSAction = func(ctx context.Context) error {
	var state state.State
	state.Caches.Init()

	dbConn, err := bundb.NewBunDBService(ctx, &state)
	if err != nil {
		return fmt.Errorf("error creating dbservice: %s", err)
	}

	// Set the state DB connection
	state.DB = dbConn

	username := config.GetAdminAccountUsername()
	if username == "" {
		return errors.New("no username set")
	}
	if err := validate.Username(username); err != nil {
		return err
	}

	password := config.GetAdminAccountPassword()
	if password == "" {
		return errors.New("no password set")
	}
	if err := validate.NewPassword(password); err != nil {
		return err
	}

	a, err := dbConn.GetAccountByUsernameDomain(ctx, username, "")
	if err != nil {
		return err
	}

	u, err := dbConn.GetUserByAccountID(ctx, a.ID)
	if err != nil {
		return err
	}

	pw, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("error hashing password: %s", err)
	}

	u.EncryptedPassword = string(pw)
	if err := dbConn.UpdateUser(ctx, u, "encrypted_password"); err != nil {
		return err
	}

	return nil
}

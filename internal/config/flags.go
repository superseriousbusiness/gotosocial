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

package config

import (
	"github.com/spf13/cobra"
)

// TODO: consolidate these methods into the Configuration{} or ConfigState{} structs.

// AddAdminAccount attaches flags pertaining to admin account actions.
func AddAdminAccount(cmd *cobra.Command) {
	name := AdminAccountUsernameFlag()
	usage := fieldtag("AdminAccountUsername", "usage")
	cmd.Flags().String(name, "", usage) // REQUIRED
	if err := cmd.MarkFlagRequired(name); err != nil {
		panic(err)
	}
}

// AddAdminAccountPassword attaches flags pertaining to admin account password reset.
func AddAdminAccountPassword(cmd *cobra.Command) {
	name := AdminAccountPasswordFlag()
	usage := fieldtag("AdminAccountPassword", "usage")
	cmd.Flags().String(name, "", usage) // REQUIRED
	if err := cmd.MarkFlagRequired(name); err != nil {
		panic(err)
	}
}

// AddAdminAccountCreate attaches flags pertaining to admin account creation.
func AddAdminAccountCreate(cmd *cobra.Command) {
	// Requires both account and password
	AddAdminAccount(cmd)
	AddAdminAccountPassword(cmd)

	name := AdminAccountEmailFlag()
	usage := fieldtag("AdminAccountEmail", "usage")
	cmd.Flags().String(name, "", usage) // REQUIRED
	if err := cmd.MarkFlagRequired(name); err != nil {
		panic(err)
	}
}

// AddAdminTrans attaches flags pertaining to import/export commands.
func AddAdminTrans(cmd *cobra.Command) {
	name := AdminTransPathFlag()
	usage := fieldtag("AdminTransPath", "usage")
	cmd.Flags().String(name, "", usage) // REQUIRED
	if err := cmd.MarkFlagRequired(name); err != nil {
		panic(err)
	}
}

// AddAdminMediaList attaches flags pertaining to media list commands.
func AddAdminMediaList(cmd *cobra.Command) {
	localOnly := AdminMediaListLocalOnlyFlag()
	localOnlyUsage := fieldtag("AdminMediaListLocalOnly", "usage")
	cmd.Flags().Bool(localOnly, false, localOnlyUsage)

	remoteOnly := AdminMediaListRemoteOnlyFlag()
	remoteOnlyUsage := fieldtag("AdminMediaListRemoteOnly", "usage")
	cmd.Flags().Bool(remoteOnly, false, remoteOnlyUsage)
}

// AddAdminMediaPrune attaches flags pertaining to media storage prune commands.
func AddAdminMediaPrune(cmd *cobra.Command) {
	name := AdminMediaPruneDryRunFlag()
	usage := fieldtag("AdminMediaPruneDryRun", "usage")
	cmd.Flags().Bool(name, true, usage)
}

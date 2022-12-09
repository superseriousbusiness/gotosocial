/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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

package main

import (
	"github.com/spf13/cobra"
	"github.com/superseriousbusiness/gotosocial/cmd/gotosocial/action/admin/account"
	"github.com/superseriousbusiness/gotosocial/cmd/gotosocial/action/admin/media/prune"
	"github.com/superseriousbusiness/gotosocial/cmd/gotosocial/action/admin/trans"
	"github.com/superseriousbusiness/gotosocial/internal/config"
)

func adminCommands() *cobra.Command {
	adminCmd := &cobra.Command{
		Use:   "admin",
		Short: "gotosocial admin-related tasks",
	}

	/*
	   ADMIN ACCOUNT COMMANDS
	*/

	adminAccountCmd := &cobra.Command{
		Use:   "account",
		Short: "admin commands related to local (this instance) accounts",
	}
	config.AddAdminAccount(adminAccountCmd)

	adminAccountCreateCmd := &cobra.Command{
		Use:   "create",
		Short: "create a new local account",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return preRun(preRunArgs{cmd: cmd})
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), account.Create)
		},
	}
	config.AddAdminAccountCreate(adminAccountCreateCmd)
	adminAccountCmd.AddCommand(adminAccountCreateCmd)

	adminAccountConfirmCmd := &cobra.Command{
		Use:   "confirm",
		Short: "confirm an existing local account manually, thereby skipping email confirmation",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return preRun(preRunArgs{cmd: cmd})
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), account.Confirm)
		},
	}
	config.AddAdminAccount(adminAccountConfirmCmd)
	adminAccountCmd.AddCommand(adminAccountConfirmCmd)

	adminAccountPromoteCmd := &cobra.Command{
		Use:   "promote",
		Short: "promote a local account to admin",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return preRun(preRunArgs{cmd: cmd})
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), account.Promote)
		},
	}
	config.AddAdminAccount(adminAccountPromoteCmd)
	adminAccountCmd.AddCommand(adminAccountPromoteCmd)

	adminAccountDemoteCmd := &cobra.Command{
		Use:   "demote",
		Short: "demote a local account from admin to normal user",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return preRun(preRunArgs{cmd: cmd})
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), account.Demote)
		},
	}
	config.AddAdminAccount(adminAccountDemoteCmd)
	adminAccountCmd.AddCommand(adminAccountDemoteCmd)

	adminAccountDisableCmd := &cobra.Command{
		Use:   "disable",
		Short: "prevent a local account from signing in or posting etc, but don't delete anything",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return preRun(preRunArgs{cmd: cmd})
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), account.Disable)
		},
	}
	config.AddAdminAccount(adminAccountDisableCmd)
	adminAccountCmd.AddCommand(adminAccountDisableCmd)

	adminAccountPasswordCmd := &cobra.Command{
		Use:   "password",
		Short: "set a new password for the given local account",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return preRun(preRunArgs{cmd: cmd})
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), account.Password)
		},
	}
	config.AddAdminAccount(adminAccountPasswordCmd)
	config.AddAdminAccountPassword(adminAccountPasswordCmd)
	adminAccountCmd.AddCommand(adminAccountPasswordCmd)

	adminCmd.AddCommand(adminAccountCmd)

	/*
	   ADMIN IMPORT/EXPORT COMMANDS
	*/

	adminExportCmd := &cobra.Command{
		Use:   "export",
		Short: "export data from the database to file at the given path",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return preRun(preRunArgs{cmd: cmd})
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), trans.Export)
		},
	}
	config.AddAdminTrans(adminExportCmd)
	adminCmd.AddCommand(adminExportCmd)

	adminImportCmd := &cobra.Command{
		Use:   "import",
		Short: "import data from a file into the database",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return preRun(preRunArgs{cmd: cmd})
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), trans.Import)
		},
	}
	config.AddAdminTrans(adminImportCmd)
	adminCmd.AddCommand(adminImportCmd)

	/*
		ADMIN MEDIA COMMANDS
	*/

	adminMediaCmd := &cobra.Command{
		Use:   "media",
		Short: "admin commands related stored media attachments/emojis",
	}

	/*
		ADMIN MEDIA PRUNE COMMANDS
	*/
	adminMediaPruneCmd := &cobra.Command{
		Use:   "prune",
		Short: "admin commands for pruning unused/orphaned media from storage",
	}
	config.AddAdminMediaPrune(adminMediaPruneCmd)

	adminMediaPruneOrphanedCmd := &cobra.Command{
		Use:   "orphaned",
		Short: "prune orphaned media from storage",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return preRun(preRunArgs{cmd: cmd})
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), prune.Orphaned)
		},
	}
	config.AddAdminMediaPrune(adminMediaPruneOrphanedCmd)
	adminMediaPruneCmd.AddCommand(adminMediaPruneOrphanedCmd)

	adminMediaCmd.AddCommand(adminMediaPruneCmd)

	adminCmd.AddCommand(adminMediaCmd)

	return adminCmd
}

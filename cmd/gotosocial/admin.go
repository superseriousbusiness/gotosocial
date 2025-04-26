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

package main

import (
	"code.superseriousbusiness.org/gotosocial/cmd/gotosocial/action/admin/account"
	"code.superseriousbusiness.org/gotosocial/cmd/gotosocial/action/admin/media"
	"code.superseriousbusiness.org/gotosocial/cmd/gotosocial/action/admin/media/prune"
	"code.superseriousbusiness.org/gotosocial/cmd/gotosocial/action/admin/trans"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"github.com/spf13/cobra"
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

	adminAccountListCmd := &cobra.Command{
		Use:   "list",
		Short: "list all existing local accounts",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return preRun(preRunArgs{cmd: cmd})
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), account.List)
		},
	}
	adminAccountCmd.AddCommand(adminAccountListCmd)

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
		Short: "set 'disabled' to true on a local account to prevent it from signing in or posting etc, but don't delete anything",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return preRun(preRunArgs{cmd: cmd})
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), account.Disable)
		},
	}
	config.AddAdminAccount(adminAccountDisableCmd)
	adminAccountCmd.AddCommand(adminAccountDisableCmd)

	adminAccountEnableCmd := &cobra.Command{
		Use:   "enable",
		Short: "undo a previous disable command by setting 'disabled' to false on a local account",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return preRun(preRunArgs{cmd: cmd})
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), account.Enable)
		},
	}
	config.AddAdminAccount(adminAccountEnableCmd)
	adminAccountCmd.AddCommand(adminAccountEnableCmd)

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
		Short: "admin commands related to stored media / emojis",
	}

	/*
		ADMIN MEDIA LIST COMMANDS
	*/

	adminMediaListAttachmentsCmd := &cobra.Command{
		Use:   "list-attachments",
		Short: "list local, remote, or all attachments",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return preRun(preRunArgs{cmd: cmd})
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), media.ListAttachments)
		},
	}
	config.AddAdminMediaList(adminMediaListAttachmentsCmd)
	adminMediaCmd.AddCommand(adminMediaListAttachmentsCmd)

	adminMediaListEmojisLocalCmd := &cobra.Command{
		Use:   "list-emojis",
		Short: "list local, remote, or all emojis",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return preRun(preRunArgs{cmd: cmd})
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), media.ListEmojis)
		},
	}
	config.AddAdminMediaList(adminMediaListEmojisLocalCmd)
	adminMediaCmd.AddCommand(adminMediaListEmojisLocalCmd)

	/*
		ADMIN MEDIA PRUNE COMMANDS
	*/
	adminMediaPruneCmd := &cobra.Command{
		Use:   "prune",
		Short: "admin commands for pruning media from storage",
	}

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

	adminMediaPruneRemoteCmd := &cobra.Command{
		Use:   "remote",
		Short: "prune unused / stale media from storage, older than given number of days",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return preRun(preRunArgs{cmd: cmd})
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), prune.Remote)
		},
	}
	config.AddAdminMediaPrune(adminMediaPruneRemoteCmd)
	adminMediaPruneCmd.AddCommand(adminMediaPruneRemoteCmd)

	adminMediaPruneAllCmd := &cobra.Command{
		Use:   "all",
		Short: "perform all media and emoji prune / cleaning commands",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return preRun(preRunArgs{cmd: cmd})
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), prune.All)
		},
	}
	config.AddAdminMediaPrune(adminMediaPruneAllCmd)
	adminMediaPruneCmd.AddCommand(adminMediaPruneAllCmd)

	adminMediaCmd.AddCommand(adminMediaPruneCmd)

	adminCmd.AddCommand(adminMediaCmd)

	return adminCmd
}

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
	"code.superseriousbusiness.org/gotosocial/cmd/gotosocial/action/migration"
	"github.com/spf13/cobra"
)

// migrationCommands returns the 'migrations' subcommand
func migrationCommands() *cobra.Command {
	migrationCmd := &cobra.Command{
		Use:   "migrations",
		Short: "gotosocial migrations-related tasks",
	}
	migrationRunCmd := &cobra.Command{
		Use:   "run",
		Short: "starts and stops the database, running any outstanding migrations",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return preRun(preRunArgs{cmd: cmd})
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), migration.Run)
		},
	}
	migrationCmd.AddCommand(migrationRunCmd)
	return migrationCmd
}

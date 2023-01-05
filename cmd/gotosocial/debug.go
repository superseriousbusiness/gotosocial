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

package main

import (
	"github.com/spf13/cobra"
	configaction "github.com/superseriousbusiness/gotosocial/cmd/gotosocial/action/debug/config"
	"github.com/superseriousbusiness/gotosocial/internal/config"
)

func debugCommands() *cobra.Command {
	debugCmd := &cobra.Command{
		Use:   "debug",
		Short: "gotosocial debug-related tasks",
	}

	debugConfigCmd := &cobra.Command{
		Use:   "config",
		Short: "print the collated config (derived from env, flag, and config file) to stdout",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return preRun(preRunArgs{cmd: cmd, skipValidation: true}) // don't do validation for debugging config
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), configaction.Config)
		},
	}
	config.AddServerFlags(debugConfigCmd)
	debugCmd.AddCommand(debugConfigCmd)
	return debugCmd
}

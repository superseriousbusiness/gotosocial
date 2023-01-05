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
	"github.com/superseriousbusiness/gotosocial/cmd/gotosocial/action/server"
	"github.com/superseriousbusiness/gotosocial/internal/config"
)

// serverCommands returns the 'server' subcommand
func serverCommands() *cobra.Command {
	serverCmd := &cobra.Command{
		Use:   "server",
		Short: "gotosocial server-related tasks",
	}
	serverStartCmd := &cobra.Command{
		Use:   "start",
		Short: "start the gotosocial server",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return preRun(preRunArgs{cmd: cmd})
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), server.Start)
		},
	}
	config.AddServerFlags(serverStartCmd)
	serverCmd.AddCommand(serverStartCmd)
	return serverCmd
}

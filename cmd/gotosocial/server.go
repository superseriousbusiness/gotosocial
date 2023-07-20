// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-FileCopyrightText: 2023 GoToSocial Authors <admin@gotosocial.org>
//
// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-License-Identifier: AGPL-3.0-or-later

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

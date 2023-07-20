// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-FileCopyrightText: 2023 GoToSocial Authors <admin@gotosocial.org>
//
// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-License-Identifier: AGPL-3.0-or-later

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

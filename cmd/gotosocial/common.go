// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-FileCopyrightText: 2023 GoToSocial Authors <admin@gotosocial.org>
//
// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/superseriousbusiness/gotosocial/cmd/gotosocial/action"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

type preRunArgs struct {
	cmd            *cobra.Command
	skipValidation bool
}

// preRun should be run in the pre-run stage of every cobra command.
// The goal here is to initialize the viper config store, and also read in
// the config file (if present).
//
// Config then undergoes basic validation if 'skipValidation' is not true.
//
// The order of these is important: the init-config function reads the location
// of the config file from the viper store so that it can be picked up by either
// env vars or cli flag.
func preRun(a preRunArgs) error {
	if err := config.BindFlags(a.cmd); err != nil {
		return fmt.Errorf("error binding flags: %s", err)
	}

	if err := config.Reload(); err != nil {
		return fmt.Errorf("error reloading config: %s", err)
	}

	if !a.skipValidation {
		if err := config.Validate(); err != nil {
			return fmt.Errorf("invalid config: %s", err)
		}
	}

	return nil
}

// run should be used during the run stage of every cobra command.
// The idea here is to take a GTSAction and run it with the given
// context, after initializing any last-minute things like loggers etc.
func run(ctx context.Context, action action.GTSAction) error {
	// Set the global log level from configuration
	if err := log.ParseLevel(config.GetLogLevel()); err != nil {
		return fmt.Errorf("error parsing log level: %w", err)
	}

	if config.GetSyslogEnabled() {
		// Enable logging to syslog
		if err := log.EnableSyslog(
			config.GetSyslogProtocol(),
			config.GetSyslogAddress(),
		); err != nil {
			return fmt.Errorf("error enabling syslogging: %w", err)
		}
	}

	return action(ctx)
}

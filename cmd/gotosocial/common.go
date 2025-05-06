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
	"context"
	"fmt"

	"code.superseriousbusiness.org/gotosocial/cmd/gotosocial/action"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"github.com/spf13/cobra"
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
		return fmt.Errorf("error binding flags: %w", err)
	}

	if err := config.LoadConfigFile(); err != nil {
		return fmt.Errorf("error loading config file: %w", err)
	}

	if !a.skipValidation {
		if err := config.Validate(); err != nil {
			return fmt.Errorf("invalid config: %w", err)
		}
	}

	return nil
}

// run should be used during the run stage of every cobra command.
// The idea here is to take a GTSAction and run it with the given
// context, after initializing any last-minute things like loggers etc.
func run(ctx context.Context, action action.GTSAction) error {
	log.SetTimeFormat(config.GetLogTimestampFormat())
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

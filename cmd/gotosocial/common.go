/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

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
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/superseriousbusiness/gotosocial/internal/cliactions"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

// preRun should be run in the pre-run stage of every cobra command.
// The goal here is to initialize the viper config store, and also read in
// the config file (if present).
//
// The order of these is important: the init-config function reads the location
// of the config file from the viper store so that it can be picked up by either
// env vars or cli flags.
func preRun(cmd *cobra.Command, version string) error {
	if err := config.InitViper(cmd.Flags(), version); err != nil {
		return err
	}
	if err := config.InitConfig(); err != nil {
		return err
	}

	return nil
}

// run should be used during the run stage of every cobra command.
// The idea here is to take a GTSAction and run it with the given
// context, after initializing any last-minute things like loggers etc.
func run(ctx context.Context, action cliactions.GTSAction) error {
	if err := log.Initialize(viper.GetString(config.FlagNames.LogLevel)); err != nil {
		return err
	}

	if err := action(ctx); err != nil {
		return err
	}

	return nil
}

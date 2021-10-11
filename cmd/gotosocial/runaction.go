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
	"fmt"

	"github.com/superseriousbusiness/gotosocial/internal/cliactions"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/urfave/cli/v2"
)

// runAction builds up the config and logger necessary for any
// gotosocial action, and then executes the action.
func runAction(c *cli.Context, a cliactions.GTSAction) error {

	// create a new *config.Config based on the config path provided...
	conf, err := config.FromFile(c.String(config.GetFlagNames().ConfigPath))
	if err != nil {
		return fmt.Errorf("error creating config: %s", err)
	}
	// ... and the flags set on the *cli.Context by urfave
	if err := conf.ParseCLIFlags(c, c.App.Version); err != nil {
		return fmt.Errorf("error parsing config: %s", err)
	}

	// initialize the global logger to the log level, with formatting and output splitter already set
	err = log.Initialize(conf.LogLevel)
	if err != nil {
		return fmt.Errorf("error creating logger: %s", err)
	}

	return a(c.Context, conf)
}

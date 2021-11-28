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
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	_ "github.com/superseriousbusiness/gotosocial/docs"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

// Version is the software version of GtS being used
var Version string

// Commit is the git commit of GtS being used
var Commit string

//go:generate swagger generate spec
func main() {
	var v string
	if Commit == "" {
		v = Version
	} else {
		v = Version + " " + Commit[:7]
	}

	// instantiate the root command
	cmd := &cobra.Command{
		Use:     "gotosocial",
		Short:   "a fediverse social media server",
		Version: v,
	}

	// add subcommands
	cmd.AddCommand(serverCommands())

	// initialize viper config
	if err := config.InitViper(cmd.Flags(), v); err != nil {
		logrus.Fatalf("error initializing config: %s", err)
	}

	// initialize the global logger to the provided log level, with formatting and output splitter already set
	if err := log.Initialize(viper.GetString(config.FlagNames.LogLevel)); err != nil {
		logrus.Fatalf("error creating logger: %s", err)
	}

	// run the damn diggity thing
	if err := cmd.Execute(); err != nil {
		logrus.Fatalf("error executing command: %s", err)
	}
}

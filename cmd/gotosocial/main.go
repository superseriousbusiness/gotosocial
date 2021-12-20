/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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

	"github.com/superseriousbusiness/gotosocial/cmd/gotosocial/flag"
	_ "github.com/superseriousbusiness/gotosocial/docs"
	"github.com/superseriousbusiness/gotosocial/internal/config"
)

// Version is the software version of GtS being used.
var Version string

// Commit is the git commit of GtS being used.
var Commit string

//go:generate swagger generate spec
func main() {
	var v string
	if len(Commit) < 7 {
		v = Version
	} else {
		v = Version + " " + Commit[:7]
	}

	// override software version in viper store
	viper.Set(config.Keys.SoftwareVersion, v)

	// instantiate the root command
	rootCmd := &cobra.Command{
		Use:           "gotosocial",
		Short:         "GoToSocial - a fediverse social media server",
		Long:          "GoToSocial - a fediverse social media server\n\nFor help, see: https://docs.gotosocial.org.\n\nCode: https://github.com/superseriousbusiness/gotosocial",
		Version:       v,
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	// attach global flags to the root command so that they can be accessed from any subcommand
	flag.Global(rootCmd, config.Defaults)

	// bind the config-path flag to viper early so that we can call it in the pre-run of following commands
	if err := viper.BindPFlag(config.Keys.ConfigPath, rootCmd.PersistentFlags().Lookup(config.Keys.ConfigPath)); err != nil {
		logrus.Fatalf("error attaching config flag: %s", err)
	}

	// add subcommands
	rootCmd.AddCommand(serverCommands())
	rootCmd.AddCommand(testrigCommands())
	rootCmd.AddCommand(debugCommands())
	rootCmd.AddCommand(adminCommands())

	// run
	if err := rootCmd.Execute(); err != nil {
		logrus.Fatalf("error executing command: %s", err)
	}
}

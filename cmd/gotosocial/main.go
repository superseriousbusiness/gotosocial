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
	"runtime/debug"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	_ "github.com/superseriousbusiness/gotosocial/docs"
	"github.com/superseriousbusiness/gotosocial/internal/config"
)

// Version is the version of GoToSocial being used.
// It's injected into the binary by the build script.
var Version string

//go:generate swagger generate spec
func main() {
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		panic("could not read buildinfo")
	}

	var commit string
	for _, s := range buildInfo.Settings {
		if s.Key == "vcs.revision" {
			commit = s.Value[:7]
			continue
		}
	}

	var versionString string

	if Version != "" {
		Version = "devel"
	}

	// override software version in config store
	config.SetSoftwareVersion(Version + " " + commit)

	// instantiate the root command
	rootCmd := &cobra.Command{
		Use:     "gotosocial",
		Short:   "GoToSocial - a fediverse social media server",
		Long:    "GoToSocial - a fediverse social media server\n\nFor help, see: https://docs.gotosocial.org.\n\nCode: https://github.com/superseriousbusiness/gotosocial",
		Version: versionString,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// before running any other cmd funcs, we must load config-path
			return config.LoadEarlyFlags(cmd)
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	// attach global flags to the root command so that they can be accessed from any subcommand
	config.AddGlobalFlags(rootCmd)

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

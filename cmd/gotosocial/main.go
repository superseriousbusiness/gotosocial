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
	"log"
	"runtime/debug"
	"strings"

	"github.com/spf13/cobra"

	_ "github.com/superseriousbusiness/gotosocial/docs"
	"github.com/superseriousbusiness/gotosocial/internal/config"
)

// Version is the version of GoToSocial being used.
// It's injected into the binary by the build script.
var Version string

//go:generate swagger generate spec
func main() {
	// Load version string
	version := version()

	// override version in config store
	config.SetSoftwareVersion(version)

	// instantiate the root command
	rootCmd := &cobra.Command{
		Use:     "gotosocial",
		Short:   "GoToSocial - a fediverse social media server",
		Long:    "GoToSocial - a fediverse social media server\n\nFor help, see: https://docs.gotosocial.org.\n\nCode: https://github.com/superseriousbusiness/gotosocial",
		Version: version,
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
		log.Fatalf("error executing command: %s", err)
	}
}

// version will build a version string from binary's stored build information.
func version() string {
	// Read build information from binary
	build, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}

	// Define easy getter to fetch build settings
	getSetting := func(key string) string {
		for i := 0; i < len(build.Settings); i++ {
			if build.Settings[i].Key == key {
				return build.Settings[i].Value
			}
		}
		return ""
	}

	var info []string

	if Version != "" {
		// Append version if set
		info = append(info, Version)
	}

	if vcs := getSetting("vcs"); vcs != "" {
		// A VCS type was set (99.9% probably git)

		if commit := getSetting("vcs.revision"); commit != "" {
			if len(commit) > 7 {
				// Truncate commit
				commit = commit[:7]
			}

			// Append VCS + commit if set
			info = append(info, vcs+"-"+commit)
		}
	}

	return strings.Join(info, " ")
}

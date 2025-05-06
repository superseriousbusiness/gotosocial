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
	"log"
	"os"
	godebug "runtime/debug"
	"strings"

	_ "code.superseriousbusiness.org/gotosocial/docs"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"github.com/spf13/cobra"
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

	rootCmd := new(cobra.Command)
	rootCmd.Use = "gotosocial"
	rootCmd.Short = "GoToSocial - a fediverse social media server"
	rootCmd.Long = "GoToSocial - a fediverse social media server\n\nFor help, see: https://docs.gotosocial.org.\n\nCode: https://codeberg.org/superseriousbusiness/gotosocial"
	rootCmd.Version = version
	rootCmd.SilenceErrors = true
	rootCmd.SilenceUsage = true

	// Register global flags with root.
	config.RegisterGlobalFlags(rootCmd)

	// Add subcommands with their flags.
	rootCmd.AddCommand(serverCommands())
	rootCmd.AddCommand(debugCommands())
	rootCmd.AddCommand(adminCommands())

	// Testrigcmd will only be set when debug is enabled.
	if testrigCmd := testrigCommands(); testrigCmd != nil {
		rootCmd.AddCommand(testrigCmd)
	} else if len(os.Args) > 1 && os.Args[1] == "testrig" {
		log.Fatal("gotosocial must be built and run with the DEBUG enviroment variable set to enable and access testrig")
	}

	// Run the prepared root command.
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("error executing command: %s", err)
	}
}

// version will build a version string from binary's stored build information.
// It is SemVer-compatible so long as Version is SemVer-compatible.
func version() string {
	// Read build information from binary
	build, ok := godebug.ReadBuildInfo()
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

	return strings.Join(info, "+")
}

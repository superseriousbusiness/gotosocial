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
	"fmt"
	"runtime/debug"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/superseriousbusiness/gotosocial/cmd/gotosocial/flag"
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

	goVersion := buildInfo.GoVersion
	var commit string
	var time string
	for _, s := range buildInfo.Settings {
		if s.Key == "vcs.revision" {
			commit = s.Value[:7]
		}
		if s.Key == "vcs.time" {
			time = s.Value
		}
	}

	var versionString string
	if Version != "" {
		versionString = fmt.Sprintf("%s %s %s [%s]", Version, commit, time, goVersion)
	}

	// override software version in viper store
	viper.Set(config.Keys.SoftwareVersion, versionString)

	// instantiate the root command
	rootCmd := &cobra.Command{
		Use:           "gotosocial",
		Short:         "GoToSocial - a fediverse social media server",
		Long:          "GoToSocial - a fediverse social media server\n\nFor help, see: https://docs.gotosocial.org.\n\nCode: https://github.com/superseriousbusiness/gotosocial",
		Version:       versionString,
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

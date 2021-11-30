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
)

// Version is the software version of GtS being used
var Version string

// Commit is the git commit of GtS being used
var Commit string

//go:generate swagger generate spec
func main() {
	cobra.OnInitialize(config.InitConfig)

	var v string
	if Commit == "" {
		v = Version
	} else {
		v = Version + " " + Commit[:7]
	}

	// instantiate the root rootCommand
	rootCommand := &cobra.Command{
		Use:           "gotosocial",
		Short:         "GoToSocial - a fediverse social media server",
		Long:          "GoToSocial - a fediverse social media server\n\nFor help, see: https://docs.gotosocial.org.\n\nCode: https://github.com/superseriousbusiness/gotosocial",
		Version:       v,
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	config.AttachCommonFlags(rootCommand.PersistentFlags(), config.Defaults)
	viper.BindPFlag(config.FlagNames.ConfigPath, rootCommand.PersistentFlags().Lookup(config.FlagNames.ConfigPath))

	// add subcommands
	rootCommand.AddCommand(serverCommands(v))
	rootCommand.AddCommand(testrigCommands(v))
	rootCommand.AddCommand(debugCommands(v))

	// run the damn diggity thing
	if err := rootCommand.Execute(); err != nil {
		logrus.Fatalf("error executing command: %s", err)
	}
}

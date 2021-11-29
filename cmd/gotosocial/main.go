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

	_ "github.com/superseriousbusiness/gotosocial/docs"
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
		Use:           "gotosocial",
		Short:         "GoToSocial - a fediverse social media server",
		Long:          "GoToSocial - a fediverse social media server\n\nFor help, see: https://docs.gotosocial.org.\n\nCode: https://github.com/superseriousbusiness/gotosocial",
		Version:       v,
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	// add subcommands
	cmd.AddCommand(serverCommands(v))
	cmd.AddCommand(testrigCommands(v))

	// run the damn diggity thing
	if err := cmd.Execute(); err != nil {
		logrus.Fatalf("error executing command: %s", err)
	}
}

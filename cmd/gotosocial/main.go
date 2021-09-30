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
	"os"

	"github.com/sirupsen/logrus"

	_ "github.com/superseriousbusiness/gotosocial/docs"
	"github.com/urfave/cli/v2"
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

	app := &cli.App{
		Version:  v,
		Usage:    "a fediverse social media server",
		Flags:    getFlags(),
		Commands: getCommands(),
	}

	err := app.Run(os.Args)
	if err != nil {
		logrus.Fatal(err)
	}
}

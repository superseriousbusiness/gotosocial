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
	"code.superseriousbusiness.org/gotosocial/cmd/gotosocial/action/testrig"
	"github.com/spf13/cobra"
)

func testrigCommands() *cobra.Command {
	if testrig.Start != nil {
		testrigCmd := &cobra.Command{
			Use:   "testrig",
			Short: "gotosocial testrig-related tasks",
		}

		testrigStartCmd := &cobra.Command{
			Use:   "start",
			Short: "start the gotosocial testrig server",
			RunE: func(cmd *cobra.Command, args []string) error {
				return run(cmd.Context(), testrig.Start)
			},
		}

		testrigCmd.AddCommand(testrigStartCmd)
		return testrigCmd
	}
	return nil
}

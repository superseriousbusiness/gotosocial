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

package flag

import (
	"github.com/spf13/cobra"
	"github.com/superseriousbusiness/gotosocial/internal/config"
)

// AdminAccount attaches flags pertaining to admin account actions.
func AdminAccount(cmd *cobra.Command, values config.Values) {
	cmd.Flags().String(config.Keys.AdminAccountUsername, "", usage.AdminAccountUsername) // REQUIRED
	if err := cmd.MarkFlagRequired(config.Keys.AdminAccountUsername); err != nil {
		panic(err)
	}
}

// AdminAccountPassword attaches flags pertaining to admin account password reset.
func AdminAccountPassword(cmd *cobra.Command, values config.Values) {
	AdminAccount(cmd, values)
	cmd.Flags().String(config.Keys.AdminAccountPassword, "", usage.AdminAccountPassword) // REQUIRED
	if err := cmd.MarkFlagRequired(config.Keys.AdminAccountPassword); err != nil {
		panic(err)
	}
}

// AdminAccountCreate attaches flags pertaining to admin account creation.
func AdminAccountCreate(cmd *cobra.Command, values config.Values) {
	AdminAccount(cmd, values)
	cmd.Flags().String(config.Keys.AdminAccountPassword, "", usage.AdminAccountPassword) // REQUIRED
	if err := cmd.MarkFlagRequired(config.Keys.AdminAccountPassword); err != nil {
		panic(err)
	}
	cmd.Flags().String(config.Keys.AdminAccountEmail, "", usage.AdminAccountEmail) // REQUIRED
	if err := cmd.MarkFlagRequired(config.Keys.AdminAccountEmail); err != nil {
		panic(err)
	}
}

// AdminTrans attaches flags pertaining to import/export commands.
func AdminTrans(cmd *cobra.Command, values config.Values) {
	cmd.Flags().String(config.Keys.AdminTransPath, "", usage.AdminTransPath) // REQUIRED
	if err := cmd.MarkFlagRequired(config.Keys.AdminTransPath); err != nil {
		panic(err)
	}
}

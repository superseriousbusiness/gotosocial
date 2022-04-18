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

// Global attaches flags that are common to all commands, aka persistent commands.
func Global(cmd *cobra.Command, values config.Values) {
	// general stuff
	cmd.PersistentFlags().String(config.Keys.ApplicationName, values.ApplicationName, usage.ApplicationName)
	cmd.PersistentFlags().String(config.Keys.Host, values.Host, usage.Host)
	cmd.PersistentFlags().String(config.Keys.AccountDomain, values.AccountDomain, usage.AccountDomain)
	cmd.PersistentFlags().String(config.Keys.Protocol, values.Protocol, usage.Protocol)
	cmd.PersistentFlags().String(config.Keys.LogLevel, values.LogLevel, usage.LogLevel)
	cmd.PersistentFlags().Bool(config.Keys.LogDbQueries, values.LogDbQueries, usage.LogDbQueries)
	cmd.PersistentFlags().String(config.Keys.ConfigPath, values.ConfigPath, usage.ConfigPath)

	// database stuff
	cmd.PersistentFlags().String(config.Keys.DbType, values.DbType, usage.DbType)
	cmd.PersistentFlags().String(config.Keys.DbAddress, values.DbAddress, usage.DbAddress)
	cmd.PersistentFlags().Int(config.Keys.DbPort, values.DbPort, usage.DbPort)
	cmd.PersistentFlags().String(config.Keys.DbUser, values.DbUser, usage.DbUser)
	cmd.PersistentFlags().String(config.Keys.DbPassword, values.DbPassword, usage.DbPassword)
	cmd.PersistentFlags().String(config.Keys.DbDatabase, values.DbDatabase, usage.DbDatabase)
	cmd.PersistentFlags().String(config.Keys.DbTLSMode, values.DbTLSMode, usage.DbTLSMode)
	cmd.PersistentFlags().String(config.Keys.DbTLSCACert, values.DbTLSCACert, usage.DbTLSCACert)
}

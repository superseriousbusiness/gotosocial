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
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	configcliactions "github.com/superseriousbusiness/gotosocial/internal/cliactions/debug/config"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

func debugCommands(version string) *cobra.Command {
	command := &cobra.Command{
		Use:   "debug",
		Short: "gotosocial debug-related tasks",
	}

	conf := &cobra.Command{
		Use:   "config",
		Short: "print the collated config (derived from env, flags, and config file) to stdout",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return config.InitViper(cmd.Flags(), version)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := log.Initialize(viper.GetString(config.FlagNames.LogLevel)); err != nil {
				return err
			}
			return configcliactions.Config(cmd.Context())
		},
	}
	config.AttachServerFlags(conf.Flags(), config.Defaults)

	command.AddCommand(conf)
	return command
}

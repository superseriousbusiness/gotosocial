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
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/urfave/cli/v2"
)

func accountsFlags(flagNames, envNames config.Flags, defaults config.Defaults) []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{
			Name:    flagNames.AccountsOpenRegistration,
			Usage:   "Allow anyone to submit an account signup request. If false, server will be invite-only.",
			Value:   defaults.AccountsOpenRegistration,
			EnvVars: []string{envNames.AccountsOpenRegistration},
		},
		&cli.BoolFlag{
			Name:    flagNames.AccountsApprovalRequired,
			Usage:   "Do account signups require approval by an admin or moderator before user can log in? If false, new registrations will be automatically approved.",
			Value:   defaults.AccountsRequireApproval,
			EnvVars: []string{envNames.AccountsApprovalRequired},
		},
		&cli.BoolFlag{
			Name:    flagNames.AccountsReasonRequired,
			Usage:   "Do new account signups require a reason to be submitted on registration?",
			Value:   defaults.AccountsReasonRequired,
			EnvVars: []string{envNames.AccountsReasonRequired},
		},
	}
}

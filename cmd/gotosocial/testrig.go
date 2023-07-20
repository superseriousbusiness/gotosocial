// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-FileCopyrightText: 2023 GoToSocial Authors <admin@gotosocial.org>
//
// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"github.com/spf13/cobra"
	"github.com/superseriousbusiness/gotosocial/cmd/gotosocial/action/testrig"
)

func testrigCommands() *cobra.Command {
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

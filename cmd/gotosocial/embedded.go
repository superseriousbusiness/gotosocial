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
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/superseriousbusiness/gotosocial/cmd/gotosocial/action/embedded"
)

func embeddedCommands() *cobra.Command {
	embeddedCmd := &cobra.Command{
		Use:   "embedded",
		Short: "work with assets embedded in the gotosocial executable for customization",
	}

	embeddedListCmd := &cobra.Command{
		Use:   "list",
		Short: "generate list of assets embedded in this executable",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return preRun(preRunArgs{cmd: cmd, skipValidation: true})
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), embedded.ListEmbeddedFiles)
		},
	}
	embeddedCmd.AddCommand(embeddedListCmd)

	embeddedViewCmd := &cobra.Command{
		Use:   "view <embedded-path>",
		Short: "print content of an embedded asset",
		Args:  cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return preRun(preRunArgs{cmd: cmd, skipValidation: true})
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), func(ctx context.Context) error {
				return embedded.ViewEmbeddedFile(args[0])
			})
		},
	}
	embeddedCmd.AddCommand(embeddedViewCmd)

	var targetBaseDirFlag string
	embeddedExtractCmd := &cobra.Command{
		Use:     "extract <embedded-path>...",
		Short:   "write content of an embedded asset to a file on disk, creating a backup if the target already existed.",
		Example: "gotosocial embedded extract template/index.tmpl",
		Args:    cobra.MinimumNArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return preRun(preRunArgs{cmd: cmd, skipValidation: true})
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), func(ctx context.Context) error {
				for _, path := range args {
					fmt.Println(path)
					err := embedded.ExtractEmbeddedFile(path, targetBaseDirFlag)
					if err != nil {
						return err
					}
				}
				return nil
			})
		},
	}
	embeddedExtractCmd.Flags().StringVar(&targetBaseDirFlag, "target-base-dir", "", "Destination to write the file(s) to.\nIf left empty, files will be written to web-{template/asset}-base-dir as defined in config")
	embeddedCmd.AddCommand(embeddedExtractCmd)

	return embeddedCmd
}

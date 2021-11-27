package main

import (
	"github.com/spf13/cobra"
	"github.com/superseriousbusiness/gotosocial/internal/config"
)

func rootCommand(version string) *cobra.Command {
	command := &cobra.Command{
		Use:     "gotosocial",
		Short:   "a fediverse social media server",
		Version: version,
	}

	flagNames := config.FlagNames()
	defaults := config.Defaults()
	flagUsage := config.FlagUsage()

	command.Flags().Int(flagNames.Port, defaults.Port, flagUsage.Port)

	return command
}

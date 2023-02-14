package cli

import (
	"fmt"

	"github.com/katbyte/gogo-azurerm-info/version"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func ValidateParams(params []string) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		for _, p := range params {
			if viper.GetString(p) == "" {
				return fmt.Errorf(p + " parameter can't be empty")
			}
		}

		return nil
	}
}

func Make(cmdName string) (*cobra.Command, error) {
	// todo should this be a no-op to avoid accidentally triggering broken builds on malformed commands ?
	root := &cobra.Command{
		Use:           cmdName + " [command]",
		Short:         cmdName + "is a small utility to TODO",
		Long:          `TODO`,
		SilenceErrors: true,
		PreRunE:       ValidateParams([]string{"token", "org", "repo", "cache"}),
		RunE: func(cmd *cobra.Command, args []string) error {
			// f := GetFlags()
			// r := gh.NewRepo(f.Owner, f.Repo, f.Token)

			// what should default be?

			// fetch
			// calculate
			// stats

			// ??

			return nil
		},
	}

	// cmds:
	// report path (# services/resources pandora/track1)
	// report list [track1|parse|not-typed]
	// graph path

	// lint? probably not

	root.AddCommand(&cobra.Command{
		Use:           "report [repo path] [pandora-sdk-issue]",
		Short:         cmdName + " calculates a report for the provider (services, resources, datasources, sdk in use etc)",
		Args:          cobra.RangeArgs(1, 2),
		SilenceErrors: true,
		// PreRunE:       ValidateParams([]string{"cache"}),
		RunE: CmdReport,
	})

	root.AddCommand(&cobra.Command{
		Use:           "list [repo path] [track1|typed|create-update]",
		Short:         cmdName + " list resources that need migration",
		Args:          cobra.ExactArgs(2),
		SilenceErrors: true,
		// PreRunE:       ValidateParams([]string{"cache"}),
		RunE: CmdList,
	})

	root.AddCommand(&cobra.Command{
		Use:           "graphs",
		Args:          cobra.MaximumNArgs(2),
		SilenceErrors: true,
		RunE:          CmdGraphs,
	})

	// todo emoji stats/counter

	root.AddCommand(&cobra.Command{
		Use:           "version",
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(cmdName + " v" + version.Version + "-" + version.GitCommit)
		},
	})

	if err := configureFlags(root); err != nil {
		return nil, fmt.Errorf("unable to configure flags: %w", err)
	}

	return root, nil
}

package cli

import (
	"github.com/spf13/cobra"
)

type FlagData struct {
}

func configureFlags(root *cobra.Command) error {
	return nil
}

func GetFlags() FlagData {
	// there has to be an easier way....
	return FlagData{}
}

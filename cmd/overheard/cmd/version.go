package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of overheard",
	Long:  `All software has versions. This is overheard's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("overheard v1.0 -- HEAD")
	},
}

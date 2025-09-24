package cmd

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "overheard",
	Short: "A command-line utility for scheduling and recording audio from internet radio streams.",
	Long: `overheard is a CLI tool that allows you to schedule recordings from
internet radio streams. Configure your streams in the config file
and use the record command to start the process.`,
}

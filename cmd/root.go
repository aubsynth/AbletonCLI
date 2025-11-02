package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var debugMode bool

func init() {
	rootCmd.PersistentFlags().BoolVarP(&debugMode, "debug", "d", false, "Show debug logs")
}

var rootCmd = &cobra.Command{
	Use:     "abletoncli",
	Short:   "AbletonCLI is a command-line interface for Ableton Live",
	Long:    `AbletonCLI is a command-line interface for Ableton Live built with Cobra in Go.`,
	Version: "0.1.0",
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		cleanUpTempFiles(".")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

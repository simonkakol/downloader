package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const initialFileSize = 1

var rootCmd = &cobra.Command{
	Use:   "downloader",
	Short: "Downloader can download files from using multiple protocols",
}

// Execute executes the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

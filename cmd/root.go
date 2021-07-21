package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use: "xiaolongbaoblog",
}

// Execute the goc tool
func Execute() {
	if err := rootCmd.Execute(); err != nil {
	}
}

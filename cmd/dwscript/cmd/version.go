package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  `Display detailed version information including commit hash and build date.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("dwscript version %s\n", Version)
		fmt.Printf("Git Commit: %s\n", GitCommit)
		fmt.Printf("Build Date: %s\n", BuildDate)
		fmt.Println("\nA Go port of DWScript - https://github.com/cwbudde/go-dws")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

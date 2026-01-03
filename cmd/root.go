package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Global flags
	profile string
	region  string

	// Root command
	rootCmd = &cobra.Command{
		Use:   "dms-manager",
		Short: "Manage AWS DMS replication tasks",
		Long: `A command-line tool for managing AWS Database Migration Service (DMS) replication tasks.
		
Supports both CLI commands and an interactive TUI for listing, describing, and controlling
DMS tasks across different AWS profiles and regions.`,
	}
)

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags available to all commands
	rootCmd.PersistentFlags().StringVarP(&profile, "profile", "p", "", "AWS profile to use (default: default profile)")
	rootCmd.PersistentFlags().StringVarP(&region, "r", "", "", "AWS region (default: from profile or AWS_REGION)")
}

// GetProfile returns the global profile flag value
func GetProfile() string {
	return profile
}

// GetRegion returns the global region flag value
func GetRegion() string {
	return region
}

// exitWithError prints an error and exits
func exitWithError(err error) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	os.Exit(1)
}

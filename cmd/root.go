package cmd

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Long: "bucket is a simple docker",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(listCommand)
	rootCmd.AddCommand(logCmd)
	rootCmd.AddCommand(execCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(commitCmd)
}

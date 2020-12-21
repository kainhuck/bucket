package cmd

import (
	"bucket/container"
	"bucket/log"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "it",
	Long:  "init",
	Run: func(d *cobra.Command, args []string) {
		if err := container.RunContainerInitProcess(); err != nil {
			log.ConsoleLog.Fatal("init error: %v", err)
		}
	},
}

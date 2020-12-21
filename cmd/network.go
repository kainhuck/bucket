package cmd

import (
	"bucket/log"
	"bucket/network"
	"github.com/spf13/cobra"
)

var driver string
var subnet string

var networkCmd = &cobra.Command{
	Use: "network",
	Short: "network",
	Long: "container network op",
}

var netCreateCmd = &cobra.Command{
	Use: "create",
	Short: "create container network",
	Long: "create container network",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			log.ConsoleLog.Fatal("Missing network name")
			return
		}
		_ = network.Init()
		err := network.CreateNetwork(driver, subnet,args[0])
		if err != nil {
			log.ConsoleLog.Fatal("create network error: %+v", err)
		}
	},
}

var netListCmd = &cobra.Command{
	Use: "list",
	Short: "list container network",
	Long: "list container network",
	Aliases: []string{"ls"},
	Run: func(cmd *cobra.Command, args []string) {
		_ = network.Init()
		network.ListNetwork()
	},
}

var netRemoveCmd = &cobra.Command{
	Use: "remove",
	Short: "remove container network",
	Long: "remove container network",
	Aliases: []string{"rm"},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			log.ConsoleLog.Fatal("Missing network name")
			return
		}
		_ = network.Init()
		err := network.DeleteNetwork(args[0])
		if err != nil {
			log.ConsoleLog.Fatal("remove network error: %+v", err)
		}
	},
}

func init() {
	netCreateCmd.Flags().StringVarP(&driver, "driver", "d", "bridge", "network driver")
	netCreateCmd.Flags().StringVarP(&subnet, "subnet", "s", "192.168.0.1/24", "subnet driver")
	networkCmd.AddCommand(netCreateCmd)
	networkCmd.AddCommand(netListCmd)
	networkCmd.AddCommand(netRemoveCmd)
}
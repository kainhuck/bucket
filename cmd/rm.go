package cmd

import (
	"bucket/container"
	"bucket/log"
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var removeCmd = &cobra.Command{
	Use:   "rm",
	Short: "remove containers",
	Long:  "remove containers",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			log.ConsoleLog.Fatal("Missing container name")
			return
		}
		containerName := args[0]
		removeContainer(containerName)
	},
}

func removeContainer(containerName string) {
	containerInfo, err := getContainerInfoByName(containerName)
	if err != nil {
		log.ConsoleLog.Error("Get container %s info error %v", containerName, err)
		return
	}
	if containerInfo.Status != container.STOP {
		log.ConsoleLog.Error("Couldn't remove running container")
		return
	}
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	if err := os.RemoveAll(dirURL); err != nil {
		log.ConsoleLog.Error("Remove file %s error %v", dirURL, err)
		return
	}
	container.DeleteWorkSpace(containerInfo.Volume, containerName)
}

package cmd

import (
	"bucket/container"
	"bucket/log"
	"fmt"
	"github.com/spf13/cobra"
	"os/exec"
)

var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "commit a container into image",
	Long:  "commit a container into image",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			log.ConsoleLog.Fatal("Missing container name or image name")
			return
		}
		containerName := args[0]
		imageName := args[1]
		commitContainer(containerName, imageName)
	},
}

func commitContainer(containerName, imageName string) {
	mntURL := fmt.Sprintf(container.MntUrl, containerName)
	mntURL += "/"

	imageTar := container.RootUrl + "/" + imageName + ".tar"

	if _, err := exec.Command("tar", "-czf", imageTar, "-C", mntURL, ".").CombinedOutput(); err != nil {
		log.ConsoleLog.Error("Tar folder %s error %v", mntURL, err)
	}
}

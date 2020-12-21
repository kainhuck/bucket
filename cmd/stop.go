package cmd

import (
	"bucket/container"
	"bucket/log"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"strconv"
	"syscall"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "stop a container",
	Long:  "stop a container",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			log.ConsoleLog.Fatal("Missing container name")
			return
		}
		containerName := args[0]
		stopContainer(containerName)
	},
}

func stopContainer(containerName string) {
	pid, err := GetContainerPidByName(containerName)
	if err != nil {
		log.ConsoleLog.Error("Get contaienr pid by name %s error %v", containerName, err)
		return
	}
	pidInt, err := strconv.Atoi(pid)
	if err != nil {
		log.ConsoleLog.Error("Conver pid from string to int error %v", err)
		return
	}
	if err := syscall.Kill(pidInt, syscall.SIGTERM); err != nil {
		log.ConsoleLog.Error("Stop container %s error %v", containerName, err)
		return
	}
	containerInfo, err := getContainerInfoByName(containerName)
	if err != nil {
		log.ConsoleLog.Error("Get container %s info error %v", containerName, err)
		return
	}
	containerInfo.Status = container.STOP
	containerInfo.Pid = " "
	newContentBytes, err := json.Marshal(containerInfo)
	if err != nil {
		log.ConsoleLog.Error("Json marshal %s error %v", containerName, err)
		return
	}
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	configFilePath := dirURL + container.ConfigName
	if err := ioutil.WriteFile(configFilePath, newContentBytes, 0622); err != nil {
		log.ConsoleLog.Error("Write file %s error", configFilePath, err)
	}
}

func getContainerInfoByName(containerName string) (*container.ContainerInfo, error) {
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	configFilePath := dirURL + container.ConfigName
	contentBytes, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		log.ConsoleLog.Error("Read file %s error %v", configFilePath, err)
		return nil, err
	}
	var containerInfo container.ContainerInfo
	if err := json.Unmarshal(contentBytes, &containerInfo); err != nil {
		log.ConsoleLog.Error("GetContainerInfoByName unmarshal error %v", err)
		return nil, err
	}
	return &containerInfo, nil
}

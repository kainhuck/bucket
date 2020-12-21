package cmd

import (
	"bucket/container"
	"bucket/log"
	_ "bucket/nsenter"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

const ENV_EXEC_PID = "bucket_pid"
const ENV_EXEC_CMD = "bucket_cmd"

var execCmd = &cobra.Command{
	Use:   "exec",
	Short: "exec a command into container",
	Long:  "exec a command into container",
	Run: func(cmd *cobra.Command, args []string) {
		if os.Getenv(ENV_EXEC_PID) != "" {
			log.ConsoleLog.Info("get callback pid: %v", os.Getpid())
		}

		if len(args) < 2 {
			log.ConsoleLog.Fatal("Missing container name or command")
			return
		}

		containerName := args[0]
		commandList := args[1:]
		ExecContainer(containerName, commandList)
	},
}

func ExecContainer(containerName string, comArray []string) {
	pid, err := GetContainerPidByName(containerName)
	if err != nil {
		log.ConsoleLog.Error("Exec container getContainerPidByName %s error %v", containerName, err)
		return
	}

	cmdStr := strings.Join(comArray, " ")
	log.ConsoleLog.Info("container pid %s", pid)
	log.ConsoleLog.Info("command %s", cmdStr)

	cmd := exec.Command("/proc/self/exe", "exec")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	os.Setenv(ENV_EXEC_PID, pid)
	os.Setenv(ENV_EXEC_CMD, cmdStr)
	containerEnvs := getEnvsByPid(pid)
	cmd.Env = append(os.Environ(), containerEnvs...)

	if err := cmd.Run(); err != nil {
		log.ConsoleLog.Error("Exec container %s error %v", containerName, err)
	}
}

func GetContainerPidByName(containerName string) (string, error) {
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	configFilePath := dirURL + container.ConfigName
	contentBytes, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return "", err
	}
	var containerInfo container.ContainerInfo
	if err := json.Unmarshal(contentBytes, &containerInfo); err != nil {
		return "", err
	}
	return containerInfo.Pid, nil
}

func getEnvsByPid(pid string) []string {
	path := fmt.Sprintf("/proc/%s/environ", pid)
	contentBytes, err := ioutil.ReadFile(path)
	if err != nil {
		log.ConsoleLog.Error("Read file %s error %v", path, err)
		return nil
	}
	//env split by \u0000
	envs := strings.Split(string(contentBytes), "\u0000")
	return envs
}

package cmd

import (
	"bucket/cgroups"
	"bucket/cgroups/subsystems"
	"bucket/container"
	"bucket/log"
	"bucket/network"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

var tty bool
var input bool
var detach bool
var name string
var volume string
var memory string
var cpuSet string
var cpuShare string
var envList []string
var net string
var portMapping []string

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "运行容器",
	Long:  "运行容器命令",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			log.ConsoleLog.Fatal("missing image name")
			return
		}

		imageName := args[0]
		cmdList := args[1:]

		if detach && tty {
			log.ConsoleLog.Fatal("t and d parameter can not both provided")
			return
		}

		resConf := &subsystems.ResourceConfig{
			MemoryLimit: memory,
			CpuShare:    cpuShare,
			CpuSet:      cpuSet,
		}

		Run(input, tty, cmdList, resConf, name, volume, imageName, envList, net, portMapping)
	},
}

func init() {
	runCmd.Flags().BoolVarP(&tty, "tty", "t", true, "enable tty")
	runCmd.Flags().BoolVarP(&input, "input", "i", true, "pen std input")
	runCmd.Flags().BoolVarP(&detach, "detach", "d", false, "detach container")
	runCmd.Flags().StringVarP(&name, "name", "n", "", "set container Name")
	runCmd.Flags().StringVarP(&volume, "volume", "v", "", "set container volume")
	runCmd.Flags().StringVarP(&memory, "memory", "m", "", "set container memory limit")
	runCmd.Flags().StringVarP(&cpuSet, "cpuset", "x", "", "set container cpuset")
	runCmd.Flags().StringVarP(&cpuShare, "cpushare", "y", "", "set container cpushare")
	runCmd.Flags().StringVarP(&net, "net", "z", "", "set container network")
	runCmd.Flags().StringSliceVarP(&portMapping, "port", "p", []string{}, "set container port")
	runCmd.Flags().StringSliceVarP(&envList, "environment", "e", []string{}, "set container env")
}

func Run(input, tty bool, comArray []string, res *subsystems.ResourceConfig, containerName, volume, imageName string, envSlice []string,
	nw string, portMapping []string) {
	containerID := randStringBytes(10)
	if containerName == "" {
		containerName = containerID
	}

	parent, writePipe := container.NewContainerProcess(input, tty, containerName, volume, imageName, envSlice)
	if parent == nil {
		log.ConsoleLog.Error("New parent process error")
		return
	}
	if err := parent.Start(); err != nil {
		log.ConsoleLog.Error("%v", err)
	}

	//record container info
	containerName, err := recordContainerInfo(parent.Process.Pid, comArray, containerName, containerID, volume)
	if err != nil {
		log.ConsoleLog.Error("Record container info error %v", err)
		return
	}

	// use bucket-cgroup as cgroup name
	cgroupManager := cgroups.NewCgroupManager("bucket-cgroup")
	defer cgroupManager.Destroy()
	cgroupManager.Set(res)
	cgroupManager.Apply(parent.Process.Pid)

	if nw != "" {
		// config container network
		_ = network.Init()
		containerInfo := &container.ContainerInfo{
			Id:          containerID,
			Pid:         strconv.Itoa(parent.Process.Pid),
			Name:        containerName,
			PortMapping: portMapping,
		}
		if err := network.Connect(nw, containerInfo); err != nil {
			log.ConsoleLog.Error("Error Connect Network %v", err)
			return
		}
	}

	sendInitCommand(comArray, writePipe)

	if tty {
		parent.Wait()
		deleteContainerInfo(containerName)
		container.DeleteWorkSpace(volume, containerName)
	}

}

func sendInitCommand(comArray []string, writePipe *os.File) {
	command := strings.Join(comArray, " ")
	log.ConsoleLog.Info("command all is %s", command)
	writePipe.WriteString(command)
	writePipe.Close()
}

func recordContainerInfo(containerPID int, commandArray []string, containerName, id, volume string) (string, error) {
	createTime := time.Now().Format("2006-01-02 15:04:05")
	command := strings.Join(commandArray, "")
	containerInfo := &container.ContainerInfo{
		Id:          id,
		Pid:         strconv.Itoa(containerPID),
		Command:     command,
		CreatedTime: createTime,
		Status:      container.RUNNING,
		Name:        containerName,
		Volume:      volume,
	}

	jsonBytes, err := json.Marshal(containerInfo)
	if err != nil {
		log.ConsoleLog.Error("Record container info error %v", err)
		return "", err
	}
	jsonStr := string(jsonBytes)

	dirUrl := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	if err := os.MkdirAll(dirUrl, 0622); err != nil {
		log.ConsoleLog.Error("Mkdir error %s error %v", dirUrl, err)
		return "", err
	}
	fileName := dirUrl + "/" + container.ConfigName
	file, err := os.Create(fileName)
	defer file.Close()
	if err != nil {
		log.ConsoleLog.Error("Create file %s error %v", fileName, err)
		return "", err
	}
	if _, err := file.WriteString(jsonStr); err != nil {
		log.ConsoleLog.Error("File write string error %v", err)
		return "", err
	}

	return containerName, nil
}

func deleteContainerInfo(containerId string) {
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, containerId)
	if err := os.RemoveAll(dirURL); err != nil {
		log.ConsoleLog.Error("Remove dir %s error %v", dirURL, err)
	}
}

func randStringBytes(n int) string {
	letterBytes := "1234567890"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

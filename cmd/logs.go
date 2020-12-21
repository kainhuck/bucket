package cmd

import (
	"bucket/container"
	"bucket/log"
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
)

var logCmd = &cobra.Command{
	Use:   "logs",
	Short: "print logs of container",
	Long:  "print logs of container",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			log.ConsoleLog.Fatal("please input container name")
			return
		}
		logContainer(args[0])
	},
}

func logContainer(containerName string) {
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	logFileLocation := dirURL + container.ContainerLogFile
	file, err := os.Open(logFileLocation)
	defer file.Close()
	if err != nil {
		log.ConsoleLog.Error("Log container open file %s error %v", logFileLocation, err)
		return
	}
	content, err := ioutil.ReadAll(file)
	if err != nil {
		log.ConsoleLog.Error("Log container read file %s error %v", logFileLocation, err)
		return
	}
	fmt.Fprint(os.Stdout, string(content))
}

package container

import (
	"bucket/log"
	"os/exec"
	"testing"
)

func TestInit(t *testing.T) {
	path, err := exec.LookPath("ls")
	if err != nil {
		log.ConsoleLog.Fatal("error: %v", err)
		return
	}
	log.ConsoleLog.Info(path)
}
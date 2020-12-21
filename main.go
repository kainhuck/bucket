package main

import (
	"bucket/cmd"
	"bucket/log"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.ConsoleLog.Fatal("bucket run failed: %v", err)
	}
}

package main

import (
	arukas "github.com/arukasio/cli"
	"log"
)

func stopContainer(stopContainerID string) {
	client := arukas.NewClientWithOsExitOnErr()

	if err := client.Delete("/containers/" + stopContainerID + "/power"); err != nil {
		client.Println(nil, "Failed to stop the container")
		log.Println(err)
		ExitCode = 1
	} else {
		client.Println(nil, "Stopping...")
	}
}

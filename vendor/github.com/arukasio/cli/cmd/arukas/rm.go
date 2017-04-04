package main

import (
	arukas "github.com/arukasio/cli"
	"log"
)

func removeContainer(containerID string) {
	client := arukas.NewClientWithOsExitOnErr()
	var container arukas.Container

	if err := client.Get(&container, "/containers/"+containerID); err != nil {
		client.Println(nil, "Failed to rm the container")
		log.Println(err)
		ExitCode = 1
		return
	}

	if err := client.Delete("/apps/" + container.App.ID); err != nil {
		client.Println(nil, "Failed to rm the container")
		log.Println(err)
		ExitCode = 1
		return
	}
}

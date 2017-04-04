package main

import (
	arukas "github.com/arukasio/cli"
)

func createAndRunContainer(name string, image string, instances int, mem int, envs []string, ports []string, cmd string, appName string) {
	client := arukas.NewClientWithOsExitOnErr()
	var appSet arukas.AppSet

	// create an app
	newApp := arukas.App{Name: appName}

	var parsedEnvs arukas.Envs
	var parsedPorts arukas.Ports

	if len(envs) > 0 {
		var err error
		parsedEnvs, err = arukas.ParseEnv(envs)
		if err != nil {
			client.Println(nil, err)
			ExitCode = 1
			return
		}
	}

	if len(ports) > 0 {
		var err error
		parsedPorts, err = arukas.ParsePort(ports)
		if err != nil {
			client.Println(nil, err)
			ExitCode = 1
			return
		}
	}

	newContainer := arukas.Container{
		Envs:      parsedEnvs,
		Ports:     parsedPorts,
		ImageName: image,
		Mem:       mem,
		Instances: instances,
		Cmd:       cmd,
		Name:      name,
	}

	newAppSet := arukas.AppSet{
		App:       newApp,
		Container: newContainer,
	}

	if err := client.Post(&appSet, "/app-sets", newAppSet); err != nil {
		client.Println(nil, err)
		ExitCode = 1
		return
	}

	startContainer(appSet.Container.ID, true)

	client.Println(nil, "ID", "IMAGE", "CREATED", "STATUS", "NAME", "ENDPOINT")
	client.Println(nil, appSet.Container.ID, appSet.Container.ImageName, appSet.Container.CreatedAt.String(),
		appSet.Container.StatusText, appSet.Container.Name, appSet.Container.Endpoint)
}

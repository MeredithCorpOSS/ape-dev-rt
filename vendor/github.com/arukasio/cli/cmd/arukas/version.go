package main

import (
	arukas "github.com/arukasio/cli"
)

func displayVersion() {
	client := arukas.NewClientWithOsExitOnErr()
	client.Println(nil, arukas.VERSION)
}

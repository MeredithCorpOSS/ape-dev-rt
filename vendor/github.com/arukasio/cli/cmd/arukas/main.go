package main

import (
	"github.com/joho/godotenv"
	"os"
)

func main() {
	godotenv.Load()
	exitCode := Run(os.Args)
	os.Exit(exitCode)
}

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/ttacon/chalk"
	"github.com/urfave/cli"
)

var errorStyle = chalk.Red.NewStyle().WithTextStyle(chalk.Bold).Style

func main() {
	app := cli.NewApp()
	app.Name = "release tool"
	app.Usage = "For amazing releases"
	app.EnableBashCompletion = true
	app.HideVersion = true
	app.Metadata = make(map[string]interface{})
	app.Flags = []cli.Flag{
		flags.Config,
		flags.Profile,
		flags.Verbose,
		flags.EnableFileLogging,
		flags.AwsProfile,
		flags.Module,
	}

	app.Before = func(c *cli.Context) error {
		if !c.Bool("verbose") {
			log.SetOutput(ioutil.Discard)
		}

		if c.BoolT("enable-file-logging") {
			logFile, err := createLogFile()
			if err != nil {
				return err
			}
			log.SetFlags(log.LstdFlags | log.Llongfile)
			log.SetOutput(logFile)
		}

		return nil
	}

	app.Commands = Commands

	err := app.Run(os.Args)
	if err != nil {
		// Since logging setup can fail too, we just use stderr
		fmt.Fprintln(os.Stderr, errorStyle("[ERROR] "+err.Error()))
		os.Exit(1)
	}
}

func createLogFile() (*os.File, error) {
	hd, err := homedir.Dir()
	if err != nil {
		return nil, fmt.Errorf("Unable to resolve current user homedir: %s", err)
	}

	logDir := filepath.Join(hd, ".rt", "logs")
	err = os.MkdirAll(logDir, 0755)
	if err != nil {
		return nil, err
	}

	layout := "2006-01-02_T15-04-05_Z07-00"
	fileName := "rt-" + time.Now().Format(layout) + ".log"
	logPath := filepath.Join(logDir, fileName)
	logFile, err := os.Create(logPath)
	if err != nil {
		return nil, err
	}

	return logFile, nil
}

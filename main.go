package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/terraform/helper/logging"
	"github.com/mitchellh/go-homedir"
	"github.com/mitchellh/panicwrap"
	"github.com/ttacon/chalk"
	"github.com/urfave/cli"
	"syscall"
)

var errorStyle = chalk.Red.NewStyle().WithTextStyle(chalk.Bold).Style

func main() {
	os.Exit(realMain())
}

func realMain() int {
	var wrapConfig panicwrap.WrapConfig
	if !panicwrap.Wrapped(&wrapConfig) {
		// Determine where logs should go in general (requested by the user)
		logWriter, err := logging.LogOutput()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Couldn't setup log output: %s", err)
			return 1
		}

		// We always send logs to a temporary file that we use in case
		// there is a panic. Otherwise, we delete it.
		logTempFile, err := ioutil.TempFile("", "terraform-log")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Couldn't setup logging tempfile: %s", err)
			return 1
		}
		defer os.Remove(logTempFile.Name())
		defer logTempFile.Close()

		// Setup the prefixed readers that send data properly to
		// stdout/stderr.
		doneCh := make(chan struct{})
		outR, outW := io.Pipe()
		go copyOutput(outR, doneCh)

		// Create the configuration for panicwrap and wrap our executable
		wrapConfig.Handler = panicHandler(logTempFile)
		wrapConfig.Writer = io.MultiWriter(logTempFile, logWriter)
		wrapConfig.Stdout = outW
		wrapConfig.IgnoreSignals = []os.Signal{os.Interrupt}
		wrapConfig.ForwardSignals = []os.Signal{syscall.SIGTERM}
		exitStatus, err := panicwrap.Wrap(&wrapConfig)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Couldn't start Terraform: %s", err)
			return 1
		}
		// If >= 0, we're the parent, so just exit
		if exitStatus >= 0 {
			// Close the stdout writer so that our copy process can finish
			outW.Close()

			// Wait for the output copying to finish
			<-doneCh

			return exitStatus
		}

		// We're the child, so just close the tempfile we made in order to
		// save file handles since the tempfile is only used by the parent.
		logTempFile.Close()
	}
	return wrappedMain()
}

func wrappedMain() int {
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
		return 1
	}

	return 0
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

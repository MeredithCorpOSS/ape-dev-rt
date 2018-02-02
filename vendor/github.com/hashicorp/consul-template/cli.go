package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/consul-template/config"
	"github.com/hashicorp/consul-template/logging"
	"github.com/hashicorp/consul-template/manager"
	"github.com/hashicorp/consul-template/signals"
	"github.com/hashicorp/consul-template/watch"
)

// Exit codes are int values that represent an exit code for a particular error.
// Sub-systems may check this unique error to determine the cause of an error
// without parsing the output or help text.
//
// Errors start at 10
const (
	ExitCodeOK int = 0

	ExitCodeError = 10 + iota
	ExitCodeInterrupt
	ExitCodeParseFlagsError
	ExitCodeRunnerError
	ExitCodeConfigError
)

/// ------------------------- ///

// CLI is the main entry point for Consul Template.
type CLI struct {
	sync.Mutex

	// outSteam and errStream are the standard out and standard error streams to
	// write messages from the CLI.
	outStream, errStream io.Writer

	// stopCh is an internal channel used to trigger a shutdown of the CLI.
	stopCh  chan struct{}
	stopped bool
}

// NewCLI creates a new CLI object with the given stdout and stderr streams.
func NewCLI(out, err io.Writer) *CLI {
	return &CLI{
		outStream: out,
		errStream: err,
		stopCh:    make(chan struct{}),
	}
}

// Run accepts a slice of arguments and returns an int representing the exit
// status from the command.
func (cli *CLI) Run(args []string) int {
	// Parse the flags
	config, once, dry, version, err := cli.parseFlags(args[1:])
	if err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return cli.handleError(err, ExitCodeParseFlagsError)
	}

	// Save original config (defaults + parsed flags) for handling reloads
	baseConfig := config.Copy()

	// Setup the config and logging
	config, err = cli.setup(config)
	if err != nil {
		return cli.handleError(err, ExitCodeConfigError)
	}

	// Print version information for debugging
	log.Printf("[INFO] %s", humanVersion)

	// If the version was requested, return an "error" containing the version
	// information. This might sound weird, but most *nix applications actually
	// print their version on stderr anyway.
	if version {
		log.Printf("[DEBUG] (cli) version flag was given, exiting now")
		fmt.Fprintf(cli.errStream, "%s\n", humanVersion)
		return ExitCodeOK
	}

	// Initial runner
	runner, err := manager.NewRunner(config, dry, once)
	if err != nil {
		return cli.handleError(err, ExitCodeRunnerError)
	}
	go runner.Start()

	// Listen for signals
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh)

	for {
		select {
		case err := <-runner.ErrCh:
			// Check if the runner's error returned a specific exit status, and return
			// that value. If no value was given, return a generic exit status.
			code := ExitCodeRunnerError
			if typed, ok := err.(manager.ErrExitable); ok {
				code = typed.ExitStatus()
			}
			return cli.handleError(err, code)
		case <-runner.DoneCh:
			return ExitCodeOK
		case s := <-signalCh:
			log.Printf("[DEBUG] (cli) receiving signal %q", s)

			switch s {
			case config.ReloadSignal:
				fmt.Fprintf(cli.errStream, "Reloading configuration...\n")
				runner.Stop()

				// Load the new configuration from disk
				config, err = cli.setup(baseConfig)
				if err != nil {
					return cli.handleError(err, ExitCodeConfigError)
				}

				runner, err = manager.NewRunner(config, dry, once)
				if err != nil {
					return cli.handleError(err, ExitCodeRunnerError)
				}
				go runner.Start()
			case config.DumpSignal:
				runner.Stop()
				debug.PrintStack()
				return ExitCodeInterrupt
			case config.KillSignal:
				fmt.Fprintf(cli.errStream, "Cleaning up...\n")
				runner.Stop()
				return ExitCodeInterrupt
			case signals.SignalLookup["SIGCHLD"]:
				// The SIGCHLD signal is sent to the parent of a child process when it
				// exits, is interrupted, or resumes after being interrupted. We ignore
				// this signal because the child process is monitored on its own.
				//
				// Also, the reason we do a lookup instead of a direct syscall.SIGCHLD
				// is because that isn't defined on Windows.
			default:
				// Propogate the signal to the child process
				runner.Signal(s)
			}
		case <-cli.stopCh:
			return ExitCodeOK
		}
	}
}

// stop is used internally to shutdown a running CLI
func (cli *CLI) stop() {
	cli.Lock()
	defer cli.Unlock()

	if cli.stopped {
		return
	}

	close(cli.stopCh)
	cli.stopped = true
}

// parseFlags is a helper function for parsing command line flags using Go's
// Flag library. This is extracted into a helper to keep the main function
// small, but it also makes writing tests for parsing command line arguments
// much easier and cleaner.
func (cli *CLI) parseFlags(args []string) (*config.Config, bool, bool, bool, error) {
	var dry, once, version bool
	conf := config.DefaultConfig()

	// Parse the flags and options
	flags := flag.NewFlagSet(Name, flag.ContinueOnError)
	flags.SetOutput(cli.errStream)
	flags.Usage = func() { fmt.Fprintf(cli.errStream, usage, Name) }

	flags.Var((funcVar)(func(s string) error {
		conf.Consul = s
		conf.Set("consul")
		return nil
	}), "consul", "")

	flags.Var((funcVar)(func(s string) error {
		conf.Token = s
		conf.Set("token")
		return nil
	}), "token", "")

	flags.Var((funcVar)(func(s string) error {
		if s == "" {
			conf.ReloadSignal = nil
			conf.Set("reload_signal")
			return nil
		}

		sig, err := signals.Parse(s)
		if err != nil {
			return err
		}
		conf.ReloadSignal = sig
		conf.Set("reload_signal")
		return nil
	}), "reload-signal", "")

	flags.Var((funcVar)(func(s string) error {
		if s == "" {
			conf.DumpSignal = nil
			conf.Set("dump_signal")
			return nil
		}

		sig, err := signals.Parse(s)
		if err != nil {
			return err
		}
		conf.DumpSignal = sig
		conf.Set("dump_signal")
		return nil
	}), "dump-signal", "")

	flags.Var((funcVar)(func(s string) error {
		if s == "" {
			conf.KillSignal = nil
			conf.Set("kill_signal")
			return nil
		}

		sig, err := signals.Parse(s)
		if err != nil {
			return err
		}
		conf.KillSignal = sig
		conf.Set("kill_signal")
		return nil
	}), "kill-signal", "")

	flags.Var((funcVar)(func(s string) error {
		conf.Auth.Enabled = true
		conf.Set("auth.enabled")
		if strings.Contains(s, ":") {
			split := strings.SplitN(s, ":", 2)
			conf.Auth.Username = split[0]
			conf.Set("auth.username")
			conf.Auth.Password = split[1]
			conf.Set("auth.password")
		} else {
			conf.Auth.Username = s
			conf.Set("auth.username")
		}
		return nil
	}), "auth", "")

	flags.Var((funcBoolVar)(func(b bool) error {
		conf.SSL.Enabled = b
		conf.Set("ssl")
		conf.Set("ssl.enabled")
		return nil
	}), "ssl", "")

	flags.Var((funcBoolVar)(func(b bool) error {
		conf.SSL.Verify = b
		conf.Set("ssl")
		conf.Set("ssl.verify")
		return nil
	}), "ssl-verify", "")

	flags.Var((funcVar)(func(s string) error {
		conf.SSL.Cert = s
		conf.Set("ssl")
		conf.Set("ssl.cert")
		return nil
	}), "ssl-cert", "")

	flags.Var((funcVar)(func(s string) error {
		conf.SSL.Key = s
		conf.Set("ssl")
		conf.Set("ssl.key")
		return nil
	}), "ssl-key", "")

	flags.Var((funcVar)(func(s string) error {
		conf.SSL.CaCert = s
		conf.Set("ssl")
		conf.Set("ssl.ca_cert")
		return nil
	}), "ssl-ca-cert", "")

	flags.Var((funcDurationVar)(func(d time.Duration) error {
		conf.MaxStale = d
		conf.Set("max_stale")
		return nil
	}), "max-stale", "")

	flags.Var((funcVar)(func(s string) error {
		t, err := config.ParseConfigTemplate(s)
		if err != nil {
			return err
		}
		if conf.ConfigTemplates == nil {
			conf.ConfigTemplates = make([]*config.ConfigTemplate, 0, 1)
		}
		conf.ConfigTemplates = append(conf.ConfigTemplates, t)
		return nil
	}), "template", "")

	flags.Var((funcBoolVar)(func(b bool) error {
		conf.Syslog.Enabled = b
		conf.Set("syslog")
		conf.Set("syslog.enabled")
		return nil
	}), "syslog", "")

	flags.Var((funcVar)(func(s string) error {
		conf.Syslog.Facility = s
		conf.Set("syslog.facility")
		return nil
	}), "syslog-facility", "")

	flags.Var((funcBoolVar)(func(b bool) error {
		conf.Deduplicate.Enabled = b
		conf.Set("deduplicate")
		conf.Set("deduplicate.enabled")
		return nil
	}), "dedup", "")

	flags.Var((funcVar)(func(s string) error {
		conf.Exec.Command = s
		conf.Set("exec")
		conf.Set("exec.command")
		return nil
	}), "exec", "")

	flags.Var((funcDurationVar)(func(d time.Duration) error {
		conf.Exec.Splay = d
		conf.Set("exec.splay")
		return nil
	}), "exec-splay", "")

	flags.Var((funcVar)(func(s string) error {
		sig, err := signals.Parse(s)
		if err != nil {
			return err
		}
		conf.Exec.ReloadSignal = sig
		conf.Set("exec.reload_signal")
		return nil
	}), "exec-reload-signal", "")

	flags.Var((funcVar)(func(s string) error {
		sig, err := signals.Parse(s)
		if err != nil {
			return err
		}
		conf.Exec.KillSignal = sig
		conf.Set("exec.kill_signal")
		return nil
	}), "exec-kill-signal", "")

	flags.Var((funcDurationVar)(func(d time.Duration) error {
		conf.Exec.KillTimeout = d
		conf.Set("exec.kill_timeout")
		return nil
	}), "exec-kill-timeout", "")

	flags.Var((funcVar)(func(s string) error {
		w, err := watch.ParseWait(s)
		if err != nil {
			return err
		}
		conf.Wait.Min = w.Min
		conf.Wait.Max = w.Max
		conf.Set("wait")
		return nil
	}), "wait", "")

	flags.Var((funcDurationVar)(func(d time.Duration) error {
		conf.Retry = d
		conf.Set("retry")
		return nil
	}), "retry", "")

	flags.Var((funcVar)(func(s string) error {
		conf.Path = s
		conf.Set("path")
		return nil
	}), "config", "")

	flags.Var((funcVar)(func(s string) error {
		conf.PidFile = s
		conf.Set("pid_file")
		return nil
	}), "pid-file", "")

	flags.Var((funcVar)(func(s string) error {
		conf.LogLevel = s
		conf.Set("log_level")
		return nil
	}), "log-level", "")

	flags.Var((funcVar)(func(s string) error {
		conf.Vault.Address = s
		conf.Set("vault")
		conf.Set("vault.address")
		return nil
	}), "vault-addr", "")

	flags.Var((funcVar)(func(s string) error {
		conf.Vault.Token = s
		conf.Set("vault")
		conf.Set("vault.token")
		return nil
	}), "vault-token", "")

	flags.Var((funcBoolVar)(func(b bool) error {
		conf.Vault.RenewToken = b
		conf.Set("vault")
		conf.Set("vault.renew_token")
		return nil
	}), "vault-renew-token", "")

	flags.Var((funcBoolVar)(func(b bool) error {
		conf.Vault.UnwrapToken = b
		conf.Set("vault")
		conf.Set("vault.unwrap_token")
		return nil
	}), "vault-unwrap-token", "")

	flags.BoolVar(&once, "once", false, "")
	flags.BoolVar(&dry, "dry", false, "")
	flags.BoolVar(&version, "v", false, "")
	flags.BoolVar(&version, "version", false, "")

	// If there was a parser error, stop
	if err := flags.Parse(args); err != nil {
		return nil, false, false, false, err
	}

	// Error if extra arguments are present
	args = flags.Args()
	if len(args) > 0 {
		return nil, false, false, false, fmt.Errorf("cli: extra argument(s): %q",
			args)
	}

	return conf, once, dry, version, nil
}

// handleError outputs the given error's Error() to the errStream and returns
// the given exit status.
func (cli *CLI) handleError(err error, status int) int {
	fmt.Fprintf(cli.errStream, "Consul Template returned errors:\n%s\n", err)
	return status
}

func (cli *CLI) setup(conf *config.Config) (*config.Config, error) {
	if conf.Path != "" {
		newConfig, err := config.FromPath(conf.Path)
		if err != nil {
			return nil, err
		}

		// Merge ensuring that the CLI options still take precedence
		newConfig.Merge(conf)
		conf = newConfig
	}

	// Setup the logging
	if err := logging.Setup(&logging.Config{
		Name:           Name,
		Level:          conf.LogLevel,
		Syslog:         conf.Syslog.Enabled,
		SyslogFacility: conf.Syslog.Facility,
		Writer:         cli.errStream,
	}); err != nil {
		return nil, err
	}

	return conf, nil
}

const usage = `
Usage: %s [options]

  Watches a series of templates on the file system, writing new changes when
  Consul is updated. It runs until an interrupt is received unless the -once
  flag is specified.

Options:

  -auth=<username[:password]>
      Set the basic authentication username (and password)

  -config=<path>
      Sets the path to a configuration file on disk

  -consul=<address>
      Sets the address of the Consul instance

  -dedup
      Enable de-duplication mode - reduces load on Consul when many instances of
      Consul Template are rendering a common template

  -dry
      Print generated templates to stdout instead of rendering

  -dump-signal=<signal>
      Signal to listen to initiate a core dump and terminate the process

  -exec=<command>
      Enable exec mode to run as a supervisor-like process - the given command
      will receive all signals provided to the parent process and will receive a
      signal when templates change

  -exec-kill-signal=<signal>
      Signal to send when gracefully killing the process

  -exec-kill-timeout=<duration>
      Amount of time to wait before force-killing the child

  -exec-reload-signal=<signal>
      Signal to send when a reload takes place

  -exec-splay=<duration>
      Amount of time to wait before sending signals

  -kill-signal=<signal>
      Signal to listen to gracefully terminate the process

  -log-level=<level>
      Set the logging level - values are "debug", "info", "warn", and "err"

  -max-stale=<duration>
      Set the maximum staleness and allow stale queries to Consul which will
      distribute work among all servers instead of just the leader

  -once
      Do not run the process as a daemon

  -pid-file=<path>
      Path on disk to write the PID of the process

  -reload-signal=<signal>
      Signal to listen to reload configuration

  -retry=<duration>
      The amount of time to wait if Consul returns an error when communicating
      with the API

  -ssl
      Use SSL when connecting to Consul

  -ssl-ca-cert
      Validate server certificate against this CA certificate file list

  -ssl-cert
      SSL client certificate to send to server

  -ssl-key
      SSL/TLS private key for use in client authentication key exchange

  -ssl-verify
      Verify certificates when connecting via SSL

  -syslog
      Send the output to syslog instead of standard error and standard out. The
      syslog facility defaults to LOCAL0 and can be changed using a
      configuration file

  -syslog-facility=<facility>
      Set the facility where syslog should log - if this attribute is supplied,
      the -syslog flag must also be supplied

  -template=<template>
       Adds a new template to watch on disk in the format 'in:out(:command)'

  -token=<token>
      Sets the Consul API token

  -vault-addr=<address>
      Sets the address of the Vault server

  -vault-token=<token>
      Sets the Vault API token

  -vault-renew-token
      Periodically renew the provided Vault API token - this defaults to "true"
      and will renew the token at half of the lease duration

  -vault-unwrap-token
      Unwrap the provided Vault API token (see Vault documentation for more
      information on this feature)

  -wait=<duration>
      Sets the 'min(:max)' amount of time to wait before writing a template (and
      triggering a command)

  -v, -version
      Print the version of this daemon
`

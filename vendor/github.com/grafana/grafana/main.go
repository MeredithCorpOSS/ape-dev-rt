package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"github.com/grafana/grafana/pkg/cmd"
	"github.com/grafana/grafana/pkg/log"
	"github.com/grafana/grafana/pkg/login"
	"github.com/grafana/grafana/pkg/metrics"
	"github.com/grafana/grafana/pkg/plugins"
	"github.com/grafana/grafana/pkg/services/eventpublisher"
	"github.com/grafana/grafana/pkg/services/notifications"
	"github.com/grafana/grafana/pkg/services/search"
	"github.com/grafana/grafana/pkg/services/sqlstore"
	"github.com/grafana/grafana/pkg/setting"
	"github.com/grafana/grafana/pkg/social"
)

var version = "master"
var commit = "NA"
var buildstamp string
var build_date string

var configFile = flag.String("config", "", "path to config file")
var homePath = flag.String("homepath", "", "path to grafana install/home path, defaults to working directory")
var pidFile = flag.String("pidfile", "", "path to pid file")
var exitChan = make(chan int)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {

	v := flag.Bool("v", false, "prints current version and exits")
	flag.Parse()
	if *v {
		fmt.Printf("Version %s (commit: %s)\n", version, commit)
		os.Exit(0)
	}

	buildstampInt64, _ := strconv.ParseInt(buildstamp, 10, 64)

	setting.BuildVersion = version
	setting.BuildCommit = commit
	setting.BuildStamp = buildstampInt64

	go listenToSystemSignels()

	flag.Parse()
	writePIDFile()
	initRuntime()

	search.Init()
	login.Init()
	social.NewOAuthService()
	eventpublisher.Init()
	plugins.Init()

	if err := notifications.Init(); err != nil {
		log.Fatal(3, "Notification service failed to initialize", err)
	}

	if setting.ReportingEnabled {
		go metrics.StartUsageReportLoop()
	}

	cmd.StartServer()
	exitChan <- 0
}

func initRuntime() {
	err := setting.NewConfigContext(&setting.CommandLineArgs{
		Config:   *configFile,
		HomePath: *homePath,
		Args:     flag.Args(),
	})

	if err != nil {
		log.Fatal(3, err.Error())
	}

	log.Info("Starting Grafana")
	log.Info("Version: %v, Commit: %v, Build date: %v", setting.BuildVersion, setting.BuildCommit, time.Unix(setting.BuildStamp, 0))
	setting.LogConfigurationInfo()

	sqlstore.NewEngine()
	sqlstore.EnsureAdminUser()
}

func writePIDFile() {
	if *pidFile == "" {
		return
	}

	// Ensure the required directory structure exists.
	err := os.MkdirAll(filepath.Dir(*pidFile), 0700)
	if err != nil {
		log.Fatal(3, "Failed to verify pid directory", err)
	}

	// Retrieve the PID and write it.
	pid := strconv.Itoa(os.Getpid())
	if err := ioutil.WriteFile(*pidFile, []byte(pid), 0644); err != nil {
		log.Fatal(3, "Failed to write pidfile", err)
	}
}

func listenToSystemSignels() {
	signalChan := make(chan os.Signal, 1)
	code := 0

	signal.Notify(signalChan, os.Interrupt)
	signal.Notify(signalChan, os.Kill)
	signal.Notify(signalChan, syscall.SIGTERM)

	select {
	case sig := <-signalChan:
		log.Info("Received signal %s. shutting down", sig)
	case code = <-exitChan:
		switch code {
		case 0:
			log.Info("Shutting down")
		default:
			log.Warn("Shutting down")
		}
	}

	log.Close()
	os.Exit(code)
}

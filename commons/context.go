package commons

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/urfave/cli"
	"gopkg.in/ini.v1"
)

type Validator func(name string, c interface{}) error

type ValidatedFlag interface {
	cli.Flag
	Validate(c *Context) error
}

type Float64Flag struct {
	cli.Float64Flag
	Validator Validator
}

func (f Float64Flag) Validate(c *Context) error {
	n := f.Name
	if f.Validator == nil {
		return fmt.Errorf("Flag %q does not contain a Validator.", f.Name)
	}
	return f.Validator(n, c.Float64(n))
}

type StringFlag struct {
	cli.StringFlag
	Validator Validator
}

func (f StringFlag) Validate(c *Context) error {
	n := f.Name
	if f.Validator == nil {
		return fmt.Errorf("Flag %q does not contain a Validator.", f.Name)
	}
	return f.Validator(n, c.String(n))
}

type BoolFlag struct {
	cli.StringFlag
	Validator Validator
}

func (f BoolFlag) Validate(c *Context) error {
	n := f.Name
	if f.Validator == nil {
		return fmt.Errorf("Flag %q does not contain a Validator.", f.Name)
	}
	return f.Validator(n, c.Bool(n))
}

type Context struct {
	CliContext *cli.Context
	config     *ini.File
}

func NewContext(cliContext *cli.Context) *Context {
	configPath := cliContext.GlobalString("config-path")
	config := readConfiguration(configPath)
	return &Context{CliContext: cliContext, config: config}
}

func (c *Context) Int(name string) int {
	cli := c.CliContext.Int(name)
	if c.CliContext.IsSet(name) {
		return cli
	}

	profile, err := c.profile().Key(name).Int()
	if err != nil {
		return cli
	}

	return profile
}

func (c *Context) GlobalInt(name string) int {
	cli := c.CliContext.GlobalInt(name)
	if c.CliContext.GlobalIsSet(name) {
		return cli
	}

	profile, err := c.profile().Key(name).Int()
	if err != nil {
		return cli
	}

	return profile
}

func (c *Context) Duration(name string) time.Duration {
	cli := c.CliContext.Duration(name)
	return cli
}

func (c *Context) GlobalDuration(name string) time.Duration {
	cli := c.CliContext.GlobalDuration(name)
	return cli
}

func (c *Context) Float64(name string) float64 {
	cli := c.CliContext.Float64(name)
	if c.CliContext.IsSet(name) {
		return cli
	}

	profile, err := c.profile().Key(name).Float64()
	if err != nil {
		return cli
	}

	return profile
}

// TODO GlobalFloat64?

func (c *Context) Bool(name string) bool {
	cli := c.CliContext.Bool(name)
	if c.CliContext.IsSet(name) {
		return cli
	}

	profile, err := c.profile().Key(name).Bool()
	if err != nil {
		return cli
	}

	return profile
}

func (c *Context) GlobalBool(name string) bool {
	cli := c.CliContext.Bool(name)
	if c.CliContext.GlobalIsSet(name) {
		return cli
	}

	profile, err := c.profile().Key(name).Bool()
	if err != nil {
		return cli
	}

	return profile
}

func (c *Context) BoolT(name string) bool {
	cli := c.CliContext.BoolT(name)
	if c.CliContext.IsSet(name) {
		return cli
	}

	profile, err := c.profile().Key(name).Bool()
	if err != nil {
		return cli
	}

	return profile
}

// TODO GlobalBoolT?

func (c *Context) String(name string) string {
	cli := c.CliContext.String(name)
	if c.CliContext.IsSet(name) {
		return cli
	}

	profile := c.profile().Key(name).String()
	if profile != "" {
		return profile
	}

	return cli
}

func (c *Context) GlobalString(name string) string {
	cli := c.CliContext.GlobalString(name)
	if c.CliContext.GlobalIsSet(name) {
		return cli
	}

	profile := c.profile().Key(name).String()
	if profile != "" {
		return profile
	}

	return cli
}

func (c *Context) StringSlice(name string) []string {
	cli := c.CliContext.StringSlice(name)
	if c.CliContext.IsSet(name) {
		return cli
	}

	profile := c.profile().Key(name).Strings(",")
	if len(profile) > 0 {
		return profile
	}

	return cli
}

func (c *Context) GlobalStringSlice(name string) []string {
	cli := c.CliContext.GlobalStringSlice(name)
	if c.CliContext.GlobalIsSet(name) {
		return cli
	}

	profile := c.profile().Key(name).Strings(",")
	if len(profile) > 0 {
		return profile
	}

	return cli
}

func (c *Context) IntSlice(name string) []int {
	cli := c.CliContext.IntSlice(name)
	if c.CliContext.IsSet(name) {
		return cli
	}

	profile := c.profile().Key(name).Ints(",")
	if len(profile) > 0 {
		return profile
	}

	return cli
}

func (c *Context) GlobalIntSlice(name string) []int {
	cli := c.CliContext.GlobalIntSlice(name)
	if c.CliContext.GlobalIsSet(name) {
		return cli
	}

	profile := c.profile().Key(name).Ints(",")
	if len(profile) > 0 {
		return profile
	}

	return cli
}

func (c *Context) Generic(name string) interface{} {
	cli := c.CliContext.Generic(name)
	return cli
}

func (c *Context) GlobalGeneric(name string) interface{} {
	cli := c.CliContext.GlobalGeneric(name)
	return cli
}

func (c *Context) profile() *ini.Section {
	name := c.CliContext.GlobalString("profile")
	profile, err := c.config.GetSection(name)
	if err != nil {
		log.Printf("[ERROR] loading rt profile, profile: %s error: %#v", name, err)
		return c.config.Section("")
	}
	return profile
}

func readConfiguration(configDir string) *ini.File {
	if configDir == "" {
		hd, err := homedir.Dir()
		if err != nil {
			log.Printf("[ERROR] resolving current user homedir: %s", err)
			return ini.Empty()
		}
		configDir = filepath.Join(hd, ".rt")
	}

	if cd, err := os.Stat(configDir); err != nil {
		log.Printf("[WARN] configDir does not exist %s, error: %#v", configDir, err)
		return ini.Empty()
	} else if !cd.Mode().IsDir() {
		log.Printf("[WARN] configDir is not a directory %s", configDir)
		return ini.Empty()
	}

	configPath := filepath.Join(configDir, "config")
	if _, err := os.Stat(configPath); err != nil {
		log.Printf("[WARN] configPath does not exist %s, error: %#v", configPath, err)
		return ini.Empty()
	}

	config, err := ini.Load(configPath)
	if err != nil {
		log.Printf("[ERROR] loading configPath %s, error: %#v", configPath, err)
		return ini.Empty()
	}

	return config
}

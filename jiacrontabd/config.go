package jiacrontabd

import (
	"jiacrontab/pkg/util"
	"os"
	"reflect"

	"fmt"

	ini "gopkg.in/ini.v1"
)

const (
	appname = "jiacrontabd"
)

// var cfg *Config

type Config struct {
	LogLevel         string `opt:"log_level" default:"warn" comment:"日志等级"`
	VerboseJobLog    bool   `opt:"verbose_job_log" default:"true"`
	ListenAddr       string `opt:"listen_addr" default:":20001"`
	LocalAddr        string `opt:"local_addr" default:"127.0.0.1:20002"`
	AdminAddr        string `opt:"admin_addr" default:"127.0.0.1:20003"`
	LogPath          string `opt:"log_path" default:"./logs"`
	PprofAddr        string `opt:"pprof_addr" default:"127.0.0.1:20004"`
	AutoCleanTaskLog bool   `opt:"auto_clean_task_log" default:"true"`
	Hostname         string `opt:"hostname" default:""`
	CfgPath          string
	iniFile          *ini.File
	DriverName       string `opt:"driver_name" default:"sqlite3"`
	DSN              string `opt:"dsn" default:"data/jiacrontabd.db"`
}

func (c *Config) Resolve() error {
	c.iniFile = loadConfig(c.CfgPath)
	defer c.iniFile.SaveTo(c.CfgPath)
	val := reflect.ValueOf(c).Elem()
	typ := val.Type()
	hostname := util.GetHostname()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		opt := field.Tag.Get("opt")
		if opt == "" {
			continue
		}
		key := c.iniFile.Section("jiacrontabd").Key(opt)
		defaultVal := key.String()
		comment := field.Tag.Get("comment")
		key.Comment = comment

		if opt == "hostname" {
			val.Field(i).SetString(hostname)
			key.SetValue(hostname)
			continue
		}

		switch field.Type.Kind() {
		case reflect.Bool:
			setVal := false
			if defaultVal == "true" {
				setVal = true
			}
			if val.Field(i).Bool() == false {
				val.Field(i).SetBool(setVal)
			}
			key.SetValue(fmt.Sprint(val.Field(i).Bool()))
		case reflect.String:
			if val.Field(i).String() == "" {
				val.Field(i).SetString(defaultVal)
			}
			key.SetValue(val.Field(i).String())
		}
	}
	return nil
}

func NewConfig() *Config {
	return &Config{
		LogLevel:         "warn",
		VerboseJobLog:    true,
		ListenAddr:       "127.0.0.1:20001",
		LocalAddr:        "127.0.0.1:20002",
		AdminAddr:        "127.0.0.1:20003",
		LogPath:          "./logs",
		PprofAddr:        "127.0.0.1:20004",
		AutoCleanTaskLog: true,
		CfgPath:          "./jiacrontabd.ini",
		DriverName:       "sqlite3",
		DSN:              "data/jiacrontabd.db",
	}
}

func loadConfig(path string) *ini.File {
	f, err := util.TryOpen(path, os.O_CREATE)
	if err != nil {
		panic(err)
	}
	f.Close()

	iniFile, err := ini.Load(path)
	if err != nil {
		panic(err)
	}
	return iniFile
}

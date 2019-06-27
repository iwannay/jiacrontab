package jiacrontabd

import (
	"jiacrontab/pkg/file"
	"jiacrontab/pkg/util"
	"reflect"

	"github.com/iwannay/log"

	ini "gopkg.in/ini.v1"
)

const (
	appname = "jiacrontabd"
)

type Config struct {
	LogLevel         string `opt:"log_level"`
	VerboseJobLog    bool   `opt:"verbose_job_log"`
	ListenAddr       string `opt:"listen_addr"`
	AdminAddr        string `opt:"admin_addr"`
	LogPath          string `opt:"log_path"`
	PprofAddr        string `opt:"pprof_addr"`
	AutoCleanTaskLog bool   `opt:"auto_clean_task_log"`
	Hostname         string `opt:"hostname"`
	BoardcastAddr    string `opt:"boardcast_addr"`
	CfgPath          string
	Debug            bool `opt:"debug"`
	iniFile          *ini.File
	DriverName       string `opt:"driver_name"`
	DSN              string `opt:"dsn"`
}

func (c *Config) Resolve() error {
	c.iniFile = c.loadConfig(c.CfgPath)

	val := reflect.ValueOf(c).Elem()
	typ := val.Type()
	hostname := util.GetHostname()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		opt := field.Tag.Get("opt")
		if opt == "" {
			continue
		}
		sec := c.iniFile.Section("jiacrontabd")

		if opt == "hostname" && sec.Key(opt).String() == "" {
			val.Field(i).SetString(hostname)
		}

		if !sec.HasKey(opt) {
			continue
		}

		key := sec.Key(opt)
		switch field.Type.Kind() {
		case reflect.Bool:
			v, err := key.Bool()
			if err != nil {
				log.Errorf("cannot resolve ini field %s err(%v)", opt, err)
			}
			val.Field(i).SetBool(v)
		case reflect.String:
			val.Field(i).SetString(key.String())
		}
	}
	return nil
}

func NewConfig() *Config {
	return &Config{
		LogLevel:         "warn",
		VerboseJobLog:    true,
		ListenAddr:       "127.0.0.1:20001",
		AdminAddr:        "127.0.0.1:20003",
		LogPath:          "./logs",
		PprofAddr:        "127.0.0.1:20002",
		AutoCleanTaskLog: true,
		CfgPath:          "./jiacrontabd.ini",
		DriverName:       "sqlite3",
		BoardcastAddr:    util.InternalIP() + ":20001",
		DSN:              "data/jiacrontabd.db",
	}
}

func (c *Config) loadConfig(path string) *ini.File {
	if !file.Exist(path) {
		f, err := file.CreateFile(path)
		if err != nil {
			panic(err)
		}
		f.Close()
	}

	iniFile, err := ini.Load(path)
	if err != nil {
		panic(err)
	}
	return iniFile
}

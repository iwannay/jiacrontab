package main

import (
	"fmt"
	admin "jiacrontab/jiacrontab_admin"
	"jiacrontab/pkg/pprof"
	"os"

	"flag"

	"jiacrontab/pkg/version"

	"jiacrontab/pkg/util"

	"github.com/iwannay/log"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

func parseFlag(opt *admin.Config) *flag.FlagSet {

	var (
		debug    bool
		cfgPath  string
		logLevel string
	)

	flagSet := flag.NewFlagSet("jiacrontab_admin", flag.ExitOnError)
	// app options
	flagSet.Bool("version", false, "打印版本信息")
	flagSet.Bool("help", false, "帮助信息")
	flagSet.StringVar(&logLevel, "log_level", "warn", "日志级别(debug|info|warn|error)")
	flagSet.BoolVar(&debug, "debug", false, "开启debug模式")
	flagSet.StringVar(&cfgPath, "config", "./jiacrontab_admin.ini", "配置文件路径")
	// jwt options
	flagSet.Parse(os.Args[1:])

	if flagSet.Lookup("version").Value.(flag.Getter).Get().(bool) {
		fmt.Println(version.String("jiacrontab_admin"))
		os.Exit(0)
	}
	if flagSet.Lookup("help").Value.(flag.Getter).Get().(bool) {
		flagSet.Usage()
		os.Exit(0)
	}

	opt.CfgPath = cfgPath

	opt.Resolve()

	if util.HasFlagName(flagSet, "debug") {
		opt.App.Debug = debug
	}

	if util.HasFlagName(flagSet, "log_level") {
		opt.App.LogLevel = logLevel
	}

	return flagSet
}

func main() {
	cfg := admin.NewConfig()
	parseFlag(cfg)
	log.SetLevel(map[string]int{
		"debug": 0,
		"info":  1,
		"warn":  2,
		"error": 3,
	}[cfg.App.LogLevel])
	pprof.ListenPprof()
	admin := admin.New(cfg)
	admin.Main()
}

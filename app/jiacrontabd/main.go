package main

import (
	"flag"
	"fmt"
	"jiacrontab/jiacrontabd"
	"jiacrontab/pkg/pprof"
	"jiacrontab/pkg/util"
	"jiacrontab/pkg/version"

	"os"

	"github.com/iwannay/log"
)

func parseFlag(opt *jiacrontabd.Config) *flag.FlagSet {

	var (
		debug         bool
		cfgPath       string
		logLevel      string
		boardcastAddr string
	)

	flagSet := flag.NewFlagSet("jiacrontabd", flag.ExitOnError)
	flagSet.Bool("version", false, "打印版本信息")
	flagSet.Bool("help", false, "帮助信息")
	flagSet.StringVar(&logLevel, "log_level", "warn", "日志级别(debug|info|warn|error)")
	flagSet.BoolVar(&debug, "debug", false, "开启debug模式")
	flagSet.StringVar(&boardcastAddr, "boardcast_addr", "", fmt.Sprintf("广播地址(default: %s:20001)", util.InternalIP()))
	flagSet.StringVar(&cfgPath, "config", "./jiacrontabd.ini", "配置文件路径")
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

	// TODO： can be better
	if util.HasFlagName(flagSet, "log_level") {
		opt.LogLevel = logLevel
	}

	if util.HasFlagName(flagSet, "debug") {
		opt.Debug = debug
	}

	if util.HasFlagName(flagSet, "boardcast_addr") {
		opt.BoardcastAddr = boardcastAddr
	}

	if debug {
		log.JSON("debug config:", opt)
	}

	return flagSet
}

func main() {
	cfg := jiacrontabd.NewConfig()
	parseFlag(cfg)
	log.SetLevel(map[string]int{
		"debug": 0,
		"info":  1,
		"warn":  2,
		"error": 3,
	}[cfg.LogLevel])
	pprof.ListenPprof()
	jiad := jiacrontabd.New(cfg)
	jiad.Main()
}

package main

import (
	"fmt"
	"jiacrontab/jiacrontabd"
	"jiacrontab/pkg/pprof"
	"jiacrontab/pkg/version"

	"flag"

	"os"

	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

func parseFlag(opt *jiacrontabd.Config) *flag.FlagSet {
	flagSet := flag.NewFlagSet("jiacrontabd", flag.ExitOnError)

	flagSet.Bool("version", false, "打印版本信息")
	flagSet.Bool("help", false, "帮助信息")
	flagSet.Bool("init", false, "初始化配置文件")

	flagSet.StringVar(&opt.CfgPath, "config", "./jiacrontabd.ini", "配置文件路径")
	flagSet.Parse(os.Args[1:])
	if flagSet.Lookup("version").Value.(flag.Getter).Get().(bool) {
		fmt.Println(version.String("jiacrontab_admin"))
		os.Exit(0)
	}
	if flagSet.Lookup("help").Value.(flag.Getter).Get().(bool) {
		flagSet.Usage()
		os.Exit(0)
	}
	if flagSet.Lookup("init").Value.(flag.Getter).Get().(bool) {
		opt.Resolve()
		os.Exit(0)
	}
	opt.Resolve()
	return flagSet
}

func main() {
	cfg := jiacrontabd.NewConfig()
	parseFlag(cfg)

	pprof.ListenPprof()
	jiad := jiacrontabd.New(cfg)
	jiad.Main()
}

package main

import (
	"fmt"
	admin "jiacrontab/jiacrontab_admin"
	"jiacrontab/pkg/pprof"
	"os"

	"flag"

	"jiacrontab/pkg/version"

	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

func parseFlag(opt *admin.Config) *flag.FlagSet {
	flagSet := flag.NewFlagSet("jiacrontab_admin", flag.ExitOnError)

	// app options
	flagSet.Bool("version", false, "打印版本信息")
	flagSet.Bool("help", false, "帮助信息")
	flagSet.Bool("init", false, "初始化配置文件")
	flagSet.StringVar(&opt.CfgPath, "config", "./jiacrontab_admin.ini", "配置文件路径")
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

	if flagSet.Lookup("init").Value.(flag.Getter).Get().(bool) {
		opt.Resolve()
		os.Exit(0)
	}

	opt.Resolve()
	return flagSet
}

func main() {
	cfg := admin.NewConfig()
	parseFlag(cfg)

	pprof.ListenPprof()
	admin := admin.New(cfg)
	admin.Main()
}

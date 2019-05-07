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

	flagSet.StringVar(&opt.CfgPath, "config", "./jiacrontabd.ini", "配置文件路径")
	flagSet.BoolVar(&opt.VerboseJobLog, "verbose_job_log", opt.VerboseJobLog, "job日志记录冗余信息")
	flagSet.StringVar(&opt.ListenAddr, "listen_addr", opt.ListenAddr, "rpc服务监听地址")
	flagSet.StringVar(&opt.LocalAddr, "local_addr", opt.LocalAddr, "本机广播地址，用于服务器和node通信")
	flagSet.StringVar(&opt.AdminAddr, "admin_addr", opt.AdminAddr, "jiacrontab_admin通信地址")
	flagSet.StringVar(&opt.LogPath, "log_path", opt.LogPath, "job日志记录地址")
	flagSet.StringVar(&opt.PprofAddr, "pprof_addr", opt.PprofAddr, "pprof地址")
	flagSet.BoolVar(&opt.AutoCleanTaskLog, "auto_clean_task_log", opt.AutoCleanTaskLog, "自动清理时间大于一个月或者容量大于1G的job日志")
	flagSet.StringVar(&opt.Hostname, "hostname", opt.Hostname, "节点名")
	flagSet.Parse(os.Args[1:])
	if flagSet.Lookup("version").Value.(flag.Getter).Get().(bool) {
		fmt.Println(version.String("jiacrontab_admin"))
		os.Exit(0)
	}
	if flagSet.Lookup("help").Value.(flag.Getter).Get().(bool) {
		flagSet.Usage()
		os.Exit(0)
	}
	return flagSet
}

func main() {
	cfg := jiacrontabd.NewConfig()
	parseFlag(cfg)
	cfg.Resolve()

	pprof.ListenPprof()
	jiad := jiacrontabd.New(cfg)
	jiad.Main()
}

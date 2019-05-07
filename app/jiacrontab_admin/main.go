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
	flagSet.StringVar(&opt.CfgPath, "config", "./jiacrontab_admin.ini", "配置文件路径")
	flagSet.StringVar(&opt.App.HTTPListenAddr, "http_listen_addr", opt.App.HTTPListenAddr, "http监听端口")
	flagSet.StringVar(&opt.App.RPCListenAddr, "rpc_listen_addr", opt.App.RPCListenAddr, "rpc监听端口")
	flagSet.StringVar(&opt.Jwt.SigningKey, "signing_key", opt.App.SigningKey, "签名key")

	// jwt options
	flagSet.StringVar(&opt.Jwt.Name, "jwt_name", opt.Jwt.Name, "jwt保存名")
	flagSet.Int64Var(&opt.Jwt.Expires, "jwt_expires", opt.Jwt.Expires, "jwt过期时间")
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
	cfg := admin.NewConfig()
	parseFlag(cfg)
	cfg.Resolve()

	pprof.ListenPprof()
	admin := admin.New(cfg)
	admin.Main()
}

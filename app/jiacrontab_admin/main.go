package main

import (
	admin "jiacrontab/jiacrontab_admin"
	"jiacrontab/pkg/pprof"
	"os"

	"flag"

	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

func parseFlag(opt *admin.Config) *flag.FlagSet {
	flagSet := flag.NewFlagSet("jiacrontab_admin", flag.ExitOnError)

	// app options
	flagSet.Bool("version", false, "打印版本信息")
	flagSet.StringVar(&opt.CfgPath, "config", "./jiacrontab_admin.ini", "配置文件路径")
	flagSet.StringVar(&opt.App.HTTPListenAddr, "http_listen_addr", opt.App.HTTPListenAddr, "http监听端口")
	flagSet.StringVar(&opt.App.RPCListenAddr, "rpc_listen_addr", opt.App.RPCListenAddr, "rpc监听端口")
	flagSet.StringVar(&opt.Jwt.SigningKey, "signing_key", opt.App.SigningKey, "签名key")

	// jwt options
	flagSet.StringVar(&opt.Jwt.Name, "jwt_name", opt.Jwt.Name, "jwt保存名")
	flagSet.Int64Var(&opt.Jwt.Expires, "jwt_expires", opt.Jwt.Expires, "jwt过期时间")
	flagSet.Parse(os.Args[1:])

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

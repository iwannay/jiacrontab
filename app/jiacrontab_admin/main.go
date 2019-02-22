package main

import (
	admin "jiacrontab/jiacrontab_admin"
	"jiacrontab/pkg/pprof"

	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

func main() {
	pprof.ListenPprof()
	admin := admin.New()
	admin.Main()
}

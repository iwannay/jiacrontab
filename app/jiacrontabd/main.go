package main

import (
	"jiacrontab/jiacrontabd"
	"jiacrontab/pkg/pprof"

	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

func main() {

	pprof.ListenPprof()
	jiad := jiacrontabd.New()
	jiad.Main()
}

package main

import (
	"jiacrontab/jiacrontabd"

	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

func main() {
	jiad := jiacrontabd.New()
	jiad.Main()
}

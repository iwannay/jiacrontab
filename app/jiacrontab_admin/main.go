package main

import (
	admin "jiacrontab/jiacrontab_admin"

	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

func main() {
	admin := admin.New()
	admin.Main()
}

package model

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"os"
	"path/filepath"
)

var db *gorm.DB

func CreateDB(dialect string, args ...interface{}) {

	switch dialect {
	case "sqlite3":
		createSqlite3(dialect, args...)
	}

}

func createSqlite3(dialect string, args ...interface{}) {
	var err error
	if args[0] == nil {
		panic("sqlite3:db file cannot empty")
	}

	dbDir := filepath.Dir(filepath.Clean(fmt.Sprint(args[0])))
	err = os.MkdirAll(dbDir, 0755)
	if err != nil {
		panic(fmt.Errorf("sqlite3:%s", err))
	}

	db, err = gorm.Open("sqlite3", "data/jiacrontab_client.db")
	if err != nil {
		panic(err)
	}

}

func DB() *gorm.DB {
	if db == nil {
		panic("you must call CreateDb first")
	}
	return db
}

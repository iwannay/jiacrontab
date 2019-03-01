package models

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jinzhu/gorm"
)

// D alias DB
type D = gorm.DB

var db *D

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

	db, err = gorm.Open(dialect, args...)
	if err != nil {
		panic(err)
	}

}

func DB() *D {
	if db == nil {
		panic("you must call CreateDb first")
	}
	return db
}

func Transactions(fn func(tx *gorm.DB) error) error {
	if fn == nil {
		return errors.New("fn is nil")
	}
	tx := DB().Begin()
	defer func() {
		if err := recover(); err != nil {
			DB().Rollback()
		}
	}()

	if fn(tx) != nil {
		tx.Rollback()
	}
	return tx.Commit().Error
}

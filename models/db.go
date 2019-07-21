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

var (
	db        *D
	debugMode bool
)

func CreateDB(dialect string, args ...interface{}) error {
	switch dialect {
	case "sqlite3":
		return createSqlite3(dialect, args...)
	case "postgres", "mysql":
		var err error
		db, err = gorm.Open(dialect, args...)
		return err
	}
	return nil
}

func createSqlite3(dialect string, args ...interface{}) error {
	var err error
	if args[0] == nil {
		errors.New("sqlite3:db file cannot empty")
	}

	dbDir := filepath.Dir(filepath.Clean(fmt.Sprint(args[0])))
	err = os.MkdirAll(dbDir, 0755)
	if err != nil {
		return fmt.Errorf("sqlite3:%s", err)
	}

	db, err = gorm.Open(dialect, args...)
	if err == nil {
		db.DB().SetMaxOpenConns(1)
	}
	return err
}

func DB() *D {
	if db == nil {
		panic("you must call CreateDb first")
	}

	if debugMode {
		return db.Debug()
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

func InitModel(driverName string, dsn string, debug bool) error {
	if driverName == "" || dsn == "" {
		return errors.New("driverName and dsn cannot empty")
	}

	if err := CreateDB(driverName, dsn); err != nil {
		return err
	}

	debugMode = debug

	// DB().CreateTable(&Node{}, &Group{}, &User{}, &Event{}, &JobHistory{})
	DB().AutoMigrate(&Node{}, &Group{}, &User{}, &Event{}, &JobHistory{})

	DB().Create(&SuperGroup)
	return nil
}

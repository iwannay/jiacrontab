package models

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/iwannay/log"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// D alias DB
type D = gorm.DB

var (
	db        *D
	debugMode bool
)

func CreateDB(dialect string, dsn string) (err error) {
	switch dialect {
	case "sqlite3":
		return createSqlite(dsn)
	case "mysql":
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			PrepareStmt:                              true,
			DisableForeignKeyConstraintWhenMigrating: true,
		})
		return
	case "postgres":
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			PrepareStmt:                              true,
			DisableForeignKeyConstraintWhenMigrating: true,
		})
		return
	}
	return fmt.Errorf("unknow database type %s", dialect)
}

func createSqlite(dsn string) error {
	var err error
	if dsn == "" {
		return errors.New("sqlite:db file cannot empty")
	}

	dbDir := filepath.Dir(filepath.Clean(dsn))
	err = os.MkdirAll(dbDir, 0755)
	if err != nil {
		return fmt.Errorf("sqlite: makedir failed %s", err)
	}
	db, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err == nil {
		d, err := db.DB()
		if err != nil {
			panic(err)
		}
		d.SetMaxOpenConns(1)
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
	AutoMigrate()
	return nil
}

func AutoMigrate() {
	if err := DB().AutoMigrate(&SysSetting{}, &Node{}, &Group{}, &User{}, &Event{}, &JobHistory{}); err != nil {
		log.Fatal(err)
	}
	if err := DB().FirstOrCreate(&SuperGroup).Error; err != nil {
		log.Fatal(err)
	}
}

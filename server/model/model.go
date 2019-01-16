package model

import (
	"log"
)

var innerStore *Store

func InitStore(path string) {
	innerStore = NewStore(path)
	innerStore.Load()
}

func recordError(err error) {
	if err != nil {
		log.Println(err)
	}
}

type Model struct {
}

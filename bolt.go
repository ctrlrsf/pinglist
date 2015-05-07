package main

import (
	"github.com/boltdb/bolt"
)

const defaultHostDbFile = "host.db"

type BoltDbContext struct {
	db *bolt.DB
}

func (ctx *BoltDbContext) NewBoltDbContext(fileName string) {
	var err error
	ctx.db, err = bolt.Open(fileName, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
}

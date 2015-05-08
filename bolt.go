package main

import (
	"github.com/boltdb/bolt"
)

const defaultHostDbFile = "host.db"

type BoltDbContext struct {
	db *bolt.DB
}

// NewBoltDbContext creates a new BoltDbContext
func (ctx *BoltDbContext) NewBoltDbContext(fileName string) {
	var err error
	ctx.db, err = bolt.Open(fileName, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
}

// Close closes an open boltdb file
func (ctx *BoltDbContext) Close() {
	ctx.db.Close()
}

// SaveHost saves a host in the boltdb file
func (ctx *BoltDbContext) SaveHost(host Host) {
	ctx.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Hosts"))
		err := b.Put([]byte(host.Address), []byte(host.Description))
		return err
	})
}

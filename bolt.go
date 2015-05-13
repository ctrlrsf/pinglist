package main

import (
	"errors"
	"github.com/boltdb/bolt"
)

const defaultHostDbFile = "host.db"
const hostsBucketName = "hosts"

type BoltDbContext struct {
	db *bolt.DB
}

// NewBoltDbContext creates a new BoltDbContext
func NewBoltDbContext(fileName string) *BoltDbContext {
	ctx := BoltDbContext{}

	var err error
	ctx.db, err = bolt.Open(fileName, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	return &ctx
}

// Close closes an open boltdb file
func (ctx *BoltDbContext) Close() {
	ctx.db.Close()
}

// MakeHostsBucket creates the initial Hosts bucket
// Only needs to be called once when DB is created
func (ctx *BoltDbContext) MakeHostsBucket() error {
	var err error
	return ctx.db.Update(func(tx *bolt.Tx) error {
		_, err = tx.CreateBucket([]byte(hostsBucketName))
		return err
	})
}

// SaveHost saves a host in the boltdb file
func (ctx *BoltDbContext) SaveHost(host Host) error {
	return ctx.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(hostsBucketName))
		err := b.Put([]byte(host.Address), host.GobEncode())
		return err
	})
}

// GetHost retrieves a host from the boltdb file
func (ctx *BoltDbContext) GetHost(address string) (*Host, error) {
	var hostBytes []byte
	ctx.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(hostsBucketName))
		hostBytes = b.Get([]byte(address))
		return nil
	})

	if hostBytes == nil {
		return nil, errors.New("Host not found.")
	}

	host, err := GobDecodeHost(hostBytes)
	return &host, err
}

// GetAllHosts retrieves all hosts from the boltdb file
func (ctx *BoltDbContext) GetAllHosts() ([]Host, error) {
	hosts := make([]Host, 0)

	err := ctx.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(hostsBucketName))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			host, err := GobDecodeHost(v)
			if err != nil {
				return err
			}
			hosts = append(hosts, host)
		}
		return nil
	})

	return hosts, err
}

// DeleteHost removes a host in the boltdb file
func (ctx *BoltDbContext) DeleteHost(address string) {
	ctx.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(hostsBucketName))
		err := b.Delete([]byte(address))
		return err
	})
}

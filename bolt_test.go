package main

import (
	"os"
	"testing"
)

func TestSaveHostToBolt(t *testing.T) {
	tempFile := os.TempDir() + "/" + "host.db"

	testHost := Host{
		Address:     "127.0.0.1",
		Description: "Localhost",
	}

	ctx := NewBoltDbContext(tempFile)
	ctx.MakeHostsBucket()
	ctx.SaveHost(testHost)
	ctx.DeleteHost(testHost.Address)
	ctx.Close()
}

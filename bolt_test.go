package main

import (
	"os"
	"testing"
)

func TestSaveHostToBolt(t *testing.T) {
	tempFile := os.TempDir() + "/" + "host.db"

	testHost := Host{
		Address:     "127.0.0.1",
		Description: "localhost",
	}

	ctx := NewBoltDbContext(tempFile)
	ctx.MakeHostsBucket()
	ctx.SaveHost(testHost)

	savedHost := ctx.GetHost("127.0.0.1")
	t.Logf("Retrieved host: %q", savedHost)

	if savedHost.Description != "localhost" {
		t.Error("Retrieved Host struct is not what was saved: %q", savedHost)
	}

	ctx.DeleteHost(testHost.Address)
	ctx.Close()
}

package main

import (
	"os"
	"testing"
)

func TestSaveHostToBolt(t *testing.T) {
	tempFile := os.TempDir() + "/" + "host.db"

	testHosts := []Host{
		{
			Address:     "127.0.0.1",
			Description: "localhost",
		},
		{
			Address:     "8.8.8.8",
			Description: "google",
		},
	}

	ctx := NewBoltDbContext(tempFile)
	ctx.MakeHostsBucket()

	// Save all test hosts
	for i := range testHosts {
		ctx.SaveHost(testHosts[i])
	}

	savedHost := ctx.GetHost("127.0.0.1")
	t.Logf("Retrieved host: %q", savedHost)

	if savedHost.Description != "localhost" {
		t.Error("Retrieved Host struct is not what was saved: %q", savedHost)
	}

	allHosts := ctx.GetAllHosts()
	t.Logf("Size of allHosts: %d\n", len(allHosts))
	for _, v := range allHosts {
		t.Logf("Host: %v", v)
	}

	ctx.DeleteHost(testHosts[0].Address)
	ctx.Close()
}

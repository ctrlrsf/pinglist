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

	badHost, err := ctx.GetHost("8.9.9.9")
	if badHost != nil {
		t.Errorf("GetHost returned non-nil pointer for host that doesn't exist: %v\n", badHost)
	}

	if err == nil {
		t.Errorf("GetHost returned nil error for host that doesn't exist: %v\n", err)
	}

	// Save all test hosts
	for i := range testHosts {
		ctx.SaveHost(testHosts[i])
	}

	savedHost, err := ctx.GetHost("127.0.0.1")
	if err != nil {
		t.Errorf("GetHost returned non-nil error: %v", err)
	}
	t.Logf("Retrieved host: %q", savedHost)

	if savedHost.Description != "localhost" {
		t.Error("Retrieved Host struct is not what was saved: %q", savedHost)
	}

	allHosts, err := ctx.GetAllHosts()
	if err != nil {
		t.Errorf("GetAllHosts returned non-nil error: %v\n", err)
	}
	t.Logf("Size of allHosts: %d\n", len(allHosts))
	for _, v := range allHosts {
		t.Logf("Host: %v", v)
	}

	ctx.DeleteHost(testHosts[0].Address)
	ctx.Close()
}

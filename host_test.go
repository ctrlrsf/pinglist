package main

import (
	"os"
	"testing"
)

func TestHostRegistryContains(t *testing.T) {
	testDbFile := os.TempDir() + "/" + "test_host_registry.db"

	hr := NewHostRegistry(testDbFile)

	ip := "1.1.1.1"
	hr.RegisterHost(Host{Address: ip})
	result := hr.Contains(ip)
	if result == false {
		t.Error("Contains() returned false, wanted true.")
	}

	result2 := hr.Contains("2.2.2.2")
	if result2 == true {
		t.Error("Contains() returned true, wanted false.")
	}
}

func TestValidIpOrHost(t *testing.T) {
	validHosts := []string{"1.1.1.1", "2.2.2.2", "google.com", "apple.com"}

	for i := range validHosts {
		host := validHosts[i]
		result := ValidIPOrHost(host)
		if result != true {
			t.Error("ValidIPOrHost(%v) returned false, wanted true.", host)
		}
	}
}

func TestInvalidIpOrHost(t *testing.T) {
	invalidHosts := []string{"./4.2423", "x$yz", "28 42klkfjs"}

	for i := range invalidHosts {
		host := invalidHosts[i]
		result := ValidIPOrHost(host)
		if result != false {
			t.Error("ValidIPOrHost(%v) returned true, wanted false.", host)
		}
	}

}

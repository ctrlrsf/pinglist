package main

import "testing"

func TestHostRegistryContains(t *testing.T) {
	hr := NewHostRegistry()

	ip := "1.1.1.1"
	hr.RegisterAddress(ip)
	result := hr.contains(ip)
	if result == false {
		t.Error("contains() returned false, wanted true.")
	}

	result2 := hr.contains("2.2.2.2")
	if result2 == true {
		t.Error("contains() returned true, wanted false.")
	}
}

func TestValidIpOrHost(t *testing.T) {
	validHosts := []string{"1.1.1.1", "2.2.2.2", "google.com", "apple.com"}
	invalidHosts := []string{"./4.2423", "x$yz", "28 42klkfjs"}

	for i := range validHosts {
		host := validHosts[i]
		result := ValidIPOrHost(host)
		if result != true {
			t.Error("ValidIPOrHost(%v) returned false, wanted true.", host)
		}
	}

	for i := range invalidHosts {
		host := invalidHosts[i]
		result := ValidIPOrHost(host)
		if result != false {
			t.Error("ValidIPOrHost(%v) returned true, wanted false.", host)
		}
	}

}

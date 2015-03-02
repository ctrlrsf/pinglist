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

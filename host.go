package main

import (
	"net"
	"regexp"
	"time"
)

const (
	UnknownStatus = iota
	OfflineStatus
	OnlineStatus
)

// Host holds information about a host that will be pinged, such as
// IP address or hostname.
type Host struct {
	Address, Description string
	Latency              time.Duration
	Status               int
}

// HostRegistry keeps track of Hosts that will be pinged.
type HostRegistry struct {
	hosts map[string]Host
}

// NewHostRegistry returns new HostRegistry where hosts can later be added.
func NewHostRegistry() *HostRegistry {
	hr := &HostRegistry{}
	hr.hosts = make(map[string]Host)
	return hr
}

// RegisterHost adds a host to the registry
func (hr *HostRegistry) RegisterHost(h *Host) {
	// Don't add duplicate address, just return if already exists.
	if hr.Contains(h.Address) {
		return
	}

	hr.hosts[h.Address] = *h
}

// Contains checks if host list already contains a host entry with same address.
func (hr *HostRegistry) Contains(address string) bool {
	_, ok := hr.hosts[address]
	return ok
}

// RemoveHost removes a host from the registry.
func (hr *HostRegistry) RemoveHost(address string) {
	delete(hr.hosts, address)
}

// ValidIPOrHost validates address is an IP or hostname
func ValidIPOrHost(address string) bool {
	// Check if we can parse IP
	ip := net.ParseIP(address)

	// is valid IP address
	if ip != nil {
		return true
	}

	hostRe := regexp.MustCompile("^[a-zA-Z0-9][a-zA-Z0-9-\\.]{1,254}$")

	return hostRe.MatchString(address)
}

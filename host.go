package main

import (
	"net"
	"time"
)

const (
	OfflineStatus = iota
	OnlineStatus
)

// Host holds information about a host that will be pinged, such as
// IP address or hostname.
type Host struct {
	Address     string
	Latency     time.Duration
	Status      byte
	Description string
}

// HostRegistry keeps track of Hosts that will be pinged.
type HostRegistry struct {
	hostList []Host
}

// NewHostRegistry returns new HostRegistry where hosts can later be added.
func NewHostRegistry() *HostRegistry {
	return &HostRegistry{}
}

// RegisterAddress creates a new Host with specified address and adds it
// to the HostRegistry.
func (hr *HostRegistry) RegisterAddress(address string) {
	// Don't add duplicate address, just return if already exists.
	if hr.contains(address) {
		return
	}

	host := &Host{Address: address}

	hr.hostList = append(hr.hostList, *host)
}

// RegisterHost adds a host to the registry
func (hr *HostRegistry) RegisterHost(h *Host) {
	// Don't add duplicate address, just return if already exists.
	if hr.contains(h.Address) {
		return
	}

	hr.hostList = append(hr.hostList, *h)
}

// Check if host list already contains a host entry with same address.
func (hr *HostRegistry) contains(address string) bool {
	for _, host := range hr.hostList {
		if host.Address == address {
			return true
		}
	}
	return false
}

// ValidIPOrHost validates address is an IP or hostname
func ValidIPOrHost(address string) bool {
	// Check if we can parse IP
	ip := net.ParseIP(address)

	// is valid IP address
	if ip != nil {
		return true
	}

	// Check if we can resolve hostname
	_, lookupErr := net.LookupHost(address)

	// is valid hostname
	if lookupErr == nil {
		return true
	}

	return false
}

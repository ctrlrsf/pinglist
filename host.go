package main

import (
	"net"
	"regexp"
	"sync"
	"time"
)

type HostStatus int

const (
	UnknownStatus = HostStatus(0)
	OfflineStatus = HostStatus(1)
	OnlineStatus  = HostStatus(2)
)

// String returns the string representation of HostStatus
func (hs HostStatus) String() string {
	switch hs {
	case OfflineStatus:
		return "Offline"
	case OnlineStatus:
		return "Online"
	}
	return "Unknown"
}

// Host holds information about a host that will be pinged, such as
// IP address or hostname.
type Host struct {
	Address, Description string
	Latency              time.Duration
	Status               HostStatus
}

// HostRegistry keeps track of Hosts that will be pinged.
type HostRegistry struct {
	mutex *sync.RWMutex
	hosts map[string]Host
}

// NewHostRegistry returns new HostRegistry where hosts can later be added.
func NewHostRegistry() *HostRegistry {
	hr := &HostRegistry{}
	hr.mutex = &sync.RWMutex{}
	hr.hosts = make(map[string]Host)
	return hr
}

// RegisterHost adds a host to the registry
func (hr *HostRegistry) RegisterHost(h *Host) {
	// Don't add duplicate address, just return if already exists.
	if hr.Contains(h.Address) {
		return
	}

	hr.mutex.Lock()
	hr.hosts[h.Address] = *h
	hr.mutex.Unlock()
}

// Contains checks if host list already contains a host entry with same address.
func (hr *HostRegistry) Contains(address string) bool {
	hr.mutex.RLock()
	defer hr.mutex.RUnlock()

	_, ok := hr.hosts[address]
	return ok
}

// RemoveHost removes a host from the registry.
func (hr *HostRegistry) RemoveHost(address string) {
	hr.mutex.Lock()
	defer hr.mutex.Unlock()

	delete(hr.hosts, address)
}

// GetHosts returns map of hosts
func (hr *HostRegistry) GetHosts() map[string]Host {
	return hr.hosts
}

// GetHostAddresses returns map of hosts
func (hr *HostRegistry) GetHostAddresses() []string {
	list := make([]string, 0)

	for key, _ := range hr.hosts {
		list = append(list, key)
	}

	return list
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

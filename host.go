package main

import (
	"bytes"
	"encoding/gob"
	"net"
	"regexp"
	"sync"
	"time"
)

type HostStatus int

const (
	Unknown = HostStatus(0)
	Offline = HostStatus(1)
	Online  = HostStatus(2)
)

// NewHostStatus creates new HostStatus from integer value
func NewHostStatus(i int) HostStatus {
	switch i {
	case 1:
		return Offline
	case 2:
		return Online
	}
	return Unknown
}

// String returns the string representation of HostStatus
func (hs HostStatus) String() string {
	switch hs {
	case Offline:
		return "Offline"
	case Online:
		return "Online"
	}
	return "Unknown"
}

// Host holds information about a host that will be pinged, such as
// IP address or hostname.
type Host struct {
	Address     string        `json:"address"`
	Description string        `json:"description"`
	Latency     time.Duration `json:"latency"`
	Status      HostStatus    `json:"status"`
}

// GobEncode encodes a Host struct into a gob and returns the bytes
func (h *Host) GobEncode() []byte {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	enc.Encode(h)

	return buf.Bytes()
}

// GobDecodeHost decodes a Host struct from an array of bytes
func GobDecodeHost(gobBytes []byte) (Host, error) {
	buf := bytes.NewBuffer(gobBytes)
	dec := gob.NewDecoder(buf)

	var h Host
	err := dec.Decode(&h)

	return h, err
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

// GetHost returns a copy of the Host sruct for host
func (hr *HostRegistry) GetHost(address string) Host {
	hr.mutex.RLock()
	defer hr.mutex.RUnlock()

	host, _ := hr.hosts[address]
	return host
}

// UpdateHost updates a host in the registry.
func (hr *HostRegistry) UpdateHost(host Host) {
	hr.mutex.Lock()

	// Only store new host if key already exists. Possible that host was deleted
	// while a ping for that host was already in progress. This confirms host is
	// still valid before storing.
	if _, ok := hr.hosts[host.Address]; ok {
		hr.hosts[host.Address] = host
	}
	hr.mutex.Unlock()
	log.Debug("HostRegistry.Hosts = %q\n", hr.hosts)
}

// RemoveHost removes a host from the registry.
func (hr *HostRegistry) RemoveHost(address string) {
	hr.mutex.Lock()
	defer hr.mutex.Unlock()

	delete(hr.hosts, address)
}

// GetHostAddresses returns map of hosts
func (hr *HostRegistry) GetHostAddresses() []string {
	list := make([]string, 0)

	hr.mutex.RLock()
	for key, _ := range hr.hosts {
		list = append(list, key)
	}
	hr.mutex.RUnlock()

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

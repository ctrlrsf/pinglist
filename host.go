package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net"
	"regexp"
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

// Host struct is the main struct that holds information about
// a pending or completed ping test. This struct is used
// to save and retrieve results to and from a database.
type Host struct {
	Address     string        `json:"address"`
	Description string        `json:"description"`
	Latency     time.Duration `json:"latency"`
	Status      HostStatus    `json:"status"`
}

// AddressHostMap is a map of hostname addresses to Host struct. This
// gets serialized into Json for REST API
type AddressHostMap map[string]Host

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
	boltDbFile string
	boltDbCtx  *BoltDbContext
}

// NewHostRegistry returns new HostRegistry where hosts can later be added.
func NewHostRegistry(boltDbFile string) *HostRegistry {
	hr := &HostRegistry{}
	hr.boltDbFile = boltDbFile
	return hr
}

// RegisterHost adds a host to the registry
func (hr *HostRegistry) RegisterHost(h Host) {
	// Don't add duplicate address, just return if already exists.
	if hr.Contains(h.Address) {
		return
	}

	hr.boltDbCtx = NewBoltDbContext(hr.boltDbFile)
	hr.boltDbCtx.SaveHost(h)
	hr.boltDbCtx.Close()
}

// Contains checks if host list already contains a host entry with same address.
func (hr *HostRegistry) Contains(address string) bool {

	hr.boltDbCtx = NewBoltDbContext(hr.boltDbFile)
	_, err := hr.boltDbCtx.GetHost(address)
	hr.boltDbCtx.Close()

	if err == nil {
		return true
	}
	return false
}

// GetHost returns a copy of the Host sruct for host
func (hr *HostRegistry) GetHost(address string) (Host, error) {
	hr.boltDbCtx = NewBoltDbContext(hr.boltDbFile)
	host, err := hr.boltDbCtx.GetHost(address)
	hr.boltDbCtx.Close()

	return *host, err
}

// UpdateHost updates a host in the registry.
func (hr *HostRegistry) UpdateHost(host Host) {
	// Only store new host if key already exists. Possible that host was deleted
	// while a ping for that host was already in progress. This confirms host is
	// still valid before storing.
	if hr.Contains(host.Address) {
		hr.boltDbCtx = NewBoltDbContext(hr.boltDbFile)
		hr.boltDbCtx.SaveHost(host)
		hr.boltDbCtx.Close()
	}
}

// RemoveHost removes a host from the registry.
func (hr *HostRegistry) RemoveHost(address string) {
	hr.boltDbCtx = NewBoltDbContext(hr.boltDbFile)
	hr.boltDbCtx.DeleteHost(address)
	hr.boltDbCtx.Close()
}

// GetHostAddresses returns list of all host addresses
func (hr *HostRegistry) GetHostAddresses() []string {
	addressList := make([]string, 0)

	hr.boltDbCtx = NewBoltDbContext(hr.boltDbFile)
	allHosts, err := hr.boltDbCtx.GetAllHosts()
	hr.boltDbCtx.Close()

	if err != nil {
		fmt.Println("Error getting all hosts.")
	}

	for idx, _ := range allHosts {
		addressList = append(addressList, allHosts[idx].Address)
	}

	return addressList
}

// GetAllHosts returns all hosts in an AddressHostMap variable
func (hr *HostRegistry) GetAllHosts() AddressHostMap {
	allHostsMap := make(AddressHostMap)

	allAddresses := hr.GetHostAddresses()
	for _, address := range allAddresses {
		host, err := hr.GetHost(address)
		if err != nil {
			log.Warning("Unable to add host %v to map due to error: %v", address, err)
		}

		allHostsMap[address] = host
	}

	return allHostsMap
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

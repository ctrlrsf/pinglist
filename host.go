package main

// Host holds information about a host that will be pinged, such as
// IP address or hostname.
type Host struct {
	Address string
}

// HostRegistry keeps track of Hosts that will be pinged.
type HostRegistry struct {
	hostList []Host
}

// NewHostRegister returns new HostRegistry where hosts can later be added.
func NewHostRegistry() *HostRegistry {
	return &HostRegistry{}
}

// RegisterAddress creates a new Host with specified address and adds it
// to the HostRegistry.
func (hr *HostRegistry) RegisterAddress(address string) {
	host := &Host{Address: address}

	hr.hostList = append(hr.hostList, *host)
}

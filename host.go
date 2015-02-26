package pinglist

type Host struct {
	Address string
}

type HostRegistry struct {
	hostList []Host
}

func NewHostRegistry() *HostRegistry {
	return &HostRegistry{}
}

func (hr *HostRegistry) RegisterAddress(address string) {
	host := &Host{Address: address}

	hr.hostList = append(hr.hostList, *host)
}

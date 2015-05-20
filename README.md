# pinglist

Simple HTTP based server to ping list of hosts

## Goals

- Monitor uptime and latency of many hosts (10s, 100s)
- Primarily ICMP ping based monitoring
- Support adding and removing hosts via simple web UI
- Keep history of status and latency of hosts for charting
- Web UI will show auto updating status of hosts and history graphs


## External Dependencies

- InfluxDB - used for historical graphs

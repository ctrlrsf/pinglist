# pinglist

Simple HTTP based server to ping list of hosts

## Requirements

- Install go-lang
- Install react-tools for jsx compilation: `npm install -g react-tools`

## Building

Just run `make`. Sample build output below.

```
$ make
"go" build -race
cp -a ./static/ ./public
jsx static/ public/
["app"]
```

## Usage

```
$ ./pinglist -h
NAME:
   pinglist - Pinglist server

USAGE:
   pinglist [global options] command [command options] [arguments...]

VERSION:
   0.0.1

AUTHOR:
  Rene Fragoso - <ctrlrsf@gmail.com>

COMMANDS:
   help, h	Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --http ":8000"			host and port HTTP server should listen on
   --interval "5"			Time interval between checking hosts (seconds)
   --timeout "2"			Seconds to wait for reply before considering host down
   --debug				Enable debug output
   --influxurl "http://localhost:8086"	URL to InfluxDB server
   --hostdbfile "host.db"		Specify host database file
   --help, -h				show help
   --version, -v			print the version
```

## Goals

- Monitor uptime and latency of many hosts (10s, 100s)
- Primarily ICMP ping based monitoring
- Support adding and removing hosts via simple web UI
- Keep history of status and latency of hosts for charting
- Web UI will show auto updating status of hosts and history graphs


## External Dependencies

- InfluxDB - used for historical graphs

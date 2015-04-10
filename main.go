package main

import (
	"fmt"
	"net"
	"os"
	"time"

	_ "net/http/pprof"

	"github.com/codegangsta/cli"
	"github.com/tatsushid/go-fastping"
)

type HostJson struct {
	Address, Description string
}

func main() {
	app := cli.NewApp()
	app.Name = "pinglist"
	app.Author = "Rene Fragoso"
	app.Email = "ctrlrsf@gmail.com"
	app.Usage = "Pinglist server"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "http",
			Value: ":8000",
			Usage: "host and port HTTP server should listen on",
		},
		cli.IntFlag{
			Name:  "interval",
			Value: 5,
			Usage: "Time interval between checking hosts (seconds)",
		},
		cli.IntFlag{
			Name:  "timeout",
			Value: 2,
			Usage: "Seconds to wait for reply before considering host down",
		},
	}

	app.Action = func(c *cli.Context) {
		var hostRegistry *HostRegistry = NewHostRegistry()

		var historyLog *HistoryLog = NewHistoryLog()

		pingInterval := time.Duration(c.Int("interval")) * time.Second
		defaultTimeout := time.Duration(c.Int("timeout")) * time.Second

		results := make(chan Host)
		go pingLoop(results, hostRegistry, pingInterval, defaultTimeout)

		go storePingResults(results, hostRegistry, historyLog)

		startHTTPServer(c.String("http"), hostRegistry, historyLog)
	}

	app.Run(os.Args)

}

// Ping all hosts and then sleep some amount of time, repeat
func pingLoop(results chan Host, hostRegistry *HostRegistry, interval time.Duration, timeout time.Duration) {
	// Loop indefinitely
	for {
		hostRegistry.mutex.RLock()
		hostAddresses := hostRegistry.GetHostAddresses()
		hostRegistry.mutex.RUnlock()

		fmt.Printf("Host addresses: %q\n", hostAddresses)

		// Ping each host
		for _, address := range hostAddresses {
			hostRegistry.mutex.RLock()
			host, hostOk := hostRegistry.hosts[address]
			hostRegistry.mutex.RUnlock()

			// host address was not found in map, it has been deleted so
			// don't continue with pinging host.
			if !hostOk {
				continue
			}

			// Ping host in a goroutine so we can ping multiple hosts concurrently
			go pingHost(results, host, timeout)
		}

		time.Sleep(interval)
	}
}

// Process and store ping results received from results channel
func storePingResults(results chan Host, hostRegistry *HostRegistry, historyLog *HistoryLog) {
	for {
		host := <-results

		/*
			// Overwrite with new host struct
			hostRegistry.mutex.Lock()
			// Only store new host if key already exists. Possible that host was deleted
			// while a ping for that host was already in progress. This confirms host is
			// still valid before storing.
			if _, ok := hostRegistry.hosts[host.Address]; ok {
				hostRegistry.hosts[host.Address] = host
			}
			hostRegistry.mutex.Unlock()
		*/
		hostRegistry.UpdateHost(host)

		historyLog.AddLogEntry(host.Address, LogEntry{host.Status, host.Latency, time.Now()})
	}
}

// pingHost pings a host and sends result down results channel for storage
func pingHost(results chan Host, host Host, timeout time.Duration) {
	isUp, rtt, err := pingHostAddress(host.Address, timeout)

	if err != nil {
		fmt.Println(err)
	}

	if isUp {
		host.Status = OnlineStatus
		host.Latency = rtt
	} else {
		host.Status = OfflineStatus
	}
	fmt.Printf("Pinged: address=%q status=%s rtt=%s\n", host.Address, host.Status, host.Latency)

	results <- host
}

// pingHostAddress pings a host to check if host is up and records network latency
//
// host arg should be a string hostname or IP
// maxRtt is how long to wait before declaring host down
//
// Returns whether host was up, latency, and/or any error
func pingHostAddress(host string, maxRtt time.Duration) (bool, time.Duration, error) {
	var retRtt time.Duration
	var isUp = false

	p := fastping.NewPinger()
	p.MaxRTT = maxRtt
	ra, err := net.ResolveIPAddr("ip4:icmp", host)

	if err != nil {
		return false, 0, err
	}

	p.AddIPAddr(ra)
	p.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
		isUp = true
		retRtt = rtt
	}

	err = p.Run()
	if err != nil {
		return false, 0, err
	}

	return isUp, retRtt, nil
}

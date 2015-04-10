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
	for {
		hostAddresses := hostRegistry.GetHostAddresses()

		fmt.Printf("Pinging these addresses: %q\n", hostAddresses)

		for _, address := range hostAddresses {
			go pingAddress(results, address, timeout)
		}

		time.Sleep(interval)
	}
}

// Process and store ping results received from results channel
func storePingResults(results chan Host, hostRegistry *HostRegistry, historyLog *HistoryLog) {
	for {
		host := <-results

		hostRegistry.UpdateHost(host)

		historyLog.AddLogEntry(host.Address, LogEntry{host.Status, host.Latency, time.Now()})
	}
}

// pingAddress pings a host and sends result down results channel for storage
func pingAddress(results chan Host, address string, timeout time.Duration) {
	isUp, rtt, err := pingWithFastping(address, timeout)

	if err != nil {
		fmt.Println(err)
	}

	host := Host{}

	if isUp {
		host.Status = OnlineStatus
		host.Latency = rtt
	} else {
		host.Status = OfflineStatus
	}
	fmt.Printf("Pinged: address=%q status=%s rtt=%s\n", host.Address, host.Status, host.Latency)

	results <- host
}

// pingWithFastping pings a device using the fastping library and determines whether
// it is up or down, and latency
//
// address arg should be a string hostname or IP
// maxRtt is how long to wait before declaring host down
//
// Returns whether host was up, latency, and/or any error
func pingWithFastping(address string, maxRtt time.Duration) (bool, time.Duration, error) {
	var retRtt time.Duration
	var isUp = false

	p := fastping.NewPinger()
	p.MaxRTT = maxRtt
	ra, err := net.ResolveIPAddr("ip4:icmp", address)

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

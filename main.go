package main

import (
	"net"
	"os"
	"time"

	_ "net/http/pprof"

	"github.com/codegangsta/cli"
	"github.com/op/go-logging"
	"github.com/tatsushid/go-fastping"
)

var log = logging.MustGetLogger("pinglist")

type PinglistConfig struct {
	pingInterval time.Duration
	pingTimeout  time.Duration
	httpAddr     string
	hostDbFile   string
}

func main() {
	app := cli.NewApp()
	app.Name = "pinglist"
	app.Author = "Rene Fragoso"
	app.Email = "ctrlrsf@gmail.com"
	app.Usage = "Pinglist server"
	app.Version = "0.0.1"

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
		cli.BoolFlag{
			Name:  "debug",
			Usage: "Enable debug output",
		},
		cli.StringFlag{
			Name:  "hostdbfile",
			Value: defaultHostDbFile,
			Usage: "Specify host database file",
		},
	}

	app.Action = func(c *cli.Context) {
		if c.Bool("debug") {
			logging.SetLevel(logging.DEBUG, "pinglist")
		}

		config := PinglistConfig{
			pingInterval: time.Duration(c.Int("interval")) * time.Second,
			pingTimeout:  time.Duration(c.Int("timeout")) * time.Second,
			httpAddr:     c.String("http"),
			hostDbFile:   c.String("hostdbfile"),
		}

		StartPinglistServer(config)
	}

	app.Run(os.Args)
}

// StartPinglistServer runs the main server go routines
func StartPinglistServer(config PinglistConfig) {
	// Create the host registry that will keep track of the hosts
	// that we're pinging.
	var hostRegistry *HostRegistry = NewHostRegistry(config.hostDbFile)

	// Results channel receives Host structs once a host has been
	// pinged so result can be stored.
	results := make(chan Host)

	go pingLoop(results, hostRegistry, config.pingInterval, config.pingTimeout)

	go storePingResults(results, hostRegistry)

	startHTTPServer(config.httpAddr, hostRegistry)
}

// Ping all hosts and then sleep some amount of time, repeat
func pingLoop(results chan Host, hostRegistry *HostRegistry, interval time.Duration, timeout time.Duration) {
	for {
		hostAddresses := hostRegistry.GetHostAddresses()

		log.Info("Pinging these addresses: %q\n", hostAddresses)

		for _, address := range hostAddresses {
			log.Debug("Pinging: %v\n", address)

			host, err := hostRegistry.GetHost(address)
			if err != nil {
				log.Warning("GetHost() returned error=%v for address=%v", err, address)
			}

			go pingAddress(results, host, timeout)
		}

		log.Debug("Started pings for all hosts. Sleeping for: %v", interval)
		time.Sleep(interval)
	}
}

// Store ping results received from results channel
func storePingResults(results chan Host, hostRegistry *HostRegistry) {
	for {
		host := <-results

		log.Info("Storing results for host: %q\n", host)

		hostRegistry.UpdateHost(host)
	}
}

// pingAddress pings a host and sends result down results channel for storage
func pingAddress(results chan Host, oldHost Host, timeout time.Duration) {
	isUp, rtt, err := pingWithFastping(oldHost.Address, timeout)

	if err != nil {
		log.Error(err.Error())
	}

	newHost := Host{}

	newHost.Address = oldHost.Address
	newHost.Description = oldHost.Description

	if isUp {
		newHost.Status = Online
		newHost.Latency = rtt
	} else {
		newHost.Status = Offline
	}
	log.Info("Pinged: address=%q status=%s rtt=%s\n", newHost.Address, newHost.Status, newHost.Latency)

	results <- newHost
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

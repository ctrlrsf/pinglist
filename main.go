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

type HostJson struct {
	Address, Description string
}

var log = logging.MustGetLogger("pinglist")

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
		cli.BoolFlag{
			Name:  "debug",
			Usage: "Enable debug output",
		},
		cli.StringFlag{
			Name:  "influxurl",
			Value: "http://localhost:8086",
			Usage: "URL to InfluxDB server",
		},
	}

	app.Action = func(c *cli.Context) {
		if c.Bool("debug") {
			logging.SetLevel(logging.DEBUG, "pinglist")
		}

		var hostRegistry *HostRegistry = NewHostRegistry()

		var historyLog *HistoryLog = NewHistoryLog()

		pingInterval := time.Duration(c.Int("interval")) * time.Second
		defaultTimeout := time.Duration(c.Int("timeout")) * time.Second

		results := make(chan Host)
		go pingLoop(results, hostRegistry, pingInterval, defaultTimeout)

		ic := NewInfluxContext(c.String("influxurl"))
		go storePingResults(results, hostRegistry, historyLog, &ic)

		startHTTPServer(c.String("http"), hostRegistry, &ic)
	}

	app.Run(os.Args)

}

// Ping all hosts and then sleep some amount of time, repeat
func pingLoop(results chan Host, hostRegistry *HostRegistry, interval time.Duration, timeout time.Duration) {
	for {
		hostAddresses := hostRegistry.GetHostAddresses()

		log.Info("Pinging these addresses: %q\n", hostAddresses)

		for _, address := range hostAddresses {
			go pingAddress(results, address, timeout)
		}

		time.Sleep(interval)
	}
}

// Store ping results received from results channel
func storePingResults(results chan Host, hostRegistry *HostRegistry,
	historyLog *HistoryLog, influxContext *InfluxContext) {

	for {
		host := <-results

		log.Info("Storing results for host: %q\n", host)

		hostRegistry.UpdateHost(host)

		historyLog.AddLogEntry(host.Address, LogEntry{host.Status, host.Latency, time.Now()})

		influxContext.WritePoint(host.Address, host.Status, host.Latency)
	}
}

// pingAddress pings a host and sends result down results channel for storage
func pingAddress(results chan Host, address string, timeout time.Duration) {
	isUp, rtt, err := pingWithFastping(address, timeout)

	if err != nil {
		log.Error(err.Error())
	}

	host := Host{}

	host.Address = address

	if isUp {
		host.Status = Online
		host.Latency = rtt
	} else {
		host.Status = Offline
	}
	log.Info("Pinged: address=%q status=%s rtt=%s\n", host.Address, host.Status, host.Latency)

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

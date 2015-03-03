package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/codegangsta/cli"
	"github.com/gorilla/mux"
	"github.com/tatsushid/go-fastping"
)

var pingInterval = 5 * time.Second
var defaultTimeout = 2 * time.Second

var hostRegistry *HostRegistry

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
		},
	}

	app.Action = func(c *cli.Context) {
		hostRegistry = NewHostRegistry()

		hostArgs := c.Args()
		for i := range hostArgs {
			hostRegistry.RegisterAddress(hostArgs[i])
		}

		go pingLoop()

		startServer(c.String("http"))
	}

	app.Run(os.Args)

}

// Ping all hosts and then sleep some amount of time, repeat
func pingLoop() {
	// Loop indefinitely
	for {
		// Ping each host
		for i := range hostRegistry.hostList {
			host := hostRegistry.hostList[i]

			fmt.Printf("Pinging: %s\n", host)

			isUp, rtt, err := pingHost(host.Address, defaultTimeout)

			if err != nil {
				fmt.Println(err)
			}

			if isUp {
				fmt.Printf("Host is up: RTT=%s\n", rtt)
			} else {
				fmt.Println("Host is down: timeout")
			}

		}
		time.Sleep(pingInterval)
	}
}

// Pings a host to check if host is up and records network latency
//
// host arg should be a string hostname or IP
// maxRtt is how long to wait before declaring host down
//
// Returns whether host was up, latency, and/or any error
func pingHost(host string, maxRtt time.Duration) (bool, time.Duration, error) {
	var retRtt time.Duration = 0
	var isUp bool = false

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

// Handle "/api/listhosts"
func apiListHostsHandler(w http.ResponseWriter, r *http.Request) {
	jsonHostsList, err := json.Marshal(hostRegistry.hostList)

	if err != nil {
		fmt.Println(err)
		http.Error(w, "error, try again later.", 500)
	}
	fmt.Fprintln(w, string(jsonHostsList))
}

// Handle "/api/addhost"
func apiAddHostHandler(w http.ResponseWriter, r *http.Request) {
	address := r.FormValue("address")
	if address == "" {
		http.Error(w, "address format error", 500)
		return
	}

	// Check if we can parse IP
	ip := net.ParseIP(address)
	// Check if we can resolve hostname
	_, lookupErr := net.LookupHost(address)

	if ip == nil && lookupErr != nil {
		http.Error(w, "invalid address", 500)
		return
	}

	fmt.Fprintln(w, "Success")
	hostRegistry.RegisterAddress(address)
}

// Start HTTP server
func startServer(listenIpPort string) {
	r := mux.NewRouter()

	r.HandleFunc("/api/addhost", apiAddHostHandler)
	r.HandleFunc("/api/listhosts", apiListHostsHandler)
	http.Handle("/", r)
	http.ListenAndServe(listenIpPort, nil)
}

package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/codegangsta/cli"
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

		go pingLoop()

		startHTTPServer(c.String("http"))
	}

	app.Run(os.Args)

}

// Ping all hosts and then sleep some amount of time, repeat
func pingLoop() {
	// Loop indefinitely
	for {
		// Ping each host
		for i := range hostRegistry.hostList {
			host := &hostRegistry.hostList[i]

			fmt.Printf("Pinging: %s\n", host.Address)

			isUp, rtt, err := pingHost(host.Address, defaultTimeout)

			if err != nil {
				fmt.Println(err)
			}

			if isUp {
				fmt.Printf("Host is up: RTT=%s\n", rtt)
				host.Status = OnlineStatus
				host.Latency = rtt
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

// Start HTTP server
func startHTTPServer(listenIPPort string) {
	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)
	router, err := rest.MakeRouter(
		&rest.Route{"GET", "/hosts", func(w rest.ResponseWriter, r *rest.Request) {
			w.WriteJson(&hostRegistry.hostList)
		}},
		&rest.Route{"PUT", "/hosts/#address", func(w rest.ResponseWriter, r *rest.Request) {
			address := r.PathParam("address")
			if !ValidIPOrHost(address) {
				rest.Error(w, "Invalid address or format", http.StatusInternalServerError)
				return
			}
			hostRegistry.RegisterAddress(address)
		}},
	)

	if err != nil {
		log.Fatal(err)
	}
	api.SetApp(router)

	http.Handle("/api/", http.StripPrefix("/api", api.MakeHandler()))

	http.Handle("/", http.FileServer(http.Dir("static/")))

	log.Fatal(http.ListenAndServe(listenIPPort, nil))
}

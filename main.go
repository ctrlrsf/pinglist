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
		},
	}

	app.Action = func(c *cli.Context) {
		var hostRegistry *HostRegistry = NewHostRegistry()

		var historyLog *HistoryLog = NewHistoryLog()

		go pingLoop(hostRegistry, historyLog)

		startHTTPServer(c.String("http"), hostRegistry, historyLog)
	}

	app.Run(os.Args)

}

// Ping all hosts and then sleep some amount of time, repeat
func pingLoop(hostRegistry *HostRegistry, historyLog *HistoryLog) {
	// Loop indefinitely
	for {
		// Ping each host
		for _, host := range hostRegistry.hosts {
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
				host.Status = OfflineStatus
				fmt.Println("Host is down: timeout")
			}

			// Overwrite with new host struct
			hostRegistry.hosts[host.Address] = host

			historyLog.AddLogEntry(host.Address, LogEntry{host.Status, host.Latency, time.Now()})
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
func startHTTPServer(listenIPPort string, hostRegistry *HostRegistry, historyLog *HistoryLog) {
	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)
	router, err := rest.MakeRouter(
		&rest.Route{"GET", "/hosts", func(w rest.ResponseWriter, r *rest.Request) {
			w.WriteJson(&hostRegistry.hosts)
		}},
		&rest.Route{"PUT", "/hosts/#address", func(w rest.ResponseWriter, r *rest.Request) {
			hostJson := HostJson{}
			err := r.DecodeJsonPayload(&hostJson)
			if err != nil {
				rest.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if hostJson.Address == "" {
				rest.Error(w, "Address required", 400)
				return
			}

			if len(hostJson.Description) > 200 {
				rest.Error(w, "Description too long", 400)
				return
			}

			if !ValidIPOrHost(hostJson.Address) {
				rest.Error(w, "Invalid address or format", http.StatusInternalServerError)
				return
			}

			h := &Host{Address: hostJson.Address, Description: hostJson.Description}
			hostRegistry.RegisterHost(h)
		}},
		&rest.Route{"DELETE", "/hosts/#address", func(w rest.ResponseWriter, r *rest.Request) {
			address := r.PathParam("address")

			if !hostRegistry.Contains(address) {
				rest.Error(w, "Host doesn't exist", http.StatusInternalServerError)
				return
			}

			hostRegistry.RemoveHost(address)
		}},
		&rest.Route{"GET", "/history/#address", func(w rest.ResponseWriter, r *rest.Request) {
			address := r.PathParam("address")

			logEntries := historyLog.GetLogEntryList(address)

			fmt.Printf("Log entires: %q\n", logEntries)

			w.WriteJson(logEntries)
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

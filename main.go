package pinglist

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/tatsushid/go-fastping"
)

const listenPort = ":8000"

//var hostList []string
var pingInterval = 5 * time.Second
var defaultTimeout = 2 * time.Second

var hostRegistry *HostRegistry

const Usage = `Usage:
  pinglist [HOSTS...]
`

func usage() {
	fmt.Println(Usage)
	os.Exit(1)
}

func main() {
	if len(os.Args[1:]) == 0 {
		usage()
	}

	hostRegistry = NewHostRegistry()

	hostArgs := os.Args[1:]
	for i := range hostArgs {
		hostRegistry.RegisterAddress(hostArgs[i])
	}

	go pingLoop()

	startServer()

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
func apiListHandler(w http.ResponseWriter, r *http.Request) {
	jsonHostsList, err := json.Marshal(hostRegistry.hostList)

	if err != nil {
		fmt.Println(err)
		http.Error(w, "error, try again later.", 500)
	}
	fmt.Fprintln(w, string(jsonHostsList))
}

// Start HTTP server
func startServer() {
	http.HandleFunc("/api/listhosts", apiListHandler)
	http.ListenAndServe(listenPort, nil)
}

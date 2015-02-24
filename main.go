package main

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/tatsushid/go-fastping"
)

var hostList []string
var pingInterval = 5 * time.Second
var defaultTimeout = 2 * time.Second

const Usage = `Usage:
  pinglist [HOSTS...]
`

func usage() {
	fmt.Println(Usage)
	os.Exit(1)
}

func main() {
	if len(os.Args[1:]) > 0 {
		hostList = os.Args[1:]
	} else {
		usage()
	}

	pingLoop()
}

// Ping all hosts and then sleep some amount of time, repeat
func pingLoop() {
	// Loop indefinitely
	for {
		// Ping each host
		for i := range hostList {
			host := hostList[i]

			fmt.Printf("Pinging: %s\n", host)

			isUp, rtt, err := pingHost(host, defaultTimeout)

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

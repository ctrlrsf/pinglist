package main

import (
	"testing"
	"time"
)

type TestHost struct {
	address string
	status  int
	latency time.Duration
}

func TestHistoryLog(t *testing.T) {
	testHosts := []TestHost{
		{"127.0.0.1", OnlineStatus, time.Second * 2},
		{"127.0.0.2", OfflineStatus, time.Second * 1},
		{"127.0.0.3", OfflineStatus, time.Second * 1},
		{"127.0.0.4", OfflineStatus, time.Second * 1},
		{"127.0.0.5", OfflineStatus, time.Second * 1},
	}

	hl := NewHistoryLog()

	for i := range testHosts {
		le := NewLogEntry(testHosts[i].status, testHosts[i].latency, time.Now())
		hl.AddLogEntry(testHosts[i].address, le)

		leList := hl.GetLogEntryList(testHosts[i].address)
		if leList.Len() < 1 {
			t.Errorf("Bad log entry list. Expected Len() >= 1, actual = %d\n", leList.Len())
		}

		firstLe := leList.Front().Value.(LogEntry)
		if firstLe.Status != testHosts[i].status {
			t.Errorf("Bad log entry status. Expected = %d, actual = %d\n",
				testHosts[i].status, firstLe.Status)
		}

		if firstLe.Latency != testHosts[i].latency {
			t.Error("Bad latency. Expected = %d, actual = %d\n",
				testHosts[i].latency, firstLe.Latency)
		}
	}
}

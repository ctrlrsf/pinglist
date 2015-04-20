package main

import (
	"testing"
	"time"
)

type FakeResult struct {
	address string
	status  HostStatus
	latency time.Duration
}

func TestHistoryLog(t *testing.T) {
	fakeResults := []FakeResult{
		{"127.0.0.1", Online, time.Second * 2},
		{"127.0.0.1", Offline, time.Second * 1},
		{"127.0.0.1", Offline, time.Second * 1},
		{"127.0.0.1", Offline, time.Second * 1},
		{"127.0.0.1", Offline, time.Second * 1},
	}

	hl := NewHistoryLog()

	for i := range fakeResults {
		le := NewLogEntry(fakeResults[i].status, fakeResults[i].latency, time.Now())
		hl.AddLogEntry(le)
	}

	logEntries := hl.GetLogEntries()
	if len(logEntries) < 1 {
		t.Errorf("Bad log entry list. Expected Len() >= 1, actual = %d\n", len(logEntries))
	}

	firstLe := logEntries[0]
	if firstLe.Status != fakeResults[0].status {
		t.Errorf("Bad log entry status. Expected = %d, actual = %d\n",
			fakeResults[0].status, firstLe.Status)
	}

	if firstLe.Latency != fakeResults[0].latency {
		t.Error("Bad latency. Expected = %d, actual = %d\n",
			fakeResults[0].latency, firstLe.Latency)
	}
}

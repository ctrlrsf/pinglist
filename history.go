package main

import (
	"container/list"
	"time"
)

type LogEntry struct {
	Status    int
	Latency   time.Duration
	Timestamp time.Time
}

type HistoryLog struct {
	Log map[string]*list.List
}

// NewLogEntry creates and returns a new LogEntry object
func NewLogEntry(status int, latency time.Duration, timestamp time.Time) LogEntry {
	return LogEntry{status, latency, timestamp}
}

// NewHistoryLog creates a new HistoryLog and map of address to
// log entries
func NewHistoryLog() HistoryLog {
	hl := HistoryLog{}
	hl.Log = make(map[string]*list.List)
	return hl
}

// AddLogEntry adds a log entry to history log for a host
func (h *HistoryLog) AddLogEntry(address string, logEntry LogEntry) {
	if _, ok := h.Log[address]; !ok {
		h.Log[address] = list.New()
	}

	h.Log[address].PushBack(logEntry)
}

// GetLogEntryList returns the list of log entries for a host
func (h *HistoryLog) GetLogEntryList(address string) *list.List {
	return h.Log[address]
}

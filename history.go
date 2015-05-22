package main

import (
	"time"
)

type LogEntry struct {
	Status    HostStatus    `json:"status"`
	Latency   time.Duration `json:"latency"`
	Timestamp time.Time     `json:"timestamp"`
}

type HistoryLog struct {
	logEntries []LogEntry
}

// AddLogEntry adds a log entry to history log for a host
func (h *HistoryLog) AddLogEntry(logEntry LogEntry) {
	h.logEntries = append(h.logEntries, logEntry)
}

// GetLogEntries returns the logEntries
func (h *HistoryLog) GetLogEntries() []LogEntry {
	return h.logEntries
}

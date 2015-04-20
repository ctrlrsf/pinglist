package main

import (
	"time"
)

type LogEntry struct {
	Status    HostStatus
	Latency   time.Duration
	Timestamp time.Time
}

type HistoryLog struct {
	logEntries []LogEntry
}

// NewLogEntry creates and returns a new LogEntry object
func NewLogEntry(status HostStatus, latency time.Duration, timestamp time.Time) LogEntry {
	return LogEntry{status, latency, timestamp}
}

// NewHistoryLog creates a new HistoryLog struct with defaults
func NewHistoryLog() HistoryLog {
	hl := HistoryLog{}
	return hl
}

// AddLogEntry adds a log entry to history log for a host
func (h *HistoryLog) AddLogEntry(logEntry LogEntry) {
	h.logEntries = append(h.logEntries, logEntry)
}

// GetLogEntries returns the logEntries
func (h *HistoryLog) GetLogEntries() []LogEntry {
	return h.logEntries
}

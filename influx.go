package main

import (
	"encoding/json"
	"fmt"
	"github.com/influxdb/influxdb/client"
	"net/url"
	"os"
	"time"
)

const (
	InfluxDatabase        = "pinglist"
	InfluxRetentionPolicy = "default"
	PingListSeriesName    = "hostHistory"
)

// Store context information for Influx server to assist with new
// connections and queries.
type InfluxContext struct {
	database        string
	config          client.Config
	client          *client.Client
	retentionPolicy string
	version         string
}

// NewInfluxContext creates a new Influx context using URL of InfluxDB server
// and environment variabls for username and password.
func NewInfluxContext(uri string) InfluxContext {
	u, err := url.Parse(uri)
	if err != nil {
		log.Fatal(err)
	}

	ic := InfluxContext{}
	ic.database = InfluxDatabase

	ic.config = client.Config{
		URL:      *u,
		Username: os.Getenv("INFLUX_USER"),
		Password: os.Getenv("INFLUX_PWD"),
	}

	ic.client, err = client.NewClient(ic.config)
	if err != nil {
		log.Fatal(err)
	}

	// Ping to determine version
	ic.Ping()

	return ic
}

// Ping pings Influx DB server to get version in case we have to
// use different client API (v0.8 vs v0.9)
func (ic *InfluxContext) Ping() error {
	duration, version, err := ic.client.Ping()
	if err != nil {
		return err
	}
	log.Info("Influx client ping: %v, %s", duration, version)

	ic.version = version

	return nil
}

// WritePoint writes a host status data point to Influx server
func (ic *InfluxContext) WritePoint(timestamp time.Time, host string,
	status HostStatus, latency time.Duration) error {
	log.Debug("Writing points: status=%q, latency=%q", status, latency)

	point := client.Point{
		Name: PingListSeriesName,
		Fields: map[string]interface{}{
			"status":  status,
			"latency": latency,
		},
		Tags: map[string]string{
			"host": host,
		},
		Timestamp: timestamp,
		Precision: "s",
	}

	batchPoints := client.BatchPoints{
		Points:          []client.Point{point},
		Database:        ic.database,
		RetentionPolicy: ic.retentionPolicy,
	}

	_, err := ic.client.Write(batchPoints)
	if err != nil {
		return err
	}

	return nil
}

// Query queries Influx DB for all history for a specific host
func (ic *InfluxContext) Query(host string) ([]client.Result, error) {
	result := []client.Result{}

	command := fmt.Sprintf("SELECT status, latency FROM %s WHERE host='%s'", PingListSeriesName, host)

	query := client.Query{
		Command:  command,
		Database: ic.database,
	}

	response, err := ic.client.Query(query)
	if err != nil {
		return result, err
	}

	if response.Error() != nil {
		return result, response.Error()
	}

	results := response.Results

	return results, nil
}

// influxResultsToHistoryLog converts the Influx DB query results into
// a HistoryLog data structure that's used by rest of application
func influxResultsToHistoryLog(results []client.Result) HistoryLog {
	historyLog := HistoryLog{}

	historyLog.logEntries = []LogEntry{}

	for resultsIndex := range results {
		result := results[resultsIndex]
		for seriesIndex := range result.Series {
			series := result.Series[seriesIndex]
			log.Debug("Series name: %q", series.Name)
			log.Debug("Series columns: %q", series.Columns)
			for valueIndex := range series.Values {
				value := series.Values[valueIndex]

				t, _ := time.Parse(time.RFC3339, value[0].(string))

				statusJsonNumber := value[1].(json.Number)
				log.Debug("statusJsonNumber: %q", statusJsonNumber)
				statusInt64, _ := statusJsonNumber.Int64()
				log.Debug("statusInt64: %q", statusInt64)
				status := NewHostStatus(int(statusInt64))
				log.Debug("status: %q", status)

				latencyDurationJsonNumber := value[2].(json.Number)
				log.Debug("latencyDurationJsonNumber: %q", latencyDurationJsonNumber)
				latencyDurationInt64, _ := latencyDurationJsonNumber.Int64()
				log.Debug("latencyDurationInt64: %q", latencyDurationInt64)
				latencyDuration := time.Duration(latencyDurationInt64)

				logEntry := LogEntry{
					Timestamp: t,
					Status:    status,
					Latency:   latencyDuration,
				}

				historyLog.logEntries = append(historyLog.logEntries, logEntry)
			}
		}
	}

	log.Debug("returning historyLog: %q", historyLog)

	return historyLog
}

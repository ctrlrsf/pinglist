package main

import (
	"fmt"
	"github.com/influxdb/influxdb/client"
	"net/url"
	"os"
	"time"
)

const (
	InfluxDatabase        = "pinglist"
	InfluxRetentionPolicy = "default"
)

type InfluxContext struct {
	database        string
	config          client.Config
	client          *client.Client
	retentionPolicy string
	version         string
}

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

func (ic *InfluxContext) Ping() error {
	duration, version, err := ic.client.Ping()
	if err != nil {
		return err
	}
	log.Info("Influx client ping: %v, %s", duration, version)

	ic.version = version

	return nil
}

func (ic *InfluxContext) WritePoint(host, name, value string) error {
	point := client.Point{
		Name: name,
		Fields: map[string]interface{}{
			"value": value,
		},
		Tags: map[string]string{
			"host": host,
		},
		Timestamp: time.Now(),
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

func (ic *InfluxContext) Query(host, name string) ([]client.Result, error) {
	result := []client.Result{}

	command := fmt.Sprintf("SELECT value FROM %s WHERE host='%s'", name, host)

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

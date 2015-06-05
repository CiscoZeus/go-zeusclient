package main

import (
	"fmt"
	. "github.com/CiscoZeus/go-zeusclient"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func randString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func main() {
	zeus := &Zeus{ApiServ: "http://api.ciscozeus.io", Token: "{Your token}"}

	logs := LogList{
		Name: randString(5),
		Logs: []Log{
			Log{"foo": "bar", "tar": "woo"},
			Log{"timestamp": time.Now().Unix(), "tar": "woo"},
		},
	}
	fmt.Println("going to send two logs:")
	fmt.Printf("%+v\n", logs)
	suc, err := zeus.PostLogs(logs)
	if err != nil {
		panic("failed to send logs")
	}
	fmt.Printf("sent 2 logs, %d successful\n", suc)

	fmt.Println("\nsleep 1 second")
	time.Sleep(1 * time.Second)
	total, rLogs, err := zeus.GetLogs(logs.Name, "", "", 0, 0, 0, 0)
	if err != nil {
		panic("failed to get logs")
	}
	fmt.Printf("received %d logs:\n", total)
	fmt.Printf("%+v\n", rLogs)

	metrics := MetricList{
		Name:    randString(5),
		Columns: []string{"col1", "col2", "col3"},
		Metrics: []Metric{
			Metric{
				Timestamp: float64(time.Now().Unix()),
				Point:     []float64{1.0, 2.0, 3.0},
			},
			Metric{
				Timestamp: float64(time.Now().Unix()),
				Point:     []float64{1.0, 4.0, 9.0},
			},
		},
	}
	fmt.Println("going to send two metric datapoints to a series")
	fmt.Printf("%+v\n", metrics)
	suc, err = zeus.PostMetrics(metrics)
	if err != nil {
		panic("failed to send metrics: " + err.Error())
	}
	fmt.Printf("sent 2 metrics, %d successful\n", suc)

	fmt.Println("\nsleep for 20 seconds")
	time.Sleep(20 * time.Second)
	name, err := zeus.GetMetricNames(metrics.Name, 0)
	if err != nil {
		panic("failed to retrieve metric name")
	}
	fmt.Printf("found log name: %v\n", name)

	rMetrics, err := zeus.GetMetricValues(metrics.Name, "", "", 0, 0, "", 0, 0)
	if err != nil {
		panic("failed to retrieve metric values")
	}
	fmt.Printf("%+v\n", rMetrics)

	fmt.Println("going to delete the series we just created")
	succ, err := zeus.DeleteMetrics(metrics.Name)
	if err != nil || succ != true {
		panic("failed to delete one series")
	}
	fmt.Println("deleted")
}

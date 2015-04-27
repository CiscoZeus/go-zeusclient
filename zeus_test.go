package zeus

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

var _ = fmt.Printf
var zeus *Zeus

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
	zeus = &Zeus{username: "marc", password: "123", apiServ: "http://api.ciscozeus.io/"}
	zeus.Auth()
}

func randString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func TestPostLogs(t *testing.T) {
	logs := make([]Log, 1)
	logs[0] = Log{Message: "Message from Go"}
	successful, err := zeus.PostLogs("go_log", logs)
	if err != nil {
		t.Error("failed to post logs:", err)
	}
	if successful != 1 {
		t.Errorf("successful=%d != 1", successful)
	}
}

func TestGetLogs(t *testing.T) {
	message := randString(10)
	total, logs, err := zeus.GetLogs(message, "", "", 0, 10)
	if total != 0 {
		t.Error("got not existing log:", logs)
	}

	expLogs := Logs{{Message: message}}
	successful, err := zeus.PostLogs("go_log", expLogs)
	if err != nil {
		t.Error("failed to post logs:", err)
	}
	if successful != 1 {
		t.Errorf("successful=%d != 1", successful)
	}

	time.Sleep(time.Second)
	total, logs, err = zeus.GetLogs(message, "", "", 0, 10)
	if err != nil {
		t.Error("failed to retrieve logs:", err)
	}
	if total == 0 || logs[0].Message != message {
		t.Error("failed to retrieve log:", expLogs)
	}
}

func TestPostMetrics(t *testing.T) {
	metricName := randString(5)
	value := rand.Float64()
	metrics := Metrics{{Value: value}}
	successful, err := zeus.PostMetrics(metricName, metrics)
	if err != nil || successful != 1 {
		t.Errorf("failed to post metrics, %d successful", successful)
	}
}

func TestGetMetricNames(t *testing.T) {
	metricName := randString(5)
	value := rand.Float64()
	metrics := Metrics{{Value: value}}
	successful, err := zeus.PostMetrics(metricName, metrics)
	if err != nil || successful != 1 {
		t.Errorf("failed to post metrics, %d successful", successful)
	}
	time.Sleep(time.Second)
	names, err := zeus.GetMetricNames(metricName, 0)
	if err != nil {
		t.Error("failed to get metrics' name:", err)
	}
	found := false
	for _, val := range names {
		if val == metricName {
			found = true
			break
		}
	}
	if found == false {
		t.Error("failed to retrieve metrics' name")
	}
}

func TestGetMetricValues(t *testing.T) {
	metricName := randString(5)
	value := rand.Float64()
	metrics := Metrics{{Value: value}}
	successful, err := zeus.PostMetrics(metricName, metrics)
	if err != nil || successful != 1 {
		t.Errorf("failed to post metrics, %d successful", successful)
	}

	time.Sleep(time.Second)
	multiMetrics, err := zeus.GetMetricValues(metricName, "", "", "", "", "", 0)
	if err != nil {
		t.Error("failed to get metric values:", err)
	}
	if len(multiMetrics) == 0 ||
		len(multiMetrics[metricName+"_Value"]) == 0 ||
		multiMetrics[metricName+"_Value"][0].Value != value {
		t.Error("failed to retrieve metric values")
	}
}

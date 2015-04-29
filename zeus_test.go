package zeus

import (
	"math/rand"
	"testing"
	"time"
)

var zeus *Zeus

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
	zeus = &Zeus{apiServ: "http://api.ciscozeus.io/", token: "a7b71e4b"}
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
	logType := randString(5)
	logs := make([]Log, 1)
	logs[0] = Log{Timestamp: time.Now().Unix(), Message: "Message from Go"}
	successful, err := zeus.PostLogs(logType, logs)
	if err != nil {
		t.Error("failed to post logs:", err)
	}
	if successful != 1 {
		t.Errorf("successful=%d != 1", successful)
	}
}

func TestGetLogs(t *testing.T) {
	message := randString(10)
	timestamp := time.Now().Unix()
	logType := randString(5)
	total, logs, err := zeus.GetLogs(logType, "", 0, 0, 0, 10)
	if total != 0 {
		t.Error("got not existing log:", logs)
	}

	expLogs := Logs{{Timestamp: timestamp, Message: message},
		{Timestamp: timestamp + 10, Message: message + "2"}}
	successful, err := zeus.PostLogs(logType, expLogs)
	if err != nil {
		t.Error("failed to post logs:", err)
	}
	if successful != 2 {
		t.Errorf("successful=%d != 2", successful)
	}

	time.Sleep(time.Second)
	total, logs, err = zeus.GetLogs(logType, message, 0, timestamp+5, 0, 10)
	if err != nil {
		t.Error("failed to retrieve logs:", err)
	}
	if total == 0 ||
		logs[0].Message != message || logs[0].Timestamp != timestamp {
		t.Errorf("failed to retrieve log: expect %#v, got %#v", expLogs, logs)
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
	multiMetrics, err := zeus.GetMetricValues(metricName, "", "", 0, 0, "", 0)
	if err != nil {
		t.Error("failed to get metric values:", err)
	}
	if len(multiMetrics) == 0 ||
		len(multiMetrics[metricName+"_value"]) == 0 ||
		multiMetrics[metricName+"_value"][0].Value != value {
		t.Errorf("failed to retrieve metric values, expect %#v, got %#v", metrics, multiMetrics)
	}
}

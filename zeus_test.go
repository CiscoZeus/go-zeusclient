package zeus

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func mock(expPath string, expParam *url.Values, code int, retBody string) (
	*httptest.Server, *Zeus) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			reqBody, _ := ioutil.ReadAll(r.Body)
			if r.Method == "GET" {
				expPath += "?" + expParam.Encode()
			}
			if expPath != r.RequestURI ||
				(r.Method == "POST" && string(reqBody) != expParam.Encode()) {
				w.WriteHeader(400)
			} else {
				w.WriteHeader(code)
			}
			fmt.Fprintln(w, retBody)
		}))

	// Initialize a Zeus client
	zeus := &Zeus{apiServ: server.URL, token: "goZeus"}
	return server, zeus
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
	logName := randString(5)
	logs := make([]Log, 1)
	logs[0] = Log{Timestamp: time.Now().Unix(), Message: "Message from Go"}

	jsonStr, _ := json.Marshal(logs)
	param := url.Values{"logs": {string(jsonStr)}}
	server, zeus := mock("/logs/goZeus/"+logName+"/", &param, 200, `{"successful": 1}`)
	defer server.Close()

	successful, err := zeus.PostLogs(logName, logs)
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
	logName := randString(5)

	param := url.Values{
		"log_name": {logName},
		"pattern":  {message},
		"from":     {strconv.FormatInt(timestamp, 10)},
		"to":       {strconv.FormatInt(timestamp+10, 10)},
		"limit":    {"10"}}
	retBody := fmt.Sprintf("{\"total\": 1,\"result\": [{\"timestamp\": %d, \"message\":\"%s\"}]}",
		timestamp, message)
	server, zeus := mock("/logs/goZeus/", &param, 200, retBody)
	defer server.Close()

	total, logs, err := zeus.GetLogs(logName, message, timestamp, timestamp+10, 0, 10)

	if total != 1 || logs[0].Message != message {
		t.Error("failed to retrieve logs:", err)
	}
}

func TestPostMetrics(t *testing.T) {
	metricName := randString(5)
	value := rand.Float64()
	metrics := Metrics{{Value: value}}
	jsonStr, _ := json.Marshal(metrics)

	param := url.Values{"metrics": {string(jsonStr)}}
	retBody := `{"successful":1}`
	server, zeus := mock("/metrics/goZeus/"+metricName+"/", &param, 200, retBody)
	defer server.Close()

	successful, err := zeus.PostMetrics(metricName, metrics)
	if err != nil || successful != 1 {
		t.Errorf("failed to post metrics, %d successful", successful)
	}
}

func TestGetMetricNames(t *testing.T) {
	metricName := randString(5)

	param := url.Values{"metric_name": {metricName}, "limit": {"1024"}}
	retBody := `["Harry", "Potter"]`
	server, zeus := mock("/metrics/goZeus/_names/", &param, 200, retBody)
	defer server.Close()

	names, err := zeus.GetMetricNames(metricName, 1024)
	if err != nil {
		t.Error("failed to get metrics' name:", err)
	}
	foundH, foundP := false, false
	for _, val := range names {
		if val == "Harry" {
			foundH = true
		} else if val == "Potter" {
			foundP = true
		}
	}
	if foundH == false || foundP == false {
		t.Error("failed to retrieve metrics' name")
	}
}

func TestGetMetricValues(t *testing.T) {
	metricName := "Jon.Snow"
	timestamp := int64(1430355869000)

	param := url.Values{
		"metric_name":      {metricName},
		"from":             {strconv.FormatInt(timestamp-int64(10*1000), 10)},
		"to":               {strconv.FormatInt(timestamp, 10)},
		"filter_condition": {"value>10"},
		"limit":            {"1024"}}
	retBody := `[{"points": [[1430355865000,144740003,58.8]],"name": "Jon.Snow","columns": ["time","sequence_number","value"]}]`
	server, zeus := mock("/metrics/goZeus/_values/", &param, 200, retBody)
	defer server.Close()

	multiMetrics, err := zeus.GetMetricValues(metricName, "", "", timestamp-int64(10*1000), timestamp, "value>10", 1024)
	if err != nil {
		t.Error("failed to get metric values:", err)
	}
	expMetric := Metric{Timestamp: 1430355865000, Value: 58.8}
	// Two colume: sequence_number and value
	if len(multiMetrics) != 2 ||
		len(multiMetrics[metricName+"_value"]) != 1 ||
		multiMetrics[metricName+"_value"][0] != expMetric {
		t.Errorf("failed to retrieve metric values, expect %#v, got %#v", expMetric, multiMetrics)
	}
}

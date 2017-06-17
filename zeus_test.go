// Copyright 2015 Cisco Systems, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// 	Unless required by applicable law or agreed to in writing, software
// 	distributed under the License is distributed on an "AS IS" BASIS,
// 	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// 	See the License for the specific language governing permissions and
// 	limitations under the License.

package zeus

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strconv"
	"testing"
	"time"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func mock(expPath string, expParam *url.Values, code int, retBody string) (
	*httptest.Server, *Zeus, string) {

	organizationAndBucket := "org1/bucket1"
	token := "goZeus"

	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			reqBody, _ := ioutil.ReadAll(r.Body)

			if r.Header["Bucket-Name"][0] != "org1/bucket1" || r.Header["Authorization"][0] != ("Bearer "+token) {
				w.WriteHeader(400)
				fmt.Fprintln(w, retBody)
				return
			}

			if r.Method == "GET" {
				expPath += "?" + expParam.Encode()
				pathMatch, _ := regexp.MatchString(`/$`, r.RequestURI) // For golang 1.6.2
				if pathMatch {
					r.RequestURI += "?"
				}
				if expPath != r.RequestURI {
					w.WriteHeader(400)
				} else {
					w.WriteHeader(code)
				}
			} else if r.Method == "POST" || r.Method == "PUT" || r.Method == "DELETE" {
				if expPath != r.RequestURI {
					w.WriteHeader(400)
				} else if string(reqBody) != expParam.Encode() {
					w.WriteHeader(400)
				} else {
					w.WriteHeader(code)
				}
			}
			fmt.Fprintln(w, retBody)
			return
		}))

	// Initialize a Zeus client
	zeus := &Zeus{ApiServ: server.URL, Token: token}
	return server, zeus, organizationAndBucket
}

func randString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func TestPostAlert(t *testing.T) {
	alert := Alert{
		Alert_name:       randString(5),
		Username:         randString(5),
		Token:            "databucket",
		Alerts_type:      "metric",
		Alert_expression: "cpu.value > 20",
		Alert_severity:   "S1",
		Metric_name:      "cpu.value",
		Emails:           "blah@blah.com",
		Status:           "active",
		Frequency:        30.0,
	}
	data := make(url.Values)
	setAlertToUrlValues(alert, &data)
	server, zeus, bucket_name := mock("/alerts/goZeus/", &data, 201, `{"successful": 1}`)
	defer server.Close()

	token := zeus.Token
	zeus.Token = ""
	_, err := zeus.bucket(bucket_name).PostAlert(Alert{})
	if err == nil {
		t.Error("should fail on empty token")
	}
	zeus.Token = token

	_, err = zeus.bucket(bucket_name).PostAlert(Alert{})
	if err != nil {
		t.Error("should fail on empty alert")
	}

	successful, err := zeus.bucket(bucket_name).PostAlert(alert)
	if err != nil {
		t.Error("failed to post logs:", err)
	}
	if successful != 1 {
		t.Errorf("successful=%d != 1", successful)
	}
}

func TestZeusGetAlerts(t *testing.T) {
	alert := Alert{
		Alert_name:       randString(5),
		Username:         randString(5),
		Token:            "databucket",
		Alerts_type:      "metric",
		Alert_expression: "cpu.value > 20",
		Alert_severity:   "S1",
		Metric_name:      "cpu.value",
		Emails:           "blah@blah.com",
		Status:           "active",
		Frequency:        30.0,
	}
	metric := ""
	param := make(url.Values)
	if len(metric) > 0 {
		param.Add("metric", metric)
	}
	jsonStr, _ := json.Marshal([]Alert{alert})
	server, zeus, bucket_name := mock("/alerts/goZeus/", &param, 200, string(jsonStr))
	defer server.Close()

	token := zeus.Token
	zeus.Token = ""
	_, _, err := zeus.bucket(bucket_name).GetAlerts()
	if err == nil {
		t.Error("should fail on empty token")
	}
	zeus.Token = token

	total, alerts, err := zeus.bucket(bucket_name).GetAlerts()
	if err != nil {
		t.Error("failed to retrieve alerts:", err)
	}
	if total != 1 || alerts[0].Token != "databucket" {
		t.Error("Retrieved alerts are wrong: ", total, alerts)
	}
}

func TestPutAlert(t *testing.T) {
	alert := Alert{
		Alert_name:       randString(5),
		Username:         randString(5),
		Token:            "databucket",
		Alerts_type:      "metric",
		Alert_expression: "cpu.value > 20",
		Alert_severity:   "S1",
		Metric_name:      "cpu.value",
		Emails:           "blah@blah.com",
		Status:           "active",
		Frequency:        30.0,
	}
	data := make(url.Values)
	setAlertToUrlValues(alert, &data)
	server, zeus, bucket_name := mock("/alerts/goZeus/1/", &data, 200, `{"successful": 1}`)
	defer server.Close()

	token := zeus.Token
	zeus.Token = ""
	_, err := zeus.bucket(bucket_name).PutAlert(1, Alert{})
	if err == nil {
		t.Error("should fail on empty token")
	}
	zeus.Token = token

	_, err = zeus.bucket(bucket_name).PutAlert(1, Alert{})
	if err != nil {
		t.Error("should fail on empty alert")
	}

	successful, err := zeus.bucket(bucket_name).PutAlert(1, alert)
	if err != nil {
		t.Error("failed to put alert:", err)
	}
	if successful != 1 {
		t.Errorf("successful=%d != 1", successful)
	}
}

func TestZeusGetAlert(t *testing.T) {
	sample_alert := Alert{
		Alert_name:       randString(5),
		Username:         randString(5),
		Token:            "databucket",
		Alerts_type:      "metric",
		Alert_expression: "cpu.value > 20",
		Alert_severity:   "S1",
		Metric_name:      "cpu.value",
		Emails:           "blah@blah.com",
		Status:           "active",
		Frequency:        30.0,
	}
	param := make(url.Values)
	jsonStr, _ := json.Marshal(sample_alert)
	server, zeus, bucket_name := mock("/alerts/goZeus/1/", &param, 200, string(jsonStr))
	defer server.Close()

	token := zeus.Token
	zeus.Token = ""
	_, err := zeus.bucket(bucket_name).GetAlert(1)
	if err == nil {
		t.Error("should fail on empty token")
	}
	zeus.Token = token

	alert, err := zeus.bucket(bucket_name).GetAlert(1)
	if err != nil {
		t.Error("failed to retrieve alert:", err)
	}
	if alert.Token != "databucket" {
		t.Error("Retrieved alerts are wrong: ", alert)
	}
}

func TestZeusDeleteAlert(t *testing.T) {
	param := make(url.Values)
	server, zeus, bucket_name := mock("/alerts/goZeus/1/", &param, 204, `{"successful": 1}`)

	defer server.Close()

	token := zeus.Token
	zeus.Token = ""
	_, err := zeus.bucket(bucket_name).DeleteAlert(1)
	if err == nil {
		t.Error("should fail on empty token")
	}
	zeus.Token = token

	successful, err := zeus.bucket(bucket_name).DeleteAlert(1)

	if err != nil {
		t.Error("failed to delete logs:", err)
	}
	if successful != 1 {
		t.Errorf("successful=%d != 1", successful)
	}
}

// Logs test
func TestPostLogs(t *testing.T) {
	logName := randString(5)
	log := Log{"timestamp": time.Now().Unix(), "message": "Message from Go"}
	logs := LogList{Name: logName, Logs: []Log{log}}

	jsonStr, _ := json.Marshal(logs)
	param := url.Values{"logs": {string(jsonStr)}}
	server, zeus, bucket_name := mock("/logs/goZeus/"+logName+"/", &param, 200, `{"successful": 1}`)
	defer server.Close()

	token := zeus.Token
	zeus.Token = ""
	_, err := zeus.bucket(bucket_name).PostLogs(LogList{})
	if err == nil {
		t.Error("should fail on empty token")
	}
	zeus.Token = token

	successful, err := zeus.bucket(bucket_name).PostLogs(LogList{})
	if err == nil {
		t.Error("should fail on empty log")
	}

	successful, err = zeus.bucket(bucket_name).PostLogs(logs)
	if err != nil {
		t.Error("failed to post logs:", err)
	}
	if successful != 1 {
		t.Errorf("successful=%d != 1", successful)
	}
}

func TestGetLogs(t *testing.T) {
	pattern := randString(10)
	timestamp := time.Now().Unix()
	logName := randString(5)

	param := url.Values{
		"log_name":       {logName},
		"attribute_name": {"message"},
		"pattern":        {pattern},
		"from":           {strconv.FormatInt(timestamp, 10)},
		"to":             {strconv.FormatInt(timestamp+10, 10)},
		"limit":          {"10"}}
	retBody := fmt.Sprintf("{\"total\": 1,\"result\": [{\"timestamp\": %d, \"message\":\"%s\"}]}",
		timestamp, pattern)
	server, zeus, bucket_name := mock("/logs/goZeus/", &param, 200, retBody)
	defer server.Close()

	token := zeus.Token
	zeus.Token = ""
	_, _, err := zeus.bucket(bucket_name).GetLogs("", "", "", 0, 0, 0, 0)
	if err == nil {
		t.Error("should fail on empty token")
	}
	zeus.Token = token

	total, logs, err := zeus.bucket(bucket_name).GetLogs("", "", "", 0, 0, 0, 0)
	if err == nil {
		t.Error("test should failed because of missing parameters")
	}

	total, logs, err = zeus.bucket(bucket_name).GetLogs(logName, "message", pattern, timestamp,
		timestamp+10, 0, 10)

	if total != 1 || logs.Logs[0]["message"] != pattern {
		t.Error("failed to retrieve logs:", err)
	}
}

func TestPostMetrics(t *testing.T) {
	metricName := randString(5)
	metrics := MetricList{
		Name:    metricName,
		Columns: []string{"col1", "col2", "col3"},
		Metrics: []Metric{
			Metric{
				Point: []float64{1.1, 2.2, 3.3},
			},
		},
	}
	jsonStr, _ := json.Marshal(metrics)

	param := url.Values{"metrics": {string(jsonStr)}}
	retBody := `{"successful":1}`
	server, zeus, bucket_name := mock("/metrics/goZeus/"+metricName+"/", &param, 200, retBody)
	defer server.Close()

	token := zeus.Token
	zeus.Token = ""
	_, err := zeus.bucket(bucket_name).PostMetrics(MetricList{})
	if err == nil {
		t.Error("should fail on empty token")
	}
	zeus.Token = token

	successful, err := zeus.bucket(bucket_name).PostMetrics(MetricList{})
	if err == nil {
		t.Error("should fail on empty metrics")
	}

	successful, err = zeus.bucket(bucket_name).PostMetrics(metrics)
	if err != nil || successful != 1 {
		t.Errorf("failed to post metrics, %d successful", successful)
	}
}

func TestGetMetricNames(t *testing.T) {
	metricName := randString(5)

	param := url.Values{
		"metric_name": {metricName},
		"offset":      {"1"},
		"limit":       {"1024"}}
	retBody := `["Harry", "Potter"]`
	server, zeus, bucket_name := mock("/metrics/goZeus/_names/", &param, 200, retBody)
	defer server.Close()

	token := zeus.Token
	zeus.Token = ""
	_, err := zeus.bucket(bucket_name).GetMetricNames(metricName, 1, 1024)
	if err == nil {
		t.Error("should fail on empty token")
	}
	zeus.Token = token

	names, err := zeus.bucket(bucket_name).GetMetricNames(metricName, 1, 1024)
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
	timestamp := float64(1430355869.123)

	param := url.Values{
		"metric_name":         {metricName},
		"aggregator_function": {"max"},
		"aggregator_column":   {"age"},
		"group_interval":      {"1s"},
		"from":                {strconv.FormatFloat(timestamp-10.0, 'f', 3, 64)},
		"to":                  {strconv.FormatFloat(timestamp, 'f', 3, 64)},
		"filter_condition":    {"value>10"},
		"limit":               {"1024"}}
	retBody := `[{"points": [[1430355869.123,144740003,20.0]],"name": "Jon.Snow","columns": ["time","sequence_number","age"]}]`
	server, zeus, bucket_name := mock("/metrics/goZeus/_values/", &param, 200, retBody)
	defer server.Close()

	token := zeus.Token
	zeus.Token = ""
	_, err := zeus.bucket(bucket_name).GetMetricValues(metricName, "max", "age", "1s", timestamp-10.0, timestamp, "value>10", 0, 1024)
	if err == nil {
		t.Error("should fail on empty token")
	}
	zeus.Token = token

	metrics, err := zeus.bucket(bucket_name).GetMetricValues(metricName, "max", "age", "1s", timestamp-10.0, timestamp, "value>10", 0, 1024)
	if err != nil {
		t.Error("failed to get metric values:", err)
	}
	expMetric := MetricList{
		Name:    metricName,
		Columns: []string{"sequence_number", "age"},
		Metrics: []Metric{
			Metric{Timestamp: 1430355869.123, Point: []float64{144740003, 20}},
		},
	}
	// Two colume: sequence_number and value
	if expMetric.Name != metrics.Name ||
		len(expMetric.Columns) != len(metrics.Columns) ||
		expMetric.Columns[0] != metrics.Columns[0] ||
		expMetric.Columns[1] != metrics.Columns[1] ||
		expMetric.Metrics[0].Timestamp != metrics.Metrics[0].Timestamp ||
		expMetric.Metrics[0].Point[0] != metrics.Metrics[0].Point[0] ||
		expMetric.Metrics[0].Point[1] != metrics.Metrics[0].Point[1] {
		t.Errorf("failed to retrieve metric values, expect %#v, got %#v", expMetric, metrics)
	}
}

func TestDeleteMetrics(t *testing.T) {
	metricName := randString(5)
	param := url.Values{}
	retBody := `["Metric deletion successful"]`
	server, zeus, bucket_name := mock("/metrics/goZeus/"+metricName+"/", &param, 200, retBody)
	defer server.Close()

	token := zeus.Token
	zeus.Token = ""
	_, err := zeus.bucket(bucket_name).DeleteMetrics("")
	if err == nil {
		t.Error("should fail on empty token")
	}
	zeus.Token = token

	successful, err := zeus.bucket(bucket_name).DeleteMetrics("")
	if successful != false || err == nil {
		t.Error("should fail on empty metricName")
	}
	successful, err = zeus.bucket(bucket_name).DeleteMetrics(metricName)
	if err != nil || successful != true {
		t.Error("failed to delete on series")
	}
}

func TestZeusGetTrigalert(t *testing.T) {
	param := make(url.Values)
	server, zeus, bucket_name := mock("/trigalerts/goZeus/", &param, 200, `{"successful": 1}`)
	defer server.Close()

	token := zeus.Token
	zeus.Token = ""
	_, err := zeus.bucket(bucket_name).GetTrigalert()
	if err == nil {
		t.Error("should fail on empty token")
	}
	zeus.Token = token

	trigalert, err := zeus.bucket(bucket_name).GetTrigalert()
	if err != nil {
		t.Error("failed to retrieve trigalert:", err)
	}
	id, _ := trigalert["successful"].(float64)
	if id != float64(1) {
		t.Error("Retrieved trigalert is wrong:", trigalert)
	}
}

func TestZeusGetTrigalertLast24(t *testing.T) {
	param := make(url.Values)
	server, zeus, bucket_name := mock("/trigalerts/goZeus/last24/", &param, 200, `{"successful": 1}`)
	defer server.Close()

	token := zeus.Token
	zeus.Token = ""
	_, err := zeus.bucket(bucket_name).GetTrigalertLast24()
	if err == nil {
		t.Error("should fail on empty token")
	}
	zeus.Token = token

	trigalert, err := zeus.bucket(bucket_name).GetTrigalertLast24()
	if err != nil {
		t.Error("failed to retrieve trigalert_last24:", err)
	}
	id, _ := trigalert["successful"].(float64)
	if id != float64(1) {
		t.Error("Retrieved trigalert_last24 is wrong:", trigalert)
	}
}

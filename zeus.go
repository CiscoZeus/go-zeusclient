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

// Package zeusclient provides operations to post and retrieve logs/metrics.
package zeus

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// Log contains two properties of a log: timestamp of a event, and description
// of the event.
type Log struct {
	Timestamp int64  `json:"timestamp,omitempty"`
	Message   string `json:"message"`
}

// A collection of logs.
type Logs []Log

// Metric contains two properties of a metric: timestamp of a metric, and
// value of the metric.
type Metric struct {
	Value     float64 `json:"value"`
	Timestamp int64   `json:"timestamp,omitempty"`
}

// A collection of metrics.
type Metrics []Metric

// Zeus implements functions to send/receive log, send/receive metrics.
// Constructing Zeus requires URL of Zeus rest api and user token.
type Zeus struct {
	apiServ, token string
}

type postResponse struct {
	Successful int    `json:"successful"`
	Failed     int    `json:"failed"`
	Error      string `json:"error"`
}

func buildUrl(urls ...string) string {
	return strings.Join(urls, "/") + "/"
}

func (zeus *Zeus) request(method, urlStr string, data *url.Values) (
	body []byte, status int, err error) {
	if data == nil {
		data = &url.Values{}
	}
	var resp *http.Response
	if method == "POST" {
		resp, err = http.PostForm(urlStr, *data)
	} else {
		resp, err = http.Get(urlStr + "?" + data.Encode())
	}
	if err != nil {
		return []byte{}, 0, err
	}

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, 0, err
	}
	status = resp.StatusCode
	resp.Body.Close()
	return
}

// GetLogs returns a list of logs that math given constrains. logName is the
// name(or category) of the log. Pattern is a expression to match log
// message field. From and to are the starting and ending timestamp of logs
// in unix time(in seconds). Offest and limit control the pagination of the
// results.
func (zeus *Zeus) GetLogs(logName, pattern string, from, to int64, offset,
	limit int) (total int, logs Logs, err error) {
	urlStr := buildUrl(zeus.apiServ, "logs", zeus.token)
	data := make(url.Values)
	if len(logName) > 0 {
		data.Add("log_name", logName)
	}
	if len(pattern) > 0 {
		data.Add("pattern", pattern)
	}
	if from > 0 {
		data.Add("from", strconv.FormatInt(from, 10))
	}
	if to > 0 {
		data.Add("to", strconv.FormatInt(to, 10))
	}
	if offset != 0 {
		data.Add("offset", strconv.Itoa(offset))
	}
	if limit != 0 {
		data.Add("limit", strconv.Itoa(limit))
	}

	body, status, err := zeus.request("GET", urlStr, &data)
	if err != nil {
		return 0, Logs{}, err
	}

	if status == 200 {
		type Resp struct {
			Total  int  `json:"total"`
			Result Logs `json:"result"`
		}
		var resp Resp
		if err := json.Unmarshal(body, &resp); err != nil {
			return 0, Logs{}, err
		}
		total = resp.Total
		logs = resp.Result
	} else if status == 400 {
		return 0, Logs{}, errors.New("Bad request")
	}
	return
}

// PostLogs sends a list of logs under given log name. It returns number of
// successfully sent logs or an error.
func (zeus *Zeus) PostLogs(logName string, logs Logs) (successful int, err error) {
	urlStr := buildUrl(zeus.apiServ, "logs", zeus.token, logName)

	jsonStr, err := json.Marshal(logs)
	if err != nil {
		return 0, err
	}
	data := url.Values{"logs": {string(jsonStr)}}

	body, status, err := zeus.request("POST", urlStr, &data)
	if err != nil {
		return 0, err
	}

	var resp postResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return 0, err
	}
	if status == 200 {
		successful = resp.Successful
	} else {
		return 0, errors.New(resp.Error)
	}
	return
}

// PostMetrics sends a list of metrics under the given metricName.
func (zeus *Zeus) PostMetrics(metricName string, metrics Metrics) (
	successful int, err error) {
	urlStr := buildUrl(zeus.apiServ, "metrics", zeus.token, metricName)

	jsonStr, err := json.Marshal(metrics)
	if err != nil {
		return 0, err
	}
	data := url.Values{"metrics": {string(jsonStr)}}

	body, status, err := zeus.request("POST", urlStr, &data)
	if err != nil {
		return 0, err
	}
	var resp postResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return 0, err
	}
	if status == 200 {
		successful = resp.Successful
	} else {
		return 0, errors.New(resp.Error)
	}
	return
}

// GetMetricNames returns less than limit of metric names that match regular
// expression metricName.
func (zeus *Zeus) GetMetricNames(metricName string, limit int) (names []string,
	err error) {
	urlStr := buildUrl(zeus.apiServ, "metrics", zeus.token, "_names")
	data := make(url.Values)
	if len(metricName) > 0 {
		data.Add("metric_name", metricName)
	}
	if limit > 0 {
		data.Add("limit", strconv.Itoa(limit))
	}

	body, status, err := zeus.request("GET", urlStr, &data)
	if err != nil {
		return []string{}, err
	}
	if status == 200 {
		if err := json.Unmarshal(body, &names); err != nil {
			return []string{}, err
		}
	}
	return
}

// GetMetricValues returns less than limit of metric values under the name
// metricName, The returned values' timestamp greater than from and smaller
// than to. Values can be aggreated by a function(count, min, max, sum, mean,
// mode, median). Values can also be gouped by a group_interval or filtered by
// filter_condition(value > 0).
func (zeus *Zeus) GetMetricValues(metricName string, aggregator string,
	groupInterval string, from, to int64, filterCondition string, limit int) (
	multiMetrics map[string]Metrics, err error) {
	urlStr := buildUrl(zeus.apiServ, "metrics", zeus.token, "_values")
	data := make(url.Values)
	if len(metricName) > 0 {
		data.Add("metric_name", metricName)
	}
	if len(aggregator) > 0 {
		data.Add("aggregator_function", aggregator)
	}
	if len(groupInterval) > 0 {
		data.Add("group_interval", groupInterval)
	}
	if from > 0 {
		data.Add("from", strconv.FormatInt(from, 10))
	}
	if to > 0 {
		data.Add("to", strconv.FormatInt(to, 10))
	}
	if len(filterCondition) > 0 {
		data.Add("filter_condition", filterCondition)
	}
	if limit > 0 {
		data.Add("limit", strconv.Itoa(limit))
	}

	body, status, err := zeus.request("GET", urlStr, &data)
	if err != nil {
		return map[string]Metrics{}, err
	}
	if status == 200 {
		type JsonMetric struct {
			Points  [][]float64 `json:"points"`
			Name    string      `json:"name"`
			Columns []string    `json:"columns"`
		}
		var resp []JsonMetric
		if err := json.Unmarshal(body, &resp); err != nil {
			return map[string]Metrics{}, err
		}
		multiMetrics = make(map[string]Metrics)
		for _, metric := range resp {
			for _, col := range metric.Columns {
				multiMetrics[metric.Name+"_"+col] = make(Metrics, len(metric.Points))
			}
			for i, vals := range metric.Points {
				for idx, val := range vals {
					key := metric.Name + "_" + metric.Columns[idx]
					multiMetrics[key][i] = Metric{Value: val}
				}
			}
			if _, pres := multiMetrics[metric.Name+"_time"]; pres {
				if _, pres := multiMetrics[metric.Name+"_value"]; pres {
					ms := multiMetrics[metric.Name+"_value"]
					ts := multiMetrics[metric.Name+"_time"]
					for idx, _ := range ms {
						ms[idx].Timestamp = int64(ts[idx].Value)
					}
				}
				delete(multiMetrics, metric.Name+"_time")
			}
		}
	}
	return
}

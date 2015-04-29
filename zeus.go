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

type Log struct {
	Timestamp int64  `json:"timestamp,omitempty"`
	Message   string `json:"message"`
}

type Logs []Log

type Metric struct {
	Value     float64 `json:"value"`
	Timestamp int64   `json:"timestamp,omitempty"`
}

type Metrics []Metric

type Zeus struct {
	apiServ, token string
}

type postResponse struct {
	Successful int    `json:"successful"`
	Failed     int    `json:"failed"`
	Error      string `json:"error"`
}

func (zeus *Zeus) request(method, urlStr string, data *url.Values) (
	body []byte, status int, err error) {
	if data == nil {
		data = &url.Values{}
	}
	var req *http.Request
	if method == "post" {
		req, err = http.NewRequest(method, urlStr, strings.NewReader(data.Encode()))
	} else {
		req, err = http.NewRequest(method, urlStr+"?"+data.Encode(), nil)
	}
	if err != nil {
		return []byte{}, 0, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := new(http.Client)
	resp, err := client.Do(req)
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

func (zeus *Zeus) GetLogs(logType, pattern string, from, to int64, offset,
	limit int) (total int, logs Logs, err error) {
	urlStr := zeus.apiServ + "logs/" + zeus.token + "/"
	data := make(url.Values)
	if len(logType) > 0 {
		data.Add("log_type", logType)
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

	body, status, err := zeus.request("get", urlStr, &data)
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

func (zeus *Zeus) PostLogs(logType string, logs Logs) (successful int, err error) {
	urlStr := zeus.apiServ + "logs/" + zeus.token + "/" + logType + "/"

	jsonStr, err := json.Marshal(logs)
	if err != nil {
		return 0, err
	}
	data := url.Values{"logs": {string(jsonStr)}}

	body, status, err := zeus.request("post", urlStr, &data)
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

func (zeus *Zeus) PostMetrics(metricName string, metrics Metrics) (
	successful int, err error) {
	urlStr := zeus.apiServ + "metrics/" + zeus.token + "/" + metricName + "/"

	jsonStr, err := json.Marshal(metrics)
	if err != nil {
		return 0, err
	}
	data := url.Values{"metrics": {string(jsonStr)}}

	body, status, err := zeus.request("post", urlStr, &data)
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

func (zeus *Zeus) GetMetricNames(pattern string, limit int) (names []string,
	err error) {
	urlStr := zeus.apiServ + "metrics/" + zeus.token + "/_names/"
	data := make(url.Values)
	if len(pattern) > 0 {
		data.Add("pattern", pattern)
	}
	if limit > 0 {
		data.Add("limit", strconv.Itoa(limit))
	}

	body, status, err := zeus.request("get", urlStr, &data)
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

func (zeus *Zeus) GetMetricValues(pattern string, aggregator string,
	groupInterval string, from, to int64, filterCondition string, limit int) (
	multiMetrics map[string]Metrics, err error) {
	urlStr := zeus.apiServ + "metrics/" + zeus.token + "/_values/"
	data := make(url.Values)
	if len(pattern) > 0 {
		data.Add("pattern", pattern)
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

	body, status, err := zeus.request("get", urlStr, &data)
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
					for idx, m := range ms {
						m.Timestamp = int64(ts[idx].Value)
					}
				}
				delete(multiMetrics, metric.Name+"_time")
			}
		}
	}
	return
}

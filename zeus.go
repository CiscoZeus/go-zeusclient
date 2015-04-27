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

type Token struct {
	Access_token  string `access_token`
	Token_type    string `token_type`
	Refresh_token string `refresh_token`
	Expires_in    int    `expires_in`
}

type Log struct {
	Timestamp string `json:"timestamp"`
	Message   string `json:"message"`
}

type Logs []Log

type Metric struct {
	Name      string  `json:"name,omitempty"`
	Value     float64 `json:"value"`
	Timestamp int64   `json:"timestamp,omitempty"`
}

type Metrics []Metric

type Zeus struct {
	apiServ, username, password string
	token                       *Token
}

type postResponse struct {
	Successful int    `json:"successful"`
	Failed     int    `json:"failed"`
	Error      string `json:"error"`
}

func (zeus *Zeus) request(method, urlStr string, data *url.Values) (
	body []byte, status int, err error) {
	var req *http.Request
	if method == "post" {
		req, _ = http.NewRequest(method, urlStr, strings.NewReader(data.Encode()))
	} else {
		req, _ = http.NewRequest(method, urlStr+"?"+data.Encode(), nil)
	}

	if method == "post" {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	req.Header.Set("Authorization", "Bearer "+zeus.token.Access_token)

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

func (zeus *Zeus) Auth() (bool, error) {
	u := zeus.apiServ + "login"
	resp, err := http.PostForm(u,
		url.Values{"username": {zeus.username}, "password": {zeus.password}})
	if err != nil {
		return false, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	dec := json.NewDecoder(strings.NewReader(string(body)))
	var token Token
	if err := dec.Decode(&token); err != nil {
		return false, err
	}

	if resp.StatusCode != 200 {
		var resp map[string]string
		if err := json.Unmarshal(body, &resp); err != nil {
			return false, err
		}
		if _, repre := resp["error"]; repre != false {
			return false, errors.New(resp["error_description"])
		}
		return false, errors.New("Authentication error")
	}

	zeus.token = &token
	return true, nil
}

func (zeus *Zeus) GetLogs(pattern, from, to string, offset, pagesize int) (
	total int, logs Logs, err error) {
	urlStr := zeus.apiServ + "logs/"
	data := make(url.Values)
	if len(pattern) > 0 {
		data.Add("pattern", pattern)
	}
	if len(from) > 0 {
		data.Add("from", from)
	}
	if len(to) > 0 {
		data.Add("to", to)
	}
	if offset != 0 {
		data.Add("offset", strconv.Itoa(offset))
	}
	if pagesize != 0 {
		data.Add("pagesize", strconv.Itoa(pagesize))
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
	}
	return
}

func (zeus *Zeus) PostLogs(tag string, logs Logs) (successful int, err error) {
	urlStr := zeus.apiServ + "logs/" + tag + "/"

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

func (zeus *Zeus) PostMetrics(tag string, metrics Metrics) (successful int, err error) {
	urlStr := zeus.apiServ + "metrics/" + tag + "/"

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

func (zeus *Zeus) GetMetricNames(pattern string, limit int) (names []string, err error) {
	urlStr := zeus.apiServ + "metric_names/"
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
		//FIXME(kai) delete following 2 lines after api returns JSON
		strBody := strings.Trim(string(body), "\"")
		strBody = strings.Replace(strBody, "\\", "", -1)
		if err := json.Unmarshal([]byte(strBody), &names); err != nil {
			return []string{}, err
		}
	}
	return
}

func (zeus *Zeus) GetMetricValues(pattern string, aggregator string,
	groupInterval string, from string, to string, filterCondition string,
	limit int) (multiMetrics map[string]Metrics, err error) {
	urlStr := zeus.apiServ + "metric_values/"
	data := make(url.Values)
	if len(pattern) > 0 {
		data.Add("pattern", pattern)
	}
	if len(aggregator) > 0 {
		data.Add("aggregator", aggregator)
	}
	if len(groupInterval) > 0 {
		data.Add("group_interval", groupInterval)
	}
	if len(from) > 0 {
		data.Add("from", from)
	}
	if len(to) > 0 {
		data.Add("to", to)
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
		//FIXME(kai) delete following 2 lines after api returns JSON
		strBody := strings.Trim(string(body), "\"")
		strBody = strings.Replace(strBody, "\\", "", -1)
		var resp []JsonMetric
		if err := json.Unmarshal([]byte(strBody), &resp); err != nil {
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
				if _, pres := multiMetrics[metric.Name+"_Value"]; pres {
					ms := multiMetrics[metric.Name+"_Value"]
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

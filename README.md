# Zeus Go Client
[![Godoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/CiscoZeus/go-zeusclient) [![license](https://img.shields.io/hexpm/l/plug.svg)]
Go client for CiscoZeus.io. it allows a user to send and receive data to and from Zeus.

* Free software: [Apache License Version 2.0](http://www.apache.org/licenses/LICENSE-2.0)
* Documentation: https://godoc.org/github.com/CiscoZeus/go-zeusclient

## Features
* Send logs and metrics to Zeus
* Query both logs and metrics.

## Sample
* Generate a Zeus client object:
`zeus := &Zeus{apiServ: REST_API_URL, token: "goZeus"}`

* Send a log
```
logs := make([]Log, 1)
logs[0] = Log{Timestamp: time.Now().Unix(), Message: "Message from Go"}
successful, err := zeus.PostLogs("Hello_log", logs)
```

* Retrieve logs
```
timestamp := 
total, logs, err := zeus.GetLogs("apache", "GET", 1431711563, 1431711863, 0, 10)
```

* Send a metric
```
metrics := Metrics{{Value: 123}}
successful, err := zeus.PostMetrics("Hello_metric", metrics)
```

* Query metric name
```
names, err := zeus.GetMetricNames("Hello_metric", 1024)
```

* Query metric values
```
timestamp := int64(1430355869000)
multiMetrics, err := zeus.GetMetricValues("Hello_metric", "", "", timestamp-int64(10*1000), timestamp, "value>10", 1024)
```

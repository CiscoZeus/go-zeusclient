# Zeus Go Client

![Alt text](https://raw.githubusercontent.com/CiscoZeus/go-zeusclient/master/docs/images/zeus_logo.png "Zeus Logo")

[![Build Status](https://travis-ci.org/CiscoZeus/go-zeusclient.svg)](https://travis-ci.org/CiscoZeus/go-zeusclient) [![Coverage Status](https://coveralls.io/repos/CiscoZeus/go-zeusclient/badge.svg)](https://coveralls.io/r/CiscoZeus/go-zeusclient) [![Godoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/CiscoZeus/go-zeusclient) [![license](https://img.shields.io/hexpm/l/plug.svg)](http://www.apache.org/licenses/LICENSE-2.0)

Go client for CiscoZeus.io. it allows a user to send and receive data to and from Zeus.

## Features
* Send logs and metrics to Zeus
* Query both logs and metrics.

## Get Code
```
go get github.com/CiscoZeus/go-zeusclient
```

## Usage
At first, generate a Zeus client object:
```go
zeus := &Zeus{ApiServ: "http://api.ciscozeus.io", Token: "{Your token}"}
```
Then, you can call methods as
```go
zeus.bucket("{Your organization's name}/{Your bucket's name}").SomeMethod()
// "bucket" method sets your bucket's information. It's necessary
```

## Examples
After initialize 'zeus' as [Usage](##Usage), 
* Send a log
```go
logs := LogList{
    Name: "syslog",
    Logs: []Log{
        Log{"foo": "bar", "tar": "woo"},
    },
}
suc, err := zeus.bucket("org1/bucket1").PostLogs(logs)
```

* Retrieve logs
```go
total, logs, err := zeus.bucket("org1/bucket1").GetLogs("syslog", "", "", 0, 0, 0, 0)
```

* Send a metric
```go
metrics := MetricList{
    Name:    "sample",
    Columns: []string{"col1", "col2", "col3"},
    Metrics: []Metric{
        Metric{
            Timestamp: float64(time.Now().Unix()),
            Point:     []float64{1.0, 2.0, 3.0},
        },
    },
}
suc, err := zeus.bucket("org1/bucket1").PostMetrics(metrics)
```

* Query metric name
```go
name, err := zeus.bucket("org1/bucket1").GetMetricNames("sample*", 0, 0)
```

* Query metric values
```go
timestamp := 1430355869.123
rMetrics, err := zeus.bucket("org1/bucket1").GetMetricValues("sample", "", "", "", timestamp-10.0, timestamp, "col2>1", 0, 1024)
```

For more examples, please refer to sample/sample.go

## Contributing

1. Fork it ( https://github.com/CiscoZeus/go-zeusclient/fork )
2. Create your feature branch (`git checkout -b my-new-feature`)
3. Commit your changes (`git commit -am 'Add some feature'`)
4. Push to the branch (`git push origin my-new-feature`)
5. Create a new Pull Request

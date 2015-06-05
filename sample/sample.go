package main

import (
	"fmt"
	. "github.com/CiscoZeus/go-zeusclient"
	"math/rand"
	"time"
)

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
	fmt.Println("Going to send two logs:")
	fmt.Printf("%+v\n", logs)
	suc, err := zeus.PostLogs(logs)
	if err != nil {
		panic("failed to send logs")
	}
	fmt.Printf("Sent 2 logs, %d successful\n", suc)

	time.Sleep(1)
	total, rLogs, err := zeus.GetLogs(logs.Name, "", "", 0, 0, 0, 0)
	if err != nil {
		panic("failed to get logs")
	}
	fmt.Printf("Received %d logs:\n", total)
	fmt.Printf("%+v\n", rLogs)
}

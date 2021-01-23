package main

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	reter "github.com/skvoch/reter/scheduler"
	"time"
)

func main() {
	scheduler, err := reter.New(reter.Zerolog(log.Logger), &reter.Options{
		Etcd: reter.EtcdOptions{
			Endpoints:   []string{"127.0.0.1:2379"},
			DialTimeout: time.Second * 5,
		},
		LockTtl: time.Minute * 1,
	})
	if err != nil {
		log.Fatal()
	}
	if err := scheduler.Every(2).Seconds().Do("get_data", func() {
		fmt.Println("doing work")
		time.Sleep(time.Second * 5)
	}); err != nil {

	}
	scheduler.Run(context.Background())

	for {

	}
}

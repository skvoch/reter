package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/skvoch/reter/scheduler"
	"golang.org/x/sync/errgroup"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	ErrSigint = errors.New("sigint or sigterm")
)

func NotifySigterm() error {
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM)
	<-s
	return ErrSigint
}

func main() {
	s, err := scheduler.New(scheduler.Zerolog(log.Logger), &scheduler.Options{
		Etcd: scheduler.EtcdOptions{
			Endpoints: []string{"127.0.0.1:2379"},
		},
		LockTTL: time.Minute * 1,
		Timeout: time.Second * 10,
	})

	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to etcd")
	}
	g, ctx := errgroup.WithContext(context.Background())

	g.Go(NotifySigterm)

	g.Go(func() error {
		return s.Every(30).Seconds().Do(ctx, "get_data", func() {
			fmt.Println("doing work")
		})
	})

	if err := g.Wait(); err != nil {
		if errors.Is(err, ErrSigint) {
			log.Info().Msg("graceful shutdown")
		} else {
			log.Error().Err(err).Msg("err group waiting")
		}
	}
}

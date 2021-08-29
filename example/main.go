package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/skvoch/reter/scheduler/logger/zerologadapter"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/skvoch/reter/scheduler"
	"golang.org/x/sync/errgroup"
)

var (
	ErrSigint = errors.New("sigint or sigterm")
)

func NotifySigterm(ctx context.Context) error {
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-ctx.Done():
			return nil

		case <-s:
			return ErrSigint
		}
	}
}

func main() {
	s, err := scheduler.New(zerologadapter.NewLogger(log.Logger), &scheduler.Options{
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

	g.Go(func() error {
		return NotifySigterm(ctx)
	})

	g.Go(func() error {
		return s.Every(10).Seconds().Do(ctx, "seconds", func() {
			fmt.Println("print every 10 second")
		})
	})

	g.Go(func() error {
		return s.Every().Interval(time.Second*3).Do(ctx, "interval", func() {
			fmt.Println("print every 3 second")
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

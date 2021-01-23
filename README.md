# Reter - task scheduler with ectd locks

### Example
```go
s, err := scheduler.New(scheduler.Zerolog(log.Logger), &scheduler.Options{
	Etcd: scheduler.EtcdOptions{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: time.Second * 5,
	},
	LockTtl: time.Minute * 1,
})
	
if err != nil {
	log.Fatal().Err(err).Msg("failed to connect to etcd")
}

g, ctx := errgroup.WithContext(context.Background())
g.Go(func() error {
	return s.Every(2).Seconds().Do(ctx, "task_name", func() {
		fmt.Println("doing work")
		time.Sleep(time.Second * 5)
	})
})
```

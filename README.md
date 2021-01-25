# Reter - task scheduler with ectd locks

### Algorithm
1. Checking the time since last action
2. If difference between time.Now() and last action time less than interval - skipping
3. Locking and calling a handler func
4. Updating last action time, and unlocking


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

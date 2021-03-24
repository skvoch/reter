# Reter - task scheduler with ectd locks
![Tests](https://github.com/skvoch/reter/workflows/tests/badge.svg)

### Installation
```bash
go get github.com/skvoch/reter
```

### Algorithm
1. Checking the time since the last action
2. If the difference between time.Now() and last action time less than interval - skipping
3. Locking and calling a handler function
4. Updating last action time, and unlocking

### Example
```go
s, err := scheduler.New(logger.Zerolog(log.Logger), &scheduler.Options{
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
	return s.Every(30).Seconds().Do(ctx, "print 1", func() {
		fmt.Println("print 1")
	})
})

g.Go(func() error {
	return s.Every().Interval(time.Second).Do(ctx, "print 2", func() {
		fmt.Println("print 2")
	})
})
```


### Warning
If you have some issue with building your application, please put these lines to *go.mod* file
```go
replace (
	github.com/coreos/bbolt => go.etcd.io/bbolt v1.3.3
	go.etcd.io/etcd => go.etcd.io/etcd v0.0.0-20200520232829-54ba9589114f
	google.golang.org/grpc v1.32.0 => google.golang.org/grpc v1.26.0
)
```

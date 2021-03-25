module github.com/skvoch/reter

go 1.15

require (
	bou.ke/monkey v1.0.2
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f // indirect
	github.com/go-kit/kit v0.10.0
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/mock v1.5.0
	github.com/golang/protobuf v1.4.2 // indirect
	github.com/google/uuid v1.2.0 // indirect
	github.com/jackc/pgx/v4 v4.11.0 // indirect
	github.com/rs/zerolog v1.20.0
	github.com/sirupsen/logrus v1.6.0
	github.com/skvoch/go-etcd-lock/v5 v5.0.12
	github.com/stretchr/testify v1.7.0
	go.etcd.io/etcd v3.3.25+incompatible // indirect
	go.etcd.io/etcd/v3 v3.3.0-rc.0.0.20200518175753-732df43cf85b
	go.uber.org/zap v1.16.0
	golang.org/x/sync v0.0.0-20201020160332-67f06af15bc9
	gopkg.in/errgo.v1 v1.0.1 // indirect
)

replace go.etcd.io/etcd => go.etcd.io/etcd v0.0.0-20200520232829-54ba9589114f

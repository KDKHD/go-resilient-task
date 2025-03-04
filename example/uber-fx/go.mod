module github.com/KDKHD/alarm

go 1.23.5

replace github.com/KDKHD/go-resilient-task/modules/go-resilient-task => ../../modules/go-resilient-task

require (
	github.com/KDKHD/go-resilient-task/modules/go-resilient-task v0.0.0-20250303214559-469595d66da9
	github.com/jackc/pgx/v5 v5.7.2
	github.com/twmb/franz-go v1.18.1
	go.uber.org/fx v1.23.0
	go.uber.org/zap v1.27.0
)

require (
	github.com/google/uuid v1.6.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/klauspost/compress v1.17.11 // indirect
	github.com/pierrec/lz4/v4 v4.1.22 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/twmb/franz-go/pkg/kmsg v1.9.0 // indirect
	go.uber.org/dig v1.18.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/crypto v0.32.0 // indirect
	golang.org/x/sync v0.10.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
	golang.org/x/text v0.21.0 // indirect
)

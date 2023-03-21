package worker

import (
	"time"
)

type WorkerType int

var (
	Asynq WorkerType = 0
)

type WorkerRedisCredentials struct {
	Addresses    []string
	Username     string
	Password     string
	Database     int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type EnqueuerOption func(*EnqueuerOptions)

type EnqueuerOptions struct {
	MaxRetry int
	Timeout  time.Duration
}

func NewEnqueuerOptions(opts ...EnqueuerOption) EnqueuerOptions {
	opt := EnqueuerOptions{
		MaxRetry: 3,
		Timeout:  0 * time.Second,
	}

	for _, o := range opts {
		o(&opt)
	}

	return opt
}

func WithMaxRetry(val int) EnqueuerOption {
	return func(eo *EnqueuerOptions) {
		if val > 0 {
			eo.MaxRetry = val
		}
	}
}

func WithTimeout(val time.Duration) EnqueuerOption {
	return func(eo *EnqueuerOptions) {
		eo.Timeout = val
	}
}

package worker

import (
	"context"

	"github.com/ONLYOFFICE/onlyoffice-pipedrive/pkg/config"
)

type BackgroundEnqueuer interface {
	Enqueue(pattern string, task []byte, opts ...EnqueuerOption) error
	EnqueueContext(ctx context.Context, pattern string, task []byte, opts ...EnqueuerOption) error
	Close() error
}

func NewBackgroundEnqueuer(config *config.WorkerConfig) BackgroundEnqueuer {
	switch config.Worker.Type {
	case 0:
		return newAsynqEnqueuer(config)
	default:
		return newAsynqEnqueuer(config)
	}
}

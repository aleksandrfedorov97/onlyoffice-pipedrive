package worker

import (
	"context"

	"github.com/ONLYOFFICE/onlyoffice-pipedrive/pkg/config"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/pkg/log"
)

type BackgroundWorker interface {
	Register(pattern string, handler func(ctx context.Context, payload []byte) error)
	Run()
}

func NewBackgroundWorker(config *config.WorkerConfig, logger log.Logger) BackgroundWorker {
	switch config.Worker.Type {
	case 0:
		return newAsynqWorker(config, logger)
	default:
		return newAsynqWorker(config, logger)
	}
}

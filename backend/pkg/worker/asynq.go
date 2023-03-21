package worker

import (
	"context"
	"log"
	"time"

	"github.com/ONLYOFFICE/onlyoffice-pipedrive/pkg/config"
	plog "github.com/ONLYOFFICE/onlyoffice-pipedrive/pkg/log"
	"github.com/hibiken/asynq"
)

type asynqWorker struct {
	enabled bool
	srv     *asynq.Server
	mux     *asynq.ServeMux
}

type asynqEnqueuer struct {
	enabled bool
	client  *asynq.Client
}

func newAsynqWorker(config *config.WorkerConfig, logger plog.Logger) BackgroundWorker {
	var workerOpts asynq.RedisConnOpt = asynq.RedisClientOpt{
		Addr:         config.Worker.RedisAddresses[0],
		Username:     config.Worker.RedisUsername,
		Password:     config.Worker.RedisPassword,
		ReadTimeout:  4 * time.Second,
		WriteTimeout: 7 * time.Second,
	}
	if len(config.Worker.RedisAddresses) > 1 {
		workerOpts = asynq.RedisClusterClientOpt{
			Addrs:        config.Worker.RedisAddresses,
			Username:     config.Worker.RedisUsername,
			Password:     config.Worker.RedisPassword,
			ReadTimeout:  4 * time.Second,
			WriteTimeout: 7 * time.Second,
		}
	}

	return asynqWorker{
		enabled: config.Worker.Enable,
		srv: asynq.NewServer(workerOpts, asynq.Config{
			Concurrency: config.Worker.MaxConcurrency,
			Logger:      logger,
		}),
		mux: asynq.NewServeMux(),
	}
}

func (w asynqWorker) Register(pattern string, handler func(ctx context.Context, payload []byte) error) {
	if w.enabled {
		w.mux.Handle(pattern, asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
			return handler(ctx, t.Payload())
		}))
	}
}

func (w asynqWorker) Run() {
	if w.enabled {
		go func() {
			if err := w.srv.Run(w.mux); err != nil {
				log.Fatal(err.Error())
			}
		}()
	}
}

func newAsynqEnqueuer(config *config.WorkerConfig) BackgroundEnqueuer {
	var enqOpts asynq.RedisConnOpt = asynq.RedisClientOpt{
		Addr:         config.Worker.RedisAddresses[0],
		Username:     config.Worker.RedisUsername,
		Password:     config.Worker.RedisPassword,
		ReadTimeout:  4 * time.Second,
		WriteTimeout: 7 * time.Second,
	}
	if len(config.Worker.RedisAddresses) > 1 {
		enqOpts = asynq.RedisClusterClientOpt{
			Addrs:        config.Worker.RedisAddresses,
			Username:     config.Worker.RedisUsername,
			Password:     config.Worker.RedisPassword,
			ReadTimeout:  4 * time.Second,
			WriteTimeout: 7 * time.Second,
		}
	}

	return asynqEnqueuer{
		enabled: config.Worker.Enable,
		client:  asynq.NewClient(enqOpts),
	}
}

func (e asynqEnqueuer) Enqueue(pattern string, task []byte, opts ...EnqueuerOption) error {
	if e.enabled {
		options := NewEnqueuerOptions(opts...)
		t := asynq.NewTask(pattern, task)

		_, err := e.client.Enqueue(t, asynq.MaxRetry(options.MaxRetry), asynq.Timeout(options.Timeout))
		return err
	}

	return nil
}

func (e asynqEnqueuer) EnqueueContext(ctx context.Context, pattern string, task []byte, opts ...EnqueuerOption) error {
	if e.enabled {
		options := NewEnqueuerOptions(opts...)
		t := asynq.NewTask(pattern, task)

		_, err := e.client.EnqueueContext(ctx, t, asynq.MaxRetry(options.MaxRetry), asynq.Timeout(options.Timeout))
		return err
	}

	return nil
}

func (e asynqEnqueuer) Close() error {
	if e.enabled {
		return e.Close()
	}

	return nil
}

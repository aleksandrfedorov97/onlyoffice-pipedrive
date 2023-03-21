package web

import (
	"net/http"

	"github.com/ONLYOFFICE/onlyoffice-pipedrive/pkg/config"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/pkg/log"
	chttp "github.com/ONLYOFFICE/onlyoffice-pipedrive/pkg/service/http"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/pkg/worker"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/callback/web/controller"
	workerh "github.com/ONLYOFFICE/onlyoffice-pipedrive/services/callback/web/worker"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/shared"
	"github.com/gin-gonic/gin"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"go-micro.dev/v4/cache"
	"go-micro.dev/v4/client"
)

type CallbackService struct {
	namespace     string
	mux           *chi.Mux
	client        client.Client
	logger        log.Logger
	worker        worker.BackgroundWorker
	enqueuer      worker.BackgroundEnqueuer
	maxSize       int64
	uploadTimeout int
}

// ApplyMiddleware useed to apply http server middlewares.
func (s CallbackService) ApplyMiddleware(middlewares ...func(http.Handler) http.Handler) {
	s.mux.Use(middlewares...)
}

// NewService initializes http server with options.
func NewServer(
	serverConfig *config.ServerConfig,
	workerConfig *config.WorkerConfig,
	onlyofficeConfig *shared.OnlyofficeConfig,
	logger log.Logger,
) chttp.ServerEngine {
	gin.SetMode(gin.ReleaseMode)

	service := CallbackService{
		namespace:     serverConfig.Namespace,
		mux:           chi.NewRouter(),
		logger:        logger,
		worker:        worker.NewBackgroundWorker(workerConfig, logger),
		enqueuer:      worker.NewBackgroundEnqueuer(workerConfig),
		maxSize:       onlyofficeConfig.Onlyoffice.Callback.MaxSize,
		uploadTimeout: onlyofficeConfig.Onlyoffice.Callback.UploadTimeout,
	}

	return service
}

// NewHandler returns http server engine.
func (s CallbackService) NewHandler(client client.Client, cache cache.Cache) interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
} {
	return s.InitializeServer(client)
}

// InitializeServer sets all injected dependencies.
func (s *CallbackService) InitializeServer(c client.Client) *chi.Mux {
	s.client = c
	s.worker.Register("pipedrive-callback-upload", workerh.NewCallbackWorker(s.namespace, c, s.uploadTimeout, s.logger).UploadFile)
	s.InitializeRoutes()
	s.worker.Run()
	return s.mux
}

// InitializeRoutes builds all http routes.
func (s *CallbackService) InitializeRoutes() {
	callbackController := controller.NewCallbackController(s.namespace, s.maxSize, s.logger, s.client)
	s.mux.Group(func(r chi.Router) {
		r.Use(chimiddleware.Recoverer)
		r.NotFound(func(rw http.ResponseWriter, r *http.Request) {
			http.Redirect(rw, r.WithContext(r.Context()), "https://onlyoffice.com", http.StatusMovedPermanently)
		})
		r.Post("/callback", callbackController.BuildPostHandleCallback(s.enqueuer))
	})
}

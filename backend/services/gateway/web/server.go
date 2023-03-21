package web

import (
	"net/http"

	"github.com/ONLYOFFICE/onlyoffice-pipedrive/pkg/config"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/pkg/log"
	shttp "github.com/ONLYOFFICE/onlyoffice-pipedrive/pkg/service/http"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/gateway/web/controller"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/gateway/web/middleware"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/shared"
	pclient "github.com/ONLYOFFICE/onlyoffice-pipedrive/services/shared/client"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/shared/crypto"
	"github.com/gin-gonic/gin"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/sessions"
	"go-micro.dev/v4/cache"
	"go-micro.dev/v4/client"
)

type PipedriveHTTPService struct {
	namespace        string
	mux              *chi.Mux
	client           client.Client
	logger           log.Logger
	store            sessions.Store
	clientID         string
	clientSecret     string
	redirectURI      string
	allowedDownloads int
}

// NewService initializes http server with options.
func NewServer(
	serverConfig *config.ServerConfig,
	credentialsConfig *config.OAuthCredentialsConfig,
	onlyofficeConfig *shared.OnlyofficeConfig,
	logger log.Logger,
) shttp.ServerEngine {
	gin.SetMode(gin.ReleaseMode)

	service := PipedriveHTTPService{
		namespace:        serverConfig.Namespace,
		mux:              chi.NewRouter(),
		logger:           logger,
		clientID:         credentialsConfig.Credentials.ClientID,
		clientSecret:     credentialsConfig.Credentials.ClientSecret,
		redirectURI:      credentialsConfig.Credentials.RedirectURI,
		store:            sessions.NewCookieStore([]byte(credentialsConfig.Credentials.ClientSecret)),
		allowedDownloads: onlyofficeConfig.Onlyoffice.Builder.AllowedDownloads,
	}

	return service
}

// ApplyMiddleware useed to apply http server middlewares.
func (s PipedriveHTTPService) ApplyMiddleware(middlewares ...func(http.Handler) http.Handler) {
	s.mux.Use(middlewares...)
}

// NewHandler returns http server engine.
func (s PipedriveHTTPService) NewHandler(client client.Client, cache cache.Cache) interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
} {
	return s.InitializeServer(client)
}

// InitializeServer sets all injected dependencies.
func (s *PipedriveHTTPService) InitializeServer(c client.Client) *chi.Mux {
	s.client = c
	s.InitializeRoutes()
	return s.mux
}

// InitializeRoutes builds all http routes.
func (s *PipedriveHTTPService) InitializeRoutes() {
	jwtManager := crypto.NewOnlyofficeJwtManager()
	authMiddleware := middleware.BuildHandleAuthMiddleware(s.clientID, s.clientSecret, s.logger)
	tokenMiddleware := middleware.BuildHandleContextMiddleware(s.clientSecret, jwtManager, s.logger)

	authController := controller.NewAuthController(s.namespace, s.redirectURI, s.client, pclient.NewPipedriveAuthClient(s.clientID, s.clientSecret), s.logger)
	apiController := controller.NewApiController(s.namespace, s.client, jwtManager, s.allowedDownloads, s.logger)

	s.mux.Group(func(r chi.Router) {
		r.Use(chimiddleware.Recoverer)
		r.NotFound(func(rw http.ResponseWriter, cr *http.Request) {
			http.Redirect(rw, cr.WithContext(cr.Context()), "https://onlyoffice.com", http.StatusMovedPermanently)
		})

		r.Route("/oauth", func(cr chi.Router) {
			cr.Use(chimiddleware.NoCache)
			cr.Get("/auth", authController.BuildGetAuth())
			cr.Delete("/auth", authMiddleware(authController.BuildDeleteAuth()))
		})

		r.Route("/api", func(cr chi.Router) {
			cr.Use(func(h http.Handler) http.Handler {
				return tokenMiddleware(h)
			})
			cr.Get("/me", apiController.BuildGetMe())
			cr.Get("/config", apiController.BuildGetConfig())
			cr.Post("/settings", apiController.BuildPostSettings())
			cr.Get("/settings", apiController.BuildGetSettings())
		})

		r.Get("/download", apiController.BuildGetFile())
	})
}

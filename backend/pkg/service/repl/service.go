package repl

import (
	"fmt"
	"net/http"
	"net/http/pprof"
	"strconv"
	"time"

	"github.com/ONLYOFFICE/onlyoffice-pipedrive/pkg/config"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/pkg/middleware"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/hellofresh/health-go/v5"
	"github.com/justinas/alice"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// NewService Initializes repl service with options.
func NewService(
	replConfig *config.ServerConfig,
	corsConfig *config.CORSConfig,
) *http.Server {
	mux := http.NewServeMux()
	h, _ := health.New(health.WithComponent(health.Component{
		Name:    fmt.Sprintf("%s:%s", replConfig.Namespace, replConfig.Name),
		Version: fmt.Sprintf("v%d", replConfig.Version),
	}))

	mux.Handle("/metrics", promhttp.Handler())
	mux.Handle("/health", h.Handler())

	if replConfig.Debug {
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}

	return &http.Server{
		Addr: replConfig.ReplAddress,
		Handler: alice.New(
			chimiddleware.RealIP,
			middleware.NewRateLimiter(1000, 1*time.Second, middleware.WithKeyFuncAll),
			chimiddleware.RequestID,
			middleware.Cors(corsConfig.CORS.AllowedOrigins, corsConfig.CORS.AllowedMethods, corsConfig.CORS.AllowedHeaders, corsConfig.CORS.AllowCredentials),
			middleware.Secure,
			middleware.NoCache,
			middleware.Version(strconv.Itoa(replConfig.Version)),
		).Then(mux),
	}
}

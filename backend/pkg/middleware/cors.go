package middleware

import (
	"net/http"

	corsmiddleware "github.com/go-chi/cors"
)

// Cors creates a new CORS middleware.
func Cors(AllowedOrigins, AllowedMethods, AllowedHeaders []string, AllowCredentials bool) func(http.Handler) http.Handler {
	return corsmiddleware.Handler(corsmiddleware.Options{
		AllowedOrigins:   AllowedOrigins,
		AllowedMethods:   AllowedMethods,
		AllowedHeaders:   AllowedHeaders,
		AllowCredentials: AllowCredentials,
	})
}

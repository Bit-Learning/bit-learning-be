package app

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/lcaohoanq/bit-learning-be-v2/internal/modules/auth"
	"github.com/lcaohoanq/bit-learning-be-v2/internal/modules/user"
	"github.com/lcaohoanq/bit-learning-be-v2/internal/platform/httpx"
	platformmw "github.com/lcaohoanq/bit-learning-be-v2/internal/platform/middleware"
)

func NewRouter(authHandler *auth.Handler, userHandler *user.Handler, tokens platformmw.TokenParser, logger *slog.Logger) http.Handler {
	router := chi.NewRouter()
	router.Use(chimiddleware.RequestID, chimiddleware.RealIP, chimiddleware.Recoverer)
	router.Use(platformmw.RequestLogger(logger))
	router.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		httpx.WriteData(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	router.Route("/v1", func(router chi.Router) {
		router.Post("/auth/register", authHandler.Register)
		router.Post("/auth/login", authHandler.Login)
		router.Group(func(router chi.Router) {
			router.Use(platformmw.Authenticate(tokens))
			router.Get("/users/me", userHandler.GetMe)
			router.Patch("/users/me", userHandler.UpdateMe)
		})
	})
	return otelhttp.NewHandler(router, "http.server")
}

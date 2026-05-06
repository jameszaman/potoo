package server

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/notifylayer/notifylayer/internal/api/handlers"
	"github.com/notifylayer/notifylayer/internal/api/middleware"
	"github.com/notifylayer/notifylayer/internal/auth"
	api "github.com/notifylayer/notifylayer/internal/gen/openapi"
	"github.com/notifylayer/notifylayer/internal/queue"
)

func New(pool *pgxpool.Pool, q *queue.Client, allowedOrigin string) http.Handler {
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{allowedOrigin},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	r.Use(chimiddleware.RealIP)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/docs", http.RedirectHandler("/docs/", http.StatusMovedPermanently).ServeHTTP)
	r.Get("/docs/", swaggerUI)
	r.Get("/docs/openapi.json", serveSpec)

	h := handlers.New(pool, q)

	// Public routes — no API key required.
	r.Post("/v1/auth/register", h.RegisterHTTP)
	r.Post("/v1/auth/login", h.LoginHTTP)
	r.Post("/v1/auth/logout", h.LogoutHTTP)
	r.Post("/v1/auth/refresh", h.RefreshSessionHTTP)
	r.Get("/v1/auth/me", h.GetMeHTTP)
	r.Get("/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	r.Post("/v1/webhooks/email/{provider}", h.IngestEmailWebhookHTTP)

	// Authenticated API routes on a separate router, mounted at /.
	// This avoids chi route conflicts with the public auth routes above.
	apiRouter := chi.NewRouter()
	apiRouter.Use(auth.Authenticate(pool))
	api.HandlerWithOptions(api.NewStrictHandler(h, nil), api.ChiServerOptions{
		BaseRouter: apiRouter,
	})
	r.Mount("/", apiRouter)

	return r
}

func swaggerUI(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, `<!DOCTYPE html>
<html>
<head>
  <title>NotifyLayer API</title>
  <meta charset="utf-8"/>
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist/swagger-ui.css">
</head>
<body>
<div id="swagger-ui"></div>
<script src="https://unpkg.com/swagger-ui-dist/swagger-ui-bundle.js"></script>
<script>
  SwaggerUIBundle({
    url: "/docs/openapi.json",
    dom_id: "#swagger-ui",
    presets: [SwaggerUIBundle.presets.apis, SwaggerUIBundle.SwaggerUIStandalonePreset],
    layout: "BaseLayout",
    deepLinking: true,
  })
</script>
</body>
</html>`)
}

func serveSpec(w http.ResponseWriter, _ *http.Request) {
	spec, err := api.GetSpecJSON()
	if err != nil {
		http.Error(w, "could not load spec", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(spec) //nolint:errcheck
}

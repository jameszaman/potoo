package server

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/notifylayer/notifylayer/internal/api/handlers"
	"github.com/notifylayer/notifylayer/internal/api/middleware"
	"github.com/notifylayer/notifylayer/internal/auth"
	api "github.com/notifylayer/notifylayer/internal/gen/openapi"
	"github.com/notifylayer/notifylayer/internal/queue"
)

func New(pool *pgxpool.Pool, q *queue.Client) http.Handler {
	r := chi.NewRouter()

	r.Use(chimiddleware.RealIP)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/docs", http.RedirectHandler("/docs/", http.StatusMovedPermanently).ServeHTTP)
	r.Get("/docs/", swaggerUI)
	r.Get("/docs/openapi.json", serveSpec)

	h := handlers.New(pool, q)

	// Public webhook routes — no auth, raw body needed for signature verification.
	r.Post("/v1/webhooks/email/{provider}", h.IngestEmailWebhookHTTP)

	// All API routes are registered once. Auth middleware skips /v1/health.
	r.Group(func(r chi.Router) {
		r.Use(auth.AuthenticateExcept("/v1/health")(pool))
		api.HandlerWithOptions(api.NewStrictHandler(h, nil), api.ChiServerOptions{
			BaseRouter: r,
		})
	})

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

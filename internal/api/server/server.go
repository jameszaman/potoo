package server

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/potoo/potoo/internal/api/handlers"
	"github.com/potoo/potoo/internal/api/middleware"
	"github.com/potoo/potoo/internal/auth"
	api "github.com/potoo/potoo/internal/gen/openapi"
	"github.com/potoo/potoo/internal/queue"
	"github.com/potoo/potoo/internal/storage"
)

func New(pool *pgxpool.Pool, q *queue.Client, store storage.Driver, allowedOrigin string) http.Handler {
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{allowedOrigin},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
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

	// Serve locally uploaded files. In S3 mode this route is unused
	// (files are served directly from S3 URLs).
	if os.Getenv("STORAGE_DRIVER") != "s3" {
		uploadDir := envOr("STORAGE_LOCAL_PATH", "./uploads")
		absDir, _ := filepath.Abs(uploadDir)
		r.Get("/uploads/{file}", func(w http.ResponseWriter, r *http.Request) {
			name := chi.URLParam(r, "file")
			// Prevent path traversal
			if filepath.Base(name) != name {
				http.NotFound(w, r)
				return
			}
			http.ServeFile(w, r, filepath.Join(absDir, name))
		})
	}

	h := handlers.New(pool, q, store)

	// ── Tier 1: Fully public routes ────────────────────────────────────────
	r.Get("/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	r.Get("/v1/platform/status", h.GetPlatformStatusHTTP)
	r.Post("/v1/setup", h.SetupHTTP)
	r.Post("/v1/auth/register", h.RegisterHTTP)
	r.Post("/v1/auth/accept-invite", h.AcceptInviteHTTP)
	r.Get("/v1/invite/{token}", h.GetInviteHTTP)
	r.Post("/v1/auth/login", h.LoginHTTP)
	r.Post("/v1/auth/logout", h.LogoutHTTP)
	r.Post("/v1/auth/refresh", h.RefreshSessionHTTP)
	r.Get("/v1/auth/me", h.GetMeHTTP)
	r.Post("/v1/auth/select-org", h.SelectOrgHTTP)
	r.Post("/v1/webhooks/email/{provider}", h.IngestEmailWebhookHTTP)

	// ── Tier 2: Session-authenticated routes (cookie, no API key) ──────────
	r.Group(func(r chi.Router) {
		r.Use(auth.AuthenticateSession(pool))
		r.Get("/v1/me/orgs", h.ListMyOrgsHTTP)
		r.Get("/v1/api-keys", h.ListAPIKeysHTTP)
		r.Post("/v1/api-keys", h.CreateAPIKeyHTTP)
		r.Delete("/v1/api-keys/{keyId}", h.RevokeAPIKeyHTTP)
	})

	// ── Tier 3: Session + platform owner ────────────────────────────────────
	r.Group(func(r chi.Router) {
		r.Use(auth.AuthenticateSession(pool))
		r.Use(auth.RequirePlatformOwner(pool))
		r.Get("/v1/platform/orgs", h.ListOrgsHTTP)
		r.Post("/v1/platform/orgs", h.CreateOrgHTTP)
		r.Patch("/v1/platform/orgs/{orgId}", h.UpdateOrgHTTP)
		r.Post("/v1/platform/invites", h.CreateInviteHTTP)
		r.Get("/v1/platform/invites", h.ListInvitesHTTP)
	})

	// ── Tier 4: API key or session cookie — all remaining strict routes ──────
	apiRouter := chi.NewRouter()
	apiRouter.Use(auth.AuthenticateEither(pool))

	// Upload is not in the OpenAPI spec so it must be registered on apiRouter
	// directly before HandlerWithOptions registers its routes, so chi's trie
	// matches this explicit path before the spec routes are evaluated.
	apiRouter.With(auth.AuthenticateSession(pool)).Post("/v1/uploads/image", h.UploadImageHTTP)

	api.HandlerWithOptions(api.NewStrictHandler(h, nil), api.ChiServerOptions{
		BaseRouter: apiRouter,
	})
	r.Mount("/", apiRouter)

	return r
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func swaggerUI(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, `<!DOCTYPE html>
<html>
<head>
  <title>Potoo API</title>
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

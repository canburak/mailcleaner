package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

// NewRouter creates a new chi router with all routes configured
func NewRouter(h *Handler) *chi.Mux {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)

	// CORS for frontend
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://localhost:3000", "http://127.0.0.1:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Requested-With"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// API routes
	r.Route("/api", func(r chi.Router) {
		// Health check
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
		})

		// Account routes
		r.Route("/accounts", func(r chi.Router) {
			r.Get("/", h.ListAccounts)
			r.Post("/", h.CreateAccount)
			r.Post("/test", h.TestAccountDirect)

			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", h.GetAccount)
				r.Put("/", h.UpdateAccount)
				r.Delete("/", h.DeleteAccount)
				r.Post("/test", h.TestAccount)
				r.Get("/folders", h.GetAccountFolders)
				r.Post("/folders", h.CreateFolder)

				// Rules for this account
				r.Route("/rules", func(r chi.Router) {
					r.Get("/", h.ListRules)
					r.Post("/", h.CreateRule)
				})

				// Preview and apply
				r.Get("/preview", h.PreviewRules)
				r.Post("/apply", h.ApplyRules)
			})
		})

		// Rule routes (for direct access)
		r.Route("/rules", func(r chi.Router) {
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", h.GetRule)
				r.Put("/", h.UpdateRule)
				r.Delete("/", h.DeleteRule)
			})
		})
	})

	return r
}

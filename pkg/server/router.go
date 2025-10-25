// Package server provides HTTP server implementation
package server

import (
	"encoding/json"
	"net/http"

	"github.com/thanhphuchuynh/dear-future/pkg/auth"
	"github.com/thanhphuchuynh/dear-future/pkg/composition"
	"github.com/thanhphuchuynh/dear-future/pkg/handlers"
	"github.com/thanhphuchuynh/dear-future/pkg/middleware"
)

// NewRouter creates a new HTTP router with all handlers and middleware configured
func NewRouter(app *composition.App) *http.ServeMux {
	mux := http.NewServeMux()

	// Initialize JWT service
	jwtService := auth.NewJWTService(
		app.Config().JWTSecret,
		app.Config().JWTExpirationTime,
		app.Config().RefreshTokenLifetime,
	)

	// Initialize handlers
	userHandler := handlers.NewUserHandler(app, jwtService)
	messageHandler := handlers.NewMessageHandler(app)
	attachmentHandler := handlers.NewAttachmentHandler(app)
	analyticsHandler := handlers.NewAnalyticsHandler(app)

	// Create middleware chain
	authMiddleware := middleware.AuthMiddleware(jwtService)
	corsMiddleware := middleware.CORSMiddleware(app.Config().CORSOrigins)
	loggingMiddleware := middleware.LoggingMiddleware()
	recoveryMiddleware := middleware.RecoveryMiddleware()
	securityMiddleware := middleware.SecurityHeadersMiddleware()

	// Apply global middleware
	globalMiddleware := chain(
		recoveryMiddleware,
		loggingMiddleware,
		corsMiddleware,
		securityMiddleware,
	)

	// Public routes (no authentication required)
	mux.Handle("/health", globalMiddleware(http.HandlerFunc(healthHandler(app))))
	mux.Handle("/environment/current", globalMiddleware(http.HandlerFunc(environmentHandler(app))))

	// Auth routes (public)
	mux.Handle("/api/v1/auth/register", globalMiddleware(http.HandlerFunc(userHandler.Register)))
	mux.Handle("/api/v1/auth/login", globalMiddleware(http.HandlerFunc(userHandler.Login)))
	mux.Handle("/api/v1/auth/refresh", globalMiddleware(http.HandlerFunc(userHandler.RefreshToken)))

	// User routes (authenticated)
	mux.Handle("/api/v1/user/profile", chain(globalMiddleware, authMiddleware)(http.HandlerFunc(userHandler.GetProfile)))
	mux.Handle("/api/v1/user/update", chain(globalMiddleware, authMiddleware)(http.HandlerFunc(userHandler.UpdateProfile)))

	// Message routes (authenticated)
	mux.Handle("/api/v1/messages", chain(globalMiddleware, authMiddleware)(http.HandlerFunc(handleMessagesRoute(messageHandler))))
	mux.Handle("/api/v1/messages/create", chain(globalMiddleware, authMiddleware)(http.HandlerFunc(messageHandler.CreateMessage)))
	mux.Handle("/api/v1/messages/attachments", chain(globalMiddleware, authMiddleware)(http.HandlerFunc(handleAttachmentsRoute(attachmentHandler))))
	mux.Handle("/api/v1/analytics/summary", chain(globalMiddleware, authMiddleware)(http.HandlerFunc(analyticsHandler.GetSummary)))

	// API info route
	mux.Handle("/api/v1/", globalMiddleware(http.HandlerFunc(apiInfoHandler(app))))

	// Home page (development only)
	if app.Config().IsDevelopment() {
		mux.Handle("/", globalMiddleware(http.HandlerFunc(homeHandler())))
	}

	return mux
}

// handleMessagesRoute routes message requests based on method and query params
func handleMessagesRoute(h *handlers.MessageHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			// Check if ID is provided
			if r.URL.Query().Get("id") != "" {
				h.GetMessage(w, r)
			} else {
				h.GetMessages(w, r)
			}
		case http.MethodPost:
			h.CreateMessage(w, r)
		case http.MethodPut:
			h.UpdateMessage(w, r)
		case http.MethodDelete:
			h.DeleteMessage(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func handleAttachmentsRoute(h *handlers.AttachmentHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			h.Upload(w, r)
		case http.MethodGet:
			h.List(w, r)
		case http.MethodDelete:
			h.Delete(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// apiInfoHandler returns API information
func apiInfoHandler(app *composition.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"name":    "Dear Future API",
			"version": "1.0.0",
			"status":  "operational",
			"endpoints": map[string]interface{}{
				"health": map[string]string{
					"path":   "/health",
					"method": "GET",
				},
				"auth": map[string]interface{}{
					"register": map[string]string{
						"path":   "/api/v1/auth/register",
						"method": "POST",
					},
					"login": map[string]string{
						"path":   "/api/v1/auth/login",
						"method": "POST",
					},
					"refresh": map[string]string{
						"path":   "/api/v1/auth/refresh",
						"method": "POST",
					},
				},
				"user": map[string]interface{}{
					"profile": map[string]string{
						"path":   "/api/v1/user/profile",
						"method": "GET",
					},
					"update": map[string]string{
						"path":   "/api/v1/user/update",
						"method": "PUT",
					},
				},
				"messages": map[string]interface{}{
					"list": map[string]string{
						"path":   "/api/v1/messages",
						"method": "GET",
					},
					"create": map[string]string{
						"path":   "/api/v1/messages",
						"method": "POST",
					},
					"get": map[string]string{
						"path":   "/api/v1/messages?id={id}",
						"method": "GET",
					},
					"update": map[string]string{
						"path":   "/api/v1/messages?id={id}",
						"method": "PUT",
					},
					"delete": map[string]string{
						"path":   "/api/v1/messages?id={id}",
						"method": "DELETE",
					},
				},
			},
		}

		respondWithJSON(w, http.StatusOK, response)
	}
}

// chain creates a middleware chain
func chain(middlewares ...func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(final http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			final = middlewares[i](final)
		}
		return final
	}
}

// Helper functions

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

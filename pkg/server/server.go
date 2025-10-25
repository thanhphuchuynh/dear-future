// Package server provides HTTP server implementation
package server

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/thanhphuchuynh/dear-future/pkg/composition"
	"github.com/thanhphuchuynh/dear-future/pkg/config"
	"github.com/thanhphuchuynh/dear-future/pkg/domain/common"
)

// NewServer creates a new HTTP server with all routes configured
func NewServer(cfg *config.Config, app *composition.App) common.Result[*http.Server] {
	// Create router with all handlers and middleware
	router := NewRouter(app)

	// Create server with timeouts
	server := &http.Server{
		Addr:         cfg.GetServerAddress(),
		Handler:      router,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	return common.Ok(server)
}

// healthHandler returns the health status of the application
func healthHandler(app *composition.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		healthResult := app.Health(ctx)
		if healthResult.IsErr() {
			http.Error(w, "Health check failed", http.StatusInternalServerError)
			return
		}

		health := healthResult.Value()
		w.Header().Set("Content-Type", "application/json")

		// Set appropriate status code
		if health.Status == "healthy" {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
		}

		json.NewEncoder(w).Encode(health)
	}
}

// apiHandler handles all API routes
func apiHandler(app *composition.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Add CORS headers for development
		if app.Config().IsDevelopment() {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		}

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		// For now, return a simple API info response
		response := map[string]interface{}{
			"message": "Dear Future API",
			"version": "1.0.0",
			"status":  "operational",
			"endpoints": map[string]string{
				"health": "/health",
				"api":    "/api/v1/",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

// homeHandler serves a simple homepage for development
func homeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		html := `
<!DOCTYPE html>
<html>
<head>
    <title>Dear Future - Your Message to Tomorrow</title>
    <style>
        body { 
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif; 
            max-width: 800px; 
            margin: 0 auto; 
            padding: 40px 20px; 
            line-height: 1.6;
        }
        .header { text-align: center; margin-bottom: 40px; }
        .status { color: #28a745; font-weight: bold; }
        .section { margin: 20px 0; padding: 20px; background: #f8f9fa; border-radius: 8px; }
        .endpoint { font-family: monospace; background: #e9ecef; padding: 4px 8px; border-radius: 4px; }
    </style>
</head>
<body>
    <div class="header">
        <h1>🚀 Dear Future</h1>
        <p>Your Message to Tomorrow</p>
        <p class="status">✅ Server Running</p>
    </div>

    <div class="section">
        <h2>🎯 What is Dear Future?</h2>
        <p>A web application for sending messages to your future self. Built with functional programming principles in Go.</p>
    </div>

    <div class="section">
        <h2>🛠️ Architecture Highlights</h2>
        <ul>
            <li><strong>Functional Programming</strong>: Pure business logic with Result/Option monads</li>
            <li><strong>Immutable Data</strong>: All domain entities are immutable by design</li>
            <li><strong>Clean Architecture</strong>: Clear separation between business logic and side effects</li>
            <li><strong>Migration-Ready</strong>: Designed to scale from Lambda to Kubernetes</li>
        </ul>
    </div>

    <div class="section">
        <h2>🔗 API Endpoints</h2>
        <p><strong>Health Check:</strong> <a href="/health" class="endpoint">/health</a></p>
        <p><strong>Environment Info:</strong> <a href="/environment/current" class="endpoint">/environment/current</a></p>
        <p><strong>API Info:</strong> <a href="/api/v1/" class="endpoint">/api/v1/</a></p>
    </div>

    <div class="section">
        <h2>📊 Development Status</h2>
        <ul>
            <li>✅ Functional Programming Foundation</li>
            <li>✅ Domain Models (User, Message)</li>
            <li>✅ Business Logic Functions</li>
            <li>✅ Configuration Management</li>
            <li>✅ HTTP Server</li>
            <li>🔄 Database Integration (Next)</li>
            <li>🔄 Frontend Development (Next)</li>
        </ul>
    </div>

    <div class="section">
        <h2>🧪 Testing</h2>
        <p>Run tests with: <code class="endpoint">go test ./...</code></p>
        <p>All tests are passing with comprehensive coverage of functional programming patterns.</p>
    </div>
</body>
</html>`

		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(html))
	}
}

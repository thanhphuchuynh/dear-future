// Package server provides HTTP handlers for the Dear Future API
package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/thanhphuchuynh/dear-future/pkg/composition"
	"github.com/thanhphuchuynh/dear-future/pkg/config"
)

// EnvironmentResponse represents the response for /environment/current endpoint
type EnvironmentResponse struct {
	Environment string          `json:"environment"`
	Version     string          `json:"version"`
	Platform    string          `json:"platform"`
	Debug       bool            `json:"debug"`
	Features    map[string]bool `json:"features"`
	Uptime      string          `json:"uptime"`
	Timestamp   time.Time       `json:"timestamp"`
	BuildInfo   BuildInfo       `json:"build_info"`
	Config      ConfigSummary   `json:"config"`
}

// BuildInfo represents build and version information
type BuildInfo struct {
	Version   string `json:"version"`
	GoVersion string `json:"go_version"`
	GitCommit string `json:"git_commit,omitempty"`
	BuildTime string `json:"build_time,omitempty"`
	BuildUser string `json:"build_user,omitempty"`
}

// ConfigSummary represents a summary of current configuration
type ConfigSummary struct {
	DatabaseConnected bool   `json:"database_connected"`
	CacheEnabled      bool   `json:"cache_enabled"`
	HTTPSEnabled      bool   `json:"https_enabled"`
	MetricsEnabled    bool   `json:"metrics_enabled"`
	ServerPort        string `json:"server_port"`
}

// Application start time for uptime calculation
var startTime = time.Now()

// environmentHandler handles GET /environment/current requests
func environmentHandler(app *composition.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Get configuration
		cfg := app.Config()

		// Build response
		response := EnvironmentResponse{
			Environment: cfg.Environment,
			Version:     "1.0.0",
			Platform:    cfg.PlatformName,
			Debug:       cfg.Debug,
			Features:    buildFeaturesMap(cfg),
			Uptime:      formatUptime(time.Since(startTime)),
			Timestamp:   time.Now(),
			BuildInfo:   buildBuildInfo(),
			Config:      buildConfigSummary(cfg),
		}

		// Set response headers
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("X-Environment", cfg.Environment)
		w.Header().Set("X-Version", response.Version)

		// Return response
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}

// buildFeaturesMap creates a map of enabled features
func buildFeaturesMap(cfg *config.Config) map[string]bool {
	if cfg == nil {
		return map[string]bool{}
	}

	return map[string]bool{
		"push_notifications":  cfg.Features.EnablePushNotifications,
		"webhooks":            cfg.Features.EnableWebhooks,
		"message_templates":   cfg.Features.EnableMessageTemplates,
		"batch_processing":    cfg.Features.EnableBatchProcessing,
		"advanced_scheduling": cfg.Features.EnableAdvancedScheduling,
		"file_attachments":    cfg.Features.EnableFileAttachments,
		"email_reminders":     cfg.Features.EnableEmailReminders,
		"analytics":           cfg.Features.EnableAnalytics,
	}
}

// buildBuildInfo creates build information
func buildBuildInfo() BuildInfo {
	return BuildInfo{
		Version:   "1.0.0",
		GoVersion: "1.21+",
		GitCommit: "latest", // This would be set during build
		BuildTime: "2025-06-05T21:00:00Z",
		BuildUser: "dear-future-builder",
	}
}

// buildConfigSummary creates a configuration summary
func buildConfigSummary(cfg *config.Config) ConfigSummary {
	if cfg == nil {
		return ConfigSummary{}
	}

	return ConfigSummary{
		DatabaseConnected: cfg.DatabaseURL != "" || cfg.SupabaseURL != "",
		CacheEnabled:      cfg.CacheEnabled,
		HTTPSEnabled:      cfg.EnableHTTPS,
		MetricsEnabled:    cfg.MetricsEnabled,
		ServerPort:        cfg.Port,
	}
}

// formatUptime formats uptime duration in a human-readable format
func formatUptime(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm %ds", days, hours, minutes, seconds)
	} else if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	} else {
		return fmt.Sprintf("%ds", seconds)
	}
}

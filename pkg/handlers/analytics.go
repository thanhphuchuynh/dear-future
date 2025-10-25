package handlers

import (
	"net/http"
	"sort"
	"time"

	"github.com/thanhphuchuynh/dear-future/pkg/composition"
	"github.com/thanhphuchuynh/dear-future/pkg/domain/message"
	"github.com/thanhphuchuynh/dear-future/pkg/middleware"
)

// AnalyticsHandler provides aggregated insights for dashboards.
type AnalyticsHandler struct {
	app *composition.App
}

// AnalyticsSummaryResponse represents the analytics payload.
type AnalyticsSummaryResponse struct {
	Totals         AnalyticsTotals   `json:"totals"`
	DeliveryRate   float64           `json:"delivery_rate"`
	Upcoming       *UpcomingSummary  `json:"upcoming,omitempty"`
	RecentMessages []MessageOverview `json:"recent_messages"`
	AttachmentCount int              `json:"attachment_count"`
}

// AnalyticsTotals groups message counts by state.
type AnalyticsTotals struct {
	Total     int `json:"total"`
	Scheduled int `json:"scheduled"`
	Delivered int `json:"delivered"`
	Failed    int `json:"failed"`
	Cancelled int `json:"cancelled"`
}

// UpcomingSummary describes the next scheduled delivery.
type UpcomingSummary struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	DeliveryDate string `json:"delivery_date"`
	DeliveryMethod string `json:"delivery_method"`
}

// MessageOverview is a condensed representation of a message.
type MessageOverview struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	Status       string `json:"status"`
	DeliveryDate string `json:"delivery_date"`
	DeliveryMethod string `json:"delivery_method"`
}

// NewAnalyticsHandler constructs the handler.
func NewAnalyticsHandler(app *composition.App) *AnalyticsHandler {
	return &AnalyticsHandler{app: app}
}

// GetSummary handles GET /api/v1/analytics/summary
func (h *AnalyticsHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	messagesResult := h.app.Database().FindMessagesByUserID(r.Context(), userID, 100, 0)
	if messagesResult.IsErr() {
		respondWithError(w, http.StatusInternalServerError, "failed to load analytics")
		return
	}

	messages := messagesResult.Value()
	stats := message.CalculateDeliveryStats(messages)

	attachmentTotal := 0
	for _, msg := range messages {
		attachmentsResult := h.app.Database().FindAttachmentsByMessageID(r.Context(), msg.ID())
		if attachmentsResult.IsOk() {
			attachmentTotal += len(attachmentsResult.Value())
		}
	}

	recent := buildRecentOverview(messages)
	upcoming := buildUpcomingSummary(messages)

	response := AnalyticsSummaryResponse{
		Totals: AnalyticsTotals{
			Total:     stats.Total,
			Scheduled: stats.Scheduled,
			Delivered: stats.Delivered,
			Failed:    stats.Failed,
			Cancelled: stats.Cancelled,
		},
		DeliveryRate:   stats.DeliveryRate,
		Upcoming:       upcoming,
		RecentMessages: recent,
		AttachmentCount: attachmentTotal,
	}

	respondWithJSON(w, http.StatusOK, response)
}

func buildRecentOverview(messages []message.Message) []MessageOverview {
	sorted := make([]message.Message, len(messages))
	copy(sorted, messages)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].UpdatedAt().After(sorted[j].UpdatedAt())
	})

	limit := 5
	if len(sorted) < limit {
		limit = len(sorted)
	}

	result := make([]MessageOverview, 0, limit)
	for i := 0; i < limit; i++ {
		msg := sorted[i]
		result = append(result, MessageOverview{
			ID:            msg.ID().String(),
			Title:         msg.Title(),
			Status:        string(msg.Status()),
			DeliveryDate:  msg.DeliveryDate().Format(time.RFC3339),
			DeliveryMethod: string(msg.DeliveryMethod()),
		})
	}

	return result
}

func buildUpcomingSummary(messages []message.Message) *UpcomingSummary {
	var upcoming *message.Message
	for _, msg := range messages {
		if msg.Status() != message.StatusScheduled {
			continue
		}
		if msg.DeliveryDate().Before(time.Now()) {
			continue
		}
		if upcoming == nil || msg.DeliveryDate().Before(upcoming.DeliveryDate()) {
			temp := msg
			upcoming = &temp
		}
	}

	if upcoming == nil {
		return nil
	}

	return &UpcomingSummary{
		ID:            upcoming.ID().String(),
		Title:         upcoming.Title(),
		DeliveryDate:  upcoming.DeliveryDate().Format(time.RFC3339),
		DeliveryMethod: string(upcoming.DeliveryMethod()),
	}
}

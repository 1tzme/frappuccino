package handler

import (
	"net/http"

	"hot-coffee/internal/service"
	"hot-coffee/pkg/logger"
)

type AggregationHandler struct {
	aggregationService service.AggregationServiceInterface
	logger             *logger.Logger
}

func NewAggregationHandler(s service.AggregationServiceInterface, log *logger.Logger) *AggregationHandler {
	return &AggregationHandler{
		aggregationService: s,
		logger:             log.WithComponent("aggregation_handler"),
	}
}

// GetTotalSales handles GET /api/v1/reports/total-sales
func (h *AggregationHandler) GetTotalSales(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Handling get total sales report")

	report, err := h.aggregationService.GetTotalSales()
	if err != nil {
		h.logger.Error("Failed to get total sales report", "error", err)
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to generate sales report")
		return
	}

	writeJSONResponse(w, http.StatusOK, report)
}

// GetPopularItems handles GET /api/v1/reports/popular-items
func (h *AggregationHandler) GetPopularItems(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Handling get popular items request")

	report, err := h.aggregationService.GetPopularItems()
	if err != nil {
		h.logger.Error("Failed to get popular items report", "error", err)
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to get popular items")
		return
	}

	writeJSONResponse(w, http.StatusOK, report)
}

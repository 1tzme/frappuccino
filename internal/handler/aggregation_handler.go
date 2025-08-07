package handler

// TODO: Transition State: JSON → PostgreSQL
// DEPRECATED: Update aggregation handling for PostgreSQL-based data operations
// 1. Replace in-memory aggregation logic with SQL aggregate queries
// 2. Implement database view-based reports for better performance
// 3. Add proper database error handling for aggregation operations
// 4. Update response format to handle database query results efficiently
// 5. Implement database-specific optimization for reporting queries

import (
	"net/http"
	"time"

	"frappuccino/internal/service"
	"frappuccino/pkg/logger"
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
	reqCtx := &logger.RequestContext{
		Method:     r.Method,
		Path:       r.URL.Path,
		RemoteAddr: r.RemoteAddr,
		StartTime:  time.Now(),
	}
	h.logger.LogRequest(reqCtx)

	report, err := h.aggregationService.GetTotalSales()
	if err != nil {
		h.logger.Error("Failed to get total sales report", "error", err)
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to generate sales report")
		reqCtx.StatusCode = http.StatusInternalServerError
		h.logger.LogResponse(reqCtx)
		return
	}

	writeJSONResponse(w, http.StatusOK, report)
	reqCtx.StatusCode = http.StatusOK
	h.logger.LogResponse(reqCtx)
}

// GetPopularItems handles GET /api/v1/reports/popular-items
func (h *AggregationHandler) GetPopularItems(w http.ResponseWriter, r *http.Request) {
	reqCtx := &logger.RequestContext{
		Method:     r.Method,
		Path:       r.URL.Path,
		RemoteAddr: r.RemoteAddr,
		StartTime:  time.Now(),
	}
	h.logger.LogRequest(reqCtx)

	report, err := h.aggregationService.GetPopularItems()
	if err != nil {
		h.logger.Error("Failed to get popular items report", "error", err)
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to get popular items")
		reqCtx.StatusCode = http.StatusInternalServerError
		h.logger.LogResponse(reqCtx)
		return
	}

	writeJSONResponse(w, http.StatusOK, report)
	reqCtx.StatusCode = http.StatusOK
	h.logger.LogResponse(reqCtx)
}

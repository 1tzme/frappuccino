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
	"strconv"
	"strings"
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

// SearchFullText handles GET /reports/search
func (h *AggregationHandler) SearchFullText(w http.ResponseWriter, r *http.Request) {
	reqCtx := &logger.RequestContext{
		Method:     r.Method,
		Path:       r.URL.Path,
		RemoteAddr: r.RemoteAddr,
		StartTime:  time.Now(),
	}
	h.logger.LogRequest(reqCtx)

	query := r.URL.Query().Get("q")
	if query == "" {
		h.logger.Warn("Search query parameter 'q' is required")
		writeErrorResponse(w, http.StatusBadRequest, "Query parameter 'q' is required")
		reqCtx.StatusCode = http.StatusBadRequest
		h.logger.LogResponse(reqCtx)
		return
	}

	filters := []string{}
	if filterParam := r.URL.Query().Get("filter"); filterParam != "" {
		filters = strings.Split(filterParam, ",")
		for i, filter := range filters {
			filters[i] = strings.TrimSpace(filter)
		}
	}

	var minPrice, maxPrice *float64
	if minPriceStr := r.URL.Query().Get("minPrice"); minPriceStr != "" {
		if price, err := strconv.ParseFloat(minPriceStr, 64); err == nil {
			minPrice = &price
		} else {
			h.logger.Warn("Invalid minPrice parameter", "value", minPriceStr, "error", err)
			writeErrorResponse(w, http.StatusBadRequest, "Invalid minPrice parameter")
			reqCtx.StatusCode = http.StatusBadRequest
			h.logger.LogResponse(reqCtx)
			return
		}
	}

	if maxPriceStr := r.URL.Query().Get("maxPrice"); maxPriceStr != "" {
		if price, err := strconv.ParseFloat(maxPriceStr, 64); err == nil {
			maxPrice = &price
		} else {
			h.logger.Warn("Invalid maxPrice parameter", "value", maxPriceStr, "error", err)
			writeErrorResponse(w, http.StatusBadRequest, "Invalid maxPrice parameter")
			reqCtx.StatusCode = http.StatusBadRequest
			h.logger.LogResponse(reqCtx)
			return
		}
	}

	searchReq := service.SearchRequest{
		Query:    query,
		Filters:  filters,
		MinPrice: minPrice,
		MaxPrice: maxPrice,
	}

	result, err := h.aggregationService.SearchFullText(searchReq)
	if err != nil {
		h.logger.Error("Failed to perform full text search", "error", err)
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to perform search")
		reqCtx.StatusCode = http.StatusInternalServerError
		h.logger.LogResponse(reqCtx)
		return
	}

	writeJSONResponse(w, http.StatusOK, result)
	reqCtx.StatusCode = http.StatusOK
	h.logger.LogResponse(reqCtx)
}

// GetOrderedItemsByPeriod handles GET /reports/orderedItemsByPeriod
func (h *AggregationHandler) GetOrderedItemsByPeriod(w http.ResponseWriter, r *http.Request) {
	reqCtx := &logger.RequestContext{
		Method:     r.Method,
		Path:       r.URL.Path,
		RemoteAddr: r.RemoteAddr,
		StartTime:  time.Now(),
	}
	h.logger.LogRequest(reqCtx)

	period := r.URL.Query().Get("period")
	if period == "" {
		h.logger.Warn("Period parameter is required")
		writeErrorResponse(w, http.StatusBadRequest, "Period parameter is required")
		reqCtx.StatusCode = http.StatusBadRequest
		h.logger.LogResponse(reqCtx)
		return
	}

	month := r.URL.Query().Get("month")
	year := r.URL.Query().Get("year")

	periodReq := service.OrderedItemsByPeriodRequest{
		Period: period,
		Month:  month,
		Year:   year,
	}

	result, err := h.aggregationService.GetOrderedItemsByPeriod(periodReq)
	if err != nil {
		h.logger.Error("Failed to get ordered items by period", "error", err)
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to get ordered items by period")
		reqCtx.StatusCode = http.StatusInternalServerError
		h.logger.LogResponse(reqCtx)
		return
	}

	writeJSONResponse(w, http.StatusOK, result)
	reqCtx.StatusCode = http.StatusOK
	h.logger.LogResponse(reqCtx)
}

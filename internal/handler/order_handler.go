package handler

// TODO: Transition State: JSON → PostgreSQL
// DEPRECATED: Update error handling for PostgreSQL integration
// 1. Handle database constraint violations (foreign key, unique constraints)
// 2. Implement proper transaction error handling for order operations
// 3. Add database connection timeout and retry logic
// 4. Replace file I/O error responses with database-specific error messages
// 5. Update HTTP status codes for database operation failures

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"frappuccino/internal/service"
	"frappuccino/models"
	"frappuccino/pkg/logger"
)

// OrderHandler struct
type OrderHandler struct {
	orderService service.OrderServiceInterface
	logger       *logger.Logger
}

// NewOrderHandler creates a new OrderHandler with the given service and logger
func NewOrderHandler(orderService service.OrderServiceInterface, logger *logger.Logger) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
		logger:       logger.WithComponent("order_handler"),
	}
}

// CreateOrder handles POST /api/v1/orders
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	reqCtx := &logger.RequestContext{
		Method:     r.Method,
		Path:       r.URL.Path,
		RemoteAddr: r.RemoteAddr,
		StartTime:  time.Now(),
	}
	h.logger.LogRequest(reqCtx)

	var createReq service.CreateOrderRequest
	if err := h.parseRequestBody(r, &createReq); err != nil {
		h.logger.Warn("Invalid request body for create order", "error", err)
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		reqCtx.StatusCode = http.StatusBadRequest
		h.logger.LogResponse(reqCtx)
		return
	}

	order, err := h.orderService.CreateOrder(createReq)
	if err != nil {
		h.logger.Warn("Failed to create order", "error", err)
		statusCode := http.StatusBadRequest

		if strings.Contains(err.Error(), "not found in menu") {
			statusCode = http.StatusNotFound
		} else if strings.Contains(err.Error(), "insufficient inventory") {
			statusCode = http.StatusConflict
		} else if strings.Contains(err.Error(), "foreign key") || strings.Contains(err.Error(), "violates") {
			statusCode = http.StatusUnprocessableEntity
		}

		h.writeErrorResponse(w, statusCode, err.Error())
		reqCtx.StatusCode = statusCode
		h.logger.LogResponse(reqCtx)
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, order)
	reqCtx.StatusCode = http.StatusCreated
	h.logger.LogResponse(reqCtx)
}

// GetAllOrders handles GET /api/v1/orders
func (h *OrderHandler) GetAllOrders(w http.ResponseWriter, r *http.Request) {
	reqCtx := &logger.RequestContext{
		Method:     r.Method,
		Path:       r.URL.Path,
		RemoteAddr: r.RemoteAddr,
		StartTime:  time.Now(),
	}
	h.logger.LogRequest(reqCtx)

	orders, err := h.orderService.GetAllOrders()
	if err != nil {
		h.logger.Error("Failed to get all orders", "error", err)
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to fetch orders")
		reqCtx.StatusCode = http.StatusInternalServerError
		h.logger.LogResponse(reqCtx)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, orders)
	reqCtx.StatusCode = http.StatusOK
	h.logger.LogResponse(reqCtx)
}

// GetOrderByID handles GET /api/v1/orders/{id}
func (h *OrderHandler) GetOrderByID(w http.ResponseWriter, r *http.Request) {
	reqCtx := &logger.RequestContext{
		Method:     r.Method,
		Path:       r.URL.Path,
		RemoteAddr: r.RemoteAddr,
		StartTime:  time.Now(),
	}
	h.logger.LogRequest(reqCtx)

	id := h.extractIDFromPath(r)
	if err := h.validateOrderID(id); err != nil {
		h.logger.Warn("Invalid order ID", "id", id, "error", err)
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid order ID")
		reqCtx.StatusCode = http.StatusBadRequest
		h.logger.LogResponse(reqCtx)
		return
	}

	order, err := h.orderService.GetOrderByID(id)
	if err != nil {
		h.logger.Warn("Order not found", "id", id, "error", err)
		h.writeErrorResponse(w, http.StatusNotFound, "Order not found")
		reqCtx.StatusCode = http.StatusNotFound
		h.logger.LogResponse(reqCtx)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, order)
	reqCtx.StatusCode = http.StatusOK
	h.logger.LogResponse(reqCtx)
}

// UpdateOrder handles PUT /api/v1/orders/{id}
func (h *OrderHandler) UpdateOrder(w http.ResponseWriter, r *http.Request) {
	reqCtx := &logger.RequestContext{
		Method:     r.Method,
		Path:       r.URL.Path,
		RemoteAddr: r.RemoteAddr,
		StartTime:  time.Now(),
	}
	h.logger.LogRequest(reqCtx)

	id := h.extractIDFromPath(r)
	if err := h.validateOrderID(id); err != nil {
		h.logger.Warn("Invalid order ID", "id", id, "error", err)
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid order ID")
		reqCtx.StatusCode = http.StatusBadRequest
		h.logger.LogResponse(reqCtx)
		return
	}

	var updateReq service.UpdateOrderRequest
	if err := h.parseRequestBody(r, &updateReq); err != nil {
		h.logger.Warn("Invalid request body for update order", "error", err)
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		reqCtx.StatusCode = http.StatusBadRequest
		h.logger.LogResponse(reqCtx)
		return
	}

	err := h.orderService.UpdateOrder(id, updateReq)
	if err != nil {
		h.logger.Warn("Failed to update order", "id", id, "error", err)
		statusCode := http.StatusBadRequest

		if strings.Contains(err.Error(), "not found") {
			statusCode = http.StatusNotFound
		} else if strings.Contains(err.Error(), "insufficient inventory") {
			statusCode = http.StatusConflict
		} else if strings.Contains(err.Error(), "cannot update closed order") {
			statusCode = http.StatusConflict
		} else if strings.Contains(err.Error(), "foreign key") || strings.Contains(err.Error(), "violates") {
			statusCode = http.StatusUnprocessableEntity
		}

		h.writeErrorResponse(w, statusCode, err.Error())
		reqCtx.StatusCode = statusCode
		h.logger.LogResponse(reqCtx)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{"order_id": id, "message": "Order updated"})
	reqCtx.StatusCode = http.StatusOK
	h.logger.LogResponse(reqCtx)
}

// DeleteOrder handles DELETE /api/v1/orders/{id}
func (h *OrderHandler) DeleteOrder(w http.ResponseWriter, r *http.Request) {
	reqCtx := &logger.RequestContext{
		Method:     r.Method,
		Path:       r.URL.Path,
		RemoteAddr: r.RemoteAddr,
		StartTime:  time.Now(),
	}
	h.logger.LogRequest(reqCtx)

	id := h.extractIDFromPath(r)
	if err := h.validateOrderID(id); err != nil {
		h.logger.Warn("Invalid order ID", "id", id, "error", err)
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid order ID")
		reqCtx.StatusCode = http.StatusBadRequest
		h.logger.LogResponse(reqCtx)
		return
	}

	err := h.orderService.DeleteOrder(id)
	if err != nil {
		h.logger.Warn("Failed to delete order", "id", id, "error", err)
		statusCode := http.StatusNotFound
		if strings.Contains(err.Error(), "foreign key") || strings.Contains(err.Error(), "violates") {
			statusCode = http.StatusConflict
		}
		h.writeErrorResponse(w, statusCode, "Order not found")
		reqCtx.StatusCode = statusCode
		h.logger.LogResponse(reqCtx)
		return
	}

	h.writeJSONResponse(w, http.StatusNoContent, map[string]interface{}{"order_id": id, "message": "Order deleted"})
	reqCtx.StatusCode = http.StatusNoContent
	h.logger.LogResponse(reqCtx)
}

// CloseOrder handles POST /api/v1/orders/{id}/close
func (h *OrderHandler) CloseOrder(w http.ResponseWriter, r *http.Request) {
	reqCtx := &logger.RequestContext{
		Method:     r.Method,
		Path:       r.URL.Path,
		RemoteAddr: r.RemoteAddr,
		StartTime:  time.Now(),
	}
	h.logger.LogRequest(reqCtx)

	id := h.extractIDFromPath(r)
	if err := h.validateOrderID(id); err != nil {
		h.logger.Warn("Invalid order ID", "id", id, "error", err)
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid order ID")
		reqCtx.StatusCode = http.StatusBadRequest
		h.logger.LogResponse(reqCtx)
		return
	}

	err := h.orderService.CloseOrder(id)
	if err != nil {
		h.logger.Warn("Failed to close order", "id", id, "error", err)
		statusCode := http.StatusNotFound
		if strings.Contains(err.Error(), "already closed") {
			statusCode = http.StatusConflict
		}
		h.writeErrorResponse(w, statusCode, err.Error())
		reqCtx.StatusCode = statusCode
		h.logger.LogResponse(reqCtx)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{"order_id": id, "message": "Order closed"})
	reqCtx.StatusCode = http.StatusOK
	h.logger.LogResponse(reqCtx)
}

// GetNumberOfOrderedItems handles GET /api/v1/orders/numberOfOrderedItems
func (h *OrderHandler) GetNumberOfOrderedItems(w http.ResponseWriter, r *http.Request) {
	reqCtx := &logger.RequestContext{
		Method:     r.Method,
		Path:       r.URL.Path,
		RemoteAddr: r.RemoteAddr,
		StartTime:  time.Now(),
	}
	h.logger.LogRequest(reqCtx)

	query := r.URL.Query()
	startDate := query.Get("startDate")
	endDate := query.Get("endDate")

	h.logger.Debug("Processing numberOfOrderedItems request", "startDate", startDate, "endDate", endDate)

	result, err := h.orderService.GetNumberOfOrderedItems(startDate, endDate)
	if err != nil {
		h.logger.Warn("Failed to get number of ordered items", "error", err)
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		reqCtx.StatusCode = http.StatusBadRequest
		h.logger.LogResponse(reqCtx)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, result)
	reqCtx.StatusCode = http.StatusOK
	h.logger.LogResponse(reqCtx)
}

// BatchProcessOrders handles POST /api/v1/orders/batch-process
func (h *OrderHandler) BatchProcessOrders(w http.ResponseWriter, r *http.Request) {
	reqCtx := &logger.RequestContext{
		Method:     r.Method,
		Path:       r.URL.Path,
		RemoteAddr: r.RemoteAddr,
		StartTime:  time.Now(),
	}
	h.logger.LogRequest(reqCtx)

	var batchReq models.BatchOrderRequest
	if err := h.parseRequestBody(r, &batchReq); err != nil {
		h.logger.Warn("Invalid request body for batch process orders", "error", err)
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		reqCtx.StatusCode = http.StatusBadRequest
		h.logger.LogResponse(reqCtx)
		return
	}

	if len(batchReq.Orders) == 0 {
		h.logger.Warn("Empty batch orders request")
		h.writeErrorResponse(w, http.StatusBadRequest, "No orders provided for processing")
		reqCtx.StatusCode = http.StatusBadRequest
		h.logger.LogResponse(reqCtx)
		return
	}

	if len(batchReq.Orders) > 100 {
		h.logger.Warn("Batch size too large", "size", len(batchReq.Orders))
		h.writeErrorResponse(w, http.StatusBadRequest, "Batch size cannot exceed 100 orders")
		reqCtx.StatusCode = http.StatusBadRequest
		h.logger.LogResponse(reqCtx)
		return
	}

	response, err := h.orderService.BatchProcessOrders(batchReq)
	if err != nil {
		h.logger.Error("Failed to batch process orders", "error", err)
		statusCode := http.StatusInternalServerError

		if strings.Contains(err.Error(), "validation failed") {
			statusCode = http.StatusBadRequest
		} else if strings.Contains(err.Error(), "not found") {
			statusCode = http.StatusNotFound
		} else if strings.Contains(err.Error(), "insufficient inventory") {
			statusCode = http.StatusConflict
		}

		h.writeErrorResponse(w, statusCode, err.Error())
		reqCtx.StatusCode = statusCode
		h.logger.LogResponse(reqCtx)
		return
	}

	h.logger.Info("Batch processing completed",
		"total_orders", response.Summary.TotalOrders,
		"accepted", response.Summary.Accepted,
		"rejected", response.Summary.Rejected,
		"total_revenue", response.Summary.TotalRevenue)

	h.writeJSONResponse(w, http.StatusOK, response)
	reqCtx.StatusCode = http.StatusOK
	h.logger.LogResponse(reqCtx)
}

// Private helper methods

// writeJSONResponse writes JSON response with given status code and data
func (h *OrderHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			h.logger.Error("Failed to encode JSON response", "error", err)
			http.Error(w, `{"error":"failed to encode response"}`, http.StatusInternalServerError)
		}
	}
}

// writeErrorResponse writes an error response with given status code and message
func (h *OrderHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	resp := map[string]string{"error": message}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error("Failed to encode error response", "error", err)
	}
}

// parseRequestBody parses JSON request body into the target struct
func (h *OrderHandler) parseRequestBody(r *http.Request, target interface{}) error {
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

// extractIDFromPath extracts ID from URL path (expects /api/v1/orders/{id} or /api/v1/orders/{id}/close)
func (h *OrderHandler) extractIDFromPath(r *http.Request) string {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/orders/")

	// Handle /api/v1/orders/{id}/close case
	path = strings.TrimSuffix(path, "/close")

	// Split by '/' and return the first segment (the ID)
	parts := strings.Split(path, "/")
	if len(parts) > 0 && parts[0] != "" {
		return parts[0]
	}

	return ""
}

// validateOrderID validates order ID format
func (h *OrderHandler) validateOrderID(id string) error {
	if id == "" {
		return fmt.Errorf("order ID cannot be empty")
	}

	if len(id) < 36 || len(id) > 36 {
		if !strings.HasPrefix(id, "order") {
			return fmt.Errorf("invalid order ID format")
		}
	}

	return nil
}

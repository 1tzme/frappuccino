package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"hot-coffee/internal/service"
	"hot-coffee/pkg/logger"
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
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		reqCtx.StatusCode = http.StatusBadRequest
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
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		reqCtx.StatusCode = http.StatusBadRequest
		h.logger.LogResponse(reqCtx)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{"id": id, "message": "Order updated"})
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
		h.writeErrorResponse(w, http.StatusNotFound, "Order not found")
		reqCtx.StatusCode = http.StatusNotFound
		h.logger.LogResponse(reqCtx)
		return
	}

	h.writeJSONResponse(w, http.StatusNoContent, map[string]interface{}{"id": id, "message": "Order deleted"})
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
		h.writeErrorResponse(w, http.StatusNotFound, "Order not found")
		reqCtx.StatusCode = http.StatusNotFound
		h.logger.LogResponse(reqCtx)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{"id": id, "message": "Order closed"})
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

	// Basic validation - ID should start with "order" prefix
	if !strings.HasPrefix(id, "order") {
		return fmt.Errorf("invalid order ID format")
	}

	// Check that what follows "order" is a number
	numStr := strings.TrimPrefix(id, "order")
	if numStr == "" {
		return fmt.Errorf("order ID missing number")
	}

	// Validate that the suffix is a valid number
	if _, err := strconv.Atoi(numStr); err != nil {
		return fmt.Errorf("invalid order ID format: %s", id)
	}

	return nil
}

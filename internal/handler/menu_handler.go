package handler

import (
	"net/http"
	"regexp"
	"strings"
	"time"

	"hot-coffee/internal/service"
	"hot-coffee/pkg/logger"
)

type MenuHandler struct {
	menuService service.MenuServiceInterface
	logger      *logger.Logger
}

func NewMenuHandler(menuService service.MenuServiceInterface, logger *logger.Logger) *MenuHandler {
	return &MenuHandler{
		menuService: menuService,
		logger:      logger.WithComponent("menu_handler"),
	}
}

// GetAllMenuItems handles GET /api/v1/menu
func (h *MenuHandler) GetAllMenuItems(w http.ResponseWriter, r *http.Request) {
	reqCtx := &logger.RequestContext{
		Method:     r.Method,
		Path:       r.URL.Path,
		RemoteAddr: r.RemoteAddr,
		StartTime:  time.Now(),
	}
	h.logger.LogRequest(reqCtx)

	items, err := h.menuService.GetAllMenuItems()
	if err != nil {
		h.logger.Error("Failed to get all menu items")
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to fetch menu items")
		reqCtx.StatusCode = http.StatusInternalServerError
		h.logger.LogResponse(reqCtx)
		return
	}

	writeJSONResponse(w, http.StatusOK, items)
	reqCtx.StatusCode = http.StatusOK
	h.logger.LogResponse(reqCtx)
}

// GetMenuItem handles GET /api/v1/menu/{id}
func (h *MenuHandler) GetMenuItem(w http.ResponseWriter, r *http.Request) {
	reqCtx := &logger.RequestContext{
		Method:     r.Method,
		Path:       r.URL.Path,
		RemoteAddr: r.RemoteAddr,
		StartTime:  time.Now(),
	}
	h.logger.LogRequest(reqCtx)

	id := extractIDFromPath(r)
	item, err := h.menuService.GetMenuItem(id)
	if err != nil {
		h.logger.Warn("Menu item not found", "id", id, "error", err)
		writeErrorResponse(w, http.StatusNotFound, err.Error())
		reqCtx.StatusCode = http.StatusNotFound
		h.logger.LogResponse(reqCtx)
		return
	}

	writeJSONResponse(w, http.StatusOK, item)
	reqCtx.StatusCode = http.StatusOK
	h.logger.LogResponse(reqCtx)
}

// CreateMenuItem handles POST /api/v1/menu
func (h *MenuHandler) CreateMenuItem(w http.ResponseWriter, r *http.Request) {
	reqCtx := &logger.RequestContext{
		Method:     r.Method,
		Path:       r.URL.Path,
		RemoteAddr: r.RemoteAddr,
		StartTime:  time.Now(),
	}
	h.logger.LogRequest(reqCtx)

	var createdReq service.CreateMenuItemRequest
	if err := parseRequestBody(r, &createdReq); err != nil {
		h.logger.Warn("Invalid request body for create", "error", err)
		writeErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		reqCtx.StatusCode = http.StatusBadRequest
		h.logger.LogResponse(reqCtx)
		return
	}

	newID := generateMenuItemID(createdReq.Name)
	item, err := h.menuService.CreateMenuItem(newID, createdReq)
	if err != nil {
		h.logger.Warn("Failed to create menu item", "error", err)
		writeErrorResponse(w, http.StatusBadRequest, err.Error())
		reqCtx.StatusCode = http.StatusBadRequest
		h.logger.LogResponse(reqCtx)
		return
	}

	writeJSONResponse(w, http.StatusCreated, map[string]interface{}{"id": newID, "message": "Menu item created", "item": item})
	reqCtx.StatusCode = http.StatusCreated
	h.logger.LogResponse(reqCtx)
}

// UpdateMenuItem handles PUT /api/v1/menu/{id}
func (h *MenuHandler) UpdateMenuItem(w http.ResponseWriter, r *http.Request) {
	reqCtx := &logger.RequestContext{
		Method:     r.Method,
		Path:       r.URL.Path,
		RemoteAddr: r.RemoteAddr,
		StartTime:  time.Now(),
	}
	h.logger.LogRequest(reqCtx)

	id := extractIDFromPath(r)

	updateReq := service.UpdateMenuItemRequest{}
	if err := parseRequestBody(r, &updateReq); err != nil {
		h.logger.Warn("Invalid request body", "error", err)
		writeErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		reqCtx.StatusCode = http.StatusBadRequest
		h.logger.LogResponse(reqCtx)
		return
	}

	if err := h.menuService.UpdateMenuItem(id, updateReq); err != nil {
		h.logger.Warn("Failed to update menu item", "id", id, "error", err)
		if strings.Contains(err.Error(), "not found") {
			writeErrorResponse(w, http.StatusNotFound, err.Error())
			reqCtx.StatusCode = http.StatusNotFound
		} else {
			writeErrorResponse(w, http.StatusBadRequest, err.Error())
			reqCtx.StatusCode = http.StatusBadRequest
		}
		h.logger.LogResponse(reqCtx)
		return
	}

	writeJSONResponse(w, http.StatusOK, map[string]interface{}{"id": id, "message": "Menu item updated"})
	reqCtx.StatusCode = http.StatusOK
	h.logger.LogResponse(reqCtx)
}

// DeleteMenuItem handles DELETE /api/v1/menu/{id}
func (h *MenuHandler) DeleteMenuItem(w http.ResponseWriter, r *http.Request) {
	reqCtx := &logger.RequestContext{
		Method:     r.Method,
		Path:       r.URL.Path,
		RemoteAddr: r.RemoteAddr,
		StartTime:  time.Now(),
	}
	h.logger.LogRequest(reqCtx)

	id := extractIDFromPath(r)

	if err := h.menuService.DeleteMenuItem(id); err != nil {
		h.logger.Warn("Failed to delete menu item", "id", id, "error", err)
		writeErrorResponse(w, http.StatusNotFound, err.Error())
		reqCtx.StatusCode = http.StatusNotFound
		h.logger.LogResponse(reqCtx)
		return
	}

	writeJSONResponse(w, http.StatusNoContent, nil)
	reqCtx.StatusCode = http.StatusNoContent
	h.logger.LogResponse(reqCtx)
}

// TODO: Implement GetPopularItems HTTP handler - GET /api/v1/menu/aggregations/popular
// - Call menu service for popular items aggregation
// - Return 200 OK with aggregation data or 500 on error
// - Log HTTP request/response
// func (h *MenuHandler) GetPopularItems(w http.ResponseWriter, r *http.Request)

// Private helper methods

// generateMenuItemID - Generates menu item ID based on item name
func generateMenuItemID(name string) string {
	cleaned := strings.ToLower(name)
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	cleaned = reg.ReplaceAllString(cleaned, "_")
	cleaned = strings.Trim(cleaned, "_")

	if cleaned == "" {
		cleaned = "ingredient"
	}

	return cleaned
}

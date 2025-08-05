package router

import (
	"net/http"
	"strings"

	"hot-coffee/internal/handler"
)

func NewRouter(orderHandler *handler.OrderHandler, menuHandler *handler.MenuHandler, inventoryHandler *handler.InventoryHandler, aggregationHandler *handler.AggregationHandler) *http.ServeMux {
	mux := http.NewServeMux()

	api := "/api/v1"
	// Aggregation routes: GET total sales, GET popular items
	mux.HandleFunc(api+"/reports/total-sales", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			aggregationHandler.GetTotalSales(w, r)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})

	mux.HandleFunc(api+"/reports/popular-items", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			aggregationHandler.GetPopularItems(w, r)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})

	// Order collection routes: POST (create), GET (all)
	mux.HandleFunc(api+"/orders", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			orderHandler.CreateOrder(w, r)
			return
		}
		if r.Method == http.MethodGet {
			orderHandler.GetAllOrders(w, r)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})

	// Order item routes: GET (by id), PUT (update), DELETE (delete)
	mux.HandleFunc(api+"/orders/", func(w http.ResponseWriter, r *http.Request) {
		// Check if it's a close order request: POST /api/v1/orders/{id}/close
		if r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/close") {
			orderHandler.CloseOrder(w, r)
			return
		}

		// Regular order operations: GET, PUT, DELETE /api/v1/orders/{id}
		if r.Method == http.MethodGet {
			orderHandler.GetOrderByID(w, r)
			return
		}
		if r.Method == http.MethodPut {
			orderHandler.UpdateOrder(w, r)
			return
		}
		if r.Method == http.MethodDelete {
			orderHandler.DeleteOrder(w, r)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})

	// Menu collection routes: POST (create), GET (all)
	mux.HandleFunc(api+"/menu", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			menuHandler.CreateMenuItem(w, r)
			return
		}
		if r.Method == http.MethodGet {
			menuHandler.GetAllMenuItems(w, r)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})

	// Menu item routes: GET (by id), PUT (update), DELETE (delete)
	mux.HandleFunc(api+"/menu/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			menuHandler.GetMenuItem(w, r)
			return
		}
		if r.Method == http.MethodPut {
			menuHandler.UpdateMenuItem(w, r)
			return
		}
		if r.Method == http.MethodDelete {
			menuHandler.DeleteMenuItem(w, r)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})

	// Inventory collection routes: POST (create), GET (all)
	mux.HandleFunc(api+"/inventory", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			inventoryHandler.CreateInventoryItem(w, r)
			return
		}
		if r.Method == http.MethodGet {
			inventoryHandler.GetAllInventoryItems(w, r)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})

	// Inventory item routes: GET (by id), PUT (update), DELETE (delete)
	mux.HandleFunc(api+"/inventory/", func(w http.ResponseWriter, r *http.Request) {
		// /api/v1/inventory/{id}
		if r.Method == http.MethodGet {
			inventoryHandler.GetInventoryItem(w, r)
			return
		}
		if r.Method == http.MethodPut {
			inventoryHandler.UpdateInventoryItem(w, r)
			return
		}
		if r.Method == http.MethodDelete {
			inventoryHandler.DeleteInventoryItem(w, r)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})
	return mux
}

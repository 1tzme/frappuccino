# 🚀 Hot Coffee Logger - Quick Reference Card

## 📋 **TL;DR - Copy & Paste**

### **Component Setup (Constructor Pattern)**
```go
// Repository
func NewOrderRepository(logger *logger.Logger) *OrderRepository {
    return &OrderRepository{
        logger: logger.WithComponent("order_repository"),
    }
}

// Service  
func NewOrderService(repo *OrderRepository, logger *logger.Logger) *OrderService {
    return &OrderService{
        repo: repo,
        logger: logger.WithComponent("order_service"),
    }
}

// Handler
func NewOrderHandler(service *OrderService, logger *logger.Logger) *OrderHandler {
    return &OrderHandler{
        service: service,
        logger: logger.WithComponent("order_handler"),
    }
}
```

### **Main Application Setup**
```go
// cmd/main.go
func main() {
    // Logger setup
    appLogger := logger.New(logger.Config{
        Level:  logger.LevelInfo,
        Format: "json",
        Output: "stdout",
    })
    
    // Dependency injection chain
    orderRepo := dal.NewOrderRepository(appLogger)
    orderService := service.NewOrderService(orderRepo, appLogger)
    orderHandler := handler.NewOrderHandler(orderService, appLogger)
    
    // HTTP server with middleware
    mux := http.NewServeMux()
    mux.HandleFunc("/api/orders", orderHandler.CreateOrder)
    
    server := &http.Server{
        Addr: ":8080",
        Handler: appLogger.HTTPMiddleware(mux), // Auto HTTP logging
    }
    
    appLogger.Info("Server starting", "port", 8080)
    server.ListenAndServe()
}
```

## 🎯 **What to Log When**

| **Situation** | **Level** | **Code** |
|---------------|-----------|----------|
| **HTTP Request** | AUTO | `// Middleware handles this` |
| **Business Success** | INFO | `s.logger.Info("Order created", "order_id", id)` |
| **Validation Error** | WARN | `s.logger.Warn("Invalid data", "field", "email")` |
| **System Error** | ERROR | `s.logger.Error("DB failed", "error", err)` |
| **Debug Info** | DEBUG | `s.logger.Debug("Processing", "step", "validation")` |

## 🏗️ **Common Patterns**

### **Error Handling**
```go
if err != nil {
    s.logger.Error("Operation failed",
        "error", err,
        "operation", "create_order",
        "entity_id", orderID)
    return err
}
```

### **Business Events**
```go
s.logger.LogBusinessEvent("order_created", orderID, map[string]interface{}{
    "customer_id": customerID,
    "amount": 99.99,
})
```

### **Performance Tracking** (Future feature - currently commented out)
```go
// start := time.Now()
// // ... operation
// s.logger.LogPerformance(logger.Performance{
//     Operation: "database_query",
//     Duration: time.Since(start),
// })
```

### **Security Events** (Future feature - currently commented out)
```go
// s.logger.LogSecurity("login_failed", "medium", map[string]interface{}{
//     "username": username,
//     "ip": clientIP,
// })
```

## ⚡ **Layer-Specific Quick Guides**

### **Handler Layer** - Log requests/responses + validation
```go
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
    // Validation logging
    if customerID == "" {
        h.logger.Warn("Missing customer ID", "endpoint", "/orders")
        return
    }
    
    // Service call logging
    order, err := h.service.CreateOrder(req)
    if err != nil {
        h.logger.Error("Service call failed", "error", err)
        return
    }
    
    h.logger.Info("Order created", "order_id", order.ID)
}
```

### **Service Layer** - Log business logic + events
```go
func (s *OrderService) CreateOrder(req CreateOrderRequest) (*Order, error) {
    s.logger.Info("Processing order", "customer_id", req.CustomerID)
    
    // Business validation
    if len(req.Items) == 0 {
        s.logger.Warn("Empty order rejected", "customer_id", req.CustomerID)
        return nil, errors.New("no items")
    }
    
    // Success
    s.logger.LogBusinessEvent("order_created", order.ID, map[string]interface{}{
        "customer_id": req.CustomerID,
        "total": order.Total,
    })
    
    return order, nil
}
```

### **Repository Layer** - Log data operations
```go
func (r *OrderRepository) Create(order *Order) error {
    r.logger.Debug("Creating order", "order_id", order.ID)
    
    err := r.writeToFile(order)
    if err != nil {
        r.logger.Error("File write failed", 
            "error", err, 
            "order_id", order.ID)
        return err
    }
    
    r.logger.Info("Order persisted", "order_id", order.ID)
    return nil
}
```

## 🚫 **DON'T Do This**

```go
// ❌ Global logger
log.Println("Order created")

// ❌ No context
logger.Info("Error occurred")

// ❌ Sensitive data
logger.Info("User login", "password", password)

// ❌ String concatenation
logger.Info("Order " + orderID + " created")

// ❌ Loop logging
for _, item := range items {
    logger.Debug("Processing item", "item", item)
}
```

## ✅ **DO This Instead**

```go
// ✅ Injected logger
s.logger.Info("Order created", "order_id", orderID)

// ✅ Rich context
s.logger.Error("Database error", 
    "error", err,
    "operation", "create_order",
    "table", "orders")

// ✅ Safe logging
s.logger.Info("User authenticated", "user_id", userID)

// ✅ Structured logging
s.logger.Info("Order created", "order_id", orderID, "amount", 99.99)

// ✅ Batch logging
s.logger.Info("Processing items", "count", len(items))
// ... process all items
s.logger.Info("Items processed", "success", successCount, "failed", failCount)
```

## 🔧 **Environment Configs**

### **Development**
```go
logger.Config{
    Level: logger.LevelDebug,
    Format: "text",
    Output: "stdout",
    EnableColors: true,
    EnableCaller: true,
}
```

### **Production**
```go
logger.Config{
    Level: logger.LevelInfo,
    Format: "json", 
    Output: "/var/log/app.log",
    EnableColors: false,
    EnableCaller: false,
    SensitiveKeys: []string{"password", "token"},
}
```

## 📊 **Health & Metrics**

```go
// Health check endpoint
http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
    health := appLogger.HealthCheck()
    json.NewEncoder(w).Encode(health)
})

// Get metrics
metrics := appLogger.GetMetrics()
fmt.Printf("Total logs: %d, Error rate: %.2f%%", 
    metrics.TotalLogs, metrics.ErrorRate)
```

---

**💡 Remember**: Logger flows down through dependency injection: `main.go → Repository → Service → Handler`

**🎯 Key Rule**: Each layer adds its own context, logs its own concerns, passes the logger down.

**📖 Full Guide**: See `DEVELOPER_GUIDE.md` for complete examples and patterns.

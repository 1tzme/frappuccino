# Hot Coffee - Three-Layered Architecture Documentation

## Project Overview

Hot Coffee is a coffee shop management system built with Go using a three-layered architecture pattern. The application provides HTTP endpoints for managing orders, menu items, and inventory with data persistence in JSON files.

## Commit Message Format

```
<type>(optional-scope): <description>
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix  
- `docs`: Documentation only changes
- `style`: Code style changes (formatting, missing semi)
- `refactor`: Code change that isn't a feature or bug fix
- `perf`: Performance improvement
- `test`: Adding or updating tests
- `chore`: Misc tasks (build process, config, deps)


## Architecture Overview

The project follows a three-layered architecture pattern:

1. **Presentation Layer (Handlers)** - HTTP request/response handling
2. **Business Logic Layer (Services)** - Core business logic and rules
3. **Data Access Layer (Repositories)** - Data storage and retrieval

## Project Structure

```
hot-coffee/
├── cmd/                            # Entry point of the application
│   └── main.go                     # Application bootstrap and HTTP server setup
├── internal/                       # Application code divided into layers
│   ├── handler/                    # Presentation layer handling HTTP requests
│   │   ├── order_handler.go        # Order-related HTTP endpoints
│   │   ├── menu_handler.go         # Menu-related HTTP endpoints
│   │   └── inventory_handler.go    # Inventory-related HTTP endpoints
│   ├── service/                    # Business logic layer with interfaces and implementations
│   │   ├── order_service.go        # Order business logic
│   │   ├── menu_service.go         # Menu business logic
│   │   └── inventory_service.go    # Inventory business logic
│   └── dal/                        # Data access layer with interfaces and implementations
│       ├── order_repository.go     # Order data operations
│       ├── menu_repository.go      # Menu data operations
│       └── inventory_repository.go # Inventory data operations
├── models/                         # Data models shared across layers
│   ├── order.go                    # Order-related data structures
│   ├── menu.go                     # Menu-related data structures
│   └── inventory.go                # Inventory-related data structures
├── data/                           # JSON data storage directory
│   ├── orders.json                 # Orders data persistence
│   ├── menu_items.json             # Menu items data persistence
│   └── inventory.json              # Inventory data persistence
├── go.mod                          # Go module definition
├── go.sum                          # Go module dependencies checksum
└── README.md                       # Project documentation
```

## Layer Responsibilities

### 1. Presentation Layer (Handlers)

**Location**: `internal/handler/`

**Responsibilities**:
- Handle HTTP requests and responses
- Parse input data and format output data
- Invoke appropriate methods from the Business Logic Layer
- Validate input data and return meaningful error messages
- Return appropriate HTTP status codes

**Implementation Details**:
- Organized by entity (orders, menu, inventory)
- Uses Gin framework for HTTP routing
- Handles JSON serialization/deserialization
- Implements proper error handling and logging

**Key Files**:
- `order_handler.go` - Order CRUD operations
- `menu_handler.go` - Menu item management
- `inventory_handler.go` - Inventory management

### 2. Business Logic Layer (Services)

**Location**: `internal/service/`

**Responsibilities**:
- Implement core business logic and rules
- Define interfaces for services to promote decoupling
- Perform data processing and validation
- Handle aggregations and computations
- Coordinate between handlers and repositories

**Implementation Details**:
- Service interfaces for testability and flexibility
- Business rule validation
- Data aggregation methods
- Error handling and logging
- Independent and testable components

**Key Files**:
- `order_service.go` - Order business logic and validation
- `menu_service.go` - Menu management and availability logic
- `inventory_service.go` - Inventory tracking and low-stock alerts

### 3. Data Access Layer (Repositories)

**Location**: `internal/dal/`

**Responsibilities**:
- Manage data storage and retrieval operations
- Interact with JSON files to persist and read data
- Ensure data integrity and consistency
- Provide interfaces for flexibility

**Implementation Details**:
- Repository interfaces for each entity
- JSON file-based persistence
- Thread-safe operations using mutexes
- Data validation and error handling
- Atomic file operations

**Key Files**:
- `order_repository.go` - Order data persistence
- `menu_repository.go` - Menu item data persistence
- `inventory_repository.go` - Inventory data persistence

## API Endpoints

### Order Management

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/orders` | Create a new order |
| GET | `/api/v1/orders` | Get all orders |
| GET | `/api/v1/orders/:id` | Get order by ID |
| PUT | `/api/v1/orders/:id/status` | Update order status |
| GET | `/api/v1/orders/aggregations/sales` | Get total sales |

### Menu Management

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/menu` | Get all menu items |
| POST | `/api/v1/menu` | Create new menu item |
| PUT | `/api/v1/menu/:id` | Update menu item |
| DELETE | `/api/v1/menu/:id` | Delete menu item |
| GET | `/api/v1/menu/aggregations/popular` | Get popular menu items |

### Inventory Management

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/inventory` | Get all inventory items |
| PUT | `/api/v1/inventory/:id` | Update inventory item |
| GET | `/api/v1/inventory/low-stock` | Get low stock items |

## Data Models

### Order Model

```go
type Order struct {
    ID          string      `json:"id"`
    CustomerID  string      `json:"customer_id"`
    Items       []OrderItem `json:"items"`
    TotalAmount float64     `json:"total_amount"`
    Status      OrderStatus `json:"status"`
    CreatedAt   time.Time   `json:"created_at"`
    UpdatedAt   time.Time   `json:"updated_at"`
}

type OrderStatus string
const (
    OrderStatusPending    OrderStatus = "pending"
    OrderStatusPreparing  OrderStatus = "preparing"
    OrderStatusReady      OrderStatus = "ready"
    OrderStatusCompleted  OrderStatus = "completed"
    OrderStatusCancelled  OrderStatus = "cancelled"
)
```

### Menu Model

```go
type MenuItem struct {
    ID          string       `json:"id"`
    Name        string       `json:"name"`
    Description string       `json:"description"`
    Category    MenuCategory `json:"category"`
    Price       float64      `json:"price"`
    Available   bool         `json:"available"`
    CreatedAt   time.Time    `json:"created_at"`
    UpdatedAt   time.Time    `json:"updated_at"`
}

type MenuCategory string
const (
    CategoryCoffee    MenuCategory = "coffee"
    CategoryTea       MenuCategory = "tea"
    CategoryPastry    MenuCategory = "pastry"
    CategorySandwich  MenuCategory = "sandwich"
    CategoryDrink     MenuCategory = "drink"
)
```

### Inventory Model

```go
type InventoryItem struct {
    ID           string    `json:"id"`
    Name         string    `json:"name"`
    Description  string    `json:"description"`
    Quantity     int       `json:"quantity"`
    MinThreshold int       `json:"min_threshold"`
    Unit         string    `json:"unit"`
    LastUpdated  time.Time `json:"last_updated"`
}
```

## Data Persistence

### JSON File Storage

Data is persisted in JSON files located in the `data/` directory:

- `data/orders.json` - Contains all order records
- `data/menu_items.json` - Contains all menu item records
- `data/inventory.json` - Contains all inventory item records

### File Operations

- **Atomic Writes**: Files are written atomically to prevent corruption
- **Backup**: Previous versions are backed up before updates
- **Thread Safety**: Concurrent access is handled with mutexes
- **Error Recovery**: Graceful handling of file system errors

## Configuration

### Environment Variables

The application can be configured using environment variables. You can set these in your environment or create a `.env` file in the project root.

| Variable | Default | Description |
|----------|---------|-------------|
| `HOST` | `localhost` | HTTP server host address |
| `PORT` | `8080` | HTTP server port |
| `LOG_LEVEL` | `info` | Logging level (`debug`, `info`, `warn`, `error`) |
| `LOG_FORMAT` | `json` | Log format (`json`, `text`, `console`) |
| `LOG_OUTPUT` | `stdout` | Log output (`stdout`, `stderr`, or file path) |
| `LOG_ENABLE_CALLER` | `true` | Include file and line info in logs |
| `LOG_ENABLE_COLORS` | `false` | Enable colored output for console format |
| `ENVIRONMENT` | `development` | Application environment |
| `DATA_DIR` | `./data` | Data storage directory |

### Environment File Setup

1. Copy the example environment file:
```bash
cp .env.example .env
```

2. Edit `.env` with your preferred settings:
```bash
# Example .env file
HOST=localhost
PORT=8080
LOG_LEVEL=debug
LOG_FORMAT=text
LOG_ENABLE_COLORS=true
ENVIRONMENT=development
```

### Server Configuration

The application starts an HTTP server with configurable settings:

```go
// Default configuration
config := Config{
    Port:    getEnv("PORT", "8080"),
    DataDir: getEnv("DATA_DIR", "./data"),
    LogLevel: getEnv("LOG_LEVEL", "info"),
}
```

## Logging

The application uses Go's `log/slog` package for structured logging:

### Log Levels

- **DEBUG**: Detailed debugging information
- **INFO**: General information about application flow
- **WARN**: Warning conditions that don't prevent operation
- **ERROR**: Error conditions that require attention

### Log Events

- HTTP request/response logging
- Business logic operations
- Data access operations
- Error conditions
- Performance metrics

### Example Log Entries

```json
{
  "time": "2025-07-10T10:30:00Z",
  "level": "INFO",
  "msg": "Order created successfully",
  "order_id": "ord-123",
  "customer_id": "cust-456",
  "total_amount": 15.50
}

{
  "time": "2025-07-10T10:31:00Z",
  "level": "ERROR",
  "msg": "Failed to update inventory",
  "item_id": "inv-789",
  "error": "file not found"
}
```

## Logging Architecture

### Logger Placement Strategy

The logger should be properly positioned across all layers of the application following dependency injection principles:

#### 1. Centralized Logger Configuration (`pkg/logger/`)

Create a centralized logger package that provides:
- Logger configuration and initialization
- Structured logging setup with slog
- Environment-based log level configuration
- Consistent log formatting across the application

```go
// pkg/logger/logger.go
package logger

import (
    "log/slog"
    "os"
)

type Logger struct {
    *slog.Logger
}

func New(config Config) *Logger {
    // Logger initialization logic
}
```

#### 2. Application Bootstrap (`cmd/main.go`)

Initialize the logger at the application entry point and inject it into all layers:

```go
// cmd/main.go
func main() {
    // Initialize logger
    loggerConfig := logger.Config{
        Level:  getEnv("LOG_LEVEL", "info"),
        Format: getEnv("LOG_FORMAT", "json"),
    }
    appLogger := logger.New(loggerConfig)
    
    // Inject logger into repositories
    orderRepo := dal.NewOrderRepository(appLogger)
    
    // Inject logger into services  
    orderService := service.NewOrderService(orderRepo, appLogger)
    
    // Inject logger into handlers
    orderHandler := handler.NewOrderHandler(orderService, appLogger)
}
```

#### 3. Data Access Layer (`internal/dal/`)

Repositories should accept logger via constructor and log:
- Data operations (create, read, update, delete)
- File I/O operations
- Error conditions
- Performance metrics

```go
// internal/dal/order_repository.go
type OrderRepository struct {
    orders map[string]*models.Order
    mutex  sync.RWMutex
    logger *logger.Logger  // Injected logger
}

func NewOrderRepository(logger *logger.Logger) *OrderRepository {
    return &OrderRepository{
        orders: make(map[string]*models.Order),
        logger: logger.WithContext("component", "order_repository"),
    }
}

func (r *OrderRepository) Create(order *models.Order) error {
    r.logger.Info("Creating order", "order_id", order.ID)
    // ... implementation
    r.logger.Info("Order created successfully", "order_id", order.ID)
}
```

#### 4. Business Logic Layer (`internal/service/`)

Services should log:
- Business logic operations
- Validation failures
- Business rule violations
- Important state changes

```go
// internal/service/order_service.go
type OrderService struct {
    orderRepo     *dal.OrderRepository
    inventoryRepo *dal.InventoryRepository
    logger        *logger.Logger  // Injected logger
}

func NewOrderService(orderRepo *dal.OrderRepository, logger *logger.Logger) *OrderService {
    return &OrderService{
        orderRepo: orderRepo,
        logger:    logger.WithContext("component", "order_service"),
    }
}

func (s *OrderService) CreateOrder(req CreateOrderRequest) (*models.Order, error) {
    s.logger.Info("Processing order creation", "customer_id", req.CustomerID)
    // ... business logic
    s.logger.Info("Order processed successfully", "order_id", order.ID)
}
```

#### 5. Presentation Layer (`internal/handler/`)

Handlers should log:
- HTTP requests and responses
- Input validation errors
- HTTP status codes returned
- Request processing time

```go
// internal/handler/order_handler.go
type OrderHandler struct {
    orderService *service.OrderService
    logger       *logger.Logger  // Injected logger
}

func NewOrderHandler(orderService *service.OrderService, logger *logger.Logger) *OrderHandler {
    return &OrderHandler{
        orderService: orderService,
        logger:       logger.WithContext("component", "order_handler"),
    }
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
    h.logger.Info("Received order creation request", "method", c.Request.Method, "path", c.Request.URL.Path)
    // ... handler logic
}
```

### Logger Best Practices

#### 1. **Dependency Injection**
- Pass logger as a constructor parameter to all components
- Don't use global logger instances in business logic
- Create logger context for each component

#### 2. **Structured Logging**
- Use key-value pairs for log context
- Include relevant IDs (order_id, customer_id, etc.)
- Add component context to identify log source

#### 3. **Log Levels**
- **DEBUG**: Detailed debugging information
- **INFO**: General application flow
- **WARN**: Potentially harmful situations
- **ERROR**: Error events that don't stop the application

#### 4. **Context Enrichment**
```go
// Add context to logger for better traceability
logger := baseLogger.WithContext(
    "component", "order_service",
    "version", "1.0.0",
    "environment", "production",
)
```

#### 5. **Request Tracing**
```go
// Add request ID for request tracing
func (h *OrderHandler) CreateOrder(c *gin.Context) {
    requestID := c.GetHeader("X-Request-ID")
    logger := h.logger.WithContext("request_id", requestID)
    
    logger.Info("Processing order creation request")
    // ... rest of handler
}
```

#### 6. **Error Logging**
```go
// Log errors with full context
if err := h.orderService.CreateOrder(req); err != nil {
    h.logger.Error("Failed to create order",
        "error", err,
        "customer_id", req.CustomerID,
        "item_count", len(req.Items),
    )
    return
}
```

### Logger Configuration

The logger configuration should be environment-specific:

```go
// Development
logger.Config{
    Level:  "debug",
    Format: "text",
}

// Production
logger.Config{
    Level:  "info", 
    Format: "json",
}
```

This approach ensures:
- **Consistency**: All components use the same logging format
- **Traceability**: Logs can be traced through the entire request lifecycle
- **Testability**: Logger can be mocked for unit tests
- **Maintainability**: Centralized logging configuration
- **Performance**: Appropriate log levels for different environments
## Aggregations

### Sales Aggregations

- **Total Sales**: Sum of all completed orders
- **Sales by Period**: Daily, weekly, monthly sales
- **Sales by Category**: Revenue breakdown by menu category

### Menu Aggregations

- **Popular Items**: Most frequently ordered items
- **Revenue by Item**: Total revenue per menu item
- **Category Performance**: Sales performance by category

### Inventory Aggregations

- **Low Stock Items**: Items below minimum threshold
- **Inventory Value**: Total value of current inventory
- **Usage Patterns**: Consumption rates by item

## Getting Started

### Prerequisites

- Go 1.21 or higher
- Git

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd hot-coffee
```

2. Set up environment configuration:
```bash
cp .env.example .env
```

3. Edit `.env` file with your preferred settings:
```bash
# Example customizations
LOG_LEVEL=debug
LOG_FORMAT=text
LOG_ENABLE_COLORS=true
```

4. Build and run the application:
```bash
go build -o hot-coffee ./cmd/main.go
./hot-coffee
```

### Usage

```bash
$ ./hot-coffee --help
Coffee Shop Management System

Usage:
  hot-coffee [--port <N>] [--dir <S>] 
  hot-coffee --help

Options:
  --help       Show this screen.
  --port N     Port number.
  --dir S      Path to the data directory.
```

### Development

1. Follow the layered architecture pattern
2. Add new features by creating corresponding files in each layer
3. Ensure proper error handling and logging
4. Write unit tests for business logic
5. Document API changes

## Best Practices

### Code Organization

- Keep layers separate and well-defined
- Use interfaces for dependency injection
- Follow Go naming conventions
- Implement proper error handling

### Data Management

- Validate data at service layer
- Use atomic file operations
- Implement proper backup strategies
- Handle concurrent access safely

### API Design

- Use RESTful conventions
- Return consistent error formats
- Implement proper HTTP status codes
- Version your APIs

### Testing

- Write unit tests for business logic
- Mock dependencies for isolation
- Test error conditions
- Implement integration tests

## Development
````markdown
This documentation provides a comprehensive guide for understanding, developing, and maintaining the Hot Coffee application using the three-layered architecture pattern.
````

# Frappuccino - Coffee Shop Management System

## Project Overview

Frappuccino is a coffee shop management system built with Go using a three-layered architecture pattern. The application provides HTTP endpoints for managing orders, menu items, and inventory with **PostgreSQL database persistence**, comprehensive transaction management, and business analytics capabilities.

## 🎉 **MIGRATION COMPLETED: JSON → PostgreSQL**

✅ **Successfully migrated from JSON file-based storage to PostgreSQL database**

### ✅ **Completed Migration Features:**

#### 1. Database Infrastructure ✅
- ✅ PostgreSQL 15 database with comprehensive schema
- ✅ Docker Compose orchestration with pgAdmin integration
- ✅ Advanced database schema with 8 tables, UUID primary keys, JSONB fields
- ✅ Custom PostgreSQL enums and proper foreign key relationships
- ✅ Performance-optimized indexes and database connection pooling

#### 2. Repository Layer Migration ✅
- ✅ **Order Repository**: Full PostgreSQL implementation with robust transaction management
- ✅ **Menu Repository**: Complete SQL queries with ingredient relationships
- ✅ **Inventory Repository**: Database operations with batch update capabilities
- ✅ **Aggregation Repository**: Advanced reporting queries with period-based analytics
- ✅ Removed all JSON file operations (loadFromFile, saveToFile, backupFile)

#### 3. **Advanced Transaction Management** ✅
- ✅ **Atomic Operations**: All multi-table operations use proper database transactions
- ✅ **Error Handling**: Comprehensive transaction rollback with detailed logging
- ✅ **Consistency**: Order creation with items, menu updates, inventory batch updates
- ✅ **Improved Rollback Pattern**: Conditional rollback to eliminate unnecessary warnings
- ✅ **Transaction Logging**: Detailed success/failure logging for all database operations

#### 4. Service Layer Enhancements ✅
- ✅ Database-specific error handling and validation
- ✅ Advanced business logic leveraging database features
- ✅ Connection management and retry logic
- ✅ Complex aggregation and reporting capabilities

#### 5. Data Model Evolution ✅
- ✅ Complete PostgreSQL column mappings with proper data types
- ✅ Comprehensive timestamp handling with PostgreSQL TIMESTAMPTZ
- ✅ Advanced relationships and database constraints
- ✅ JSONB support for flexible data structures

#### 6. **Production-Ready Configuration** ✅
- ✅ Environment-based database configuration
- ✅ Docker containerization with health checks
- ✅ Database connection pooling and optimization
- ✅ Comprehensive setup and deployment documentation

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

The project follows a **three-layered architecture pattern** with **PostgreSQL database persistence**:

1. **Presentation Layer (Handlers)** - HTTP request/response handling with comprehensive error management
2. **Business Logic Layer (Services)** - Core business logic with database transaction coordination  
3. **Data Access Layer (Repositories)** - PostgreSQL database operations with advanced transaction management

## 🛠 **Current Technology Stack**

- **Backend**: Go 1.21+ with Gin HTTP framework
- **Database**: PostgreSQL 15 with advanced schema design
- **Containerization**: Docker Compose with multi-container orchestration
- **Database Admin**: pgAdmin 4 for database management
- **Logging**: Structured logging with Go's slog package
- **Architecture**: Three-layer pattern with dependency injection

## Project Structure

```
frappuccino/
├── cmd/                            # Application entry point
│   └── main.go                     # Bootstrap with database connection
├── internal/                       # Core application layers
│   ├── handler/                    # HTTP request handlers with database error handling
│   │   ├── order_handler.go        # Order management endpoints
│   │   ├── menu_handler.go         # Menu CRUD operations
│   │   ├── inventory_handler.go    # Inventory management with pagination
│   │   └── aggregation_handler.go  # Business analytics and reporting
│   ├── service/                    # Business logic with transaction coordination
│   │   ├── order_service.go        # Order processing and validation
│   │   ├── menu_service.go         # Menu management logic
│   │   ├── inventory_service.go    # Inventory tracking and alerts
│   │   └── aggregation_service.go  # Advanced analytics and reporting
│   └── repositories/               # Database access layer with PostgreSQL
│       ├── order_repository.go     # Order data operations with transactions
│       ├── menu_repository.go      # Menu persistence with ingredients
│       ├── inventory_repository.go # Inventory operations with batch updates
│       └── aggregations_repository.go # Complex reporting queries
├── models/                         # Data models with PostgreSQL mapping
│   ├── order.go                    # Order and related structures
│   ├── menu.go                     # Menu items with ingredients
│   └── inventory.go                # Inventory items and transactions
├── pkg/                            # Shared packages and utilities
│   ├── database/                   # PostgreSQL connection management
│   │   └── connection.go           # Connection pooling and configuration
│   ├── logger/                     # Centralized logging system
│   │   ├── logger.go               # Core logging functionality
│   │   └── logger_helper.go        # Logging utilities and helpers
│   ├── envconfig/                  # Environment configuration
│   ├── flags/                      # Command-line argument parsing
│   └── shutdownsetup/              # Graceful shutdown handling
├── docker-compose.yml              # Multi-container orchestration
├── Dockerfile                      # Application containerization
├── init.sql                        # Database schema initialization
├── sample_data_fixed.sql           # Comprehensive sample data
├── .env                            # Environment configuration
└── README.md                       # Project documentation
```
│   ├── menu_items.json             # Menu items data persistence
│   └── inventory.json              # Inventory data persistence
├── go.mod                          # Go module definition
├── go.sum                          # Go module dependencies checksum
└── README.md                       # Project documentation
```

## Layer Responsibilities

### 1. **Presentation Layer (Handlers)**

**Location**: `internal/handler/`

**Responsibilities**:
- Handle HTTP requests and responses with comprehensive error management
- Parse input data and format output data with proper validation
- Coordinate with Business Logic Layer for database operations
- Implement proper HTTP status codes and database error mapping
- Provide comprehensive API endpoints with pagination and filtering

**Implementation Details**:
- **Advanced Error Handling**: Database-specific error responses with proper HTTP codes
- **Input Validation**: Request body parsing with detailed validation messages  
- **Response Formatting**: Consistent JSON responses with error details
- **Database Integration**: Proper handling of PostgreSQL constraints and transactions
- **Business Analytics**: Comprehensive reporting endpoints with period-based queries

**Key Files**:
- `order_handler.go` - Order CRUD with transaction management
- `menu_handler.go` - Menu management with ingredient relationships
- `inventory_handler.go` - Inventory operations with pagination and sorting
- `aggregation_handler.go` - Business analytics and reporting endpoints

### 2. **Business Logic Layer (Services)**

**Location**: `internal/service/`

**Responsibilities**:
- Implement core business logic with database transaction coordination
- Define service interfaces for dependency injection and testability
- Handle complex data processing, validation, and business rules
- Coordinate database operations across multiple repositories
- Provide advanced analytics and aggregation capabilities

**Implementation Details**:
- **Transaction Coordination**: Managing complex multi-table database operations
- **Business Rule Validation**: Advanced validation with database constraint checking
- **Data Aggregation**: Complex reporting and analytics with period-based grouping
- **Error Management**: Comprehensive error handling with database rollback support
- **Performance Optimization**: Efficient database queries with connection pooling

**Key Files**:
- `order_service.go` - Order processing with inventory integration
- `menu_service.go` - Menu management with ingredient validation
- `inventory_service.go` - Advanced inventory tracking with batch operations
- `aggregation_service.go` - Complex business analytics and reporting

### 3. **Data Access Layer (Repositories)**

**Location**: `internal/repositories/`

**Responsibilities**:
- **PostgreSQL Database Operations**: Advanced SQL queries with transaction management
- **Data Integrity**: Comprehensive constraint validation and referential integrity
- **Transaction Management**: Robust atomic operations with proper rollback handling
- **Performance Optimization**: Efficient queries with proper indexing and connection pooling
- **Complex Relationships**: Advanced JOIN operations and data aggregation

**Implementation Details**:
- **Advanced SQL Queries**: Complex reporting queries with window functions and CTEs
- **Transaction Safety**: Proper Begin/Commit/Rollback patterns with detailed logging
- **Connection Management**: Optimized database connection pooling and retry logic
- **Data Validation**: Database constraint validation with meaningful error messages
- **Batch Operations**: Efficient bulk operations for inventory and order processing

**Key Files**:
- `order_repository.go` - Order persistence with atomic transaction management
- `menu_repository.go` - Menu operations with ingredient relationship handling
- `inventory_repository.go` - Inventory management with batch update capabilities
- `aggregations_repository.go` - Complex analytical queries and business reporting

## 🚀 **API Endpoints**

### **Order Management**

| Method | Endpoint | Description | Features |
|--------|----------|-------------|----------|
| POST | `/api/v1/orders` | Create a new order | Transaction-safe with inventory validation |
| GET | `/api/v1/orders` | Get all orders | Comprehensive order details with items |
| GET | `/api/v1/orders/:id` | Get order by ID | Complete order information |
| PUT | `/api/v1/orders/:id` | Update order | Atomic updates with item management |
| DELETE | `/api/v1/orders/:id` | Delete order | Safe cascade deletion |

### **Menu Management**

| Method | Endpoint | Description | Features |
|--------|----------|-------------|----------|
| GET | `/api/v1/menu` | Get all menu items | Complete menu with availability |
| POST | `/api/v1/menu` | Create new menu item | Ingredient relationship management |
| PUT | `/api/v1/menu/:id` | Update menu item | Transaction-safe updates |
| DELETE | `/api/v1/menu/:id` | Delete menu item | Cascade deletion with dependencies |

### **Inventory Management**

| Method | Endpoint | Description | Features |
|--------|----------|-------------|----------|
| GET | `/api/v1/inventory` | Get all inventory items | Complete inventory status |
| PUT | `/api/v1/inventory/:id` | Update inventory item | Atomic quantity updates |
| GET | `/api/v1/inventory/getLeftOvers?sortBy={value}&page={page}&pageSize={pageSize}` | Get inventory with pagination | Advanced sorting and pagination |

### **📊 Business Analytics & Reporting**

| Method | Endpoint | Description | Features |
|--------|----------|-------------|----------|
| GET | `/api/v1/reports/orderedItemsByPeriod?period=day&month=august` | Get orders by day | Period-based analytics |
| GET | `/api/v1/reports/orderedItemsByPeriod?period=month&year=2025` | Get orders by month | Yearly reporting |

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/inventory` | Get all inventory items |
| PUT | `/api/v1/inventory/:id` | Update inventory item |
| GET | `/api/v1/inventory/low-stock` | Get low stock items |

## 🗄️ **PostgreSQL Database Schema**

### **Advanced Database Design**

**Database**: PostgreSQL 15 with comprehensive schema design
**Key Features**: UUID primary keys, JSONB support, custom enums, optimized indexes

### **Core Tables**

#### **Orders Table**
```sql
CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_name VARCHAR(255) NOT NULL,
    status order_status NOT NULL DEFAULT 'pending',
    total_amount DECIMAL(10,2) NOT NULL,
    special_instructions TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

#### **Menu Items Table**
```sql
CREATE TABLE menu_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    category menu_category NOT NULL,
    price DECIMAL(8,2) NOT NULL,
    available BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

#### **Inventory Table**
```sql
CREATE TABLE inventory (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    quantity DECIMAL(10,3) NOT NULL DEFAULT 0,
    unit VARCHAR(50) NOT NULL,
    min_threshold DECIMAL(10,3) NOT NULL DEFAULT 0,
    last_updated TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

### **Advanced Features**

#### **Custom Enums**
```sql
CREATE TYPE order_status AS ENUM ('pending', 'preparing', 'ready', 'closed', 'cancelled');
CREATE TYPE menu_category AS ENUM ('coffee', 'tea', 'pastry', 'sandwich', 'beverage', 'dessert');
CREATE TYPE transaction_type AS ENUM ('purchase', 'usage', 'waste', 'adjustment', 'return');
```

#### **Relationship Tables**
- **`order_items`**: Links orders to menu items with quantities and pricing
- **`menu_ingredients`**: Manages menu item ingredient relationships  
- **`inventory_transactions`**: Tracks all inventory movements with full audit trail
- **`order_status_history`**: Complete order status change tracking

#### **Performance Optimization**
- **Indexes**: Optimized indexes on frequently queried columns
- **Foreign Keys**: Proper referential integrity constraints
- **Connection Pooling**: Advanced connection management for high performance

## 🔧 **Database Transaction Management**

### **Advanced Transaction Features**

#### **✅ Atomic Operations**
All multi-table operations use comprehensive database transactions:

```go
// Example: Order Creation with Items
tx, err := r.db.Begin()
if err != nil {
    return fmt.Errorf("failed to begin transaction: %v", err)
}
defer func() {
    if err != nil {
        r.logger.Warn("Rolling back order creation transaction", "error", err)
        tx.Rollback()
    }
}()

// Insert order and all order items atomically
// ... database operations ...

err = tx.Commit()
if err != nil {
    return fmt.Errorf("failed to commit transaction: %v", err)
}
r.logger.Info("Successfully committed transaction", "order_id", order.ID)
```

#### **🛡️ Transaction Safety Features**
- **Proper Rollback**: Conditional rollback prevents unnecessary warnings
- **Comprehensive Logging**: Detailed transaction success/failure logging  
- **Error Recovery**: Graceful handling of database constraint violations
- **Connection Management**: Optimized connection pooling and retry logic

#### **📊 Transaction Scope**
- **Order Operations**: Order creation, updates, and item management
- **Menu Management**: Menu item creation with ingredient relationships
- **Inventory Updates**: Batch inventory operations with quantity tracking
- **Complex Analytics**: Multi-table reporting queries with consistency

## 🐳 **Docker & Development Environment**

### **Container Orchestration**

**Docker Compose Setup**:
```yaml
version: '3.8'
services:
  app:
    build: .
    ports:
      - "8080:8080"
    depends_on:
      - db
    environment:
      - DB_HOST=db
      - DB_PORT=5432
      - DB_USER=postgres
      - DB_NAME=hotcoffee

  db:
    image: postgres:15
    ports:
      - "5432:5432"
    environment:
      POSTGRES_DB: hotcoffee
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: password
    volumes:
      - ./init.sql:/docker-entrypoint-initdb.d/init.sql
      - ./sample_data_fixed.sql:/docker-entrypoint-initdb.d/sample_data_fixed.sql

  pgadmin:
    image: dpage/pgadmin4
    ports:
      - "5050:80"
    environment:
      PGADMIN_DEFAULT_EMAIL: admin@frappuccino.com
      PGADMIN_DEFAULT_PASSWORD: admin
```

### **Quick Start Commands**
```bash
# Start all services
docker-compose up -d

# Rebuild application with changes
docker-compose build --no-cache app

# View logs
docker-compose logs -f app

# Access database
docker-compose exec db psql -U postgres -d hotcoffee
```

## ⚙️ **Configuration**

### **Database Environment Variables**

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_HOST` | `localhost` | PostgreSQL host address |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_USER` | `postgres` | Database username |
| `DB_PASSWORD` | `password` | Database password |
| `DB_NAME` | `hotcoffee` | Database name |
| `DB_SSLMODE` | `disable` | SSL connection mode |

### **Application Configuration**

| Variable | Default | Description |
|----------|---------|-------------|
| `HOST` | `localhost` | HTTP server host address |
| `PORT` | `8080` | HTTP server port |
| `LOG_LEVEL` | `info` | Logging level (`debug`, `info`, `warn`, `error`) |
| `LOG_FORMAT` | `json` | Log format (`json`, `text`, `console`) |
| `ENVIRONMENT` | `development` | Application environment |

### **Environment Setup**

1. **Copy environment template:**
```bash
cp .env.example .env
```

2. **Configure database connection:**
```bash
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=hotcoffee
DB_SSLMODE=disable

# Application Configuration  
HOST=localhost
PORT=8080
LOG_LEVEL=debug
LOG_FORMAT=text
ENVIRONMENT=development
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

## 🚀 **Getting Started**

### **Prerequisites**

- **Go 1.21+** - Modern Go version with advanced features
- **Docker & Docker Compose** - Container orchestration
- **PostgreSQL 15** - Advanced database features (included in Docker setup)
- **Git** - Version control

### **Quick Start**

#### **1. Clone and Setup**
```bash
git clone <repository-url>
cd frappuccino
cp .env.example .env
```

#### **2. Start with Docker (Recommended)**
```bash
# Start all services (app, database, pgAdmin)
docker-compose up -d

# Check service status
docker-compose ps

# View application logs
docker-compose logs -f app
```

#### **3. Verify Installation**
```bash
# Test API endpoint
curl http://localhost:8080/api/v1/orders

# Access pgAdmin (optional)
# Open http://localhost:5050
# Email: admin@frappuccino.com, Password: admin
```

#### **4. Development Workflow**
```bash
# Make code changes, then rebuild
docker-compose build --no-cache app
docker-compose restart app

# View logs
docker-compose logs -f app

# Database operations
docker-compose exec db psql -U postgres -d hotcoffee
```

### **🗂️ Sample Data**

The application comes with comprehensive sample data:
- **15 Menu Items** - Various coffee drinks, pastries, and beverages
- **20 Inventory Items** - Complete ingredient and supply tracking
- **15 Orders** - Diverse order examples with multiple items
- **36 Order Items** - Detailed order compositions
- **25 Status History Entries** - Complete order lifecycle tracking

### **📊 API Testing Examples**

#### **Get All Orders**
```bash
curl -X GET "http://localhost:8080/api/v1/orders"
```

#### **Business Analytics**
```bash
# Daily order analytics
curl -X GET "http://localhost:8080/api/v1/reports/orderedItemsByPeriod?period=day&month=august"

# Monthly reporting
curl -X GET "http://localhost:8080/api/v1/reports/orderedItemsByPeriod?period=month&year=2025"
```

#### **Inventory Management**
```bash
# Get inventory with pagination
curl -X GET "http://localhost:8080/api/v1/inventory/getLeftOvers?sortBy=quantity&page=1&pageSize=10"
```

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

# Go Project Design Patterns & Architecture Blueprint

**Version:** 1.0
**Last Updated:** 2025-11-08
**Project:** Go Gin Boilerplate - Universal Starter Kit

---

> **Note on repository examples:** many snippets below use the old global `database.DB`. The code now injects the connection: repositories are built with `NewXRepository(db *gorm.DB)`, every method takes `ctx context.Context` first, and queries start from `r.db.WithContext(ctx)`. See `internal/domain/repositories/user_repo.go` for the current pattern.

## ⚠️ FOR AI AGENTS - READ THIS FIRST

> **🚨 CRITICAL: Read [`00_AI_CRITICAL_RULES.md`](./00_AI_CRITICAL_RULES.md) BEFORE reading this document!**
>
> **This document is 2479 lines.** The critical rules file is 100 lines and contains NON-NEGOTIABLE patterns.
>
> **Skip = Fail:** The #1 mistake is not reading the struct-based patterns.

### 🔥 Critical Sections in This Document (MUST READ)

After reading `00_AI_CRITICAL_RULES.md`, focus on these sections:

- **Lines 900-948:** Struct-based Controller Pattern (MANDATORY)
- **Lines 984-1016:** Struct-based Service Pattern (MANDATORY)
- **Lines 1070-1100:** Response Utility Pattern (MANDATORY)
- **Lines 439-492:** Dependency Injection Pattern (MANDATORY)

### 📖 How to Use This Document

1. ✅ Read `00_AI_CRITICAL_RULES.md` first (100 lines)
2. ✅ Read critical sections listed above
3. ⚠️  Skim other sections for context
4. 📚 Use as reference for detailed patterns

---

## 📋 Table of Contents

> **Quick Navigation:** Use Ctrl+F with line numbers (e.g., "L900") or search keywords

### 🔥 Critical Sections (MUST READ)

| Section | Lines | Time | Keywords |
|---------|-------|------|----------|
| [6. Implementation Patterns](#6-implementation-patterns) | L919-1173 | 8 min | `struct`, `controller`, `service`, `repository` |
| → [6.1 Struct-Based Controller](#61-struct-based-controller-pattern) | L921-1005 | 3 min | `NewController`, `DI`, `struct methods` |
| → [6.2 Struct-Based Service](#62-struct-based-service-pattern) | L1006-1040 | 2 min | `NewService`, `business logic` |
| → [6.3 Function-Based Repository](#63-function-based-repository-pattern) | L1041-1092 | 2 min | `CRUD`, `GetByID`, `Create`, `Update` |
| [13. Anti-Patterns](#13-anti-patterns-to-avoid) | L2301-2460 | 4 min | `wrong`, `bad`, `avoid`, `anti-pattern` |

### 📚 All Sections

| # | Section | Lines | Keywords |
|---|---------|-------|----------|
| 1 | [Overview](#1-overview) | L53-98 | `philosophy`, `goals`, `principles` |
| 1.1 | [Architecture Philosophy](#11-architecture-philosophy) | L55-100 | `clean architecture`, `why patterns` |
| 2 | [Project Architecture](#2-project-architecture) | L99-153 | `layers`, `architecture`, `separation` |
| 2.1 | [Clean Architecture Layers](#21-clean-architecture-layers) | L101-134 | `layer`, `controller`, `service`, `repository` |
| 2.2 | [Dependency Flow Rules](#22-dependency-flow-rules) | L135-153 | `dependency`, `direction`, `flow` |
| 3 | [Core Design Patterns](#3-core-design-patterns) | L154-516 | `patterns`, `repository`, `factory`, `DI` |
| 3.1 | [Repository Pattern](#31-repository-pattern) | L156-227 | `repository`, `data access`, `CRUD` |
| 3.2 | [Service Layer Pattern](#32-service-layer-pattern) | L228-303 | `service`, `business logic`, `orchestration` |
| 3.3 | [DTO Pattern](#33-dto-data-transfer-object-pattern) | L304-368 | `DTO`, `data transfer`, `request`, `response` |
| 3.4 | [Factory Pattern](#34-factory-pattern) | L369-409 | `factory`, `New*`, `constructor` |
| 3.5 | [Middleware Pattern](#35-middleware-pattern) | L410-462 | `middleware`, `gin.HandlerFunc`, `c.Next()` |
| 3.6 | [Dependency Injection](#36-dependency-injection-pattern) | L463-516 | `DI`, `injection`, `constructor` |
| 4 | [Directory Structure](#4-directory-structure-standard) | L517-682 | `directory`, `folder`, `structure`, `tree` |
| 4.1 | [Complete Project Structure](#41-complete-project-structure) | L519-638 | `project tree`, `full structure` |
| 4.2 | [Package Organization](#42-package-organization-rules) | L639-682 | `package`, `internal`, `pkg`, `cmd` |
| 5 | [Layer Responsibilities](#5-layer-responsibilities) | L683-918 | `responsibilities`, `what`, `where` |
| 5.1 | [Controller Layer](#51-controller-layer-thin-layer) | L685-731 | `controller`, `thin`, `HTTP`, `validation` |
| 5.2 | [Service Layer](#52-service-layer-fat-layer) | L732-810 | `service`, `fat`, `business logic` |
| 5.3 | [Repository Layer](#53-repository-layer-data-layer) | L811-918 | `repository`, `database`, `CRUD`, `GORM` |
| 6 | [Implementation Patterns](#6-implementation-patterns) | L919-1173 | `how to`, `implementation`, `code` |
| 6.1 | [Struct-Based Controller](#61-struct-based-controller-pattern) | L921-1005 | `controller struct`, `NewController`, `methods` |
| 6.2 | [Struct-Based Service](#62-struct-based-service-pattern) | L1006-1040 | `service struct`, `NewService`, `methods` |
| 6.3 | [Function-Based Repository](#63-function-based-repository-pattern) | L1041-1092 | `repository functions`, `CreateUser`, `GetByID` |
| 6.4 | [Response Utility Pattern](#64-response-utility-pattern) | L1093-1173 | `utils.Ok`, `response`, `standard format` |
| 7 | [Request Flow Patterns](#7-request-flow-patterns) | L1174-1362 | `flow`, `request`, `lifecycle`, `pipeline` |
| 7.1 | [Standard CRUD Flow](#71-standard-crud-flow) | L1176-1268 | `CRUD`, `create`, `read`, `update`, `delete` |
| 7.2 | [Authentication Flow](#72-authentication-flow-pattern) | L1269-1318 | `auth`, `login`, `JWT`, `token` |
| 7.3 | [Transaction Flow](#73-transaction-flow-pattern) | L1319-1362 | `transaction`, `Begin()`, `Commit()`, `Rollback()` |
| 8 | [Data Flow Patterns](#8-data-flow-patterns) | L1363-1484 | `data`, `transformation`, `mapping` |
| 8.1 | [Request → Response Transform](#81-request--response-data-transformation) | L1365-1433 | `transform`, `DTO`, `model`, `mapping` |
| 8.2 | [DTO vs Model Usage](#82-dto-vs-model-usage) | L1434-1484 | `when`, `DTO`, `model`, `difference` |
| 9 | [Error Handling Patterns](#9-error-handling-patterns) | L1485-1613 | `error`, `handling`, `recovery`, `logging` |
| 9.1 | [Error Flow](#91-error-flow-pattern) | L1487-1516 | `error flow`, `propagation` |
| 9.2 | [Custom Error Types](#92-custom-error-types) | L1517-1562 | `custom error`, `typed error` |
| 9.3 | [Error Wrapping](#93-error-wrapping-pattern) | L1563-1613 | `%w`, `fmt.Errorf`, `wrapping` |
| 10 | [Testing Patterns](#10-testing-patterns) | L1614-1780 | `test`, `testing`, `mock`, `coverage` |
| 10.1 | [Service Layer Testing](#101-service-layer-testing-pattern) | L1616-1736 | `service test`, `mock`, `setup` |
| 10.2 | [Table-Driven Testing](#102-table-driven-testing-pattern) | L1737-1780 | `table test`, `subtests`, `t.Run` |
| 11 | [Feature Implementation Guide](#11-complete-feature-implementation-guide) | L1781-2196 | `step by step`, `guide`, `example`, `product` |
| 11.x | [Step-by-Step Product Feature](#step-by-step-adding-a-new-product-feature) | L1783-2196 | `complete example`, `full feature` |
| 12 | [Pattern Examples](#12-pattern-examples-from-codebase) | L2197-2300 | `examples`, `real code`, `reference` |
| 12.1 | [Auth Pattern Example](#121-auth-pattern-from-auth_controllergo) | L2199-2267 | `auth example`, `authentication` |
| 12.2 | [Client Management Example](#122-client-management-pattern-from-client_controllergo) | L2268-2300 | `client example`, `management` |
| 13 | [Anti-Patterns to Avoid](#13-anti-patterns-to-avoid) | L2301-2460 | `wrong`, `bad`, `avoid`, `don't` |
| 13.1 | [Business Logic in Controller](#131--business-logic-in-controller) | L2303-2347 | `controller anti-pattern`, `fat controller` |
| 13.2 | [Direct DB in Service](#132--direct-database-access-in-service) | L2348-2378 | `service anti-pattern`, `tight coupling` |
| 13.3 | [Standalone Functions](#133--standalone-controller-functions) | L2379-2413 | `standalone`, `function anti-pattern` |
| 13.4 | [God Service](#134--god-service-too-many-responsibilities) | L2414-2440 | `god service`, `SRP violation` |
| 13.5 | [Circular Dependencies](#135--circular-dependencies) | L2441-2460 | `import cycle`, `circular` |

### 🎯 Quick Lookups by Task

**Implementing Controllers:**
- L921-1005: Struct-Based Controller Pattern (CRITICAL)
- L685-731: Controller Responsibilities
- L2303-2347: What NOT to do

**Implementing Services:**
- L1006-1040: Struct-Based Service Pattern (CRITICAL)
- L732-810: Service Responsibilities
- L2348-2378: What NOT to do

**Implementing Repositories:**
- L1041-1092: Function-Based Repository Pattern
- L811-918: Repository Responsibilities
- L156-227: Repository Pattern Theory

**Understanding Flow:**
- L1176-1268: Standard CRUD Flow
- L1269-1318: Authentication Flow
- L1319-1362: Transaction Flow

**Complete Example:**
- L1783-2196: Step-by-Step Product Feature (Full Implementation)

---

## 1. OVERVIEW

### 1.1 Architecture Philosophy

This project follows **Clean Architecture** principles with a **Layered Architecture** implementation:

```
┌─────────────────────────────────────────────┐
│          External Systems / HTTP            │
└──────────────────┬──────────────────────────┘
                   │
┌──────────────────▼──────────────────────────┐
│         Controllers (Thin Layer)            │  ← HTTP Handlers
│  • Parse requests                           │
│  • Call services                            │
│  • Return responses                         │
└──────────────────┬──────────────────────────┘
                   │
┌──────────────────▼──────────────────────────┐
│         Services (Fat Layer)                │  ← Business Logic
│  • Business rules                           │
│  • Validation                               │
│  • Orchestration                            │
│  • External API calls                       │
└──────────────────┬──────────────────────────┘
                   │
┌──────────────────▼──────────────────────────┐
│         Repositories (Data Layer)           │  ← Data Access
│  • CRUD operations                          │
│  • Database queries                         │
│  • Transaction management                   │
└──────────────────┬──────────────────────────┘
                   │
┌──────────────────▼──────────────────────────┐
│              Database / Models              │  ← Data Storage
└─────────────────────────────────────────────┘
```

**Key Principles:**
1. **Dependency Inversion** - High-level modules don't depend on low-level modules
2. **Separation of Concerns** - Each layer has ONE responsibility
3. **Dependency Direction** - Dependencies flow INWARD only
4. **Testability** - Each layer can be tested independently

---

## 2. PROJECT ARCHITECTURE

### 2.1 Clean Architecture Layers

```
┌──────────────────────────────────────────────────────┐
│                  PRESENTATION LAYER                  │
│  • Controllers (HTTP handlers)                       │
│  • Middlewares (Cross-cutting concerns)             │
│  • Routers (Route definitions)                       │
│  • DTOs (Request/Response structures)                │
└──────────────────────────────────────────────────────┘
                         ↓
┌──────────────────────────────────────────────────────┐
│                 APPLICATION LAYER                     │
│  • Services (Business logic)                         │
│  • Use Cases (Application workflows)                 │
│  • Orchestration (Coordinating multiple operations)  │
└──────────────────────────────────────────────────────┘
                         ↓
┌──────────────────────────────────────────────────────┐
│                   DOMAIN LAYER                        │
│  • Models (Domain entities)                          │
│  • Repositories (Data access interfaces)             │
│  • Domain logic (Pure business rules)                │
└──────────────────────────────────────────────────────┘
                         ↓
┌──────────────────────────────────────────────────────┐
│              INFRASTRUCTURE LAYER                     │
│  • Database adapters                                 │
│  • External API clients                              │
│  • File system access                                │
│  • Third-party integrations                          │
└──────────────────────────────────────────────────────┘
```

### 2.2 Dependency Flow Rules

**✅ ALLOWED:**
```
Controllers → Services → Repositories → Database
     ↓            ↓            ↓
   DTOs       Utils/Enums   Models
```

**❌ FORBIDDEN:**
```
Services → Controllers      // Services cannot depend on HTTP layer
Repositories → Services     // Repositories cannot depend on business logic
Models → Repositories       // Models should be pure data structures
pkg/enums → internal/*      // Public packages cannot import internal
```

---

## 3. CORE DESIGN PATTERNS

### 3.1 Repository Pattern

**Purpose:** Abstract data access logic from business logic.

**Structure:**
```go
// 1. Define interface in domain layer
// File: internal/domain/repositories/user_repository.go
package repositories

type UserRepository interface {
    Create(user *models.User) error
    FindByID(id uint) (*models.User, error)
    FindByEmail(email string) (*models.User, error)
    Update(user *models.User) error
    Delete(id uint) error
    List(page, pageSize int) ([]*models.User, int64, error)
}

// 2. Implement in infrastructure layer
// File: internal/adapters/database/user_repo_impl.go (OR just use functions)
package repositories

import "gorm.io/gorm"

// Option A: Function-based (simpler, used in this project)
func CreateUser(user *models.User) error {
    if err := database.DB.Create(user).Error; err != nil {
        logger.Errorf("failed to create user: %v", err)
        return err
    }
    return nil
}

func GetUserByID(id uint) (*models.User, error) {
    var user models.User
    if err := database.DB.Where("id = ?", id).First(&user).Error; err != nil {
        logger.Errorf("failed to get user %d: %v", id, err)
        return nil, err
    }
    return &user, nil
}

// Option B: Struct-based (more complex, better for mocking)
type userRepoImpl struct {
    db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
    return &userRepoImpl{db: db}
}

func (r *userRepoImpl) Create(user *models.User) error {
    return r.db.Create(user).Error
}
```

**✅ DO:**
- Keep repositories focused on data access ONLY
- Use GORM or parameterized queries (never string concatenation)
- Return domain models, not DTOs
- Handle database-specific errors here
- Use transactions for multi-step operations

**❌ DON'T:**
- Put business logic in repositories
- Call other repositories directly
- Import service layer packages
- Transform data for API responses (that's DTOs job)

---

### 3.2 Service Layer Pattern

**Purpose:** Encapsulate business logic and orchestrate operations.

**Structure:**
```go
// File: internal/app/services/user_service.go
package services

type UserService struct {
    // Dependencies injected via constructor
}

func NewUserService() *UserService {
    return &UserService{}
}

// Public methods implement business operations
func (s *UserService) CreateUser(req *dto.CreateUserRequest) error {
    // 1. Validate business rules
    if err := s.validateUser(req); err != nil {
        return fmt.Errorf("validation failed: %w", err)
    }

    // 2. Check business constraints
    existing, _ := repositories.GetUserByEmail(req.Email)
    if existing != nil {
        return fmt.Errorf("email already exists")
    }

    // 3. Hash password
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
        return fmt.Errorf("failed to hash password: %w", err)
    }

    // 4. Build domain model
    user := &models.User{
        Name:     req.Name,
        Email:    req.Email,
        Password: string(hashedPassword),
    }

    // 5. Save via repository
    if err := repositories.CreateUser(user); err != nil {
        logger.Errorf("failed to create user: %v", err)
        return fmt.Errorf("failed to create user: %w", err)
    }

    logger.Infof("User created successfully: %s", user.Email)
    return nil
}

// Private helper methods
func (s *UserService) validateUser(req *dto.CreateUserRequest) error {
    // Validation logic
    return nil
}
```

**✅ DO:**
- Implement ALL business logic here
- Validate input according to business rules
- Orchestrate multiple repository calls
- Handle external API calls
- Transform data between layers
- Log important business events

**❌ DON'T:**
- Handle HTTP concerns (gin.Context, HTTP status codes)
- Access database directly (use repositories)
- Import controller packages
- Return HTTP responses

---

### 3.3 DTO (Data Transfer Object) Pattern

**Purpose:** Define contracts for API requests/responses, prevent tight coupling.

**Structure:**
```go
// File: internal/app/dto/user_dto.go
package dto

// CreateUserRequest represents user registration payload
type CreateUserRequest struct {
    // Name is the full name of the user
    Name string `json:"name" binding:"required,min=3,max=255"`

    // Email must be unique (validated at service layer)
    Email string `json:"email" binding:"required,email"`

    // Password will be hashed before storage
    Password string `json:"password" binding:"required,min=8"`

    // RoleID references the role table
    RoleID uint `json:"role_id" binding:"required"`
}

// UserResponse represents user data returned to client
type UserResponse struct {
    ID        uint      `json:"id"`
    Name      string    `json:"name"`
    Email     string    `json:"email"`
    RoleID    uint      `json:"role_id"`
    Role      string    `json:"role"`
    CreatedAt time.Time `json:"created_at"`
}

// UpdateUserRequest represents user update payload
type UpdateUserRequest struct {
    Name  string `json:"name" binding:"omitempty,min=3,max=255"`
    Email string `json:"email" binding:"omitempty,email"`
}

// ListUsersResponse represents paginated user list
type ListUsersResponse struct {
    Users      []UserResponse `json:"users"`
    Page       int            `json:"page"`
    PageSize   int            `json:"page_size"`
    Total      int64          `json:"total"`
    TotalPages int            `json:"total_pages"`
}
```

**✅ DO:**
- Define separate DTOs for Request and Response
- Use struct tags for validation (`binding:`, `validate:`)
- Add comments explaining each field
- Keep DTOs in `internal/app/dto/` directory
- Use DTOs only for API layer (not internal logic)

**❌ DON'T:**
- Expose database models directly via API
- Add business logic to DTOs (they're pure data structures)
- Reuse request DTOs as response DTOs
- Put DTOs in service or repository packages

---

### 3.4 Factory Pattern

**Purpose:** Centralize object creation and initialization.

**Structure:**
```go
// Constructor pattern for services
func NewUserService() *UserService {
    return &UserService{
        // Initialize with dependencies
    }
}

// Constructor pattern for controllers
func NewUserController(service *UserService) *UserController {
    return &UserController{
        service: service,
    }
}

// Constructor pattern for repositories (if using struct-based)
func NewUserRepository(db *gorm.DB) UserRepository {
    return &userRepoImpl{
        db: db,
    }
}
```

**✅ DO:**
- Use `New*` functions for all constructors
- Initialize dependencies via constructor
- Return concrete types (not interfaces) from controllers/services
- Return interfaces from repository constructors (for mocking)

**❌ DON'T:**
- Create objects with `&Struct{}` directly in business logic
- Use `init()` functions for dependency initialization
- Create global instances

---

### 3.5 Middleware Pattern

**Purpose:** Handle cross-cutting concerns (auth, logging, etc.).

**Structure:**
```go
// File: internal/app/middlewares/auth_middleware.go
package middlewares

func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. Extract token
        token := c.GetHeader("Authorization")
        if token == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing token"})
            c.Abort()
            return
        }

        // 2. Validate token
        claims, err := utils.ValidateJWT(token)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
            c.Abort()
            return
        }

        // 3. Set user context
        c.Set("user_id", claims.UserID)
        c.Set("email", claims.Email)

        // 4. Continue to next handler
        c.Next()
    }
}

// Usage in router
route.Use(middlewares.AuthMiddleware())
route.GET("/api/users", controllers.ListUsers)
```

**Common Middleware Types:**
- **Authentication** - Verify JWT tokens
- **Authorization** - Check user permissions
- **Rate Limiting** - Prevent abuse
- **Logging** - Log requests/responses
- **CORS** - Handle cross-origin requests
- **Recovery** - Catch panics
- **Validation** - Validate requests
- **IP Whitelist** - Restrict access by IP

---

### 3.6 Dependency Injection Pattern

**Purpose:** Decouple components and improve testability.

**Structure:**
```go
// 1. Define dependencies in struct
type UserService struct {
    userRepo    repositories.UserRepository
    emailClient *EmailClient
    logger      *Logger
}

// 2. Inject via constructor
func NewUserService(
    userRepo repositories.UserRepository,
    emailClient *EmailClient,
    logger *Logger,
) *UserService {
    return &UserService{
        userRepo:    userRepo,
        emailClient: emailClient,
        logger:      logger,
    }
}

// 3. Use in router registration
func RegisterRoutes(router *gin.Engine) {
    // Initialize dependencies
    db := database.GetDB()
    userRepo := repositories.NewUserRepository(db)
    emailClient := NewEmailClient()
    logger := logger.NewLogger()

    // Inject into service
    userService := services.NewUserService(userRepo, emailClient, logger)

    // Inject into controller
    userController := controllers.NewUserController(userService)

    // Register routes
    router.GET("/api/users", userController.List)
    router.POST("/api/users", userController.Create)
}
```

**✅ Benefits:**
- Easy to test (inject mocks)
- Easy to change implementations
- Clear dependencies
- No global state

---

## 4. DIRECTORY STRUCTURE STANDARD

### 4.1 Complete Project Structure

```
project-name/
│
├── cmd/                                # Application entry points
│   ├── migrate/                        # Database migration tool
│   │   └── main.go
│   └── seeder/                         # Database seeder tool
│       └── seeder.go
│
├── internal/                           # Private application code
│   │
│   ├── adapters/                       # External adapters
│   │   └── database/                   # Database adapter
│   │       ├── db.go                   # Connection management
│   │       ├── migrations/             # SQL migration files
│   │       │   ├── 001_create_users.sql
│   │       │   └── 002_create_clients.sql
│   │       └── seeders/                # Seed data
│   │           └── users_seeder.go
│   │
│   ├── app/                            # Application layer
│   │   │
│   │   ├── controllers/                # HTTP handlers (thin)
│   │   │   ├── auth/                   # Auth controller (grouped)
│   │   │   │   └── auth_controller.go
│   │   │   ├── user_controller.go      # User CRUD
│   │   │   ├── client_controller.go    # Client CRUD
│   │   │   └── transaction_controller.go
│   │   │
│   │   ├── dto/                        # Data Transfer Objects
│   │   │   ├── auth_dto.go             # Auth requests/responses
│   │   │   ├── user_dto.go             # User requests/responses
│   │   │   ├── client_dto.go           # Client requests/responses
│   │   │   └── transaction_dto.go
│   │   │
│   │   ├── middlewares/                # HTTP middlewares
│   │   │   ├── auth_middleware.go      # JWT authentication
│   │   │   ├── rate_limit_middleware.go
│   │   │   ├── ip_whitelist_middleware.go
│   │   │   ├── activity_logger_middleware.go
│   │   │   ├── cors_middleware.go
│   │   │   └── recovery_middleware.go
│   │   │
│   │   ├── routers/                    # Route definitions
│   │   │   └── index.go                # Main router
│   │   │
│   │   └── services/                   # Business logic (fat)
│   │       ├── auth_service.go         # Authentication logic
│   │       ├── user_service.go         # User business logic
│   │       ├── order_service.go        # Order processing logic
│   │       ├── email_service.go        # Email sending
│   │       ├── payment_service.go      # Payment processing
│   │       └── notification_service.go # Push notifications
│   │
│   └── domain/                         # Domain layer
│       │
│       ├── models/                     # Database entities
│       │   ├── user.go                 # User model
│       │   ├── client.go               # Client model
│       │   ├── transaction.go          # Transaction model
│       │   ├── role.go                 # Role model
│       │   └── otp_session.go          # OTP session model
│       │
│       └── repositories/               # Data access layer
│           ├── auth_repo.go            # Auth-specific queries
│           ├── user_repo.go            # User CRUD operations
│           ├── client_repo.go          # Client CRUD operations
│           └── transaction_repo.go     # Transaction CRUD operations
│
├── pkg/                                # Public reusable packages
│   ├── config/                         # Configuration loader
│   │   └── config.go
│   │
│   ├── datatable/                      # DataTable utilities
│   │   └── datatable.go
│   │
│   ├── enums/                          # Constants and enums
│   │   ├── roles.go                    # Role constants
│   │   ├── statuses.go                 # Status constants
│   │   └── providers.go                # Provider constants
│   │
│   ├── logger/                         # Logging utilities
│   │   └── logger.go
│   │
│   ├── types/                          # Shared types
│   │   ├── errors.go                   # Custom error types
│   │   └── common.go                   # Common types
│   │
│   └── utils/                          # Helper functions
│       ├── jwt.go                      # JWT utilities
│       ├── hash.go                     # Password hashing
│       ├── response.go                 # HTTP response helpers
│       └── validation.go               # Validation helpers
│
├── docs/                               # Documentation
│   ├── DESIGN_PATTERNS.md              # This file
│   ├── CODING_STANDARDS.md             # Coding standards
│   ├── AI_AGENT_RULES.md               # AI agent rules
│   └── API_DOCUMENTATION.md            # API docs
│
├── tests/                              # Test files
│   ├── unit/                           # Unit tests
│   ├── integration/                    # Integration tests
│   └── fixtures/                       # Test data
│
├── .env.example                        # Environment template
├── .env                                # Local environment (gitignored)
├── .gitignore                          # Git ignore rules
├── docker-compose-dev.yml              # Docker dev config
├── docker-compose-prod.yml             # Docker prod config
├── Dockerfile                          # Docker image
├── go.mod                              # Go dependencies
├── go.sum                              # Dependency checksums
├── main.go                             # Application entry point
├── Makefile                            # Build automation
└── README.md                           # Project documentation
```

### 4.2 Package Organization Rules

**✅ RULES:**

1. **One package per directory**
   ```
   ✅ internal/app/services/   → package services
   ❌ internal/app/           → package users, package clients (multiple packages)
   ```

2. **Group multi-file controllers/services into a feature folder**
   ```
   ✅ CORRECT:
   internal/app/controllers/auth/
       ├── auth_controller.go
       ├── auth_controller_handlers.go
       └── auth_controller_validators.go

   internal/app/services/user/
       ├── user_service.go
       └── user_service_rules.go

   ❌ WRONG:
   internal/app/controllers/
       ├── auth_controller.go
       ├── auth_controller_handlers.go
       └── user_controller.go            // Mixed features at the same directory level
   ```

2. **Package naming:**
   - Lowercase, single word
   - No underscores, no dashes
   - Plural for collections: `services`, `controllers`, `repositories`
   - Singular for utilities: `logger`, `config`, `utils`

3. **Import paths:**
   ```go
   import (
       "github.com/your-org/project-name/internal/app/services"
       "github.com/your-org/project-name/internal/domain/models"
       "github.com/your-org/project-name/pkg/utils"
   )
   ```

4. **Grouping imports:**
   ```go
   import (
       // Standard library
       "fmt"
       "time"

       // External dependencies
       "github.com/gin-gonic/gin"
       "gorm.io/gorm"

       // Internal packages
       "github.com/your-org/project-name/internal/app/dto"
       "github.com/your-org/project-name/pkg/utils"
   )
   ```

---

## 5. LAYER RESPONSIBILITIES

### 5.1 Controller Layer (Thin Layer)

**Responsibility:** Handle HTTP concerns ONLY.

**What Controllers SHOULD Do:**
```go
func (ctrl *UserController) Create(c *gin.Context) {
    // 1. Parse and bind request
    var req dto.CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        utils.BadRequest(c, err, "Invalid request")
        return
    }

    // 2. Call service
    if err := ctrl.service.CreateUser(&req); err != nil {
        utils.InternalServerError(c, err, "Failed to create user")
        return
    }

    // 3. Return response
    utils.Created(c, nil, "User created successfully")
}
```

**✅ Controllers SHOULD:**
- Parse request parameters (URL params, query params, body)
- Validate request format (binding, JSON parsing)
- Call service methods
- Format HTTP responses
- Handle HTTP status codes
- Use response utility functions (`utils.Ok`, `utils.BadRequest`, etc.)

**❌ Controllers MUST NOT:**
- Contain business logic
- Access database directly
- Transform business data
- Call repositories directly
- Have functions >50 lines
- Import repository packages

**Size Limits:**
- **Maximum 50 lines** per controller function
- **Maximum 300 lines** per controller file

---

### 5.2 Service Layer (Fat Layer)

**Responsibility:** Implement ALL business logic.

**What Services SHOULD Do:**
```go
func (s *UserService) CreateUser(req *dto.CreateUserRequest) error {
    // 1. Business validation
    if err := s.validateBusinessRules(req); err != nil {
        return fmt.Errorf("validation failed: %w", err)
    }

    // 2. Check business constraints
    existing, _ := repositories.GetUserByEmail(req.Email)
    if existing != nil {
        return fmt.Errorf("email already exists")
    }

    // 3. Apply business logic
    hashedPassword, err := bcrypt.GenerateFromPassword(
        []byte(req.Password),
        bcrypt.DefaultCost,
    )
    if err != nil {
        return fmt.Errorf("failed to hash password: %w", err)
    }

    // 4. Build domain model
    user := &models.User{
        Name:     req.Name,
        Email:    req.Email,
        Password: string(hashedPassword),
        RoleID:   req.RoleID,
    }

    // 5. Orchestrate repository calls
    if err := repositories.CreateUser(user); err != nil {
        logger.Errorf("failed to create user: %v", err)
        return fmt.Errorf("failed to create user: %w", err)
    }

    // 6. Trigger side effects (emails, webhooks, etc.)
    go s.sendWelcomeEmail(user.Email)

    // 7. Log business events
    logger.Infof("User created: ID=%d, Email=%s", user.ID, user.Email)

    return nil
}
```

**✅ Services SHOULD:**
- Implement ALL business logic and rules
- Validate business constraints
- Orchestrate multiple repository calls
- Call external APIs
- Transform data between layers
- Handle business errors gracefully
- Log important business events
- Trigger side effects (emails, webhooks, notifications)
- Use transactions for multi-step operations

**❌ Services MUST NOT:**
- Handle HTTP concerns (`gin.Context`, status codes)
- Access database directly (use repositories)
- Import controller packages
- Return HTTP responses
- Have knowledge of request/response formats

**Size Limits:**
- **Maximum 100 lines** per service function
- **Maximum 400 lines** per service file
- If exceeding, split into multiple files:
  - `user_service.go` - main logic
  - `user_service_helpers.go` - helper functions
  - `user_service_validators.go` - validation logic

---

### 5.3 Repository Layer (Data Layer)

**Responsibility:** Handle data persistence ONLY.

**What Repositories SHOULD Do:**
```go
// CRUD operations
func CreateUser(user *models.User) error {
    if err := database.DB.Create(user).Error; err != nil {
        logger.Errorf("failed to create user: %v", err)
        return err
    }
    return nil
}

func GetUserByID(id uint) (*models.User, error) {
    var user models.User
    if err := database.DB.Where("id = ?", id).
        Preload("UserRoles.Role").
        Preload("UserRoles.Client").
        First(&user).Error; err != nil {
        logger.Errorf("failed to get user %d: %v", id, err)
        return nil, err
    }
    return &user, nil
}

func UpdateUser(user *models.User) error {
    if err := database.DB.Save(user).Error; err != nil {
        logger.Errorf("failed to update user: %v", err)
        return err
    }
    return nil
}

// Complex queries with optimization
func ListUsers(page, pageSize int) ([]*models.User, int64, error) {
    var users []*models.User
    var total int64

    // Count total
    database.DB.Model(&models.User{}).Count(&total)

    // Get paginated results
    offset := (page - 1) * pageSize
    if err := database.DB.
        Preload("UserRoles.Role").
        Offset(offset).
        Limit(pageSize).
        Find(&users).Error; err != nil {
        return nil, 0, err
    }

    return users, total, nil
}

// Transaction example
func CreateUserWithRole(user *models.User, roleID uint, clientID uint) error {
    tx := database.DB.Begin()
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
        }
    }()

    // Step 1: Create user
    if err := tx.Create(user).Error; err != nil {
        tx.Rollback()
        return err
    }

    // Step 2: Create user role
    userRole := models.UserRole{
        UserID:   user.ID,
        RoleID:   roleID,
        ClientID: clientID,
    }
    if err := tx.Create(&userRole).Error; err != nil {
        tx.Rollback()
        return err
    }

    return tx.Commit().Error
}
```

**✅ Repositories SHOULD:**
- Perform CRUD operations
- Build database queries
- Use GORM or parameterized queries
- Handle database-specific errors
- Use transactions for multi-step operations
- Optimize queries (use Preload, Select, etc.)
- Return domain models

**❌ Repositories MUST NOT:**
- Contain business logic
- Validate business rules
- Call other repositories
- Import service packages
- Transform data for API responses

**Size Limits:**
- **Maximum 30 lines** per repository function
- **Maximum 300 lines** per repository file

---

## 6. IMPLEMENTATION PATTERNS

### 6.1 Struct-Based Controller Pattern

**✅ REQUIRED Pattern:**
```go
// File: internal/app/controllers/user_controller.go
package controllers

import (
    "github.com/gin-gonic/gin"
    "github.com/your-org/project/internal/app/dto"
    "github.com/your-org/project/internal/app/services"
    "github.com/your-org/project/pkg/utils"
)

// UserController handles user-related HTTP requests
type UserController struct {
    service *services.UserService
}

// NewUserController creates a new UserController instance
func NewUserController(service *services.UserService) *UserController {
    return &UserController{
        service: service,
    }
}

// List handles GET /api/users
func (ctrl *UserController) List(c *gin.Context) {
    // Implementation
}

// Get handles GET /api/users/:id
func (ctrl *UserController) Get(c *gin.Context) {
    // Implementation
}

// Create handles POST /api/users
func (ctrl *UserController) Create(c *gin.Context) {
    // Implementation
}

// Update handles PUT /api/users/:id
func (ctrl *UserController) Update(c *gin.Context) {
    // Implementation
}

// Delete handles DELETE /api/users/:id
func (ctrl *UserController) Delete(c *gin.Context) {
    // Implementation
}
```

**❌ WRONG Pattern (Standalone Functions):**
```go
// DON'T DO THIS - Old pattern
func CreateUser(c *gin.Context) {
    // Direct implementation without struct
}
```

**Router Registration:**
```go
// File: internal/app/routers/index.go
func RegisterRoutes(router *gin.Engine) {
    // Initialize dependencies
    userService := services.NewUserService()
    userController := controllers.NewUserController(userService)

    // Register routes with controller methods
    api := router.Group("/api")
    {
        users := api.Group("/users")
        {
            users.GET("", userController.List)
            users.GET("/:id", userController.Get)
            users.POST("", userController.Create)
            users.PUT("/:id", userController.Update)
            users.DELETE("/:id", userController.Delete)
        }
    }
}
```

---

### 6.2 Struct-Based Service Pattern

**✅ REQUIRED Pattern:**
```go
// File: internal/app/services/user_service.go
package services

type UserService struct {
    // Dependencies (if needed)
}

func NewUserService() *UserService {
    return &UserService{}
}

// Public methods - implement business operations
func (s *UserService) CreateUser(req *dto.CreateUserRequest) error {
    // Business logic
    return nil
}

func (s *UserService) GetUserByID(id uint) (*models.User, error) {
    // Business logic
    return nil, nil
}

// Private helper methods
func (s *UserService) validateUser(req *dto.CreateUserRequest) error {
    // Validation logic
    return nil
}
```

---

### 6.3 Function-Based Repository Pattern

**✅ CURRENT Pattern (Simple, Direct):**
```go
// File: internal/domain/repositories/user_repo.go
package repositories

// Direct functions for data access
func CreateUser(user *models.User) error {
    return database.DB.Create(user).Error
}

func GetUserByID(id uint) (*models.User, error) {
    var user models.User
    err := database.DB.Where("id = ?", id).First(&user).Error
    return &user, err
}

func GetUserByEmail(email string) (*models.User, error) {
    var user models.User
    err := database.DB.Where("email = ?", email).First(&user).Error
    return &user, err
}

func UpdateUser(user *models.User) error {
    return database.DB.Save(user).Error
}

func DeleteUser(id uint) error {
    return database.DB.Delete(&models.User{}, id).Error
}
```

**Alternative: Interface-Based Repository (For Testing):**
```go
// If you need mocking for tests
type UserRepository interface {
    Create(user *models.User) error
    GetByID(id uint) (*models.User, error)
}

type userRepoImpl struct {
    db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
    return &userRepoImpl{db: db}
}
```

---

### 6.4 Response Utility Pattern

**✅ MANDATORY: Always Use Response Utilities**

```go
// File: internal/app/controllers/user_controller.go

import "github.com/your-org/project/pkg/utils"

func (ctrl *UserController) Create(c *gin.Context) {
    var req dto.CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        // ✅ Use utility for validation errors
        utils.BadRequest(c, err, "Validation failed")
        return
    }

    user, err := ctrl.service.CreateUser(&req)
    if err != nil {
        // ✅ Use utility for errors
        utils.InternalServerError(c, err, "Failed to create user")
        return
    }

    // ✅ Use utility for success
    utils.Created(c, user, "User created successfully")
}

func (ctrl *UserController) Get(c *gin.Context) {
    id := c.Param("id")
    user, err := ctrl.service.GetUserByID(id)
    if err != nil {
        // ✅ Use utility for not found
        utils.NotFound(c, err, "User not found")
        return
    }

    // ✅ Use utility for success
    utils.Ok(c, user, "User retrieved successfully")
}
```

**Available Response Functions:**
```go
// Success responses
utils.Ok(c, data, message)              // 200 OK
utils.Created(c, data, message)         // 201 Created
utils.NoContent(c)                      // 204 No Content

// Error responses
utils.BadRequest(c, err, message)       // 400 Bad Request
utils.Unauthorized(c, err, message)     // 401 Unauthorized
utils.Forbidden(c, err, message)        // 403 Forbidden
utils.NotFound(c, err, message)         // 404 Not Found
utils.Conflict(c, err, message)         // 409 Conflict
utils.UnprocessableEntity(c, err, msg)  // 422 Unprocessable
utils.TooManyRequests(c, err, message)  // 429 Too Many Requests
utils.InternalServerError(c, err, msg)  // 500 Internal Server Error
utils.BadGateway(c, err, message)       // 502 Bad Gateway

// Generic
utils.HandleSuccess(c, statusCode, data, message)
utils.HandleErrors(c, statusCode, err, message)
```

**Response Format (Consistent):**
```json
{
    "success": true,
    "message": "User created successfully",
    "data": {
        "id": 1,
        "name": "John Doe",
        "email": "john@example.com"
    },
    "errors": null
}
```

---

## 7. REQUEST FLOW PATTERNS

### 7.1 Standard CRUD Flow

**Complete Request Flow for CREATE Operation:**

```
1. HTTP POST /api/users
   Body: {"name": "John", "email": "john@example.com", "password": "secret"}
       ↓
2. [Middleware Stack]
   • CORS Middleware
   • Rate Limit Middleware
   • Auth Middleware (extract user from JWT)
   • Activity Logger Middleware
       ↓
3. [Router]
   Match route: POST /api/users → userController.Create
       ↓
4. [Controller] - user_controller.go
   func (ctrl *UserController) Create(c *gin.Context) {
       // Parse request
       var req dto.CreateUserRequest
       if err := c.ShouldBindJSON(&req); err != nil {
           utils.BadRequest(c, err, "Invalid input")
           return
       }

       // Call service
       user, err := ctrl.service.CreateUser(&req)
       if err != nil {
           utils.InternalServerError(c, err, "Failed to create")
           return
       }

       // Return response
       utils.Created(c, user, "User created")
   }
       ↓
5. [Service] - user_service.go
   func (s *UserService) CreateUser(req *dto.CreateUserRequest) (*models.User, error) {
       // 1. Validate business rules
       if err := s.validate(req); err != nil {
           return nil, err
       }

       // 2. Check duplicates
       existing, _ := repositories.GetUserByEmail(req.Email)
       if existing != nil {
           return nil, fmt.Errorf("email exists")
       }

       // 3. Hash password
       hashed, _ := bcrypt.GenerateFromPassword([]byte(req.Password))

       // 4. Build model
       user := &models.User{
           Name:     req.Name,
           Email:    req.Email,
           Password: string(hashed),
       }

       // 5. Save to DB
       if err := repositories.CreateUser(user); err != nil {
           return nil, err
       }

       return user, nil
   }
       ↓
6. [Repository] - user_repo.go
   func CreateUser(user *models.User) error {
       return database.DB.Create(user).Error
   }
       ↓
7. [Database]
   INSERT INTO users (name, email, password) VALUES (?, ?, ?)
       ↓
8. [Response Back Up The Stack]
   Repository → Service → Controller → Middleware → HTTP Response
       ↓
9. HTTP 201 Created
   Body: {
       "success": true,
       "message": "User created successfully",
       "data": {
           "id": 1,
           "name": "John",
           "email": "john@example.com"
       }
   }
```

---

### 7.2 Authentication Flow Pattern

```
1. POST /login
   Body: {"email": "user@example.com", "password": "secret"}
       ↓
2. [AuthController] → authService.Login(req)
       ↓
3. [AuthService]
   • Get user by email (via repository)
   • Compare password hash
   • Generate OTP code
   • Create OTP session (via repository)
   • Send OTP email (background job)
   • Return OTP token
       ↓
4. Response: {"otpToken": "abc123", "role": "Admin"}
       ↓
5. POST /verify-otp
   Body: {"otpToken": "abc123", "code": "123456"}
       ↓
6. [AuthController] → authService.VerifyOTP(req)
       ↓
7. [AuthService]
   • Get OTP session by token (via repository)
   • Check expiration
   • Validate code
   • Get user by ID (via repository)
   • Generate JWT token
   • Delete OTP session (via repository)
   • Return JWT + user data
       ↓
8. Response: {"token": "eyJhbGc...", "user": {...}}
       ↓
9. Subsequent Requests with JWT:
   GET /api/users
   Header: Authorization: Bearer eyJhbGc...
       ↓
10. [AuthMiddleware]
    • Extract token from header
    • Validate JWT
    • Extract claims (user_id, email, role)
    • Set in context: c.Set("user_id", id)
    • Call c.Next()
       ↓
11. [Controller] can access: c.GetUint("user_id")
```

---

### 7.3 Transaction Flow Pattern

**Multi-Step Operation with Database Transaction:**

```go
func (s *TransactionService) ProcessPayment(req *dto.PaymentRequest) error {
    // Use GORM transaction
    return database.DB.Transaction(func(tx *gorm.DB) error {
        // Step 1: Create transaction record
        transaction := &models.Transaction{
            ClientID: req.ClientID,
            Amount:   req.Amount,
            Status:   "pending",
        }
        if err := tx.Create(transaction).Error; err != nil {
            return fmt.Errorf("create transaction: %w", err)
        }

        // Step 2: Deduct client balance
        if err := tx.Model(&models.Client{}).
            Where("id = ?", req.ClientID).
            Update("balance", gorm.Expr("balance - ?", req.Amount)).
            Error; err != nil {
            return fmt.Errorf("update balance: %w", err)
        }

        // Step 3: Create transaction log
        log := &models.TransactionLog{
            TransactionID: transaction.ID,
            Action:        "payment_processed",
        }
        if err := tx.Create(log).Error; err != nil {
            return fmt.Errorf("create log: %w", err)
        }

        // All steps succeeded - commit
        return nil
    })
    // If any step fails, all changes are rolled back automatically
}
```

---

## 8. DATA FLOW PATTERNS

### 8.1 Request → Response Data Transformation

```
HTTP Request (JSON)
       ↓
[Controller] Parse to DTO
       ↓
dto.CreateUserRequest {
    Name:     "John Doe"
    Email:    "john@example.com"
    Password: "plaintext"
}
       ↓
[Service] Transform to Domain Model
       ↓
models.User {
    Name:     "John Doe"
    Email:    "john@example.com"
    Password: "$2a$10$hashed..." (bcrypt)
    RoleID:   2
}
       ↓
[Repository] Save to Database
       ↓
Database Row:
    id=1, name="John Doe", email="john@...", password="$2a$10$..."
       ↓
[Repository] Return Domain Model
       ↓
models.User {
    ID:       1
    Name:     "John Doe"
    Email:    "john@example.com"
    Password: "$2a$10$hashed..."
    CreatedAt: 2025-11-08T10:30:00Z
}
       ↓
[Service] Return Model (or transform to DTO if needed)
       ↓
[Controller] Transform to Response DTO
       ↓
dto.UserResponse {
    ID:    1
    Name:  "John Doe"
    Email: "john@example.com"
    // Password NOT included in response
}
       ↓
HTTP Response (JSON)
{
    "success": true,
    "message": "User created",
    "data": {
        "id": 1,
        "name": "John Doe",
        "email": "john@example.com"
    }
}
```

**Key Transformations:**
1. **Controller**: JSON → DTO
2. **Service**: DTO → Domain Model (apply business logic)
3. **Repository**: Domain Model ↔ Database
4. **Service**: Domain Model → Domain Model (or DTO)
5. **Controller**: Domain Model/DTO → Response JSON

---

### 8.2 DTO vs Model Usage

**✅ When to Use DTOs:**
- API request payloads
- API response payloads
- External API communication
- Data validation at boundaries

**✅ When to Use Models:**
- Internal business logic
- Database operations
- Domain rules enforcement
- Repository layer

**Example:**
```go
// DTO for API (external boundary)
type CreateUserRequest struct {
    Name     string `json:"name" binding:"required"`
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=8"`
}

// Model for domain (internal)
type User struct {
    ID        uint      `gorm:"primarykey"`
    Name      string    `gorm:"size:255;not null"`
    Email     string    `gorm:"size:255;unique;not null"`
    Password  string    `gorm:"size:255;not null"` // Hashed
    RoleID    uint      `gorm:"not null"`
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`

    // Relationships
    UserRoles *UserRole `gorm:"foreignKey:UserID"`
}

// DTO for API response
type UserResponse struct {
    ID        uint      `json:"id"`
    Name      string    `json:"name"`
    Email     string    `json:"email"`
    Role      string    `json:"role"`
    CreatedAt time.Time `json:"created_at"`
    // Password NEVER included in response!
}
```

---

## 9. ERROR HANDLING PATTERNS

### 9.1 Error Flow Pattern

```
[Repository] Database Error
       ↓
    err = "record not found"
       ↓
[Repository] Wrap with context
       ↓
    return fmt.Errorf("failed to get user %d: %w", id, err)
       ↓
[Service] Catch error, add business context
       ↓
    return fmt.Errorf("user retrieval failed: %w", err)
       ↓
[Controller] Handle error, return appropriate HTTP status
       ↓
    if err != nil {
        utils.NotFound(c, err, "User not found")
        return
    }
       ↓
HTTP Response: 404 Not Found
{
    "success": false,
    "message": "User not found",
    "errors": {...}
}
```

### 9.2 Custom Error Types

```go
// File: pkg/types/errors.go
package types

import "errors"

// Business errors
var (
    ErrNotFound          = errors.New("resource not found")
    ErrUnauthorized      = errors.New("unauthorized access")
    ErrForbidden         = errors.New("forbidden access")
    ErrInvalidInput      = errors.New("invalid input data")
    ErrDuplicateEntry    = errors.New("duplicate entry")
    ErrExternalAPIFailed = errors.New("external API call failed")
    ErrInsufficientBalance = errors.New("insufficient balance")
)

// Usage in service
func (s *UserService) GetUserByID(id uint) (*models.User, error) {
    user, err := repositories.GetUserByID(id)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, types.ErrNotFound
        }
        return nil, fmt.Errorf("get user: %w", err)
    }
    return user, nil
}

// Usage in controller
func (ctrl *UserController) Get(c *gin.Context) {
    user, err := ctrl.service.GetUserByID(id)
    if err != nil {
        if errors.Is(err, types.ErrNotFound) {
            utils.NotFound(c, err, "User not found")
            return
        }
        utils.InternalServerError(c, err, "Failed to get user")
        return
    }
    utils.Ok(c, user, "User retrieved")
}
```

### 9.3 Error Wrapping Pattern

**✅ ALWAYS wrap errors with context:**

```go
// Repository layer
func GetUserByID(id uint) (*models.User, error) {
    var user models.User
    if err := database.DB.First(&user, id).Error; err != nil {
        // Add context: what failed + which resource
        return nil, fmt.Errorf("failed to get user %d: %w", id, err)
    }
    return &user, nil
}

// Service layer
func (s *UserService) GetUserByID(id uint) (*models.User, error) {
    user, err := repositories.GetUserByID(id)
    if err != nil {
        logger.Errorf("user retrieval failed for ID %d: %v", id, err)
        // Add business context
        return nil, fmt.Errorf("user retrieval failed: %w", err)
    }
    return user, nil
}

// Controller layer
func (ctrl *UserController) Get(c *gin.Context) {
    user, err := ctrl.service.GetUserByID(id)
    if err != nil {
        // Convert to HTTP response
        utils.InternalServerError(c, err, "Failed to get user")
        return
    }
    utils.Ok(c, user, "User retrieved")
}
```

**Error Chain Example:**
```
Original: record not found
    ↓
Repository: failed to get user 123: record not found
    ↓
Service: user retrieval failed: failed to get user 123: record not found
    ↓
Controller: HTTP 404 + "User not found"
```

---

## 10. TESTING PATTERNS

### 10.1 Service Layer Testing Pattern

```go
// File: internal/app/services/user_service_test.go
package services_test

import (
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"

    "your-project/internal/app/dto"
    "your-project/internal/app/services"
    "your-project/internal/domain/models"
    "your-project/pkg/types"
)

// Mock repository
type MockUserRepository struct {
    mock.Mock
}

func (m *MockUserRepository) GetUserByEmail(email string) (*models.User, error) {
    args := m.Called(email)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) CreateUser(user *models.User) error {
    args := m.Called(user)
    return args.Error(0)
}

// Test: Success case
func TestCreateUser_Success(t *testing.T) {
    // Arrange
    mockRepo := new(MockUserRepository)
    service := services.NewUserService(mockRepo)

    req := &dto.CreateUserRequest{
        Name:     "John Doe",
        Email:    "john@example.com",
        Password: "secret123",
    }

    // Expect: Check email doesn't exist
    mockRepo.On("GetUserByEmail", req.Email).
        Return(nil, types.ErrNotFound)

    // Expect: Create user
    mockRepo.On("CreateUser", mock.AnythingOfType("*models.User")).
        Return(nil)

    // Act
    err := service.CreateUser(req)

    // Assert
    assert.NoError(t, err)
    mockRepo.AssertExpectations(t)
}

// Test: Duplicate email
func TestCreateUser_DuplicateEmail(t *testing.T) {
    // Arrange
    mockRepo := new(MockUserRepository)
    service := services.NewUserService(mockRepo)

    req := &dto.CreateUserRequest{
        Name:     "John Doe",
        Email:    "existing@example.com",
        Password: "secret123",
    }

    existingUser := &models.User{
        ID:    1,
        Email: req.Email,
    }

    // Expect: Email already exists
    mockRepo.On("GetUserByEmail", req.Email).
        Return(existingUser, nil)

    // Act
    err := service.CreateUser(req)

    // Assert
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "email already exists")
    mockRepo.AssertExpectations(t)
}

// Test: Repository error
func TestCreateUser_RepositoryError(t *testing.T) {
    // Arrange
    mockRepo := new(MockUserRepository)
    service := services.NewUserService(mockRepo)

    req := &dto.CreateUserRequest{
        Name:     "John Doe",
        Email:    "john@example.com",
        Password: "secret123",
    }

    mockRepo.On("GetUserByEmail", req.Email).
        Return(nil, types.ErrNotFound)

    mockRepo.On("CreateUser", mock.AnythingOfType("*models.User")).
        Return(errors.New("database error"))

    // Act
    err := service.CreateUser(req)

    // Assert
    assert.Error(t, err)
    mockRepo.AssertExpectations(t)
}
```

### 10.2 Table-Driven Testing Pattern

```go
func TestValidateEmail(t *testing.T) {
    tests := []struct {
        name    string
        email   string
        wantErr bool
    }{
        {
            name:    "valid email",
            email:   "user@example.com",
            wantErr: false,
        },
        {
            name:    "missing @",
            email:   "userexample.com",
            wantErr: true,
        },
        {
            name:    "empty email",
            email:   "",
            wantErr: true,
        },
        {
            name:    "missing domain",
            email:   "user@",
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateEmail(tt.email)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateEmail() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

---

## 11. COMPLETE FEATURE IMPLEMENTATION GUIDE

### Step-by-Step: Adding a New "Product" Feature

#### Step 1: Create Model
```go
// File: internal/domain/models/product.go
package models

import (
    "time"
    "gorm.io/gorm"
)

type Product struct {
    ID          uint           `gorm:"primarykey"`
    Name        string         `gorm:"size:255;not null"`
    Description string         `gorm:"type:text"`
    Price       float64        `gorm:"not null"`
    Stock       int            `gorm:"default:0"`
    CategoryID  uint           `gorm:"not null"`
    IsActive    bool           `gorm:"default:true"`
    CreatedAt   time.Time
    UpdatedAt   time.Time
    DeletedAt   gorm.DeletedAt `gorm:"index"`

    // Relationships
    Category    *Category `gorm:"foreignKey:CategoryID"`
}
```

#### Step 2: Create DTOs
```go
// File: internal/app/dto/product_dto.go
package dto

type CreateProductRequest struct {
    Name        string  `json:"name" binding:"required,min=3,max=255"`
    Description string  `json:"description"`
    Price       float64 `json:"price" binding:"required,gt=0"`
    Stock       int     `json:"stock" binding:"gte=0"`
    CategoryID  uint    `json:"category_id" binding:"required"`
}

type UpdateProductRequest struct {
    Name        string  `json:"name" binding:"omitempty,min=3,max=255"`
    Description string  `json:"description"`
    Price       float64 `json:"price" binding:"omitempty,gt=0"`
    Stock       int     `json:"stock" binding:"omitempty,gte=0"`
    CategoryID  uint    `json:"category_id" binding:"omitempty"`
    IsActive    bool    `json:"is_active"`
}

type ProductResponse struct {
    ID          uint      `json:"id"`
    Name        string    `json:"name"`
    Description string    `json:"description"`
    Price       float64   `json:"price"`
    Stock       int       `json:"stock"`
    CategoryID  uint      `json:"category_id"`
    Category    string    `json:"category"`
    IsActive    bool      `json:"is_active"`
    CreatedAt   time.Time `json:"created_at"`
}
```

#### Step 3: Create Repository
```go
// File: internal/domain/repositories/product_repo.go
package repositories

import (
    "github.com/your-org/project/internal/adapters/database"
    "github.com/your-org/project/internal/domain/models"
    "github.com/your-org/project/pkg/logger"
)

func CreateProduct(product *models.Product) error {
    if err := database.DB.Create(product).Error; err != nil {
        logger.Errorf("failed to create product: %v", err)
        return err
    }
    return nil
}

func GetProductByID(id uint) (*models.Product, error) {
    var product models.Product
    if err := database.DB.
        Preload("Category").
        Where("id = ?", id).
        First(&product).Error; err != nil {
        logger.Errorf("failed to get product %d: %v", id, err)
        return nil, err
    }
    return &product, nil
}

func UpdateProduct(product *models.Product) error {
    if err := database.DB.Save(product).Error; err != nil {
        logger.Errorf("failed to update product: %v", err)
        return err
    }
    return nil
}

func DeleteProduct(id uint) error {
    if err := database.DB.Delete(&models.Product{}, id).Error; err != nil {
        logger.Errorf("failed to delete product: %v", err)
        return err
    }
    return nil
}

func ListProducts(page, pageSize int, filters map[string]interface{}) ([]*models.Product, int64, error) {
    var products []*models.Product
    var total int64

    query := database.DB.Model(&models.Product{}).Preload("Category")

    // Apply filters
    if categoryID, ok := filters["category_id"]; ok {
        query = query.Where("category_id = ?", categoryID)
    }
    if isActive, ok := filters["is_active"]; ok {
        query = query.Where("is_active = ?", isActive)
    }

    // Count total
    query.Count(&total)

    // Paginate
    offset := (page - 1) * pageSize
    if err := query.Offset(offset).Limit(pageSize).Find(&products).Error; err != nil {
        return nil, 0, err
    }

    return products, total, nil
}
```

#### Step 4: Create Service
```go
// File: internal/app/services/product_service.go
package services

import (
    "fmt"

    "github.com/your-org/project/internal/app/dto"
    "github.com/your-org/project/internal/domain/models"
    "github.com/your-org/project/internal/domain/repositories"
    "github.com/your-org/project/pkg/logger"
    "github.com/your-org/project/pkg/types"
)

type ProductService struct{}

func NewProductService() *ProductService {
    return &ProductService{}
}

func (s *ProductService) CreateProduct(req *dto.CreateProductRequest) (*models.Product, error) {
    // 1. Validate business rules
    if err := s.validateProduct(req); err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }

    // 2. Check category exists
    category, err := repositories.GetCategoryByID(req.CategoryID)
    if err != nil {
        return nil, fmt.Errorf("invalid category: %w", err)
    }

    // 3. Build product
    product := &models.Product{
        Name:        req.Name,
        Description: req.Description,
        Price:       req.Price,
        Stock:       req.Stock,
        CategoryID:  req.CategoryID,
        IsActive:    true,
    }

    // 4. Save product
    if err := repositories.CreateProduct(product); err != nil {
        logger.Errorf("failed to create product: %v", err)
        return nil, fmt.Errorf("failed to create product: %w", err)
    }

    logger.Infof("Product created: ID=%d, Name=%s", product.ID, product.Name)
    return product, nil
}

func (s *ProductService) GetProductByID(id uint) (*models.Product, error) {
    product, err := repositories.GetProductByID(id)
    if err != nil {
        return nil, types.ErrNotFound
    }
    return product, nil
}

func (s *ProductService) UpdateProduct(id uint, req *dto.UpdateProductRequest) (*models.Product, error) {
    // 1. Get existing product
    product, err := repositories.GetProductByID(id)
    if err != nil {
        return nil, types.ErrNotFound
    }

    // 2. Update fields
    if req.Name != "" {
        product.Name = req.Name
    }
    if req.Description != "" {
        product.Description = req.Description
    }
    if req.Price > 0 {
        product.Price = req.Price
    }
    if req.Stock >= 0 {
        product.Stock = req.Stock
    }
    if req.CategoryID > 0 {
        product.CategoryID = req.CategoryID
    }
    product.IsActive = req.IsActive

    // 3. Save
    if err := repositories.UpdateProduct(product); err != nil {
        return nil, fmt.Errorf("failed to update product: %w", err)
    }

    return product, nil
}

func (s *ProductService) DeleteProduct(id uint) error {
    // Check exists
    _, err := repositories.GetProductByID(id)
    if err != nil {
        return types.ErrNotFound
    }

    // Delete
    if err := repositories.DeleteProduct(id); err != nil {
        return fmt.Errorf("failed to delete product: %w", err)
    }

    logger.Infof("Product deleted: ID=%d", id)
    return nil
}

// Private helper
func (s *ProductService) validateProduct(req *dto.CreateProductRequest) error {
    if req.Price <= 0 {
        return fmt.Errorf("price must be greater than 0")
    }
    if req.Stock < 0 {
        return fmt.Errorf("stock cannot be negative")
    }
    return nil
}
```

#### Step 5: Create Controller
```go
// File: internal/app/controllers/product_controller.go
package controllers

import (
    "strconv"

    "github.com/gin-gonic/gin"
    "github.com/your-org/project/internal/app/dto"
    "github.com/your-org/project/internal/app/services"
    "github.com/your-org/project/pkg/utils"
)

type ProductController struct {
    service *services.ProductService
}

func NewProductController(service *services.ProductService) *ProductController {
    return &ProductController{service: service}
}

// List handles GET /api/products
func (ctrl *ProductController) List(c *gin.Context) {
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

    products, total, err := ctrl.service.ListProducts(page, pageSize)
    if err != nil {
        utils.InternalServerError(c, err, "Failed to list products")
        return
    }

    response := map[string]interface{}{
        "products":    products,
        "page":        page,
        "page_size":   pageSize,
        "total":       total,
        "total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
    }

    utils.Ok(c, response, "Products retrieved successfully")
}

// Get handles GET /api/products/:id
func (ctrl *ProductController) Get(c *gin.Context) {
    id, err := strconv.ParseUint(c.Param("id"), 10, 32)
    if err != nil {
        utils.BadRequest(c, err, "Invalid product ID")
        return
    }

    product, err := ctrl.service.GetProductByID(uint(id))
    if err != nil {
        utils.NotFound(c, err, "Product not found")
        return
    }

    utils.Ok(c, product, "Product retrieved successfully")
}

// Create handles POST /api/products
func (ctrl *ProductController) Create(c *gin.Context) {
    var req dto.CreateProductRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        utils.BadRequest(c, err, "Invalid request")
        return
    }

    product, err := ctrl.service.CreateProduct(&req)
    if err != nil {
        utils.InternalServerError(c, err, "Failed to create product")
        return
    }

    utils.Created(c, product, "Product created successfully")
}

// Update handles PUT /api/products/:id
func (ctrl *ProductController) Update(c *gin.Context) {
    id, err := strconv.ParseUint(c.Param("id"), 10, 32)
    if err != nil {
        utils.BadRequest(c, err, "Invalid product ID")
        return
    }

    var req dto.UpdateProductRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        utils.BadRequest(c, err, "Invalid request")
        return
    }

    product, err := ctrl.service.UpdateProduct(uint(id), &req)
    if err != nil {
        utils.InternalServerError(c, err, "Failed to update product")
        return
    }

    utils.Ok(c, product, "Product updated successfully")
}

// Delete handles DELETE /api/products/:id
func (ctrl *ProductController) Delete(c *gin.Context) {
    id, err := strconv.ParseUint(c.Param("id"), 10, 32)
    if err != nil {
        utils.BadRequest(c, err, "Invalid product ID")
        return
    }

    if err := ctrl.service.DeleteProduct(uint(id)); err != nil {
        utils.InternalServerError(c, err, "Failed to delete product")
        return
    }

    utils.NoContent(c)
}
```

#### Step 6: Register Routes
```go
// File: internal/app/routers/index.go
func RegisterRoutes(router *gin.Engine) {
    // ... existing routes ...

    // Initialize product dependencies
    productService := services.NewProductService()
    productController := controllers.NewProductController(productService)

    // Register product routes
    api := router.Group("/api")
    api.Use(middlewares.AuthMiddleware()) // Protected routes
    {
        products := api.Group("/products")
        {
            products.GET("", productController.List)
            products.GET("/:id", productController.Get)
            products.POST("", productController.Create)
            products.PUT("/:id", productController.Update)
            products.DELETE("/:id", productController.Delete)
        }
    }
}
```

#### Step 7: Create Tests
```go
// File: internal/app/services/product_service_test.go
package services_test

// Add comprehensive tests here
```

---

## 12. PATTERN EXAMPLES FROM CODEBASE

### 12.1 Auth Pattern (from auth_controller.go)

**✅ Struct-Based Controller:**
```go
type AuthController struct {
    service *services.AuthService
}

func NewAuthController(service *services.AuthService) *AuthController {
    return &AuthController{service: service}
}

func (ctrl *AuthController) Login(c *gin.Context) {
    var req dto.LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        utils.BadRequest(c, err, "Invalid request")
        return
    }

    response, err := ctrl.service.Login(&req)
    if err != nil {
        utils.Unauthorized(c, err, err.Error())
        return
    }

    utils.Ok(c, response, "Login successful")
}
```

**✅ Service with Business Logic:**
```go
func (s *AuthService) Login(req *dto.LoginRequest) (*dto.LoginResponse, error) {
    // 1. Get user
    user, err := repositories.GetUserByEmail(req.Email)
    if err != nil {
        return nil, fmt.Errorf("invalid email or password")
    }

    // 2. Validate password
    if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
        return nil, fmt.Errorf("invalid email or password")
    }

    // 3. Generate OTP
    code := time.Now().Nanosecond()%900000 + 100000
    otp := &models.OTPSession{
        ID:        utils.GenerateRandomString(32),
        UserID:    user.ID,
        Email:     user.Email,
        Code:      utils.IntToZeroPaddedString(code, 6),
        ExpiresAt: time.Now().Add(10 * time.Minute),
    }

    // 4. Save OTP
    if err := repositories.CreateOTPSession(otp); err != nil {
        return nil, fmt.Errorf("could not create OTP session")
    }

    // 5. Send email
    go s.sendEmailOTP(otp.Code, user.Email)

    return &dto.LoginResponse{
        OTPToken: otp.ID,
        RoleID:   user.UserRoles.RoleID,
        Role:     string(user.UserRoles.Role.Name),
    }, nil
}
```

### 12.2 Client Management Pattern (from client_controller.go)

**Complete CRUD Implementation:**
```go
type ClientController struct {
    service *services.ClientService
}

func NewClientController(service *services.ClientService) *ClientController {
    return &ClientController{service: service}
}

// List with DataTable support
func (ctrl *ClientController) List(c *gin.Context) {
    var dtRequest datatable.DataTableRequest
    if err := c.ShouldBindJSON(&dtRequest); err != nil {
        utils.BadRequest(c, err, "Invalid request")
        return
    }

    response := ctrl.service.GetClientDataTable(&dtRequest)
    utils.Ok(c, response, "OK")
}

// Standard CRUD methods
func (ctrl *ClientController) Get(c *gin.Context) { }
func (ctrl *ClientController) Create(c *gin.Context) { }
func (ctrl *ClientController) Update(c *gin.Context) { }
func (ctrl *ClientController) Delete(c *gin.Context) { }
```

---

## 13. ANTI-PATTERNS TO AVOID

### 13.1 ❌ Business Logic in Controller

**WRONG:**
```go
func CreateUser(c *gin.Context) {
    var req dto.CreateUserRequest
    c.ShouldBindJSON(&req)

    // ❌ Password hashing in controller!
    hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(req.Password))

    // ❌ Direct database access!
    user := &models.User{
        Name:     req.Name,
        Email:    req.Email,
        Password: string(hashedPassword),
    }
    database.DB.Create(user)

    c.JSON(200, user)
}
```

**CORRECT:**
```go
func (ctrl *UserController) Create(c *gin.Context) {
    var req dto.CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        utils.BadRequest(c, err, "Invalid request")
        return
    }

    // ✅ Delegate to service
    user, err := ctrl.service.CreateUser(&req)
    if err != nil {
        utils.InternalServerError(c, err, "Failed to create user")
        return
    }

    utils.Created(c, user, "User created")
}
```

---

### 13.2 ❌ Direct Database Access in Service

**WRONG:**
```go
func (s *UserService) CreateUser(req *dto.CreateUserRequest) error {
    // ❌ Direct DB access in service!
    var existingUser models.User
    database.DB.Where("email = ?", req.Email).First(&existingUser)

    user := &models.User{...}
    database.DB.Create(user)
    return nil
}
```

**CORRECT:**
```go
func (s *UserService) CreateUser(req *dto.CreateUserRequest) error {
    // ✅ Use repository
    existing, _ := repositories.GetUserByEmail(req.Email)
    if existing != nil {
        return fmt.Errorf("email already exists")
    }

    user := &models.User{...}
    return repositories.CreateUser(user)
}
```

---

### 13.3 ❌ Standalone Controller Functions

**WRONG:**
```go
// ❌ Old pattern - standalone function
func Login(c *gin.Context) {
    // Direct implementation
}

// Router registration
route.POST("/login", Login)
```

**CORRECT:**
```go
// ✅ Struct-based controller
type AuthController struct {
    service *services.AuthService
}

func NewAuthController(service *services.AuthService) *AuthController {
    return &AuthController{service: service}
}

func (ctrl *AuthController) Login(c *gin.Context) {
    // Implementation
}

// Router registration
authController := NewAuthController(authService)
route.POST("/login", authController.Login)
```

---

### 13.4 ❌ God Service (Too Many Responsibilities)

**WRONG:**
```go
// ❌ One service doing everything
type AppService struct {}

func (s *AppService) CreateUser() error { }
func (s *AppService) ProcessPayment() error { }
func (s *AppService) SendEmail() error { }
func (s *AppService) GenerateReport() error { }
func (s *AppService) CalculateTax() error { }
// ... 50 more methods
```

**CORRECT:**
```go
// ✅ Separate services with single responsibility
type UserService struct {}
type PaymentService struct {}
type EmailService struct {}
type ReportService struct {}
type TaxService struct {}
```

---

### 13.5 ❌ Circular Dependencies

**WRONG:**
```go
// pkg/enums/constants.go
import "internal/app/dto"  // ❌ pkg importing internal!

var NotificationMapping = map[string]dto.NotificationRoute{...}
```

**CORRECT:**
```go
// internal/app/services/notification_router.go
import "internal/app/dto"  // ✅ services can import dto

var NotificationMapping = map[string]dto.NotificationRoute{...}
```

---

## 🎯 SUMMARY: Key Takeaways

### ✅ DO:
1. **Follow Clean Architecture** - Layers have clear responsibilities
2. **Use Struct-Based Controllers** - Never standalone functions
3. **Keep Controllers Thin** - Only HTTP handling (<50 lines)
4. **Keep Services Fat** - All business logic goes here
5. **Use Repositories** - Abstract all database access
6. **Use DTOs** - Never expose models directly
7. **Use Response Utilities** - Consistent API responses
8. **Handle Errors Properly** - Always wrap with context
9. **Write Tests** - Minimum 70% coverage for services
10. **Follow Dependency Direction** - Always inward

### ❌ DON'T:
1. **Business Logic in Controllers** - Move to services
2. **Database Access in Services** - Use repositories
3. **Standalone Controller Functions** - Use struct-based
4. **God Services** - Split by domain
5. **Circular Dependencies** - pkg cannot import internal
6. **Ignore Errors** - Always handle and wrap
7. **Hardcode Secrets** - Use environment variables
8. **Skip Tests** - Required for all services
9. **Large Files** - Max 300 lines per file
10. **Large Functions** - Max 100 lines per function

---

**END OF DESIGN PATTERNS DOCUMENT**

*Last Updated: 2025-11-08*
*Version: 1.0*
*Next Review: Every 3 months or after major architectural changes*

---

## 📚 Related Documentation

- **CODING_STANDARDS.md** - Detailed coding standards and guidelines
- **AI_AGENT_RULES.md** - Quick reference rules for AI agents
- **README.md** - Project overview and setup instructions
- **API_DOCUMENTATION.md** - API endpoint documentation

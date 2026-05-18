# Golang Middleware Masterclass
## From Beginner to Expert

---

## Table of Contents

1. [Introduction to Middleware](#introduction)
2. [HTTP Fundamentals in Go](#http-fundamentals)
3. [Basic Middleware Patterns](#basic-middleware)
4. [The http.Handler Interface](#handler-interface)
5. [Building Your First Middleware](#first-middleware)
6. [Middleware Chaining](#middleware-chaining)
7. [Context and Request/Response Manipulation](#context-manipulation)
8. [Common Middleware Patterns](#common-patterns)
9. [Advanced Middleware Techniques](#advanced-techniques)
10. [Third-Party Middleware Libraries](#third-party)
11. [Performance Optimization](#performance)
12. [Testing Middleware](#testing)
13. [Production Patterns](#production-patterns)
14. [Real-World Examples](#real-world)
15. [Anti-Patterns and Pitfalls](#anti-patterns)

---

## 1. Introduction to Middleware {#introduction}

### What is Middleware?

Middleware is a function that wraps HTTP handlers to provide cross-cutting functionality. Think of it as a layer between the HTTP request and your actual business logic.

**Analogy**: Middleware is like security checkpoints at an airport. Each checkpoint (middleware) performs a specific task (check ID, scan bags, verify ticket) before you reach your destination (handler).

### Why Use Middleware?

- **Separation of Concerns**: Keep authentication, logging, CORS separate from business logic
- **Reusability**: Write once, use across multiple routes
- **Composability**: Chain multiple middlewares together
- **Maintainability**: Easier to test and modify individual concerns

### The Middleware Pipeline

```
Request → Middleware 1 → Middleware 2 → Middleware 3 → Handler → Response
           (Logging)      (Auth)         (CORS)        (Business)
```

---

## 2. HTTP Fundamentals in Go {#http-fundamentals}

### The http.Handler Interface

This is the foundation of Go's HTTP system:

```go
type Handler interface {
    ServeHTTP(ResponseWriter, *Request)
}
```

Any type that implements `ServeHTTP` is an `http.Handler`.

### http.HandlerFunc

A function type that implements `http.Handler`:

```go
type HandlerFunc func(ResponseWriter, *Request)

func (f HandlerFunc) ServeHTTP(w ResponseWriter, r *Request) {
    f(w, r)
}
```

### Example: Basic Handler

```go
package main

import (
    "fmt"
    "net/http"
)

// Method 1: Struct implementing Handler
type HelloHandler struct{}

func (h HelloHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello from struct handler!")
}

// Method 2: HandlerFunc
func helloHandlerFunc(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello from HandlerFunc!")
}

func main() {
    // Using struct handler
    http.Handle("/hello1", HelloHandler{})
    
    // Using HandlerFunc (automatic conversion)
    http.HandleFunc("/hello2", helloHandlerFunc)
    
    // Using Handler directly
    http.Handle("/hello3", http.HandlerFunc(helloHandlerFunc))
    
    http.ListenAndServe(":8080", nil)
}
```

---

## 3. Basic Middleware Patterns {#basic-middleware}

### Pattern 1: Function Wrapping Function

The most common pattern - a function that takes a handler and returns a handler:

```go
func middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Before handler
        fmt.Println("Before")
        
        // Call the next handler
        next.ServeHTTP(w, r)
        
        // After handler
        fmt.Println("After")
    })
}
```

### Pattern 2: HandlerFunc Wrapping

```go
func middleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Before
        next(w, r)
        // After
    }
}
```

### Pattern 3: Adapter Pattern

```go
type Adapter func(http.Handler) http.Handler

func (a Adapter) Then(h http.Handler) http.Handler {
    return a(h)
}
```

---

## 4. The http.Handler Interface Deep Dive {#handler-interface}

### Understanding the Interface

```go
type Handler interface {
    ServeHTTP(ResponseWriter, *Request)
}
```

### ResponseWriter Interface

```go
type ResponseWriter interface {
    Header() http.Header          // Get/Set headers
    Write([]byte) (int, error)    // Write response body
    WriteHeader(statusCode int)   // Set HTTP status code
}
```

**Critical**: `WriteHeader()` must be called before `Write()`. Once you call `Write()`, headers are automatically sent with status 200.

### Common Mistakes

```go
// WRONG - Headers sent after WriteHeader
w.WriteHeader(http.StatusOK)
w.Header().Set("X-Custom", "value") // Too late!
w.Write([]byte("response"))

// CORRECT - Headers before WriteHeader
w.Header().Set("X-Custom", "value")
w.WriteHeader(http.StatusOK)
w.Write([]byte("response"))
```

---

## 5. Building Your First Middleware {#first-middleware}

### Simple Logging Middleware

```go
package main

import (
    "log"
    "net/http"
    "time"
)

func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        // Call the next handler
        next.ServeHTTP(w, r)
        
        // Log after handler completes
        log.Printf(
            "%s %s %s",
            r.Method,
            r.RequestURI,
            time.Since(start),
        )
    })
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Welcome Home!"))
}

func main() {
    // Wrap the handler with middleware
    http.Handle("/", loggingMiddleware(http.HandlerFunc(homeHandler)))
    http.ListenAndServe(":8080", nil)
}
```

### Enhanced Logging with Status Code

To capture the status code, we need a custom ResponseWriter:

```go
type responseWriter struct {
    http.ResponseWriter
    status      int
    wroteHeader bool
}

func wrapResponseWriter(w http.ResponseWriter) *responseWriter {
    return &responseWriter{ResponseWriter: w}
}

func (rw *responseWriter) Status() int {
    return rw.status
}

func (rw *responseWriter) WriteHeader(code int) {
    if rw.wroteHeader {
        return
    }

    rw.status = code
    rw.ResponseWriter.WriteHeader(code)
    rw.wroteHeader = true
}

func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        wrapped := wrapResponseWriter(w)
        
        next.ServeHTTP(wrapped, r)
        
        log.Printf(
            "status=%d method=%s path=%s duration=%s",
            wrapped.status,
            r.Method,
            r.URL.Path,
            time.Since(start),
        )
    })
}
```

---

## 6. Middleware Chaining {#middleware-chaining}

### Manual Chaining

```go
func main() {
    handler := homeHandler
    handler = loggingMiddleware(handler)
    handler = authMiddleware(handler)
    handler = corsMiddleware(handler)
    
    http.Handle("/", handler)
    http.ListenAndServe(":8080", nil)
}
```

Order matters! Execution flows:
```
Request → CORS → Auth → Logging → Handler → Logging → Auth → CORS → Response
```

### Chain Helper Function

```go
func Chain(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
    for i := len(middlewares) - 1; i >= 0; i-- {
        h = middlewares[i](h)
    }
    return h
}

// Usage
handler := Chain(
    http.HandlerFunc(homeHandler),
    loggingMiddleware,
    authMiddleware,
    corsMiddleware,
)
```

### Alice - Popular Chaining Library

```go
import "github.com/justinas/alice"

chain := alice.New(
    loggingMiddleware,
    authMiddleware,
    corsMiddleware,
).Then(http.HandlerFunc(homeHandler))

http.Handle("/", chain)
```

### Building Your Own Fluent Chain

```go
type Chain struct {
    middlewares []func(http.Handler) http.Handler
}

func NewChain(middlewares ...func(http.Handler) http.Handler) Chain {
    return Chain{middlewares: middlewares}
}

func (c Chain) Then(h http.Handler) http.Handler {
    for i := len(c.middlewares) - 1; i >= 0; i-- {
        h = c.middlewares[i](h)
    }
    return h
}

func (c Chain) Append(middlewares ...func(http.Handler) http.Handler) Chain {
    newMiddlewares := make([]func(http.Handler) http.Handler, 
        len(c.middlewares)+len(middlewares))
    copy(newMiddlewares, c.middlewares)
    copy(newMiddlewares[len(c.middlewares):], middlewares)
    return Chain{middlewares: newMiddlewares}
}

// Usage
commonChain := NewChain(loggingMiddleware, recoveryMiddleware)
apiChain := commonChain.Append(authMiddleware, rateLimitMiddleware)

http.Handle("/api/users", apiChain.Then(http.HandlerFunc(usersHandler)))
```

---

## 7. Context and Request/Response Manipulation {#context-manipulation}

### Using Context for Request-Scoped Values

Context is how you pass values between middleware layers:

```go
package main

import (
    "context"
    "fmt"
    "net/http"
)

type contextKey string

const userIDKey contextKey = "userID"

// Middleware that extracts user ID from header
func authMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        userID := r.Header.Get("X-User-ID")
        
        if userID == "" {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        
        // Add userID to context
        ctx := context.WithValue(r.Context(), userIDKey, userID)
        
        // Pass the request with new context to next handler
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func handler(w http.ResponseWriter, r *http.Request) {
    // Retrieve userID from context
    userID, ok := r.Context().Value(userIDKey).(string)
    if !ok {
        http.Error(w, "Internal error", http.StatusInternalServerError)
        return
    }
    
    fmt.Fprintf(w, "Hello, user %s!", userID)
}

func main() {
    http.Handle("/", authMiddleware(http.HandlerFunc(handler)))
    http.ListenAndServe(":8080", nil)
}
```

### Context Best Practices

**DO:**
```go
// Use custom types for keys to avoid collisions
type contextKey string
const requestIDKey contextKey = "requestID"

// Type-safe context getters
func GetRequestID(ctx context.Context) (string, bool) {
    id, ok := ctx.Value(requestIDKey).(string)
    return id, ok
}
```

**DON'T:**
```go
// Avoid string keys - they can collide
ctx = context.WithValue(ctx, "userID", id)

// Don't store large objects in context
ctx = context.WithValue(ctx, "database", db) // BAD!
```

### Cancellation and Timeouts

```go
func timeoutMiddleware(timeout time.Duration) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            ctx, cancel := context.WithTimeout(r.Context(), timeout)
            defer cancel()
            
            r = r.WithContext(ctx)
            
            // Channel to signal handler completion
            done := make(chan struct{})
            
            go func() {
                next.ServeHTTP(w, r)
                close(done)
            }()
            
            select {
            case <-done:
                return
            case <-ctx.Done():
                http.Error(w, "Request timeout", http.StatusGatewayTimeout)
            }
        })
    }
}
```

---

## 8. Common Middleware Patterns {#common-patterns}

### Authentication Middleware

```go
func jwtAuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        authHeader := r.Header.Get("Authorization")
        
        if authHeader == "" {
            http.Error(w, "Missing authorization header", http.StatusUnauthorized)
            return
        }
        
        // Extract token from "Bearer <token>"
        parts := strings.Split(authHeader, " ")
        if len(parts) != 2 || parts[0] != "Bearer" {
            http.Error(w, "Invalid authorization format", http.StatusUnauthorized)
            return
        }
        
        token := parts[1]
        
        // Validate JWT (simplified)
        claims, err := validateJWT(token)
        if err != nil {
            http.Error(w, "Invalid token", http.StatusUnauthorized)
            return
        }
        
        // Add claims to context
        ctx := context.WithValue(r.Context(), "claims", claims)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

### CORS Middleware

```go
func corsMiddleware(allowedOrigins []string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            origin := r.Header.Get("Origin")
            
            // Check if origin is allowed
            allowed := false
            for _, allowedOrigin := range allowedOrigins {
                if origin == allowedOrigin || allowedOrigin == "*" {
                    allowed = true
                    break
                }
            }
            
            if allowed {
                w.Header().Set("Access-Control-Allow-Origin", origin)
                w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
                w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
                w.Header().Set("Access-Control-Allow-Credentials", "true")
            }
            
            // Handle preflight
            if r.Method == "OPTIONS" {
                w.WriteHeader(http.StatusOK)
                return
            }
            
            next.ServeHTTP(w, r)
        })
    }
}
```

### Rate Limiting Middleware

```go
import (
    "golang.org/x/time/rate"
    "sync"
)

func rateLimitMiddleware(requestsPerSecond int) func(http.Handler) http.Handler {
    // Map of IP addresses to rate limiters
    limiters := make(map[string]*rate.Limiter)
    mu := &sync.Mutex{}
    
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            ip := r.RemoteAddr
            
            mu.Lock()
            limiter, exists := limiters[ip]
            if !exists {
                limiter = rate.NewLimiter(rate.Limit(requestsPerSecond), requestsPerSecond*2)
                limiters[ip] = limiter
            }
            mu.Unlock()
            
            if !limiter.Allow() {
                http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
                return
            }
            
            next.ServeHTTP(w, r)
        })
    }
}
```

### Request ID Middleware

```go
import "github.com/google/uuid"

func requestIDMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        requestID := r.Header.Get("X-Request-ID")
        
        if requestID == "" {
            requestID = uuid.New().String()
        }
        
        // Add to response headers
        w.Header().Set("X-Request-ID", requestID)
        
        // Add to context
        ctx := context.WithValue(r.Context(), "requestID", requestID)
        
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

### Recovery Middleware (Panic Handler)

```go
func recoveryMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                // Log the error with stack trace
                log.Printf("Panic: %v\n%s", err, debug.Stack())
                
                // Return 500 to client
                w.WriteHeader(http.StatusInternalServerError)
                w.Write([]byte("Internal Server Error"))
            }
        }()
        
        next.ServeHTTP(w, r)
    })
}
```

### Compression Middleware

```go
import (
    "compress/gzip"
    "io"
    "strings"
)

type gzipResponseWriter struct {
    io.Writer
    http.ResponseWriter
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
    return w.Writer.Write(b)
}

func compressionMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Check if client accepts gzip
        if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
            next.ServeHTTP(w, r)
            return
        }
        
        // Set response header
        w.Header().Set("Content-Encoding", "gzip")
        
        // Create gzip writer
        gz := gzip.NewWriter(w)
        defer gz.Close()
        
        // Wrap response writer
        gzipWriter := gzipResponseWriter{Writer: gz, ResponseWriter: w}
        
        next.ServeHTTP(gzipWriter, r)
    })
}
```

---

## 9. Advanced Middleware Techniques {#advanced-techniques}

### Conditional Middleware

Apply middleware only when certain conditions are met:

```go
func conditionalMiddleware(
    condition func(*http.Request) bool,
    middleware func(http.Handler) http.Handler,
) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if condition(r) {
                middleware(next).ServeHTTP(w, r)
            } else {
                next.ServeHTTP(w, r)
            }
        })
    }
}

// Usage: Only apply auth to /api/* routes
isAPIRoute := func(r *http.Request) bool {
    return strings.HasPrefix(r.URL.Path, "/api/")
}

handler := conditionalMiddleware(isAPIRoute, authMiddleware)(finalHandler)
```

### Middleware with Configuration

```go
type AuthConfig struct {
    Secret       string
    TokenLookup  string // "header:Authorization" or "query:token"
    SkipPaths    []string
}

func authMiddleware(config AuthConfig) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Check if path should be skipped
            for _, path := range config.SkipPaths {
                if r.URL.Path == path {
                    next.ServeHTTP(w, r)
                    return
                }
            }
            
            // Extract token based on TokenLookup
            var token string
            parts := strings.Split(config.TokenLookup, ":")
            
            switch parts[0] {
            case "header":
                token = r.Header.Get(parts[1])
            case "query":
                token = r.URL.Query().Get(parts[1])
            case "cookie":
                cookie, err := r.Cookie(parts[1])
                if err == nil {
                    token = cookie.Value
                }
            }
            
            if token == "" {
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                return
            }
            
            // Validate token with secret
            claims, err := validateToken(token, config.Secret)
            if err != nil {
                http.Error(w, "Invalid token", http.StatusUnauthorized)
                return
            }
            
            ctx := context.WithValue(r.Context(), "claims", claims)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

// Usage
authMW := authMiddleware(AuthConfig{
    Secret:      "my-secret-key",
    TokenLookup: "header:Authorization",
    SkipPaths:   []string{"/health", "/login"},
})
```

### Middleware Factory Pattern

```go
type MiddlewareFactory struct {
    logger *log.Logger
    db     *sql.DB
    config *Config
}

func NewMiddlewareFactory(logger *log.Logger, db *sql.DB, config *Config) *MiddlewareFactory {
    return &MiddlewareFactory{
        logger: logger,
        db:     db,
        config: config,
    }
}

func (mf *MiddlewareFactory) Logging() func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()
            next.ServeHTTP(w, r)
            mf.logger.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
        })
    }
}

func (mf *MiddlewareFactory) DBTransaction() func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            tx, err := mf.db.Begin()
            if err != nil {
                http.Error(w, "Database error", http.StatusInternalServerError)
                return
            }
            
            defer tx.Rollback() // Rollback if not committed
            
            ctx := context.WithValue(r.Context(), "tx", tx)
            
            next.ServeHTTP(w, r.WithContext(ctx))
            
            // Commit if no error occurred (you'd need custom response writer to track this)
        })
    }
}

// Usage
factory := NewMiddlewareFactory(logger, db, config)
http.Handle("/", factory.Logging()(factory.DBTransaction()(handler)))
```

### Per-Route Middleware with gorilla/mux

```go
import "github.com/gorilla/mux"

func setupRouter() *mux.Router {
    r := mux.NewRouter()
    
    // Global middleware
    r.Use(loggingMiddleware)
    r.Use(recoveryMiddleware)
    
    // Public routes - no auth
    r.HandleFunc("/login", loginHandler).Methods("POST")
    r.HandleFunc("/register", registerHandler).Methods("POST")
    
    // Protected API routes
    api := r.PathPrefix("/api").Subrouter()
    api.Use(authMiddleware)
    api.Use(rateLimitMiddleware(100))
    
    api.HandleFunc("/users", listUsersHandler).Methods("GET")
    api.HandleFunc("/users/{id}", getUserHandler).Methods("GET")
    
    // Admin routes - additional admin check
    admin := api.PathPrefix("/admin").Subrouter()
    admin.Use(adminMiddleware)
    
    admin.HandleFunc("/stats", statsHandler).Methods("GET")
    
    return r
}
```

### Middleware with Metrics Collection

```go
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    httpRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "path", "status"},
    )
    
    httpRequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "http_request_duration_seconds",
            Help: "HTTP request duration in seconds",
        },
        []string{"method", "path"},
    )
)

func metricsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        wrapped := wrapResponseWriter(w)
        next.ServeHTTP(wrapped, r)
        
        duration := time.Since(start).Seconds()
        
        httpRequestsTotal.WithLabelValues(
            r.Method,
            r.URL.Path,
            fmt.Sprintf("%d", wrapped.status),
        ).Inc()
        
        httpRequestDuration.WithLabelValues(
            r.Method,
            r.URL.Path,
        ).Observe(duration)
    })
}
```

---

## 10. Third-Party Middleware Libraries {#third-party}

### Negroni - Idiomatic Middleware

```go
import "github.com/urfave/negroni"

func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/", homeHandler)
    
    n := negroni.Classic() // Includes Recovery, Logger, Static
    n.Use(negroni.HandlerFunc(customMiddleware))
    n.UseHandler(mux)
    
    http.ListenAndServe(":8080", n)
}

func customMiddleware(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
    // Before
    next(w, r)
    // After
}
```

### Chi - Lightweight Router with Middleware

```go
import "github.com/go-chi/chi/v5"

func main() {
    r := chi.NewRouter()
    
    // Global middleware
    r.Use(middleware.RequestID)
    r.Use(middleware.RealIP)
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)
    r.Use(middleware.Timeout(60 * time.Second))
    
    r.Get("/", homeHandler)
    
    // Grouped routes with additional middleware
    r.Group(func(r chi.Router) {
        r.Use(authMiddleware)
        r.Get("/protected", protectedHandler)
    })
    
    http.ListenAndServe(":8080", r)
}
```

### Echo Framework

```go
import "github.com/labstack/echo/v4"

func main() {
    e := echo.New()
    
    // Middleware
    e.Use(middleware.Logger())
    e.Use(middleware.Recover())
    e.Use(middleware.CORS())
    
    // Routes
    e.GET("/", homeHandler)
    
    e.Logger.Fatal(e.Start(":8080"))
}

func homeHandler(c echo.Context) error {
    return c.String(http.StatusOK, "Hello, World!")
}
```

### Gin Framework

```go
import "github.com/gin-gonic/gin"

func main() {
    r := gin.Default() // Includes Logger and Recovery
    
    r.Use(customMiddleware())
    
    r.GET("/", func(c *gin.Context) {
        c.JSON(200, gin.H{"message": "Hello"})
    })
    
    r.Run(":8080")
}

func customMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Before
        c.Next()
        // After
    }
}
```

---

## 11. Performance Optimization {#performance}

### Avoid Allocations in Hot Paths

```go
// BAD - Allocates on every request
func badLoggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        next.ServeHTTP(w, r)
        
        // String concatenation allocates
        msg := r.Method + " " + r.URL.Path + " " + time.Since(start).String()
        log.Println(msg)
    })
}

// GOOD - Use strings.Builder or fmt.Sprintf
func goodLoggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        next.ServeHTTP(w, r)
        
        log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
    })
}
```

### Sync.Pool for Repeated Allocations

```go
var responseWriterPool = sync.Pool{
    New: func() interface{} {
        return &responseWriter{}
    },
}

func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        wrapped := responseWriterPool.Get().(*responseWriter)
        wrapped.ResponseWriter = w
        wrapped.status = 0
        wrapped.wroteHeader = false
        
        defer responseWriterPool.Put(wrapped)
        
        next.ServeHTTP(wrapped, r)
    })
}
```

### Context Value Performance

```go
// SLOW - Type assertion on every access
userID := r.Context().Value("userID").(string)

// FASTER - Custom context type with typed methods
type requestContext struct {
    context.Context
    userID string
}

func (rc *requestContext) UserID() string {
    return rc.userID
}
```

### Benchmarking Middleware

```go
func BenchmarkLoggingMiddleware(b *testing.B) {
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("OK"))
    })
    
    middleware := loggingMiddleware(handler)
    
    req := httptest.NewRequest("GET", "/", nil)
    w := httptest.NewRecorder()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        middleware.ServeHTTP(w, req)
    }
}
```

---

## 12. Testing Middleware {#testing}

### Unit Testing Middleware

```go
import (
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestLoggingMiddleware(t *testing.T) {
    // Create a test handler
    testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("test response"))
    })
    
    // Wrap with middleware
    handler := loggingMiddleware(testHandler)
    
    // Create test request and recorder
    req := httptest.NewRequest("GET", "/test", nil)
    w := httptest.NewRecorder()
    
    // Execute
    handler.ServeHTTP(w, req)
    
    // Assert response
    if w.Code != http.StatusOK {
        t.Errorf("Expected status 200, got %d", w.Code)
    }
    
    if w.Body.String() != "test response" {
        t.Errorf("Expected 'test response', got %s", w.Body.String())
    }
}
```

### Testing Authentication Middleware

```go
func TestAuthMiddleware(t *testing.T) {
    tests := []struct {
        name           string
        authHeader     string
        expectedStatus int
        expectNext     bool
    }{
        {
            name:           "valid token",
            authHeader:     "Bearer valid-token",
            expectedStatus: http.StatusOK,
            expectNext:     true,
        },
        {
            name:           "missing token",
            authHeader:     "",
            expectedStatus: http.StatusUnauthorized,
            expectNext:     false,
        },
        {
            name:           "invalid format",
            authHeader:     "InvalidFormat",
            expectedStatus: http.StatusUnauthorized,
            expectNext:     false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            nextCalled := false
            testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                nextCalled = true
                w.WriteHeader(http.StatusOK)
            })
            
            handler := authMiddleware(testHandler)
            
            req := httptest.NewRequest("GET", "/", nil)
            if tt.authHeader != "" {
                req.Header.Set("Authorization", tt.authHeader)
            }
            w := httptest.NewRecorder()
            
            handler.ServeHTTP(w, req)
            
            if w.Code != tt.expectedStatus {
                t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
            }
            
            if nextCalled != tt.expectNext {
                t.Errorf("Expected nextCalled=%v, got %v", tt.expectNext, nextCalled)
            }
        })
    }
}
```

### Testing Context Values

```go
func TestUserIDMiddleware(t *testing.T) {
    expectedUserID := "user123"
    
    testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        userID, ok := r.Context().Value(userIDKey).(string)
        
        if !ok {
            t.Error("userID not found in context")
        }
        
        if userID != expectedUserID {
            t.Errorf("Expected userID %s, got %s", expectedUserID, userID)
        }
        
        w.WriteHeader(http.StatusOK)
    })
    
    handler := userIDMiddleware(testHandler)
    
    req := httptest.NewRequest("GET", "/", nil)
    req.Header.Set("X-User-ID", expectedUserID)
    w := httptest.NewRecorder()
    
    handler.ServeHTTP(w, req)
}
```

### Integration Testing Middleware Chain

```go
func TestMiddlewareChain(t *testing.T) {
    executionOrder := []string{}
    
    middleware1 := func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            executionOrder = append(executionOrder, "m1-before")
            next.ServeHTTP(w, r)
            executionOrder = append(executionOrder, "m1-after")
        })
    }
    
    middleware2 := func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            executionOrder = append(executionOrder, "m2-before")
            next.ServeHTTP(w, r)
            executionOrder = append(executionOrder, "m2-after")
        })
    }
    
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        executionOrder = append(executionOrder, "handler")
    })
    
    chain := middleware1(middleware2(handler))
    
    req := httptest.NewRequest("GET", "/", nil)
    w := httptest.NewRecorder()
    
    chain.ServeHTTP(w, req)
    
    expected := []string{"m1-before", "m2-before", "handler", "m2-after", "m1-after"}
    
    if !reflect.DeepEqual(executionOrder, expected) {
        t.Errorf("Expected order %v, got %v", expected, executionOrder)
    }
}
```

---

## 13. Production Patterns {#production-patterns}

### Structured Logging

```go
import "go.uber.org/zap"

func structuredLoggingMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()
            
            wrapped := wrapResponseWriter(w)
            
            // Add request ID to logger
            requestID := r.Context().Value("requestID").(string)
            reqLogger := logger.With(zap.String("request_id", requestID))
            
            // Add logger to context
            ctx := context.WithValue(r.Context(), "logger", reqLogger)
            
            next.ServeHTTP(wrapped, r.WithContext(ctx))
            
            duration := time.Since(start)
            
            reqLogger.Info("HTTP request",
                zap.String("method", r.Method),
                zap.String("path", r.URL.Path),
                zap.Int("status", wrapped.status),
                zap.Duration("duration", duration),
                zap.String("user_agent", r.UserAgent()),
                zap.String("ip", r.RemoteAddr),
            )
        })
    }
}
```

### Circuit Breaker Middleware

```go
import "github.com/sony/gobreaker"

func circuitBreakerMiddleware(cb *gobreaker.CircuitBreaker) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            _, err := cb.Execute(func() (interface{}, error) {
                wrapped := wrapResponseWriter(w)
                next.ServeHTTP(wrapped, r)
                
                if wrapped.status >= 500 {
                    return nil, fmt.Errorf("server error: %d", wrapped.status)
                }
                
                return nil, nil
            })
            
            if err != nil {
                http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
                return
            }
        })
    }
}

// Usage
cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
    Name:        "HTTP",
    MaxRequests: 3,
    Interval:    time.Minute,
    Timeout:     time.Minute * 5,
})

http.Handle("/", circuitBreakerMiddleware(cb)(handler))
```

### Security Headers Middleware

```go
func securityHeadersMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Prevent MIME type sniffing
        w.Header().Set("X-Content-Type-Options", "nosniff")
        
        // Prevent clickjacking
        w.Header().Set("X-Frame-Options", "DENY")
        
        // Enable XSS protection
        w.Header().Set("X-XSS-Protection", "1; mode=block")
        
        // Strict transport security
        w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
        
        // Content security policy
        w.Header().Set("Content-Security-Policy", "default-src 'self'")
        
        // Referrer policy
        w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
        
        // Permissions policy
        w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
        
        next.ServeHTTP(w, r)
    })
}
```

### Request Validation Middleware

```go
func validateJSONMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.Method == "POST" || r.Method == "PUT" || r.Method == "PATCH" {
            contentType := r.Header.Get("Content-Type")
            
            if !strings.HasPrefix(contentType, "application/json") {
                http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
                return
            }
            
            // Limit request body size
            r.Body = http.MaxBytesReader(w, r.Body, 1048576) // 1MB
        }
        
        next.ServeHTTP(w, r)
    })
}
```

### Distributed Tracing

```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/trace"
)

func tracingMiddleware(tracer trace.Tracer) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            ctx, span := tracer.Start(r.Context(), r.URL.Path)
            defer span.End()
            
            span.SetAttributes(
                attribute.String("http.method", r.Method),
                attribute.String("http.url", r.URL.String()),
                attribute.String("http.user_agent", r.UserAgent()),
            )
            
            wrapped := wrapResponseWriter(w)
            
            next.ServeHTTP(wrapped, r.WithContext(ctx))
            
            span.SetAttributes(attribute.Int("http.status_code", wrapped.status))
            
            if wrapped.status >= 500 {
                span.SetStatus(codes.Error, "Server Error")
            }
        })
    }
}
```

---

## 14. Real-World Examples {#real-world}

### Complete API Server with Middleware

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "time"
    
    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
)

type Server struct {
    router *chi.Mux
    logger *log.Logger
}

func NewServer(logger *log.Logger) *Server {
    s := &Server{
        router: chi.NewRouter(),
        logger: logger,
    }
    
    s.setupMiddleware()
    s.setupRoutes()
    
    return s
}

func (s *Server) setupMiddleware() {
    // Core middleware (order matters!)
    s.router.Use(middleware.RequestID)
    s.router.Use(middleware.RealIP)
    s.router.Use(s.loggingMiddleware())
    s.router.Use(middleware.Recoverer)
    s.router.Use(securityHeadersMiddleware)
    s.router.Use(middleware.Timeout(60 * time.Second))
}

func (s *Server) setupRoutes() {
    // Public routes
    s.router.Get("/health", s.healthHandler)
    s.router.Post("/login", s.loginHandler)
    
    // Protected API routes
    s.router.Route("/api", func(r chi.Router) {
        r.Use(s.authMiddleware)
        r.Use(rateLimitMiddleware(100))
        
        r.Get("/users", s.listUsersHandler)
        r.Post("/users", s.createUserHandler)
        r.Get("/users/{id}", s.getUserHandler)
        
        // Admin only
        r.Group(func(r chi.Router) {
            r.Use(s.adminMiddleware)
            r.Get("/admin/stats", s.statsHandler)
        })
    })
}

func (s *Server) loggingMiddleware() func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()
            wrapped := wrapResponseWriter(w)
            
            next.ServeHTTP(wrapped, r)
            
            s.logger.Printf(
                "method=%s path=%s status=%d duration=%v request_id=%s",
                r.Method,
                r.URL.Path,
                wrapped.status,
                time.Since(start),
                middleware.GetReqID(r.Context()),
            )
        })
    }
}

func (s *Server) authMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := r.Header.Get("Authorization")
        
        if token == "" {
            s.jsonError(w, "Missing authorization", http.StatusUnauthorized)
            return
        }
        
        // Validate token and extract user
        userID := "user123" // Simplified
        
        ctx := context.WithValue(r.Context(), "userID", userID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func (s *Server) adminMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        userID := r.Context().Value("userID").(string)
        
        // Check if user is admin
        if !s.isAdmin(userID) {
            s.jsonError(w, "Forbidden", http.StatusForbidden)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
    s.jsonResponse(w, map[string]string{"status": "healthy"}, http.StatusOK)
}

func (s *Server) listUsersHandler(w http.ResponseWriter, r *http.Request) {
    users := []map[string]string{
        {"id": "1", "name": "John"},
        {"id": "2", "name": "Jane"},
    }
    s.jsonResponse(w, users, http.StatusOK)
}

func (s *Server) jsonResponse(w http.ResponseWriter, data interface{}, status int) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(data)
}

func (s *Server) jsonError(w http.ResponseWriter, message string, status int) {
    s.jsonResponse(w, map[string]string{"error": message}, status)
}

func (s *Server) isAdmin(userID string) bool {
    // Implement admin check
    return true
}

func (s *Server) loginHandler(w http.ResponseWriter, r *http.Request) {}
func (s *Server) createUserHandler(w http.ResponseWriter, r *http.Request) {}
func (s *Server) getUserHandler(w http.ResponseWriter, r *http.Request) {}
func (s *Server) statsHandler(w http.ResponseWriter, r *http.Request) {}

func main() {
    logger := log.New(os.Stdout, "[API] ", log.LstdFlags)
    server := NewServer(logger)
    
    logger.Println("Server starting on :8080")
    if err := http.ListenAndServe(":8080", server.router); err != nil {
        logger.Fatal(err)
    }
}
```

### Microservice with Observability

```go
package main

import (
    "context"
    "net/http"
    
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "go.opentelemetry.io/otel"
    "go.uber.org/zap"
)

type ObservableServer struct {
    logger *zap.Logger
    tracer trace.Tracer
}

func NewObservableServer() (*ObservableServer, error) {
    logger, _ := zap.NewProduction()
    tracer := otel.Tracer("api-service")
    
    return &ObservableServer{
        logger: logger,
        tracer: tracer,
    }, nil
}

func (s *ObservableServer) setupRouter() http.Handler {
    mux := http.NewServeMux()
    
    // Metrics endpoint
    mux.Handle("/metrics", promhttp.Handler())
    
    // Application endpoints
    handler := http.HandlerFunc(s.businessHandler)
    
    // Apply middleware stack
    handler = s.tracingMiddleware(handler)
    handler = s.metricsMiddleware(handler)
    handler = s.loggingMiddleware(handler)
    
    mux.Handle("/api/", handler)
    
    return mux
}

func (s *ObservableServer) loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ctx := r.Context()
        
        // Extract trace ID for correlation
        span := trace.SpanFromContext(ctx)
        traceID := span.SpanContext().TraceID().String()
        
        // Create request-scoped logger
        reqLogger := s.logger.With(
            zap.String("trace_id", traceID),
            zap.String("method", r.Method),
            zap.String("path", r.URL.Path),
        )
        
        ctx = context.WithValue(ctx, "logger", reqLogger)
        
        reqLogger.Info("Request started")
        
        next.ServeHTTP(w, r.WithContext(ctx))
        
        reqLogger.Info("Request completed")
    })
}

func (s *ObservableServer) businessHandler(w http.ResponseWriter, r *http.Request) {
    // Business logic here
    w.Write([]byte("OK"))
}
```

---

## 15. Anti-Patterns and Pitfalls {#anti-patterns}

### Anti-Pattern 1: Writing Headers After Body

```go
// WRONG
func badHandler(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("response"))
    w.Header().Set("X-Custom", "value") // Too late!
    w.WriteHeader(http.StatusOK)        // Does nothing
}

// CORRECT
func goodHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("X-Custom", "value")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("response"))
}
```

### Anti-Pattern 2: Not Calling Next Handler

```go
// WRONG - Blocks all requests
func badMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Do some work
        log.Println("Request received")
        // Forgot to call next.ServeHTTP!
    })
}

// CORRECT
func goodMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        log.Println("Request received")
        next.ServeHTTP(w, r) // Always call next!
    })
}
```

### Anti-Pattern 3: Goroutine Leaks

```go
// WRONG - Goroutine may leak
func badTimeoutMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        go func() {
            time.Sleep(5 * time.Second)
            // This might run after request is done!
            log.Println("Delayed log")
        }()
        next.ServeHTTP(w, r)
    })
}

// CORRECT - Use context cancellation
func goodTimeoutMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
        defer cancel()
        
        done := make(chan struct{})
        go func() {
            select {
            case <-ctx.Done():
                log.Println("Request cancelled")
            case <-done:
                log.Println("Request completed")
            }
        }()
        
        next.ServeHTTP(w, r.WithContext(ctx))
        close(done)
    })
}
```

### Anti-Pattern 4: Mutating Request

```go
// WRONG - Modifying shared request
func badMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        r.Header.Set("X-Modified", "yes") // Modifies original!
        next.ServeHTTP(w, r)
    })
}

// CORRECT - Clone if you need to modify
func goodMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Create shallow copy
        r2 := r.Clone(r.Context())
        r2.Header.Set("X-Modified", "yes")
        next.ServeHTTP(w, r2)
    })
}
```

### Anti-Pattern 5: Storing Pointers in Context

```go
// WRONG - Race condition potential
type Stats struct {
    Count int
}

func badMiddleware(next http.Handler) http.Handler {
    stats := &Stats{}
    
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ctx := context.WithValue(r.Context(), "stats", stats)
        stats.Count++ // Race condition!
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// CORRECT - Use atomic or mutex
func goodMiddleware(next http.Handler) http.Handler {
    var count int64
    
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        atomic.AddInt64(&count, 1)
        ctx := context.WithValue(r.Context(), "count", atomic.LoadInt64(&count))
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

### Anti-Pattern 6: Ignoring Errors

```go
// WRONG
func badMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        body, _ := io.ReadAll(r.Body) // Ignoring error!
        // Use body...
        next.ServeHTTP(w, r)
    })
}

// CORRECT
func goodMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        body, err := io.ReadAll(r.Body)
        if err != nil {
            http.Error(w, "Bad request", http.StatusBadRequest)
            return
        }
        // Use body...
        next.ServeHTTP(w, r)
    })
}
```

### Common Pitfalls Summary

1. **Order matters** - Middleware executes in order of application
2. **Always call next** - Unless you intentionally want to short-circuit
3. **Headers before body** - Set headers before writing response
4. **Context immutability** - Don't modify request, use context for values
5. **Goroutine safety** - Be careful with shared state
6. **Error handling** - Always handle errors appropriately
7. **Resource cleanup** - Use defer for cleanup (close, unlock, etc.)
8. **Testing** - Test middleware in isolation and in chains

---

## Conclusion

You now have a comprehensive understanding of Go middleware from basics to production-grade patterns. Key takeaways:

1. **Middleware is just function composition** - Functions wrapping http.Handler
2. **Use context for request-scoped values** - Never mutate the request
3. **Order matters** - Think through your middleware chain carefully
4. **Test thoroughly** - Middleware is infrastructure code
5. **Keep it simple** - Don't over-engineer unless you need it
6. **Performance matters** - Profile and optimize hot paths
7. **Security is critical** - Always validate, sanitize, and protect

Practice building your own middleware, study production codebases, and remember: good middleware is invisible to your application logic but invaluable to your infrastructure.

---

## Additional Resources

- [Official Go net/http documentation](https://pkg.go.dev/net/http)
- [Justinas/alice - Middleware chaining](https://github.com/justinas/alice)
- [Go-chi/chi - Lightweight router](https://github.com/go-chi/chi)
- [Gorilla toolkit](https://github.com/gorilla)
- [OWASP Security Headers](https://owasp.org/www-project-secure-headers/)

**Happy coding!** 🚀
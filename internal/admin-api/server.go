// Copyright 2025 James Ross
package adminapi

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/flyingrobots/go-redis-work-queue/internal/config"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Server represents the admin API server
type Server struct {
	cfg      *Config
	appCfg   *config.Config
	rdb      *redis.Client
	logger   *zap.Logger
	server   *http.Server
	auditLog *AuditLogger
}

// NewServer creates a new admin API server
func NewServer(cfg *Config, appCfg *config.Config, rdb *redis.Client, logger *zap.Logger) (*Server, error) {
	var auditLog *AuditLogger
	var err error

	if cfg.AuditEnabled {
		auditLog, err = NewAuditLogger(cfg.AuditLogPath, cfg.AuditRotateSize, cfg.AuditMaxBackups)
		if err != nil {
			return nil, fmt.Errorf("failed to create audit logger: %w", err)
		}
	}

	return &Server{
		cfg:      cfg,
		appCfg:   appCfg,
		rdb:      rdb,
		logger:   logger,
		auditLog: auditLog,
	}, nil
}

// Start starts the API server
func (s *Server) Start() error {
	handler := s.SetupRoutes()

	// Apply middleware chain
	handler = s.applyMiddleware(handler)

	s.server = &http.Server{
		Addr:         s.cfg.ListenAddr,
		Handler:      handler,
		ReadTimeout:  s.cfg.ReadTimeout,
		WriteTimeout: s.cfg.WriteTimeout,
	}

	s.logger.Info("Starting admin API server",
		zap.String("addr", s.cfg.ListenAddr),
		zap.Bool("auth_enabled", s.cfg.RequireAuth),
		zap.Bool("rate_limit_enabled", s.cfg.RateLimitEnabled))

	if s.cfg.TLSEnabled {
		return s.server.ListenAndServeTLS(s.cfg.TLSCertFile, s.cfg.TLSKeyFile)
	}

	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	if s.auditLog != nil {
		s.auditLog.Close()
	}

	if s.server != nil {
		return s.server.Shutdown(ctx)
	}

	return nil
}

// SetupRoutes configures the API routes (exported for testing)
func (s *Server) SetupRoutes() http.Handler {
	mux := http.NewServeMux()
	h := NewHandler(s.appCfg, s.cfg, s.rdb, s.logger, s.auditLog)

	// Health check endpoint (no auth required)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	})

	// API v1 endpoints
	mux.HandleFunc("/api/v1/stats", methodHandler("GET", h.GetStats))
	mux.HandleFunc("/api/v1/stats/keys", methodHandler("GET", h.GetStatsKeys))
	mux.HandleFunc("/api/v1/queues/", func(w http.ResponseWriter, r *http.Request) {
		// Route based on path suffix
		path := r.URL.Path
		switch {
		case r.Method == "GET" && contains(path, "/peek"):
			h.PeekQueue(w, r)
		case r.Method == "DELETE" && contains(path, "/dlq"):
			h.PurgeDLQ(w, r)
		case r.Method == "DELETE" && contains(path, "/all"):
			h.PurgeAll(w, r)
		default:
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Endpoint not found")
		}
	})
	mux.HandleFunc("/api/v1/bench", methodHandler("POST", h.RunBenchmark))

	// OpenAPI spec endpoint
	mux.HandleFunc("/api/v1/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-yaml")
		w.Write([]byte(openAPISpec))
	})

	return mux
}

// applyMiddleware applies the middleware chain
func (s *Server) applyMiddleware(handler http.Handler) http.Handler {
	// Apply in reverse order (outermost first)

	// Recovery middleware (outermost)
	handler = RecoveryMiddleware(s.logger)(handler)

	// Request ID middleware
	handler = RequestIDMiddleware()(handler)

	// CORS middleware
	if s.cfg.CORSEnabled {
		handler = CORSMiddleware(s.cfg.CORSAllowOrigins)(handler)
	}

	// Audit middleware
	if s.cfg.AuditEnabled && s.auditLog != nil {
		handler = AuditMiddleware(s.auditLog, s.logger)(handler)
	}

	// Rate limiting middleware
	if s.cfg.RateLimitEnabled {
		handler = RateLimitMiddleware(s.cfg.RateLimitPerMinute, s.cfg.RateLimitBurst, s.logger)(handler)
	}

	// Auth middleware
	if s.cfg.RequireAuth {
		handler = AuthMiddleware(s.cfg.JWTSecret, s.cfg.DenyByDefault, s.logger)(handler)
	}

	return handler
}

// Helper function to create method-specific handlers
func methodHandler(method string, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
			return
		}
		handler(w, r)
	}
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
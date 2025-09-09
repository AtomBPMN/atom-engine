/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package restapi

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"atom-engine/src/core/auth"
	"atom-engine/src/core/interfaces"
	"atom-engine/src/core/logger"
	"atom-engine/src/core/restapi/handlers"
	"atom-engine/src/core/restapi/middleware"
	"atom-engine/src/core/restapi/models"
)

// Config holds REST API server configuration
type Config struct {
	Host      string                      `yaml:"host"`
	Port      int                         `yaml:"port"`
	CORS      *middleware.CORSConfig      `yaml:"cors"`
	Logging   *middleware.LoggingConfig   `yaml:"logging"`
	RateLimit *middleware.RateLimitConfig `yaml:"rate_limit"`
	Swagger   *SwaggerConfig              `yaml:"swagger"`
}

// SwaggerConfig holds Swagger documentation configuration
type SwaggerConfig struct {
	Enabled bool   `yaml:"enabled"`
	Path    string `yaml:"path"`
	Title   string `yaml:"title"`
	Version string `yaml:"version"`
}

// DefaultConfig returns default REST API configuration
func DefaultConfig() *Config {
	return &Config{
		Host:      "localhost",
		Port:      27555,
		CORS:      middleware.DefaultCORSConfig(),
		Logging:   middleware.DefaultLoggingConfig(),
		RateLimit: middleware.DefaultRateLimitConfig(),
		Swagger: &SwaggerConfig{
			Enabled: true,
			Path:    "/api/docs",
			Title:   "Atom Engine REST API",
			Version: "1.0.0",
		},
	}
}

// Server represents REST API server
type Server struct {
	config        *Config
	httpServer    *http.Server
	router        *gin.Engine
	coreInterface CoreInterface
	authComponent auth.Component

	// Middleware instances
	authMiddleware      *middleware.AuthMiddleware
	corsMiddleware      *middleware.CORSMiddleware
	loggingMiddleware   *middleware.LoggingMiddleware
	rateLimitMiddleware *middleware.RateLimitMiddleware

	// Handler instances
	storageHandler    *handlers.StorageHandler
	parserHandler     *handlers.ParserHandler
	processHandler    *handlers.ProcessHandler
	tokensHandler     *handlers.TokensHandler
	timerHandler      *handlers.TimerHandler
	jobsHandler       *handlers.JobsHandler
	messagesHandler   *handlers.MessagesHandler
	expressionHandler *handlers.ExpressionHandler
	incidentsHandler  *handlers.IncidentsHandler
	systemHandler     *handlers.SystemHandler
}

// Import the unified core interface (with typed support)
// Импортируем унифицированный интерфейс core (с поддержкой типизации)
type CoreInterface = interfaces.CoreTypedInterface

// Response types (simplified for REST)
type StorageStatusResponse struct {
	IsConnected   bool   `json:"is_connected"`
	IsHealthy     bool   `json:"is_healthy"`
	Status        string `json:"status"`
	UptimeSeconds int64  `json:"uptime_seconds"`
}

type StorageInfoResponse struct {
	TotalSizeBytes int64             `json:"total_size_bytes"`
	UsedSizeBytes  int64             `json:"used_size_bytes"`
	FreeSizeBytes  int64             `json:"free_size_bytes"`
	TotalKeys      int64             `json:"total_keys"`
	DatabasePath   string            `json:"database_path"`
	Statistics     map[string]string `json:"statistics"`
}

type TimewheelStatsResponse struct {
	TotalTimers     int32            `json:"total_timers"`
	PendingTimers   int32            `json:"pending_timers"`
	FiredTimers     int32            `json:"fired_timers"`
	CancelledTimers int32            `json:"cancelled_timers"`
	CurrentTick     int64            `json:"current_tick"`
	SlotsCount      int32            `json:"slots_count"`
	TimerTypes      map[string]int32 `json:"timer_types"`
}

type TimersListResponse struct {
	Timers     []TimerInfo `json:"timers"`
	TotalCount int32       `json:"total_count"`
}

type TimerInfo struct {
	TimerID           string `json:"timer_id"`
	ElementID         string `json:"element_id"`
	ProcessInstanceID string `json:"process_instance_id"`
	TimerType         string `json:"timer_type"`
	Status            string `json:"status"`
	ScheduledAt       int64  `json:"scheduled_at"`
	CreatedAt         int64  `json:"created_at"`
	TimeDuration      string `json:"time_duration"`
	TimeCycle         string `json:"time_cycle"`
	RemainingSeconds  int64  `json:"remaining_seconds"`
	WheelLevel        int32  `json:"wheel_level"`
}

// Interfaces (same as gRPC)
type TimewheelComponentInterface interface {
	ProcessMessage(ctx context.Context, messageJSON string) error
	GetResponseChannel() <-chan string
	GetTimerInfo(timerID string) (level int, remainingSeconds int64, found bool)
}

type StorageComponentInterface interface {
	LoadAllTokens() ([]*Token, error)
	LoadTokensByState(state TokenState) ([]*Token, error)
	LoadToken(tokenID string) (*Token, error)
}

type ProcessComponentInterface interface {
	StartProcessInstance(processKey string, variables map[string]interface{}) (*ProcessInstanceResult, error)
	GetProcessInstanceStatus(instanceID string) (*ProcessInstanceResult, error)
	CancelProcessInstance(instanceID string, reason string) error
	ListProcessInstances(statusFilter string, processKeyFilter string, limit int) ([]*ProcessInstanceResult, error)
	GetActiveTokens(instanceID string) ([]*Token, error)
}

// Placeholder types
type Token struct {
	ID                string     `json:"id"`
	State             TokenState `json:"state"`
	ElementID         string     `json:"element_id"`
	ProcessInstanceID string     `json:"process_instance_id"`
}

type TokenState string

const (
	TokenStateActive    TokenState = "ACTIVE"
	TokenStateCompleted TokenState = "COMPLETED"
	TokenStateCancelled TokenState = "CANCELLED"
)

type ProcessInstanceResult struct {
	InstanceID      string                 `json:"instance_id"`
	ProcessID       string                 `json:"process_id"`
	ProcessName     string                 `json:"process_name"`
	State           string                 `json:"state"`
	CurrentActivity string                 `json:"current_activity"`
	StartedAt       int64                  `json:"started_at"`
	UpdatedAt       int64                  `json:"updated_at"`
	CompletedAt     int64                  `json:"completed_at,omitempty"`
	Variables       map[string]interface{} `json:"variables"`
}

// NewServer creates new REST API server instance
func NewServer(config *Config, coreInterface CoreInterface) *Server {
	if config == nil {
		config = DefaultConfig()
	}

	server := &Server{
		config:        config,
		coreInterface: coreInterface,
	}

	// Get auth component from core if available
	if authComp := coreInterface.GetAuthComponent(); authComp != nil {
		if authComponent, ok := authComp.(auth.Component); ok {
			server.authComponent = authComponent
		}
	}

	server.setupHandlers()
	server.setupRouter()
	return server
}

// setupHandlers initializes all request handlers
func (s *Server) setupHandlers() {
	s.storageHandler = handlers.NewStorageHandler(s.coreInterface)
	s.parserHandler = handlers.NewParserHandler(s.coreInterface)
	s.processHandler = handlers.NewProcessHandler(s.coreInterface)
	s.tokensHandler = handlers.NewTokensHandler(s.coreInterface)
	s.timerHandler = handlers.NewTimerHandler(s.coreInterface)
	s.jobsHandler = handlers.NewJobsHandler(s.coreInterface)
	s.messagesHandler = handlers.NewMessagesHandler(s.coreInterface)
	s.expressionHandler = handlers.NewExpressionHandler(s.coreInterface)
	s.incidentsHandler = handlers.NewIncidentsHandler(s.coreInterface)
	s.systemHandler = handlers.NewSystemHandler(s.coreInterface)
}

// setupRouter configures Gin router and middleware
func (s *Server) setupRouter() {
	// Set Gin mode based on log level
	gin.SetMode(gin.ReleaseMode) // Default to release mode

	// Create router
	s.router = gin.New()

	// Setup middleware
	s.setupMiddleware()

	// Setup routes
	s.setupRoutes()
}

// setupMiddleware configures all middleware
func (s *Server) setupMiddleware() {
	// Recovery middleware (built-in)
	s.router.Use(gin.Recovery())

	// CORS middleware
	if s.config.CORS != nil {
		s.corsMiddleware = middleware.NewCORSMiddleware(s.config.CORS)
		s.router.Use(s.corsMiddleware.Handler())
	}

	// Logging middleware
	if s.config.Logging != nil {
		s.loggingMiddleware = middleware.NewLoggingMiddleware(s.config.Logging)
		s.router.Use(s.loggingMiddleware.Handler())
	}

	// Rate limiting middleware
	if s.config.RateLimit != nil {
		s.rateLimitMiddleware = middleware.NewRateLimitMiddleware(s.config.RateLimit, s.authComponent)
		s.router.Use(s.rateLimitMiddleware.Handler())
	}

	// Auth middleware
	if s.authComponent != nil {
		s.authMiddleware = middleware.NewAuthMiddleware(s.authComponent)
		s.router.Use(s.authMiddleware.Authenticate())
	}
}

// setupRoutes configures all API routes
func (s *Server) setupRoutes() {
	// Health check endpoint (no auth required)
	s.router.GET("/health", s.healthHandler)

	// API v1 routes
	v1 := s.router.Group("/api/v1")
	{
		// Daemon management (basic handlers)
		daemon := v1.Group("/daemon")
		{
			daemon.GET("/status", s.daemonStatusHandler)
			daemon.POST("/start", s.daemonStartHandler)
			daemon.POST("/stop", s.daemonStopHandler)
			daemon.GET("/events", s.daemonEventsHandler)
		}

		// Register handlers with their routes
		s.storageHandler.RegisterRoutes(v1, s.authMiddleware)
		s.parserHandler.RegisterRoutes(v1, s.authMiddleware)
		s.processHandler.RegisterRoutes(v1, s.authMiddleware)
		s.tokensHandler.RegisterRoutes(v1, s.authMiddleware)
		s.timerHandler.RegisterRoutes(v1, s.authMiddleware)
		s.jobsHandler.RegisterRoutes(v1, s.authMiddleware)
		s.messagesHandler.RegisterRoutes(v1, s.authMiddleware)
		s.expressionHandler.RegisterRoutes(v1, s.authMiddleware)
		s.incidentsHandler.RegisterRoutes(v1, s.authMiddleware)
		s.systemHandler.RegisterRoutes(v1, s.authMiddleware)
	}

	// Swagger documentation
	if s.config.Swagger != nil && s.config.Swagger.Enabled {
		s.router.GET(s.config.Swagger.Path, s.swaggerHandler)
		s.router.Static(s.config.Swagger.Path+"/static", "./docs/swagger")
	}
}

// Start starts the REST API server
func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)

	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	logger.Info("Starting REST API server",
		logger.String("address", addr),
		logger.Int("port", s.config.Port))

	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("REST API server failed", logger.String("error", err.Error()))
		}
	}()

	return nil
}

// Stop stops the REST API server
func (s *Server) Stop() error {
	if s.httpServer == nil {
		return nil
	}

	logger.Info("Stopping REST API server")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return s.httpServer.Shutdown(ctx)
}

// IsReady returns server ready status
func (s *Server) IsReady() bool {
	return s.httpServer != nil
}

// Basic handlers (more will be in separate handler files)

// healthHandler handles health check requests
func (s *Server) healthHandler(c *gin.Context) {
	response := models.HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Checks: map[string]interface{}{
			"server": "ok",
		},
	}

	c.JSON(http.StatusOK, models.SuccessResponse(response, "health"))
}

// swaggerHandler serves Swagger documentation
func (s *Server) swaggerHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "swagger.html", gin.H{
		"title":   s.config.Swagger.Title,
		"version": s.config.Swagger.Version,
	})
}

// Daemon handlers - implemented REST endpoints
func (s *Server) daemonStatusHandler(c *gin.Context) {
	requestID := s.getRequestID(c)

	// Get system status from core
	status, err := s.coreInterface.GetSystemStatus()
	if err != nil {
		logger.Error("Failed to get daemon status", logger.String("error", err.Error()))
		apiErr := models.InternalServerError("Failed to get daemon status")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(apiErr, requestID))
		return
	}

	// Return daemon status information
	response := map[string]interface{}{
		"status":     status.Status,
		"health":     status.Health,
		"uptime":     status.Uptime,
		"version":    status.Version,
		"components": status.ComponentsTotal,
	}

	c.JSON(http.StatusOK, models.SuccessResponse(response, requestID))
}

func (s *Server) daemonStartHandler(c *gin.Context) {
	requestID := s.getRequestID(c)

	// Daemon start/stop operations are managed at system level, not through REST API
	// These operations affect the entire system and should be done through CLI
	response := map[string]interface{}{
		"message": "Daemon is already running if you can reach this endpoint",
		"note":    "Use CLI 'atomd start' to start daemon, 'atomd stop' to stop",
		"status":  "running",
	}

	c.JSON(http.StatusOK, models.SuccessResponse(response, requestID))
}

func (s *Server) daemonStopHandler(c *gin.Context) {
	requestID := s.getRequestID(c)

	// Graceful shutdown would stop this endpoint itself, so we can't implement it here
	// Direct shutdown should be done through CLI or system signals
	response := map[string]interface{}{
		"message": "Graceful shutdown should be initiated through CLI or system signals",
		"note":    "Use CLI 'atomd stop' for graceful shutdown",
		"warning": "REST endpoint cannot stop itself",
	}

	c.JSON(http.StatusOK, models.SuccessResponse(response, requestID))
}

func (s *Server) daemonEventsHandler(c *gin.Context) {
	requestID := s.getRequestID(c)

	// Get system events from storage through core
	storageComp := s.coreInterface.GetStorageTyped()
	if storageComp == nil {
		apiErr := models.InternalServerError("Storage component not available")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(apiErr, requestID))
		return
	}

	// For now, return system status and info as "events"
	// This can be enhanced later to get actual event logs from storage
	status, err := s.coreInterface.GetSystemStatus()
	if err != nil {
		logger.Error("Failed to get system status for events", logger.String("error", err.Error()))
		apiErr := models.InternalServerError("Failed to get system events")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(apiErr, requestID))
		return
	}

	info, err := s.coreInterface.GetSystemInfo()
	if err != nil {
		logger.Error("Failed to get system info for events", logger.String("error", err.Error()))
		apiErr := models.InternalServerError("Failed to get system events")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(apiErr, requestID))
		return
	}

	// Return recent system events information
	response := map[string]interface{}{
		"events": []map[string]interface{}{
			{
				"type":      "system_status",
				"timestamp": time.Now(),
				"status":    status.Status,
				"health":    status.Health,
				"message":   "System status check",
			},
			{
				"type":      "system_info",
				"timestamp": time.Now(),
				"version":   info.Version,
				"uptime":    info.Uptime,
				"message":   "System information",
			},
		},
		"note": "Use CLI 'atomd events' for detailed event logs from storage",
	}

	c.JSON(http.StatusOK, models.SuccessResponse(response, requestID))
}

// getRequestID extracts request ID from context or generates one
func (s *Server) getRequestID(c *gin.Context) string {
	if requestID := c.GetHeader("X-Request-ID"); requestID != "" {
		return requestID
	}
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

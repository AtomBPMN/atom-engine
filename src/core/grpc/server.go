/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package grpc

import (
	"context"
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"atom-engine/proto/jobs/jobspb"
	"atom-engine/proto/messages/messagespb"
	"atom-engine/proto/parser/parserpb"
	"atom-engine/proto/process/processpb"
	"atom-engine/proto/timewheel/timewheelpb"
	"atom-engine/src/core/logger"
	"atom-engine/src/core/models"
)

// Server represents gRPC server
// Представляет gRPC сервер
type Server struct {
	grpcServer *grpc.Server
	listener   net.Listener
	port       int
	core       CoreInterface
}

// CoreInterface defines methods that core must provide to gRPC
// Определяет методы которые core должен предоставить для gRPC
type CoreInterface interface {
	GetStorageStatus() (*StorageStatusResponse, error)
	GetStorageInfo() (*StorageInfoResponse, error)
	GetTimewheelComponent() TimewheelComponentInterface
	GetTimewheelStats() (*timewheelpb.GetTimeWheelStatsResponse, error)
	GetTimersList(statusFilter string, limit int32) (*timewheelpb.ListTimersResponse, error)

	GetProcessComponent() ProcessComponentInterface
	GetStorageComponent() StorageComponentInterface
	GetMessagesComponent() interface{}
	GetJobsComponent() interface{}
	GetParserComponent() interface{}
	GetExpressionComponent() interface{}
	GetStorage() interface{}

	// JSON Message Routing
	SendMessage(componentName, messageJSON string) error

	// Response Handling
	WaitForParserResponse(timeoutMs int) (string, error)
	WaitForJobsResponse(timeoutMs int) (string, error)
	WaitForMessagesResponse(timeoutMs int) (string, error)
}

// TimewheelComponentInterface defines timewheel component interface
// Определяет интерфейс timewheel компонента
type TimewheelComponentInterface interface {
	ProcessMessage(ctx context.Context, messageJSON string) error
	GetResponseChannel() <-chan string
	GetTimerInfo(timerID string) (level int, remainingSeconds int64, found bool)
}

// StorageComponentInterface defines storage component interface
// Определяет интерфейс storage компонента
type StorageComponentInterface interface {
	LoadAllTokens() ([]*models.Token, error)
	LoadTokensByState(state models.TokenState) ([]*models.Token, error)
	LoadToken(tokenID string) (*models.Token, error)
}

// ProcessComponentInterface defines process component interface
// Определяет интерфейс process компонента
type ProcessComponentInterface interface {
	StartProcessInstance(processKey string, variables map[string]interface{}) (*ProcessInstanceResult, error)
	GetProcessInstanceStatus(instanceID string) (*ProcessInstanceResult, error)
	CancelProcessInstance(instanceID string, reason string) error
	ListProcessInstances(statusFilter string, processKeyFilter string, limit int) ([]*ProcessInstanceResult, error)
	GetActiveTokens(instanceID string) ([]*models.Token, error)
}

// MessageStats represents message statistics
type MessageStats struct {
	TotalMessages         int32 `json:"total_messages"`
	BufferedMessages      int32 `json:"buffered_messages"`
	ExpiredMessages       int32 `json:"expired_messages"`
	PublishedToday        int32 `json:"published_today"`
	InstancesCreatedToday int32 `json:"instances_created_today"`
}

// MessagesComponentInterface defines messages component interface
// Определяет интерфейс messages компонента
type MessagesComponentInterface interface {
	PublishMessage(ctx context.Context, tenantID, messageName, correlationKey string, variables map[string]interface{}, ttl *time.Duration) (*models.MessageCorrelationResult, error)
	ListBufferedMessages(ctx context.Context, tenantID string, limit, offset int) ([]*models.BufferedMessage, error)
	ListMessageSubscriptions(ctx context.Context, tenantID string, limit, offset int) ([]*models.ProcessMessageSubscription, error)
	GetMessageStats(ctx context.Context, tenantID string) (*MessageStats, error)
	CleanupExpiredMessages(ctx context.Context) (int, error)
}

// JobsComponentInterface defines jobs component interface
// Определяет интерфейс jobs компонента
type JobsComponentInterface interface {
	CreateJob(jobType, processInstanceID string, variables map[string]interface{}) (string, error)
	GetJobStats() (interface{}, error)
}

// ProcessInstanceResult represents process instance result for gRPC
// Представляет результат экземпляра процесса для gRPC
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

// Config holds gRPC server configuration
// Конфигурация gRPC сервера
type Config struct {
	Port int `yaml:"port"`
}

// NewServer creates new gRPC server instance
// Создает новый экземпляр gRPC сервера
func NewServer(config *Config, core CoreInterface) *Server {
	return &Server{
		port: config.Port,
		core: core,
	}
}

// Start starts gRPC server
// Запускает gRPC сервер
func (s *Server) Start() error {
	logger.Info("Starting gRPC server", logger.Int("port", s.port))

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", s.port, err)
	}
	s.listener = listener

	s.grpcServer = grpc.NewServer()

	// Register storage service
	RegisterStorageServiceServer(s.grpcServer, &storageServiceServer{core: s.core})

	// Register timewheel service
	timewheelpb.RegisterTimeWheelServiceServer(s.grpcServer, &timewheelServiceServer{core: s.core})
	parserpb.RegisterParserServiceServer(s.grpcServer, NewParserService(s.core))

	// Register process service
	processpb.RegisterProcessServiceServer(s.grpcServer, &processServiceServer{core: s.core})

	// Register messages service
	messagespb.RegisterMessagesServiceServer(s.grpcServer, &messagesServiceServer{core: s.core})

	// Register jobs service
	jobspb.RegisterJobsServiceServer(s.grpcServer, &jobsServiceServer{core: s.core})

	// Enable reflection for development
	reflection.Register(s.grpcServer)

	logger.Info("gRPC server started successfully", logger.Int("port", s.port))

	go func() {
		if err := s.grpcServer.Serve(listener); err != nil {
			logger.Error("gRPC server failed", logger.String("error", err.Error()))
		}
	}()

	return nil
}

// Stop stops gRPC server
// Останавливает gRPC сервер
func (s *Server) Stop() error {
	if s.grpcServer != nil {
		logger.Info("Stopping gRPC server")
		s.grpcServer.GracefulStop()
	}
	if s.listener != nil {
		s.listener.Close()
	}
	return nil
}

// IsReady returns server ready status
// Возвращает статус готовности сервера
func (s *Server) IsReady() bool {
	return s.grpcServer != nil && s.listener != nil
}

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

	"atom-engine/proto/expression/expressionpb"
	"atom-engine/proto/incidents/incidentspb"
	"atom-engine/proto/jobs/jobspb"
	"atom-engine/proto/messages/messagespb"
	"atom-engine/proto/parser/parserpb"
	"atom-engine/proto/process/processpb"
	"atom-engine/proto/timewheel/timewheelpb"
	"atom-engine/src/core/auth"
	"atom-engine/src/core/interfaces"
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
// Import the unified core interface
// Импортируем унифицированный интерфейс core
type CoreInterface = interfaces.CoreInterface

// Use interfaces from the unified interfaces package
// Используем интерфейсы из унифицированного пакета interfaces
type TimewheelComponentInterface = interfaces.TimewheelComponentInterface
type StorageComponentInterface = interfaces.StorageComponentInterface
type ProcessComponentInterface = interfaces.ProcessComponentInterface

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
	PublishMessage(
		ctx context.Context,
		tenantID, messageName, correlationKey string,
		variables map[string]interface{},
		ttl *time.Duration,
	) (*models.MessageCorrelationResult, error)
	ListBufferedMessages(
		ctx context.Context,
		tenantID string,
		limit, offset int,
	) ([]*models.BufferedMessage, error)
	ListMessageSubscriptions(
		ctx context.Context,
		tenantID string,
		limit, offset int,
	) ([]*models.ProcessMessageSubscription, error)
	GetMessageStats(ctx context.Context, tenantID string) (*MessageStats, error)
	CleanupExpiredMessages(ctx context.Context) (int, error)
}

// JobsComponentInterface defines jobs component interface
// Определяет интерфейс jobs компонента
type JobsComponentInterface interface {
	CreateJob(jobType, processInstanceID string, variables map[string]interface{}) (string, error)
	GetJobStats() (interface{}, error)
}

// Removed ProcessInstanceResult - now using unified interfaces

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

	// Setup interceptors
	var opts []grpc.ServerOption

	// Add auth interceptor if auth component is available
	if authComp := s.core.GetAuthComponent(); authComp != nil {
		if authComponent, ok := authComp.(auth.Component); ok {
			authInterceptor := NewAuthInterceptor(authComponent)
			opts = append(opts,
				grpc.UnaryInterceptor(authInterceptor.UnaryInterceptor()),
				grpc.StreamInterceptor(authInterceptor.StreamInterceptor()),
			)
			logger.Info("Auth interceptors enabled for gRPC server")
		}
	}

	s.grpcServer = grpc.NewServer(opts...)

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

	// Register incidents service
	incidentspb.RegisterIncidentsServiceServer(s.grpcServer, &incidentsServiceServer{core: s.core})

	// Register expression service
	expressionpb.RegisterExpressionServiceServer(s.grpcServer, &expressionServiceServer{core: s.core})

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

// GetLoopbackConnection returns a loopback gRPC connection to this server
// Возвращает loopback gRPC соединение к этому серверу
func (s *Server) GetLoopbackConnection() (*grpc.ClientConn, error) {
	if !s.IsReady() {
		return nil, fmt.Errorf("gRPC server is not ready")
	}

	// Create a connection to localhost on the server port
	// Создаем соединение к localhost на порту сервера
	target := fmt.Sprintf("localhost:%d", s.port)
	conn, err := grpc.Dial(target, grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("failed to create loopback connection to %s: %w", target, err)
	}

	return conn, nil
}

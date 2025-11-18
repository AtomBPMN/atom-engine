/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package grpc

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"atom-engine/src/core/auth"
	"atom-engine/src/core/logger"
)

// AuthInterceptor provides gRPC authentication interceptor
type AuthInterceptor struct {
	authComponent auth.Component
	bypassMethods []string // Methods that bypass authentication
}

// NewAuthInterceptor creates a new auth interceptor
func NewAuthInterceptor(authComponent auth.Component) *AuthInterceptor {
	return &AuthInterceptor{
		authComponent: authComponent,
		bypassMethods: []string{
			// Health check and status endpoints
			"/grpc.health.v1.Health/Check",
			"/grpc.health.v1.Health/Watch",
			// Add other methods that should bypass auth
		},
	}
}

// UnaryInterceptor returns unary server interceptor for authentication
func (ai *AuthInterceptor) UnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Check if auth is enabled
		if !ai.authComponent.IsEnabled() {
			// Auth disabled - allow all requests
			return handler(ctx, req)
		}

		// Check if method should bypass auth
		if ai.shouldBypassAuth(info.FullMethod) {
			return handler(ctx, req)
		}

		// Create auth context from gRPC request
		authCtx := auth.CreateAuthContextFromGRPC(ctx, info.FullMethod)

		// Validate auth context
		if err := auth.ValidateAuthContext(authCtx); err != nil {
			logger.Warn("Invalid auth context",
				logger.String("method", info.FullMethod),
				logger.String("error", err.Error()))
			return nil, status.Error(codes.InvalidArgument, "Invalid request context")
		}

		// Authenticate
		authResult, err := ai.authComponent.Authenticate(authCtx)
		if err != nil {
			logger.Error("Authentication error",
				logger.String("method", info.FullMethod),
				logger.String("client_ip", authCtx.ClientIP),
				logger.String("error", err.Error()))
			return nil, status.Error(codes.Internal, "Authentication failed")
		}

		if !authResult.Authenticated {
			logger.Warn("Authentication failed",
				logger.String("method", info.FullMethod),
				logger.String("client_ip", authCtx.ClientIP),
				logger.String("reason", authResult.Reason))

			// Return appropriate error based on reason
			switch {
			case strings.Contains(authResult.Reason, "Rate limit"):
				return nil, status.Error(codes.ResourceExhausted, "Rate limit exceeded")
			case strings.Contains(authResult.Reason, "IP"):
				return nil, status.Error(codes.PermissionDenied, "IP address not allowed")
			case strings.Contains(authResult.Reason, "API key"):
				return nil, status.Error(codes.Unauthenticated, "Invalid or missing API key")
			default:
				return nil, status.Error(codes.Unauthenticated, "Authentication failed")
			}
		}

		// Add auth result to context for downstream use
		newCtx := context.WithValue(ctx, authContextKey, authResult)

		logger.Debug("Request authenticated",
			logger.String("method", info.FullMethod),
			logger.String("client_ip", authCtx.ClientIP),
			logger.String("api_key_name", authResult.APIKeyName))

		return handler(newCtx, req)
	}
}

// StreamInterceptor returns stream server interceptor for authentication
func (ai *AuthInterceptor) StreamInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		stream grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		// Check if auth is enabled
		if !ai.authComponent.IsEnabled() {
			// Auth disabled - allow all requests
			return handler(srv, stream)
		}

		// Check if method should bypass auth
		if ai.shouldBypassAuth(info.FullMethod) {
			return handler(srv, stream)
		}

		// Create auth context from gRPC stream
		authCtx := auth.CreateAuthContextFromGRPC(stream.Context(), info.FullMethod)

		// Validate auth context
		if err := auth.ValidateAuthContext(authCtx); err != nil {
			logger.Warn("Invalid auth context for stream",
				logger.String("method", info.FullMethod),
				logger.String("error", err.Error()))
			return status.Error(codes.InvalidArgument, "Invalid request context")
		}

		// Authenticate
		authResult, err := ai.authComponent.Authenticate(authCtx)
		if err != nil {
			logger.Error("Stream authentication error",
				logger.String("method", info.FullMethod),
				logger.String("client_ip", authCtx.ClientIP),
				logger.String("error", err.Error()))
			return status.Error(codes.Internal, "Authentication failed")
		}

		if !authResult.Authenticated {
			logger.Warn("Stream authentication failed",
				logger.String("method", info.FullMethod),
				logger.String("client_ip", authCtx.ClientIP),
				logger.String("reason", authResult.Reason))

			// Return appropriate error based on reason
			switch {
			case strings.Contains(authResult.Reason, "Rate limit"):
				return status.Error(codes.ResourceExhausted, "Rate limit exceeded")
			case strings.Contains(authResult.Reason, "IP"):
				return status.Error(codes.PermissionDenied, "IP address not allowed")
			case strings.Contains(authResult.Reason, "API key"):
				return status.Error(codes.Unauthenticated, "Invalid or missing API key")
			default:
				return status.Error(codes.Unauthenticated, "Authentication failed")
			}
		}

		// Create wrapped stream with auth context
		wrappedStream := &authServerStream{
			ServerStream: stream,
			ctx:          context.WithValue(stream.Context(), authContextKey, authResult),
		}

		logger.Debug("Stream request authenticated",
			logger.String("method", info.FullMethod),
			logger.String("client_ip", authCtx.ClientIP),
			logger.String("api_key_name", authResult.APIKeyName))

		return handler(srv, wrappedStream)
	}
}

// shouldBypassAuth checks if method should bypass authentication
func (ai *AuthInterceptor) shouldBypassAuth(method string) bool {
	for _, bypassMethod := range ai.bypassMethods {
		if method == bypassMethod {
			return true
		}
	}
	return false
}

// AddBypassMethod adds a method to bypass authentication
func (ai *AuthInterceptor) AddBypassMethod(method string) {
	ai.bypassMethods = append(ai.bypassMethods, method)
}

// authServerStream wraps grpc.ServerStream with authentication context
type authServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

// Context returns the context with authentication information
func (s *authServerStream) Context() context.Context {
	return s.ctx
}

// Context key for storing auth result in context
type contextKey string

const authContextKey contextKey = "auth_result"

// GetAuthResultFromContext extracts auth result from context
func GetAuthResultFromContext(ctx context.Context) (*auth.AuthResult, bool) {
	authResult, ok := ctx.Value(authContextKey).(*auth.AuthResult)
	return authResult, ok
}

// RequirePermission checks if the authenticated user has required permission
func RequirePermission(ctx context.Context, permission string) error {
	authResult, ok := GetAuthResultFromContext(ctx)
	if !ok {
		return status.Error(codes.Internal, "Authentication context not found")
	}

	if !auth.HasPermission(authResult.Permissions, permission) {
		return status.Error(codes.PermissionDenied, "Insufficient permissions")
	}

	return nil
}

/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package auth

import (
	"context"
	"errors"
	"net"
	"strings"
	"time"

	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

var (
	ErrMissingClientIP = errors.New("missing client IP address")
)

// CreateAuthContextFromGRPC creates AuthContext from gRPC request
func CreateAuthContextFromGRPC(ctx context.Context, method string) AuthContext {
	authCtx := AuthContext{
		Protocol:    "grpc",
		Method:      method,
		RequestPath: method,
		Timestamp:   time.Now(),
	}

	// Extract client IP from peer info
	if p, ok := peer.FromContext(ctx); ok {
		if addr, ok := p.Addr.(*net.TCPAddr); ok {
			authCtx.ClientIP = addr.IP.String()
		} else {
			// Handle other address types or use string representation
			authCtx.ClientIP = p.Addr.String()
			// Remove port if present
			if host, _, err := net.SplitHostPort(authCtx.ClientIP); err == nil {
				authCtx.ClientIP = host
			}
		}
	}

	// Extract metadata
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		// Extract API key from authorization header
		if auth := md.Get("authorization"); len(auth) > 0 {
			authCtx.APIKey = extractBearerToken(auth[0])
		}

		// Extract user agent if available
		if ua := md.Get("user-agent"); len(ua) > 0 {
			authCtx.UserAgent = ua[0]
		}
	}

	return authCtx
}

// CreateAuthContextFromHTTP creates AuthContext from HTTP request (future use)
func CreateAuthContextFromHTTP(clientIP, userAgent, method, path, authHeader string) AuthContext {
	return AuthContext{
		ClientIP:    extractClientIP(clientIP),
		APIKey:      extractBearerToken(authHeader),
		UserAgent:   userAgent,
		RequestPath: path,
		Method:      method,
		Protocol:    "http",
		Timestamp:   time.Now(),
	}
}

// extractBearerToken extracts token from "Bearer <token>" format
func extractBearerToken(authHeader string) string {
	const bearerPrefix = "Bearer "
	if strings.HasPrefix(authHeader, bearerPrefix) {
		return strings.TrimSpace(authHeader[len(bearerPrefix):])
	}
	return ""
}

// extractClientIP extracts client IP handling various headers
func extractClientIP(rawIP string) string {
	// Handle common proxy headers format
	// X-Forwarded-For: client, proxy1, proxy2
	// X-Real-IP: client

	// Split by comma and take first IP (original client)
	if strings.Contains(rawIP, ",") {
		ips := strings.Split(rawIP, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Remove port if present
	if host, _, err := net.SplitHostPort(rawIP); err == nil {
		return host
	}

	return rawIP
}

// ValidateAuthContext performs basic validation on AuthContext
func ValidateAuthContext(ctx AuthContext) error {
	// Basic validation - can be extended
	if ctx.ClientIP == "" {
		return ErrMissingClientIP
	}
	return nil
}

// SanitizeAuthContext sanitizes sensitive data in AuthContext for logging
func SanitizeAuthContext(ctx AuthContext) AuthContext {
	sanitized := ctx

	// Mask API key for logging (show only first 8 characters)
	if len(sanitized.APIKey) > 8 {
		sanitized.APIKey = sanitized.APIKey[:8] + "..."
	}

	return sanitized
}

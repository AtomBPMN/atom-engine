/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package auth

import (
	"net"
	"strings"

	"atom-engine/src/core/logger"
)

// ipValidator implements IPValidator interface
type ipValidator struct {
	globalAllowedHosts []string
}

// NewIPValidator creates a new IP validator
func NewIPValidator(globalAllowedHosts []string) IPValidator {
	return &ipValidator{
		globalAllowedHosts: globalAllowedHosts,
	}
}

// ValidateIP checks if IP is allowed based on configuration
func (v *ipValidator) ValidateIP(ip string, allowedHosts []string) bool {
	// Always allow localhost
	if IsLocalhost(ip) {
		return true
	}

	// Check specific allowed hosts first (per-key restrictions)
	if len(allowedHosts) > 0 {
		return v.checkIPInList(ip, allowedHosts)
	}

	// Fall back to global allowed hosts
	return v.IsAllowedGlobally(ip)
}

// IsAllowedGlobally checks if IP is in global whitelist
func (v *ipValidator) IsAllowedGlobally(ip string) bool {
	// Always allow localhost
	if IsLocalhost(ip) {
		return true
	}

	return v.checkIPInList(ip, v.globalAllowedHosts)
}

// checkIPInList checks if IP matches any entry in the allowed list
func (v *ipValidator) checkIPInList(ip string, allowedList []string) bool {
	clientIP := net.ParseIP(ip)
	if clientIP == nil {
		logger.Warn("Invalid IP address format", logger.String("ip", ip))
		return false
	}

	for _, allowed := range allowedList {
		// Handle CIDR notation
		if strings.Contains(allowed, "/") {
			if v.checkIPInCIDR(clientIP, allowed) {
				return true
			}
		} else {
			// Handle specific IP address
			if v.checkSpecificIP(clientIP, allowed) {
				return true
			}
		}
	}

	return false
}

// checkIPInCIDR checks if IP is within CIDR range
func (v *ipValidator) checkIPInCIDR(clientIP net.IP, cidr string) bool {
	_, network, err := net.ParseCIDR(cidr)
	if err != nil {
		logger.Warn("Invalid CIDR format",
			logger.String("cidr", cidr),
			logger.String("error", err.Error()))
		return false
	}

	return network.Contains(clientIP)
}

// checkSpecificIP checks if IP matches specific allowed IP
func (v *ipValidator) checkSpecificIP(clientIP net.IP, allowedIP string) bool {
	allowed := net.ParseIP(allowedIP)
	if allowed == nil {
		logger.Warn("Invalid allowed IP format", logger.String("ip", allowedIP))
		return false
	}

	return clientIP.Equal(allowed)
}

// GetAllowedHosts returns the global allowed hosts list
func (v *ipValidator) GetAllowedHosts() []string {
	return v.globalAllowedHosts
}

// UpdateAllowedHosts updates the global allowed hosts list
func (v *ipValidator) UpdateAllowedHosts(hosts []string) {
	v.globalAllowedHosts = hosts
}

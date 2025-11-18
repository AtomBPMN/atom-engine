/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// GenerateSecureRequestID generates a cryptographically secure request ID using UUID v4 standard
func GenerateSecureRequestID(prefix string) string {
	uuid := generateUUIDv4()
	if uuid == "" {
		// Fallback to timestamp-based ID if UUID generation fails
		timestamp := time.Now().UnixNano()
		return fmt.Sprintf("%s_%d", prefix, timestamp%1000000)
	}

	return fmt.Sprintf("%s_%s", prefix, uuid)
}

// generateUUIDv4 generates a UUID v4 (random) compliant with RFC 4122
func generateUUIDv4() string {
	// UUID v4 requires 16 random bytes
	uuid := make([]byte, 16)
	if _, err := rand.Read(uuid); err != nil {
		return "" // Indicate failure
	}

	// Set version (4) in the most significant 4 bits of the 7th byte
	uuid[6] = (uuid[6] & 0x0f) | 0x40 // Version 4

	// Set variant (2 bits) in the most significant 2 bits of the 9th byte
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // Variant bits 10

	// Format as UUID string: xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		uuid[0:4],   // time_low
		uuid[4:6],   // time_mid
		uuid[6:8],   // time_hi_and_version
		uuid[8:10],  // clock_seq_hi_and_reserved + clock_seq_low
		uuid[10:16]) // node
}

// GenerateSecureRandomString generates a cryptographically secure random string
func GenerateSecureRandomString(length int) string {
	if length <= 0 {
		return ""
	}

	// Generate enough random bytes
	numBytes := (length + 1) / 2 // Each byte gives 2 hex chars
	randomBytes := make([]byte, numBytes)
	if _, err := rand.Read(randomBytes); err != nil {
		// Fallback to timestamp-based string if crypto/rand fails
		timestamp := time.Now().UnixNano()
		fallback := fmt.Sprintf("%d", timestamp)
		if len(fallback) > length {
			return fallback[:length]
		}
		return strings.Repeat(fallback, (length/len(fallback))+1)[:length]
	}

	// Convert to hex and truncate to desired length
	hexStr := hex.EncodeToString(randomBytes)
	if len(hexStr) > length {
		return hexStr[:length]
	}
	return hexStr
}

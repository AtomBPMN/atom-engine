/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package models

import (
	"crypto/rand"
	"math/big"
	"strings"
	"sync"
	"time"
)

// Global instance name for ID generation
// Глобальное имя инстанса для генерации ID
var (
	instanceName string
	instanceMu   sync.RWMutex
)

// SetInstanceName sets instance name for ID generation
// Устанавливает имя инстанса для генерации ID
func SetInstanceName(name string) {
	instanceMu.Lock()
	defer instanceMu.Unlock()
	instanceName = name
}

// GetInstanceName returns current instance name
// Возвращает текущее имя инстанса
func GetInstanceName() string {
	instanceMu.RLock()
	defer instanceMu.RUnlock()
	return instanceName
}

// GenerateID generates unique ID with node prefix + NanoID
// Генерирует уникальный ID с префиксом узла + NanoID
func GenerateID() string {
	nodePrefix := getNodePrefix()
	nanoID := generateNanoID(18)
	return nodePrefix + "-" + nanoID
}

// getNodePrefix gets 4-character node prefix from instance name
// Получает 4-символьный префикс узла из имени инстанса
func getNodePrefix() string {
	instanceMu.RLock()
	name := instanceName
	instanceMu.RUnlock()

	if name == "" {
		name = "unkn" // fallback if not set
	}

	// Clean instance name and take first 4 chars
	// Очищаем имя инстанса и берем первые 4 символа
	cleaned := strings.ToLower(strings.ReplaceAll(name, ".", ""))
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	cleaned = strings.ReplaceAll(cleaned, "_", "")

	if len(cleaned) >= 4 {
		return cleaned[:4]
	}

	// Pad with zeros if name is too short
	// Дополняем нулями если имя слишком короткое
	for len(cleaned) < 4 {
		cleaned += "0"
	}

	return cleaned
}

// generateNanoID generates NanoID-like string with custom alphabet
// First and last characters are always alphanumeric (no special chars)
// Генерирует NanoID-подобную строку с кастомным алфавитом
// Первый и последний символы всегда буквенно-цифровые (без спецсимволов)
func generateNanoID(length int) string {
	// URL-safe alphabet for middle characters
	// URL-safe алфавит для средних символов
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_-"
	// Alphanumeric only for first and last characters
	// Только буквенно-цифровые для первого и последнего символов
	const alphanumeric = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	result := make([]byte, length)
	for i := range result {
		var alphabet string
		// Use alphanumeric for first and last positions
		// Используем буквенно-цифровые для первой и последней позиций
		if i == 0 || i == length-1 {
			alphabet = alphanumeric
		} else {
			alphabet = charset
		}

		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			// Fallback to time-based if crypto/rand fails
			// Фоллбэк на time-based если crypto/rand не работает
			result[i] = alphabet[time.Now().UnixNano()%int64(len(alphabet))]
		} else {
			result[i] = alphabet[num.Int64()]
		}
	}

	return string(result)
}


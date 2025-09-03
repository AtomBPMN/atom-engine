/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package timewheel

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ISO8601DurationParser parses ISO8601 duration strings
// Парсер ISO8601 строк длительности
type ISO8601DurationParser struct{}

// NewISO8601DurationParser creates new parser
// Создает новый парсер
func NewISO8601DurationParser() *ISO8601DurationParser {
	return &ISO8601DurationParser{}
}

// ParseDuration parses ISO8601 duration string like "PT30S", "P1DT2H"
// Парсит ISO8601 строку длительности типа "PT30S", "P1DT2H"
func (p *ISO8601DurationParser) ParseDuration(durationStr string) (time.Duration, error) {
	if durationStr == "" {
		return 0, fmt.Errorf("empty duration string")
	}

	// Convert to uppercase for case-insensitive parsing
	// Преобразуем в верхний регистр для регистронезависимого парсинга
	durationStr = strings.ToUpper(durationStr)

	// ISO8601 duration regex: P[nY][nM][nD][T[nH][nM][nS]]
	regex := regexp.MustCompile(`^P(?:(\d+)Y)?(?:(\d+)M)?(?:(\d+)D)?(?:T(?:(\d+)H)?(?:(\d+)M)?(?:(\d+(?:\.\d+)?)S)?)?$`)
	matches := regex.FindStringSubmatch(durationStr)

	if matches == nil {
		return 0, fmt.Errorf("invalid ISO8601 duration format: %s", durationStr)
	}

	var totalDuration time.Duration

	// Years (approximate as 365 days)
	// Годы (приблизительно 365 дней)
	if matches[1] != "" {
		if years, err := strconv.Atoi(matches[1]); err == nil {
			totalDuration += time.Duration(years) * 365 * 24 * time.Hour
		}
	}

	// Months (approximate as 30 days)
	// Месяцы (приблизительно 30 дней)
	if matches[2] != "" {
		if months, err := strconv.Atoi(matches[2]); err == nil {
			totalDuration += time.Duration(months) * 30 * 24 * time.Hour
		}
	}

	// Days
	// Дни
	if matches[3] != "" {
		if days, err := strconv.Atoi(matches[3]); err == nil {
			totalDuration += time.Duration(days) * 24 * time.Hour
		}
	}

	// Hours
	// Часы
	if matches[4] != "" {
		if hours, err := strconv.Atoi(matches[4]); err == nil {
			totalDuration += time.Duration(hours) * time.Hour
		}
	}

	// Minutes
	// Минуты
	if matches[5] != "" {
		if minutes, err := strconv.Atoi(matches[5]); err == nil {
			totalDuration += time.Duration(minutes) * time.Minute
		}
	}

	// Seconds (can be decimal)
	// Секунды (могут быть десятичными)
	if matches[6] != "" {
		if seconds, err := strconv.ParseFloat(matches[6], 64); err == nil {
			totalDuration += time.Duration(seconds * float64(time.Second))
		}
	}

	return totalDuration, nil
}

// ParseRepeatingInterval parses repeating interval like "R5/PT30S"
// Парсит повторяющийся интервал типа "R5/PT30S"
func (p *ISO8601DurationParser) ParseRepeatingInterval(intervalStr string) (repeatCount int, interval time.Duration, err error) {
	if intervalStr == "" {
		return 0, 0, fmt.Errorf("empty interval string")
	}

	// Convert to uppercase for case-insensitive parsing
	// Преобразуем в верхний регистр для регистронезависимого парсинга
	intervalStr = strings.ToUpper(intervalStr)

	// Check if it starts with R
	// Проверяем начинается ли с R
	if !strings.HasPrefix(intervalStr, "R") {
		return 0, 0, fmt.Errorf("repeating interval must start with 'R': %s", intervalStr)
	}

	// Split by '/'
	// Разделяем по '/'
	parts := strings.Split(intervalStr, "/")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid repeating interval format: %s", intervalStr)
	}

	// Parse repeat count
	// Парсим количество повторений
	repeatStr := strings.TrimPrefix(parts[0], "R")
	if repeatStr == "" {
		// Infinite repetition
		// Бесконечное повторение
		repeatCount = -1
	} else {
		var parseErr error
		repeatCount, parseErr = strconv.Atoi(repeatStr)
		if parseErr != nil {
			return 0, 0, fmt.Errorf("invalid repeat count: %s", repeatStr)
		}
	}

	// Parse duration
	// Парсим длительность
	interval, err = p.ParseDuration(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid duration in repeating interval: %w", err)
	}

	return repeatCount, interval, nil
}

// ParseDate parses ISO8601 date string like "2025-12-31T23:59:59Z"
// Парсит ISO8601 строку даты типа "2025-12-31T23:59:59Z"
func (p *ISO8601DurationParser) ParseDate(dateStr string) (time.Time, error) {
	if dateStr == "" {
		return time.Time{}, fmt.Errorf("empty date string")
	}

	// Try different ISO8601 formats
	// Пробуем разные форматы ISO8601
	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05",
		"2006-01-02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid date format: %s", dateStr)
}

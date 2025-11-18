/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package utils

import (
	"math"
	"strconv"

	"atom-engine/src/core/restapi/models"
)

// ParsePaginationParams parses pagination parameters from query string
func ParsePaginationParams(pageStr, limitStr string) models.PaginationParams {
	params := models.GetDefaultPagination()

	if pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			params.Page = page
		}
	}

	if limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 1000 {
			params.Limit = limit
		}
	}

	return params
}

// CalculatePaginationInfo calculates pagination metadata
func CalculatePaginationInfo(page, limit, totalCount int) *models.PaginationInfo {
	if totalCount == 0 {
		return &models.PaginationInfo{
			Page:    page,
			Limit:   limit,
			Total:   0,
			Pages:   0,
			HasNext: false,
			HasPrev: false,
		}
	}

	totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))

	return &models.PaginationInfo{
		Page:    page,
		Limit:   limit,
		Total:   totalCount,
		Pages:   totalPages,
		HasNext: page < totalPages,
		HasPrev: page > 1,
	}
}

// GetOffset calculates offset for database queries
func GetOffset(page, limit int) int {
	if page <= 0 {
		page = 1
	}
	return (page - 1) * limit
}

// ValidatePaginationParams validates pagination parameters
func ValidatePaginationParams(params models.PaginationParams) *models.APIError {
	if params.Page < 1 {
		return models.BadRequestError("page must be greater than 0")
	}

	if params.Limit < 1 {
		return models.BadRequestError("limit must be greater than 0")
	}

	if params.Limit > 1000 {
		return models.BadRequestError("limit cannot exceed 1000")
	}

	return nil
}

// ApplyPagination applies pagination to a slice
func ApplyPagination(items interface{}, page, limit int) (interface{}, *models.PaginationInfo) {
	switch v := items.(type) {
	case []interface{}:
		return applySlicePagination(v, page, limit)
	case []string:
		return applyStringSlicePagination(v, page, limit)
	case []map[string]interface{}:
		return applyMapSlicePagination(v, page, limit)
	default:
		// Return original if type not supported
		return items, CalculatePaginationInfo(page, limit, 0)
	}
}

// applySlicePagination applies pagination to []interface{}
func applySlicePagination(items []interface{}, page, limit int) ([]interface{}, *models.PaginationInfo) {
	totalCount := len(items)
	offset := GetOffset(page, limit)

	// Handle empty slice or offset beyond bounds
	if totalCount == 0 || offset >= totalCount {
		return []interface{}{}, CalculatePaginationInfo(page, limit, totalCount)
	}

	end := offset + limit
	if end > totalCount {
		end = totalCount
	}

	paginatedItems := items[offset:end]
	paginationInfo := CalculatePaginationInfo(page, limit, totalCount)

	return paginatedItems, paginationInfo
}

// applyStringSlicePagination applies pagination to []string
func applyStringSlicePagination(items []string, page, limit int) ([]string, *models.PaginationInfo) {
	totalCount := len(items)
	offset := GetOffset(page, limit)

	if totalCount == 0 || offset >= totalCount {
		return []string{}, CalculatePaginationInfo(page, limit, totalCount)
	}

	end := offset + limit
	if end > totalCount {
		end = totalCount
	}

	paginatedItems := items[offset:end]
	paginationInfo := CalculatePaginationInfo(page, limit, totalCount)

	return paginatedItems, paginationInfo
}

// applyMapSlicePagination applies pagination to []map[string]interface{}
func applyMapSlicePagination(
	items []map[string]interface{},
	page, limit int,
) ([]map[string]interface{}, *models.PaginationInfo) {
	totalCount := len(items)
	offset := GetOffset(page, limit)

	if totalCount == 0 || offset >= totalCount {
		return []map[string]interface{}{}, CalculatePaginationInfo(page, limit, totalCount)
	}

	end := offset + limit
	if end > totalCount {
		end = totalCount
	}

	paginatedItems := items[offset:end]
	paginationInfo := CalculatePaginationInfo(page, limit, totalCount)

	return paginatedItems, paginationInfo
}

// PaginationHelper provides pagination utilities
type PaginationHelper struct{}

// NewPaginationHelper creates new pagination helper
func NewPaginationHelper() *PaginationHelper {
	return &PaginationHelper{}
}

// ParseAndValidate parses and validates pagination parameters
func (h *PaginationHelper) ParseAndValidate(pageStr, limitStr string) (models.PaginationParams, *models.APIError) {
	params := ParsePaginationParams(pageStr, limitStr)

	if err := ValidatePaginationParams(params); err != nil {
		return params, err
	}

	return params, nil
}

// CreateResponse creates paginated response
func (h *PaginationHelper) CreateResponse(
	items interface{},
	totalCount int,
	params models.PaginationParams,
	requestID string,
) *models.PaginatedResponse {
	paginationInfo := CalculatePaginationInfo(params.Page, params.Limit, totalCount)

	return models.PaginatedSuccessResponse(items, paginationInfo, requestID)
}

// FilterAndPaginate filters and paginates items based on criteria
func FilterAndPaginate(
	items []map[string]interface{},
	filters map[string]interface{},
	page, limit int,
) ([]map[string]interface{}, *models.PaginationInfo) {
	// Apply filters
	filteredItems := make([]map[string]interface{}, 0)

	for _, item := range items {
		match := true
		for key, value := range filters {
			if itemValue, exists := item[key]; !exists || itemValue != value {
				match = false
				break
			}
		}
		if match {
			filteredItems = append(filteredItems, item)
		}
	}

	// Apply pagination
	return applyMapSlicePagination(filteredItems, page, limit)
}

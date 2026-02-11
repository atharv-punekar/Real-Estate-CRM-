package utils

import (
	"fmt"
	"strings"
)

// PaginationMeta represents standardized pagination metadata
type PaginationMeta struct {
	Page        int   `json:"page"`
	Limit       int   `json:"limit"`
	Total       int64 `json:"total_count"`
	TotalPages  int   `json:"total_pages"`
	OffsetStart int   `json:"offset_start"`
	OffsetEnd   int   `json:"offset_end"`
}

// CalculatePagination generates consistent pagination metadata
func CalculatePagination(page, limit int, total int64) PaginationMeta {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}

	offsetStart := (page-1)*limit + 1
	offsetEnd := page * limit
	if offsetEnd > int(total) {
		offsetEnd = int(total)
	}

	// Handle empty results
	if total == 0 {
		offsetStart = 0
		offsetEnd = 0
	}

	return PaginationMeta{
		Page:        page,
		Limit:       limit,
		Total:       total,
		TotalPages:  totalPages,
		OffsetStart: offsetStart,
		OffsetEnd:   offsetEnd,
	}
}

// ValidatePaginationParams validates and returns safe pagination parameters
func ValidatePaginationParams(page, limit int) (int, int) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return page, limit
}

// BuildSortQuery builds a safe SQL ORDER BY clause
func BuildSortQuery(sortBy, sortOrder, defaultSort string, allowedFields []string) string {
	// Validate sortOrder
	sortOrder = strings.ToUpper(sortOrder)
	if sortOrder != "ASC" && sortOrder != "DESC" {
		sortOrder = "DESC"
	}

	// Validate sortBy is in allowed fields
	if sortBy != "" {
		for _, field := range allowedFields {
			if sortBy == field {
				return fmt.Sprintf("%s %s", sortBy, sortOrder)
			}
		}
	}

	// Return default sort
	return defaultSort
}

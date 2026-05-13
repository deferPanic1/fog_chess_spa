package httphandler

import (
	"net/http"
	"strconv"
)

const (
	defaultPage  = 1
	defaultLimit = 20
	maxLimit     = 100
)

type paginationParams struct {
	Page   int `json:"page"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type paginationMeta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"totalPages"`
	HasPrev    bool  `json:"hasPrev"`
	HasNext    bool  `json:"hasNext"`
}

func readPagination(r *http.Request) paginationParams {
	page := defaultPage
	if rawPage := r.URL.Query().Get("page"); rawPage != "" {
		if parsed, err := strconv.Atoi(rawPage); err == nil && parsed > 0 {
			page = parsed
		}
	}

	limit := defaultLimit
	if rawLimit := r.URL.Query().Get("limit"); rawLimit != "" {
		if parsed, err := strconv.Atoi(rawLimit); err == nil && parsed > 0 && parsed <= maxLimit {
			limit = parsed
		}
	}

	return paginationParams{
		Page:   page,
		Limit:  limit,
		Offset: (page - 1) * limit,
	}
}

func newPaginationMeta(page, limit int, total int64) paginationMeta {
	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(limit) - 1) / int64(limit))
	}

	return paginationMeta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
		HasPrev:    page > 1 && totalPages > 0,
		HasNext:    totalPages > 0 && page < totalPages,
	}
}

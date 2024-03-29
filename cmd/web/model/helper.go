package model

// PageResponse paginated response
type PageResponse[T any] struct {
	Items      []T   `json:"items"`
	TotalCount int64 `json:"total_count"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
}

// NewPageResponse new page response
func NewPageResponse[T any](items []T, totalCount int64, page, pageSize int) *PageResponse[T] {
	return &PageResponse[T]{
		Items:      items,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
	}
}

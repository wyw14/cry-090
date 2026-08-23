package common

import "strings"

type PageRequest struct {
	Cursor  string
	Page    int
	Limit   int
	Sort    string
	Filters map[string]string
}

func (p PageRequest) Normalize(allowedSorts, allowedFilters map[string]struct{}) (PageRequest, error) {
	if p.Page < 0 {
		return PageRequest{}, FieldError("page", "must not be negative")
	}
	if p.Limit == 0 {
		p.Limit = 20
	}
	if p.Limit < 1 || p.Limit > 100 {
		return PageRequest{}, FieldError("limit", "must be between 1 and 100")
	}
	if p.Cursor != "" && p.Page != 0 {
		return PageRequest{}, FieldError("cursor", "cannot be combined with page")
	}
	p.Sort = strings.TrimSpace(p.Sort)
	if p.Sort != "" {
		key := strings.TrimPrefix(p.Sort, "-")
		if _, ok := allowedSorts[key]; !ok {
			return PageRequest{}, FieldError("sort", "unsupported sort field")
		}
	}
	clean := make(map[string]string, len(p.Filters))
	for key, value := range p.Filters {
		if _, ok := allowedFilters[key]; !ok {
			return PageRequest{}, FieldError("filter."+key, "unsupported filter")
		}
		clean[key] = strings.TrimSpace(value)
	}
	p.Filters = clean
	return p, nil
}

type Page[T any] struct {
	Items      []T    `json:"items"`
	NextCursor string `json:"next_cursor,omitempty"`
	Page       int    `json:"page,omitempty"`
	Total      int    `json:"total"`
}

package httpx

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// ListPage membungkus item daftar dengan jumlah dan parameter halamannya.
type ListPage[T any] struct {
	Items []T `json:"items"`
	Total int `json:"total"`
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

// PaginateList menerapkan parameter page/limit pada daftar yang sudah diambil.
func PaginateList[T any](c *gin.Context, all []T) ListPage[T] {
	total := len(all)
	if all == nil {
		all = []T{}
	}
	pageQuery, hasPage := c.GetQuery("page")
	limitQuery, hasLimit := c.GetQuery("limit")
	if !hasPage && !hasLimit {
		return ListPage[T]{Items: all, Total: total, Page: 1, Limit: total}
	}

	page := 1
	if parsed, err := strconv.Atoi(pageQuery); hasPage && err == nil && parsed > 0 {
		page = parsed
	}
	limit := 20
	if parsed, err := strconv.Atoi(limitQuery); hasLimit && err == nil && parsed > 0 {
		limit = parsed
	}
	if limit > 200 {
		limit = 200
	}

	if page-1 > len(all)/limit {
		return ListPage[T]{Items: []T{}, Total: total, Page: page, Limit: limit}
	}
	start := (page - 1) * limit
	if start >= len(all) {
		return ListPage[T]{Items: []T{}, Total: total, Page: page, Limit: limit}
	}
	end := len(all)
	if limit <= len(all)-start {
		end = start + limit
	}
	return ListPage[T]{Items: all[start:end], Total: total, Page: page, Limit: limit}
}

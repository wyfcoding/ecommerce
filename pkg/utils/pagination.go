// Package utils 提供了通用的工具函数集合。
// 此文件定义了用于处理API响应中分页逻辑的结构体和辅助函数。
package utils

import "math"

// Pagination 结构体封装了所有与分页相关的信息。
// 它通常用于API响应中，告知客户端当前数据集合的分页状态。
type Pagination struct {
	// Page 是当前页码，从1开始。
	Page int `json:"page"`
	// PageSize 是每页包含的条目数。
	PageSize int `json:"page_size"`
	// TotalItems 是符合查询条件的总条目数。
	TotalItems int64 `json:"total_items"`
	// TotalPages 是根据 TotalItems 和 PageSize 计算出的总页数。
	TotalPages int `json:"total_pages"`
}

// NewPagination 是一个构造函数，用于创建一个新的 Pagination 实例。
// 它接收当前页码、每页大小和总条目数作为输入，并自动计算总页数。
//
// 参数:
//
//	page - 客户端请求的页码。如果小于1，将被强制设置为1。
//	pageSize - 客户端请求的每页大小。如果小于1，将被强制设置为默认值10。
//	totalItems - 数据库中查询到的总记录数。
//
// 返回:
//
//	一个初始化并计算好总页数的 *Pagination 指针。
func NewPagination(page, pageSize int, totalItems int64) *Pagination {
	// 对非法的页码和页面大小进行修正
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		// 设置一个合理的默认页面大小
		pageSize = 10
	}

	// 计算总页数，使用 math.Ceil 向上取整，确保最后不足一页的数据也能被正确计入
	totalPages := int(math.Ceil(float64(totalItems) / float64(pageSize)))

	return &Pagination{
		Page:       page,
		PageSize:   pageSize,
		TotalItems: totalItems,
		TotalPages: totalPages,
	}
}

// Offset 计算并返回用于数据库查询的偏移量（offset）。
// 偏移量是根据当前页码和每页大小计算得出的，用于 LIMIT...OFFSET... 类型的SQL查询。
// 例如，第1页的偏移量是0，第2页的偏移量是 PageSize，以此类推。
func (p *Pagination) Offset() int {
	return (p.Page - 1) * p.PageSize
}

// HasNext 检查是否存在下一页。
// 如果当前页码小于总页数，则返回 true。
func (p *Pagination) HasNext() bool {
	return p.Page < p.TotalPages
}

// HasPrev 检查是否存在上一页。
// 如果当前页码大于1, 则返回 true。
func (p *Pagination) HasPrev() bool {
	return p.Page > 1
}

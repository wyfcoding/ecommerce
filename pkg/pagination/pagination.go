package pagination

// Page 分页参数
type Page struct {
	PageNum  int `json:"page_num" form:"page_num"`   // 页码，从1开始
	PageSize int `json:"page_size" form:"page_size"` // 每页数量
}

// Validate 验证并设置默认值
func (p *Page) Validate() {
	if p.PageNum <= 0 {
		p.PageNum = 1
	}
	if p.PageSize <= 0 {
		p.PageSize = 10
	}
	if p.PageSize > 100 {
		p.PageSize = 100
	}
}

// Offset 计算偏移量
func (p *Page) Offset() int {
	return (p.PageNum - 1) * p.PageSize
}

// Limit 返回限制数量
func (p *Page) Limit() int {
	return p.PageSize
}

// PageResult 分页结果
type PageResult struct {
	Total    int64       `json:"total"`     // 总记录数
	PageNum  int         `json:"page_num"`  // 当前页码
	PageSize int         `json:"page_size"` // 每页数量
	Data     interface{} `json:"data"`      // 数据列表
}

// NewPageResult 创建分页结果
func NewPageResult(total int64, page *Page, data interface{}) *PageResult {
	return &PageResult{
		Total:    total,
		PageNum:  page.PageNum,
		PageSize: page.PageSize,
		Data:     data,
	}
}

// TotalPages 计算总页数
func (r *PageResult) TotalPages() int {
	if r.PageSize == 0 {
		return 0
	}
	pages := int(r.Total) / r.PageSize
	if int(r.Total)%r.PageSize > 0 {
		pages++
	}
	return pages
}

// HasNext 是否有下一页
func (r *PageResult) HasNext() bool {
	return r.PageNum < r.TotalPages()
}

// HasPrev 是否有上一页
func (r *PageResult) HasPrev() bool {
	return r.PageNum > 1
}

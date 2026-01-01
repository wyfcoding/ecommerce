package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/dataprocessing/domain"
)

// DataProcessingQuery 处理所有数据处理相关的查询操作（Queries）。
type DataProcessingQuery struct {
	repo domain.DataProcessingRepository
}

// NewDataProcessingQuery 构造函数。
func NewDataProcessingQuery(repo domain.DataProcessingRepository) *DataProcessingQuery {
	return &DataProcessingQuery{repo: repo}
}

// ListTasks 获取数据处理任务列表。
func (q *DataProcessingQuery) ListTasks(ctx context.Context, workflowID uint64, status domain.TaskStatus, page, pageSize int) ([]*domain.ProcessingTask, int64, error) {
	offset := (page - 1) * pageSize
	return q.repo.ListTasks(ctx, workflowID, status, offset, pageSize)
}

// ListWorkflows 获取数据处理工作流列表。
func (q *DataProcessingQuery) ListWorkflows(ctx context.Context, page, pageSize int) ([]*domain.ProcessingWorkflow, int64, error) {
	offset := (page - 1) * pageSize
	return q.repo.ListWorkflows(ctx, offset, pageSize)
}

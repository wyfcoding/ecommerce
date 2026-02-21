package domain

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	pb "github.com/wyfcoding/ecommerce/go-api/order/v1"
	"github.com/wyfcoding/pkg/worker"
)

type ExportStatus string

const (
	ExportStatusPending    ExportStatus = "PENDING"
	ExportStatusProcessing ExportStatus = "PROCESSING"
	ExportStatusCompleted  ExportStatus = "COMPLETED"
	ExportStatusFailed     ExportStatus = "FAILED"
	ExportStatusCancelled  ExportStatus = "CANCELLED"
)

// ExportFormat 导出格式
type ExportFormat string

const (
	ExportFormatCSV   ExportFormat = "csv"
	ExportFormatExcel ExportFormat = "excel"
	ExportFormatJSON  ExportFormat = "json"
)

// OrderExportTask 异步导出任务实体
type OrderExportTask struct {
	ID           uint64       `json:"id"`
	TaskNo       string       `json:"task_no"`
	UserID       uint64       `json:"user_id"`
	Filter       string       `json:"filter"` // JSON 存储的搜索过滤条件
	Format       ExportFormat `json:"format"`
	Status       ExportStatus `json:"status"`
	FileURL      string       `json:"file_url"`
	FileSize     int64        `json:"file_size"`
	TotalRecords int64        `json:"total_records"`
	Processed    int64        `json:"processed"`
	ErrorMsg     string       `json:"error_msg"`
	CreatedAt    time.Time    `json:"created_at"`
	StartedAt    *time.Time   `json:"started_at"`
	CompletedAt  *time.Time   `json:"completed_at"`
	ExpiresAt    time.Time    `json:"expires_at"`
	Progress     int          `json:"progress"` // 0-100
}

// ExportQueryParams 导出查询参数
type ExportQueryParams struct {
	StartTime      *time.Time          `json:"start_time"`
	EndTime        *time.Time          `json:"end_time"`
	Status         []pb.OrderStatus    `json:"status"`
	PaymentStatus  []pb.PaymentStatus  `json:"payment_status"`
	ShippingStatus []pb.ShippingStatus `json:"shipping_status"`
	MerchantIDs    []uint64            `json:"merchant_ids"`
	ProductIDs     []uint64            `json:"product_ids"`
	UserIDs        []uint64            `json:"user_ids"`
	OrderTypes     []pb.OrderType      `json:"order_types"`
	MinAmount      int64               `json:"min_amount"`
	MaxAmount      int64               `json:"max_amount"`
	Keywords       string              `json:"keywords"`
	SortBy         string              `json:"sort_by"`
	SortOrder      string              `json:"sort_order"`
	Limit          int                 `json:"limit"`
	Offset         int                 `json:"offset"`
}

// ExportService 导出服务接口
type ExportService interface {
	// CreateExportTask 创建导出任务
	CreateExportTask(ctx context.Context, userID uint64, format ExportFormat, params *ExportQueryParams) (*OrderExportTask, error)

	// GetExportTask 获取导出任务状态
	GetExportTask(ctx context.Context, taskID string) (*OrderExportTask, error)

	// ListExportTasks 列出用户的导出任务
	ListExportTasks(ctx context.Context, userID uint64, status []ExportStatus, limit, offset int) ([]*OrderExportTask, int64, error)

	// CancelExportTask 取消导出任务
	CancelExportTask(ctx context.Context, taskID string) error

	// DownloadExportFile 下载导出文件
	DownloadExportFile(ctx context.Context, taskID string) ([]byte, error)

	// CleanupExpiredExports 清理过期的导出文件
	CleanupExpiredExports(ctx context.Context) error
}

// ExportManager 导出管理器
type ExportManager struct {
	orderRepo     OrderRepository
	exportRepo    OrderExportRepository
	workerPool    *worker.Pool
	storagePath   string
	maxFileSize   int64
	retentionDays int
	maxConcurrent int
	taskQueue     chan *OrderExportTask
	mu            sync.RWMutex
	activeTasks   map[string]*OrderExportTask
}

// NewExportManager 创建导出管理器
func NewExportManager(orderRepo OrderRepository, exportRepo OrderExportRepository, storagePath string) *ExportManager {
	return &ExportManager{
		orderRepo:     orderRepo,
		exportRepo:    exportRepo,
		storagePath:   storagePath,
		maxFileSize:   100 * 1024 * 1024, // 100MB
		retentionDays: 7,
		maxConcurrent: 5,
		taskQueue:     make(chan *OrderExportTask, 100),
		activeTasks:   make(map[string]*OrderExportTask),
	}
}

// CreateExportTask 创建导出任务
func (m *ExportManager) CreateExportTask(ctx context.Context, userID uint64, format ExportFormat, params *ExportQueryParams) (*OrderExportTask, error) {
	// 验证导出格式
	if !isValidExportFormat(format) {
		return nil, fmt.Errorf("unsupported export format: %s", format)
	}

	// 序列化查询参数
	var filterJSON string
	if params != nil {
		paramsBytes, err := json.Marshal(params)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal query params: %w", err)
		}
		filterJSON = string(paramsBytes)
	}

	// 创建任务
	task := &OrderExportTask{
		TaskNo:    generateExportTaskNo(),
		UserID:    userID,
		Filter:    filterJSON,
		Format:    format,
		Status:    ExportStatusPending,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(time.Duration(m.retentionDays) * 24 * time.Hour),
		Progress:  0,
	}

	// 保存任务
	err := m.exportRepo.Save(task)
	if err != nil {
		return nil, fmt.Errorf("failed to save export task: %w", err)
	}

	// 异步处理任务
	go m.processExportTask(task)

	return task, nil
}

// GetExportTask 获取导出任务状态
func (m *ExportManager) GetExportTask(ctx context.Context, taskID string) (*OrderExportTask, error) {
	task, err := m.exportRepo.GetByTaskNo(taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get export task: %w", err)
	}

	// TODO: 检查权限
	// if task.UserID != getUserIDFromContext(ctx) {
	//     return nil, fmt.Errorf("unauthorized access to export task")
	// }

	return task, nil
}

// CancelExportTask 取消导出任务
func (m *ExportManager) CancelExportTask(ctx context.Context, taskID string) error {
	task, err := m.exportRepo.GetByTaskNo(taskID)
	if err != nil {
		return fmt.Errorf("failed to get export task: %w", err)
	}

	// 检查权限
	// if task.UserID != getUserIDFromContext(ctx) {
	//     return fmt.Errorf("unauthorized access to export task")
	// }

	// 只能取消待处理或处理中的任务
	if task.Status != ExportStatusPending && task.Status != ExportStatusProcessing {
		return fmt.Errorf("cannot cancel task with status: %s", task.Status)
	}

	// 更新任务状态
	task.Status = ExportStatusCancelled
	task.ErrorMsg = "Cancelled by user"

	err = m.exportRepo.UpdateStatus(task.ID, task.Status, task.FileURL, task.ErrorMsg)
	if err != nil {
		return fmt.Errorf("failed to update export task: %w", err)
	}

	return nil
}

// processExportTask 处理导出任务
func (m *ExportManager) processExportTask(task *OrderExportTask) {
	m.mu.Lock()
	m.activeTasks[task.TaskNo] = task
	m.mu.Unlock()

	defer func() {
		m.mu.Lock()
		delete(m.activeTasks, task.TaskNo)
		m.mu.Unlock()
	}()

	// 更新任务状态为处理中
	task.Status = ExportStatusProcessing
	startedAt := time.Now()
	task.StartedAt = &startedAt

	// 保存状态更新
	m.exportRepo.UpdateStatus(task.ID, task.Status, task.FileURL, task.ErrorMsg)

	// 解析查询参数
	var params ExportQueryParams
	if task.Filter != "" {
		err := json.Unmarshal([]byte(task.Filter), &params)
		if err != nil {
			task.ErrorMsg = fmt.Sprintf("failed to parse filter: %v", err)
			task.Status = ExportStatusFailed
			m.exportRepo.UpdateStatus(task.ID, task.Status, task.FileURL, task.ErrorMsg)
			return
		}
	}

	// 执行导出
	err := m.executeExport(task, &params)
	if err != nil {
		task.ErrorMsg = fmt.Sprintf("export failed: %v", err)
		task.Status = ExportStatusFailed
	} else {
		task.Status = ExportStatusCompleted
		completedAt := time.Now()
		task.CompletedAt = &completedAt
		task.Progress = 100
	}

	m.exportRepo.UpdateStatus(task.ID, task.Status, task.FileURL, task.ErrorMsg)
}

// executeExport 执行导出
func (m *ExportManager) executeExport(task *OrderExportTask, params *ExportQueryParams) error {
	// 根据格式选择导出方法
	switch task.Format {
	case ExportFormatCSV:
		return m.exportToCSV(task, params)
	case ExportFormatJSON:
		return m.exportToJSON(task, params)
	case ExportFormatExcel:
		return m.exportToExcel(task, params)
	default:
		return fmt.Errorf("unsupported export format: %s", task.Format)
	}
}

// exportToCSV 导出为CSV
func (m *ExportManager) exportToCSV(task *OrderExportTask, params *ExportQueryParams) error {
	// 获取订单数据
	orders, err := m.orderRepo.Search(context.Background(), params)
	if err != nil {
		return fmt.Errorf("failed to search orders: %w", err)
	}

	task.TotalRecords = int64(len(orders))
	// 创建CSV文件
	filePath := filepath.Join(m.storagePath, task.TaskNo+".csv")
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 写入CSV头部
	header := []string{
		"订单号", "用户ID", "状态", "支付状态", "物流状态",
		"总金额", "实付金额", "运费", "优惠金额",
		"支付方式", "收货人", "联系电话", "收货地址",
		"创建时间", "支付时间", "发货时间", "完成时间",
	}

	err = writer.Write(header)
	if err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// 写入实际订单数据
	for i, order := range orders {
		record := []string{
			order.OrderNo,
			fmt.Sprintf("%d", order.UserID),
			order.Status.String(),
			order.PaymentStatus.String(),
			order.ShippingStatus.String(),
			fmt.Sprintf("%.2f", float64(order.TotalAmount)/100),
			fmt.Sprintf("%.2f", float64(order.ActualAmount)/100),
			fmt.Sprintf("%.2f", float64(order.ShippingFee)/100),
			fmt.Sprintf("%.2f", float64(order.DiscountAmount)/100),
			order.PaymentMethod,
			order.ShippingAddress.RecipientName,
			order.ShippingAddress.PhoneNumber,
			fmt.Sprintf("%s%s%s%s",
				order.ShippingAddress.Province,
				order.ShippingAddress.City,
				order.ShippingAddress.District,
				order.ShippingAddress.DetailedAddress),
			order.CreatedAt.Format("2006-01-02 15:04:05"),
			formatTime(order.PaidAt),
			formatTime(order.ShippedAt),
			formatTime(order.CompletedAt),
		}
		err = writer.Write(record)
		if err != nil {
			return fmt.Errorf("failed to write CSV record: %w", err)
		}
		task.Processed = int64(i + 1)
		task.Progress = int(float64(i+1) / float64(len(orders)) * 100)
		// 定期更新任务状态
		if (i+1)%100 == 0 {
			m.exportRepo.UpdateStatus(task.ID, task.Status, task.FileURL, task.ErrorMsg)
		}
	}
	// 更新文件信息
	fileInfo, err := file.Stat()
	if err == nil {
		task.FileSize = fileInfo.Size()
		task.FileURL = fmt.Sprintf("/exports/%s", task.TaskNo)
	}

	return nil
}

// exportToJSON 导出为JSON
func (m *ExportManager) exportToJSON(task *OrderExportTask, params *ExportQueryParams) error {
	// 获取订单数据
	orders, err := m.orderRepo.Search(context.Background(), params)
	if err != nil {
		return fmt.Errorf("failed to search orders: %w", err)
	}

	task.TotalRecords = int64(len(orders))
	// 创建JSON文件
	filePath := filepath.Join(m.storagePath, task.TaskNo+".json")
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create JSON file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	// 写入JSON数据
	err = encoder.Encode(orders)
	if err != nil {
		return fmt.Errorf("failed to write JSON data: %w", err)
	}

	// 更新文件信息
	fileInfo, err := file.Stat()
	if err == nil {
		task.FileSize = fileInfo.Size()
		task.FileURL = fmt.Sprintf("/exports/%s", task.TaskNo)
	}

	return nil
}

// exportToExcel 导出为Excel
func (m *ExportManager) exportToExcel(task *OrderExportTask, params *ExportQueryParams) error {
	// TODO: 实现Excel导出
	// 可以使用第三方库如 tealeg/xlsx 或 unidoc/unioffice

	return fmt.Errorf("Excel export not implemented yet")
}

// Helper functions

func isValidExportFormat(format ExportFormat) bool {
	switch format {
	case ExportFormatCSV, ExportFormatExcel, ExportFormatJSON:
		return true
	default:
		return false
	}
}

func generateExportTaskNo() string {
	return fmt.Sprintf("EXPORT%d", time.Now().UnixNano())
}

func formatTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}

type OrderExportRepository interface {
	Save(task *OrderExportTask) error
	GetByID(id uint64) (*OrderExportTask, error)
	GetByTaskNo(no string) (*OrderExportTask, error)
	UpdateStatus(id uint64, status ExportStatus, fileURL, errorMsg string) error
}

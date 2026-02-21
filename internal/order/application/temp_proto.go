package application

// 临时protobuf消息定义，用于编译通过

type MergeOrdersRequest struct {
	OrderIds []uint64
	Operator string
}

type MergeOrdersResponse struct {
	BatchNo     string
	OrderCount  int32
	TotalAmount int64
	Status      string
}

type SplitMergeBatchRequest struct {
	BatchId  string
	Reason   string
	Operator string
}

type SplitMergeBatchResponse struct {
	Success bool
	Message string
}

type CreateExportTaskRequest struct {
	UserId uint64
	Format string
	Filter string
}

type CreateExportTaskResponse struct {
	TaskId    string
	Status    string
	CreatedAt string
	ExpiresAt string
}

type GetExportTaskRequest struct {
	TaskId string
}

type GetExportTaskResponse struct {
	TaskId       string
	Status       string
	Format       string
	FileUrl      string
	FileSize     int64
	TotalRecords int64
	Processed    int64
	Progress     int32
	ErrorMsg     string
	CreatedAt    string
	StartedAt    string
	CompletedAt  string
}

type AddOrderTagRequest struct {
	OrderId    uint64
	TagIds     []uint64
	OperatorId uint64
	Remark     string
}

type AddOrderTagResponse struct {
	Success bool
	Message string
}

type RemoveOrderTagRequest struct {
	OrderId uint64
	TagIds  []uint64
}

type RemoveOrderTagResponse struct {
	Success bool
	Message string
}

type GetOrderTagsRequest struct {
	OrderId uint64
}

type OrderTag struct {
	Id       uint64
	Name     string
	Color    string
	Type     string
	Priority int32
}

type GetOrderTagsResponse struct {
	Tags []*OrderTag
}

type AddOrderNoteRequest struct {
	OrderId   uint64
	UserId    uint64
	UserName  string
	UserType  string
	Content   string
	NoteType  string
	Priority  string
	IsPrivate bool
}

type AddOrderNoteResponse struct {
	NoteId  uint64
	Success bool
	Message string
}

type GetOrderNotesRequest struct {
	OrderId        uint64
	IncludePrivate bool
	NoteTypes      []string
	Page           int32
	PageSize       int32
}

type OrderNoteDetail struct {
	Id        uint64
	UserId    uint64
	UserName  string
	UserType  string
	Content   string
	NoteType  string
	Priority  string
	IsPrivate bool
	CreatedAt string
}

type GetOrderNotesResponse struct {
	Notes []*OrderNoteDetail
	Total int64
	Page  int32
}

type AssessOrderRiskRequest struct {
	OrderId   uint64
	UserId    uint64 // 添加缺失的 UserId 字段
	IpAddress string
	DeviceId  string
}

type AssessOrderRiskResponse struct {
	RiskScore      float32
	RiskLevel      string
	RiskAction     string
	ReviewRequired bool
	ReviewReason   string
	BlockReason    string
	AssessedAt     string
}

type CreateTimeoutTaskRequest struct {
	OrderId    uint64
	UserId     uint64 // 添加缺失的 UserId 字段
	PolicyType string
}

type CreateTimeoutTaskResponse struct {
	TaskId    string
	ExecuteAt string
	Status    string
	Success   bool
	Message   string
}

type RequestOrderModificationRequest struct {
	OrderId          uint64
	UserId           uint64
	ModificationType string
	Reason           string
	RequesterType    string
}

type RequestOrderModificationResponse struct {
	RequestNo        string
	ModificationType string
	Status           string
	ReviewRequired   bool
	AutoApprove      bool
	CreatedAt        string
}

type ApproveModificationRequest struct {
	RequestId  uint64
	ReviewerId uint64
	ReviewNote string
}

type ApproveModificationResponse struct {
	Success bool
	Message string
}

type RejectModificationRequest struct {
	RequestId  uint64
	ReviewerId uint64
	ReviewNote string
}

type RejectModificationResponse struct {
	Success bool
	Message string
}

type GetModificationHistoryRequest struct {
	OrderId uint64
	Page    int32
}

type GetModificationHistoryResponse struct {
}

package gateway

// PaymentRequest 支付请求
type PaymentRequest struct {
	OrderNo string // 订单号
	Amount  uint64 // 支付金额（分）
	Subject string // 商品标题
	Body    string // 商品描述
	UserID  uint64 // 用户ID
}

// PaymentResponse 支付响应
type PaymentResponse struct {
	PaymentURL    string // 支付URL（PC网站支付）
	OrderString   string // 订单字符串（APP支付）
	QRCode        string // 二维码内容（扫码支付）
	TransactionID string // 交易流水号
}

// PaymentQueryResponse 支付查询响应
type PaymentQueryResponse struct {
	OrderNo       string // 订单号
	TransactionID string // 交易流水号
	Status        string // 支付状态
	Amount        string // 支付金额
	PaidAt        string // 支付时间
}

// RefundRequest 退款请求
type RefundRequest struct {
	OrderNo      string // 订单号
	RefundNo     string // 退款单号
	RefundAmount uint64 // 退款金额（分）
	RefundReason string // 退款原因
}

// RefundResponse 退款响应
type RefundResponse struct {
	RefundNo      string // 退款单号
	TransactionID string // 交易流水号
	RefundAmount  string // 退款金额
	Success       bool   // 是否成功
	Message       string // 消息
}

// PaymentGateway 支付网关接口
type PaymentGateway interface {
	// CreatePayment 创建支付订单
	CreatePayment(req *PaymentRequest) (*PaymentResponse, error)
	
	// QueryPayment 查询支付订单
	QueryPayment(orderNo string) (*PaymentQueryResponse, error)
	
	// CreateRefund 创建退款
	CreateRefund(req *RefundRequest) (*RefundResponse, error)
	
	// VerifyNotify 验证异步通知
	VerifyNotify(params map[string]string) (bool, error)
}

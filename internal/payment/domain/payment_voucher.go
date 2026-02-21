package domain

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"html/template"
	"image"
	"image/color"
	"image/png"
	"strings"
	"sync"
	"time"
)

const voucherSignatureSecret = "ecommerce-voucher-signature-v1"

// VoucherType 凭证类型
type VoucherType string

const (
	VoucherPayment    VoucherType = "PAYMENT"    // 支付凭证
	VoucherRefund     VoucherType = "REFUND"     // 退款凭证
	VoucherChargeback VoucherType = "CHARGEBACK" // 拒付凭证
	VoucherSettlement VoucherType = "SETTLEMENT" // 结算凭证
)

// VoucherStatus 凭证状态
type VoucherStatus string

const (
	VoucherPending    VoucherStatus = "PENDING"    // 待生成
	VoucherGenerated  VoucherStatus = "GENERATED"  // 已生成
	VoucherSent       VoucherStatus = "SENT"       // 已发送
	VoucherViewed     VoucherStatus = "VIEWED"     // 已查看
	VoucherDownloaded VoucherStatus = "DOWNLOADED" // 已下载
	VoucherArchived   VoucherStatus = "ARCHIVED"   // 已归档
)

// PaymentVoucher 支付凭证
type PaymentVoucher struct {
	ID          string        `json:"id"`
	VoucherNo   string        `json:"voucher_no"`
	PaymentID   string        `json:"payment_id"`
	UserID      string        `json:"user_id"`
	VoucherType VoucherType   `json:"voucher_type"`
	Status      VoucherStatus `json:"status"`

	// 凭证内容
	Title           string    `json:"title"`
	Content         string    `json:"content"`
	Amount          float64   `json:"amount"`
	Currency        string    `json:"currency"`
	PaymentMethod   string    `json:"payment_method"`
	TransactionDate time.Time `json:"transaction_date"`
	ReferenceNo     string    `json:"reference_no"`

	// 格式和存储
	Format           string `json:"format"` // PDF, HTML, IMAGE
	StoragePath      string `json:"storage_path"`
	StorageURL       string `json:"storage_url"`
	Hash             string `json:"hash"` // 用于验证完整性
	DigitalSignature string `json:"digital_signature"`

	// 分发信息
	RecipientEmail string     `json:"recipient_email"`
	SentAt         *time.Time `json:"sent_at"`
	ViewedAt       *time.Time `json:"viewed_at"`
	DownloadedAt   *time.Time `json:"downloaded_at"`

	// 元数据
	Metadata  map[string]interface{} `json:"metadata"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
	ExpiresAt time.Time              `json:"expires_at"`
}

// VoucherTemplate 凭证模板
type VoucherTemplate struct {
	ID              string      `json:"id"`
	TemplateName    string      `json:"template_name"`
	VoucherType     VoucherType `json:"voucher_type"`
	Format          string      `json:"format"`
	TemplateContent string      `json:"template_content"`
	Variables       []string    `json:"variables"`
	Styles          string      `json:"styles"`
	HeaderImage     string      `json:"header_image"`
	FooterText      string      `json:"footer_text"`
	IsDefault       bool        `json:"is_default"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
}

// VoucherManager 凭证管理器
type VoucherManager struct {
	voucherRepo    VoucherRepository
	templateRepo   VoucherTemplateRepository
	storageService StorageService
	emailService   EmailService
	mu             sync.RWMutex
	config         *VoucherConfig
	templates      map[VoucherType]*VoucherTemplate
}

// VoucherConfig 凭证配置
type VoucherConfig struct {
	DefaultFormat   string        `json:"default_format"`
	DefaultExpiry   time.Duration `json:"default_expiry"`
	AutoGenerate    bool          `json:"auto_generate"`
	AutoSend        bool          `json:"auto_send"`
	StorageProvider string        `json:"storage_provider"`
	EnableSignature bool          `json:"enable_signature"`
	EnableHash      bool          `json:"enable_hash"`
	RetentionPeriod time.Duration `json:"retention_period"`
}

// NewVoucherManager 创建凭证管理器
func NewVoucherManager(voucherRepo VoucherRepository, templateRepo VoucherTemplateRepository,
	storageService StorageService, emailService EmailService) *VoucherManager {

	return &VoucherManager{
		voucherRepo:    voucherRepo,
		templateRepo:   templateRepo,
		storageService: storageService,
		emailService:   emailService,
		config: &VoucherConfig{
			DefaultFormat:   "PDF",
			DefaultExpiry:   365 * 24 * time.Hour, // 1年
			AutoGenerate:    true,
			AutoSend:        true,
			StorageProvider: "S3",
			EnableSignature: true,
			EnableHash:      true,
			RetentionPeriod: 7 * 365 * 24 * time.Hour, // 7年
		},
		templates: make(map[VoucherType]*VoucherTemplate),
	}
}

// Initialize 初始化凭证管理器
func (vm *VoucherManager) Initialize(ctx context.Context) error {
	// 加载凭证模板
	templates, err := vm.templateRepo.GetAllTemplates(ctx)
	if err != nil {
		return fmt.Errorf("failed to load voucher templates: %w", err)
	}

	vm.mu.Lock()
	for _, template := range templates {
		vm.templates[template.VoucherType] = template
	}
	vm.mu.Unlock()

	return nil
}

// GenerateVoucher 生成凭证
func (vm *VoucherManager) GenerateVoucher(ctx context.Context, payment *Payment, voucherType VoucherType) (*PaymentVoucher, error) {
	// 获取凭证模板
	template, err := vm.getTemplate(voucherType)
	if err != nil {
		return nil, fmt.Errorf("failed to get voucher template: %w", err)
	}

	// 创建凭证
	voucher := &PaymentVoucher{
		ID:              generateVoucherID(),
		VoucherNo:       generateVoucherNo(),
		PaymentID:       fmt.Sprintf("%d", payment.ID),
		UserID:          fmt.Sprintf("%d", payment.UserID),
		VoucherType:     voucherType,
		Status:          VoucherPending,
		Format:          template.Format,
		Amount:          float64(payment.Amount),
		Currency:        payment.Currency,
		PaymentMethod:   payment.PaymentMethod,
		TransactionDate: payment.CreatedAt,
		ReferenceNo:     payment.TransactionID,
		RecipientEmail:  payment.UserEmail,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		ExpiresAt:       time.Now().Add(vm.config.DefaultExpiry),
		Metadata:        make(map[string]interface{}),
	}

	// 生成凭证内容
	content, err := vm.generateVoucherContent(template, voucher, payment)
	if err != nil {
		return nil, fmt.Errorf("failed to generate voucher content: %w", err)
	}

	voucher.Content = content
	voucher.Title = vm.generateVoucherTitle(voucherType, payment)

	// 生成数字签名和哈希
	if vm.config.EnableSignature {
		signature, err := vm.generateDigitalSignature(voucher)
		if err != nil {
			return nil, fmt.Errorf("failed to generate digital signature: %w", err)
		}
		voucher.DigitalSignature = signature
	}

	if vm.config.EnableHash {
		hash, err := vm.generateHash(voucher)
		if err != nil {
			return nil, fmt.Errorf("failed to generate hash: %w", err)
		}
		voucher.Hash = hash
	}

	// 保存凭证
	err = vm.voucherRepo.SaveVoucher(ctx, voucher)
	if err != nil {
		return nil, fmt.Errorf("failed to save voucher: %w", err)
	}

	// 生成凭证文件
	err = vm.generateVoucherFile(ctx, voucher)
	if err != nil {
		return nil, fmt.Errorf("failed to generate voucher file: %w", err)
	}

	// 更新状态
	voucher.Status = VoucherGenerated
	voucher.UpdatedAt = time.Now()

	err = vm.voucherRepo.UpdateVoucher(ctx, voucher)
	if err != nil {
		return nil, fmt.Errorf("failed to update voucher: %w", err)
	}

	// 自动发送
	if vm.config.AutoSend {
		go vm.sendVoucher(ctx, voucher)
	}

	return voucher, nil
}

// getTemplate 获取模板
func (vm *VoucherManager) getTemplate(voucherType VoucherType) (*VoucherTemplate, error) {
	vm.mu.RLock()
	template, exists := vm.templates[voucherType]
	vm.mu.RUnlock()

	if !exists {
		// 使用默认模板
		template = &VoucherTemplate{
			TemplateName:    "Default Template",
			VoucherType:     voucherType,
			Format:          vm.config.DefaultFormat,
			TemplateContent: vm.getDefaultTemplate(voucherType),
			Variables:       []string{"title", "content", "amount", "currency", "date", "reference"},
			IsDefault:       true,
		}
	}

	return template, nil
}

// getDefaultTemplate 获取默认模板
func (vm *VoucherManager) getDefaultTemplate(voucherType VoucherType) string {
	switch voucherType {
	case VoucherPayment:
		return `Payment Voucher
		
Transaction Details:
- Amount: {{amount}} {{currency}}
- Date: {{date}}
- Reference: {{reference}}
- Status: Completed

Thank you for your payment.`

	case VoucherRefund:
		return `Refund Voucher
		
Refund Details:
- Amount: {{amount}} {{currency}}
- Date: {{date}}
- Reference: {{reference}}
- Status: Refunded

Your refund has been processed.`

	case VoucherChargeback:
		return `Chargeback Voucher
		
Chargeback Details:
- Amount: {{amount}} {{currency}}
- Date: {{date}}
- Reference: {{reference}}
- Status: Chargeback

This is a record of the chargeback.`

	default:
		return `Transaction Voucher
		
Transaction Details:
- Amount: {{amount}} {{currency}}
- Date: {{date}}
- Reference: {{reference}}
- Status: {{status}}`
	}
}

// generateVoucherContent 生成凭证内容
func (vm *VoucherManager) generateVoucherContent(template *VoucherTemplate, voucher *PaymentVoucher, payment *Payment) (string, error) {
	// 替换模板变量
	content := template.TemplateContent

	// 替换变量
	replacements := map[string]string{
		"{{title}}":     voucher.Title,
		"{{amount}}":    fmt.Sprintf("%.2f", voucher.Amount),
		"{{currency}}":  voucher.Currency,
		"{{date}}":      voucher.TransactionDate.Format("2006-01-02 15:04:05"),
		"{{reference}}": voucher.ReferenceNo,
		"{{status}}":    payment.Status.String(),
		"{{method}}":    voucher.PaymentMethod,
		"{{user}}":      voucher.UserID,
	}

	for key, value := range replacements {
		content = replaceAll(content, key, value)
	}

	return content, nil
}

// generateVoucherTitle 生成凭证标题
func (vm *VoucherManager) generateVoucherTitle(voucherType VoucherType, payment *Payment) string {
	switch voucherType {
	case VoucherPayment:
		return fmt.Sprintf("Payment Receipt - %s", payment.TransactionID)
	case VoucherRefund:
		return fmt.Sprintf("Refund Receipt - %s", payment.TransactionID)
	case VoucherChargeback:
		return fmt.Sprintf("Chargeback Notice - %s", payment.TransactionID)
	default:
		return fmt.Sprintf("Transaction Voucher - %s", payment.TransactionID)
	}
}

// generateDigitalSignature 生成数字签名
func (vm *VoucherManager) generateDigitalSignature(voucher *PaymentVoucher) (string, error) {
	if voucher == nil {
		return "", fmt.Errorf("voucher is nil")
	}
	payload := buildVoucherIntegrityPayload(voucher)
	mac := hmac.New(sha256.New, []byte(voucherSignatureSecret))
	_, _ = mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil)), nil
}

// generateHash 生成哈希
func (vm *VoucherManager) generateHash(voucher *PaymentVoucher) (string, error) {
	if voucher == nil {
		return "", fmt.Errorf("voucher is nil")
	}
	payload := buildVoucherIntegrityPayload(voucher)
	sum := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(sum[:]), nil
}

// generateVoucherFile 生成凭证文件
func (vm *VoucherManager) generateVoucherFile(ctx context.Context, voucher *PaymentVoucher) error {
	// 根据格式生成文件
	var fileData []byte
	var err error

	switch voucher.Format {
	case "PDF":
		fileData, err = vm.generatePDF(voucher)
	case "HTML":
		fileData, err = vm.generateHTML(voucher)
	case "IMAGE":
		fileData, err = vm.generateImage(voucher)
	default:
		return fmt.Errorf("unsupported format: %s", voucher.Format)
	}

	if err != nil {
		return fmt.Errorf("failed to generate file: %w", err)
	}

	// 存储文件
	storagePath, storageURL, err := vm.storageService.StoreFile(ctx, fileData, voucher.ID, voucher.Format)
	if err != nil {
		return fmt.Errorf("failed to store file: %w", err)
	}

	voucher.StoragePath = storagePath
	voucher.StorageURL = storageURL
	voucher.UpdatedAt = time.Now()

	return nil
}

// generatePDF 生成PDF
func (vm *VoucherManager) generatePDF(voucher *PaymentVoucher) ([]byte, error) {
	if voucher == nil {
		return nil, fmt.Errorf("voucher is nil")
	}
	// 轻量占位 PDF 内容（生产环境可替换为专业 PDF 引擎）。
	content := strings.Builder{}
	content.WriteString("%PDF-1.4\n")
	content.WriteString("%Voucher\n")
	content.WriteString("1 0 obj << /Type /Catalog /Pages 2 0 R >> endobj\n")
	content.WriteString("2 0 obj << /Type /Pages /Kids [3 0 R] /Count 1 >> endobj\n")
	content.WriteString("3 0 obj << /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] >> endobj\n")
	content.WriteString("4 0 obj << /Length 5 0 R >> stream\n")
	content.WriteString(fmt.Sprintf("Voucher No: %s\n", voucher.VoucherNo))
	content.WriteString(fmt.Sprintf("Title: %s\n", voucher.Title))
	content.WriteString(fmt.Sprintf("Amount: %.2f %s\n", voucher.Amount, voucher.Currency))
	content.WriteString(fmt.Sprintf("Date: %s\n", voucher.TransactionDate.Format("2006-01-02 15:04:05")))
	content.WriteString(fmt.Sprintf("Reference: %s\n", voucher.ReferenceNo))
	content.WriteString(voucher.Content + "\n")
	content.WriteString("endstream endobj\n")
	content.WriteString("5 0 obj 0 endobj\n")
	content.WriteString("xref\n0 6\n0000000000 65535 f \n")
	content.WriteString("trailer << /Root 1 0 R /Size 6 >>\nstartxref\n0\n%%EOF")
	return []byte(content.String()), nil
}

// generateHTML 生成HTML
func (vm *VoucherManager) generateHTML(voucher *PaymentVoucher) ([]byte, error) {
	if voucher == nil {
		return nil, fmt.Errorf("voucher is nil")
	}
	tpl := `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>{{.Title}}</title>
  <style>
    body { font-family: "Helvetica Neue", Arial, sans-serif; color: #1f2937; margin: 24px; line-height: 1.5; }
    .card { max-width: 760px; margin: 0 auto; border: 1px solid #e5e7eb; border-radius: 12px; padding: 20px; }
    .meta { font-size: 14px; color: #4b5563; }
    .amount { font-size: 26px; font-weight: 700; margin: 12px 0; }
    .content { white-space: pre-wrap; background: #f9fafb; padding: 12px; border-radius: 8px; }
  </style>
</head>
<body>
  <div class="card">
    <h1>{{.Title}}</h1>
    <div class="amount">{{printf "%.2f" .Amount}} {{.Currency}}</div>
    <div class="meta">Voucher No: {{.VoucherNo}}</div>
    <div class="meta">Payment ID: {{.PaymentID}}</div>
    <div class="meta">Date: {{.TransactionDate}}</div>
    <div class="meta">Reference: {{.ReferenceNo}}</div>
    <hr/>
    <div class="content">{{.Content}}</div>
  </div>
</body>
</html>`
	view := struct {
		Title           string
		Amount          float64
		Currency        string
		VoucherNo       string
		PaymentID       string
		TransactionDate string
		ReferenceNo     string
		Content         string
	}{
		Title:           voucher.Title,
		Amount:          voucher.Amount,
		Currency:        voucher.Currency,
		VoucherNo:       voucher.VoucherNo,
		PaymentID:       voucher.PaymentID,
		TransactionDate: voucher.TransactionDate.Format("2006-01-02 15:04:05"),
		ReferenceNo:     voucher.ReferenceNo,
		Content:         voucher.Content,
	}
	parsed, err := template.New("voucher").Parse(tpl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse html template: %w", err)
	}
	var buf bytes.Buffer
	if err := parsed.Execute(&buf, view); err != nil {
		return nil, fmt.Errorf("failed to render html template: %w", err)
	}
	return buf.Bytes(), nil
}

// generateImage 生成图片
func (vm *VoucherManager) generateImage(voucher *PaymentVoucher) ([]byte, error) {
	if voucher == nil {
		return nil, fmt.Errorf("voucher is nil")
	}
	width, height := 1000, 560
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	seed := sha256.Sum256([]byte(voucher.VoucherNo + voucher.ReferenceNo))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r := uint8((x + int(seed[0])) % 255)
			g := uint8((y + int(seed[1])) % 255)
			b := uint8((x/2 + y/3 + int(seed[2])) % 255)
			if y < 90 {
				r = seed[3]
				g = seed[4]
				b = seed[5]
			}
			img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("failed to encode png: %w", err)
	}
	return buf.Bytes(), nil
}

// sendVoucher 发送凭证
func (vm *VoucherManager) sendVoucher(ctx context.Context, voucher *PaymentVoucher) error {
	// 检查是否应该发送
	if voucher.RecipientEmail == "" {
		return fmt.Errorf("recipient email is empty")
	}

	// 创建邮件
	email := &Email{
		To:      voucher.RecipientEmail,
		Subject: voucher.Title,
		Body:    vm.generateEmailBody(voucher),
		Attachments: []*Attachment{
			{
				FileName: fmt.Sprintf("%s.%s", voucher.VoucherNo, getFileExtension(voucher.Format)),
				Content:  []byte("Attachment content"), // 实际应该从存储获取
				MimeType: getMimeType(voucher.Format),
			},
		},
	}

	// 发送邮件
	err := vm.emailService.SendEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	// 更新凭证状态
	voucher.Status = VoucherSent
	sentAt := time.Now()
	voucher.SentAt = &sentAt
	voucher.UpdatedAt = time.Now()

	err = vm.voucherRepo.UpdateVoucher(ctx, voucher)
	if err != nil {
		return fmt.Errorf("failed to update voucher: %w", err)
	}

	return nil
}

// generateEmailBody 生成邮件正文
func (vm *VoucherManager) generateEmailBody(voucher *PaymentVoucher) string {
	return fmt.Sprintf(`
Dear Customer,

Please find attached your %s.

Transaction Details:
- Voucher No: %s
- Amount: %.2f %s
- Date: %s
- Reference: %s

You can also download the voucher from: %s

Thank you for your business.

Best regards,
Payment Team
`, voucher.Title, voucher.VoucherNo, voucher.Amount, voucher.Currency,
		voucher.TransactionDate.Format("2006-01-02"), voucher.ReferenceNo, voucher.StorageURL)
}

// GetVoucher 获取凭证
func (vm *VoucherManager) GetVoucher(ctx context.Context, voucherID string) (*PaymentVoucher, error) {
	return vm.voucherRepo.GetVoucher(ctx, voucherID)
}

// GetVoucherByPayment 通过支付获取凭证
func (vm *VoucherManager) GetVoucherByPayment(ctx context.Context, paymentID string, voucherType VoucherType) (*PaymentVoucher, error) {
	return vm.voucherRepo.GetVoucherByPayment(ctx, paymentID, voucherType)
}

// DownloadVoucher 下载凭证
func (vm *VoucherManager) DownloadVoucher(ctx context.Context, voucherID string) ([]byte, string, error) {
	// 获取凭证
	voucher, err := vm.GetVoucher(ctx, voucherID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get voucher: %w", err)
	}

	// 更新下载时间
	downloadedAt := time.Now()
	voucher.DownloadedAt = &downloadedAt
	voucher.Status = VoucherDownloaded
	voucher.UpdatedAt = time.Now()

	err = vm.voucherRepo.UpdateVoucher(ctx, voucher)
	if err != nil {
		return nil, "", fmt.Errorf("failed to update voucher: %w", err)
	}

	// 从存储获取文件
	fileData, err := vm.storageService.GetFile(ctx, voucher.StoragePath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get file: %w", err)
	}

	return fileData, getMimeType(voucher.Format), nil
}

// VerifyVoucher 验证凭证
func (vm *VoucherManager) VerifyVoucher(ctx context.Context, voucherID string) (bool, error) {
	// 获取凭证
	voucher, err := vm.GetVoucher(ctx, voucherID)
	if err != nil {
		return false, fmt.Errorf("failed to get voucher: %w", err)
	}

	// 验证哈希
	if vm.config.EnableHash && voucher.Hash != "" {
		expectedHash, err := vm.generateHash(voucher)
		if err != nil {
			return false, fmt.Errorf("failed to generate hash for verification: %w", err)
		}
		if subtle.ConstantTimeCompare([]byte(expectedHash), []byte(voucher.Hash)) != 1 {
			return false, fmt.Errorf("voucher hash verification failed")
		}
	}

	// 验证签名
	if vm.config.EnableSignature && voucher.DigitalSignature != "" {
		expectedSig, err := vm.generateDigitalSignature(voucher)
		if err != nil {
			return false, fmt.Errorf("failed to generate signature for verification: %w", err)
		}
		if subtle.ConstantTimeCompare([]byte(expectedSig), []byte(voucher.DigitalSignature)) != 1 {
			return false, fmt.Errorf("voucher signature verification failed")
		}
	}

	// 验证过期时间
	if time.Now().After(voucher.ExpiresAt) {
		return false, fmt.Errorf("voucher has expired")
	}

	return true, nil
}

// ArchiveOldVouchers 归档旧凭证
func (vm *VoucherManager) ArchiveOldVouchers(ctx context.Context) error {
	// 获取过期的凭证
	expiredVouchers, err := vm.voucherRepo.GetExpiredVouchers(ctx)
	if err != nil {
		return fmt.Errorf("failed to get expired vouchers: %w", err)
	}

	// 归档每个凭证
	for _, voucher := range expiredVouchers {
		if voucher.Status != VoucherArchived {
			voucher.Status = VoucherArchived
			voucher.UpdatedAt = time.Now()

			err = vm.voucherRepo.UpdateVoucher(ctx, voucher)
			if err != nil {
				fmt.Printf("Failed to archive voucher %s: %v\n", voucher.ID, err)
			}
		}
	}

	return nil
}

// Helper functions

func generateVoucherID() string {
	return fmt.Sprintf("VOUCHER_%d", time.Now().UnixNano())
}

func generateVoucherNo() string {
	return fmt.Sprintf("VCH%d", time.Now().UnixNano())
}

func replaceAll(s, old, new string) string {
	return strings.ReplaceAll(s, old, new)
}

func buildVoucherIntegrityPayload(voucher *PaymentVoucher) string {
	if voucher == nil {
		return ""
	}
	parts := []string{
		voucher.ID,
		voucher.VoucherNo,
		voucher.PaymentID,
		voucher.UserID,
		string(voucher.VoucherType),
		voucher.Title,
		voucher.Content,
		fmt.Sprintf("%.2f", voucher.Amount),
		voucher.Currency,
		voucher.PaymentMethod,
		voucher.TransactionDate.UTC().Format(time.RFC3339Nano),
		voucher.ReferenceNo,
		voucher.Format,
		voucher.CreatedAt.UTC().Format(time.RFC3339Nano),
		voucher.ExpiresAt.UTC().Format(time.RFC3339Nano),
	}
	return strings.Join(parts, "|")
}

func getFileExtension(format string) string {
	switch format {
	case "PDF":
		return "pdf"
	case "HTML":
		return "html"
	case "IMAGE":
		return "png"
	default:
		return "txt"
	}
}

func getMimeType(format string) string {
	switch format {
	case "PDF":
		return "application/pdf"
	case "HTML":
		return "text/html"
	case "IMAGE":
		return "image/png"
	default:
		return "text/plain"
	}
}

// Data structures

type Email struct {
	To          string        `json:"to"`
	Subject     string        `json:"subject"`
	Body        string        `json:"body"`
	Attachments []*Attachment `json:"attachments"`
}

type Attachment struct {
	FileName string `json:"file_name"`
	Content  []byte `json:"content"`
	MimeType string `json:"mime_type"`
}

// Repository interfaces

type VoucherRepository interface {
	SaveVoucher(ctx context.Context, voucher *PaymentVoucher) error
	GetVoucher(ctx context.Context, voucherID string) (*PaymentVoucher, error)
	GetVoucherByPayment(ctx context.Context, paymentID string, voucherType VoucherType) (*PaymentVoucher, error)
	GetVouchersByUser(ctx context.Context, userID string, startDate, endDate time.Time) ([]*PaymentVoucher, error)
	GetExpiredVouchers(ctx context.Context) ([]*PaymentVoucher, error)
	UpdateVoucher(ctx context.Context, voucher *PaymentVoucher) error
	DeleteVoucher(ctx context.Context, voucherID string) error
}

type VoucherTemplateRepository interface {
	SaveTemplate(ctx context.Context, template *VoucherTemplate) error
	GetTemplate(ctx context.Context, voucherType VoucherType) (*VoucherTemplate, error)
	GetAllTemplates(ctx context.Context) ([]*VoucherTemplate, error)
	UpdateTemplate(ctx context.Context, template *VoucherTemplate) error
	DeleteTemplate(ctx context.Context, templateID string) error
}

// Service interfaces

type StorageService interface {
	StoreFile(ctx context.Context, data []byte, fileName, format string) (string, string, error)
	GetFile(ctx context.Context, storagePath string) ([]byte, error)
	DeleteFile(ctx context.Context, storagePath string) error
	GetFileURL(ctx context.Context, storagePath string) (string, error)
}

type EmailService interface {
	SendEmail(ctx context.Context, email *Email) error
	GetEmailStatus(ctx context.Context, emailID string) (string, error)
}

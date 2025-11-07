package gateway

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"go.uber.org/zap"
)

// AlipayConfig 支付宝配置
type AlipayConfig struct {
	AppID           string // 应用ID
	PrivateKey      string // 应用私钥
	AlipayPublicKey string // 支付宝公钥
	NotifyURL       string // 异步通知地址
	ReturnURL       string // 同步返回地址
	SignType        string // 签名类型，默认RSA2
	Charset         string // 字符集，默认utf-8
	Format          string // 数据格式，默认JSON
	GatewayURL      string // 支付宝网关地址
}

// AlipayGateway 支付宝支付网关
type AlipayGateway struct {
	config     *AlipayConfig
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	logger     *zap.Logger
}

// NewAlipayGateway 创建支付宝支付网关实例
func NewAlipayGateway(config *AlipayConfig, logger *zap.Logger) (*AlipayGateway, error) {
	// 解析私钥
	privateKey, err := parsePrivateKey(config.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("解析私钥失败: %w", err)
	}

	// 解析公钥
	publicKey, err := parsePublicKey(config.AlipayPublicKey)
	if err != nil {
		return nil, fmt.Errorf("解析公钥失败: %w", err)
	}

	// 设置默认值
	if config.SignType == "" {
		config.SignType = "RSA2"
	}
	if config.Charset == "" {
		config.Charset = "utf-8"
	}
	if config.Format == "" {
		config.Format = "JSON"
	}
	if config.GatewayURL == "" {
		config.GatewayURL = "https://openapi.alipay.com/gateway.do"
	}

	return &AlipayGateway{
		config:     config,
		privateKey: privateKey,
		publicKey:  publicKey,
		logger:     logger,
	}, nil
}

// CreatePayment 创建支付订单
func (g *AlipayGateway) CreatePayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error) {
	// 构建业务参数
	bizContent := map[string]interface{}{
		"out_trade_no": req.OrderNo,
		"total_amount": fmt.Sprintf("%.2f", float64(req.Amount)/100), // 分转元
		"subject":      req.Subject,
		"product_code": "FAST_INSTANT_TRADE_PAY",
	}

	if req.Body != "" {
		bizContent["body"] = req.Body
	}

	bizContentJSON, err := json.Marshal(bizContent)
	if err != nil {
		return nil, err
	}

	// 构建公共参数
	params := map[string]string{
		"app_id":      g.config.AppID,
		"method":      "alipay.trade.page.pay", // PC网站支付
		"format":      g.config.Format,
		"charset":     g.config.Charset,
		"sign_type":   g.config.SignType,
		"timestamp":   time.Now().Format("2006-01-02 15:04:05"),
		"version":     "1.0",
		"notify_url":  g.config.NotifyURL,
		"return_url":  g.config.ReturnURL,
		"biz_content": string(bizContentJSON),
	}

	// 生成签名
	sign, err := g.sign(params)
	if err != nil {
		return nil, err
	}
	params["sign"] = sign

	// 构建支付URL
	paymentURL := g.buildURL(params)

	g.logger.Info("创建支付宝支付订单",
		zap.String("orderNo", req.OrderNo),
		zap.Uint64("amount", req.Amount))

	return &PaymentResponse{
		PaymentURL:    paymentURL,
		TransactionID: req.OrderNo,
	}, nil
}

// CreateAppPayment 创建APP支付订单
func (g *AlipayGateway) CreateAppPayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error) {
	bizContent := map[string]interface{}{
		"out_trade_no": req.OrderNo,
		"total_amount": fmt.Sprintf("%.2f", float64(req.Amount)/100),
		"subject":      req.Subject,
		"product_code": "QUICK_MSECURITY_PAY",
	}

	bizContentJSON, err := json.Marshal(bizContent)
	if err != nil {
		return nil, err
	}

	params := map[string]string{
		"app_id":      g.config.AppID,
		"method":      "alipay.trade.app.pay",
		"format":      g.config.Format,
		"charset":     g.config.Charset,
		"sign_type":   g.config.SignType,
		"timestamp":   time.Now().Format("2006-01-02 15:04:05"),
		"version":     "1.0",
		"notify_url":  g.config.NotifyURL,
		"biz_content": string(bizContentJSON),
	}

	sign, err := g.sign(params)
	if err != nil {
		return nil, err
	}
	params["sign"] = sign

	// APP支付返回签名后的字符串
	orderString := g.buildOrderString(params)

	return &PaymentResponse{
		OrderString:   orderString,
		TransactionID: req.OrderNo,
	}, nil
}

// QueryPayment 查询支付订单
func (g *AlipayGateway) QueryPayment(ctx context.Context, orderNo string) (*PaymentQueryResponse, error) {
	bizContent := map[string]interface{}{
		"out_trade_no": orderNo,
	}

	bizContentJSON, err := json.Marshal(bizContent)
	if err != nil {
		return nil, err
	}

	params := map[string]string{
		"app_id":      g.config.AppID,
		"method":      "alipay.trade.query",
		"format":      g.config.Format,
		"charset":     g.config.Charset,
		"sign_type":   g.config.SignType,
		"timestamp":   time.Now().Format("2006-01-02 15:04:05"),
		"version":     "1.0",
		"biz_content": string(bizContentJSON),
	}

	sign, err := g.sign(params)
	if err != nil {
		return nil, err
	}
	params["sign"] = sign

	// 发送请求
	resp, err := g.doRequest(params)
	if err != nil {
		return nil, err
	}

	// 解析响应
	var result struct {
		AlipayTradeQueryResponse struct {
			Code       string `json:"code"`
			Msg        string `json:"msg"`
			TradeNo    string `json:"trade_no"`
			TradeStatus string `json:"trade_status"`
			TotalAmount string `json:"total_amount"`
		} `json:"alipay_trade_query_response"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	if result.AlipayTradeQueryResponse.Code != "10000" {
		return nil, fmt.Errorf("查询失败: %s", result.AlipayTradeQueryResponse.Msg)
	}

	return &PaymentQueryResponse{
		OrderNo:       orderNo,
		TransactionID: result.AlipayTradeQueryResponse.TradeNo,
		Status:        result.AlipayTradeQueryResponse.TradeStatus,
		Amount:        result.AlipayTradeQueryResponse.TotalAmount,
	}, nil
}

// CreateRefund 创建退款
func (g *AlipayGateway) CreateRefund(ctx context.Context, req *RefundRequest) (*RefundResponse, error) {
	bizContent := map[string]interface{}{
		"out_trade_no":   req.OrderNo,
		"refund_amount":  fmt.Sprintf("%.2f", float64(req.RefundAmount)/100),
		"out_request_no": req.RefundNo,
	}

	if req.RefundReason != "" {
		bizContent["refund_reason"] = req.RefundReason
	}

	bizContentJSON, err := json.Marshal(bizContent)
	if err != nil {
		return nil, err
	}

	params := map[string]string{
		"app_id":      g.config.AppID,
		"method":      "alipay.trade.refund",
		"format":      g.config.Format,
		"charset":     g.config.Charset,
		"sign_type":   g.config.SignType,
		"timestamp":   time.Now().Format("2006-01-02 15:04:05"),
		"version":     "1.0",
		"biz_content": string(bizContentJSON),
	}

	sign, err := g.sign(params)
	if err != nil {
		return nil, err
	}
	params["sign"] = sign

	resp, err := g.doRequest(params)
	if err != nil {
		return nil, err
	}

	var result struct {
		AlipayTradeRefundResponse struct {
			Code         string `json:"code"`
			Msg          string `json:"msg"`
			TradeNo      string `json:"trade_no"`
			RefundFee    string `json:"refund_fee"`
			GmtRefundPay string `json:"gmt_refund_pay"`
		} `json:"alipay_trade_refund_response"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	if result.AlipayTradeRefundResponse.Code != "10000" {
		return nil, fmt.Errorf("退款失败: %s", result.AlipayTradeRefundResponse.Msg)
	}

	g.logger.Info("支付宝退款成功",
		zap.String("orderNo", req.OrderNo),
		zap.String("refundNo", req.RefundNo),
		zap.Uint64("amount", req.RefundAmount))

	return &RefundResponse{
		RefundNo:      req.RefundNo,
		TransactionID: result.AlipayTradeRefundResponse.TradeNo,
		RefundAmount:  result.AlipayTradeRefundResponse.RefundFee,
		Success:       true,
	}, nil
}

// VerifyNotify 验证异步通知
func (g *AlipayGateway) VerifyNotify(params map[string]string) (bool, error) {
	// 获取签名
	sign := params["sign"]
	if sign == "" {
		return false, errors.New("签名为空")
	}

	// 移除sign和sign_type
	delete(params, "sign")
	delete(params, "sign_type")

	// 验证签名
	return g.verify(params, sign)
}

// sign 生成签名
func (g *AlipayGateway) sign(params map[string]string) (string, error) {
	// 排序参数
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 拼接字符串
	var signStr strings.Builder
	for i, k := range keys {
		if i > 0 {
			signStr.WriteString("&")
		}
		signStr.WriteString(k)
		signStr.WriteString("=")
		signStr.WriteString(params[k])
	}

	// 计算签名
	var hash crypto.Hash
	if g.config.SignType == "RSA2" {
		hash = crypto.SHA256
	} else {
		hash = crypto.SHA1
	}

	h := hash.New()
	h.Write([]byte(signStr.String()))
	hashed := h.Sum(nil)

	signature, err := rsa.SignPKCS1v15(rand.Reader, g.privateKey, hash, hashed)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(signature), nil
}

// verify 验证签名
func (g *AlipayGateway) verify(params map[string]string, sign string) (bool, error) {
	// 排序参数
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 拼接字符串
	var signStr strings.Builder
	for i, k := range keys {
		if i > 0 {
			signStr.WriteString("&")
		}
		signStr.WriteString(k)
		signStr.WriteString("=")
		signStr.WriteString(params[k])
	}

	// 解码签名
	signBytes, err := base64.StdEncoding.DecodeString(sign)
	if err != nil {
		return false, err
	}

	// 验证签名
	var hash crypto.Hash
	if g.config.SignType == "RSA2" {
		hash = crypto.SHA256
	} else {
		hash = crypto.SHA1
	}

	h := hash.New()
	h.Write([]byte(signStr.String()))
	hashed := h.Sum(nil)

	err = rsa.VerifyPKCS1v15(g.publicKey, hash, hashed, signBytes)
	return err == nil, err
}

// buildURL 构建支付URL
func (g *AlipayGateway) buildURL(params map[string]string) string {
	values := url.Values{}
	for k, v := range params {
		values.Set(k, v)
	}
	return g.config.GatewayURL + "?" + values.Encode()
}

// buildOrderString 构建订单字符串（用于APP支付）
func (g *AlipayGateway) buildOrderString(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var orderStr strings.Builder
	for i, k := range keys {
		if i > 0 {
			orderStr.WriteString("&")
		}
		orderStr.WriteString(k)
		orderStr.WriteString("=")
		orderStr.WriteString(url.QueryEscape(params[k]))
	}

	return orderStr.String()
}

// doRequest 发送HTTP请求
func (g *AlipayGateway) doRequest(params map[string]string) ([]byte, error) {
	values := url.Values{}
	for k, v := range params {
		values.Set(k, v)
	}

	resp, err := http.PostForm(g.config.GatewayURL, values)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// parsePrivateKey 解析私钥
func parsePrivateKey(privateKeyStr string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(privateKeyStr))
	if block == nil {
		return nil, errors.New("私钥格式错误")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		// 尝试PKCS8格式
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		privateKey = key.(*rsa.PrivateKey)
	}

	return privateKey, nil
}

// parsePublicKey 解析公钥
func parsePublicKey(publicKeyStr string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(publicKeyStr))
	if block == nil {
		return nil, errors.New("公钥格式错误")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	publicKey, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("公钥类型错误")
	}

	return publicKey, nil
}

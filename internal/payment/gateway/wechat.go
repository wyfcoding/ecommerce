package gateway

import (
	"context"
	"crypto/md5"
	"crypto/tls"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"go.uber.org/zap"
)

// WechatConfig 微信支付配置
type WechatConfig struct {
	AppID      string // 应用ID
	MchID      string // 商户号
	APIKey     string // API密钥
	NotifyURL  string // 异步通知地址
	CertFile   string // 证书文件路径
	KeyFile    string // 密钥文件路径
	GatewayURL string // 微信支付网关地址
}

// WechatGateway 微信支付网关
type WechatGateway struct {
	config     *WechatConfig
	httpClient *http.Client
	logger     *zap.Logger
}

// NewWechatGateway 创建微信支付网关实例
func NewWechatGateway(config *WechatConfig, logger *zap.Logger) (*WechatGateway, error) {
	// 设置默认值
	if config.GatewayURL == "" {
		config.GatewayURL = "https://api.mch.weixin.qq.com"
	}

	// 创建HTTP客户端
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	// 如果有证书，配置双向认证
	if config.CertFile != "" && config.KeyFile != "" {
		cert, err := tls.LoadX509KeyPair(config.CertFile, config.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("加载证书失败: %w", err)
		}

		httpClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates: []tls.Certificate{cert},
			},
		}
	}

	return &WechatGateway{
		config:     config,
		httpClient: httpClient,
		logger:     logger,
	}, nil
}

// CreatePayment 创建支付订单（统一下单）
func (g *WechatGateway) CreatePayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error) {
	// 构建请求参数
	params := map[string]string{
		"appid":            g.config.AppID,
		"mch_id":           g.config.MchID,
		"nonce_str":        generateNonceStr(),
		"body":             req.Subject,
		"out_trade_no":     req.OrderNo,
		"total_fee":        fmt.Sprintf("%d", req.Amount), // 单位：分
		"spbill_create_ip": "127.0.0.1",                   // TODO: 获取真实IP
		"notify_url":       g.config.NotifyURL,
		"trade_type":       "NATIVE", // 扫码支付
	}

	// 生成签名
	params["sign"] = g.sign(params)

	// 发送请求
	resp, err := g.doRequest(g.config.GatewayURL+"/pay/unifiedorder", params)
	if err != nil {
		return nil, err
	}

	// 解析响应
	var result UnifiedOrderResponse
	if err := xml.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	if result.ReturnCode != "SUCCESS" {
		return nil, fmt.Errorf("请求失败: %s", result.ReturnMsg)
	}

	if result.ResultCode != "SUCCESS" {
		return nil, fmt.Errorf("下单失败: %s", result.ErrCodeDes)
	}

	g.logger.Info("创建微信支付订单",
		zap.String("orderNo", req.OrderNo),
		zap.Uint64("amount", req.Amount))

	return &PaymentResponse{
		QRCode:        result.CodeURL,
		TransactionID: result.PrepayID,
	}, nil
}

// CreateAppPayment 创建APP支付订单
func (g *WechatGateway) CreateAppPayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error) {
	params := map[string]string{
		"appid":            g.config.AppID,
		"mch_id":           g.config.MchID,
		"nonce_str":        generateNonceStr(),
		"body":             req.Subject,
		"out_trade_no":     req.OrderNo,
		"total_fee":        fmt.Sprintf("%d", req.Amount),
		"spbill_create_ip": "127.0.0.1",
		"notify_url":       g.config.NotifyURL,
		"trade_type":       "APP",
	}

	params["sign"] = g.sign(params)

	resp, err := g.doRequest(g.config.GatewayURL+"/pay/unifiedorder", params)
	if err != nil {
		return nil, err
	}

	var result UnifiedOrderResponse
	if err := xml.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	if result.ReturnCode != "SUCCESS" || result.ResultCode != "SUCCESS" {
		return nil, fmt.Errorf("下单失败")
	}

	// 构建APP调起支付参数
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	appParams := map[string]string{
		"appid":     g.config.AppID,
		"partnerid": g.config.MchID,
		"prepayid":  result.PrepayID,
		"package":   "Sign=WXPay",
		"noncestr":  generateNonceStr(),
		"timestamp": timestamp,
	}
	appParams["sign"] = g.sign(appParams)

	// 将参数转换为字符串
	orderString := g.buildOrderString(appParams)

	return &PaymentResponse{
		OrderString:   orderString,
		TransactionID: result.PrepayID,
	}, nil
}

// QueryPayment 查询支付订单
func (g *WechatGateway) QueryPayment(ctx context.Context, orderNo string) (*PaymentQueryResponse, error) {
	params := map[string]string{
		"appid":         g.config.AppID,
		"mch_id":        g.config.MchID,
		"out_trade_no":  orderNo,
		"nonce_str":     generateNonceStr(),
	}

	params["sign"] = g.sign(params)

	resp, err := g.doRequest(g.config.GatewayURL+"/pay/orderquery", params)
	if err != nil {
		return nil, err
	}

	var result OrderQueryResponse
	if err := xml.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	if result.ReturnCode != "SUCCESS" {
		return nil, fmt.Errorf("查询失败: %s", result.ReturnMsg)
	}

	if result.ResultCode != "SUCCESS" {
		return nil, fmt.Errorf("查询失败: %s", result.ErrCodeDes)
	}

	return &PaymentQueryResponse{
		OrderNo:       orderNo,
		TransactionID: result.TransactionID,
		Status:        result.TradeState,
		Amount:        result.TotalFee,
		PaidAt:        result.TimeEnd,
	}, nil
}

// CreateRefund 创建退款
func (g *WechatGateway) CreateRefund(ctx context.Context, req *RefundRequest) (*RefundResponse, error) {
	params := map[string]string{
		"appid":         g.config.AppID,
		"mch_id":        g.config.MchID,
		"nonce_str":     generateNonceStr(),
		"out_trade_no":  req.OrderNo,
		"out_refund_no": req.RefundNo,
		"total_fee":     fmt.Sprintf("%d", req.RefundAmount), // 原订单金额
		"refund_fee":    fmt.Sprintf("%d", req.RefundAmount), // 退款金额
	}

	if req.RefundReason != "" {
		params["refund_desc"] = req.RefundReason
	}

	params["sign"] = g.sign(params)

	// 退款需要使用证书
	resp, err := g.doRequest(g.config.GatewayURL+"/secapi/pay/refund", params)
	if err != nil {
		return nil, err
	}

	var result RefundResponseXML
	if err := xml.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	if result.ReturnCode != "SUCCESS" {
		return nil, fmt.Errorf("退款失败: %s", result.ReturnMsg)
	}

	if result.ResultCode != "SUCCESS" {
		return nil, fmt.Errorf("退款失败: %s", result.ErrCodeDes)
	}

	g.logger.Info("微信退款成功",
		zap.String("orderNo", req.OrderNo),
		zap.String("refundNo", req.RefundNo),
		zap.Uint64("amount", req.RefundAmount))

	return &RefundResponse{
		RefundNo:      req.RefundNo,
		TransactionID: result.RefundID,
		RefundAmount:  result.RefundFee,
		Success:       true,
	}, nil
}

// VerifyNotify 验证异步通知
func (g *WechatGateway) VerifyNotify(params map[string]string) (bool, error) {
	// 获取签名
	sign := params["sign"]
	if sign == "" {
		return false, errors.New("签名为空")
	}

	// 移除sign
	delete(params, "sign")

	// 验证签名
	expectedSign := g.sign(params)
	return sign == expectedSign, nil
}

// sign 生成签名
func (g *WechatGateway) sign(params map[string]string) string {
	// 排序参数
	keys := make([]string, 0, len(params))
	for k := range params {
		if k != "sign" && params[k] != "" {
			keys = append(keys, k)
		}
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

	// 添加API密钥
	signStr.WriteString("&key=")
	signStr.WriteString(g.config.APIKey)

	// MD5加密
	hash := md5.Sum([]byte(signStr.String()))
	return strings.ToUpper(hex.EncodeToString(hash[:]))
}

// buildOrderString 构建订单字符串
func (g *WechatGateway) buildOrderString(params map[string]string) string {
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
		orderStr.WriteString(params[k])
	}

	return orderStr.String()
}

// doRequest 发送HTTP请求
func (g *WechatGateway) doRequest(url string, params map[string]string) ([]byte, error) {
	// 构建XML请求体
	xmlData := buildXML(params)

	req, err := http.NewRequest("POST", url, strings.NewReader(xmlData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/xml")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// buildXML 构建XML
func buildXML(params map[string]string) string {
	var xml strings.Builder
	xml.WriteString("<xml>")
	for k, v := range params {
		xml.WriteString("<")
		xml.WriteString(k)
		xml.WriteString("><![CDATA[")
		xml.WriteString(v)
		xml.WriteString("]]></")
		xml.WriteString(k)
		xml.WriteString(">")
	}
	xml.WriteString("</xml>")
	return xml.String()
}

// generateNonceStr 生成随机字符串
func generateNonceStr() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// 响应结构体
type UnifiedOrderResponse struct {
	XMLName    xml.Name `xml:"xml"`
	ReturnCode string   `xml:"return_code"`
	ReturnMsg  string   `xml:"return_msg"`
	ResultCode string   `xml:"result_code"`
	ErrCode    string   `xml:"err_code"`
	ErrCodeDes string   `xml:"err_code_des"`
	PrepayID   string   `xml:"prepay_id"`
	CodeURL    string   `xml:"code_url"`
}

type OrderQueryResponse struct {
	XMLName       xml.Name `xml:"xml"`
	ReturnCode    string   `xml:"return_code"`
	ReturnMsg     string   `xml:"return_msg"`
	ResultCode    string   `xml:"result_code"`
	ErrCode       string   `xml:"err_code"`
	ErrCodeDes    string   `xml:"err_code_des"`
	TransactionID string   `xml:"transaction_id"`
	TradeState    string   `xml:"trade_state"`
	TotalFee      string   `xml:"total_fee"`
	TimeEnd       string   `xml:"time_end"`
}

type RefundResponseXML struct {
	XMLName    xml.Name `xml:"xml"`
	ReturnCode string   `xml:"return_code"`
	ReturnMsg  string   `xml:"return_msg"`
	ResultCode string   `xml:"result_code"`
	ErrCode    string   `xml:"err_code"`
	ErrCodeDes string   `xml:"err_code_des"`
	RefundID   string   `xml:"refund_id"`
	RefundFee  string   `xml:"refund_fee"`
}

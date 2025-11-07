package gateway

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.uber.org/zap"
)

// SFExpressConfig 顺丰快递配置
type SFExpressConfig struct {
	AppID      string // 应用ID
	AppSecret  string // 应用密钥
	CustomerID string // 客户编码
	GatewayURL string // 网关地址
	Sandbox    bool   // 是否沙箱环境
}

// SFExpressGateway 顺丰快递网关
type SFExpressGateway struct {
	config     *SFExpressConfig
	httpClient *http.Client
	logger     *zap.Logger
}

// NewSFExpressGateway 创建顺丰快递网关实例
func NewSFExpressGateway(config *SFExpressConfig, logger *zap.Logger) *SFExpressGateway {
	if config.GatewayURL == "" {
		if config.Sandbox {
			config.GatewayURL = "https://sfapi-sbox.sf-express.com/std/service"
		} else {
			config.GatewayURL = "https://sfapi.sf-express.com/std/service"
		}
	}

	return &SFExpressGateway{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// CreateWaybill 创建电子面单
func (g *SFExpressGateway) CreateWaybill(ctx context.Context, req *CreateWaybillRequest) (*CreateWaybillResponse, error) {
	// 构建请求参数
	bizData := map[string]interface{}{
		"orderId": fmt.Sprintf("SF%d", time.Now().UnixNano()),
		"expressType": "1", // 标准快递
		"payMethod": "1",   // 寄方付
		"cargoDetails": []map[string]interface{}{
			{
				"name":  req.GoodsName,
				"count": 1,
			},
		},
		"consigneeInfo": map[string]interface{}{
			"contact":  req.ReceiverName,
			"tel":      req.ReceiverPhone,
			"province": req.ReceiverProvince,
			"city":     req.ReceiverCity,
			"county":   req.ReceiverDistrict,
			"address":  req.ReceiverAddress,
		},
		"deliverInfo": map[string]interface{}{
			"contact":  req.SenderName,
			"tel":      req.SenderPhone,
			"province": req.SenderProvince,
			"city":     req.SenderCity,
			"county":   req.SenderDistrict,
			"address":  req.SenderAddress,
		},
	}

	if req.Remark != "" {
		bizData["remark"] = req.Remark
	}

	// 发送请求
	resp, err := g.doRequest(ctx, "EXP_RECE_CREATE_ORDER", bizData)
	if err != nil {
		return nil, err
	}

	// 解析响应
	var result struct {
		APIResultData struct {
			Success      bool   `json:"success"`
			ErrorCode    string `json:"errorCode"`
			ErrorMsg     string `json:"errorMsg"`
			WaybillNo    string `json:"waybillNo"`
			OriginCode   string `json:"originCode"`
			DestCode     string `json:"destCode"`
			FilterResult string `json:"filterResult"`
		} `json:"apiResultData"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	if !result.APIResultData.Success {
		return nil, fmt.Errorf("创建运单失败: %s - %s", 
			result.APIResultData.ErrorCode, 
			result.APIResultData.ErrorMsg)
	}

	g.logger.Info("创建顺丰运单成功",
		zap.String("waybillNo", result.APIResultData.WaybillNo))

	return &CreateWaybillResponse{
		TrackingNo: result.APIResultData.WaybillNo,
		WaybillNo:  result.APIResultData.WaybillNo,
		CreatedAt:  time.Now(),
	}, nil
}

// QueryTracking 查询物流轨迹
func (g *SFExpressGateway) QueryTracking(ctx context.Context, company ExpressCompany, trackingNo string) (*TrackingInfo, error) {
	// 构建请求参数
	bizData := map[string]interface{}{
		"trackingType": "1", // 根据运单号查询
		"trackingNumber": []string{trackingNo},
		"methodType": "1", // 标准路由查询
	}

	// 发送请求
	resp, err := g.doRequest(ctx, "EXP_RECE_SEARCH_ROUTES", bizData)
	if err != nil {
		return nil, err
	}

	// 解析响应
	var result struct {
		APIResultData struct {
			Success   bool   `json:"success"`
			ErrorCode string `json:"errorCode"`
			ErrorMsg  string `json:"errorMsg"`
			RouteResps []struct {
				MailNo string `json:"mailNo"`
				Routes []struct {
					AcceptTime    string `json:"acceptTime"`
					AcceptAddress string `json:"acceptAddress"`
					Remark        string `json:"remark"`
					OpCode        string `json:"opCode"`
				} `json:"routes"`
			} `json:"routeResps"`
		} `json:"apiResultData"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	if !result.APIResultData.Success {
		return nil, fmt.Errorf("查询物流失败: %s - %s",
			result.APIResultData.ErrorCode,
			result.APIResultData.ErrorMsg)
	}

	if len(result.APIResultData.RouteResps) == 0 {
		return nil, fmt.Errorf("未找到物流信息")
	}

	// 转换为标准格式
	routeResp := result.APIResultData.RouteResps[0]
	traces := make([]*TrackingTrace, 0, len(routeResp.Routes))
	
	for _, route := range routeResp.Routes {
		acceptTime, _ := time.Parse("2006-01-02 15:04:05", route.AcceptTime)
		traces = append(traces, &TrackingTrace{
			Time:        acceptTime,
			Status:      route.OpCode,
			Description: route.Remark,
			Location:    route.AcceptAddress,
		})
	}

	// 判断当前状态
	status := g.parseStatus(routeResp.Routes)

	return &TrackingInfo{
		Company:    ExpressCompanySF,
		TrackingNo: trackingNo,
		Status:     status,
		Traces:     traces,
		UpdatedAt:  time.Now(),
	}, nil
}

// CancelWaybill 取消运单
func (g *SFExpressGateway) CancelWaybill(ctx context.Context, company ExpressCompany, trackingNo string) error {
	bizData := map[string]interface{}{
		"orderId":   trackingNo,
		"waybillNo": trackingNo,
		"dealType":  "2", // 取消订单
	}

	resp, err := g.doRequest(ctx, "EXP_RECE_UPDATE_ORDER", bizData)
	if err != nil {
		return err
	}

	var result struct {
		APIResultData struct {
			Success   bool   `json:"success"`
			ErrorCode string `json:"errorCode"`
			ErrorMsg  string `json:"errorMsg"`
		} `json:"apiResultData"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}

	if !result.APIResultData.Success {
		return fmt.Errorf("取消运单失败: %s - %s",
			result.APIResultData.ErrorCode,
			result.APIResultData.ErrorMsg)
	}

	return nil
}

// CalculateFee 计算运费
func (g *SFExpressGateway) CalculateFee(ctx context.Context, req *CalculateFeeRequest) (*CalculateFeeResponse, error) {
	// 顺丰的运费计算需要调用专门的接口
	// 这里简化处理，实际应该调用顺丰的运费查询接口
	
	// 基础运费（首重1kg）
	baseFee := uint64(1500) // 15元
	
	// 续重费用
	var weightFee uint64
	if req.Weight > 1.0 {
		extraWeight := req.Weight - 1.0
		weightFee = uint64(extraWeight * 500) // 每kg 5元
	}
	
	// 保价费用（货值的1%）
	var insuranceFee uint64
	if req.DeclaredValue > 0 {
		insuranceFee = req.DeclaredValue / 100
	}
	
	totalFee := baseFee + weightFee + insuranceFee

	return &CalculateFeeResponse{
		BaseFee:      baseFee,
		WeightFee:    weightFee,
		InsuranceFee: insuranceFee,
		TotalFee:     totalFee,
	}, nil
}

// doRequest 发送HTTP请求
func (g *SFExpressGateway) doRequest(ctx context.Context, serviceCode string, bizData interface{}) ([]byte, error) {
	// 序列化业务数据
	bizDataJSON, err := json.Marshal(bizData)
	if err != nil {
		return nil, err
	}

	// 构建请求参数
	timestamp := time.Now().UnixMilli()
	msgData := base64.StdEncoding.EncodeToString(bizDataJSON)
	
	// 生成签名
	msgDigest := g.sign(msgData, timestamp)

	// 构建请求体
	params := url.Values{}
	params.Set("partnerID", g.config.CustomerID)
	params.Set("requestID", fmt.Sprintf("%d", timestamp))
	params.Set("serviceCode", serviceCode)
	params.Set("timestamp", fmt.Sprintf("%d", timestamp))
	params.Set("msgDigest", msgDigest)
	params.Set("msgData", msgData)

	// 发送请求
	req, err := http.NewRequestWithContext(ctx, "POST", g.config.GatewayURL, 
		strings.NewReader(params.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// sign 生成签名
func (g *SFExpressGateway) sign(msgData string, timestamp int64) string {
	// 拼接签名字符串
	signStr := fmt.Sprintf("%s%s%d", msgData, g.config.AppSecret, timestamp)
	
	// MD5加密
	hash := md5.Sum([]byte(signStr))
	return strings.ToUpper(hex.EncodeToString(hash[:]))
}

// parseStatus 解析物流状态
func (g *SFExpressGateway) parseStatus(routes []struct {
	AcceptTime    string `json:"acceptTime"`
	AcceptAddress string `json:"acceptAddress"`
	Remark        string `json:"remark"`
	OpCode        string `json:"opCode"`
}) TrackingStatus {
	if len(routes) == 0 {
		return TrackingStatusCollected
	}

	// 根据最新的opCode判断状态
	latestOpCode := routes[0].OpCode
	
	switch latestOpCode {
	case "50": // 已签收
		return TrackingStatusDelivered
	case "44": // 派送中
		return TrackingStatusDelivering
	case "31": // 到达目的地
		return TrackingStatusInTransit
	case "11": // 已收件
		return TrackingStatusCollected
	case "99": // 异常
		return TrackingStatusException
	default:
		return TrackingStatusInTransit
	}
}

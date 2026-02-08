package application

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	pb "github.com/wyfcoding/ecommerce/goapi/openapi/v1"
	"github.com/wyfcoding/ecommerce/internal/openapi/domain"
)

type OpenApiService struct {
	repo domain.OpenApiRepository
}

func NewOpenApiService(repo domain.OpenApiRepository) *OpenApiService {
	return &OpenApiService{repo: repo}
}

func (s *OpenApiService) CreateApp(ctx context.Context, req *pb.CreateAppRequest) (*pb.CreateAppResponse, error) {
	appID := fmt.Sprintf("app_%d", time.Now().UnixNano())
	apiKey := generateRandomString(16)
	apiSecret := generateRandomString(32)

	app := &domain.OpenApiApp{
		AppID:       appID,
		UserID:      req.UserId,
		AppName:     req.AppName,
		Description: req.Description,
		APIKey:      apiKey,
		APISecret:   apiSecret,
		Status:      domain.StatusActive,
		Scopes:      "product.read,order.read",
	}

	if err := s.repo.SaveApp(ctx, app); err != nil {
		return nil, err
	}

	return &pb.CreateAppResponse{
		AppId:     appID,
		ApiKey:    apiKey,
		ApiSecret: apiSecret,
	}, nil
}

func (s *OpenApiService) GetAppStatus(ctx context.Context, apiKey string) (*pb.GetAppStatusResponse, error) {
	app, err := s.repo.GetAppByKey(ctx, apiKey)
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, fmt.Errorf("app not found for key %s", apiKey)
	}

	return &pb.GetAppStatusResponse{
		AppId:  app.AppID,
		Status: string(app.Status),
		Scopes: strings.Split(app.Scopes, ","),
	}, nil
}

func (s *OpenApiService) InvokeApi(ctx context.Context, req *pb.InvokeApiRequest) (*pb.InvokeApiResponse, error) {
	app, err := s.repo.GetAppByKey(ctx, req.ApiKey)
	if err != nil || app == nil || app.Status != domain.StatusActive {
		return &pb.InvokeApiResponse{Code: 403, ErrorMsg: "invalid or disabled api key"}, nil
	}

	// 模拟代理内部 API 调用
	return &pb.InvokeApiResponse{
		Code: 200,
		Data: fmt.Sprintf("Successfully invoked %s %s via app %s", req.Method, req.Path, app.AppName),
	}, nil
}

func generateRandomString(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

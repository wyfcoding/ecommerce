package interfaces

import (
	"context"

	pb "github.com/wyfcoding/ecommerce/go-api/openapi/v1"
	"github.com/wyfcoding/ecommerce/internal/openapi/application"
)

type OpenApiHandler struct {
	pb.UnimplementedOpenApiServiceServer
	app *application.OpenApiService
}

func NewOpenApiHandler(app *application.OpenApiService) *OpenApiHandler {
	return &OpenApiHandler{app: app}
}

func (h *OpenApiHandler) CreateApp(ctx context.Context, req *pb.CreateAppRequest) (*pb.CreateAppResponse, error) {
	return h.app.CreateApp(ctx, req)
}

func (h *OpenApiHandler) GetAppStatus(ctx context.Context, req *pb.GetAppStatusRequest) (*pb.GetAppStatusResponse, error) {
	return h.app.GetAppStatus(ctx, req.ApiKey)
}

func (h *OpenApiHandler) InvokeApi(ctx context.Context, req *pb.InvokeApiRequest) (*pb.InvokeApiResponse, error) {
	return h.app.InvokeApi(ctx, req)
}

package grpc

import (
	"context"
	"encoding/json"
	pb "github.com/wyfcoding/ecommerce/api/search/v1"
	"github.com/wyfcoding/ecommerce/internal/search/application"
	"github.com/wyfcoding/ecommerce/internal/search/domain/entity"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	pb.UnimplementedSearchServiceServer
	app *application.SearchService
}

func NewServer(app *application.SearchService) *Server {
	return &Server{app: app}
}

func (s *Server) SearchProducts(ctx context.Context, req *pb.SearchProductsRequest) (*pb.SearchProductsResponse, error) {
	page := int(req.PageToken)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	filter := &entity.SearchFilter{
		Keyword:  req.Query,
		Page:     page,
		PageSize: pageSize,
	}

	// UserID is not in proto, using 0 (anonymous)
	result, err := s.app.Search(ctx, 0, filter)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbProducts := make([]*pb.Product, 0, len(result.Items))
	for _, item := range result.Items {
		// Item is interface{}, likely map[string]interface{} from Elasticsearch or struct
		// We need to marshal/unmarshal or type assert to map to proto Product
		// Assuming it can be marshaled to JSON and unmarshaled to proto Product for simplicity/robustness
		// or better, if we know the type.
		// Let's try to marshal/unmarshal as a generic way to handle interface{} -> struct mapping

		bytes, err := json.Marshal(item)
		if err != nil {
			continue
		}

		var p pb.Product
		if err := json.Unmarshal(bytes, &p); err != nil {
			continue
		}
		pbProducts = append(pbProducts, &p)
	}

	return &pb.SearchProductsResponse{
		Products:      pbProducts,
		TotalSize:     int32(result.Total),
		NextPageToken: int32(page + 1),
	}, nil
}

package grpc

import (
	"context"
	pb "ecommerce/api/data_ingestion/v1"
	"ecommerce/internal/data_ingestion/application"

	"google.golang.org/protobuf/types/known/emptypb"
)

type Server struct {
	pb.UnimplementedDataIngestionServer
	app *application.DataIngestionService
}

func NewServer(app *application.DataIngestionService) *Server {
	return &Server{app: app}
}

func (s *Server) IngestEvent(ctx context.Context, req *pb.IngestEventRequest) (*emptypb.Empty, error) {
	// Service currently supports Pull model (RegisterSource -> TriggerIngestion).
	// Proto defines Push model (IngestEvent).
	// For now, we just log the event or return success to satisfy interface.
	// In a real scenario, we might want to add a PushIngestion method to the service.

	// s.app.PushEvent(ctx, req.EventType, req.EventData, req.Timestamp.AsTime(), req.Source)

	return &emptypb.Empty{}, nil
}

func (s *Server) IngestBatch(ctx context.Context, req *pb.IngestBatchRequest) (*emptypb.Empty, error) {
	// Same as IngestEvent, but for batch.
	return &emptypb.Empty{}, nil
}

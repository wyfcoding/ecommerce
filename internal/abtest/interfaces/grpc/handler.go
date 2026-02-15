package grpc

import (
	"context"
	"log/slog"

	pb "github.com/wyfcoding/ecommerce/go-api/abtest/v1"
	"github.com/wyfcoding/ecommerce/internal/abtest/application"
	"github.com/wyfcoding/ecommerce/internal/abtest/domain"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	pb.UnimplementedABTestServiceServer
	svc    *application.ABTestService
	logger *slog.Logger
}

func NewServer(svc *application.ABTestService, logger *slog.Logger) pb.ABTestServiceServer {
	return &Server{
		svc:    svc,
		logger: logger.With("component", "abtest_grpc"),
	}
}

func (s *Server) CreateExperiment(ctx context.Context, req *pb.CreateExperimentRequest) (*pb.Experiment, error) {
	exp, err := s.svc.CreateExperiment(ctx, req)
	if err != nil {
		return nil, err
	}
	return toExperimentProto(exp), nil
}

func (s *Server) GetAssignment(ctx context.Context, req *pb.GetAssignmentRequest) (*pb.AssignmentResponse, error) {
	key, val, err := s.svc.GetAssignment(ctx, req.UserId, req.ExperimentName)
	if err != nil {
		return nil, err
	}
	return &pb.AssignmentResponse{
		VariationKey:   key,
		VariationValue: val,
	}, nil
}

func (s *Server) TrackEvent(ctx context.Context, req *pb.TrackEventRequest) (*emptypb.Empty, error) {
	if err := s.svc.TrackEvent(ctx, req.UserId, req.ExperimentName, req.EventName, req.Value); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) GetResults(ctx context.Context, req *pb.GetResultsRequest) (*pb.ExperimentResults, error) {
	results, err := s.svc.GetResults(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	protoResults := make([]*pb.VariationResult, len(results))
	for i, r := range results {
		protoResults[i] = &pb.VariationResult{
			VariationKey:   r.VariationKey,
			Participants:   r.Participants,
			Conversions:    r.Conversions,
			ConversionRate: r.ConversionRate,
		}
	}

	return &pb.ExperimentResults{
		Id:      req.Id,
		Results: protoResults,
	}, nil
}

func (s *Server) UpdateExperimentStatus(ctx context.Context, req *pb.UpdateExperimentStatusRequest) (*pb.Experiment, error) {
	exp, err := s.svc.UpdateStatus(ctx, req.Id, req.Status)
	if err != nil {
		return nil, err
	}
	return toExperimentProto(exp), nil
}

func toExperimentProto(exp *domain.Experiment) *pb.Experiment {
	variations := make([]*pb.Variation, len(exp.Variations))
	for i, v := range exp.Variations {
		variations[i] = &pb.Variation{
			Key:    v.Key,
			Value:  v.Value,
			Weight: v.Weight,
		}
	}

	p := &pb.Experiment{
		Id:                exp.ID,
		Name:              exp.Name,
		Description:       exp.Description,
		Variations:        variations,
		TrafficPercentage: exp.TrafficPercentage,
		Status:            exp.Status,
		CreatedAt:         timestamppb.New(exp.CreatedAt),
	}
	if exp.StartedAt != nil {
		p.StartedAt = timestamppb.New(*exp.StartedAt)
	}
	if exp.EndedAt != nil {
		p.EndedAt = timestamppb.New(*exp.EndedAt)
	}
	return p
}

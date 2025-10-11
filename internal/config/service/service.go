package service

import (
	"context"
	"errors"

	v1 "ecommerce/api/config/v1"
	"ecommerce/internal/config/biz"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ConfigService is the gRPC service implementation for configuration management.
type ConfigService struct {
	v1.UnimplementedConfigServiceServer
	uc *biz.ConfigUsecase
}

// NewConfigService creates a new ConfigService.
func NewConfigService(uc *biz.ConfigUsecase) *ConfigService {
	return &ConfigService{uc: uc}
}

// bizConfigEntryToProto converts biz.ConfigEntry to v1.ConfigEntry.
func bizConfigEntryToProto(entry *biz.ConfigEntry) *v1.ConfigEntry {
	if entry == nil {
		return nil
	}
	return &v1.ConfigEntry{
		Key:         entry.Key,
		Value:       entry.Value,
		Description: entry.Description,
		UpdatedAt:   timestamppb.New(entry.UpdatedAt),
	}
}

// GetConfig implements the GetConfig RPC.
func (s *ConfigService) GetConfig(ctx context.Context, req *v1.GetConfigRequest) (*v1.ConfigEntry, error) {
	if req.Key == "" {
		return nil, status.Error(codes.InvalidArgument, "key is required")
	}

	entry, err := s.uc.GetConfig(ctx, req.Key)
	if err != nil {
		if errors.Is(err, biz.ErrConfigNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Errorf(codes.Internal, "failed to get config: %v", err)
	}

	return bizConfigEntryToProto(entry), nil
}

// SetConfig implements the SetConfig RPC.
func (s *ConfigService) SetConfig(ctx context.Context, req *v1.SetConfigRequest) (*v1.ConfigEntry, error) {
	if req.Key == "" || req.Value == "" {
		return nil, status.Error(codes.InvalidArgument, "key and value are required")
	}

	entry, err := s.uc.SetConfig(ctx, req.Key, req.Value, req.Description)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to set config: %v", err)
	}

	return bizConfigEntryToProto(entry), nil
}

// ListConfigs implements the ListConfigs RPC.
func (s *ConfigService) ListConfigs(ctx context.Context, req *v1.ListConfigsRequest) (*v1.ListConfigsResponse, error) {
	entries, err := s.uc.ListConfigs(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list configs: %v", err)
	}

	protoEntries := make([]*v1.ConfigEntry, len(entries))
	for i, entry := range entries {
		protoEntries[i] = bizConfigEntryToProto(entry)
	}

	return &v1.ListConfigsResponse{Configs: protoEntries}, nil
}

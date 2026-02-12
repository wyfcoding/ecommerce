package interfaces

import (
	"context"

	pb "github.com/wyfcoding/ecommerce/go-api/i18n/v1"
	"github.com/wyfcoding/ecommerce/internal/i18n/application"
	"github.com/wyfcoding/ecommerce/internal/i18n/domain"
)

type I18nHandler struct {
	pb.UnimplementedI18NServiceServer
	app  *application.I18nService
	repo domain.I18nRepository
}

func NewI18nHandler(app *application.I18nService, repo domain.I18nRepository) *I18nHandler {
	return &I18nHandler{app: app, repo: repo}
}

func (h *I18nHandler) GetTranslation(ctx context.Context, req *pb.GetTranslationRequest) (*pb.GetTranslationResponse, error) {
	value, found, err := h.app.GetTranslation(ctx, req.LangCode, req.Key)
	if err != nil {
		return nil, err
	}
	return &pb.GetTranslationResponse{Value: value, Found: found}, nil
}

func (h *I18nHandler) ListTranslations(ctx context.Context, req *pb.ListTranslationsRequest) (*pb.ListTranslationsResponse, error) {
	translations, err := h.app.ListTranslations(ctx, req.LangCode, req.Keys, req.Namespace)
	if err != nil {
		return nil, err
	}
	return &pb.ListTranslationsResponse{Translations: translations}, nil
}

func (h *I18nHandler) PutTranslation(ctx context.Context, req *pb.PutTranslationRequest) (*pb.PutTranslationResponse, error) {
	err := h.app.PutTranslation(ctx, req.LangCode, req.Key, req.Value, req.Context, req.Namespace)
	if err != nil {
		return nil, err
	}
	return &pb.PutTranslationResponse{Success: true}, nil
}

func (h *I18nHandler) ListLanguages(ctx context.Context, req *pb.ListLanguagesRequest) (*pb.ListLanguagesResponse, error) {
	langs, err := h.app.ListLanguages(ctx)
	if err != nil {
		return nil, err
	}

	var pbLangs []*pb.Language
	for _, l := range langs {
		pbLangs = append(pbLangs, &pb.Language{
			Code:      l.Code,
			Name:      l.Name,
			Direction: l.Direction,
		})
	}
	return &pb.ListLanguagesResponse{Languages: pbLangs}, nil
}

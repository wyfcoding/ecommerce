package domain

import (
	"context"
)

type OpenApiRepository interface {
	SaveApp(ctx context.Context, app *OpenApiApp) error
	GetAppByID(ctx context.Context, appID string) (*OpenApiApp, error)
	GetAppByKey(ctx context.Context, apiKey string) (*OpenApiApp, error)
	ListAppsByUserID(ctx context.Context, userID string) ([]*OpenApiApp, error)
}

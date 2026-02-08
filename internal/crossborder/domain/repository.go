package domain

import (
	"context"
)

type CrossBorderRepository interface {
	SaveDeclaration(ctx context.Context, decl *CustomsDeclaration) error
	GetDeclaration(ctx context.Context, id string) (*CustomsDeclaration, error)

	GetHSCode(ctx context.Context, code string) (*HSCode, error)
	// For simulation/mocking
	SaveHSCode(ctx context.Context, hs *HSCode) error
}

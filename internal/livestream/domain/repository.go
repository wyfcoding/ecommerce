package domain

import (
	"context"
)

type LivestreamRepository interface {
	SaveRoom(ctx context.Context, room *Room) error
	GetRoom(ctx context.Context, roomID string) (*Room, error)
	ListRooms(ctx context.Context, status string, limit, offset int) ([]*Room, error)
	AddProduct(ctx context.Context, product *Product) error
}

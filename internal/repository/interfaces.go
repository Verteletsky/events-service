package repository

import (
	"context"

	"github.com/godev/events-service/internal/model"
)

type IEventRepository interface {
	Create(ctx context.Context, event *model.Event) error
	FindUnfinishedByType(ctx context.Context, eventType string) (*model.Event, error)
	Update(ctx context.Context, event *model.Event) error
	List(ctx context.Context, eventType string, offset, limit int64) ([]model.Event, error)
}

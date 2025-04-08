package service

import (
	"context"

	"github.com/godev/events-service/internal/model"
)

type IEventService interface {
	ListEvents(ctx context.Context, eventType string, offset, limit int64) ([]model.Event, error)
	StartEvent(ctx context.Context, eventType string) error
	FinishEvent(ctx context.Context, eventType string) error
}

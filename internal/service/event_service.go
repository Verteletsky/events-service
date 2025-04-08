package service

import (
	"context"
	"errors"
	"regexp"
	"time"

	"github.com/godev/events-service/internal/model"
	"github.com/godev/events-service/internal/repository"
)

var (
	ErrInvalidEventType = errors.New("тип события может содержать только строчные буквы и цифры")
	ErrInvalidLimit     = errors.New("limit не может быть больше 100")
	ErrEventNotFound    = errors.New("no unfinished event found")
	eventTypeRegex      = regexp.MustCompile("^[a-z0-9]+$")
)

type EventService struct {
	repo repository.IEventRepository
}

func NewEventService(repo repository.IEventRepository) IEventService {
	return &EventService{
		repo: repo,
	}
}

func (s *EventService) ListEvents(ctx context.Context, eventType string, offset, limit int64) ([]model.Event, error) {
	if limit > 100 {
		return nil, ErrInvalidLimit
	}

	if eventType != "" && !eventTypeRegex.MatchString(eventType) {
		return nil, ErrInvalidEventType
	}

	return s.repo.List(ctx, eventType, offset, limit)
}

func (s *EventService) StartEvent(ctx context.Context, eventType string) error {
	if !eventTypeRegex.MatchString(eventType) {
		return ErrInvalidEventType
	}

	existingEvent, err := s.repo.FindUnfinishedByType(ctx, eventType)
	if err != nil {
		return err
	}

	if existingEvent != nil {
		return nil
	}

	event := &model.Event{
		Type:      eventType,
		State:     model.EventStateStarted,
		StartedAt: time.Now(),
	}

	return s.repo.Create(ctx, event)
}

func (s *EventService) FinishEvent(ctx context.Context, eventType string) error {
	if !eventTypeRegex.MatchString(eventType) {
		return ErrInvalidEventType
	}

	event, err := s.repo.FindUnfinishedByType(ctx, eventType)
	if err != nil {
		return err
	}

	if event == nil {
		return ErrEventNotFound
	}

	now := time.Now()
	event.State = model.EventStateFinished
	event.FinishedAt = &now

	return s.repo.Update(ctx, event)
}

package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/godev/events-service/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MockEventRepository struct {
	mock.Mock
}

func (m *MockEventRepository) List(ctx context.Context, eventType string, offset, limit int64) ([]model.Event, error) {
	args := m.Called(ctx, eventType, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Event), args.Error(1)
}

func (m *MockEventRepository) Create(ctx context.Context, event *model.Event) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventRepository) Update(ctx context.Context, event *model.Event) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventRepository) FindById(ctx context.Context, id primitive.ObjectID) (*model.Event, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Event), args.Error(1)
}

func (m *MockEventRepository) FindUnfinishedByType(ctx context.Context, eventType string) (*model.Event, error) {
	args := m.Called(ctx, eventType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Event), args.Error(1)
}

func TestEventService_ListEvents(t *testing.T) {
	tests := []struct {
		name          string
		eventType     string
		offset        int64
		limit         int64
		mockEvents    []model.Event
		mockError     error
		expectedError error
	}{
		{
			name:      "successful list",
			eventType: "test123",
			offset:    0,
			limit:     10,
			mockEvents: []model.Event{
				{
					ID:        primitive.NewObjectID(),
					Type:      "test123",
					State:     model.EventStateFinished,
					StartedAt: time.Now().Add(-time.Hour),
					FinishedAt: func() *time.Time {
						t := time.Now()
						return &t
					}(),
				},
			},
			mockError:     nil,
			expectedError: nil,
		},
		{
			name:          "repository error",
			eventType:     "test123",
			offset:        0,
			limit:         10,
			mockEvents:    nil,
			mockError:     errors.New("repository error"),
			expectedError: errors.New("repository error"),
		},
		{
			name:          "invalid event type",
			eventType:     "Test-123",
			offset:        0,
			limit:         10,
			mockEvents:    nil,
			mockError:     nil,
			expectedError: ErrInvalidEventType,
		},
		{
			name:          "limit exceeds maximum",
			eventType:     "test123",
			offset:        0,
			limit:         101,
			mockEvents:    nil,
			mockError:     nil,
			expectedError: ErrInvalidLimit,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockEventRepository)
			service := NewEventService(mockRepo)

			if !errors.Is(tt.expectedError, ErrInvalidEventType) && !errors.Is(tt.expectedError, ErrInvalidLimit) {
				mockRepo.On("List", mock.Anything, tt.eventType, tt.offset, tt.limit).Return(tt.mockEvents, tt.mockError)
			}

			events, err := service.ListEvents(context.Background(), tt.eventType, tt.offset, tt.limit)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.mockEvents, events)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestEventService_StartEvent(t *testing.T) {
	tests := []struct {
		name      string
		eventType string
		mockEvent *model.Event
		mockErr   error
		expectErr error
	}{
		{
			name:      "successful start",
			eventType: "test123",
			mockEvent: nil,
			mockErr:   nil,
			expectErr: nil,
		},
		{
			name:      "repository error",
			eventType: "test123",
			mockEvent: nil,
			mockErr:   errors.New("repository error"),
			expectErr: errors.New("repository error"),
		},
		{
			name:      "invalid event type",
			eventType: "Test-123",
			mockEvent: nil,
			mockErr:   nil,
			expectErr: ErrInvalidEventType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockEventRepository)
			service := NewEventService(mockRepo)

			if !errors.Is(tt.expectErr, ErrInvalidEventType) {
				mockRepo.On("FindUnfinishedByType", mock.Anything, tt.eventType).Return(tt.mockEvent, tt.mockErr)
				if tt.expectErr == nil {
					mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
				}
			}

			err := service.StartEvent(context.Background(), tt.eventType)
			if tt.expectErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestEventService_FinishEvent(t *testing.T) {
	tests := []struct {
		name      string
		eventType string
		mockEvent *model.Event
		mockErr   error
		expectErr error
	}{
		{
			name:      "successful finish",
			eventType: "test123",
			mockEvent: &model.Event{
				ID:      primitive.NewObjectID(),
				Type:    "test123",
				State:   model.EventStateFinished,
				Version: 2,
			},
			mockErr:   nil,
			expectErr: nil,
		},
		{
			name:      "repository error",
			eventType: "test123",
			mockEvent: nil,
			mockErr:   errors.New("repository error"),
			expectErr: errors.New("repository error"),
		},
		{
			name:      "invalid event type",
			eventType: "Test-123",
			mockEvent: nil,
			mockErr:   nil,
			expectErr: ErrInvalidEventType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockEventRepository)
			service := NewEventService(mockRepo)

			if !errors.Is(tt.expectErr, ErrInvalidEventType) {
				mockRepo.On("FindUnfinishedByType", mock.Anything, tt.eventType).Return(tt.mockEvent, tt.mockErr)
				if tt.mockEvent != nil {
					mockRepo.On("Update", mock.Anything, tt.mockEvent).Return(nil)
				}
			}

			err := service.FinishEvent(context.Background(), tt.eventType)
			if tt.expectErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

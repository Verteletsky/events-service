package queue

import (
	"context"

	"github.com/godev/events-service/internal/service"
	"go.uber.org/zap"
)

type Worker struct {
	Id           int
	EventService service.IEventService
	Log          *zap.Logger
}

func NewWorker(id int, eventService service.IEventService, log *zap.Logger) *Worker {
	return &Worker{
		Id:           id,
		EventService: eventService,
		Log:          log,
	}
}

func (w *Worker) Process(ctx context.Context, task EventTask) {
	w.Log.Info("Worker processing task",
		zap.Int("worker_id", w.Id),
		zap.String("event_type", task.EventType),
		zap.String("action", task.Action),
	)

	var err error
	switch task.Action {
	case "start":
		err = w.EventService.StartEvent(ctx, task.EventType)
	case "finish":
		err = w.EventService.FinishEvent(ctx, task.EventType)
	}

	if err != nil {
		w.Log.Error("Failed to process event",
			zap.Error(err),
			zap.String("event_type", task.EventType),
			zap.String("action", task.Action),
		)
	}
}

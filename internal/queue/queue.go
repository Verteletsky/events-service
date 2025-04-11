package queue

import (
	"context"
	"fmt"
	"sync"

	"github.com/godev/events-service/internal/service"
	"go.uber.org/zap"
)

type EventQueue struct {
	tasks        chan EventTask
	workers      []*Worker
	log          *zap.Logger
	eventService service.IEventService
	wg           sync.WaitGroup
	bufferSize   int
}

func NewEventQueue(ctx context.Context, eventService service.IEventService, log *zap.Logger, numWorkers int, bufferSize int) *EventQueue {
	queue := &EventQueue{
		tasks:        make(chan EventTask, bufferSize),
		eventService: eventService,
		log:          log,
		bufferSize:   bufferSize,
	}

	for i := 0; i < numWorkers; i++ {
		worker := NewWorker(i, eventService, log)
		queue.workers = append(queue.workers, worker)
		queue.wg.Add(1)
		go queue.runWorker(ctx, worker)
	}

	return queue
}

func (p *EventQueue) runWorker(ctx context.Context, worker *Worker) {
	defer p.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case task, ok := <-p.tasks:
			if !ok {
				return
			}
			worker.Process(ctx, task)
		}
	}
}

func (p *EventQueue) ProcessEvent(ctx context.Context, eventType string, action string) error {
	task := EventTask{
		EventType: eventType,
		Action:    action,
	}

	select {
	case <-ctx.Done():
		return fmt.Errorf("context done")
	case p.tasks <- task:
		fmt.Println("Task added to channel")
		return nil
	default:
		return fmt.Errorf("channel is full")
	}
}

func (p *EventQueue) Close() {
	close(p.tasks)
	p.wg.Wait()
}

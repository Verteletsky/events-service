package mongo

import (
	"context"
	"errors"

	"github.com/godev/events-service/internal/repository"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/godev/events-service/internal/model"
)

type EventRepository struct {
	collection *mongo.Collection
}

func NewEventRepository(db *mongo.Database) repository.IEventRepository {
	collection := db.Collection("events")

	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "type", Value: 1},
				{Key: "state", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "started_at", Value: -1},
			},
		},
	}

	_, err := collection.Indexes().CreateMany(context.Background(), indexes)
	if err != nil {
		panic(err)
	}

	return &EventRepository{
		collection: collection,
	}
}

func (r *EventRepository) Create(ctx context.Context, event *model.Event) error {
	session, err := r.collection.Database().Client().StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		event.Version = 1
		_, err := r.collection.InsertOne(sessCtx, event)
		return nil, err
	})
	return err
}

func (r *EventRepository) FindUnfinishedByType(ctx context.Context, eventType string) (*model.Event, error) {
	var event model.Event
	err := r.collection.FindOne(ctx, bson.M{
		"type":  eventType,
		"state": model.EventStateStarted,
	}).Decode(&event)

	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}

	return &event, err
}

func (r *EventRepository) Update(ctx context.Context, event *model.Event) error {
	session, err := r.collection.Database().Client().StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		filter := bson.M{
			"_id":     event.ID,
			"version": event.Version,
		}

		update := bson.M{
			"$set": bson.M{
				"state":       event.State,
				"finished_at": event.FinishedAt,
				"version":     event.Version + 1,
			},
		}

		result, err := r.collection.UpdateOne(sessCtx, filter, update)
		if err != nil {
			return nil, err
		}

		if result.ModifiedCount == 0 {
			return nil, errors.New("event was modified by another process")
		}

		return nil, nil
	})
	return err
}

func (r *EventRepository) List(ctx context.Context, eventType string, offset, limit int64) ([]model.Event, error) {
	filter := bson.M{}
	if eventType != "" {
		filter["type"] = eventType
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "started_at", Value: -1}}).
		SetSkip(offset).
		SetLimit(limit)

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var events []model.Event
	if err = cursor.All(ctx, &events); err != nil {
		return nil, err
	}

	return events, nil
}

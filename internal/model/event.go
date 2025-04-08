package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EventState int

const (
	EventStateStarted EventState = iota
	EventStateFinished
)

type Event struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Type       string             `bson:"type" json:"type"`
	State      EventState         `bson:"state" json:"state"`
	StartedAt  time.Time          `bson:"started_at" json:"startedAt"`
	FinishedAt *time.Time         `bson:"finished_at,omitempty" json:"finishedAt,omitempty"`
	Version    int64              `bson:"version" json:"version"`
}

type EventRequest struct {
	Type string `json:"type" validate:"required,regexp=^[a-z0-9]+$"`
}

type EventsResponse struct {
	Events []Event `json:"events"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

type EventFilter struct {
	EventType string
	Offset    int64
	Limit     int64
}

package eventing

import (
	uuid "github.com/satori/go.uuid"
	"time"
)

type Event interface {
	GetEntityID() string
	GetEventID() string
	GetEventName() string
	GetCreatedAt() time.Time
}

type BasicEvent struct {
	EntityID  string    `json:"entity_id" bson:"entity_id"`
	EventID   string    `json:"event_id" bson:"event_id"`
	EventName string    `json:"event_type" bson:"event_type"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
}

func NewBasicEvent(entityID string, name string) BasicEvent {
	return BasicEvent{
		EntityID:  entityID,
		EventID:   uuid.NewV4().String(),
		EventName: name,
		CreatedAt: time.Now(),
	}
}

func (basicEvent BasicEvent) GetEntityID() string {
	return basicEvent.EntityID
}

func (basicEvent BasicEvent) GetEventID() string {
	return basicEvent.EventID
}

func (basicEvent BasicEvent) GetEventName() string {
	return basicEvent.EventName
}

func (basicEvent BasicEvent) GetCreatedAt() time.Time {
	return basicEvent.CreatedAt
}

package eventing

import (
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

type Projection interface {
	Apply(Event)
	GetCollectionName() string
	GetEntityID() string
	GetLastEventID() string
}

type BasicProjection struct {
	EntityID      string    `json:"entity_id" bson:"entity_id"`
	LastEventID   string    `json:"last_event_id" bson:"last_event_id"`
	LastEventTime time.Time `json:"last_event_time" bson:"last_event_time"`
}

func BasicProjectionGenerator(result *mongo.SingleResult) (Projection, error) {
	var proj BasicProjection
	if err := result.Decode(&proj); err != nil {
		return nil, err
	}
	return &proj, nil
}

func (basicProjection *BasicProjection) Apply(event Event) {
	basicProjection.EntityID = event.GetEntityID()
	basicProjection.LastEventID = event.GetEventID()
	basicProjection.LastEventTime = event.GetCreatedAt()
}

func (BasicProjection) GetCollectionName() string {
	return "basicProjection"
}

func (basicProjection *BasicProjection) GetEntityID() string {
	return basicProjection.EntityID
}

func (basicProjection *BasicProjection) GetLastEventID() string {
	return basicProjection.LastEventID
}

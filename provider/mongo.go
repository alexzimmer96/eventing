package provider

import (
	"context"
	"errors"
	"fmt"
	"github.com/alexzimmer96/eventing"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoEventGenerator func(cursor *mongo.Cursor) (eventing.Event, error)
type MongoProjectionGenerator func(result *mongo.SingleResult) (eventing.Projection, error)

type MongoStorageProvider struct {
	db                  *mongo.Database
	collectionName      string
	eventRegistry       map[string]MongoEventGenerator
	projectionGenerator MongoProjectionGenerator
}

func NewMongoStorageProvider(db *mongo.Database, collection string, generator MongoProjectionGenerator) *MongoStorageProvider {
	return &MongoStorageProvider{
		db:                  db,
		collectionName:      collection,
		eventRegistry:       map[string]MongoEventGenerator{},
		projectionGenerator: generator,
	}
}

func (provider *MongoStorageProvider) WithEvent(name string, generator MongoEventGenerator) *MongoStorageProvider {
	provider.eventRegistry[name] = generator
	return provider
}

func (provider *MongoStorageProvider) SaveEvent(ctx context.Context, event eventing.Event) error {
	_, err := provider.db.Collection(provider.collectionName).InsertOne(ctx, event)
	return err
}

func (provider *MongoStorageProvider) SaveProjection(ctx context.Context, projection eventing.Projection) error {
	_, err := provider.db.Collection(projection.GetCollectionName()).ReplaceOne(
		ctx,
		bson.M{"entity_id": projection.GetEntityID()},
		projection,
		options.Replace().SetUpsert(true),
	)
	return err
}

func (provider *MongoStorageProvider) GetProjection(ctx context.Context, entityID string, projection eventing.Projection) (eventing.Projection, error) {
	result := provider.db.Collection(projection.GetCollectionName()).FindOne(
		ctx,
		bson.D{{"entity_id", entityID}},
		options.FindOne().SetSort(bson.D{{"last_event_time", -1}}), // ordering by -1 means newest first
	)
	if result.Err() == mongo.ErrNoDocuments {
		return nil, nil
	}
	if result.Err() != nil {
		return nil, result.Err()
	}
	return provider.projectionGenerator(result)
}

func (provider *MongoStorageProvider) GetLatestEventIDForEntityID(ctx context.Context, entityID string) (string, error) {
	result := provider.db.Collection(provider.collectionName).FindOne(
		ctx,
		bson.D{{"entity_id", entityID}},
		options.FindOne().SetSort(bson.D{{"created_at", -1}}), // ordering by -1 means newest first
	)
	if result.Err() != nil {
		return "", result.Err()
	}
	m := make(map[string]interface{})
	if err := result.Decode(&m); err != nil {
		return "", err
	}
	eventID := fmt.Sprintf("%v", m["event_id"])
	if len(eventID) > 0 {
		return eventID, nil
	}
	return "", nil
}

func (provider *MongoStorageProvider) GetSortedEventsForEntityID(ctx context.Context, entityID string) ([]eventing.Event, error) {
	cursor, err := provider.db.Collection(provider.collectionName).Find(
		context.TODO(),
		bson.D{{"entity_id", entityID}},
		options.Find().SetSort(bson.D{{"created_at", 1}}), // ordering by 1 means oldest first
	)
	if err != nil {
		return nil, err
	}
	var events []eventing.Event
	for cursor.Next(ctx) {
		generator, err := provider.getEventFromRaw(cursor)
		if err != nil {
			return nil, errors.New("no fetchedEvent found for this aggregate")
		}
		events = append(events, generator)
	}
	return events, nil
}

func (provider *MongoStorageProvider) getEventFromRaw(raw *mongo.Cursor) (eventing.Event, error) {
	rawEventType, err := raw.Current.LookupErr("event_type")
	if err != nil {
		return nil, errors.New("could not process event")
	}
	eventType := rawEventType.StringValue()
	generator, ok := provider.eventRegistry[eventType]
	if ok == false {
		return nil, errors.New(fmt.Sprintf("event is not registered in mongo storage provider: %s", eventType))
	}
	event, err := generator(raw)
	// Error while generating event from Raw mongo entry
	if err != nil {
		return nil, err
	}
	return event, nil
}

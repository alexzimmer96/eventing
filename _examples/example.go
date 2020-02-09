package main

import (
	"context"
	"fmt"
	"github.com/alexzimmer96/eventing"
	"github.com/alexzimmer96/eventing/provider"
	uuid "github.com/satori/go.uuid"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
)

func main() {
	database := connectToMongo("localhost", "27017", "root", "root", "testEventing")
	storage := provider.NewMongoStorageProvider(database, "events", eventing.BasicProjectionGenerator).
		WithEvent("ExampleEvent", exampleEventFromCursor)

	projectionController := eventing.NewController(&ExampleProjection{}, buildExampleProjection, storage)
	exampleEvent := NewExampleEvent(ExampleEventData{SomeString: "test"})

	err := projectionController.SaveEvent(context.Background(), exampleEvent)
	if err != nil {
		log.Fatal("could not save event")
	}

	_, err = projectionController.GetLatestProjection(context.Background(), exampleEvent.EntityID)
	if err != nil {
		log.Fatal("could not save event")
	}
}

//======================================================================================================================

type ExampleEvent struct {
	eventing.BasicEvent `bson:",inline"`
	Data                ExampleEventData `json:"data" bson:"data"`
}

type ExampleEventData struct {
	SomeString string `json:"some_string" bson:"some_string"`
}

func NewExampleEvent(data ExampleEventData) *ExampleEvent {
	return &ExampleEvent{
		BasicEvent: eventing.NewBasicEvent(uuid.NewV4().String(), "ExampleEvent"),
		Data:       data,
	}
}

func exampleEventFromCursor(raw *mongo.Cursor) (eventing.Event, error) {
	e := &ExampleEvent{}
	err := raw.Decode(e)
	if err != nil {
		return nil, err
	}
	return e, nil
}

//======================================================================================================================

type ExampleProjection struct {
	eventing.BasicProjection `bson:",inline"`
	SomeString               string `json:"some_string" bson:"some_string"`
}

func buildExampleProjection(ctx context.Context, events []eventing.Event) (eventing.Projection, error) {
	ex := &ExampleProjection{}
	for _, event := range events {
		ex.Apply(event)
	}
	return ex, nil
}

func (ex *ExampleProjection) Apply(event eventing.Event) {
	ex.BasicProjection.Apply(event)
	switch v := event.(type) {
	case *ExampleEvent:
		ex.SomeString = v.Data.SomeString
	default:
		fmt.Println("could not handle event")
	}
}

//======================================================================================================================

func connectToMongo(host, port, user, password, dbName string) *mongo.Database {
	opts := options.Client().ApplyURI(fmt.Sprintf("mongodb://%s:%s", host, port))
	opts.SetAuth(options.Credential{
		Username: user,
		Password: password,
	})

	client, err := mongo.Connect(context.Background(), opts)
	if err != nil {
		log.Fatal("could not connect to database")
	}

	if err := client.Ping(context.Background(), nil); err != nil {
		log.Fatal("could not ping database")
	}

	return client.Database(dbName)
}

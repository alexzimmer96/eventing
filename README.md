[![MIT license](http://img.shields.io/badge/license-MIT-brightgreen.svg)](http://opensource.org/licenses/MIT) 
[![codecov](https://codecov.io/gh/alexzimmer96/eventing/branch/master/graph/badge.svg)](https://codecov.io/gh/alexzimmer96/eventing)
[![Go Report Card](https://goreportcard.com/badge/github.com/alexzimmer96/eventing)](https://goreportcard.com/report/github.com/alexzimmer96/eventing)
[![](https://godoc.org/github.com/alexzimmer96/eventing?status.svg)](http://godoc.org/github.com/alexzimmer96/eventing)

# Eventing
This package make working with events easy.
Its provides functions for store events and derive projections (your entities) from the list of events.

> This package is under heavy development. The API may encounter breaking changes.
> You are welcome to contribute.

## Wording
|Term|Description|
|---|---|
|Entity|An entity is an object which is defined through its identity and not through its attributes. You can read more about in any Domain Driven Design book. A Entity always have a unique id.|
|Event|Describes an relevant event for a referenced entity.|
|Projection|The current state of an Entity, based on all belonging events.|
|Controller|Provides a interface for managing a Projection and its Events. Every Event that is registered and saved through the Controller automatically updates the Projection.|
|StorageProvider|Provides an interface for storing Events and Projections to a Datastore (e.g. database).|

## Basic Functionality
You can find a combined usage-example in the [example directory](_examples)
### Event
A Event needs to implement the `Event`-interface.
This is typically done by embedding the `BasicEvent` struct and providing a constructor function:
```go
type ExampleEvent struct {
	eventing.BasicEvent `bson:",inline"` // Using inline here is recommended
	MyData       string `json:"my_data" bson:"my_data"` // Providing custom json and bson-keys for MongoDB
}

func NewExampleEvent(data string) *ExampleEvent {
	return &ExampleEvent{
		BasicEvent: eventing.NewBasicEvent(uuid.NewV4().String(), "ExampleEvent"),
		MyData:     data,
	}
}
```

### Projection
A Projection needs to implement the `Projection`-interface.
This is typically done by embedding the `BasicProjection` struct and providing some functions:
* Implement / override the Apply function to apply specific events. Its highly recommended to call the Apply function of
    the embedded `BasicProjection` first, to apply basic event data.
* A (private) function to build the concrete Projection by a list of Events.
    This is necessary as long as go does not support of generics.
```go
type ExampleProjection struct {
	eventing.BasicProjection `bson:",inline"`
	SomeString               string `json:"some_string" bson:"some_string"`
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

func buildExampleProjection(ctx context.Context, events []eventing.Event) (eventing.Projection, error) {
	ex := &ExampleProjection{}
	for _, event := range events {
		ex.Apply(event)
	}
	return ex, nil
}
```

### Controller
You should create one Controller per Projection. A Controller can be initialized like this:
```go
func main() {
    projectionController := eventing.NewController(&ExampleProjection{}, buildExampleProjection, SOME_STORAGE_PROVIDER)
}
```

## Storage Providers
To store Events and Projections you will need a StorageProvider, e.g. A database. 
This project holds a example provider for MongoDB which can be easily be used:
```go
func main() {
    database := connectToMongo("localhost", "27017", "root", "root", "testEventing")
    storage := provider.NewMongoStorageProvider(database, "events", eventing.BasicProjectionGenerator)
}
```

### Registering Events
You may need to provide a function to build a concrete Event by the raw Database output.
The provided MongoDB-implementation depends on. 
Make sure to register all your events.
This can be done by using the fluent API:
```go
func main() {
    storage := provider.NewMongoStorageProvider(database, "events", eventing.BasicProjectionGenerator).
            WithEvent("ExampleEvent", exampleEventFromCursor)
}

func exampleEventFromCursor(raw *mongo.Cursor) (eventing.Event, error) {
	e := &ExampleEvent{}
	err := raw.Decode(e)
	if err != nil {
		return nil, err
	}
	return e, nil
}
```

package eventing

import (
	"context"
)

type BuildProjectionFunc func(ctx context.Context, events []Event) (Projection, error)

type StorageProvider interface {
	SaveEvent(ctx context.Context, event Event) error
	SaveProjection(ctx context.Context, projection Projection) error
	GetProjection(ctx context.Context, entityID string, projection Projection) (Projection, error)
	GetLatestEventIDForEntityID(ctx context.Context, entityID string) (string, error)
	GetSortedEventsForEntityID(ctx context.Context, entityID string) ([]Event, error)
}

// Controllers are the primary way to save events and build projections based upon these.
type Controller struct {
	provider        StorageProvider
	projection      Projection
	buildProjection BuildProjectionFunc
}

// NewController creates a new instance of an controller by collection all necessary data.
func NewController(projection Projection, buildProjection BuildProjectionFunc, provider StorageProvider) *Controller {
	return &Controller{
		provider:        provider,
		projection:      projection,
		buildProjection: buildProjection,
	}
}

// GetLatestProjection retrieves the latest projection the managed projection-entity.
func (controller *Controller) GetLatestProjection(ctx context.Context, entityID string) (Projection, error) {
	latestProjection, err := controller.provider.GetProjection(ctx, entityID, controller.projection)
	if err != nil {
		return nil, err
	}
	return latestProjection, nil
}

// SaveEvent stores an event and builds an up-to-date projection.
func (controller *Controller) SaveEvent(ctx context.Context, event Event) error {
	if err := controller.provider.SaveEvent(ctx, event); err != nil {
		return err
	}
	events, err := controller.provider.GetSortedEventsForEntityID(ctx, event.GetEntityID())
	if err != nil {
		return err
	}
	proj, err := controller.buildProjection(ctx, events)
	return controller.provider.SaveProjection(ctx, proj)
}

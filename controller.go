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

type Controller struct {
	provider        StorageProvider
	projection      Projection
	buildProjection BuildProjectionFunc
}

func NewController(projection Projection, buildProjection BuildProjectionFunc, provider StorageProvider) *Controller {
	return &Controller{
		provider:        provider,
		projection:      projection,
		buildProjection: buildProjection,
	}
}

func (controller *Controller) GetLatestProjection(ctx context.Context, entityID string) (Projection, error) {
	latestProjection, err := controller.provider.GetProjection(ctx, entityID, controller.projection)
	if err != nil {
		return nil, err
	}
	if latestProjection == nil {
		return nil, nil
	}
	latestEventID, err := controller.provider.GetLatestEventIDForEntityID(ctx, entityID)
	if err != nil {
		return nil, err
	}
	if latestProjection.GetLastEventID() != latestEventID {
		events, err := controller.provider.GetSortedEventsForEntityID(ctx, entityID)
		if err != nil {
			return nil, err
		}
		newProjection, err := controller.buildProjection(ctx, events)
		err = controller.provider.SaveProjection(ctx, newProjection)
		if err != nil {
			return nil, err
		}
		return newProjection, nil
	}
	return latestProjection, nil
}

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

package services

import (
	"context"
	"math/rand"
	"pvz-cli/internal/common/apperrors"
	"pvz-cli/internal/models"
)

var _ ActorService = (*DefaultActorService)(nil)

type DefaultActorService struct{}

func NewDefaultActorService() *DefaultActorService {
	return &DefaultActorService{}
}

// DetermineActor determines the actor (courier or client) responsible for handling an event based on its type and user ID.
func (s *DefaultActorService) DetermineActor(ctx context.Context, event models.EventType, userID uint64) (models.Actor, error) {
	if ctx.Err() != nil {
		return models.Actor{}, ctx.Err()
	}
	switch event {
	case models.EventAccepted, models.EventReturnedToWarehouse:
		courierID, err := s.FindFreeCourier(ctx)
		if err != nil {
			return models.Actor{}, err
		}
		return models.Actor{
			Type: models.ActorCourier,
			ID:   courierID,
		}, nil
	case models.EventIssued, models.EventReturnedByClient:
		return models.Actor{
			Type: models.ActorClient,
			ID:   userID,
		}, nil
	default:
		return models.Actor{}, apperrors.Newf(apperrors.ValidationFailed, "unknown event type: %s", event)
	}
}

// FindFreeCourier attempts to identify an available courier and returns their ID, or an error if the context is canceled.
func (s *DefaultActorService) FindFreeCourier(ctx context.Context) (uint64, error) {
	if ctx.Err() != nil {
		return 0, ctx.Err()
	}
	courierID := rand.Uint64() // hw7 condition, placeholder
	if courierID == 0 {
		courierID = 1
	}
	return courierID, nil
}

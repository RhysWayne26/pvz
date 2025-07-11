package services

import (
	"context"
	"pvz-cli/internal/models"
)

type ActorService interface {
	DetermineActor(ctx context.Context, event models.EventType, userID uint64) (models.Actor, error)
	FindFreeCourier(ctx context.Context) (uint64, error)
}

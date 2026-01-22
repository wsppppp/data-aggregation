package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/wsppppp/data-aggregation/internal/domain"
)

type SubscriptionFilter struct {
	UserID      *uuid.UUID
	ServiceName *string
}

type Subscriptions interface {
	Create(ctx context.Context, sub *domain.Subscription) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Subscription, error)
	Update(ctx context.Context, sub *domain.Subscription) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter SubscriptionFilter, limit, offset int) ([]domain.Subscription, error)
	FindActiveInPeriod(ctx context.Context, filter SubscriptionFilter, from, to time.Time) ([]domain.Subscription, error)
}

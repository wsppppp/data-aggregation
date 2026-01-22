package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/wsppppp/data-aggregation/internal/domain"
	"github.com/wsppppp/data-aggregation/internal/repository"
)

type SubscriptionRepository struct {
	pool *pgxpool.Pool
}

func NewSubscriptionRepository(pool *pgxpool.Pool) *SubscriptionRepository {
	return &SubscriptionRepository{pool: pool}
}

func (r *SubscriptionRepository) Create(ctx context.Context, sub *domain.Subscription) error {
	query := `
		INSERT INTO subscriptions (id, user_id, service_name, price, start_date, end_date)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.pool.Exec(ctx, query,
		sub.ID,
		sub.UserID,
		sub.ServiceName,
		sub.Price,
		sub.StartDate,
		sub.EndDate,
	)
	if err != nil {
		return fmt.Errorf("failed to create subscription: %w", err)
	}
	return nil
}

func (r *SubscriptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Subscription, error) {
	query := `
		SELECT * FROM subscriptions
		WHERE id = $1
	`
	row := r.pool.QueryRow(ctx, query, id)

	var sub domain.Subscription
	var endDate *time.Time
	if err := row.Scan(
		&sub.ID,
		&sub.UserID,
		&sub.ServiceName,
		&sub.Price,
		&sub.StartDate,
		&endDate,
	); err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}
	sub.EndDate = endDate
	return &sub, nil
}

func (r *SubscriptionRepository) Update(ctx context.Context, sub *domain.Subscription) error {
	query := `
		UPDATE subscriptions
		SET user_id = $2, service_name = $3, price = $4, start_date = $5, end_date = $6
		WHERE id = $1
	`
	ct, err := r.pool.Exec(ctx, query,
		sub.ID,
		sub.UserID,
		sub.ServiceName,
		sub.Price,
		sub.StartDate,
		sub.EndDate,
	)
	if err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("subscription not found")
	}
	return nil
}

func (r *SubscriptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM subscriptions WHERE id = $1`
	ct, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete subscription: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("subscription not found")
	}
	return nil
}

func (r *SubscriptionRepository) List(ctx context.Context, filter repository.SubscriptionFilter, limit, offset int) ([]domain.Subscription, error) {
	query := `
		SELECT * FROM subscriptions
		WHERE 1=1
		  AND ($1::uuid IS NULL OR user_id = $1::uuid)
		  AND ($2::text IS NULL OR service_name = $2::text)
		ORDER BY start_date DESC, service_name ASC
		LIMIT $3 OFFSET $4
	`
	var userID *uuid.UUID
	if filter.UserID != nil {
		userID = filter.UserID
	}
	var serviceName *string
	if filter.ServiceName != nil {
		serviceName = filter.ServiceName
	}

	rows, err := r.pool.Query(ctx, query, userID, serviceName, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list subscriptions: %w", err)
	}
	defer rows.Close()

	var result []domain.Subscription
	for rows.Next() {
		var s domain.Subscription
		var endDate *time.Time
		if err := rows.Scan(
			&s.ID, &s.UserID, &s.ServiceName, &s.Price, &s.StartDate, &endDate,
		); err != nil {
			return nil, fmt.Errorf("failed to scan subscription: %w", err)
		}
		s.EndDate = endDate
		result = append(result, s)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("rows error: %w", rows.Err())
	}
	return result, nil
}

func (r *SubscriptionRepository) FindActiveInPeriod(ctx context.Context, filter repository.SubscriptionFilter, from, to time.Time) ([]domain.Subscription, error) {
	query := `
		SELECT * FROM subscriptions
		WHERE start_date <= $2
		  AND (end_date IS NULL OR end_date >= $1)
		  AND ($3::uuid IS NULL OR user_id = $3::uuid)
		  AND ($4::text IS NULL OR service_name = $4::text)
		ORDER BY start_date ASC, service_name ASC
	`

	var userID *uuid.UUID
	if filter.UserID != nil {
		userID = filter.UserID
	}
	var serviceName *string
	if filter.ServiceName != nil {
		serviceName = filter.ServiceName
	}

	rows, err := r.pool.Query(ctx, query, from, to, userID, serviceName)
	if err != nil {
		return nil, fmt.Errorf("failed to find active subscriptions: %w", err)
	}
	defer rows.Close()

	var result []domain.Subscription
	for rows.Next() {
		var s domain.Subscription
		var endDate *time.Time
		if err := rows.Scan(
			&s.ID, &s.UserID, &s.ServiceName, &s.Price, &s.StartDate, &endDate,
		); err != nil {
			return nil, fmt.Errorf("failed to scan subscription: %w", err)
		}
		s.EndDate = endDate
		result = append(result, s)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("rows error: %w", rows.Err())
	}
	return result, nil
}

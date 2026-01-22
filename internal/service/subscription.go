package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/wsppppp/data-aggregation/internal/domain"
	"github.com/wsppppp/data-aggregation/internal/repository"
)

type SubscriptionService struct {
	repo repository.Subscriptions
}

func NewSubscriptionService(repo repository.Subscriptions) *SubscriptionService {
	return &SubscriptionService{repo: repo}
}

type CreateSubscriptionInput struct {
	ServiceName string
	Price       int
	UserID      uuid.UUID
	StartDate   time.Time
	EndDate     *time.Time // Для будущих операций обновления/создания с end_date (опционально)
}

func (s *SubscriptionService) Create(ctx context.Context, input CreateSubscriptionInput) (uuid.UUID, error) {
	id := uuid.New()
	sub := &domain.Subscription{
		ID:          id,
		UserID:      input.UserID,
		ServiceName: input.ServiceName,
		Price:       input.Price,
		StartDate:   input.StartDate,
		EndDate:     input.EndDate,
	}
	if err := s.repo.Create(ctx, sub); err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

func (s *SubscriptionService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Subscription, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *SubscriptionService) Update(ctx context.Context, sub *domain.Subscription) error {
	return s.repo.Update(ctx, sub)
}

func (s *SubscriptionService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *SubscriptionService) List(ctx context.Context, filter repository.SubscriptionFilter, limit, offset int) ([]domain.Subscription, error) {
	return s.repo.List(ctx, filter, limit, offset)
}

// _________________функции для рассчета итоговой суммы __________________

// приводит дату к первому числу месяца
func normalizeMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC) // день первый
}

func monthsBetweenInclusive(a, b time.Time) int {
	// Если правая граница раньше левой - пересечения нет
	if b.Before(a) {
		return 0
	}
	ay, am := a.Year(), int(a.Month())
	by, bm := b.Year(), int(b.Month())
	return (by-ay)*12 + (bm - am) + 1
}

func maxDate(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

func minDate(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}

// _________________ итоговая сумма _________________

func (s *SubscriptionService) TotalCost(ctx context.Context, filter repository.SubscriptionFilter, from, to time.Time) (int, error) {
	from = normalizeMonth(from)
	to = normalizeMonth(to)

	subs, err := s.repo.FindActiveInPeriod(ctx, filter, from, to)
	if err != nil {
		return 0, err
	}

	total := 0
	for _, sub := range subs {
		left := maxDate(sub.StartDate, from)

		// Если end_date = NULL - до конца
		rightCandidate := to
		if sub.EndDate != nil {
			rightCandidate = *sub.EndDate
		}
		right := minDate(rightCandidate, to)

		months := monthsBetweenInclusive(left, right)
		if months > 0 {
			total += sub.Price * months
		}
	}

	return total, nil
}

package rest

import (
	"time"

	"github.com/wsppppp/data-aggregation/internal/domain"
)

func toMonthYear(t time.Time) string {
	return t.Format(MonthYearLayout)
}

func toMonthYearPtr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format(MonthYearLayout)
	return &s
}

func toSubscriptionResponse(s *domain.Subscription) SubscriptionResponse {
	return SubscriptionResponse{
		ID:          s.ID.String(),
		UserID:      s.UserID.String(),
		ServiceName: s.ServiceName,
		Price:       s.Price,
		StartDate:   toMonthYear(s.StartDate),
		EndDate:     toMonthYearPtr(s.EndDate),
	}
}

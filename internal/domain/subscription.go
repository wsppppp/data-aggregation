package domain

import (
	"time"

	"github.com/google/uuid"
)

type Subscription struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	UserID      uuid.UUID  `json:"user_id" db:"user_id"`
	ServiceName string     `json:"service_name" db:"service_name"`
	Price       int        `json:"price" db:"price"`
	StartDate   time.Time  `json:"start_date" db:"start_date"`       // в тз было непонятно, поэтому сделаю 1 число указанного месяца
	EndDate     *time.Time `json:"end_date,omitempty" db:"end_date"` // указатель, тк конец это опционально и может быть null
}

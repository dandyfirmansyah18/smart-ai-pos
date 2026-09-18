package domain

import (
	"time"

	"github.com/google/uuid"
)

type CashShift struct {
	ID              uuid.UUID  `json:"id"`
	UserID          uuid.UUID  `json:"user_id"`
	Status          string     `json:"status"` // OPEN, CLOSED
	OpeningCash     float64    `json:"opening_cash"`
	ClosingCash     float64    `json:"closing_cash"`
	ExpectedCash    float64    `json:"expected_cash"`
	TotalCashSales  float64    `json:"total_cash_sales"`
	TotalQrisSales  float64    `json:"total_qris_sales"`
	TotalDebitSales float64    `json:"total_debit_sales"`
	Notes           string     `json:"notes"`
	OpenedAt        time.Time  `json:"opened_at"`
	ClosedAt        *time.Time `json:"closed_at"`
}

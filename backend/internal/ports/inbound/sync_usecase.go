package inbound

import (
	"context"
)

type SyncStatusResponse struct {
	IsOnline        bool   `json:"is_online"`
	PendingOrders   int    `json:"pending_orders"`
	PendingShifts   int    `json:"pending_shifts"`
	PendingPayments int    `json:"pending_payments"`
	LastSyncedAt    string `json:"last_synced_at,omitempty"`
}

type SyncUseCase interface {
	GetSyncStatus(ctx context.Context) (*SyncStatusResponse, error)
	TriggerSync(ctx context.Context, postgresDSN string) error
}

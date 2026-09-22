package outbound

import (
	"context"
)

type SyncCounts struct {
	PendingOrders   int
	PendingShifts   int
	PendingPayments int
}

type SyncRepository interface {
	GetPendingCounts(ctx context.Context) (*SyncCounts, error)
	SyncMasterDataFromPostgres(ctx context.Context, dsn string) error
	PushPendingDataToPostgres(ctx context.Context, dsn string) error
	CheckPostgresHealth(ctx context.Context, dsn string) bool
}

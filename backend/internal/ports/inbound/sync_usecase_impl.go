package inbound

import (
	"context"
	"fmt"
	"time"

	"github.com/pos-backend/config"
	"github.com/pos-backend/internal/ports/outbound"
)

type SyncUseCaseImpl struct {
	syncRepo outbound.SyncRepository
	cfg      *config.Config
}

func NewSyncUseCaseImpl(syncRepo outbound.SyncRepository, cfg *config.Config) *SyncUseCaseImpl {
	return &SyncUseCaseImpl{
		syncRepo: syncRepo,
		cfg:      cfg,
	}
}

func (uc *SyncUseCaseImpl) GetSyncStatus(ctx context.Context) (*SyncStatusResponse, error) {
	if uc.syncRepo == nil {
		return &SyncStatusResponse{IsOnline: true, PendingOrders: 0}, nil
	}

	counts, err := uc.syncRepo.GetPendingCounts(ctx)
	if err != nil || counts == nil {
		counts = &outbound.SyncCounts{}
	}

	dsn := ""
	if uc.cfg != nil {
		dsn = uc.cfg.DSN()
	}

	isOnline := uc.syncRepo.CheckPostgresHealth(ctx, dsn)

	return &SyncStatusResponse{
		IsOnline:        isOnline,
		PendingOrders:   counts.PendingOrders,
		PendingShifts:   counts.PendingShifts,
		PendingPayments: counts.PendingPayments,
		LastSyncedAt:    time.Now().Format(time.RFC3339),
	}, nil
}

func (uc *SyncUseCaseImpl) TriggerSync(ctx context.Context, postgresDSN string) error {
	if uc.syncRepo == nil {
		return fmt.Errorf("sync repository not initialized")
	}

	dsn := postgresDSN
	if dsn == "" && uc.cfg != nil {
		dsn = uc.cfg.DSN()
	}

	if !uc.syncRepo.CheckPostgresHealth(ctx, dsn) {
		return fmt.Errorf("central postgres unreachable")
	}

	// 1. Pull Master Data from Postgres to local SQLite
	if err := uc.syncRepo.SyncMasterDataFromPostgres(ctx, dsn); err != nil {
		return fmt.Errorf("failed to sync master data: %w", err)
	}

	// 2. Push Pending Data from local SQLite to Postgres
	if err := uc.syncRepo.PushPendingDataToPostgres(ctx, dsn); err != nil {
		return fmt.Errorf("failed to push pending data: %w", err)
	}

	return nil
}

var _ SyncUseCase = (*SyncUseCaseImpl)(nil)

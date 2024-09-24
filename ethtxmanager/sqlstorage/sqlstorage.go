package sqlstorage

import (
	"context"
	"database/sql"

	"github.com/0xPolygon/zkevm-ethtx-manager/types"
	"github.com/ethereum/go-ethereum/common"
	migrate "github.com/rubenv/sql-migrate"
)

const driverName = "sqlite3"

var _ types.StorageInterface = (*SqlStorage)(nil)

//nolint:revive
type SqlStorage struct {
	db *sql.DB
}

//nolint:revive
func NewSqlStorage(dbPath string) (*SqlStorage, error) {
	db, err := sql.Open(driverName, dbPath)
	if err != nil {
		return nil, err
	}

	if err := RunMigrations(driverName, db, migrate.Up); err != nil {
		return nil, err
	}

	return &SqlStorage{db: db}, nil
}

//nolint:revive
func (s *SqlStorage) Add(ctx context.Context, mTx types.MonitoredTx) error {
	panic("not implemented") // TODO: Implement
}

//nolint:revive
func (s *SqlStorage) Remove(ctx context.Context, id common.Hash) error {
	panic("not implemented") // TODO: Implement
}

//nolint:revive
func (s *SqlStorage) Get(ctx context.Context, id common.Hash) (types.MonitoredTx, error) {
	panic("not implemented") // TODO: Implement
}

//nolint:revive
func (s *SqlStorage) GetByStatus(ctx context.Context, statuses []types.MonitoredTxStatus) ([]types.MonitoredTx, error) {
	panic("not implemented") // TODO: Implement
}

//nolint:revive
func (s *SqlStorage) GetByBlock(ctx context.Context, fromBlock *uint64, toBlock *uint64) ([]types.MonitoredTx, error) {
	panic("not implemented") // TODO: Implement
}

//nolint:revive
func (s *SqlStorage) Update(ctx context.Context, mTx types.MonitoredTx) error {
	panic("not implemented") // TODO: Implement
}

//nolint:revive
func (s *SqlStorage) Empty(ctx context.Context) error {
	panic("not implemented") // TODO: Implement
}

package sqlstorage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

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

// NewSqlStorage creates and returns a new instance of SqlStorage with the given database path.
// It first opens a connection to the SQLite database and then runs the necessary migrations.
// If any error occurs during the database connection or migration process, it returns an error.
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

// Add inserts a new monitored transaction (MonitoredTx) into the database.
// It first sets the creation and updated timestamps, prepares the SQL insert query, and handles conflicts.
// If a transaction with the same ID already exists, it returns an error.
func (s *SqlStorage) Add(ctx context.Context, mTx types.MonitoredTx) error {
	// Set the creation timestamp
	mTx.CreatedAt = time.Now()
	mTx.UpdatedAt = mTx.CreatedAt

	// SQL Query for inserting a new monitored transaction
	query := `
		INSERT INTO tx_manager.monitored_txs (
			id, from_address, to_address, nonce, value, tx_data, gas, gas_offset, gas_price, blob_sidecar, blob_gas, 
			blob_gas_price, gas_tip_cap, status, block_number, history, created_at, updated_at, estimate_gas
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, 
			$12, $13, $14, $15, $16, $17, $18, $19
		) 
		ON CONFLICT (id) DO NOTHING;`

	// Prepare history (JSONB) field conversion
	historyJSON, err := json.Marshal(mTx.History)
	if err != nil {
		return err
	}

	// Execute the insert query with the provided transaction values
	result, err := s.db.ExecContext(ctx, query,
		mTx.ID,
		mTx.From.String(), // From Address (string)
		sql.NullString{ // To Address (nullable)
			String: mTx.To.String(),
			Valid:  mTx.To != nil,
		},
		mTx.Nonce,                 // Nonce
		mTx.Value.String(),        // Value (NUMERIC field)
		mTx.Data,                  // Transaction data
		mTx.Gas,                   // Gas
		mTx.GasOffset,             // Gas Offset
		mTx.GasPrice.String(),     // Gas Price (NUMERIC)
		mTx.BlobSidecar,           // Blob Sidecar (BYTEA)
		mTx.BlobGas,               // Blob Gas
		mTx.BlobGasPrice.String(), // Blob Gas Price (NUMERIC)
		mTx.GasTipCap.String(),    // Gas Tip Cap (NUMERIC)
		mTx.Status,                // Status (int)
		sql.NullString{ // Block Number (nullable big.Int)
			String: mTx.BlockNumber.String(),
			Valid:  mTx.BlockNumber != nil,
		},
		historyJSON,     // History (JSONB)
		mTx.CreatedAt,   // Created At (TIMESTAMPTZ)
		mTx.UpdatedAt,   // Updated At (TIMESTAMPTZ)
		mTx.EstimateGas, // Estimate Gas (boolean)
	)
	if err != nil {
		return err
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("transaction with ID %s already exists", mTx.ID)
	}

	return nil
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

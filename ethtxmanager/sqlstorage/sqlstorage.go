package sqlstorage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"time"

	localCommon "github.com/0xPolygon/zkevm-ethtx-manager/common"
	"github.com/0xPolygon/zkevm-ethtx-manager/types"
	"github.com/ethereum/go-ethereum/common"
	sqlite "github.com/mattn/go-sqlite3"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/russross/meddler"
)

const (
	// Insert denotes insert statement
	Insert sqlAction = iota
	// Update denotes update statement
	Update
	// Delete denotes delete statement
	Delete
)

type sqlAction int

func (s sqlAction) String() string {
	switch s {
	case Insert:
		return "INSERT"

	case Update:
		return "UPDATE"

	case Delete:
		return "DELETE"
	}

	return "UNKNOWN"
}

const (
	// baseSelectQuery represents the base select query, that retrieves all the values from the monitored_txs table
	baseSelectQuery = `SELECT id, from_address, to_address, nonce, value, tx_data, gas, gas_offset, gas_price, 
							blob_sidecar, blob_gas, blob_gas_price, gas_tip_cap, status, 
							block_number, history, created_at, updated_at, estimate_gas
						FROM monitored_txs`

	// baseDeleteStatement represents the base delete statement that deletes all the records from the monitored_txs table
	baseDeleteStatement = "DELETE FROM monitored_txs"

	// monitoredTxsTable is table name for persisting MonitoredTx objects
	monitoredTxsTable = "monitored_txs"
)

var (
	errNoRowsInResultSet = errors.New("sql: no rows in result set")
)

var _ types.StorageInterface = (*SqlStorage)(nil)

// SqlStorage encapsulates logic for MonitoredTx CRUD operations.
type SqlStorage struct {
	db *sql.DB
}

// NewStorage creates and returns a new instance of SqlStorage with the given database path.
// It first opens a connection to the SQLite database and then runs the necessary migrations.
// If any error occurs during the database connection or migration process, it returns an error.
func NewStorage(driverName, dbPath string) (*SqlStorage, error) {
	db, err := sql.Open(driverName, dbPath)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
		pragma journal_mode = WAL;
		PRAGMA foreign_keys = ON;
		pragma synchronous = normal;
		pragma journal_size_limit  = 6144000;
	`)
	if err != nil {
		return nil, err
	}

	if err := RunMigrations(driverName, db, migrate.Up); err != nil {
		return nil, err
	}

	initMeddler()

	return &SqlStorage{db: db}, nil
}

// Add persist a monitored transaction into the SQL database.
func (s *SqlStorage) Add(_ context.Context, mTx types.MonitoredTx) error {
	mTx.CreatedAt = time.Now()
	mTx.UpdatedAt = mTx.CreatedAt

	err := meddler.Insert(s.db, monitoredTxsTable, &mTx)
	if err != nil {
		sqlErr, success := UnwrapSQLiteErr(err)
		if !success {
			return err
		}

		if sqlErr.Code == sqlite.ErrConstraint {
			return fmt.Errorf("transaction with ID %s already exists", mTx.ID)
		}
	}

	return err
}

// Remove deletes a monitored transaction from the database by its ID.
// If the transaction does not exist, it returns an ErrNotFound error.
func (s *SqlStorage) Remove(ctx context.Context, id common.Hash) error {
	query := baseDeleteStatement + " WHERE id = $1"

	result, err := s.db.ExecContext(ctx, query, id.Hex())
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	// If no rows were affected, it means that the transaction was not found.
	if rowsAffected == 0 {
		return types.ErrNotFound
	}

	return nil
}

// Get retrieves a monitored transaction from the database by its ID.
// If the transaction is not found, it returns an ErrNotFound error.
func (s *SqlStorage) Get(_ context.Context, id common.Hash) (types.MonitoredTx, error) {
	query := baseSelectQuery + " WHERE id = $1"

	// Execute the query to retrieve the transaction data.
	var mTx types.MonitoredTx
	err := meddler.QueryRow(s.db, &mTx, query, id.Hex())
	if err != nil {
		if err.Error() == errNoRowsInResultSet.Error() {
			return types.MonitoredTx{}, types.ErrNotFound
		}

		return types.MonitoredTx{}, err
	}

	return mTx, nil
}

// GetByStatus retrieves monitored transactions from the database that match the provided statuses.
// If no statuses are provided, it returns all transactions.
// The transactions are ordered by their creation date (oldest first).
func (s *SqlStorage) GetByStatus(_ context.Context, statuses []types.MonitoredTxStatus) ([]types.MonitoredTx, error) {
	query := baseSelectQuery
	args := make([]interface{}, 0, len(statuses))

	if len(statuses) > 0 {
		// Build the WHERE clause for status filtering
		query += " WHERE status IN ("
		for i, status := range statuses {
			query += fmt.Sprintf("$%d", i+1)
			if i != len(statuses)-1 {
				query += ", "
			}
			args = append(args, string(status))
		}
		query += ")"
	}

	// Add ordering by creation date (oldest first)
	query += " ORDER BY created_at ASC"

	// Use meddler.QueryAll to retrieve the monitored transactions
	var transactions []*types.MonitoredTx
	if err := meddler.QueryAll(s.db, &transactions, query, args...); err != nil {
		return nil, fmt.Errorf("failed to query monitored transactions by status: %w", err)
	}

	return localCommon.SlicePtrsToSlice(transactions), nil
}

// GetByBlock loads all monitored transactions that have the blockNumber between fromBlock and toBlock.
func (s *SqlStorage) GetByBlock(ctx context.Context, fromBlock, toBlock *uint64) ([]types.MonitoredTx, error) {
	query := baseSelectQuery
	const maxArgs = 2

	args := make([]interface{}, 0, maxArgs)
	argsCounter := 1
	if fromBlock != nil {
		query += fmt.Sprintf(" WHERE block_number >= $%d", argsCounter)
		args = append(args, *fromBlock)
		argsCounter++
	}
	if toBlock != nil {
		if argsCounter > 1 {
			query += fmt.Sprintf(" AND block_number <= $%d", argsCounter)
		} else {
			query += fmt.Sprintf(" WHERE block_number <= $%d", argsCounter)
		}

		args = append(args, *toBlock)
	}

	// Use meddler.QueryAll to execute the query and scan into the result slice.
	var monitoredTxs []*types.MonitoredTx
	err := meddler.QueryAll(s.db, &monitoredTxs, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query monitored transactions by block: %w", err)
	}

	return localCommon.SlicePtrsToSlice(monitoredTxs), nil
}

// Update a persisted monitored tx
func (s *SqlStorage) Update(ctx context.Context, mTx types.MonitoredTx) error {
	mTx.UpdatedAt = time.Now()

	query := `
		UPDATE monitored_txs
		SET from_address = $1,
		    to_address = $2,
		    nonce = $3,
		    "value" = $4,
		    tx_data = $5,
		    gas = $6,
		    gas_offset = $7,
		    gas_price = $8,
		    blob_sidecar = $9,
		    blob_gas = $10,
		    blob_gas_price = $11,
		    gas_tip_cap = $12,
		    "status" = $13,
		    block_number = $14,
		    history = $15,
		    updated_at = $16,
		    estimate_gas = $17
		WHERE id = $18
	`

	args, err := prepareArgs(mTx, Update)
	if err != nil {
		return err
	}

	// Execute the query with the arguments
	result, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update monitored transaction: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return types.ErrNotFound
	}

	return nil
}

// Empty clears all the records from the monitored_txs table.
func (s *SqlStorage) Empty(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, baseDeleteStatement)
	if err != nil {
		return fmt.Errorf("failed to empty monitored_txs table: %w", err)
	}

	return nil
}

// UnwrapSQLiteErr attempts to extract a *sqlite.Error from the given error.
// It first checks if the error is directly of type *sqlite.Error, and if not,
// it tries to unwrap it from a meddler.DriverErr.
//
// Params:
//   - err: The error to check.
//
// Returns:
//   - *sqlite.Error: The extracted SQLite error, or nil if not found.
//   - bool: True if the error was successfully unwrapped as a *sqlite.Error.
func UnwrapSQLiteErr(err error) (*sqlite.Error, bool) {
	sqliteErr := &sqlite.Error{}
	if ok := errors.As(err, sqliteErr); ok {
		return sqliteErr, true
	}

	if driverErr, ok := meddler.DriverErr(err); ok {
		return sqliteErr, errors.As(driverErr, sqliteErr)
	}

	return sqliteErr, false
}

// prepareArgs prepares the arguments for the SQL query.
func prepareArgs(mTx types.MonitoredTx, action sqlAction) ([]interface{}, error) {
	toAddress := prepareNullableString(mTx.To)
	blockNumber := prepareNullableString(mTx.BlockNumber)
	value := prepareNullableString(mTx.Value)
	gasPrice := prepareNullableString(mTx.GasPrice)
	blobGasPrice := prepareNullableString(mTx.BlobGasPrice)
	gasTipCap := prepareNullableString(mTx.GasTipCap)

	historyJSON, blobSidecar, err := encodeHistoryAndBlobSidecar(mTx)
	if err != nil {
		return nil, err
	}

	args := []interface{}{
		mTx.From.Hex(),
		toAddress,
		mTx.Nonce,
		value,
		mTx.Data,
		mTx.Gas,
		mTx.GasOffset,
		gasPrice,
		blobSidecar,
		mTx.BlobGas,
		blobGasPrice,
		gasTipCap,
		mTx.Status,
		blockNumber,
		historyJSON,
		mTx.UpdatedAt.Format(time.RFC3339),
		localCommon.BoolToInteger(mTx.EstimateGas),
	}

	switch action {
	case Insert:
		args = append([]interface{}{mTx.ID.Hex()}, args...)
		args = append(args, mTx.CreatedAt.Format(time.RFC3339))

	case Update:
		args = append(args, mTx.ID.Hex())

	default:
		return nil, fmt.Errorf("unsupported SQL action provided %s", action)
	}

	return args, nil
}

// prepareNullableString prepares a sql.NullString from a nullable fields.
func prepareNullableString(value interface{}) sql.NullString {
	switch v := value.(type) {
	case *common.Address:
		if v != nil {
			return sql.NullString{Valid: true, String: v.Hex()}
		}

	case *big.Int:
		if v != nil {
			return sql.NullString{Valid: true, String: v.String()}
		}
	}

	return sql.NullString{Valid: false}
}

// encodeHistoryAndBlobSidecar marshals the history and blob sidecar into a JSON.
func encodeHistoryAndBlobSidecar(mTx types.MonitoredTx) ([]byte, []byte, error) {
	historyJSON, err := json.Marshal(mTx.History)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal transaction history to JSON: %w", err)
	}

	blobSidecarJSON, err := json.Marshal(mTx.BlobSidecar)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal blob sidecar to JSON: %w", err)
	}

	return historyJSON, blobSidecarJSON, nil
}

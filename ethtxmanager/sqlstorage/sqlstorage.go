package sqlstorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	localCommon "github.com/0xPolygon/zkevm-ethtx-manager/common"
	"github.com/0xPolygon/zkevm-ethtx-manager/types"
	"github.com/ethereum/go-ethereum/common"
	sqlite "github.com/mattn/go-sqlite3"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/russross/meddler"
)

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
			return types.ErrAlreadyExists
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

	columns, err := meddler.Columns(&mTx, false)
	if err != nil {
		return fmt.Errorf("failed to build the update statement (column names resolution failed): %w", err)
	}

	// Use strings.Builder instead of fmt.Sprintf for safer query building
	var builder strings.Builder
	builder.WriteString("UPDATE " + monitoredTxsTable + " SET ")

	// Build the SET clause
	// Skip the first column (primary key)
	for i, column := range columns[1:] {
		if i > 0 {
			builder.WriteString(", ")
		}
		builder.WriteString(column + " = $" + strconv.Itoa(i+1))
	}

	// Add the WHERE clause for the primary key
	builder.WriteString(" WHERE id = $")
	builder.WriteString(strconv.Itoa(len(columns)))

	query := builder.String()

	args, err := meddler.Values(&mTx, false)
	if err != nil {
		return err
	}

	if len(args) == 0 {
		return fmt.Errorf("failed to update monitored transaction %s, as there are no arguments", mTx.ID.Hex())
	}

	// append the primary key at the end
	args = append(args[1:], mTx.ID.Hex())

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

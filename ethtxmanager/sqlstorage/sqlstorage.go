package sqlstorage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	localCommon "github.com/0xPolygon/zkevm-ethtx-manager/common"
	"github.com/0xPolygon/zkevm-ethtx-manager/types"
	"github.com/ethereum/go-ethereum/common"
	migrate "github.com/rubenv/sql-migrate"
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
						FROM tx_manager.monitored_txs`

	// baseDeleteStatement represents the base delete statement that deletes all the records from the monitored_txs table
	baseDeleteStatement = "DELETE FROM tx_manager.monitored_txs"
)

var _ types.StorageInterface = (*SqlStorage)(nil)

// SqlStorage encapsulates logic for MonitoredTx CRUD operations.
type SqlStorage struct {
	db *sql.DB
}

// NewSqlStorage creates and returns a new instance of SqlStorage with the given database path.
// It first opens a connection to the SQLite database and then runs the necessary migrations.
// If any error occurs during the database connection or migration process, it returns an error.
func NewSqlStorage(driverName, dbPath string) (*SqlStorage, error) {
	db, err := sql.Open(driverName, dbPath)
	if err != nil {
		return nil, err
	}

	if err := RunMigrations(driverName, db, migrate.Up); err != nil {
		return nil, err
	}

	return &SqlStorage{db: db}, nil
}

// Add persist a monitored transaction into the SQL database.
func (s *SqlStorage) Add(ctx context.Context, mTx types.MonitoredTx) error {
	mTx.CreatedAt = time.Now()
	mTx.UpdatedAt = mTx.CreatedAt

	query := `
		INSERT INTO tx_manager.monitored_txs (
			id, from_address, to_address, nonce, "value", tx_data, gas, gas_offset, gas_price, blob_sidecar, 
			blob_gas, blob_gas_price, gas_tip_cap, "status", block_number, history, updated_at, estimate_gas, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, 
			$11, $12, $13, $14, $15, $16, $17, $18, $19
		)
		ON CONFLICT (id) DO NOTHING;`

	args, err := prepareArgs(mTx, Insert)
	if err != nil {
		return err
	}

	// Execute the insert query with the provided transaction values.
	result, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to insert monitored transaction: %w", err)
	}

	// Check if any rows were affected (transaction inserted).
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("transaction with ID %s already exists", mTx.ID)
	}

	return nil
}

// Remove deletes a monitored transaction from the database by its ID.
// If the transaction does not exist, it returns an ErrNotFound error.
func (s *SqlStorage) Remove(ctx context.Context, id common.Hash) error {
	query := fmt.Sprintf("%s WHERE id = $1", baseDeleteStatement)

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
func (s *SqlStorage) Get(ctx context.Context, id common.Hash) (types.MonitoredTx, error) {
	query := fmt.Sprintf("%s WHERE id = $1", baseSelectQuery)

	// Execute the query to retrieve the transaction data.
	rows, err := s.db.QueryContext(ctx, query, id.Hex())
	if err != nil {
		return types.MonitoredTx{}, err
	}
	defer rows.Close()

	if rows.Next() {
		return scanMonitoredTxRow(rows)
	}

	return types.MonitoredTx{}, types.ErrNotFound
}

// GetByStatus retrieves monitored transactions from the database that match the provided statuses.
// If no statuses are provided, it returns all transactions.
// The transactions are ordered by their creation date (oldest first).
func (s *SqlStorage) GetByStatus(ctx context.Context, statuses []types.MonitoredTxStatus) ([]types.MonitoredTx, error) {
	query := baseSelectQuery

	statusArgs := make([]interface{}, 0, len(statuses))
	if len(statuses) > 0 {
		query += " WHERE status IN ("
		for i, status := range statuses {
			query += fmt.Sprintf("$%d", i+1)
			if i != len(statuses)-1 {
				query += ", "
			}
			statusArgs = append(statusArgs, string(status))
		}
		query += ")"
	}

	// Add ordering by creation date (oldest first)
	query += " ORDER BY created_at ASC"

	// Execute the query and handle the result.
	rows, err := s.db.QueryContext(ctx, query, statusArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanMonitoredTxRows(rows)
}

// GetByBlock loads all monitored transactions that have the blockNumber between fromBlock and toBlock.
func (s *SqlStorage) GetByBlock(ctx context.Context, fromBlock, toBlock *uint64) ([]types.MonitoredTx, error) {
	query := baseSelectQuery

	args := []interface{}{}
	if fromBlock != nil {
		query += " AND block_number >= ?"
		args = append(args, *fromBlock)
	}
	if toBlock != nil {
		query += " AND block_number <= ?"
		args = append(args, *toBlock)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	return scanMonitoredTxRows(rows)
}

// Update a persisted monitored tx
func (s *SqlStorage) Update(ctx context.Context, mTx types.MonitoredTx) error {
	mTx.UpdatedAt = time.Now()

	query := `
		UPDATE tx_manager.monitored_txs
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
	_, err = s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update monitored transaction: %w", err)
	}

	return nil
}

// Empty clears all the records from the monitored_txs table.
func (s *SqlStorage) Empty(ctx context.Context) error {
	query := "DELETE FROM tx_manager.monitored_txs"

	_, err := s.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to empty monitored_txs table: %w", err)
	}

	return nil
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
		args = append([]interface{}{mTx.ID}, args...)
		args = append(args, mTx.CreatedAt.Format(time.RFC3339))

	case Update:
		args = append(args, mTx.ID.Hex())

	default:
		return nil, fmt.Errorf("unsupported SQL action provided %s", action)
	}

	return args, nil
}

// scanMonitoredTxRow is a helper function to scan a row into a MonitoredTx object
func scanMonitoredTxRow(rows *sql.Rows) (types.MonitoredTx, error) {
	var (
		mTx             types.MonitoredTx
		toAddress       sql.NullString // Nullable field for `to_address`.
		blockNumber     sql.NullString // Nullable field for `block_number`.
		value           sql.NullString // Nullable big.Int fields.
		gasPrice        sql.NullString // Nullable big.Int fields.
		blobGasPrice    sql.NullString // Nullable big.Int fields.
		gasTipCap       sql.NullString // Nullable big.Int fields.
		blobSidecarJSON []byte
		historyJSON     []byte
		status          string
	)

	err := rows.Scan(
		&mTx.ID,
		&mTx.From,
		&toAddress,
		&mTx.Nonce,
		&value,
		&mTx.Data,
		&mTx.Gas,
		&mTx.GasOffset,
		&gasPrice,
		&blobSidecarJSON,
		&mTx.BlobGas,
		&blobGasPrice,
		&gasTipCap,
		&status,
		&blockNumber,
		&historyJSON,
		&mTx.CreatedAt,
		&mTx.UpdatedAt,
		&mTx.EstimateGas,
	)
	if err != nil {
		return types.MonitoredTx{}, err
	}

	// Set the MonitoredTxStatus from the retrieved string
	mTx.Status = types.MonitoredTxStatus(status)

	// Populate nullable fields
	mTx.PopulateNullableStrings(toAddress, blockNumber, value, gasPrice, blobGasPrice, gasTipCap)

	// Unmarshal the BlobSidecar JSON
	if err := json.Unmarshal(blobSidecarJSON, &mTx.BlobSidecar); err != nil {
		return types.MonitoredTx{}, err
	}

	// Unmarshal the history JSON back into the map
	if err := json.Unmarshal(historyJSON, &mTx.History); err != nil {
		return types.MonitoredTx{}, err
	}

	return mTx, nil
}

// scanMonitoredTxRows is a helper function to scan multiple rows
func scanMonitoredTxRows(rows *sql.Rows) ([]types.MonitoredTx, error) {
	var mTxs []types.MonitoredTx
	for rows.Next() {
		mTx, err := scanMonitoredTxRow(rows)
		if err != nil {
			return nil, err
		}
		mTxs = append(mTxs, mTx)
	}

	// Check for any errors during iteration
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return mTxs, nil
}

// encodeHistoryAndBlobSidecar marshals the history and blob sidecar into a JSON.
func encodeHistoryAndBlobSidecar(mTx types.MonitoredTx) ([]byte, []byte, error) {
	historyJSON, err := json.Marshal(mTx.History)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal transaction history: %w", err)
	}

	blobSidecar, err := mTx.EncodeBlobSidecarToJSON()
	if err != nil {
		return nil, nil, err
	}

	return historyJSON, blobSidecar, nil
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

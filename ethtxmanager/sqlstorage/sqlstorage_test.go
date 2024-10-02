package sqlstorage

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"testing"
	"time"

	localCommon "github.com/0xPolygon/zkevm-ethtx-manager/common"
	"github.com/0xPolygon/zkevm-ethtx-manager/types"
	"github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto/kzg4844"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
)

// Test for Add method
func TestSqlStorage_Add(t *testing.T) {
	ctx := context.Background()

	storage, err := NewStorage(localCommon.SQLLiteDriverName, ":memory:")
	require.NoError(t, err)
	defer storage.db.Close()

	// Define test cases
	tests := []struct {
		name        string
		mTx         types.MonitoredTx
		expectedErr error
	}{
		{
			name:        "Add new transaction",
			mTx:         newMonitoredTx("0x1", "0xSender1", "0xReceiver1", 1, types.MonitoredTxStatusCreated, 100),
			expectedErr: nil,
		},
		{
			name:        "Add new transaction (contract deployment)",
			mTx:         newMonitoredTx("0x2", "0xSender1", "", 1, types.MonitoredTxStatusCreated, 100),
			expectedErr: nil,
		},
		{
			name:        "Add duplicate transaction",
			mTx:         newMonitoredTx("0x1", "0xSender1", "0xReceiver1", 1, types.MonitoredTxStatusCreated, 100),
			expectedErr: types.ErrAlreadyExists,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := storage.Add(ctx, test.mTx)
			if test.expectedErr != nil {
				require.ErrorIs(t, err, test.expectedErr)
			} else {
				require.NoError(t, err)

				// Assert that the given monitored transaction is indeed persisted
				resultTx, err := storage.Get(ctx, test.mTx.ID)
				require.NoError(t, err)
				compareTxsWithoutDates(t, test.mTx, resultTx)
			}
		})
	}
}

// Test for Remove method
func TestSqlStorage_Remove(t *testing.T) {
	ctx := context.Background()

	storage, err := NewStorage(localCommon.SQLLiteDriverName, ":memory:")
	require.NoError(t, err)
	defer storage.db.Close()

	// Add a transaction to remove
	tx := newMonitoredTx("0x1", "0xSender1", "0xReceiver1", 1, types.MonitoredTxStatusCreated, 100)
	require.NoError(t, storage.Add(ctx, tx))

	tests := []struct {
		name        string
		id          common.Hash
		expectedErr error
	}{
		{
			name:        "Remove existing transaction",
			id:          tx.ID,
			expectedErr: nil,
		},
		{
			name:        "Remove non-existing transaction",
			id:          common.HexToHash("0x2"), // ID that does not exist
			expectedErr: types.ErrNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := storage.Remove(ctx, test.id)
			if test.expectedErr != nil {
				require.ErrorIs(t, err, test.expectedErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestSqlStorage_Get(t *testing.T) {
	ctx := context.Background()

	storage, err := NewStorage(localCommon.SQLLiteDriverName, ":memory:")
	require.NoError(t, err)
	defer storage.db.Close()

	// Add a transaction to retrieve
	tx := newMonitoredTx("0x1", "0xSender1", "0xReceiver1", 1, types.MonitoredTxStatusCreated, 100)
	require.NoError(t, storage.Add(ctx, tx))

	tests := []struct {
		name        string
		id          common.Hash
		expectedTx  types.MonitoredTx
		expectedErr error
	}{
		{
			name:        "Get existing transaction",
			id:          tx.ID,
			expectedTx:  tx,
			expectedErr: nil,
		},
		{
			name:        "Get non-existing transaction",
			id:          common.HexToHash("0x2"),
			expectedTx:  types.MonitoredTx{},
			expectedErr: types.ErrNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := storage.Get(ctx, test.id)
			if test.expectedErr != nil {
				require.ErrorIs(t, err, test.expectedErr)
			} else {
				require.NoError(t, err)
				compareTxsWithoutDates(t, test.expectedTx, result)
			}
		})
	}
}

func TestSqlStorage_GetByStatus(t *testing.T) {
	ctx := context.Background()

	storage, err := NewStorage(localCommon.SQLLiteDriverName, ":memory:")
	require.NoError(t, err)
	defer storage.db.Close()

	// Add some transactions with different statuses
	tx1 := newMonitoredTx("0x1", "0xSender1", "0xReceiver1", 1, types.MonitoredTxStatusCreated, 100)
	tx2 := newMonitoredTx("0x2", "0xSender2", "0xReceiver2", 2, types.MonitoredTxStatusMined, 101)
	tx3 := newMonitoredTx("0x3", "0xSender3", "0xReceiver3", 3, types.MonitoredTxStatusCreated, 102)
	for _, tx := range []types.MonitoredTx{tx1, tx2, tx3} {
		require.NoError(t, storage.Add(ctx, tx))
	}

	tests := []struct {
		name        string
		statuses    []types.MonitoredTxStatus
		expectedIDs []common.Hash
	}{
		{
			name:        "Get by status - Created",
			statuses:    []types.MonitoredTxStatus{types.MonitoredTxStatusCreated},
			expectedIDs: []common.Hash{tx1.ID, tx3.ID},
		},
		{
			name:        "Get by status - Mined",
			statuses:    []types.MonitoredTxStatus{types.MonitoredTxStatusMined},
			expectedIDs: []common.Hash{tx2.ID},
		},
		{
			name:        "Get by status - All",
			statuses:    nil, // No statuses provided, should return all transactions
			expectedIDs: []common.Hash{tx1.ID, tx2.ID, tx3.ID},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := storage.GetByStatus(ctx, test.statuses)
			require.NoError(t, err)

			var resultIDs []common.Hash
			for _, tx := range result {
				resultIDs = append(resultIDs, tx.ID)
			}

			require.ElementsMatch(t, test.expectedIDs, resultIDs)
		})
	}
}

func TestSqlStorage_GetByBlock(t *testing.T) {
	ctx := context.Background()

	storage, err := NewStorage(localCommon.SQLLiteDriverName, ":memory:")
	require.NoError(t, err)
	defer storage.db.Close()

	// Add some transactions with different block numbers
	tx1 := newMonitoredTx("0x1", "0xSender1", "0xReceiver1", 1, types.MonitoredTxStatusCreated, 100)
	tx2 := newMonitoredTx("0x2", "0xSender2", "0xReceiver2", 2, types.MonitoredTxStatusCreated, 101)
	tx3 := newMonitoredTx("0x3", "0xSender3", "0xReceiver3", 3, types.MonitoredTxStatusCreated, 102)
	for _, tx := range []types.MonitoredTx{tx1, tx2, tx3} {
		require.NoError(t, storage.Add(ctx, tx))
	}

	tests := []struct {
		name        string
		fromBlock   *uint64
		toBlock     *uint64
		expectedIDs []common.Hash
	}{
		{
			name:        "Get by block range 100-101",
			fromBlock:   localCommon.ToUint64Ptr(100),
			toBlock:     localCommon.ToUint64Ptr(101),
			expectedIDs: []common.Hash{tx1.ID, tx2.ID},
		},
		{
			name:        "Get by block range 102",
			fromBlock:   localCommon.ToUint64Ptr(102),
			toBlock:     localCommon.ToUint64Ptr(102),
			expectedIDs: []common.Hash{tx3.ID},
		},
		{
			name:        "Get by no block range",
			fromBlock:   nil,
			toBlock:     nil,
			expectedIDs: []common.Hash{tx1.ID, tx2.ID, tx3.ID},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := storage.GetByBlock(ctx, test.fromBlock, test.toBlock)
			require.NoError(t, err)

			// Extract IDs from the result
			var resultIDs []common.Hash
			for _, tx := range result {
				resultIDs = append(resultIDs, tx.ID)
			}

			require.ElementsMatch(t, test.expectedIDs, resultIDs)
		})
	}
}

func TestSqlStorage_Update(t *testing.T) {
	ctx := context.Background()

	// Setup a temporary SQLite database for testing
	storage, err := NewStorage(localCommon.SQLLiteDriverName, ":memory:")
	require.NoError(t, err)
	defer storage.db.Close()

	// Add an initial transaction to update
	tx := newMonitoredTx("0x1", "0x2", "0x3", 1, types.MonitoredTxStatusCreated, 100)
	require.NoError(t, storage.Add(ctx, tx))

	// Define test cases
	tests := []struct {
		name        string
		updateTx    types.MonitoredTx
		expectedErr error
	}{
		{
			name: "Update existing transaction",
			updateTx: types.MonitoredTx{
				ID:          tx.ID,
				From:        common.HexToAddress("0x123456789"),
				To:          localCommon.ToAddressPtr("0x123456789"),
				Nonce:       1,
				Value:       big.NewInt(200),
				Data:        []byte{0x4, 0x5, 0x6},
				Gas:         30000,
				GasPrice:    big.NewInt(6000000000),
				Status:      types.MonitoredTxStatusMined,
				BlockNumber: big.NewInt(200),
				CreatedAt:   time.Now(),
			},
			expectedErr: nil,
		},
		{
			name: "Update non-existing transaction",
			updateTx: types.MonitoredTx{
				ID:          common.HexToHash("0x2"),
				From:        common.HexToAddress("0xUpdatedSender2"),
				To:          localCommon.ToAddressPtr("0xabcdef987654"),
				Nonce:       1,
				Value:       big.NewInt(200),
				Data:        []byte{0x4, 0x5, 0x6},
				Gas:         30000,
				GasPrice:    big.NewInt(6000000000),
				Status:      types.MonitoredTxStatusCreated,
				BlockNumber: big.NewInt(200),
			},
			expectedErr: types.ErrNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := storage.Update(ctx, test.updateTx)
			if test.expectedErr != nil {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				// Verify that the transaction was updated correctly
				updatedTx, err := storage.Get(ctx, test.updateTx.ID)
				require.NoError(t, err)
				compareTxsWithoutDates(t, test.updateTx, updatedTx)
			}
		})
	}
}

func TestSqlStorage_Empty(t *testing.T) {
	ctx := context.Background()

	// Setup a temporary SQLite database for testing
	storage, err := NewStorage(localCommon.SQLLiteDriverName, ":memory:")
	require.NoError(t, err)
	defer storage.db.Close()

	// Add some transactions
	tx1 := newMonitoredTx("0x1", "0xSender1", "0xReceiver1", 1, types.MonitoredTxStatusCreated, 100)
	tx2 := newMonitoredTx("0x2", "0xSender2", "0xReceiver2", 2, types.MonitoredTxStatusMined, 101)
	require.NoError(t, storage.Add(ctx, tx1))
	require.NoError(t, storage.Add(ctx, tx2))

	// Ensure transactions were added
	_, err = storage.Get(ctx, tx1.ID)
	require.NoError(t, err)
	_, err = storage.Get(ctx, tx2.ID)
	require.NoError(t, err)

	// Call Empty to remove all transactions
	err = storage.Empty(ctx)
	require.NoError(t, err)

	// Ensure that the transactions are gone
	_, err = storage.Get(ctx, tx1.ID)
	require.ErrorIs(t, err, types.ErrNotFound)
	_, err = storage.Get(ctx, tx2.ID)
	require.ErrorIs(t, err, types.ErrNotFound)
}

func TestSingleReaderMultipleWriters(t *testing.T) {
	storage, err := NewStorage(localCommon.SQLLiteDriverName, ":memory:")
	require.NoError(t, err)
	defer storage.db.Close()

	ctx := context.Background()

	numWriterGoroutines := 5
	numRecordsPerGoroutine := 10

	var wg sync.WaitGroup

	// Function to insert records concurrently (producers)
	insertRecords := func(start, end int) {
		defer wg.Done()

		for i := start; i < end; i++ {
			mTx := newMonitoredTx(
				fmt.Sprintf("0x%x", i),         // ID
				fmt.Sprintf("0x%x", i+1),       // Sender
				fmt.Sprintf("0x%x", i+2),       // Receiver
				uint64(i),                      // Nonce
				types.MonitoredTxStatusCreated, // Status
				10,                             // BlockNumber
			)
			err := storage.Add(ctx, mTx)
			require.NoError(t, err)
		}
	}

	// Start writer goroutines (multiple producers)
	for i := 0; i < numWriterGoroutines; i++ {
		wg.Add(1)
		start := i * numRecordsPerGoroutine
		end := start + numRecordsPerGoroutine
		go insertRecords(start, end)
	}

	// Wait for the single reader to finish after the writers signal completion
	wg.Wait()

	// Now ensure all the records were inserted properly (using consistent ID formatting)
	for i := 0; i < numWriterGoroutines*numRecordsPerGoroutine; i++ {
		_, err := storage.Get(ctx, common.HexToHash(fmt.Sprintf("0x%x", i)))
		require.NoError(t, err, "record not found for ID %d", i)
	}
}

func TestSqlStorage_MonitoredTxTableExists(t *testing.T) {
	storage, err := NewStorage(localCommon.SQLLiteDriverName, ":memory:")
	require.NoError(t, err)
	defer storage.db.Close()

	// Check if the monitored_txs table exists
	query := `SELECT name FROM sqlite_master WHERE type='table' AND name='monitored_txs';`
	var tableName string
	err = storage.db.QueryRow(query).Scan(&tableName)
	require.NoError(t, err)
	require.Equal(t, "monitored_txs", tableName)
}

// Helper function to create a MonitoredTx for testing
func newMonitoredTx(idHex string, fromHex string, toHex string, nonce uint64, status types.MonitoredTxStatus, blockNumber int64) types.MonitoredTx {
	return types.MonitoredTx{
		ID:        common.HexToHash(idHex),
		From:      common.HexToAddress(fromHex),
		To:        localCommon.ToAddressPtr(toHex),
		Nonce:     nonce,
		Value:     big.NewInt(10),
		Data:      nil,
		Gas:       21000,
		GasOffset: 100,
		GasPrice:  big.NewInt(1000000000),
		BlobSidecar: &ethTypes.BlobTxSidecar{
			Blobs:       []kzg4844.Blob{{1, 2, 3}},
			Commitments: []kzg4844.Commitment{{4, 5, 6}},
			Proofs:      []kzg4844.Proof{{7, 8, 9}},
		},
		Status: status,
		History: map[common.Hash]bool{
			common.HexToHash("0x1"): true,
			common.HexToHash("0x2"): false,
		},
		BlockNumber: big.NewInt(blockNumber),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// compareTxsWithout dates compares the two MonitoredTx instances, but without dates, since some functions are altering it
func compareTxsWithoutDates(t *testing.T, expected, actual types.MonitoredTx) {
	t.Helper()

	expected.CreatedAt = time.Time{}
	expected.UpdatedAt = time.Time{}
	actual.CreatedAt = time.Time{}
	actual.UpdatedAt = time.Time{}

	require.Equal(t, expected, actual)
}

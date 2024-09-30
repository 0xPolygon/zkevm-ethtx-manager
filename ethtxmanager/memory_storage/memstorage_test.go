package memorystorage

import (
	"context"
	"math/big"
	"testing"
	"time"

	localCommon "github.com/0xPolygon/zkevm-ethtx-manager/common"
	"github.com/0xPolygon/zkevm-ethtx-manager/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

// TestMemStorage_Add tests adding transactions
func TestMemStorage_Add(t *testing.T) {
	ctx := context.Background()
	storage := NewStorage()

	tests := []struct {
		name         string
		setupTx      types.MonitoredTx
		expectedErr  error
		expectedSize int
	}{
		{
			name:         "Add new transaction successfully",
			setupTx:      newMonitoredTx("0x1", "0xSender1", "0xReceiver1", 1, types.MonitoredTxStatusCreated, 10),
			expectedErr:  nil,
			expectedSize: 1,
		},
		{
			name:         "Add duplicate transaction",
			setupTx:      newMonitoredTx("0x1", "0xSender1", "0xReceiver1", 1, types.MonitoredTxStatusCreated, 10), // Same ID to simulate duplicate
			expectedErr:  types.ErrAlreadyExists,
			expectedSize: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Run Add function
			err := storage.Add(ctx, test.setupTx)

			// Assert error matches expectation
			if test.expectedErr != nil {
				require.ErrorIs(t, err, test.expectedErr)
			} else {
				require.NoError(t, err)
			}

			// Assert transaction count matches expected size
			require.Len(t, storage.Transactions, test.expectedSize)
		})
	}
}

// TestMemStorage_Remove tests removing transactions
func TestMemStorage_Remove(t *testing.T) {
	ctx := context.Background()
	storage := NewStorage()

	// Pre-load some transactions into storage
	txToRemove := newMonitoredTx("0x12345", "0xSender", "0xReceiver", 1, types.MonitoredTxStatusCreated, 10)
	err := storage.Add(ctx, txToRemove)
	require.NoError(t, err)

	tests := []struct {
		name         string
		txID         common.Hash
		expectedErr  error
		expectedSize int
	}{
		{
			name:         "Remove existing transaction successfully",
			txID:         txToRemove.ID,
			expectedErr:  nil,
			expectedSize: 0, // After removal, the size should be 0
		},
		{
			name:         "Attempt to remove non-existent transaction",
			txID:         common.HexToHash("0xNonExistent"),
			expectedErr:  types.ErrNotFound,
			expectedSize: 0, // Size should remain unchanged
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Run Remove function
			err := storage.Remove(ctx, test.txID)

			// Assert error matches expectation
			if test.expectedErr != nil {
				require.ErrorIs(t, err, test.expectedErr)
			} else {
				require.NoError(t, err)
			}

			// Assert transaction count matches expected size
			require.Len(t, storage.Transactions, test.expectedSize)
		})
	}
}

// Test for Get function
func TestMemStorage_Get(t *testing.T) {
	ctx := context.Background()
	storage := NewStorage()

	tx1 := newMonitoredTx("0x1", "0xSender1", "0xReceiver1", 1, types.MonitoredTxStatusCreated, 100)
	populateTransactions(storage, tx1)

	tests := []struct {
		name        string
		txID        string
		expectedTx  types.MonitoredTx
		expectedErr error
	}{
		{
			name:        "Get existing transaction",
			txID:        "0x1",
			expectedTx:  tx1,
			expectedErr: nil,
		},
		{
			name:        "Get non-existing transaction",
			txID:        "0xNonExistent",
			expectedTx:  types.MonitoredTx{},
			expectedErr: types.ErrNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tx, err := storage.Get(ctx, common.HexToHash(test.txID))

			if test.expectedErr != nil {
				require.ErrorIs(t, err, test.expectedErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.expectedTx, tx)
			}
		})
	}
}

// Test for GetByStatus function
func TestMemStorage_GetByStatus(t *testing.T) {
	ctx := context.Background()
	storage := NewStorage()

	tx1 := newMonitoredTx("0x1", "0xSender1", "0xReceiver1", 1, types.MonitoredTxStatusCreated, 100)
	tx2 := newMonitoredTx("0x2", "0xSender2", "0xReceiver2", 2, types.MonitoredTxStatusSent, 200)
	tx3 := newMonitoredTx("0x3", "0xSender3", "0xReceiver3", 3, types.MonitoredTxStatusFailed, 300)
	tx4 := newMonitoredTx("0x4", "0xSender4", "0xReceiver4", 4, types.MonitoredTxStatusFinalized, 400)
	populateTransactions(storage, tx1, tx2, tx3, tx4)

	tests := []struct {
		name        string
		statuses    []types.MonitoredTxStatus
		expectedTxs []types.MonitoredTx
	}{
		{
			name:        "Get all statuses",
			statuses:    nil, // nil means no status filter
			expectedTxs: []types.MonitoredTx{tx1, tx2, tx3, tx4},
		},
		{
			name:        "Get Pending and Confirmed statuses",
			statuses:    []types.MonitoredTxStatus{types.MonitoredTxStatusCreated, types.MonitoredTxStatusSent},
			expectedTxs: []types.MonitoredTx{tx1, tx2},
		},
		{
			name:        "Get only Failed status",
			statuses:    []types.MonitoredTxStatus{types.MonitoredTxStatusFailed},
			expectedTxs: []types.MonitoredTx{tx3},
		},
		{
			name:        "Get non-matching status",
			statuses:    []types.MonitoredTxStatus{types.MonitoredTxStatusSafe},
			expectedTxs: []types.MonitoredTx{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			txList, err := storage.GetByStatus(ctx, test.statuses)
			require.NoError(t, err)

			require.Equal(t, test.expectedTxs, txList)
		})
	}
}

// Test for GetByBlock function
func TestMemStorage_GetByBlock(t *testing.T) {
	ctx := context.Background()
	storage := NewStorage()

	tx1 := newMonitoredTx("0x1", "0xSender1", "0xReceiver1", 1, types.MonitoredTxStatusCreated, 100)
	tx2 := newMonitoredTx("0x2", "0xSender2", "0xReceiver2", 2, types.MonitoredTxStatusSent, 150)
	tx3 := newMonitoredTx("0x3", "0xSender3", "0xReceiver3", 3, types.MonitoredTxStatusFailed, 200)

	populateTransactions(storage, tx1, tx2, tx3)

	tests := []struct {
		name        string
		fromBlock   *uint64
		toBlock     *uint64
		expectedTxs []types.MonitoredTx
	}{
		{
			name:        "Get all transactions within block range",
			fromBlock:   localCommon.ToUint64Ptr(100),
			toBlock:     localCommon.ToUint64Ptr(200),
			expectedTxs: []types.MonitoredTx{tx1, tx2, tx3},
		},
		{
			name:        "Get transactions in middle range",
			fromBlock:   localCommon.ToUint64Ptr(120),
			toBlock:     localCommon.ToUint64Ptr(180),
			expectedTxs: []types.MonitoredTx{tx2},
		},
		{
			name:        "Get transactions above a block range",
			fromBlock:   localCommon.ToUint64Ptr(170),
			toBlock:     nil, // No upper limit
			expectedTxs: []types.MonitoredTx{tx3},
		},
		{
			name:        "Get no transactions below block range",
			fromBlock:   localCommon.ToUint64Ptr(220),
			toBlock:     nil,
			expectedTxs: []types.MonitoredTx{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			txList, err := storage.GetByBlock(ctx, test.fromBlock, test.toBlock)
			require.NoError(t, err)

			require.Equal(t, test.expectedTxs, txList)
		})
	}
}

// Test for Update method
func TestMemStorage_Update(t *testing.T) {
	ctx := context.Background()
	storage := NewStorage()

	// Initial transaction
	tx1 := newMonitoredTx("0x1", "0xSender1", "0xReceiver1", 1, types.MonitoredTxStatusCreated, 100)
	err := storage.Add(ctx, tx1)
	require.NoError(t, err)

	// Define test cases
	tests := []struct {
		name        string
		tx          types.MonitoredTx
		expectedErr error
	}{
		{
			name:        "Update existing transaction",
			tx:          newMonitoredTx("0x1", "0xSender1", "0xReceiver1", 1, types.MonitoredTxStatusCreated, 200),
			expectedErr: nil,
		},
		{
			name:        "Update non-existing transaction",
			tx:          newMonitoredTx("0x2", "0xSender2", "0xReceiver2", 2, types.MonitoredTxStatusSent, 150),
			expectedErr: types.ErrNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := storage.Update(ctx, test.tx)

			if test.expectedErr != nil {
				require.ErrorIs(t, err, test.expectedErr)
			} else {
				require.NoError(t, err)
				// Verify that the transaction was updated
				updatedTx, err := storage.Get(ctx, test.tx.ID)
				require.NoError(t, err)
				compareTxsWithoutDates(t, test.tx, updatedTx)
			}
		})
	}
}

// Test for Empty method
func TestMemStorage_Empty(t *testing.T) {
	ctx := context.Background()
	storage := NewStorage()

	// Add some transactions to the storage
	tx1 := newMonitoredTx("0x1", "0xSender1", "0xReceiver1", 1, types.MonitoredTxStatusCreated, 100)
	tx2 := newMonitoredTx("0x2", "0xSender2", "0xReceiver2", 2, types.MonitoredTxStatusSent, 200)
	_ = storage.Add(ctx, tx1)
	_ = storage.Add(ctx, tx2)

	// Ensure storage has transactions before emptying
	txList, err := storage.GetByStatus(ctx, nil)
	require.NoError(t, err)
	require.Len(t, txList, 2)

	// Call Empty method
	err = storage.Empty(ctx)
	require.NoError(t, err)

	// Verify that storage is empty
	txList, err = storage.GetByStatus(ctx, nil)
	require.NoError(t, err)
	require.Len(t, txList, 0)
}

// newMonitoredTx helper function to create MonitoredTx for tests
func newMonitoredTx(id string, from string, to string, nonce uint64, status types.MonitoredTxStatus, blockNumber uint64) types.MonitoredTx {
	return types.MonitoredTx{
		ID:          common.HexToHash(id),
		From:        common.HexToAddress(from),
		To:          localCommon.ToAddressPtr(to),
		Nonce:       nonce,
		Value:       big.NewInt(1000),
		Data:        []byte("data"),
		Gas:         21000,
		GasPrice:    big.NewInt(100),
		GasTipCap:   big.NewInt(1),
		Status:      status,
		BlockNumber: big.NewInt(int64(blockNumber)),
		History:     make(map[common.Hash]bool),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// populateTransactions is a helper function that populates the transactions map
func populateTransactions(s *MemStorage, txs ...types.MonitoredTx) {
	for _, tx := range txs {
		s.Transactions[tx.ID] = tx
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

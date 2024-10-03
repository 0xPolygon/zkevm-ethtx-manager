package ethtxmanager

import (
	context "context"
	"errors"
	"math/big"
	"testing"
	"time"

	localCommon "github.com/0xPolygon/zkevm-ethtx-manager/common"
	"github.com/0xPolygon/zkevm-ethtx-manager/ethtxmanager/sqlstorage"
	"github.com/0xPolygon/zkevm-ethtx-manager/mocks"
	"github.com/0xPolygon/zkevm-ethtx-manager/types"
	common "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetMonitoredTxnIteration(t *testing.T) {
	ctx := context.Background()
	etherman := mocks.NewEthermanInterface(t)
	storage, err := sqlstorage.NewStorage(localCommon.SQLLiteDriverName, ":memory:")
	require.NoError(t, err)

	client := &Client{
		etherman: etherman,
		storage:  storage,
	}

	tests := []struct {
		name           string
		storageTxn     *types.MonitoredTx
		ethermanNonce  uint64
		shouldUpdate   bool
		expectedResult []*monitoredTxnIteration
		expectedError  error
	}{
		{
			name:           "No transactions to update",
			expectedError:  nil,
			expectedResult: []*monitoredTxnIteration{},
		},
		{
			name: "Transaction should not update nonce",
			storageTxn: &types.MonitoredTx{
				ID:          common.HexToHash("0x1"),
				From:        common.HexToAddress("0x1"),
				BlockNumber: big.NewInt(10),
				Status:      types.MonitoredTxStatusSent,
				History: map[common.Hash]bool{
					common.HexToHash("0x1"): true,
				},
			},
			shouldUpdate: false,
			expectedResult: []*monitoredTxnIteration{
				{
					MonitoredTx: &types.MonitoredTx{
						ID:          common.HexToHash("0x1"),
						From:        common.HexToAddress("0x1"),
						BlockNumber: big.NewInt(10),
						Status:      types.MonitoredTxStatusSent,
						History: map[common.Hash]bool{
							common.HexToHash("0x1"): true,
						},
					},
					confirmed:   true,
					lastReceipt: &ethtypes.Receipt{Status: ethtypes.ReceiptStatusSuccessful},
				},
			},
			expectedError: nil,
		},
		{
			name: "Transaction should update nonce",
			storageTxn: &types.MonitoredTx{
				ID:          common.HexToHash("0x1"),
				From:        common.HexToAddress("0x1"),
				Status:      types.MonitoredTxStatusCreated,
				BlockNumber: big.NewInt(10),
			},
			shouldUpdate:  true,
			ethermanNonce: 1,
			expectedResult: []*monitoredTxnIteration{
				{
					MonitoredTx: &types.MonitoredTx{
						ID:          common.HexToHash("0x1"),
						From:        common.HexToAddress("0x1"),
						Status:      types.MonitoredTxStatusCreated,
						Nonce:       1,
						BlockNumber: big.NewInt(10),
					},
				},
			},
			expectedError: nil,
		},
		{
			name: "Error getting pending nonce",
			storageTxn: &types.MonitoredTx{
				ID:          common.HexToHash("0x1"),
				From:        common.HexToAddress("0x1"),
				Status:      types.MonitoredTxStatusCreated,
				BlockNumber: big.NewInt(10),
			},
			shouldUpdate:  true,
			expectedError: errors.New("failed to get pending nonce for sender: 0x1. Error: some error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, storage.Empty(ctx))
			if tt.storageTxn != nil {
				require.NoError(t, storage.Add(ctx, *tt.storageTxn))
			}

			etherman.ExpectedCalls = nil

			if tt.shouldUpdate {
				etherman.On("PendingNonce", ctx, common.HexToAddress("0x1")).Return(tt.ethermanNonce, tt.expectedError).Once()
			} else if len(tt.expectedResult) > 0 {
				etherman.On("CheckTxWasMined", ctx, mock.Anything).Return(tt.expectedResult[0].confirmed, tt.expectedResult[0].lastReceipt, nil)
			}

			result, err := client.getMonitoredTxnIteration(ctx)
			if tt.expectedError != nil {
				require.Error(t, err)
				require.ErrorContains(t, err, tt.expectedError.Error())
			} else {
				require.NoError(t, err)
				if len(tt.expectedResult) > 0 {
					require.Len(t, result, len(tt.expectedResult))
					compareTxsWithoutDates(t, *tt.expectedResult[0].MonitoredTx, *result[0].MonitoredTx)
				} else {
					require.Empty(t, result)
				}

				// now check from storage
				if len(tt.expectedResult) > 0 {
					dbTxns, err := storage.GetByStatus(ctx, []types.MonitoredTxStatus{tt.storageTxn.Status})
					require.NoError(t, err)
					require.Len(t, dbTxns, 1)
					require.Equal(t, tt.expectedResult[0].MonitoredTx.Nonce, dbTxns[0].Nonce)
				}
			}

			etherman.AssertExpectations(t)
		})
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

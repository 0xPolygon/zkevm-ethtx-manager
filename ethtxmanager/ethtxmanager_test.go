package ethtxmanager

import (
	context "context"
	"errors"
	"testing"

	"github.com/0xPolygonHermez/zkevm-ethtx-manager/mocks"
	common "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetMonitoredTxnIteration(t *testing.T) {
	ctx := context.Background()
	etherman := mocks.NewEthermanInterface(t)
	storage := NewMemStorage("")

	client := &Client{
		etherman: etherman,
		storage:  storage,
	}

	tests := []struct {
		name           string
		storageTxn     *monitoredTx
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
			storageTxn: &monitoredTx{
				ID:     common.HexToHash("0x1"),
				From:   common.HexToAddress("0x1"),
				Status: MonitoredTxStatusSent,
				History: map[common.Hash]bool{
					common.HexToHash("0x1"): true,
				},
			},
			shouldUpdate: false,
			expectedResult: []*monitoredTxnIteration{
				{
					monitoredTx: &monitoredTx{
						ID:     common.HexToHash("0x1"),
						From:   common.HexToAddress("0x1"),
						Status: MonitoredTxStatusSent,
						History: map[common.Hash]bool{
							common.HexToHash("0x1"): true,
						},
					},
					confirmed:   true,
					lastReceipt: &types.Receipt{Status: types.ReceiptStatusSuccessful},
				},
			},
			expectedError: nil,
		},
		{
			name: "Transaction should update nonce",
			storageTxn: &monitoredTx{
				ID:     common.HexToHash("0x1"),
				From:   common.HexToAddress("0x1"),
				Status: MonitoredTxStatusCreated,
			},
			shouldUpdate:  true,
			ethermanNonce: 1,
			expectedResult: []*monitoredTxnIteration{
				{
					monitoredTx: &monitoredTx{
						ID:     common.HexToHash("0x1"),
						From:   common.HexToAddress("0x1"),
						Status: MonitoredTxStatusCreated,
						Nonce:  1,
					},
				},
			},
			expectedError: nil,
		},
		{
			name: "Error getting pending nonce",
			storageTxn: &monitoredTx{
				ID:     common.HexToHash("0x1"),
				From:   common.HexToAddress("0x1"),
				Status: MonitoredTxStatusCreated,
			},
			shouldUpdate:  true,
			expectedError: errors.New("failed to get pending nonce for sender: 0x1. Error: some error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage.Transactions = make(map[common.Hash]monitoredTx, 1)
			if tt.storageTxn != nil {
				storage.Transactions[tt.storageTxn.ID] = *tt.storageTxn
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
				require.Equal(t, tt.expectedResult, result)

				// now check from storage
				if len(tt.expectedResult) > 0 {
					dbTxns, err := storage.GetByStatus(ctx, []MonitoredTxStatus{tt.storageTxn.Status})
					require.NoError(t, err)
					require.Len(t, dbTxns, 1)
					require.Equal(t, tt.expectedResult[0].monitoredTx.Nonce, dbTxns[0].Nonce)
				}
			}

			etherman.AssertExpectations(t)
		})
	}
}

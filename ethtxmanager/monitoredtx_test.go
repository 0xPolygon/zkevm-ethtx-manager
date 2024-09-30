package ethtxmanager

import (
	"context"
	"math/big"
	"testing"

	"github.com/0xPolygonHermez/zkevm-ethtx-manager/mocks"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto/kzg4844"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTx(t *testing.T) {
	to := common.HexToAddress("0x2")
	nonce := uint64(1)
	value := big.NewInt(2)
	data := []byte("data")
	gas := uint64(3)
	gasOffset := uint64(4)
	gasPrice := big.NewInt(5)

	mTx := monitoredTx{
		To:        &to,
		Nonce:     nonce,
		Value:     value,
		Data:      data,
		Gas:       gas,
		GasOffset: gasOffset,
		GasPrice:  gasPrice,
	}

	tx := mTx.Tx()

	assert.Equal(t, &to, tx.To())
	assert.Equal(t, nonce, tx.Nonce())
	assert.Equal(t, value, tx.Value())
	assert.Equal(t, data, tx.Data())
	assert.Equal(t, gas+gasOffset, tx.Gas())
	assert.Equal(t, gasPrice, tx.GasPrice())
}

func TestBlobTx(t *testing.T) {
	client, _ := New(Config{})
	to := common.HexToAddress("0x2")
	nonce := uint64(1)
	value := big.NewInt(2)
	data := []byte{}
	gas := uint64(3)
	gasOffset := uint64(4)
	blobGas := uint64(131072)
	blobGasPrice := big.NewInt(10)

	blobBytes := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	blob, err := client.EncodeBlobData(blobBytes)
	assert.NoError(t, err)
	blobSidecar := client.MakeBlobSidecar([]kzg4844.Blob{blob})

	mTx := monitoredTx{
		To:           &to,
		Nonce:        nonce,
		Value:        value,
		Data:         data,
		Gas:          gas,
		GasOffset:    gasOffset,
		BlobSidecar:  blobSidecar,
		BlobGas:      blobGas,
		BlobGasPrice: blobGasPrice,
	}

	tx := mTx.Tx()

	assert.Equal(t, &to, tx.To())
	assert.Equal(t, nonce, tx.Nonce())
	assert.Equal(t, value, tx.Value())
	assert.Equal(t, data, tx.Data())
	assert.Equal(t, blobSidecar, tx.BlobTxSidecar())
	assert.Equal(t, blobGas, tx.BlobGas())
	assert.Equal(t, blobGasPrice, tx.BlobGasFeeCap())
}

func TestShouldUpdateNonce(t *testing.T) {
	ctx := context.Background()
	etherman := mocks.NewEthermanInterface(t)

	tests := []struct {
		name           string
		status         MonitoredTxStatus
		history        map[common.Hash]bool
		mockResponses  []*mock.Call
		expectedResult bool
	}{
		{
			name:   "StatusCreated",
			status: MonitoredTxStatusCreated,
			history: map[common.Hash]bool{
				common.HexToHash("0x1"): true,
			},
			expectedResult: true,
		},
		{
			name:   "ConfirmedTx",
			status: MonitoredTxStatusSent,
			history: map[common.Hash]bool{
				common.HexToHash("0x1"): true,
			},
			mockResponses: []*mock.Call{
				etherman.On("CheckTxWasMined", ctx, common.HexToHash("0x1")).Return(true, &types.Receipt{Status: types.ReceiptStatusSuccessful}, nil),
			},
			expectedResult: false,
		},
		{
			name:   "FailedTx",
			status: MonitoredTxStatusSent,
			history: map[common.Hash]bool{
				common.HexToHash("0x1"): true,
			},
			mockResponses: []*mock.Call{
				etherman.On("CheckTxWasMined", ctx, common.HexToHash("0x1")).Return(true, &types.Receipt{Status: types.ReceiptStatusFailed}, nil),
			},
			expectedResult: true,
		},
		{
			name:   "PendingTx",
			status: MonitoredTxStatusSent,
			history: map[common.Hash]bool{
				common.HexToHash("0x1"): true,
			},
			mockResponses: []*mock.Call{
				etherman.On("CheckTxWasMined", ctx, common.HexToHash("0x1")).Return(false, nil, nil),
			},
			expectedResult: false,
		},
		{
			name:   "MixedTx",
			status: MonitoredTxStatusSent,
			history: map[common.Hash]bool{
				common.HexToHash("0x1"): true,
				common.HexToHash("0x2"): true,
			},
			mockResponses: []*mock.Call{
				etherman.On("CheckTxWasMined", ctx, common.HexToHash("0x2")).Return(false, nil, nil),
				etherman.On("CheckTxWasMined", ctx, common.HexToHash("0x1")).Return(true, &types.Receipt{Status: types.ReceiptStatusFailed}, nil),
			},
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			etherman.ExpectedCalls = tt.mockResponses

			for _, mockResponse := range tt.mockResponses {
				mockResponse.Once()
			}

			m := &monitoredTxnIteration{
				monitoredTx: &monitoredTx{
					Status:  tt.status,
					History: tt.history,
				},
			}

			result := m.shouldUpdateNonce(ctx, etherman)
			assert.Equal(t, tt.expectedResult, result)

			etherman.AssertExpectations(t)
		})
	}
}

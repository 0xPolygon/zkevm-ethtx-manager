package types

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto/kzg4844"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTx(t *testing.T) {
	to := common.HexToAddress("0x2")
	nonce := uint64(1)
	value := big.NewInt(2)
	data := []byte("data")
	gas := uint64(3)
	gasOffset := uint64(4)
	gasPrice := big.NewInt(5)

	mTx := MonitoredTx{
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

func TestAddHistory(t *testing.T) {
	tx := types.NewTransaction(0, common.Address{}, big.NewInt(0), 0, big.NewInt(0), nil)
	mTx := MonitoredTx{
		History: make(map[common.Hash]bool),
	}

	err := mTx.AddHistory(tx)
	assert.NoError(t, err)

	// Adding the same transaction again should return an error
	err = mTx.AddHistory(tx)
	assert.ErrorContains(t, err, "already exists")

	// should have only one history
	historySlice := mTx.HistoryHashSlice()
	assert.Len(t, historySlice, 1)
}

func TestEncodeBlobSidecarToJSON(t *testing.T) {
	// Define a valid BlobTxSidecar for testing
	validBlobSidecar := &types.BlobTxSidecar{
		Blobs:       []kzg4844.Blob{{1, 2, 3}},
		Commitments: []kzg4844.Commitment{{4, 5, 6}},
		Proofs:      []kzg4844.Proof{{7, 8, 9}},
	}

	// Define test cases
	tests := []struct {
		name          string
		mTx           *MonitoredTx
		expectedBytes []byte
		expectedErr   error
	}{
		{
			name:          "BlobSidecar is nil",
			mTx:           &MonitoredTx{BlobSidecar: nil},
			expectedBytes: nil, // Should return nil for empty blob sidecar
			expectedErr:   nil,
		},
		{
			name: "BlobSidecar is valid",
			mTx:  &MonitoredTx{BlobSidecar: validBlobSidecar},
			expectedBytes: func() []byte {
				bytes, _ := json.Marshal(validBlobSidecar)
				return bytes
			}(),
			expectedErr: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Call the method
			resultBytes, err := test.mTx.EncodeBlobSidecarToJSON()

			// Check the output matches the expected results
			if test.expectedErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.expectedErr.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, test.expectedBytes, resultBytes)
			}
		})
	}
}

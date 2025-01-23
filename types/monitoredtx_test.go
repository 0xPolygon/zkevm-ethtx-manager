package types

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
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
	tx := types.NewTransaction(0, common.HexToAddress("0x123456"), big.NewInt(100), 0, big.NewInt(10), nil)
	mTx := MonitoredTx{
		History: make(map[common.Hash]bool),
	}

	found, err := mTx.AddHistory(tx)
	assert.NoError(t, err)
	assert.False(t, found)

	// Adding the same transaction again should return an error
	found, err = mTx.AddHistory(tx)
	assert.ErrorIs(t, err, ErrAlreadyExists)
	assert.True(t, found)

	// should have only one history
	historySlice := mTx.HistoryHashSlice()
	assert.Len(t, historySlice, 1)
}

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

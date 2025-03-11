package etherman

import (
	"context"
	"errors"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/0xPolygon/zkevm-ethtx-manager/mocks"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var errGenericNotFound = errors.New("not found")

func TestExploratory(t *testing.T) {
	t.Skip("skipping test")
	url := os.Getenv("L1URL")
	ethClient, err := ethclient.Dial(url)
	require.NoError(t, err)
	sut := Client{
		EthClient: ethClient,
	}
	ctx := context.TODO()
	tx, isPending, err := sut.GetTx(ctx, common.HexToHash("0x1"))
	require.Error(t, err)
	require.Equal(t, "not found", err.Error())
	require.ErrorIs(t, err, ethereum.NotFound)
	require.False(t, isPending)
	require.Nil(t, tx)
}

func TestTranslateError(t *testing.T) {
	require.ErrorIs(t, ethereum.NotFound, translateError(ethereum.NotFound))
	require.ErrorIs(t, ethereum.NotFound, translateError(errGenericNotFound))
	anotherErr := errors.New("another error")
	require.ErrorIs(t, anotherErr, translateError(anotherErr))
}

func TestGetTx(t *testing.T) {
	mockEth := mocks.NewEthereumClient(t)
	sut := Client{
		EthClient: mockEth,
	}
	ctx := context.TODO()

	mockEth.EXPECT().TransactionByHash(mock.Anything, mock.Anything).Return(nil, false, errGenericNotFound).Once()
	tx, isPending, err := sut.GetTx(ctx, common.HexToHash("0x1"))
	require.Error(t, err)
	require.Equal(t, "not found", err.Error())
	require.ErrorIs(t, err, ethereum.NotFound)
	require.False(t, isPending)
	require.Nil(t, tx)
}

func TestGetTxReceipt(t *testing.T) {
	mockEth := mocks.NewEthereumClient(t)
	sut := Client{
		EthClient: mockEth,
	}
	mockEth.EXPECT().TransactionReceipt(mock.Anything, mock.Anything).Return(nil, errGenericNotFound).Once()
	receipt, err := sut.GetTxReceipt(context.TODO(), common.HexToHash("0x1"))
	require.ErrorIs(t, err, ethereum.NotFound)
	require.Nil(t, receipt)
}

func TestGetLatestBlockNumber(t *testing.T) {
	mockEth := mocks.NewEthereumClient(t)
	sut := Client{
		EthClient: mockEth,
	}
	mockEth.EXPECT().HeaderByNumber(mock.Anything, mock.Anything).Return(nil, errGenericNotFound).Once()
	_, err := sut.GetLatestBlockNumber(context.TODO())
	require.ErrorIs(t, err, ethereum.NotFound)
}

func TestSignTx(t *testing.T) {
	mockEth := mocks.NewEthereumClient(t)
	sut := Client{
		EthClient: mockEth,
		auth:      make(map[common.Address]bind.TransactOpts),
	}
	to := common.HexToAddress("0x1")
	privateKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	signer, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(1337))
	require.NoError(t, err)
	sut.auth[to] = *signer

	tx := ethTypes.NewTx(&ethTypes.LegacyTx{To: &to, Nonce: uint64(0), Value: big.NewInt(0), Data: []byte{}})
	_, err = sut.SignTx(context.TODO(), to, tx)
	require.NoError(t, err)
}

func TestGetRevertMessage(t *testing.T) {
	mockEth := mocks.NewEthereumClient(t)
	sut := Client{
		EthClient: mockEth,
	}
	mockEth.EXPECT().TransactionReceipt(context.TODO(), mock.Anything).Return(nil, errGenericNotFound).Once()
	to := common.HexToAddress("0x1")
	tx := ethTypes.NewTx(&ethTypes.LegacyTx{To: &to, Nonce: uint64(0), Value: big.NewInt(0), Data: []byte{}})
	_, err := sut.GetRevertMessage(context.TODO(), tx)
	require.ErrorIs(t, err, ethereum.NotFound)
}

func TestWaitTxToBeMined(t *testing.T) {
	mockEth := mocks.NewEthereumClient(t)
	sut := Client{
		EthClient: mockEth,
	}
	mockEth.EXPECT().TransactionReceipt(mock.Anything, mock.Anything).Return(nil, errGenericNotFound)
	to := common.HexToAddress("0x1")
	tx := ethTypes.NewTx(&ethTypes.LegacyTx{To: &to, Nonce: uint64(0), Value: big.NewInt(0), Data: []byte{}})
	_, err := sut.WaitTxToBeMined(context.TODO(), tx, time.Second)
	require.Error(t, err)
}

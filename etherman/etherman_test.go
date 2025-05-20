package etherman

import (
	"context"
	"errors"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/0xPolygon/zkevm-ethtx-manager/mocks"
	signertypes "github.com/agglayer/go_signer/signer/types"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
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
	mockSigner := mocks.NewSigner(t)
	to := common.HexToAddress("0x1")
	ethSigners := &EthermanSigners{
		chainID: 1337,
		signers: map[common.Address]signertypes.Signer{
			to: mockSigner,
		},
	}
	sut := Client{
		EthClient: mockEth,
		auth:      ethSigners,
	}
	mockSigner.EXPECT().SignTx(context.TODO(), mock.Anything).Return(nil, nil).Once()
	tx := ethTypes.NewTx(&ethTypes.LegacyTx{To: &to, Nonce: uint64(0), Value: big.NewInt(0), Data: []byte{}})
	_, err := sut.SignTx(context.TODO(), to, tx)
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
	res, err := sut.WaitTxToBeMined(context.TODO(), tx, time.Second)
	require.False(t, res)
	require.NoError(t, err)
}

func TestCheckTxWasMined(t *testing.T) {
	mockEth := mocks.NewEthereumClient(t)
	sut := Client{
		EthClient: mockEth,
	}
	mockEth.EXPECT().TransactionReceipt(mock.Anything, mock.Anything).Return(nil, errGenericNotFound)
	to := common.HexToAddress("0x1")
	tx := ethTypes.NewTx(&ethTypes.LegacyTx{To: &to, Nonce: uint64(0), Value: big.NewInt(0), Data: []byte{}})
	res, receipt, err := sut.CheckTxWasMined(context.TODO(), tx.Hash())
	require.Nil(t, receipt)
	require.False(t, res)
	require.NoError(t, err)
}

func TestNewClient(t *testing.T) {
	mockEth := mocks.NewEthereumClient(t)
	ethclientFactoryFunc = func(url string) (EthereumClient, error) {
		return mockEth, nil
	}
	mockEth.EXPECT().ChainID(mock.Anything).Return(big.NewInt(1), nil)
	sut, err := NewClient(Config{
		URL: "http://localhost:8545",
	}, nil)
	require.NoError(t, err)
	require.NotNil(t, sut)
}

func TestNewClientDefaultConfig(t *testing.T) {
	mockEth := mocks.NewEthereumClient(t)
	ethclientFactoryFunc = func(url string) (EthereumClient, error) {
		return mockEth, nil
	}
	mockEth.EXPECT().ChainID(mock.Anything).Return(big.NewInt(1), nil)
	sut, err := NewClient(Config{
		URL: "http://localhost:8545",
	}, []signertypes.SignerConfig{
		{
			Config: map[string]interface{}{
				"path":     "../test/test.keystore",
				"password": "testonly",
			},
		},
	})
	require.NoError(t, err)
	require.NotNil(t, sut)
}

func TestPublicAddress(t *testing.T) {
	mockSigner := mocks.NewSigner(t)
	senderAddr := common.HexToAddress("0x1")
	signers := &EthermanSigners{
		signers: map[common.Address]signertypes.Signer{
			senderAddr: mockSigner,
		},
	}
	sut := Client{
		auth: signers,
	}
	mockSigner.EXPECT().PublicAddress().Return(senderAddr)
	addr, err := sut.PublicAddress()
	require.NoError(t, err)
	require.Len(t, addr, 1)
}

func TestGetL1GasPrice(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		mockSetup     func(mockEth *mocks.EthereumClient)
		expectedPrice *big.Int
		expectedError error
	}{
		{
			name: "all providers fail",
			mockSetup: func(mockEth *mocks.EthereumClient) {
				err := errors.New("failed to get gas price from all providers")
				mockEth.On("SuggestGasPrice", mock.Anything).Return(nil, err).Once()
			},
			expectedPrice: nil,
			expectedError: ErrGasPriceProviders,
		},
		{
			name: "provider returns valid gas price",
			mockSetup: func(mockEth *mocks.EthereumClient) {
				mockEth.On("SuggestGasPrice", mock.Anything).Return(big.NewInt(100), nil).Once()
			},
			expectedPrice: big.NewInt(100),
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockEth := mocks.NewEthereumClient(t)
			tt.mockSetup(mockEth)

			client := &Client{
				GasProviders: externalGasProviders{
					Providers: []ethereum.GasPricer{mockEth},
				},
			}

			price, err := client.GetL1GasPrice(ctx)

			if tt.expectedError != nil {
				require.ErrorIs(t, err, tt.expectedError)
				require.Nil(t, price)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedPrice, price)
			}

			mockEth.AssertExpectations(t)
		})
	}
}

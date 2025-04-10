package etherman

import (
	"context"
	"testing"

	"github.com/0xPolygon/zkevm-ethtx-manager/mocks"
	signertypes "github.com/agglayer/go_signer/signer/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	fileKeystorePath     = "../test/test.keystore"
	fileKeystorePassword = "testonly"
)

func TestNewEthermanSigners(t *testing.T) {
	ctx := context.TODO()
	chainID := uint64(1)
	config := []signertypes.SignerConfig{}
	got, err := NewEthermanSigners(ctx, chainID, config)
	require.NoError(t, err)
	require.NotNil(t, got)

	got, err = NewEthermanSigners(ctx, chainID, []signertypes.SignerConfig{
		{
			Method: "error",
		},
	})
	require.Error(t, err)
	require.Nil(t, got)

	_, err = NewEthermanSigners(ctx, chainID, []signertypes.SignerConfig{
		{
			Method: "local",
			Config: map[string]interface{}{
				"path":     "dontexists",
				"password": "password",
			},
		},
	})
	require.Error(t, err)
	cfg := []signertypes.SignerConfig{
		{
			Method: "local",
			Config: map[string]interface{}{
				"path":     fileKeystorePath,
				"password": fileKeystorePassword,
			},
		},
	}
	_, err = NewEthermanSigners(ctx, chainID, cfg)
	require.NoError(t, err)
	cfg = append(cfg, cfg[0])
	_, err = NewEthermanSigners(ctx, chainID, cfg)
	require.Error(t, err, "it must fail because it should detect duplicate signer with the same public address")
	require.ErrorContains(t, err, "multiple signers for address")
}

func TestEthermanSignersSignTx(t *testing.T) {
	mockSigner := mocks.NewSigner(t)
	senderAddr := common.HexToAddress("0x1")
	sut := &EthermanSigners{
		chainID: 1,
		signers: map[common.Address]signertypes.Signer{
			common.HexToAddress(senderAddr.Hex()): mockSigner,
		},
	}

	_, err := sut.SignTx(context.TODO(), common.HexToAddress("0x2"), nil)
	require.ErrorIs(t, err, ErrPrivateKeyNotFound, "for address 0x2 there are no signer")
	var nilSut *EthermanSigners = nil
	_, err = nilSut.SignTx(context.TODO(), common.HexToAddress("0x1"), nil)
	require.ErrorIs(t, err, ErrObjectIsNil, "the object is nil")
	var tx *types.Transaction = nil
	mockSigner.EXPECT().SignTx(mock.Anything, tx).Return(nil, nil)
	_, err = sut.SignTx(context.TODO(), senderAddr, tx)
	require.NoError(t, err)
}

func TestEthermanSignersPublicAddress(t *testing.T) {
	mockSigner := mocks.NewSigner(t)
	senderAddr := common.HexToAddress("0x1")
	sut := &EthermanSigners{
		chainID: 1,
		signers: map[common.Address]signertypes.Signer{
			common.HexToAddress(senderAddr.Hex()): mockSigner,
		},
	}
	mockSigner.EXPECT().PublicAddress().Return(senderAddr)
	addresses, err := sut.PublicAddress()
	require.NoError(t, err)
	require.Len(t, addresses, 1)
	require.Equal(t, senderAddr, addresses[0])

	var nilSut *EthermanSigners = nil
	addresses, err = nilSut.PublicAddress()
	require.NoError(t, err)
	require.Nil(t, addresses)
}

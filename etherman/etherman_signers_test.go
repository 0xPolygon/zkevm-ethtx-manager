package etherman

import (
	"context"
	"testing"

	"github.com/0xPolygon/zkevm-ethtx-manager/mocks"
	signertypes "github.com/agglayer/go_signer/signer/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
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

}

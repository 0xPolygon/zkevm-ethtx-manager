package etherman

import (
	"context"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"
)

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

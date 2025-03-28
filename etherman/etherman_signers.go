package etherman

import (
	"context"
	"fmt"

	"github.com/agglayer/go_signer/signer"
	signertypes "github.com/agglayer/go_signer/signer/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type EthermanSigners struct {
	chainID uint64
	signers map[common.Address]signertypes.Signer
}

func NewEthermanSigners(ctx context.Context, chainID uint64, config []signertypes.SignerConfig) (*EthermanSigners, error) {
	res := EthermanSigners{
		chainID: chainID,
		signers: make(map[common.Address]signertypes.Signer),
	}
	for i, signerConfig := range config {
		signer, err := signer.NewSigner(ctx, chainID, signerConfig, fmt.Sprintf("signer-%d", i), nil)
		if err != nil {
			return nil, err
		}
		res.signers[signer.PublicAddress()] = signer
	}
	return &res, nil
}

// getAuthByAddress tries to get an authorization from the authorizations map
func (s *EthermanSigners) getSignerByAddress(addr common.Address) (signertypes.Signer, error) {
	if s == nil {
		return nil, ErrPrivateKeyNotFound
	}
	signer, found := s.signers[addr]
	if !found {
		return nil, ErrPrivateKeyNotFound
	}
	return signer, nil
}

func (s *EthermanSigners) PublicAddress() ([]common.Address, error) {
	res := make([]common.Address, 0, len(s.signers))
	if s == nil {
		return res, nil
	}
	for _, signer := range s.signers {
		res = append(res, signer.PublicAddress())
	}
	return res, nil
}

func (s *EthermanSigners) SignTx(ctx context.Context, sender common.Address, tx *types.Transaction) (*types.Transaction, error) {
	signer, err := s.getSignerByAddress(sender)
	if err != nil {
		return nil, err
	}
	return signer.SignTx(ctx, tx)
}

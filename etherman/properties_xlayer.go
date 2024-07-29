package etherman

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
)

// GetZkEVMAddressAndL1ChainID returns the ZkEVM address and the L1 chain ID
func (etherMan *Client) GetZkEVMAddressAndL1ChainID() (common.Address, common.Address, uint64, error) {
	if etherMan == nil {
		return common.Address{}, common.Address{}, 0, fmt.Errorf("etherMan is nil")
	}
	return etherMan.cfg.ZkEVMAddr, etherMan.cfg.RollupManagerAddr, etherMan.cfg.L1ChainID, nil
}

package ethtxmanager

import (
	"context"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// MemStorage hold txs to be managed
type MemStorage struct {
	TxsMutex     sync.RWMutex
	Transactions map[common.Hash]monitoredTx
}

// NewPostgresStorage creates a new instance of storage that use
// postgres to store data
func NewMemStorage() *MemStorage {
	return &MemStorage{TxsMutex: sync.RWMutex{},
		Transactions: make(map[common.Hash]monitoredTx),
	}
}

// Add persist a monitored tx
func (s *MemStorage) Add(ctx context.Context, mTx monitoredTx) error {
	mTx.createdAt = time.Now()
	s.TxsMutex.Lock()
	defer s.TxsMutex.Unlock()
	if _, exists := s.Transactions[mTx.id]; exists {
		return ErrAlreadyExists
	}
	s.Transactions[mTx.id] = mTx
	return nil
}

// Get loads a persisted monitored tx
func (s *MemStorage) Get(ctx context.Context, id common.Hash) (monitoredTx, error) {
	s.TxsMutex.RLock()
	defer s.TxsMutex.RUnlock()
	if mTx, exists := s.Transactions[id]; exists {
		return mTx, nil
	}
	return monitoredTx{}, ErrNotFound
}

// GetByStatus loads all monitored tx that match the provided status
func (s *MemStorage) GetByStatus(ctx context.Context, statuses []MonitoredTxStatus) ([]monitoredTx, error) {
	mTxs := []monitoredTx{}
	s.TxsMutex.RLock()
	for _, mTx := range s.Transactions {
		if len(statuses) > 0 {
			for _, status := range statuses {
				if mTx.status == status {
					mTxs = append(mTxs, mTx)
				}
			}
		} else {
			mTxs = append(mTxs, mTx)
		}
	}
	defer s.TxsMutex.RUnlock()

	return mTxs, nil
}

// GetByBlock loads all monitored tx that have the blockNumber between
// fromBlock and toBlock
func (s *MemStorage) GetByBlock(ctx context.Context, fromBlock, toBlock *uint64) ([]monitoredTx, error) {
	mTxs := []monitoredTx{}
	s.TxsMutex.RLock()
	for _, mTx := range s.Transactions {
		if fromBlock != nil && mTx.blockNumber.Uint64() < *fromBlock {
			continue
		}
		if toBlock != nil && mTx.blockNumber.Uint64() > *toBlock {
			continue
		}
		mTxs = append(mTxs, mTx)
	}
	s.TxsMutex.RUnlock()
	return mTxs, nil
}

// Update a persisted monitored tx
func (s *MemStorage) Update(ctx context.Context, mTx monitoredTx) error {
	mTx.updatedAt = time.Now()
	s.TxsMutex.Lock()
	defer s.TxsMutex.Unlock()
	if _, exists := s.Transactions[mTx.id]; !exists {
		return ErrNotFound
	}
	s.Transactions[mTx.id] = mTx
	return nil
}

package ethtxmanager

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/0xPolygon/zkevm-ethtx-manager/types"
	"github.com/ethereum/go-ethereum/common"
)

// MemStorage represents a thread-safe in-memory storage for MonitoredTx object
type MemStorage struct {
	TxsMutex     sync.RWMutex
	Transactions map[common.Hash]types.MonitoredTx
}

// NewMemStorage creates a new instance of storage
func NewMemStorage() *MemStorage {
	return &MemStorage{Transactions: make(map[common.Hash]types.MonitoredTx)}
}

// Add persist a monitored tx
func (s *MemStorage) Add(ctx context.Context, mTx types.MonitoredTx) error {
	mTx.CreatedAt = time.Now()

	s.TxsMutex.Lock()
	defer s.TxsMutex.Unlock()

	if _, exists := s.Transactions[mTx.ID]; exists {
		return ErrAlreadyExists
	}
	s.Transactions[mTx.ID] = mTx
	return nil
}

// Remove a persisted monitored tx
func (s *MemStorage) Remove(ctx context.Context, id common.Hash) error {
	s.TxsMutex.Lock()
	defer s.TxsMutex.Unlock()

	if _, exists := s.Transactions[id]; !exists {
		return ErrNotFound
	}
	delete(s.Transactions, id)
	return nil
}

// Get loads a persisted monitored tx
func (s *MemStorage) Get(ctx context.Context, id common.Hash) (types.MonitoredTx, error) {
	s.TxsMutex.RLock()
	defer s.TxsMutex.RUnlock()

	if mTx, exists := s.Transactions[id]; exists {
		return mTx, nil
	}
	return types.MonitoredTx{}, ErrNotFound
}

// GetByStatus loads all monitored transactions that match the provided statuses
func (s *MemStorage) GetByStatus(ctx context.Context, statuses []types.MonitoredTxStatus) ([]types.MonitoredTx, error) {
	s.TxsMutex.RLock()
	defer s.TxsMutex.RUnlock()

	// Filter transactions based on statuses
	matchingTxs := make([]types.MonitoredTx, 0, len(s.Transactions))
	for _, mTx := range s.Transactions {
		// If no statuses are provided, add all transactions
		if len(statuses) == 0 || containsStatus(mTx.Status, statuses) {
			matchingTxs = append(matchingTxs, mTx)
		}
	}

	// Sort transactions by creation date (oldest first)
	sort.Slice(matchingTxs, func(i, j int) bool {
		return matchingTxs[i].CreatedAt.Before(matchingTxs[j].CreatedAt)
	})

	return matchingTxs, nil
}

// containsStatus checks if a status is in the statuses slice
func containsStatus(status types.MonitoredTxStatus, statuses []types.MonitoredTxStatus) bool {
	for _, s := range statuses {
		if s == status {
			return true
		}
	}
	return false
}

// GetByBlock loads all monitored tx that have the blockNumber between fromBlock and toBlock
func (s *MemStorage) GetByBlock(ctx context.Context, fromBlock, toBlock *uint64) ([]types.MonitoredTx, error) {
	mTxs := []types.MonitoredTx{}
	s.TxsMutex.RLock()
	defer s.TxsMutex.RUnlock()

	for _, mTx := range s.Transactions {
		if fromBlock != nil && mTx.BlockNumber.Uint64() < *fromBlock {
			continue
		}
		if toBlock != nil && mTx.BlockNumber.Uint64() > *toBlock {
			continue
		}

		mTxs = append(mTxs, mTx)
	}
	return mTxs, nil
}

// Update a persisted monitored tx
func (s *MemStorage) Update(ctx context.Context, mTx types.MonitoredTx) error {
	mTx.UpdatedAt = time.Now()
	s.TxsMutex.Lock()
	defer s.TxsMutex.Unlock()

	if _, exists := s.Transactions[mTx.ID]; !exists {
		return ErrNotFound
	}
	s.Transactions[mTx.ID] = mTx
	return nil
}

// Empty the storage
func (s *MemStorage) Empty(ctx context.Context) error {
	s.TxsMutex.Lock()
	defer s.TxsMutex.Unlock()

	s.Transactions = make(map[common.Hash]types.MonitoredTx)
	return nil
}

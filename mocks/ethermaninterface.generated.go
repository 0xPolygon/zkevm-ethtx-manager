// Code generated by mockery v2.45.0. DO NOT EDIT.

package mocks

import (
	context "context"
	big "math/big"

	common "github.com/ethereum/go-ethereum/common"

	mock "github.com/stretchr/testify/mock"

	time "time"

	types "github.com/ethereum/go-ethereum/core/types"
)

// EthermanInterface is an autogenerated mock type for the EthermanInterface type
type EthermanInterface struct {
	mock.Mock
}

type EthermanInterface_Expecter struct {
	mock *mock.Mock
}

func (_m *EthermanInterface) EXPECT() *EthermanInterface_Expecter {
	return &EthermanInterface_Expecter{mock: &_m.Mock}
}

// CheckTxWasMined provides a mock function with given fields: ctx, txHash
func (_m *EthermanInterface) CheckTxWasMined(ctx context.Context, txHash common.Hash) (bool, *types.Receipt, error) {
	ret := _m.Called(ctx, txHash)

	if len(ret) == 0 {
		panic("no return value specified for CheckTxWasMined")
	}

	var r0 bool
	var r1 *types.Receipt
	var r2 error
	if rf, ok := ret.Get(0).(func(context.Context, common.Hash) (bool, *types.Receipt, error)); ok {
		return rf(ctx, txHash)
	}
	if rf, ok := ret.Get(0).(func(context.Context, common.Hash) bool); ok {
		r0 = rf(ctx, txHash)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(context.Context, common.Hash) *types.Receipt); ok {
		r1 = rf(ctx, txHash)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*types.Receipt)
		}
	}

	if rf, ok := ret.Get(2).(func(context.Context, common.Hash) error); ok {
		r2 = rf(ctx, txHash)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// EthermanInterface_CheckTxWasMined_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CheckTxWasMined'
type EthermanInterface_CheckTxWasMined_Call struct {
	*mock.Call
}

// CheckTxWasMined is a helper method to define mock.On call
//   - ctx context.Context
//   - txHash common.Hash
func (_e *EthermanInterface_Expecter) CheckTxWasMined(ctx interface{}, txHash interface{}) *EthermanInterface_CheckTxWasMined_Call {
	return &EthermanInterface_CheckTxWasMined_Call{Call: _e.mock.On("CheckTxWasMined", ctx, txHash)}
}

func (_c *EthermanInterface_CheckTxWasMined_Call) Run(run func(ctx context.Context, txHash common.Hash)) *EthermanInterface_CheckTxWasMined_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(common.Hash))
	})
	return _c
}

func (_c *EthermanInterface_CheckTxWasMined_Call) Return(_a0 bool, _a1 *types.Receipt, _a2 error) *EthermanInterface_CheckTxWasMined_Call {
	_c.Call.Return(_a0, _a1, _a2)
	return _c
}

func (_c *EthermanInterface_CheckTxWasMined_Call) RunAndReturn(run func(context.Context, common.Hash) (bool, *types.Receipt, error)) *EthermanInterface_CheckTxWasMined_Call {
	_c.Call.Return(run)
	return _c
}

// CurrentNonce provides a mock function with given fields: ctx, account
func (_m *EthermanInterface) CurrentNonce(ctx context.Context, account common.Address) (uint64, error) {
	ret := _m.Called(ctx, account)

	if len(ret) == 0 {
		panic("no return value specified for CurrentNonce")
	}

	var r0 uint64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, common.Address) (uint64, error)); ok {
		return rf(ctx, account)
	}
	if rf, ok := ret.Get(0).(func(context.Context, common.Address) uint64); ok {
		r0 = rf(ctx, account)
	} else {
		r0 = ret.Get(0).(uint64)
	}

	if rf, ok := ret.Get(1).(func(context.Context, common.Address) error); ok {
		r1 = rf(ctx, account)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// EthermanInterface_CurrentNonce_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CurrentNonce'
type EthermanInterface_CurrentNonce_Call struct {
	*mock.Call
}

// CurrentNonce is a helper method to define mock.On call
//   - ctx context.Context
//   - account common.Address
func (_e *EthermanInterface_Expecter) CurrentNonce(ctx interface{}, account interface{}) *EthermanInterface_CurrentNonce_Call {
	return &EthermanInterface_CurrentNonce_Call{Call: _e.mock.On("CurrentNonce", ctx, account)}
}

func (_c *EthermanInterface_CurrentNonce_Call) Run(run func(ctx context.Context, account common.Address)) *EthermanInterface_CurrentNonce_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(common.Address))
	})
	return _c
}

func (_c *EthermanInterface_CurrentNonce_Call) Return(_a0 uint64, _a1 error) *EthermanInterface_CurrentNonce_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *EthermanInterface_CurrentNonce_Call) RunAndReturn(run func(context.Context, common.Address) (uint64, error)) *EthermanInterface_CurrentNonce_Call {
	_c.Call.Return(run)
	return _c
}

// EstimateGas provides a mock function with given fields: ctx, from, to, value, data
func (_m *EthermanInterface) EstimateGas(ctx context.Context, from common.Address, to *common.Address, value *big.Int, data []byte) (uint64, error) {
	ret := _m.Called(ctx, from, to, value, data)

	if len(ret) == 0 {
		panic("no return value specified for EstimateGas")
	}

	var r0 uint64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, common.Address, *common.Address, *big.Int, []byte) (uint64, error)); ok {
		return rf(ctx, from, to, value, data)
	}
	if rf, ok := ret.Get(0).(func(context.Context, common.Address, *common.Address, *big.Int, []byte) uint64); ok {
		r0 = rf(ctx, from, to, value, data)
	} else {
		r0 = ret.Get(0).(uint64)
	}

	if rf, ok := ret.Get(1).(func(context.Context, common.Address, *common.Address, *big.Int, []byte) error); ok {
		r1 = rf(ctx, from, to, value, data)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// EthermanInterface_EstimateGas_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'EstimateGas'
type EthermanInterface_EstimateGas_Call struct {
	*mock.Call
}

// EstimateGas is a helper method to define mock.On call
//   - ctx context.Context
//   - from common.Address
//   - to *common.Address
//   - value *big.Int
//   - data []byte
func (_e *EthermanInterface_Expecter) EstimateGas(ctx interface{}, from interface{}, to interface{}, value interface{}, data interface{}) *EthermanInterface_EstimateGas_Call {
	return &EthermanInterface_EstimateGas_Call{Call: _e.mock.On("EstimateGas", ctx, from, to, value, data)}
}

func (_c *EthermanInterface_EstimateGas_Call) Run(run func(ctx context.Context, from common.Address, to *common.Address, value *big.Int, data []byte)) *EthermanInterface_EstimateGas_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(common.Address), args[2].(*common.Address), args[3].(*big.Int), args[4].([]byte))
	})
	return _c
}

func (_c *EthermanInterface_EstimateGas_Call) Return(_a0 uint64, _a1 error) *EthermanInterface_EstimateGas_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *EthermanInterface_EstimateGas_Call) RunAndReturn(run func(context.Context, common.Address, *common.Address, *big.Int, []byte) (uint64, error)) *EthermanInterface_EstimateGas_Call {
	_c.Call.Return(run)
	return _c
}

// EstimateGasBlobTx provides a mock function with given fields: ctx, from, to, gasFeeCap, gasTipCap, value, data
func (_m *EthermanInterface) EstimateGasBlobTx(ctx context.Context, from common.Address, to *common.Address, gasFeeCap *big.Int, gasTipCap *big.Int, value *big.Int, data []byte) (uint64, error) {
	ret := _m.Called(ctx, from, to, gasFeeCap, gasTipCap, value, data)

	if len(ret) == 0 {
		panic("no return value specified for EstimateGasBlobTx")
	}

	var r0 uint64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, common.Address, *common.Address, *big.Int, *big.Int, *big.Int, []byte) (uint64, error)); ok {
		return rf(ctx, from, to, gasFeeCap, gasTipCap, value, data)
	}
	if rf, ok := ret.Get(0).(func(context.Context, common.Address, *common.Address, *big.Int, *big.Int, *big.Int, []byte) uint64); ok {
		r0 = rf(ctx, from, to, gasFeeCap, gasTipCap, value, data)
	} else {
		r0 = ret.Get(0).(uint64)
	}

	if rf, ok := ret.Get(1).(func(context.Context, common.Address, *common.Address, *big.Int, *big.Int, *big.Int, []byte) error); ok {
		r1 = rf(ctx, from, to, gasFeeCap, gasTipCap, value, data)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// EthermanInterface_EstimateGasBlobTx_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'EstimateGasBlobTx'
type EthermanInterface_EstimateGasBlobTx_Call struct {
	*mock.Call
}

// EstimateGasBlobTx is a helper method to define mock.On call
//   - ctx context.Context
//   - from common.Address
//   - to *common.Address
//   - gasFeeCap *big.Int
//   - gasTipCap *big.Int
//   - value *big.Int
//   - data []byte
func (_e *EthermanInterface_Expecter) EstimateGasBlobTx(ctx interface{}, from interface{}, to interface{}, gasFeeCap interface{}, gasTipCap interface{}, value interface{}, data interface{}) *EthermanInterface_EstimateGasBlobTx_Call {
	return &EthermanInterface_EstimateGasBlobTx_Call{Call: _e.mock.On("EstimateGasBlobTx", ctx, from, to, gasFeeCap, gasTipCap, value, data)}
}

func (_c *EthermanInterface_EstimateGasBlobTx_Call) Run(run func(ctx context.Context, from common.Address, to *common.Address, gasFeeCap *big.Int, gasTipCap *big.Int, value *big.Int, data []byte)) *EthermanInterface_EstimateGasBlobTx_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(common.Address), args[2].(*common.Address), args[3].(*big.Int), args[4].(*big.Int), args[5].(*big.Int), args[6].([]byte))
	})
	return _c
}

func (_c *EthermanInterface_EstimateGasBlobTx_Call) Return(_a0 uint64, _a1 error) *EthermanInterface_EstimateGasBlobTx_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *EthermanInterface_EstimateGasBlobTx_Call) RunAndReturn(run func(context.Context, common.Address, *common.Address, *big.Int, *big.Int, *big.Int, []byte) (uint64, error)) *EthermanInterface_EstimateGasBlobTx_Call {
	_c.Call.Return(run)
	return _c
}

// GetHeaderByNumber provides a mock function with given fields: ctx, number
func (_m *EthermanInterface) GetHeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	ret := _m.Called(ctx, number)

	if len(ret) == 0 {
		panic("no return value specified for GetHeaderByNumber")
	}

	var r0 *types.Header
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *big.Int) (*types.Header, error)); ok {
		return rf(ctx, number)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *big.Int) *types.Header); ok {
		r0 = rf(ctx, number)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.Header)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *big.Int) error); ok {
		r1 = rf(ctx, number)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// EthermanInterface_GetHeaderByNumber_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetHeaderByNumber'
type EthermanInterface_GetHeaderByNumber_Call struct {
	*mock.Call
}

// GetHeaderByNumber is a helper method to define mock.On call
//   - ctx context.Context
//   - number *big.Int
func (_e *EthermanInterface_Expecter) GetHeaderByNumber(ctx interface{}, number interface{}) *EthermanInterface_GetHeaderByNumber_Call {
	return &EthermanInterface_GetHeaderByNumber_Call{Call: _e.mock.On("GetHeaderByNumber", ctx, number)}
}

func (_c *EthermanInterface_GetHeaderByNumber_Call) Run(run func(ctx context.Context, number *big.Int)) *EthermanInterface_GetHeaderByNumber_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*big.Int))
	})
	return _c
}

func (_c *EthermanInterface_GetHeaderByNumber_Call) Return(_a0 *types.Header, _a1 error) *EthermanInterface_GetHeaderByNumber_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *EthermanInterface_GetHeaderByNumber_Call) RunAndReturn(run func(context.Context, *big.Int) (*types.Header, error)) *EthermanInterface_GetHeaderByNumber_Call {
	_c.Call.Return(run)
	return _c
}

// GetLatestBlockNumber provides a mock function with given fields: ctx
func (_m *EthermanInterface) GetLatestBlockNumber(ctx context.Context) (uint64, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for GetLatestBlockNumber")
	}

	var r0 uint64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (uint64, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) uint64); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Get(0).(uint64)
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// EthermanInterface_GetLatestBlockNumber_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetLatestBlockNumber'
type EthermanInterface_GetLatestBlockNumber_Call struct {
	*mock.Call
}

// GetLatestBlockNumber is a helper method to define mock.On call
//   - ctx context.Context
func (_e *EthermanInterface_Expecter) GetLatestBlockNumber(ctx interface{}) *EthermanInterface_GetLatestBlockNumber_Call {
	return &EthermanInterface_GetLatestBlockNumber_Call{Call: _e.mock.On("GetLatestBlockNumber", ctx)}
}

func (_c *EthermanInterface_GetLatestBlockNumber_Call) Run(run func(ctx context.Context)) *EthermanInterface_GetLatestBlockNumber_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *EthermanInterface_GetLatestBlockNumber_Call) Return(_a0 uint64, _a1 error) *EthermanInterface_GetLatestBlockNumber_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *EthermanInterface_GetLatestBlockNumber_Call) RunAndReturn(run func(context.Context) (uint64, error)) *EthermanInterface_GetLatestBlockNumber_Call {
	_c.Call.Return(run)
	return _c
}

// GetRevertMessage provides a mock function with given fields: ctx, tx
func (_m *EthermanInterface) GetRevertMessage(ctx context.Context, tx *types.Transaction) (string, error) {
	ret := _m.Called(ctx, tx)

	if len(ret) == 0 {
		panic("no return value specified for GetRevertMessage")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *types.Transaction) (string, error)); ok {
		return rf(ctx, tx)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *types.Transaction) string); ok {
		r0 = rf(ctx, tx)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, *types.Transaction) error); ok {
		r1 = rf(ctx, tx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// EthermanInterface_GetRevertMessage_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetRevertMessage'
type EthermanInterface_GetRevertMessage_Call struct {
	*mock.Call
}

// GetRevertMessage is a helper method to define mock.On call
//   - ctx context.Context
//   - tx *types.Transaction
func (_e *EthermanInterface_Expecter) GetRevertMessage(ctx interface{}, tx interface{}) *EthermanInterface_GetRevertMessage_Call {
	return &EthermanInterface_GetRevertMessage_Call{Call: _e.mock.On("GetRevertMessage", ctx, tx)}
}

func (_c *EthermanInterface_GetRevertMessage_Call) Run(run func(ctx context.Context, tx *types.Transaction)) *EthermanInterface_GetRevertMessage_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*types.Transaction))
	})
	return _c
}

func (_c *EthermanInterface_GetRevertMessage_Call) Return(_a0 string, _a1 error) *EthermanInterface_GetRevertMessage_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *EthermanInterface_GetRevertMessage_Call) RunAndReturn(run func(context.Context, *types.Transaction) (string, error)) *EthermanInterface_GetRevertMessage_Call {
	_c.Call.Return(run)
	return _c
}

// GetSuggestGasTipCap provides a mock function with given fields: ctx
func (_m *EthermanInterface) GetSuggestGasTipCap(ctx context.Context) (*big.Int, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for GetSuggestGasTipCap")
	}

	var r0 *big.Int
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (*big.Int, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) *big.Int); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*big.Int)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// EthermanInterface_GetSuggestGasTipCap_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetSuggestGasTipCap'
type EthermanInterface_GetSuggestGasTipCap_Call struct {
	*mock.Call
}

// GetSuggestGasTipCap is a helper method to define mock.On call
//   - ctx context.Context
func (_e *EthermanInterface_Expecter) GetSuggestGasTipCap(ctx interface{}) *EthermanInterface_GetSuggestGasTipCap_Call {
	return &EthermanInterface_GetSuggestGasTipCap_Call{Call: _e.mock.On("GetSuggestGasTipCap", ctx)}
}

func (_c *EthermanInterface_GetSuggestGasTipCap_Call) Run(run func(ctx context.Context)) *EthermanInterface_GetSuggestGasTipCap_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *EthermanInterface_GetSuggestGasTipCap_Call) Return(_a0 *big.Int, _a1 error) *EthermanInterface_GetSuggestGasTipCap_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *EthermanInterface_GetSuggestGasTipCap_Call) RunAndReturn(run func(context.Context) (*big.Int, error)) *EthermanInterface_GetSuggestGasTipCap_Call {
	_c.Call.Return(run)
	return _c
}

// GetTx provides a mock function with given fields: ctx, txHash
func (_m *EthermanInterface) GetTx(ctx context.Context, txHash common.Hash) (*types.Transaction, bool, error) {
	ret := _m.Called(ctx, txHash)

	if len(ret) == 0 {
		panic("no return value specified for GetTx")
	}

	var r0 *types.Transaction
	var r1 bool
	var r2 error
	if rf, ok := ret.Get(0).(func(context.Context, common.Hash) (*types.Transaction, bool, error)); ok {
		return rf(ctx, txHash)
	}
	if rf, ok := ret.Get(0).(func(context.Context, common.Hash) *types.Transaction); ok {
		r0 = rf(ctx, txHash)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.Transaction)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, common.Hash) bool); ok {
		r1 = rf(ctx, txHash)
	} else {
		r1 = ret.Get(1).(bool)
	}

	if rf, ok := ret.Get(2).(func(context.Context, common.Hash) error); ok {
		r2 = rf(ctx, txHash)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// EthermanInterface_GetTx_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetTx'
type EthermanInterface_GetTx_Call struct {
	*mock.Call
}

// GetTx is a helper method to define mock.On call
//   - ctx context.Context
//   - txHash common.Hash
func (_e *EthermanInterface_Expecter) GetTx(ctx interface{}, txHash interface{}) *EthermanInterface_GetTx_Call {
	return &EthermanInterface_GetTx_Call{Call: _e.mock.On("GetTx", ctx, txHash)}
}

func (_c *EthermanInterface_GetTx_Call) Run(run func(ctx context.Context, txHash common.Hash)) *EthermanInterface_GetTx_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(common.Hash))
	})
	return _c
}

func (_c *EthermanInterface_GetTx_Call) Return(_a0 *types.Transaction, _a1 bool, _a2 error) *EthermanInterface_GetTx_Call {
	_c.Call.Return(_a0, _a1, _a2)
	return _c
}

func (_c *EthermanInterface_GetTx_Call) RunAndReturn(run func(context.Context, common.Hash) (*types.Transaction, bool, error)) *EthermanInterface_GetTx_Call {
	_c.Call.Return(run)
	return _c
}

// GetTxReceipt provides a mock function with given fields: ctx, txHash
func (_m *EthermanInterface) GetTxReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	ret := _m.Called(ctx, txHash)

	if len(ret) == 0 {
		panic("no return value specified for GetTxReceipt")
	}

	var r0 *types.Receipt
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, common.Hash) (*types.Receipt, error)); ok {
		return rf(ctx, txHash)
	}
	if rf, ok := ret.Get(0).(func(context.Context, common.Hash) *types.Receipt); ok {
		r0 = rf(ctx, txHash)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.Receipt)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, common.Hash) error); ok {
		r1 = rf(ctx, txHash)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// EthermanInterface_GetTxReceipt_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetTxReceipt'
type EthermanInterface_GetTxReceipt_Call struct {
	*mock.Call
}

// GetTxReceipt is a helper method to define mock.On call
//   - ctx context.Context
//   - txHash common.Hash
func (_e *EthermanInterface_Expecter) GetTxReceipt(ctx interface{}, txHash interface{}) *EthermanInterface_GetTxReceipt_Call {
	return &EthermanInterface_GetTxReceipt_Call{Call: _e.mock.On("GetTxReceipt", ctx, txHash)}
}

func (_c *EthermanInterface_GetTxReceipt_Call) Run(run func(ctx context.Context, txHash common.Hash)) *EthermanInterface_GetTxReceipt_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(common.Hash))
	})
	return _c
}

func (_c *EthermanInterface_GetTxReceipt_Call) Return(_a0 *types.Receipt, _a1 error) *EthermanInterface_GetTxReceipt_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *EthermanInterface_GetTxReceipt_Call) RunAndReturn(run func(context.Context, common.Hash) (*types.Receipt, error)) *EthermanInterface_GetTxReceipt_Call {
	_c.Call.Return(run)
	return _c
}

// HeaderByNumber provides a mock function with given fields: ctx, number
func (_m *EthermanInterface) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	ret := _m.Called(ctx, number)

	if len(ret) == 0 {
		panic("no return value specified for HeaderByNumber")
	}

	var r0 *types.Header
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *big.Int) (*types.Header, error)); ok {
		return rf(ctx, number)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *big.Int) *types.Header); ok {
		r0 = rf(ctx, number)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.Header)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *big.Int) error); ok {
		r1 = rf(ctx, number)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// EthermanInterface_HeaderByNumber_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'HeaderByNumber'
type EthermanInterface_HeaderByNumber_Call struct {
	*mock.Call
}

// HeaderByNumber is a helper method to define mock.On call
//   - ctx context.Context
//   - number *big.Int
func (_e *EthermanInterface_Expecter) HeaderByNumber(ctx interface{}, number interface{}) *EthermanInterface_HeaderByNumber_Call {
	return &EthermanInterface_HeaderByNumber_Call{Call: _e.mock.On("HeaderByNumber", ctx, number)}
}

func (_c *EthermanInterface_HeaderByNumber_Call) Run(run func(ctx context.Context, number *big.Int)) *EthermanInterface_HeaderByNumber_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*big.Int))
	})
	return _c
}

func (_c *EthermanInterface_HeaderByNumber_Call) Return(_a0 *types.Header, _a1 error) *EthermanInterface_HeaderByNumber_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *EthermanInterface_HeaderByNumber_Call) RunAndReturn(run func(context.Context, *big.Int) (*types.Header, error)) *EthermanInterface_HeaderByNumber_Call {
	_c.Call.Return(run)
	return _c
}

// PendingNonce provides a mock function with given fields: ctx, account
func (_m *EthermanInterface) PendingNonce(ctx context.Context, account common.Address) (uint64, error) {
	ret := _m.Called(ctx, account)

	if len(ret) == 0 {
		panic("no return value specified for PendingNonce")
	}

	var r0 uint64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, common.Address) (uint64, error)); ok {
		return rf(ctx, account)
	}
	if rf, ok := ret.Get(0).(func(context.Context, common.Address) uint64); ok {
		r0 = rf(ctx, account)
	} else {
		r0 = ret.Get(0).(uint64)
	}

	if rf, ok := ret.Get(1).(func(context.Context, common.Address) error); ok {
		r1 = rf(ctx, account)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// EthermanInterface_PendingNonce_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'PendingNonce'
type EthermanInterface_PendingNonce_Call struct {
	*mock.Call
}

// PendingNonce is a helper method to define mock.On call
//   - ctx context.Context
//   - account common.Address
func (_e *EthermanInterface_Expecter) PendingNonce(ctx interface{}, account interface{}) *EthermanInterface_PendingNonce_Call {
	return &EthermanInterface_PendingNonce_Call{Call: _e.mock.On("PendingNonce", ctx, account)}
}

func (_c *EthermanInterface_PendingNonce_Call) Run(run func(ctx context.Context, account common.Address)) *EthermanInterface_PendingNonce_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(common.Address))
	})
	return _c
}

func (_c *EthermanInterface_PendingNonce_Call) Return(_a0 uint64, _a1 error) *EthermanInterface_PendingNonce_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *EthermanInterface_PendingNonce_Call) RunAndReturn(run func(context.Context, common.Address) (uint64, error)) *EthermanInterface_PendingNonce_Call {
	_c.Call.Return(run)
	return _c
}

// SendTx provides a mock function with given fields: ctx, tx
func (_m *EthermanInterface) SendTx(ctx context.Context, tx *types.Transaction) error {
	ret := _m.Called(ctx, tx)

	if len(ret) == 0 {
		panic("no return value specified for SendTx")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *types.Transaction) error); ok {
		r0 = rf(ctx, tx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// EthermanInterface_SendTx_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SendTx'
type EthermanInterface_SendTx_Call struct {
	*mock.Call
}

// SendTx is a helper method to define mock.On call
//   - ctx context.Context
//   - tx *types.Transaction
func (_e *EthermanInterface_Expecter) SendTx(ctx interface{}, tx interface{}) *EthermanInterface_SendTx_Call {
	return &EthermanInterface_SendTx_Call{Call: _e.mock.On("SendTx", ctx, tx)}
}

func (_c *EthermanInterface_SendTx_Call) Run(run func(ctx context.Context, tx *types.Transaction)) *EthermanInterface_SendTx_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*types.Transaction))
	})
	return _c
}

func (_c *EthermanInterface_SendTx_Call) Return(_a0 error) *EthermanInterface_SendTx_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *EthermanInterface_SendTx_Call) RunAndReturn(run func(context.Context, *types.Transaction) error) *EthermanInterface_SendTx_Call {
	_c.Call.Return(run)
	return _c
}

// SignTx provides a mock function with given fields: ctx, sender, tx
func (_m *EthermanInterface) SignTx(ctx context.Context, sender common.Address, tx *types.Transaction) (*types.Transaction, error) {
	ret := _m.Called(ctx, sender, tx)

	if len(ret) == 0 {
		panic("no return value specified for SignTx")
	}

	var r0 *types.Transaction
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, common.Address, *types.Transaction) (*types.Transaction, error)); ok {
		return rf(ctx, sender, tx)
	}
	if rf, ok := ret.Get(0).(func(context.Context, common.Address, *types.Transaction) *types.Transaction); ok {
		r0 = rf(ctx, sender, tx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.Transaction)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, common.Address, *types.Transaction) error); ok {
		r1 = rf(ctx, sender, tx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// EthermanInterface_SignTx_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SignTx'
type EthermanInterface_SignTx_Call struct {
	*mock.Call
}

// SignTx is a helper method to define mock.On call
//   - ctx context.Context
//   - sender common.Address
//   - tx *types.Transaction
func (_e *EthermanInterface_Expecter) SignTx(ctx interface{}, sender interface{}, tx interface{}) *EthermanInterface_SignTx_Call {
	return &EthermanInterface_SignTx_Call{Call: _e.mock.On("SignTx", ctx, sender, tx)}
}

func (_c *EthermanInterface_SignTx_Call) Run(run func(ctx context.Context, sender common.Address, tx *types.Transaction)) *EthermanInterface_SignTx_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(common.Address), args[2].(*types.Transaction))
	})
	return _c
}

func (_c *EthermanInterface_SignTx_Call) Return(_a0 *types.Transaction, _a1 error) *EthermanInterface_SignTx_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *EthermanInterface_SignTx_Call) RunAndReturn(run func(context.Context, common.Address, *types.Transaction) (*types.Transaction, error)) *EthermanInterface_SignTx_Call {
	_c.Call.Return(run)
	return _c
}

// SuggestedGasPrice provides a mock function with given fields: ctx
func (_m *EthermanInterface) SuggestedGasPrice(ctx context.Context) (*big.Int, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for SuggestedGasPrice")
	}

	var r0 *big.Int
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (*big.Int, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) *big.Int); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*big.Int)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// EthermanInterface_SuggestedGasPrice_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SuggestedGasPrice'
type EthermanInterface_SuggestedGasPrice_Call struct {
	*mock.Call
}

// SuggestedGasPrice is a helper method to define mock.On call
//   - ctx context.Context
func (_e *EthermanInterface_Expecter) SuggestedGasPrice(ctx interface{}) *EthermanInterface_SuggestedGasPrice_Call {
	return &EthermanInterface_SuggestedGasPrice_Call{Call: _e.mock.On("SuggestedGasPrice", ctx)}
}

func (_c *EthermanInterface_SuggestedGasPrice_Call) Run(run func(ctx context.Context)) *EthermanInterface_SuggestedGasPrice_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *EthermanInterface_SuggestedGasPrice_Call) Return(_a0 *big.Int, _a1 error) *EthermanInterface_SuggestedGasPrice_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *EthermanInterface_SuggestedGasPrice_Call) RunAndReturn(run func(context.Context) (*big.Int, error)) *EthermanInterface_SuggestedGasPrice_Call {
	_c.Call.Return(run)
	return _c
}

// WaitTxToBeMined provides a mock function with given fields: ctx, tx, timeout
func (_m *EthermanInterface) WaitTxToBeMined(ctx context.Context, tx *types.Transaction, timeout time.Duration) (bool, error) {
	ret := _m.Called(ctx, tx, timeout)

	if len(ret) == 0 {
		panic("no return value specified for WaitTxToBeMined")
	}

	var r0 bool
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *types.Transaction, time.Duration) (bool, error)); ok {
		return rf(ctx, tx, timeout)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *types.Transaction, time.Duration) bool); ok {
		r0 = rf(ctx, tx, timeout)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(context.Context, *types.Transaction, time.Duration) error); ok {
		r1 = rf(ctx, tx, timeout)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// EthermanInterface_WaitTxToBeMined_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'WaitTxToBeMined'
type EthermanInterface_WaitTxToBeMined_Call struct {
	*mock.Call
}

// WaitTxToBeMined is a helper method to define mock.On call
//   - ctx context.Context
//   - tx *types.Transaction
//   - timeout time.Duration
func (_e *EthermanInterface_Expecter) WaitTxToBeMined(ctx interface{}, tx interface{}, timeout interface{}) *EthermanInterface_WaitTxToBeMined_Call {
	return &EthermanInterface_WaitTxToBeMined_Call{Call: _e.mock.On("WaitTxToBeMined", ctx, tx, timeout)}
}

func (_c *EthermanInterface_WaitTxToBeMined_Call) Run(run func(ctx context.Context, tx *types.Transaction, timeout time.Duration)) *EthermanInterface_WaitTxToBeMined_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*types.Transaction), args[2].(time.Duration))
	})
	return _c
}

func (_c *EthermanInterface_WaitTxToBeMined_Call) Return(_a0 bool, _a1 error) *EthermanInterface_WaitTxToBeMined_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *EthermanInterface_WaitTxToBeMined_Call) RunAndReturn(run func(context.Context, *types.Transaction, time.Duration) (bool, error)) *EthermanInterface_WaitTxToBeMined_Call {
	_c.Call.Return(run)
	return _c
}

// NewEthermanInterface creates a new instance of EthermanInterface. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewEthermanInterface(t interface {
	mock.TestingT
	Cleanup(func())
}) *EthermanInterface {
	mock := &EthermanInterface{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
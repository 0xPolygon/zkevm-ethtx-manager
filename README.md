# zkevm-ethtx-manager
Stateless manager to sent transactions to L1.

## Main Funtions
### Add Transaction
`func (c *Client) Add(ctx context.Context, to *common.Address, forcedNonce *uint64, value *big.Int, data []byte) (common.Hash, error)`

Adds a transaction to be sent to L1. The returned hash is calculated over the *to*, *nonce*, *value* and *data* fields.

Parameter forcedNonce is optional, if nil is passed the current nonce is obtained from the L1 node.

### Get Transaction Status
`func (c *Client) Result(ctx context.Context, id common.Hash) (MonitoredTxResult, error)`

Result returns the current result of the transaction execution with all the details.

### Get All Transactions Status
`func (c *Client) ResultsByStatus(ctx context.Context, statuses []MonitoredTxStatus) ([]MonitoredTxResult, error)`

ResultsByStatus returns all the results for all the monitored txs matching the provided statuses.
If the statuses are empty, all the statuses are considered.

### Pending L1 Transactions
`func (c *Client) PendingL1Txs(ctx context.Context) (bool, error)`

Returns whether there are pending transactions to be mined in L1.

### Transactions statuses

- **Created**: the tx was just added to the volatile storage
- **Sent**: transaction was sent to L1
- **Failed**: the tx was already mined and failed with an error that can't be recovered automatically, ex: the data in the tx is invalid and the tx gets reverted
- **Confirmed**: the tx was already mined and the receipt status is Successful. Confirmation waits for the configured number of blocks to be considered final.

#### Legacy statuses

- **Reorged**: is used when a monitored tx was already confirmed but the L1 block where this tx was confirmed has been reorged, in this situation the caller needs to review this information and wait until it gets confirmed again in a future block.
- **Done**: the tx was set by the owner as done

// VulcanizeDB
// Copyright © 2019 Vulcanize

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.

// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package client

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type EthClient struct {
	client *ethclient.Client
}

// NewEthClient return a new EthClient
func NewEthClient(client *ethclient.Client) EthClient {
	return EthClient{client: client}
}

// BlockByNumber fetches and returns the block for a given block number
func (client EthClient) BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error) {
	return client.client.BlockByNumber(ctx, number)
}

// CallContract calls the contract with the provided call msg and returns the hex byte value returned from the contract
func (client EthClient) CallContract(ctx context.Context, msg ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	return client.client.CallContract(ctx, msg, blockNumber)
}

// FilterLogs fetches the logs which conform to the provided filter query
func (client EthClient) FilterLogs(ctx context.Context, q ethereum.FilterQuery) ([]types.Log, error) {
	return client.client.FilterLogs(ctx, q)
}

// HeaderByNumber fetchers and returns the header for a given block number
func (client EthClient) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	return client.client.HeaderByNumber(ctx, number)
}

// TransactionSender asks the node to return the sender address for a provided transaction
func (client EthClient) TransactionSender(ctx context.Context, tx *types.Transaction, block common.Hash, index uint) (common.Address, error) {
	return client.client.TransactionSender(ctx, tx, block, index)
}

// TransactionReceipt fetches the receipt that corresponds with the provided tx hash
func (client EthClient) TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	return client.client.TransactionReceipt(ctx, txHash)
}

// BalanceAt fetches the account (eth) balance for the provided account address and block number
func (client EthClient) BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error) {
	return client.client.BalanceAt(ctx, account, blockNumber)
}

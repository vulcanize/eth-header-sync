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

package fetcher

import (
	"errors"
	"math/big"
	"strconv"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"

	"github.com/vulcanize/eth-header-sync/pkg/client"
	"github.com/vulcanize/eth-header-sync/pkg/converter"
	"github.com/vulcanize/eth-header-sync/pkg/core"
)

var ErrEmptyHeader = errors.New("empty header returned over RPC")

const MAX_BATCH_SIZE = 100

// Fetcher is the underlying type which satisfies the core.Fetcher interface for go-ethereum
type Fetcher struct {
	ethClient       core.EthClient
	headerConverter converter.HeaderConverter
	node            core.Node
	rpcClient       core.RPCClient
}

// NewFetcher returns a new Fetcher
func NewFetcher(ethClient core.EthClient, rpcClient core.RPCClient, node core.Node) *Fetcher {
	return &Fetcher{
		ethClient:       ethClient,
		headerConverter: converter.HeaderConverter{},
		node:            node,
		rpcClient:       rpcClient,
	}
}

// GetHeaderByNumber fetches the header for the provided block number
func (fetcher *Fetcher) GetHeaderByNumber(blockNumber int64) (header core.Header, err error) {
	logrus.Debugf("GetHeaderByNumber called with block %d", blockNumber)
	if fetcher.node.NetworkID == string(core.KOVAN_NETWORK_ID) {
		return fetcher.getPOAHeader(blockNumber)
	}
	return fetcher.getPOWHeader(blockNumber)
}

// GetHeadersByNumbers batch fetches all of the headers for the provided block numbers
func (fetcher *Fetcher) GetHeadersByNumbers(blockNumbers []int64) (header []core.Header, err error) {
	logrus.Debug("GetHeadersByNumbers called")
	if fetcher.node.NetworkID == string(core.KOVAN_NETWORK_ID) {
		return fetcher.getPOAHeaders(blockNumbers)
	}
	return fetcher.getPOWHeaders(blockNumbers)
}

// LastBlock determines and returns the latest block number
func (fetcher *Fetcher) LastBlock() (*big.Int, error) {
	block, err := fetcher.ethClient.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return big.NewInt(0), err
	}
	return block.Number, err
}

// Node returns the node info associated with this Fetcher
func (fetcher *Fetcher) Node() core.Node {
	return fetcher.node
}

func (fetcher *Fetcher) getPOAHeader(blockNumber int64) (header core.Header, err error) {
	var POAHeader core.POAHeader
	blockNumberArg := hexutil.EncodeBig(big.NewInt(blockNumber))
	includeTransactions := false
	err = fetcher.rpcClient.CallContext(context.Background(), &POAHeader, "eth_getBlockByNumber", blockNumberArg, includeTransactions)
	if err != nil {
		return header, err
	}
	if POAHeader.Number == nil {
		return header, ErrEmptyHeader
	}
	return fetcher.headerConverter.Convert(&types.Header{
		ParentHash:  POAHeader.ParentHash,
		UncleHash:   POAHeader.UncleHash,
		Coinbase:    POAHeader.Coinbase,
		Root:        POAHeader.Root,
		TxHash:      POAHeader.TxHash,
		ReceiptHash: POAHeader.ReceiptHash,
		Bloom:       POAHeader.Bloom,
		Difficulty:  POAHeader.Difficulty.ToInt(),
		Number:      POAHeader.Number.ToInt(),
		GasLimit:    uint64(POAHeader.GasLimit),
		GasUsed:     uint64(POAHeader.GasUsed),
		Time:        uint64(POAHeader.Time),
		Extra:       POAHeader.Extra,
	}, POAHeader.Hash.String()), nil
}

func (blockChain *Fetcher) getPOAHeaders(blockNumbers []int64) (headers []core.Header, err error) {

	var batch []client.BatchElem
	var POAHeaders [MAX_BATCH_SIZE]core.POAHeader
	includeTransactions := false

	for index, blockNumber := range blockNumbers {

		if index >= MAX_BATCH_SIZE {
			break
		}

		blockNumberArg := hexutil.EncodeBig(big.NewInt(blockNumber))

		batchElem := client.BatchElem{
			Method: "eth_getBlockByNumber",
			Result: &POAHeaders[index],
			Args:   []interface{}{blockNumberArg, includeTransactions},
		}

		batch = append(batch, batchElem)
	}

	err = blockChain.rpcClient.BatchCall(batch)
	if err != nil {
		return headers, err
	}

	for _, POAHeader := range POAHeaders {
		var header core.Header
		//Header.Number of the newest block will return nil.
		if _, err := strconv.ParseUint(POAHeader.Number.ToInt().String(), 16, 64); err == nil {
			header = blockChain.headerConverter.Convert(&types.Header{
				ParentHash:  POAHeader.ParentHash,
				UncleHash:   POAHeader.UncleHash,
				Coinbase:    POAHeader.Coinbase,
				Root:        POAHeader.Root,
				TxHash:      POAHeader.TxHash,
				ReceiptHash: POAHeader.ReceiptHash,
				Bloom:       POAHeader.Bloom,
				Difficulty:  POAHeader.Difficulty.ToInt(),
				Number:      POAHeader.Number.ToInt(),
				GasLimit:    uint64(POAHeader.GasLimit),
				GasUsed:     uint64(POAHeader.GasUsed),
				Time:        uint64(POAHeader.Time),
				Extra:       POAHeader.Extra,
			}, POAHeader.Hash.String())

			headers = append(headers, header)
		}
	}

	return headers, err
}

func (fetcher *Fetcher) getPOWHeader(blockNumber int64) (header core.Header, err error) {
	gethHeader, err := fetcher.ethClient.HeaderByNumber(context.Background(), big.NewInt(blockNumber))
	if err != nil {
		return header, err
	}
	return fetcher.headerConverter.Convert(gethHeader, gethHeader.Hash().String()), nil
}

func (blockChain *Fetcher) getPOWHeaders(blockNumbers []int64) (headers []core.Header, err error) {
	var batch []client.BatchElem
	var POWHeaders [MAX_BATCH_SIZE]types.Header
	includeTransactions := false

	for index, blockNumber := range blockNumbers {

		if index >= MAX_BATCH_SIZE {
			break
		}

		blockNumberArg := hexutil.EncodeBig(big.NewInt(blockNumber))

		batchElem := client.BatchElem{
			Method: "eth_getBlockByNumber",
			Result: &POWHeaders[index],
			Args:   []interface{}{blockNumberArg, includeTransactions},
		}

		batch = append(batch, batchElem)
	}

	err = blockChain.rpcClient.BatchCall(batch)
	if err != nil {
		return headers, err
	}

	for _, POWHeader := range POWHeaders {
		if POWHeader.Number != nil {
			header := blockChain.headerConverter.Convert(&POWHeader, POWHeader.Hash().String())
			headers = append(headers, header)
		}
	}

	return headers, err
}

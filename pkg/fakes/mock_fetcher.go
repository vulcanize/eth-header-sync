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

package fakes

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
	. "github.com/onsi/gomega"

	"github.com/vulcanize/eth-header-sync/pkg/core"
)

type MockFetcher struct {
	fetchContractDataErr               error
	fetchContractDataPassedAbi         string
	fetchContractDataPassedAddress     string
	fetchContractDataPassedMethod      string
	fetchContractDataPassedMethodArgs  []interface{}
	fetchContractDataPassedResult      interface{}
	fetchContractDataPassedBlockNumber int64
	getBlockByNumberErr                error
	GetTransactionsCalled              bool
	GetTransactionsError               error
	GetTransactionsPassedHashes        []common.Hash
	logQuery                           ethereum.FilterQuery
	logQueryErr                        error
	logQueryReturnLogs                 []types.Log
	lastBlock                          *big.Int
	node                               core.Node
	accountBalanceReturnValue          *big.Int
	getAccountBalanceErr               error
}

func NewMockFetcher() *MockFetcher {
	return &MockFetcher{
		node: core.Node{GenesisBlock: "GENESIS", NetworkID: "1", ID: "x123", ClientName: "Geth"},
	}
}

func (fetcher *MockFetcher) SetFetchContractDataErr(err error) {
	fetcher.fetchContractDataErr = err
}

func (fetcher *MockFetcher) SetLastBlock(blockNumber *big.Int) {
	fetcher.lastBlock = blockNumber
}

func (fetcher *MockFetcher) SetGetBlockByNumberErr(err error) {
	fetcher.getBlockByNumberErr = err
}

func (fetcher *MockFetcher) SetGetEthLogsWithCustomQueryErr(err error) {
	fetcher.logQueryErr = err
}

func (fetcher *MockFetcher) SetGetEthLogsWithCustomQueryReturnLogs(logs []types.Log) {
	fetcher.logQueryReturnLogs = logs
}

func (fetcher *MockFetcher) FetchContractData(abiJSON string, address string, method string, methodArgs []interface{}, result interface{}, blockNumber int64) error {
	fetcher.fetchContractDataPassedAbi = abiJSON
	fetcher.fetchContractDataPassedAddress = address
	fetcher.fetchContractDataPassedMethod = method
	fetcher.fetchContractDataPassedMethodArgs = methodArgs
	fetcher.fetchContractDataPassedResult = result
	fetcher.fetchContractDataPassedBlockNumber = blockNumber
	return fetcher.fetchContractDataErr
}

func (fetcher *MockFetcher) GetEthLogsWithCustomQuery(query ethereum.FilterQuery) ([]types.Log, error) {
	fetcher.logQuery = query
	return fetcher.logQueryReturnLogs, fetcher.logQueryErr
}

func (fetcher *MockFetcher) GetHeaderByNumber(blockNumber int64) (core.Header, error) {
	return core.Header{BlockNumber: blockNumber}, nil
}

func (fetcher *MockFetcher) GetHeadersByNumbers(blockNumbers []int64) ([]core.Header, error) {
	var headers []core.Header
	for _, blockNumber := range blockNumbers {
		var header = core.Header{BlockNumber: int64(blockNumber)}
		headers = append(headers, header)
	}
	return headers, nil
}

func (fetcher *MockFetcher) CallContract(contractHash string, input []byte, blockNumber *big.Int) ([]byte, error) {
	return []byte{}, nil
}

func (fetcher *MockFetcher) LastBlock() (*big.Int, error) {
	return fetcher.lastBlock, nil
}

func (fetcher *MockFetcher) Node() core.Node {
	return fetcher.node
}

func (fetcher *MockFetcher) AssertFetchContractDataCalledWith(abiJSON string, address string, method string, methodArgs []interface{}, result interface{}, blockNumber int64) {
	Expect(fetcher.fetchContractDataPassedAbi).To(Equal(abiJSON))
	Expect(fetcher.fetchContractDataPassedAddress).To(Equal(address))
	Expect(fetcher.fetchContractDataPassedMethod).To(Equal(method))
	if methodArgs != nil {
		Expect(fetcher.fetchContractDataPassedMethodArgs).To(Equal(methodArgs))
	}
	Expect(fetcher.fetchContractDataPassedResult).To(BeAssignableToTypeOf(result))
	Expect(fetcher.fetchContractDataPassedBlockNumber).To(Equal(blockNumber))
}

func (fetcher *MockFetcher) AssertGetEthLogsWithCustomQueryCalledWith(query ethereum.FilterQuery) {
	Expect(fetcher.logQuery).To(Equal(query))
}

func (fetcher *MockFetcher) SetGetAccountBalanceErr(err error) {
	fetcher.getAccountBalanceErr = err
}

func (fetcher *MockFetcher) SetGetAccountBalance(balance *big.Int) {
	fetcher.accountBalanceReturnValue = balance
}

func (fetcher *MockFetcher) GetAccountBalance(address common.Address, blockNumber *big.Int) (*big.Int, error) {
	return fetcher.accountBalanceReturnValue, fetcher.getAccountBalanceErr
}

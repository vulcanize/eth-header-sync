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

	"github.com/vulcanize/eth-header-sync/pkg/core"
)

type MockFetcher struct {
	getBlockByNumberErr error
	lastBlock           *big.Int
	node                core.Node
}

func NewMockFetcher() *MockFetcher {
	return &MockFetcher{
		node: core.Node{GenesisBlock: "GENESIS", NetworkID: "1", ID: "x123", ClientName: "Geth"},
	}
}

func (fetcher *MockFetcher) SetLastBlock(blockNumber *big.Int) {
	fetcher.lastBlock = blockNumber
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

func (fetcher *MockFetcher) LastBlock() (*big.Int, error) {
	return fetcher.lastBlock, nil
}

func (fetcher *MockFetcher) Node() core.Node {
	return fetcher.node
}

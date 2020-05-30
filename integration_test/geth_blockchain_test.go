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

package integration_test

import (
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vulcanize/eth-header-sync/pkg/client"
	"github.com/vulcanize/eth-header-sync/pkg/core"
	"github.com/vulcanize/eth-header-sync/pkg/fetcher"
	"github.com/vulcanize/eth-header-sync/pkg/node"
	"github.com/vulcanize/eth-header-sync/test_config"
)

var _ = Describe("Reading from the Geth blockchain", func() {
	var fetch *fetcher.Fetcher

	BeforeEach(func() {
		rawRPCClient, err := rpc.Dial(test_config.TestClient.IPCPath)
		Expect(err).NotTo(HaveOccurred())
		rpcClient := client.NewRPCClient(rawRPCClient, test_config.TestClient.IPCPath)
		ethClient := ethclient.NewClient(rawRPCClient)
		blockChainClient := client.NewEthClient(ethClient)
		node := node.MakeNode(rpcClient)
		fetch = fetcher.NewFetcher(blockChainClient, rpcClient, node)
	})

	It("retrieves the genesis header and first header", func(done Done) {
		genesisBlock, err := fetch.GetHeaderByNumber(int64(0))
		Expect(err).ToNot(HaveOccurred())
		firstBlock, err := fetch.GetHeaderByNumber(int64(1))
		Expect(err).ToNot(HaveOccurred())
		lastBlockNumber, err := fetch.LastBlock()

		Expect(err).NotTo(HaveOccurred())
		Expect(genesisBlock.BlockNumber).To(Equal(int64(0)))
		Expect(firstBlock.BlockNumber).To(Equal(int64(1)))
		Expect(lastBlockNumber.Int64()).To(BeNumerically(">", 0))
		close(done)
	}, 15)

	It("retrieves the node info", func(done Done) {
		node := fetch.Node()

		Expect(node.GenesisBlock).ToNot(BeNil())
		Expect(node.NetworkID).To(Equal("1.000000"))
		Expect(len(node.ID)).ToNot(BeZero())
		Expect(node.ClientName).ToNot(BeZero())

		close(done)
	}, 15)

	//Benchmarking test: remove skip to test performance of block retrieval
	XMeasure("retrieving n headers", func(b Benchmarker) {
		b.Time("runtime", func() {
			var headers []core.Header
			n := 10
			for i := 5327459; i > 5327459-n; i-- {
				header, err := fetch.GetHeaderByNumber(int64(i))
				Expect(err).ToNot(HaveOccurred())
				headers = append(headers, header)
			}
			Expect(len(headers)).To(Equal(n))
		})
	}, 10)
})

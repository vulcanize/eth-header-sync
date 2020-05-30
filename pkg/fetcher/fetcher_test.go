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

package fetcher_test

import (
	"context"
	"math/big"

	"github.com/vulcanize/eth-header-sync/pkg/fetcher"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	vulcCore "github.com/vulcanize/eth-header-sync/pkg/core"
	"github.com/vulcanize/eth-header-sync/pkg/fakes"
)

var _ = Describe("Geth blockchain", func() {
	var (
		mockClient    *fakes.MockEthClient
		fetch         *fetcher.Fetcher
		mockRpcClient *fakes.MockRPCClient
		node          vulcCore.Node
	)

	BeforeEach(func() {
		mockClient = fakes.NewMockEthClient()
		mockRpcClient = fakes.NewMockRPCClient()
		node = vulcCore.Node{}
		fetch = fetcher.NewFetcher(mockClient, mockRpcClient, node)
	})

	Describe("getting a header", func() {
		Describe("default/mainnet", func() {
			It("fetches header from ethClient", func() {
				blockNumber := int64(100)
				mockClient.SetHeaderByNumberReturnHeader(&types.Header{Number: big.NewInt(blockNumber)})

				_, err := fetch.GetHeaderByNumber(blockNumber)

				Expect(err).NotTo(HaveOccurred())
				mockClient.AssertHeaderByNumberCalledWith(context.Background(), big.NewInt(blockNumber))
			})

			It("returns err if ethClient returns err", func() {
				mockClient.SetHeaderByNumberErr(fakes.FakeError)

				_, err := fetch.GetHeaderByNumber(100)

				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(fakes.FakeError))
			})

			It("fetches headers with multiple blocks", func() {
				_, err := fetch.GetHeadersByNumbers([]int64{100, 99})

				Expect(err).NotTo(HaveOccurred())
				mockRpcClient.AssertBatchCalledWith("eth_getBlockByNumber", 2)
			})
		})

		Describe("POA/Kovan", func() {
			It("fetches header from rpcClient", func() {
				node.NetworkID = string(vulcCore.KOVAN_NETWORK_ID)
				blockNumber := hexutil.Big(*big.NewInt(100))
				mockRpcClient.SetReturnPOAHeader(vulcCore.POAHeader{Number: &blockNumber})
				fetch = fetcher.NewFetcher(mockClient, mockRpcClient, node)

				_, err := fetch.GetHeaderByNumber(100)

				Expect(err).NotTo(HaveOccurred())
				mockRpcClient.AssertCallContextCalledWith(context.Background(), &vulcCore.POAHeader{}, "eth_getBlockByNumber")
			})

			It("returns err if rpcClient returns err", func() {
				node.NetworkID = string(vulcCore.KOVAN_NETWORK_ID)
				mockRpcClient.SetCallContextErr(fakes.FakeError)
				fetch = fetcher.NewFetcher(mockClient, mockRpcClient, node)

				_, err := fetch.GetHeaderByNumber(100)

				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(fakes.FakeError))
			})

			It("returns error if returned header is empty", func() {
				node.NetworkID = string(vulcCore.KOVAN_NETWORK_ID)
				fetch = fetcher.NewFetcher(mockClient, mockRpcClient, node)

				_, err := fetch.GetHeaderByNumber(100)

				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(fetcher.ErrEmptyHeader))
			})

			It("returns multiple headers with multiple blocknumbers", func() {
				node.NetworkID = string(vulcCore.KOVAN_NETWORK_ID)
				blockNumber := hexutil.Big(*big.NewInt(100))
				mockRpcClient.SetReturnPOAHeaders([]vulcCore.POAHeader{{Number: &blockNumber}})

				_, err := fetch.GetHeadersByNumbers([]int64{100, 99})

				Expect(err).NotTo(HaveOccurred())
				mockRpcClient.AssertBatchCalledWith("eth_getBlockByNumber", 2)
			})
		})
	})

	Describe("getting the most recent block number", func() {
		It("fetches latest header from ethClient", func() {
			blockNumber := int64(100)
			mockClient.SetHeaderByNumberReturnHeader(&types.Header{Number: big.NewInt(blockNumber)})

			result, err := fetch.LastBlock()
			Expect(err).NotTo(HaveOccurred())

			mockClient.AssertHeaderByNumberCalledWith(context.Background(), nil)
			Expect(result).To(Equal(big.NewInt(blockNumber)))
		})
	})
})

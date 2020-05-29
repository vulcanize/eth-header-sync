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

package common_test

import (
	"bytes"
	"io/ioutil"
	"log"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	vulcCommon "github.com/vulcanize/eth-header-sync/pkg/eth/converters/common"
	"github.com/vulcanize/eth-header-sync/pkg/eth/converters/rpc"
	"github.com/vulcanize/eth-header-sync/pkg/eth/fakes"
)

var _ = Describe("Conversion of GethBlock to core.Block", func() {

	It("converts basic Block metadata", func() {
		difficulty := big.NewInt(1)
		gasLimit := uint64(100000)
		gasUsed := uint64(100000)
		miner := common.HexToAddress("0x0000000000000000000000000000000000000123")
		extraData, _ := hexutil.Decode("0xe4b883e5bda9e7a59ee4bb99e9b1bc")
		nonce := types.BlockNonce{10}
		number := int64(1)
		time := uint64(140000000)

		header := types.Header{
			Difficulty: difficulty,
			GasLimit:   uint64(gasLimit),
			GasUsed:    uint64(gasUsed),
			Extra:      extraData,
			Coinbase:   miner,
			Nonce:      nonce,
			Number:     big.NewInt(number),
			ParentHash: common.Hash{64},
			Time:       time,
			UncleHash:  common.Hash{128},
		}
		block := types.NewBlock(&header, []*types.Transaction{}, []*types.Header{}, []*types.Receipt{})
		client := fakes.NewMockEthClient()
		transactionConverter := rpc.NewRPCTransactionConverter(client)
		blockConverter := vulcCommon.NewBlockConverter(transactionConverter)

		coreBlock, err := blockConverter.ToCoreBlock(block)

		Expect(err).ToNot(HaveOccurred())
		Expect(coreBlock.Difficulty).To(Equal(difficulty.Int64()))
		Expect(coreBlock.GasLimit).To(Equal(gasLimit))
		Expect(coreBlock.Miner).To(Equal(miner.Hex()))
		Expect(coreBlock.GasUsed).To(Equal(gasUsed))
		Expect(coreBlock.Hash).To(Equal(block.Hash().Hex()))
		Expect(coreBlock.Nonce).To(Equal(hexutil.Encode(header.Nonce[:])))
		Expect(coreBlock.Number).To(Equal(number))
		Expect(coreBlock.ParentHash).To(Equal(block.ParentHash().Hex()))
		Expect(coreBlock.ExtraData).To(Equal(hexutil.Encode(block.Extra())))
		Expect(coreBlock.Size).To(Equal(block.Size().String()))
		Expect(coreBlock.Time).To(Equal(time))
		Expect(coreBlock.UncleHash).To(Equal(block.UncleHash().Hex()))
		Expect(coreBlock.IsFinal).To(BeFalse())
	})

	Describe("The block and uncle rewards calculations", func() {
		It("calculates block rewards for a block", func() {
			transaction := types.NewTransaction(
				uint64(226823),
				common.HexToAddress("0x108fedb097c1dcfed441480170144d8e19bb217f"),
				big.NewInt(1080900090000000000),
				uint64(90000),
				big.NewInt(50000000000),
				[]byte{},
			)
			transactions := []*types.Transaction{transaction}

			txHash := transaction.Hash()
			receipt := types.Receipt{
				TxHash:            txHash,
				GasUsed:           uint64(21000),
				CumulativeGasUsed: uint64(21000),
			}
			receipts := []*types.Receipt{&receipt}

			client := fakes.NewMockEthClient()
			client.SetTransactionReceipts(receipts)

			number := int64(1071819)
			header := types.Header{
				Number: big.NewInt(number),
			}
			uncles := []*types.Header{{Number: big.NewInt(1071817)}, {Number: big.NewInt(1071818)}}
			block := types.NewBlock(&header, transactions, uncles, []*types.Receipt{&receipt})
			transactionConverter := rpc.NewRPCTransactionConverter(client)
			blockConverter := vulcCommon.NewBlockConverter(transactionConverter)

			coreBlock, err := blockConverter.ToCoreBlock(block)
			Expect(err).ToNot(HaveOccurred())

			expectedBlockReward := new(big.Int)
			expectedBlockReward.SetString("5313550000000000000", 10)

			blockReward := vulcCommon.CalcBlockReward(coreBlock, block.Uncles())
			Expect(blockReward.String()).To(Equal(expectedBlockReward.String()))
		})

		It("calculates the uncles reward for a block", func() {
			transaction := types.NewTransaction(
				uint64(226823),
				common.HexToAddress("0x108fedb097c1dcfed441480170144d8e19bb217f"),
				big.NewInt(1080900090000000000),
				uint64(90000),
				big.NewInt(50000000000),
				[]byte{})
			transactions := []*types.Transaction{transaction}

			receipt := types.Receipt{
				TxHash:            transaction.Hash(),
				GasUsed:           uint64(21000),
				CumulativeGasUsed: uint64(21000),
			}
			receipts := []*types.Receipt{&receipt}

			header := types.Header{
				Number: big.NewInt(int64(1071819)),
			}
			uncles := []*types.Header{
				{Number: big.NewInt(1071816)},
				{Number: big.NewInt(1071817)},
			}
			block := types.NewBlock(&header, transactions, uncles, receipts)

			client := fakes.NewMockEthClient()
			client.SetTransactionReceipts(receipts)
			transactionConverter := rpc.NewRPCTransactionConverter(client)
			blockConverter := vulcCommon.NewBlockConverter(transactionConverter)

			coreBlock, err := blockConverter.ToCoreBlock(block)
			Expect(err).ToNot(HaveOccurred())

			expectedTotalReward := new(big.Int)
			expectedTotalReward.SetString("6875000000000000000", 10)
			totalReward, coreUncles := blockConverter.ToCoreUncle(coreBlock, block.Uncles())
			Expect(totalReward.String()).To(Equal(expectedTotalReward.String()))

			Expect(len(coreUncles)).To(Equal(2))
			Expect(coreUncles[0].Reward).To(Equal("3125000000000000000"))
			Expect(coreUncles[0].Miner).To(Equal("0x0000000000000000000000000000000000000000"))
			Expect(coreUncles[0].Hash).To(Equal("0xb629de4014b6e30cf9555ee833f1806fa0d8b8516fde194405f9c98c2deb8772"))
			Expect(coreUncles[1].Reward).To(Equal("3750000000000000000"))
			Expect(coreUncles[1].Miner).To(Equal("0x0000000000000000000000000000000000000000"))
			Expect(coreUncles[1].Hash).To(Equal("0x673f5231e4888a951e0bc8a25b5774b982e6e9e258362c21affaff6e02dd5a2b"))
		})

		It("decreases the static block reward from 5 to 3 for blocks after block 4,269,999", func() {
			transactionOne := types.NewTransaction(
				uint64(8072),
				common.HexToAddress("0xebd17720aeb7ac5186c5dfa7bafeb0bb14c02551 "),
				big.NewInt(0),
				uint64(500000),
				big.NewInt(42000000000),
				[]byte{},
			)

			transactionTwo := types.NewTransaction(uint64(8071),
				common.HexToAddress("0x3cdab63d764c8c5048ed5e8f0a4e95534ba7e1ea"),
				big.NewInt(0),
				uint64(500000),
				big.NewInt(42000000000),
				[]byte{})

			transactions := []*types.Transaction{transactionOne, transactionTwo}

			receiptOne := types.Receipt{
				TxHash:            transactionOne.Hash(),
				GasUsed:           uint64(297508),
				CumulativeGasUsed: uint64(0),
			}
			receiptTwo := types.Receipt{
				TxHash:            transactionTwo.Hash(),
				GasUsed:           uint64(297508),
				CumulativeGasUsed: uint64(0),
			}
			receipts := []*types.Receipt{&receiptOne, &receiptTwo}

			number := int64(4370055)
			header := types.Header{
				Number: big.NewInt(number),
			}
			var uncles []*types.Header
			block := types.NewBlock(&header, transactions, uncles, receipts)

			client := fakes.NewMockEthClient()
			client.SetTransactionReceipts(receipts)
			transactionConverter := rpc.NewRPCTransactionConverter(client)
			blockConverter := vulcCommon.NewBlockConverter(transactionConverter)

			coreBlock, err := blockConverter.ToCoreBlock(block)
			Expect(err).ToNot(HaveOccurred())

			expectedRewards := new(big.Int)
			expectedRewards.SetString("3024990672000000000", 10)
			rewards := vulcCommon.CalcBlockReward(coreBlock, block.Uncles())
			Expect(rewards.String()).To(Equal(expectedRewards.String()))
		})
	})

	Describe("the converted transactions", func() {
		It("is empty", func() {
			header := types.Header{}
			block := types.NewBlock(&header, []*types.Transaction{}, []*types.Header{}, []*types.Receipt{})
			client := fakes.NewMockEthClient()
			transactionConverter := rpc.NewRPCTransactionConverter(client)
			blockConverter := vulcCommon.NewBlockConverter(transactionConverter)

			coreBlock, err := blockConverter.ToCoreBlock(block)

			Expect(err).ToNot(HaveOccurred())
			Expect(len(coreBlock.Transactions)).To(Equal(0))
		})

		It("converts a single transaction", func() {
			gethTransaction := types.NewTransaction(
				uint64(10000), common.Address{1},
				big.NewInt(10),
				uint64(5000),
				big.NewInt(3),
				hexutil.MustDecode("0xf7d8c8830000000000000000000000000000000000000000000000000000000000037788000000000000000000000000000000000000000000000000000000000003bd14"),
			)
			var rawTransaction bytes.Buffer
			encodeErr := gethTransaction.EncodeRLP(&rawTransaction)
			Expect(encodeErr).NotTo(HaveOccurred())

			gethReceipt := &types.Receipt{
				Bloom:             types.BytesToBloom(hexutil.MustDecode("0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")),
				ContractAddress:   common.HexToAddress("x123"),
				CumulativeGasUsed: uint64(7996119),
				GasUsed:           uint64(21000),
				Logs:              []*types.Log{},
				Status:            uint64(1),
				TxHash:            gethTransaction.Hash(),
			}

			client := fakes.NewMockEthClient()
			client.SetTransactionReceipts([]*types.Receipt{gethReceipt})

			header := types.Header{}
			block := types.NewBlock(
				&header,
				[]*types.Transaction{gethTransaction},
				[]*types.Header{},
				[]*types.Receipt{gethReceipt},
			)
			transactionConverter := rpc.NewRPCTransactionConverter(client)
			blockConverter := vulcCommon.NewBlockConverter(transactionConverter)

			coreBlock, err := blockConverter.ToCoreBlock(block)

			Expect(err).ToNot(HaveOccurred())
			Expect(len(coreBlock.Transactions)).To(Equal(1))
			coreTransaction := coreBlock.Transactions[0]
			expectedData := common.FromHex("0xf7d8c8830000000000000000000000000000000000000000000000000000000000037788000000000000000000000000000000000000000000000000000000000003bd14")
			Expect(coreTransaction.Data).To(Equal(expectedData))
			Expect(coreTransaction.To).To(Equal(gethTransaction.To().Hex()))
			Expect(coreTransaction.From).To(Equal("0x0000000000000000000000000000000000000123"))
			Expect(coreTransaction.GasLimit).To(Equal(gethTransaction.Gas()))
			Expect(coreTransaction.GasPrice).To(Equal(gethTransaction.GasPrice().Int64()))
			Expect(coreTransaction.Raw).To(Equal(rawTransaction.Bytes()))
			Expect(coreTransaction.TxIndex).To(Equal(int64(0)))
			Expect(coreTransaction.Value).To(Equal(gethTransaction.Value().String()))
			Expect(coreTransaction.Nonce).To(Equal(gethTransaction.Nonce()))

			coreReceipt := coreTransaction.Receipt
			expectedReceipt, err := vulcCommon.ToCoreReceipt(gethReceipt)
			Expect(err).ToNot(HaveOccurred())
			Expect(coreReceipt).To(Equal(expectedReceipt))
		})

		It("has an empty 'To' field when transaction creates a new contract", func() {
			gethTransaction := types.NewContractCreation(
				uint64(10000),
				big.NewInt(10),
				uint64(5000),
				big.NewInt(3),
				[]byte("1234"),
			)

			gethReceipt := &types.Receipt{
				CumulativeGasUsed: uint64(1),
				GasUsed:           uint64(1),
				TxHash:            gethTransaction.Hash(),
				ContractAddress:   common.HexToAddress("0x1023342345"),
			}

			client := fakes.NewMockEthClient()
			client.SetTransactionReceipts([]*types.Receipt{gethReceipt})

			block := types.NewBlock(
				&types.Header{},
				[]*types.Transaction{gethTransaction},
				[]*types.Header{},
				[]*types.Receipt{gethReceipt},
			)
			transactionConverter := rpc.NewRPCTransactionConverter(client)
			blockConverter := vulcCommon.NewBlockConverter(transactionConverter)

			coreBlock, err := blockConverter.ToCoreBlock(block)

			Expect(err).ToNot(HaveOccurred())
			coreTransaction := coreBlock.Transactions[0]
			Expect(coreTransaction.To).To(Equal(""))

			coreReceipt := coreTransaction.Receipt
			expectedReceipt, err := vulcCommon.ToCoreReceipt(gethReceipt)
			Expect(err).ToNot(HaveOccurred())
			Expect(coreReceipt).To(Equal(expectedReceipt))
		})
	})

	Describe("transaction error handling", func() {
		var gethTransaction *types.Transaction
		var gethReceipt *types.Receipt
		var header *types.Header
		var block *types.Block

		BeforeEach(func() {
			log.SetOutput(ioutil.Discard)
			gethTransaction = types.NewTransaction(
				uint64(0),
				common.Address{},
				big.NewInt(0),
				uint64(0),
				big.NewInt(0),
				[]byte{},
			)
			gethReceipt = &types.Receipt{}
			header = &types.Header{}
			block = types.NewBlock(
				header,
				[]*types.Transaction{gethTransaction},
				[]*types.Header{},
				[]*types.Receipt{gethReceipt},
			)

		})

		AfterEach(func() {
			defer log.SetOutput(os.Stdout)
		})

		It("returns an error when transaction sender call fails", func() {
			client := fakes.NewMockEthClient()
			client.SetTransactionSenderErr(fakes.FakeError)
			transactionConverter := rpc.NewRPCTransactionConverter(client)
			blockConverter := vulcCommon.NewBlockConverter(transactionConverter)

			_, err := blockConverter.ToCoreBlock(block)

			Expect(err).To(MatchError(fakes.FakeError))
		})

		It("returns an error when transaction receipt call fails", func() {
			client := fakes.NewMockEthClient()
			client.SetTransactionReceiptErr(fakes.FakeError)
			transactionConverter := rpc.NewRPCTransactionConverter(client)
			blockConverter := vulcCommon.NewBlockConverter(transactionConverter)

			_, err := blockConverter.ToCoreBlock(block)

			Expect(err).To(MatchError(fakes.FakeError))
		})
	})

})

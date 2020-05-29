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
	"bytes"
	"encoding/json"
	"errors"
	"math/rand"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/vulcanize/eth-header-sync/pkg/eth/core"
)

var (
	FakeAddress   = common.HexToAddress("0x" + RandomString(40))
	FakeError     = errors.New("failed")
	FakeHash      = common.BytesToHash([]byte{1, 2, 3, 4, 5})
	fakeTimestamp = rand.Int63n(1500000000)
)

var rawFakeHeader, _ = json.Marshal(types.Header{})
var FakeHeader = core.Header{
	Hash:      FakeHash.String(),
	Raw:       rawFakeHeader,
	Timestamp: strconv.FormatInt(fakeTimestamp, 10),
}

func GetFakeHeader(blockNumber int64) core.Header {
	return GetFakeHeaderWithTimestamp(fakeTimestamp, blockNumber)
}

func GetFakeHeaderWithTimestamp(timestamp, blockNumber int64) core.Header {
	return core.Header{
		Hash:        FakeHash.String(),
		BlockNumber: blockNumber,
		Raw:         rawFakeHeader,
		Timestamp:   strconv.FormatInt(timestamp, 10),
	}
}

var fakeTransaction types.Transaction
var rawTransaction bytes.Buffer
var _ = fakeTransaction.EncodeRLP(&rawTransaction)
var FakeTransaction = core.TransactionModel{
	Data:     []byte{},
	From:     "",
	GasLimit: 0,
	GasPrice: 0,
	Hash:     "",
	Nonce:    0,
	Raw:      rawTransaction.Bytes(),
	Receipt:  core.Receipt{},
	To:       "",
	TxIndex:  0,
	Value:    "0",
}

func GetFakeTransaction(hash string, receipt core.Receipt) core.TransactionModel {
	gethTransaction := types.Transaction{}
	var raw bytes.Buffer
	err := gethTransaction.EncodeRLP(&raw)
	if err != nil {
		panic("failed to marshal transaction while creating test fake")
	}
	return core.TransactionModel{
		Data:     []byte{},
		From:     "",
		GasLimit: 0,
		GasPrice: 0,
		Hash:     hash,
		Nonce:    0,
		Raw:      raw.Bytes(),
		Receipt:  receipt,
		To:       "",
		TxIndex:  0,
		Value:    "0",
	}
}

func GetFakeUncle(hash, reward string) core.Uncle {
	return core.Uncle{
		Miner:     FakeAddress.String(),
		Hash:      hash,
		Reward:    reward,
		Raw:       rawFakeHeader,
		Timestamp: strconv.FormatInt(fakeTimestamp, 10),
	}
}

func RandomString(length int) string {
	var seededRand = rand.New(
		rand.NewSource(time.Now().UnixNano()))
	charset := "abcdef1234567890"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}

	return string(b)
}

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
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/vulcanize/eth-header-sync/pkg/eth/core"
)

type MockTransactionConverter struct {
	ConvertHeaderTransactionIndexToIntCalled  bool
	ConvertBlockTransactionsToCoreCalled      bool
	ConvertBlockTransactionsToCorePassedBlock *types.Block
}

func NewMockTransactionConverter() *MockTransactionConverter {
	return &MockTransactionConverter{
		ConvertHeaderTransactionIndexToIntCalled:  false,
		ConvertBlockTransactionsToCoreCalled:      false,
		ConvertBlockTransactionsToCorePassedBlock: nil,
	}
}

func (converter *MockTransactionConverter) ConvertBlockTransactionsToCore(gethBlock *types.Block) ([]core.TransactionModel, error) {
	converter.ConvertBlockTransactionsToCoreCalled = true
	converter.ConvertBlockTransactionsToCorePassedBlock = gethBlock
	return []core.TransactionModel{}, nil
}

func (converter *MockTransactionConverter) ConvertRPCTransactionsToModels(transactions []core.RPCTransaction) ([]core.TransactionModel, error) {
	converter.ConvertHeaderTransactionIndexToIntCalled = true
	return nil, nil
}

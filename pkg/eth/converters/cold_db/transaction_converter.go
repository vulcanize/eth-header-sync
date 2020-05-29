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

package cold_db

import (
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/vulcanize/eth-header-sync/pkg/eth/core"
	"golang.org/x/sync/errgroup"
)

type ColdDbTransactionConverter struct{}

func NewColdDbTransactionConverter() *ColdDbTransactionConverter {
	return &ColdDbTransactionConverter{}
}

func (cdtc *ColdDbTransactionConverter) ConvertBlockTransactionsToCore(gethBlock *types.Block) ([]core.TransactionModel, error) {
	var g errgroup.Group
	coreTransactions := make([]core.TransactionModel, len(gethBlock.Transactions()))

	for gethTransactionIndex, gethTransaction := range gethBlock.Transactions() {
		transaction := gethTransaction
		transactionIndex := uint(gethTransactionIndex)
		g.Go(func() error {
			signer := getSigner(transaction)
			sender, err := signer.Sender(transaction)
			if err != nil {
				return err
			}
			coreTransaction := transToCoreTrans(transaction, &sender)
			coreTransactions[transactionIndex] = coreTransaction
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return coreTransactions, err
	}
	return coreTransactions, nil
}

func (cdtc *ColdDbTransactionConverter) ConvertRPCTransactionsToModels(transactions []core.RPCTransaction) ([]core.TransactionModel, error) {
	panic("converting transaction indexes to integer not supported for cold import")
}

func getSigner(tx *types.Transaction) types.Signer {
	v, _, _ := tx.RawSignatureValues()
	if v.Sign() != 0 && tx.Protected() {
		return types.NewEIP155Signer(tx.ChainId())
	}
	return types.HomesteadSigner{}
}

func transToCoreTrans(transaction *types.Transaction, from *common.Address) core.TransactionModel {
	return core.TransactionModel{
		Hash:     transaction.Hash().Hex(),
		Nonce:    transaction.Nonce(),
		To:       strings.ToLower(addressToHex(transaction.To())),
		From:     strings.ToLower(addressToHex(from)),
		GasLimit: transaction.Gas(),
		GasPrice: transaction.GasPrice().Int64(),
		Value:    transaction.Value().String(),
		Data:     transaction.Data(),
	}
}

func addressToHex(to *common.Address) string {
	if to == nil {
		return ""
	}
	return to.Hex()
}

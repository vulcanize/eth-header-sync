// Copyright 2018 Vulcanize
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package vat_fold

import (
	"bytes"
	"encoding/json"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/vulcanize/vulcanizedb/pkg/transformers/shared"
)

type Converter interface {
	ToModels(ethLogs []types.Log) ([]VatFoldModel, error)
}

type VatFoldConverter struct{}

func (VatFoldConverter) ToModels(ethLogs []types.Log) ([]VatFoldModel, error) {
	var models []VatFoldModel
	for _, ethLog := range ethLogs {
		err := verifyLog(ethLog)
		if err != nil {
			return nil, err
		}

		ilk := string(bytes.Trim(ethLog.Topics[1].Bytes(), "\x00"))
		urn := common.BytesToAddress(ethLog.Topics[2].Bytes()).String()
		rate := big.NewInt(0).SetBytes(ethLog.Topics[3].Bytes()).String()
		raw, err := json.Marshal(ethLog)

		if err != nil {
			return models, err
		}

		model := VatFoldModel{
			Ilk:              ilk,
			Urn:              urn,
			Rate:             rate,
			TransactionIndex: ethLog.TxIndex,
			Raw:              raw,
		}

		models = append(models, model)
	}
	return models, nil
}

func verifyLog(log types.Log) error {
	if len(log.Topics) < 4 {
		return errors.New("log missing topics")
	}

	sig := log.Topics[0].String()
	if sig != shared.VatFoldSignature {
		return errors.New("log is not a Vat.fold event")
	}

	return nil
}
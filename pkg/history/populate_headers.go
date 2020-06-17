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

package history

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/vulcanize/eth-header-sync/pkg/core"
	"github.com/vulcanize/eth-header-sync/pkg/repository"
)

// PopulateMissingHeaders populates missing headers in the database, it does so by finding block numbers where no header record exists
func PopulateMissingHeaders(fetcher core.Fetcher, headerRepository core.HeaderRepository, startingBlockNumber int64) (int, error) {
	lastBlock, err := fetcher.LastBlock()
	if err != nil {
		logrus.Error("PopulateMissingHeaders: Error getting last block: ", err)
		return 0, err
	}

	blockNumbers, err := headerRepository.MissingBlockNumbers(startingBlockNumber, lastBlock.Int64(), fetcher.Node().ID)
	if err != nil {
		logrus.Error("PopulateMissingHeaders: Error getting missing block numbers: ", err)
		return 0, err
	} else if len(blockNumbers) == 0 {
		return 0, nil
	}

	logrus.Debug(getBlockRangeString(blockNumbers))
	_, err = RetrieveAndUpdateHeaders(fetcher, headerRepository, blockNumbers)
	if err != nil {
		logrus.Error("PopulateMissingHeaders: Error getting/updating headers: ", err)
		return 0, err
	}
	return len(blockNumbers), nil
}

// RetrieveAndUpdateHeaders fetches the headers for the provided block numbers and upserts them into the Postgres database
func RetrieveAndUpdateHeaders(fetcher core.Fetcher, headerRepository core.HeaderRepository, blockNumbers []int64) (int, error) {
	headers, err := fetcher.GetHeadersByNumbers(blockNumbers)
	for _, header := range headers {
		_, err = headerRepository.CreateOrUpdateHeader(header)
		if err != nil {
			if err == repository.ErrValidHeaderExists {
				continue
			}
			return 0, err
		}
	}
	return len(headers), err
}

func getBlockRangeString(blockRange []int64) string {
	return fmt.Sprintf("Backfilling |%v| blocks", len(blockRange))
}

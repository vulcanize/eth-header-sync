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

package cmd

import (
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/vulcanize/eth-header-sync/pkg/eth"
	"github.com/vulcanize/eth-header-sync/pkg/eth/core"
	"github.com/vulcanize/eth-header-sync/pkg/eth/datastore"
	"github.com/vulcanize/eth-header-sync/pkg/eth/datastore/postgres/repositories"
	"github.com/vulcanize/eth-header-sync/pkg/eth/history"
	"github.com/vulcanize/eth-header-sync/utils"
)

// headerSyncCmd represents the headerSync command
var headerSyncCmd = &cobra.Command{
	Use:   "headerSync",
	Short: "Syncs VulcanizeDB with local ethereum node's block headers",
	Long: `Syncs VulcanizeDB with local ethereum node. Populates
Postgres with block headers.

./vulcanizedb headerSync --starting-block-number 0 --config public.toml

Expects ethereum node to be running and requires a .toml config:

  [database]
  name = "vulcanize_public"
  hostname = "localhost"
  port = 5432

  [client]
  ipcPath = "/Users/user/Library/Ethereum/geth.ipc"
`,
	Run: func(cmd *cobra.Command, args []string) {
		subCommand = cmd.CalledAs()
		logWithCommand = *log.WithField("SubCommand", subCommand)
		headerSync()
	},
}

func init() {
	rootCmd.AddCommand(headerSyncCmd)
	headerSyncCmd.Flags().Int64VarP(&startingBlockNumber, "starting-block-number", "s", 0, "Block number to start syncing from")
}

func backFillAllHeaders(blockchain core.BlockChain, headerRepository datastore.HeaderRepository, missingBlocksPopulated chan int, startingBlockNumber int64) {
	populated, err := history.PopulateMissingHeaders(blockchain, headerRepository, startingBlockNumber)
	if err != nil {
		// TODO Lots of possible errors in the call stack above. If errors occur, we still put
		// 0 in the channel, triggering another round
		logWithCommand.Error("backfillAllHeaders: Error populating headers: ", err)
	}
	missingBlocksPopulated <- populated
}

func headerSync() {
	ticker := time.NewTicker(pollingInterval)
	defer ticker.Stop()
	blockChain := getBlockChain()
	validateArgs(blockChain)
	db := utils.LoadPostgres(databaseConfig, blockChain.Node())

	headerRepository := repositories.NewHeaderRepository(&db)
	validator := history.NewHeaderValidator(blockChain, headerRepository, validationWindow)
	missingBlocksPopulated := make(chan int)
	go backFillAllHeaders(blockChain, headerRepository, missingBlocksPopulated, startingBlockNumber)

	for {
		select {
		case <-ticker.C:
			window, err := validator.ValidateHeaders()
			if err != nil {
				logWithCommand.Error("headerSync: ValidateHeaders failed: ", err)
			}
			logWithCommand.Debug(window.GetString())
		case n := <-missingBlocksPopulated:
			if n == 0 {
				time.Sleep(3 * time.Second)
			}
			go backFillAllHeaders(blockChain, headerRepository, missingBlocksPopulated, startingBlockNumber)
		}
	}
}

func validateArgs(blockChain *eth.BlockChain) {
	lastBlock, err := blockChain.LastBlock()
	if err != nil {
		logWithCommand.Error("validateArgs: Error getting last block: ", err)
	}
	if lastBlock.Int64() == 0 {
		logWithCommand.Fatal("geth initial: state sync not finished")
	}
	if startingBlockNumber > lastBlock.Int64() {
		logWithCommand.Fatal("starting block number > current block number")
	}
}

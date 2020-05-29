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

package repositories

import (
	"database/sql"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"

	"github.com/vulcanize/eth-header-sync/pkg/eth/core"
	"github.com/vulcanize/eth-header-sync/pkg/postgres"
)

var ErrValidHeaderExists = errors.New("valid header already exists")

type HeaderRepository struct {
	database *postgres.DB
}

func NewHeaderRepository(database *postgres.DB) HeaderRepository {
	return HeaderRepository{database: database}
}

func (repository HeaderRepository) CreateOrUpdateHeader(header core.Header) (int64, error) {
	hash, err := repository.getHeaderHash(header)
	if err != nil {
		if headerDoesNotExist(err) {
			return repository.InternalInsertHeader(header)
		}
		log.Error("CreateOrUpdateHeader: error getting header hash: ", err)
		return 0, err
	}
	if headerMustBeReplaced(hash, header) {
		return repository.replaceHeader(header)
	}
	return 0, ErrValidHeaderExists
}

func (repository HeaderRepository) CreateTransactions(headerID int64, transactions []core.TransactionModel) error {
	for _, transaction := range transactions {
		_, err := repository.database.Exec(`INSERT INTO public.header_sync_transactions
		(header_id, hash, gas_limit, gas_price, input_data, nonce, raw, tx_from, tx_index, tx_to, "value") 
		VALUES ($1, $2, $3::NUMERIC, $4::NUMERIC, $5, $6::NUMERIC, $7, $8, $9::NUMERIC, $10, $11::NUMERIC)
		ON CONFLICT DO NOTHING`, headerID, transaction.Hash, transaction.GasLimit, transaction.GasPrice,
			transaction.Data, transaction.Nonce, transaction.Raw, transaction.From, transaction.TxIndex, transaction.To,
			transaction.Value)
		if err != nil {
			return err
		}
	}
	return nil
}

func (repository HeaderRepository) CreateTransactionInTx(tx *sqlx.Tx, headerID int64, transaction core.TransactionModel) (int64, error) {
	var txID int64
	err := tx.QueryRowx(`INSERT INTO public.header_sync_transactions
		(header_id, hash, gas_limit, gas_price, input_data, nonce, raw, tx_from, tx_index, tx_to, "value")
		VALUES ($1, $2, $3::NUMERIC, $4::NUMERIC, $5, $6::NUMERIC, $7, $8, $9::NUMERIC, $10, $11::NUMERIC)
		ON CONFLICT (header_id, hash) DO UPDATE 
		SET (gas_limit, gas_price, input_data, nonce, raw, tx_from, tx_index, tx_to, "value") = ($3::NUMERIC, $4::NUMERIC, $5, $6::NUMERIC, $7, $8, $9::NUMERIC, $10, $11::NUMERIC)
		RETURNING id`,
		headerID, transaction.Hash, transaction.GasLimit, transaction.GasPrice,
		transaction.Data, transaction.Nonce, transaction.Raw, transaction.From,
		transaction.TxIndex, transaction.To, transaction.Value).Scan(&txID)
	if err != nil {
		log.Error("header_repository: error inserting transaction: ", err)
		return txID, err
	}
	return txID, err
}

func (repository HeaderRepository) GetHeader(blockNumber int64) (core.Header, error) {
	var header core.Header
	err := repository.database.Get(&header, `SELECT id, block_number, hash, raw, block_timestamp FROM headers WHERE block_number = $1 AND eth_node_fingerprint = $2`,
		blockNumber, repository.database.Node.ID)
	if err != nil {
		log.Error("GetHeader: error getting headers: ", err)
	}
	return header, err
}

func (repository HeaderRepository) MissingBlockNumbers(startingBlockNumber, endingBlockNumber int64, nodeID string) ([]int64, error) {
	numbers := make([]int64, 0)
	err := repository.database.Select(&numbers,
		`SELECT series.block_number
			FROM (SELECT generate_series($1::INT, $2::INT) AS block_number) AS series
			LEFT OUTER JOIN (SELECT block_number FROM headers
				WHERE eth_node_fingerprint = $3) AS synced
			USING (block_number)
			WHERE  synced.block_number IS NULL`,
		startingBlockNumber, endingBlockNumber, nodeID)
	if err != nil {
		log.Errorf("MissingBlockNumbers failed to get blocks between %v - %v for node %v",
			startingBlockNumber, endingBlockNumber, nodeID)
		return []int64{}, err
	}
	return numbers, nil
}

func headerMustBeReplaced(hash string, header core.Header) bool {
	return hash != header.Hash
}

func headerDoesNotExist(err error) bool {
	return err == sql.ErrNoRows
}

func (repository HeaderRepository) getHeaderHash(header core.Header) (string, error) {
	var hash string
	err := repository.database.Get(&hash, `SELECT hash FROM headers WHERE block_number = $1 AND eth_node_fingerprint = $2`,
		header.BlockNumber, repository.database.Node.ID)
	return hash, err
}

// Function is public so we can test insert being called for the same header
// Can happen when concurrent processes are inserting headers
// Otherwise should not occur since only called in CreateOrUpdateHeader
func (repository HeaderRepository) InternalInsertHeader(header core.Header) (int64, error) {
	var headerID int64
	row := repository.database.QueryRowx(
		`INSERT INTO public.headers (block_number, hash, block_timestamp, raw, node_id, eth_node_fingerprint)
		VALUES ($1, $2, $3::NUMERIC, $4, $5, $6) ON CONFLICT DO NOTHING RETURNING id`,
		header.BlockNumber, header.Hash, header.Timestamp, header.Raw, repository.database.NodeID, repository.database.Node.ID)
	err := row.Scan(&headerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, ErrValidHeaderExists
		}
		log.Error("InternalInsertHeader: error inserting header: ", err)
	}
	return headerID, err
}

func (repository HeaderRepository) replaceHeader(header core.Header) (int64, error) {
	_, err := repository.database.Exec(`DELETE FROM headers WHERE block_number = $1 AND eth_node_fingerprint = $2`,
		header.BlockNumber, repository.database.Node.ID)
	if err != nil {
		log.Error("replaceHeader: error deleting headers: ", err)
		return 0, err
	}
	return repository.InternalInsertHeader(header)
}

func (repository HeaderRepository) CreateHeaderSyncReceiptInTx(headerID, transactionID int64, receipt core.Receipt, tx *sqlx.Tx) (int64, error) {
	var receiptID int64
	addressID, getAddressErr := getOrCreateAddressInTransaction(tx, receipt.ContractAddress)
	if getAddressErr != nil {
		log.Error("createReceipt: Error getting address id: ", getAddressErr)
		return receiptID, getAddressErr
	}
	err := tx.QueryRowx(`INSERT INTO public.header_sync_receipts
               (header_id, transaction_id, contract_address_id, cumulative_gas_used, gas_used, state_root, status, tx_hash, rlp)
               VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			   ON CONFLICT (header_id, transaction_id) DO UPDATE
			   SET (contract_address_id, cumulative_gas_used, gas_used, state_root, status, tx_hash, rlp) = ($3, $4::NUMERIC, $5::NUMERIC, $6, $7, $8, $9)
               RETURNING id`,
		headerID, transactionID, addressID, receipt.CumulativeGasUsed, receipt.GasUsed, receipt.StateRoot, receipt.Status, receipt.TxHash, receipt.Rlp).Scan(&receiptID)
	if err != nil {
		log.Error("header_repository: error inserting receipt: ", err)
		return receiptID, err
	}
	return receiptID, err
}

func getOrCreateAddressInTransaction(tx *sqlx.Tx, address string) (int64, error) {
	checksumAddress := getChecksumAddress(address)
	hashedAddress := hexToKeccak256Hash(checksumAddress).Hex()

	var addressID int64
	getOrCreateErr := tx.Get(&addressID, getOrCreateAddressQuery, checksumAddress, hashedAddress)

	return addressID, getOrCreateErr
}

const getOrCreateAddressQuery = `WITH addressId AS (
			INSERT INTO addresses (address, hashed_address) VALUES ($1, $2) ON CONFLICT DO NOTHING RETURNING id
		)
		SELECT id FROM addresses WHERE address = $1
		UNION
		SELECT id FROM addressId`

func GetOrCreateAddress(db *postgres.DB, address string) (int64, error) {
	checksumAddress := getChecksumAddress(address)
	hashedAddress := hexToKeccak256Hash(checksumAddress).Hex()

	var addressID int64
	getOrCreateErr := db.Get(&addressID, getOrCreateAddressQuery, checksumAddress, hashedAddress)

	return addressID, getOrCreateErr
}

func GetOrCreateAddressInTransaction(tx *sqlx.Tx, address string) (int64, error) {
	checksumAddress := getChecksumAddress(address)
	hashedAddress := hexToKeccak256Hash(checksumAddress).Hex()

	var addressID int64
	getOrCreateErr := tx.Get(&addressID, getOrCreateAddressQuery, checksumAddress, hashedAddress)

	return addressID, getOrCreateErr
}

func GetAddressByID(db *postgres.DB, id int64) (string, error) {
	var address string
	getErr := db.Get(&address, `SELECT address FROM public.addresses WHERE id = $1`, id)
	return address, getErr
}

func getChecksumAddress(address string) string {
	stringAddressToCommonAddress := common.HexToAddress(address)
	return stringAddressToCommonAddress.Hex()
}

func hexToKeccak256Hash(hex string) common.Hash {
	return crypto.Keccak256Hash(common.FromHex(hex))
}

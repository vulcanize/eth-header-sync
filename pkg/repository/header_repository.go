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

package repository

import (
	"database/sql"
	"errors"

	log "github.com/sirupsen/logrus"

	"github.com/vulcanize/eth-header-sync/pkg/core"
	"github.com/vulcanize/eth-header-sync/pkg/postgres"
)

var ErrValidHeaderExists = errors.New("valid header already exists")

// HeaderRepository is the underlying type satisfying the core.HeaderRepository interface
type HeaderRepository struct {
	database *postgres.DB
}

// NewHeaderRepository returns a new HeaderRepository
func NewHeaderRepository(database *postgres.DB) HeaderRepository {
	return HeaderRepository{database: database}
}

// CreateOrUpdateHeader inserts a header model into the db
// If there is already a header at the height, it is replaced if the hash is not the expected value
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

// InternalInsertHeader inserts the provided header and returns its row id
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

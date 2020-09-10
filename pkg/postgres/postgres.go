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

package postgres

import (
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" //postgres driver
	"github.com/vulcanize/eth-header-sync/pkg/config"
	"github.com/vulcanize/eth-header-sync/pkg/core"
)

// DB is a wrapper around the sqlx.DB which associates node information with the connection pool
type DB struct {
	*sqlx.DB
	Node   core.Node
	NodeID int64
}

// NewDB returns a new DB for the provided database config and node info
func NewDB(databaseConfig config.Database, node core.Node) (*DB, error) {
	connectString := config.DbConnectionString(databaseConfig)
	db, connectErr := sqlx.Connect("postgres", connectString)
	if connectErr != nil {
		return &DB{}, ErrDBConnectionFailed(connectErr)
	}
	if databaseConfig.MaxOpen > 0 {
		db.SetMaxOpenConns(databaseConfig.MaxOpen)
	}
	if databaseConfig.MaxIdle > 0 {
		db.SetMaxIdleConns(databaseConfig.MaxIdle)
	}
	if databaseConfig.MaxLifetime > 0 {
		lifetime := time.Duration(databaseConfig.MaxLifetime) * time.Second
		db.SetConnMaxLifetime(lifetime)
	}
	pg := DB{DB: db, Node: node}
	nodeErr := pg.CreateNode(&node)
	if nodeErr != nil {
		return &DB{}, ErrUnableToSetNode(nodeErr)
	}
	return &pg, nil
}

// CreateNode inserts the node info into the database
func (db *DB) CreateNode(node *core.Node) error {
	var nodeID int64
	err := db.QueryRow(
		`INSERT INTO nodes (genesis_block, network_id, node_id, client_name, chain_id)
                VALUES ($1, $2, $3, $4, $5)
                ON CONFLICT (genesis_block, network_id, node_id, chain_id)
                  DO UPDATE
                    SET genesis_block = $1,
                        network_id = $2,
                        node_id = $3,
                        client_name = $4,
						chain_id = $5
                RETURNING id`,
		node.GenesisBlock, node.NetworkID, node.ID, node.ClientName, node.ChainID).Scan(&nodeID)
	if err != nil {
		return ErrUnableToSetNode(err)
	}
	db.NodeID = nodeID
	return nil
}

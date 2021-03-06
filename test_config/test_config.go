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

package test_config

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/vulcanize/eth-header-sync/pkg/config"
	"github.com/vulcanize/eth-header-sync/pkg/core"
	"github.com/vulcanize/eth-header-sync/pkg/postgres"
)

var TestConfig *viper.Viper
var DBConfig config.Database
var TestClient config.Client

func init() {
	setTestConfig()
}

func setTestConfig() {
	TestConfig = viper.New()
	TestConfig.SetConfigName("testing")
	TestConfig.AddConfigPath("$GOPATH/src/github.com/vulcanize/eth-header-sync/environments/")
	err := TestConfig.ReadInConfig()
	if err != nil {
		logrus.Fatal(err)
	}
	ipc := TestConfig.GetString("client.rpcPath")

	// If we don't have an ipc path in the config file, check the env variable
	if ipc == "" {
		TestConfig.BindEnv("url", "INFURA_URL")
		ipc = TestConfig.GetString("url")
	}
	if ipc == "" {
		logrus.Fatal(errors.New("testing.toml IPC path or $INFURA_URL env variable need to be set"))
	}

	hn := TestConfig.GetString("database.hostname")
	port := TestConfig.GetInt("database.port")
	name := TestConfig.GetString("database.name")

	DBConfig = config.Database{
		Hostname: hn,
		Name:     name,
		Port:     port,
	}
	TestClient = config.Client{
		RPCPath: ipc,
	}
}

// NewTestDB returns a new database connection pool object for testing purposes
func NewTestDB(node core.Node) *postgres.DB {
	db, err := postgres.NewDB(DBConfig, node)
	if err != nil {
		panic(fmt.Sprintf("Could not create new test db: %v", err))
	}
	return db
}

// CleanTestDB removes any of the data inserted into the test db during testing
func CleanTestDB(db *postgres.DB) {
	// can't delete from nodes since this function is called after the required node is persisted
	db.MustExec("DELETE FROM goose_db_version")
	db.MustExec("DELETE FROM headers")
}

// NewTestNode returns a new test node, with preconfigured params
func NewTestNode() core.Node {
	return core.Node{
		GenesisBlock: "GENESIS",
		NetworkID:    "1",
		ID:           "b6f90c0fdd8ec9607aed8ee45c69322e47b7063f0bfb7a29c8ecafab24d0a22d24dd2329b5ee6ed4125a03cb14e57fd584e67f9e53e6c631055cbbd82f080845",
		ClientName:   "Geth/v1.7.2-stable-1db4ecdc/darwin-amd64/go1.9",
	}
}

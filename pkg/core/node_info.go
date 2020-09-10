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

package core

// NodeType is an enum to represent different node types
type NodeType int

const (
	GETH NodeType = iota
	PARITY
	INFURA
	GANACHE
)

const (
	KOVAN_NETWORK_ID = 42
)

// Node holds params for the Ethereum client
type Node struct {
	GenesisBlock string
	NetworkID    string
	ChainID      uint64
	ID           string
	ClientName   string
}

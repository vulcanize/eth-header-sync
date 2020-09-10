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

package node_test

import (
	"github.com/spf13/viper"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vulcanize/eth-header-sync/pkg/node"
)

var EmpytHeaderHash = "0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347"

var _ = Describe("Node Info", func() {
	It("returns the genesis block for any client", func() {
		viper.Set("ethereum.genesisBlock", "0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347")
		n := node.MakeNode()
		Expect(n.GenesisBlock).To(Equal(EmpytHeaderHash))
	})

	It("returns the network id for any client", func() {
		viper.Set("ethereum.networkID", "1234.000000")
		n := node.MakeNode()
		Expect(n.NetworkID).To(Equal("1234.000000"))
	})
})

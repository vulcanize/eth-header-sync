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

package drip_drip_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/vulcanize/vulcanizedb/pkg/datastore"
	"github.com/vulcanize/vulcanizedb/pkg/datastore/postgres"
	"github.com/vulcanize/vulcanizedb/pkg/datastore/postgres/repositories"
	"github.com/vulcanize/vulcanizedb/pkg/fakes"
	"github.com/vulcanize/vulcanizedb/pkg/transformers/drip_drip"
	"github.com/vulcanize/vulcanizedb/pkg/transformers/test_data"
	"github.com/vulcanize/vulcanizedb/pkg/transformers/test_data/shared_behaviors"
	"github.com/vulcanize/vulcanizedb/test_config"
)

var _ = Describe("Drip drip repository", func() {
	var (
		db                 *postgres.DB
		dripDripRepository drip_drip.DripDripRepository
		headerRepository   datastore.HeaderRepository
	)

	BeforeEach(func() {
		db = test_config.NewTestDB(test_config.NewTestNode())
		test_config.CleanTestDB(db)
		headerRepository = repositories.NewHeaderRepository(db)
		dripDripRepository = drip_drip.DripDripRepository{}
		dripDripRepository.SetDB(db)
	})

	Describe("Create", func() {
		modelWithDifferentLogIdx := test_data.DripDripModel
		modelWithDifferentLogIdx.LogIndex++
		inputs := shared_behaviors.CreateBehaviorInputs{
			CheckedHeaderColumnName:  "drip_drip_checked",
			LogEventTableName:        "maker.drip_drip",
			TestModel:                test_data.DripDripModel,
			ModelWithDifferentLogIdx: modelWithDifferentLogIdx,
			Repository:               &dripDripRepository,
		}

		shared_behaviors.SharedRepositoryCreateBehaviors(&inputs)

		It("adds a drip drip event", func() {
			headerID, err := headerRepository.CreateOrUpdateHeader(fakes.FakeHeader)
			Expect(err).NotTo(HaveOccurred())
			err = dripDripRepository.Create(headerID, []interface{}{test_data.DripDripModel})

			Expect(err).NotTo(HaveOccurred())
			var dbDripDrip drip_drip.DripDripModel
			err = db.Get(&dbDripDrip, `SELECT ilk, log_idx, tx_idx, raw_log FROM maker.drip_drip WHERE header_id = $1`, headerID)
			Expect(err).NotTo(HaveOccurred())
			Expect(dbDripDrip.Ilk).To(Equal(test_data.DripDripModel.Ilk))
			Expect(dbDripDrip.LogIndex).To(Equal(test_data.DripDripModel.LogIndex))
			Expect(dbDripDrip.TransactionIndex).To(Equal(test_data.DripDripModel.TransactionIndex))
			Expect(dbDripDrip.Raw).To(MatchJSON(test_data.DripDripModel.Raw))
		})
	})

	Describe("MarkHeaderChecked", func() {
		inputs := shared_behaviors.MarkedHeaderCheckedBehaviorInputs{
			CheckedHeaderColumnName: "drip_drip_checked",
			Repository:              &dripDripRepository,
		}

		shared_behaviors.SharedRepositoryMarkHeaderCheckedBehaviors(&inputs)
	})

	Describe("MissingHeaders", func() {
		inputs := shared_behaviors.MissingHeadersBehaviorInputs{
			Repository:    &dripDripRepository,
			RepositoryTwo: &drip_drip.DripDripRepository{},
		}

		shared_behaviors.SharedRepositoryMissingHeadersBehaviors(&inputs)
	})
})

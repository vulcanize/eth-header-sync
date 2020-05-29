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

package repositories_test

import (
	"math/rand"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/vulcanize/eth-header-sync/pkg/eth/core"
	"github.com/vulcanize/eth-header-sync/pkg/eth/datastore"
	"github.com/vulcanize/eth-header-sync/pkg/eth/datastore/postgres/repositories"
	"github.com/vulcanize/eth-header-sync/pkg/eth/fakes"
	"github.com/vulcanize/eth-header-sync/pkg/postgres"
	"github.com/vulcanize/eth-header-sync/test_config"
)

var _ = Describe("Checked Headers repository", func() {
	var (
		db   *postgres.DB
		repo datastore.CheckedHeadersRepository
	)

	BeforeEach(func() {
		db = test_config.NewTestDB(test_config.NewTestNode())
		test_config.CleanTestDB(db)
		repo = repositories.NewCheckedHeadersRepository(db)
	})

	AfterEach(func() {
		closeErr := db.Close()
		Expect(closeErr).NotTo(HaveOccurred())
	})

	Describe("MarkHeaderChecked", func() {
		It("marks passed header as checked on insert", func() {
			headerRepository := repositories.NewHeaderRepository(db)
			headerID, headerErr := headerRepository.CreateOrUpdateHeader(fakes.FakeHeader)
			Expect(headerErr).NotTo(HaveOccurred())

			err := repo.MarkHeaderChecked(headerID)

			Expect(err).NotTo(HaveOccurred())
			var checkedCount int
			fetchErr := db.Get(&checkedCount, `SELECT check_count FROM public.headers WHERE id = $1`, headerID)
			Expect(fetchErr).NotTo(HaveOccurred())
			Expect(checkedCount).To(Equal(1))
		})

		It("increments check count on update", func() {
			headerRepository := repositories.NewHeaderRepository(db)
			headerID, headerErr := headerRepository.CreateOrUpdateHeader(fakes.FakeHeader)
			Expect(headerErr).NotTo(HaveOccurred())

			insertErr := repo.MarkHeaderChecked(headerID)
			Expect(insertErr).NotTo(HaveOccurred())

			updateErr := repo.MarkHeaderChecked(headerID)
			Expect(updateErr).NotTo(HaveOccurred())

			var checkedCount int
			fetchErr := db.Get(&checkedCount, `SELECT check_count FROM public.headers WHERE id = $1`, headerID)
			Expect(fetchErr).NotTo(HaveOccurred())
			Expect(checkedCount).To(Equal(2))
		})
	})

	Describe("MarkHeadersUnchecked", func() {
		It("removes rows for headers <= starting block number", func() {
			blockNumberOne := rand.Int63()
			blockNumberTwo := blockNumberOne + 1
			blockNumberThree := blockNumberOne + 2
			fakeHeaderOne := fakes.GetFakeHeader(blockNumberOne)
			fakeHeaderTwo := fakes.GetFakeHeader(blockNumberTwo)
			fakeHeaderThree := fakes.GetFakeHeader(blockNumberThree)
			headerRepository := repositories.NewHeaderRepository(db)
			// insert three headers with incrementing block number
			headerIdOne, insertHeaderOneErr := headerRepository.CreateOrUpdateHeader(fakeHeaderOne)
			Expect(insertHeaderOneErr).NotTo(HaveOccurred())
			headerIdTwo, insertHeaderTwoErr := headerRepository.CreateOrUpdateHeader(fakeHeaderTwo)
			Expect(insertHeaderTwoErr).NotTo(HaveOccurred())
			headerIdThree, insertHeaderThreeErr := headerRepository.CreateOrUpdateHeader(fakeHeaderThree)
			Expect(insertHeaderThreeErr).NotTo(HaveOccurred())
			// mark all headers checked
			markHeaderOneCheckedErr := repo.MarkHeaderChecked(headerIdOne)
			Expect(markHeaderOneCheckedErr).NotTo(HaveOccurred())
			markHeaderTwoCheckedErr := repo.MarkHeaderChecked(headerIdTwo)
			Expect(markHeaderTwoCheckedErr).NotTo(HaveOccurred())
			markHeaderThreeCheckedErr := repo.MarkHeaderChecked(headerIdThree)
			Expect(markHeaderThreeCheckedErr).NotTo(HaveOccurred())

			// mark headers unchecked since blockNumberTwo
			err := repo.MarkHeadersUnchecked(blockNumberTwo)

			Expect(err).NotTo(HaveOccurred())
			var headerOneCheckCount, headerTwoCheckCount, headerThreeCheckCount int
			getHeaderOneErr := db.Get(&headerOneCheckCount, `SELECT check_count FROM public.headers WHERE id = $1`, headerIdOne)
			Expect(getHeaderOneErr).NotTo(HaveOccurred())
			Expect(headerOneCheckCount).To(Equal(1))
			getHeaderTwoErr := db.Get(&headerTwoCheckCount, `SELECT check_count FROM public.headers WHERE id = $1`, headerIdTwo)
			Expect(getHeaderTwoErr).NotTo(HaveOccurred())
			Expect(headerTwoCheckCount).To(BeZero())
			getHeaderThreeErr := db.Get(&headerThreeCheckCount, `SELECT check_count FROM public.headers WHERE id = $1`, headerIdThree)
			Expect(getHeaderThreeErr).NotTo(HaveOccurred())
			Expect(headerThreeCheckCount).To(BeZero())
		})
	})

	Describe("UncheckedHeaders", func() {
		var (
			headerRepository      datastore.HeaderRepository
			startingBlockNumber   int64
			endingBlockNumber     int64
			middleBlockNumber     int64
			outOfRangeBlockNumber int64
			blockNumbers          []int64
			headerIDs             []int64
			err                   error
			uncheckedCheckCount   = int64(1)
			recheckCheckCount     = int64(2)
		)

		BeforeEach(func() {
			headerRepository = repositories.NewHeaderRepository(db)

			startingBlockNumber = rand.Int63()
			middleBlockNumber = startingBlockNumber + 1
			endingBlockNumber = startingBlockNumber + 2
			outOfRangeBlockNumber = endingBlockNumber + 1

			blockNumbers = []int64{startingBlockNumber, middleBlockNumber, endingBlockNumber, outOfRangeBlockNumber}

			headerIDs = []int64{}
			for _, n := range blockNumbers {
				headerID, err := headerRepository.CreateOrUpdateHeader(fakes.GetFakeHeader(n))
				headerIDs = append(headerIDs, headerID)
				Expect(err).NotTo(HaveOccurred())
			}
		})

		Describe("when ending block is specified", func() {
			It("excludes headers that are out of range", func() {
				headers, err := repo.UncheckedHeaders(startingBlockNumber, endingBlockNumber, uncheckedCheckCount)
				Expect(err).NotTo(HaveOccurred())

				headerBlockNumbers := getBlockNumbers(headers)
				Expect(headerBlockNumbers).To(ConsistOf(startingBlockNumber, middleBlockNumber, endingBlockNumber))
				Expect(headerBlockNumbers).NotTo(ContainElement(outOfRangeBlockNumber))
			})

			It("excludes headers that have been checked more than the check count", func() {
				_, err = db.Exec(`UPDATE public.headers SET check_count = 1 WHERE id = $1`, headerIDs[1])
				Expect(err).NotTo(HaveOccurred())

				headers, err := repo.UncheckedHeaders(startingBlockNumber, endingBlockNumber, uncheckedCheckCount)
				Expect(err).NotTo(HaveOccurred())

				headerBlockNumbers := getBlockNumbers(headers)
				Expect(headerBlockNumbers).To(ConsistOf(startingBlockNumber, endingBlockNumber))
				Expect(headerBlockNumbers).NotTo(ContainElement(middleBlockNumber))
			})

			It("does not exclude headers that have been checked less than the check count", func() {
				_, err = db.Exec(`UPDATE public.headers SET check_count = 1 WHERE id = $1`, headerIDs[1])
				Expect(err).NotTo(HaveOccurred())

				headers, err := repo.UncheckedHeaders(startingBlockNumber, endingBlockNumber, recheckCheckCount)
				Expect(err).NotTo(HaveOccurred())

				headerBlockNumbers := getBlockNumbers(headers)
				Expect(headerBlockNumbers).To(ConsistOf(startingBlockNumber, middleBlockNumber, endingBlockNumber))
			})

			It("only returns headers associated with the current node", func() {
				dbTwo := test_config.NewTestDB(core.Node{ID: "second"})
				headerRepositoryTwo := repositories.NewHeaderRepository(dbTwo)
				repoTwo := repositories.NewCheckedHeadersRepository(dbTwo)
				for _, n := range blockNumbers {
					_, err = headerRepositoryTwo.CreateOrUpdateHeader(fakes.GetFakeHeader(n + 10))
					Expect(err).NotTo(HaveOccurred())
				}

				nodeOneMissingHeaders, err := repo.UncheckedHeaders(startingBlockNumber, endingBlockNumber, uncheckedCheckCount)
				Expect(err).NotTo(HaveOccurred())
				nodeOneHeaderBlockNumbers := getBlockNumbers(nodeOneMissingHeaders)
				Expect(nodeOneHeaderBlockNumbers).To(ConsistOf(startingBlockNumber, middleBlockNumber, endingBlockNumber))

				nodeTwoMissingHeaders, err := repoTwo.UncheckedHeaders(startingBlockNumber, endingBlockNumber+10, uncheckedCheckCount)
				Expect(err).NotTo(HaveOccurred())
				nodeTwoHeaderBlockNumbers := getBlockNumbers(nodeTwoMissingHeaders)
				Expect(nodeTwoHeaderBlockNumbers).To(ConsistOf(startingBlockNumber+10, middleBlockNumber+10, endingBlockNumber+10))
			})
		})

		Describe("when ending block is -1", func() {
			var endingBlock = int64(-1)

			It("includes all non-checked headers when ending block is -1 ", func() {
				headers, err := repo.UncheckedHeaders(startingBlockNumber, endingBlock, uncheckedCheckCount)
				Expect(err).NotTo(HaveOccurred())

				headerBlockNumbers := getBlockNumbers(headers)
				Expect(headerBlockNumbers).To(ConsistOf(startingBlockNumber, middleBlockNumber, endingBlockNumber, outOfRangeBlockNumber))
			})

			It("excludes headers that have been checked more than the check count", func() {
				_, err = db.Exec(`UPDATE public.headers SET check_count = 1 WHERE id = $1`, headerIDs[1])
				Expect(err).NotTo(HaveOccurred())

				headers, err := repo.UncheckedHeaders(startingBlockNumber, endingBlock, uncheckedCheckCount)
				Expect(err).NotTo(HaveOccurred())

				headerBlockNumbers := getBlockNumbers(headers)
				Expect(headerBlockNumbers).To(ConsistOf(startingBlockNumber, endingBlockNumber, outOfRangeBlockNumber))
				Expect(headerBlockNumbers).NotTo(ContainElement(middleBlockNumber))
			})

			It("does not exclude headers that have been checked less than the check count", func() {
				_, err = db.Exec(`UPDATE public.headers SET check_count = 1 WHERE id = $1`, headerIDs[1])
				Expect(err).NotTo(HaveOccurred())

				headers, err := repo.UncheckedHeaders(startingBlockNumber, endingBlock, recheckCheckCount)
				Expect(err).NotTo(HaveOccurred())

				headerBlockNumbers := getBlockNumbers(headers)
				Expect(headerBlockNumbers).To(ConsistOf(startingBlockNumber, middleBlockNumber, endingBlockNumber, outOfRangeBlockNumber))
			})

			It("only returns headers associated with the current node", func() {
				dbTwo := test_config.NewTestDB(core.Node{ID: "second"})
				headerRepositoryTwo := repositories.NewHeaderRepository(dbTwo)
				repoTwo := repositories.NewCheckedHeadersRepository(dbTwo)
				for _, n := range blockNumbers {
					_, err = headerRepositoryTwo.CreateOrUpdateHeader(fakes.GetFakeHeader(n + 10))
					Expect(err).NotTo(HaveOccurred())
				}

				nodeOneMissingHeaders, err := repo.UncheckedHeaders(startingBlockNumber, endingBlock, uncheckedCheckCount)
				Expect(err).NotTo(HaveOccurred())
				nodeOneBlockNumbers := getBlockNumbers(nodeOneMissingHeaders)
				Expect(nodeOneBlockNumbers).To(ConsistOf(startingBlockNumber, middleBlockNumber, endingBlockNumber, outOfRangeBlockNumber))

				nodeTwoMissingHeaders, err := repoTwo.UncheckedHeaders(startingBlockNumber, endingBlock, uncheckedCheckCount)
				Expect(err).NotTo(HaveOccurred())
				nodeTwoBlockNumbers := getBlockNumbers(nodeTwoMissingHeaders)
				Expect(nodeTwoBlockNumbers).To(ConsistOf(startingBlockNumber+10, middleBlockNumber+10, endingBlockNumber+10, outOfRangeBlockNumber+10))
			})
		})
	})
})

func getBlockNumbers(headers []core.Header) []int64 {
	var headerBlockNumbers []int64
	for _, header := range headers {
		headerBlockNumbers = append(headerBlockNumbers, header.BlockNumber)
	}
	return headerBlockNumbers
}

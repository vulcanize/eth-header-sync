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
	"github.com/sirupsen/logrus"
	"github.com/vulcanize/eth-header-sync/pkg/core"
)

// HeaderValidator is the type reponsible for validating headers
type HeaderValidator struct {
	fetcher          core.Fetcher
	headerRepository core.HeaderRepository
	windowSize       int
}

// NewHeaderValidator returns a new HeaderValidator
func NewHeaderValidator(fetcher core.Fetcher, repository core.HeaderRepository, windowSize int) HeaderValidator {
	return HeaderValidator{
		fetcher:          fetcher,
		headerRepository: repository,
		windowSize:       windowSize,
	}
}

// ValidateHeaders validates headers at the head, returning the validation window used
func (validator HeaderValidator) ValidateHeaders() (ValidationWindow, error) {
	window, err := MakeValidationWindow(validator.fetcher, validator.windowSize)
	if err != nil {
		logrus.Error("ValidateHeaders: error creating validation window: ", err)
		return ValidationWindow{}, err
	}
	blockNumbers := MakeRange(window.LowerBound, window.UpperBound)
	_, err = RetrieveAndUpdateHeaders(validator.fetcher, validator.headerRepository, blockNumbers)
	if err != nil {
		logrus.Error("ValidateHeaders: error getting/updating headers: ", err)
		return ValidationWindow{}, err
	}
	return window, nil
}

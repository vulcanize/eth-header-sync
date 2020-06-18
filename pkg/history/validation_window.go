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
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/vulcanize/eth-header-sync/pkg/core"
)

// ValidationWindow represent the range of headers we validate at the header of the chain
type ValidationWindow struct {
	LowerBound int64
	UpperBound int64
}

// Size returns the size of the validation window
func (window ValidationWindow) Size() int {
	return int(window.UpperBound - window.LowerBound)
}

// MakeValidationWindow returns a validation window for the provided fetcher and window size
func MakeValidationWindow(fetcher core.Fetcher, windowSize int) (ValidationWindow, error) {
	upperBound, err := fetcher.LastBlock()
	if err != nil {
		log.Error("MakeValidationWindow: error getting LastBlock: ", err)
		return ValidationWindow{}, err
	}
	lowerBound := upperBound.Int64() - int64(windowSize)
	return ValidationWindow{lowerBound, upperBound.Int64()}, nil
}

// MakeRange creates a range from a min and max, exported for testing purposes
func MakeRange(min, max int64) []int64 {
	a := make([]int64, max-min+1)
	for i := range a {
		a[i] = min + int64(i)
	}
	return a
}

// GetString returns a string describing the current validation window
func (window ValidationWindow) GetString() string {
	return fmt.Sprintf("Validating Blocks |%v|-- Validation Window --|%v|",
		window.LowerBound, window.UpperBound)
}

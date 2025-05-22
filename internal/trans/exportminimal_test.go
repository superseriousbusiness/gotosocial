// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package trans_test

import (
	"fmt"
	"os"
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/trans"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type ExportMinimalTestSuite struct {
	TransTestSuite
}

func (suite *ExportMinimalTestSuite) TestExportMinimalOK() {
	// use a temporary file path that will be cleaned when the test is closed
	tempFilePath := fmt.Sprintf("%s/%s", suite.T().TempDir(), uuid.NewString())

	// export to the tempFilePath
	exporter := trans.NewExporter(suite.db)
	err := exporter.ExportMinimal(suite.T().Context(), tempFilePath)
	suite.NoError(err)

	// we should have some bytes in that file now
	b, err := os.ReadFile(tempFilePath)
	suite.NoError(err)
	suite.NotEmpty(b)
	fmt.Println(string(b))
}

func TestExportMinimalTestSuite(t *testing.T) {
	suite.Run(t, &ExportMinimalTestSuite{})
}

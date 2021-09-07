/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package trans

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	transmodel "github.com/superseriousbusiness/gotosocial/internal/trans/model"
)

func (i *importer) ImportMinimal(ctx context.Context, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("ImportMinimal: error opening file %s: %s", path, err)
	}

	decoder := json.NewDecoder(f)
	decoder.UseNumber()

	for {
		entry := transmodel.TransEntry{}
		err := decoder.Decode(&entry)
		if err != nil {
			if err == io.EOF {
				i.log.Infof("ImportMinimal: reached end of file")
				return neatClose(f)
			}
			return fmt.Errorf("ImportMinimal: error decoding in readLoop: %s", err)
		}
		if err := i.inputEntry(ctx, entry); err != nil {
			return fmt.Errorf("ImportMinimal: error inputting entry: %s", err)
		}
	}
}

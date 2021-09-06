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
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	transmodel "github.com/superseriousbusiness/gotosocial/internal/trans/model"
)

func (e *exporter) ExportMinimal(ctx context.Context, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}

	w := bufio.NewWriter(f)
	encoder := json.NewEncoder(w)

	accounts := []*transmodel.Account{}
	if err := e.db.GetWhere(ctx, []db.Where{{Key: "domain", Value: nil}}, &accounts); err != nil {
		return fmt.Errorf("error selecting accounts: %s", err)
	}

	for _, a := range accounts {
		encoder.Encode(a)
	}

	return neatClose(w, f)
}

func neatClose(w *bufio.Writer, f *os.File) error {
	if err := w.Flush(); err != nil {
		return fmt.Errorf("error flushing writer: %s", err)
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("error closing file: %s", err)
	}

	return nil
}

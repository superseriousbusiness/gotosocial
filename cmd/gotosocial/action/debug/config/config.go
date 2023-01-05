/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

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

package config

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/superseriousbusiness/gotosocial/cmd/gotosocial/action"
	"github.com/superseriousbusiness/gotosocial/internal/config"
)

// Config just prints the collated config out to stdout as json.
var Config action.GTSAction = func(ctx context.Context) (err error) {
	var raw map[string]interface{}

	// Marshal configuration to a raw JSON map
	config.Config(func(cfg *config.Configuration) {
		raw, err = cfg.MarshalMap()
	})
	if err != nil {
		return err
	}

	// Marshal map to JSON
	b, err := json.Marshal(raw)
	if err != nil {
		return err
	}

	// Print to stdout
	fmt.Printf("%s\n", b)
	return nil
}

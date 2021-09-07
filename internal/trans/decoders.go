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
	"crypto/rsa"
	"fmt"
	"net"
	"reflect"
	"time"

	"github.com/mitchellh/mapstructure"
	transmodel "github.com/superseriousbusiness/gotosocial/internal/trans/model"
)

func accountDecode(e transmodel.TransEntry) (*transmodel.Account, error) {
	a := &transmodel.Account{}

	decoderConfig := &mapstructure.DecoderConfig{
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeHookFunc(time.RFC3339),
			PrivateKeyHookFunc(),
		),
		Result: a,
	}
	decoder, err := mapstructure.NewDecoder(decoderConfig)
	if err != nil {
		return nil, fmt.Errorf("accountDecode: error creating decoder: %s", err)
	}

	if err := decoder.Decode(&e); err != nil {
		return nil, fmt.Errorf("accountDecode: error decoding account: %s", err)
	}

	return a, nil
}

var PrivateKeyHookFunc mapstructure.DecodeHookFunc = func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if t != reflect.TypeOf(rsa.PrivateKey{}) {
		return data, nil
	}


	rsa.

	// Convert it by parsing
	_, net, err := net.ParseCIDR(data.(string))
	return net, err

}

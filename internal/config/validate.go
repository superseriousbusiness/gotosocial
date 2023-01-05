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
	"errors"
	"fmt"
	"strings"

	"github.com/miekg/dns"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

// Validate validates global config settings which don't have defaults, to make sure they are set sensibly.
func Validate() error {
	errs := []error{}

	// host
	host := GetHost()
	if host == "" {
		errs = append(errs, fmt.Errorf("%s must be set", HostFlag()))
	}

	// accountDomain; only check if host was set, otherwise there's no point
	if host != "" {
		switch ad := GetAccountDomain(); ad {
		case "":
			SetAccountDomain(GetHost())
		default:
			if !dns.IsSubDomain(ad, host) {
				errs = append(errs, fmt.Errorf("%s was %s and %s was %s, but %s is not a valid subdomain of %s", HostFlag(), host, AccountDomainFlag(), ad, host, ad))
			}
		}
	}

	// protocol
	switch proto := GetProtocol(); proto {
	case "https":
		// no problem
		break
	case "http":
		log.Warnf("%s was set to 'http'; this should *only* be used for debugging and tests!", ProtocolFlag())
	case "":
		errs = append(errs, fmt.Errorf("%s must be set", ProtocolFlag()))
	default:
		errs = append(errs, fmt.Errorf("%s must be set to either http or https, provided value was %s", ProtocolFlag(), proto))
	}

	webAssetsBaseDir := GetWebAssetBaseDir()
	if webAssetsBaseDir == "" {
		errs = append(errs, fmt.Errorf("%s must be set", WebAssetBaseDirFlag()))
	}

	if len(errs) > 0 {
		errStrings := []string{}
		for _, err := range errs {
			errStrings = append(errStrings, err.Error())
		}
		return errors.New(strings.Join(errStrings, "; "))
	}

	return nil
}

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

package config

import (
	"fmt"
	"net/url"
	"strings"

	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/language"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"github.com/miekg/dns"
)

// Validate validates global config settings.
func Validate() error {
	// Gather all validation errors in
	// easily readable format for admins.
	var (
		errs gtserror.MultiError
		errf = func(format string, a ...any) {
			errs = append(errs, fmt.Errorf(format, a...))
		}
	)

	// `host`
	host := GetHost()
	if host == "" {
		errf("%s must be set", HostFlag())
	}

	// If `account-domain` and `host`
	// are set, `host` must be a valid
	// subdomain of `account-domain`.
	if host != "" {
		ad := GetAccountDomain()
		if ad == "" {
			// `account-domain` not set, fall
			// back by setting it to `host`.
			SetAccountDomain(GetHost())
		} else if !dns.IsSubDomain(ad, host) {
			errf(
				"%s %s is not a valid subdomain of %s %s",
				AccountDomainFlag(), ad, HostFlag(), host,
			)
		}
	}

	// Ensure `protocol` sensibly set.
	switch proto := GetProtocol(); proto {
	case "https":
		// No problem.

	case "http":
		log.Warnf(
			nil,
			"%s was set to 'http'; this should *only* be used for debugging and tests!",
			ProtocolFlag(),
		)

	case "":
		errf("%s must be set", ProtocolFlag())

	default:
		errf(
			"%s must be set to either http or https, provided value was %s",
			ProtocolFlag(), proto,
		)
	}

	// `federation-mode` should be
	// "blocklist" or "allowlist".
	switch fediMode := GetInstanceFederationMode(); fediMode {
	case InstanceFederationModeBlocklist, InstanceFederationModeAllowlist:
		// No problem.

	case "":
		errf("%s must be set", InstanceFederationModeFlag())

	default:
		errf(
			"%s must be set to either blocklist or allowlist, provided value was %s",
			InstanceFederationModeFlag(), fediMode,
		)
	}

	// Parse `instance-languages`, and
	// set enriched version into config.
	parsedLangs, err := language.InitLangs(GetInstanceLanguages().TagStrs())
	if err != nil {
		errf(
			"%s could not be parsed as an array of valid BCP47 language tags: %v",
			InstanceLanguagesFlag(), err,
		)
	} else {
		// Parsed successfully, put enriched
		// versions in config immediately.
		SetInstanceLanguages(parsedLangs)
	}

	// `instance-stats-mode` should be
	// "", "zero", "serve", or "baffle"
	switch statsMode := GetInstanceStatsMode(); statsMode {
	case InstanceStatsModeDefault, InstanceStatsModeZero, InstanceStatsModeServe, InstanceStatsModeBaffle:
		// No problem.

	default:
		errf(
			"%s must be set to empty string, zero, serve, or baffle, provided value was %s",
			InstanceFederationModeFlag(), statsMode,
		)
	}

	// `web-assets-base-dir`.
	webAssetsBaseDir := GetWebAssetBaseDir()
	if webAssetsBaseDir == "" {
		errf("%s must be set", WebAssetBaseDirFlag())
	}

	// `storage-s3-redirect-url`
	if s3RedirectURL := GetStorageS3RedirectURL(); s3RedirectURL != "" {
		if strings.HasSuffix(s3RedirectURL, "/") {
			errf(
				"%s must not end with a trailing slash",
				StorageS3RedirectURLFlag(),
			)
		}

		if url, err := url.Parse(s3RedirectURL); err != nil {
			errf(
				"%s invalid: %w",
				StorageS3RedirectURLFlag(), err,
			)
		} else if url.Scheme != "https" && url.Scheme != "http" {
			errf(
				"%s scheme must be https or http",
				StorageS3RedirectURLFlag(),
			)
		}
	}

	// Custom / LE TLS settings.
	//
	// Only one of custom certs or LE can be set,
	// and if using custom certs then all relevant
	// values must be provided.
	var (
		tlsChain     = GetTLSCertificateChain()
		tlsKey       = GetTLSCertificateKey()
		tlsChainFlag = TLSCertificateChainFlag()
		tlsKeyFlag   = TLSCertificateKeyFlag()
	)

	if GetLetsEncryptEnabled() && (tlsChain != "" || tlsKey != "") {
		errf(
			"%s cannot be true when %s and/or %s are also set",
			LetsEncryptEnabledFlag(), tlsChainFlag, tlsKeyFlag,
		)
	}

	if (tlsChain != "" && tlsKey == "") || (tlsChain == "" && tlsKey != "") {
		errf(
			"%s and %s need to both be set or unset",
			tlsChainFlag, tlsKeyFlag,
		)
	}

	return errs.Combine()
}

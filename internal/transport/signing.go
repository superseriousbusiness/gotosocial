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

package transport

import (
	"github.com/go-fed/httpsig"
)

var (
	// http signer preferences
	prefs       = []httpsig.Algorithm{httpsig.RSA_SHA256}
	digestAlgo  = httpsig.DigestSha256
	getHeaders  = []string{httpsig.RequestTarget, "host", "date"}
	postHeaders = []string{httpsig.RequestTarget, "host", "date", "digest"}
)

// NewGETSigner returns a new httpsig.Signer instance initialized with GTS GET preferences.
func NewGETSigner(expiresIn int64) (httpsig.Signer, error) {
	sig, _, err := httpsig.NewSigner(prefs, digestAlgo, getHeaders, httpsig.Signature, expiresIn)
	return sig, err
}

// NewPOSTSigner returns a new httpsig.Signer instance initialized with GTS POST preferences.
func NewPOSTSigner(expiresIn int64) (httpsig.Signer, error) {
	sig, _, err := httpsig.NewSigner(prefs, digestAlgo, postHeaders, httpsig.Signature, expiresIn)
	return sig, err
}

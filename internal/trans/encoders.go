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

package trans

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"os"

	transmodel "github.com/superseriousbusiness/gotosocial/internal/trans/model"
)

// accountEncode handles special fields like private + public keys on accounts
func (e *exporter) accountEncode(ctx context.Context, f *os.File, a *transmodel.Account) error {
	a.Type = transmodel.TransAccount

	// marshal public key
	encodedPublicKey := x509.MarshalPKCS1PublicKey(a.PublicKey)
	if encodedPublicKey == nil {
		return errors.New("could not MarshalPKCS1PublicKey")
	}
	publicKeyBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: encodedPublicKey,
	})
	a.PublicKeyString = string(publicKeyBytes)

	if a.Domain == "" {
		// marshal private key for local account
		encodedPrivateKey := x509.MarshalPKCS1PrivateKey(a.PrivateKey)
		if encodedPrivateKey == nil {
			return errors.New("could not MarshalPKCS1PrivateKey")
		}
		privateKeyBytes := pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: encodedPrivateKey,
		})
		a.PrivateKeyString = string(privateKeyBytes)
	}

	return e.simpleEncode(ctx, f, a, a.ID)
}

// simpleEncode can be used for any type that doesn't have special keys which need handling differently,
// or for types where special keys have already been handled.
//
// Beware, the 'type' key on the passed interface should already have been set, since simpleEncode won't know
// what type it is! If you try to decode stuff you've encoded with a missing type key, you're going to have a bad time.
func (e *exporter) simpleEncode(ctx context.Context, file *os.File, i interface{}, id string) error {
	_, alreadyWritten := e.writtenIDs[id]
	if alreadyWritten {
		// this exporter has already exported an entry with this ID, no need to do it twice
		return nil
	}

	err := json.NewEncoder(file).Encode(i)
	if err != nil {
		return fmt.Errorf("simpleEncode: error encoding entry with id %s: %s", id, err)
	}

	e.writtenIDs[id] = true
	return nil
}

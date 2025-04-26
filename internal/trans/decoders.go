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

package trans

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"time"

	transmodel "code.superseriousbusiness.org/gotosocial/internal/trans/model"
	"github.com/mitchellh/mapstructure"
)

func newDecoder(target interface{}) (*mapstructure.Decoder, error) {
	decoderConfig := &mapstructure.DecoderConfig{
		DecodeHook: mapstructure.StringToTimeHookFunc(time.RFC3339), // this is needed to decode time.Time entries serialized as string
		Result:     target,
	}
	return mapstructure.NewDecoder(decoderConfig)
}

func (i *importer) accountDecode(e transmodel.Entry) (*transmodel.Account, error) {
	a := &transmodel.Account{}
	if err := i.simpleDecode(e, a); err != nil {
		return nil, err
	}

	// extract public key
	publicKeyBlock, _ := pem.Decode([]byte(a.PublicKeyString))
	if publicKeyBlock == nil {
		return nil, errors.New("accountDecode: error decoding account public key")
	}
	publicKey, err := x509.ParsePKCS1PublicKey(publicKeyBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("accountDecode: error parsing account public key: %s", err)
	}
	a.PublicKey = publicKey

	if a.Domain == "" {
		// extract private key (local account)
		privateKeyBlock, _ := pem.Decode([]byte(a.PrivateKeyString))
		if privateKeyBlock == nil {
			return nil, errors.New("accountDecode: error decoding account private key")
		}
		privateKey, err := x509.ParsePKCS1PrivateKey(privateKeyBlock.Bytes)
		if err != nil {
			return nil, fmt.Errorf("accountDecode: error parsing account private key: %s", err)
		}
		a.PrivateKey = privateKey
	}

	return a, nil
}

func (i *importer) blockDecode(e transmodel.Entry) (*transmodel.Block, error) {
	b := &transmodel.Block{}
	if err := i.simpleDecode(e, b); err != nil {
		return nil, err
	}

	return b, nil
}

func (i *importer) domainBlockDecode(e transmodel.Entry) (*transmodel.DomainBlock, error) {
	b := &transmodel.DomainBlock{}
	if err := i.simpleDecode(e, b); err != nil {
		return nil, err
	}

	return b, nil
}

func (i *importer) followDecode(e transmodel.Entry) (*transmodel.Follow, error) {
	f := &transmodel.Follow{}
	if err := i.simpleDecode(e, f); err != nil {
		return nil, err
	}

	return f, nil
}

func (i *importer) followRequestDecode(e transmodel.Entry) (*transmodel.FollowRequest, error) {
	f := &transmodel.FollowRequest{}
	if err := i.simpleDecode(e, f); err != nil {
		return nil, err
	}

	return f, nil
}

func (i *importer) instanceDecode(e transmodel.Entry) (*transmodel.Instance, error) {
	inst := &transmodel.Instance{}
	if err := i.simpleDecode(e, inst); err != nil {
		return nil, err
	}

	return inst, nil
}

func (i *importer) userDecode(e transmodel.Entry) (*transmodel.User, error) {
	u := &transmodel.User{}
	if err := i.simpleDecode(e, u); err != nil {
		return nil, err
	}

	return u, nil
}

func (i *importer) simpleDecode(entry transmodel.Entry, target interface{}) error {
	decoder, err := newDecoder(target)
	if err != nil {
		return fmt.Errorf("simpleDecode: error creating decoder: %s", err)
	}

	if err := decoder.Decode(&entry); err != nil {
		return fmt.Errorf("simpleDecode: error decoding: %s", err)
	}

	return nil
}

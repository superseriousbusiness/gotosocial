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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/superseriousbusiness/gotosocial/internal/log"
	transmodel "github.com/superseriousbusiness/gotosocial/internal/trans/model"
)

func (i *importer) Import(ctx context.Context, path string) error {
	if path == "" {
		return errors.New("Export: path empty")
	}

	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("Import: couldn't export to %s: %s", path, err)
	}

	decoder := json.NewDecoder(file)
	decoder.UseNumber()

	for {
		entry := transmodel.Entry{}
		err := decoder.Decode(&entry)
		if err != nil {
			if err == io.EOF {
				log.Infof("Import: reached end of file")
				return neatClose(file)
			}
			return fmt.Errorf("Import: error decoding in readLoop: %s", err)
		}
		if err := i.inputEntry(ctx, entry); err != nil {
			return fmt.Errorf("Import: error inputting entry: %s", err)
		}
	}
}

func (i *importer) inputEntry(ctx context.Context, entry transmodel.Entry) error {
	t, ok := entry[transmodel.TypeKey].(string)
	if !ok {
		return errors.New("inputEntry: could not derive entry type: missing or malformed 'type' key in json")
	}

	switch transmodel.Type(t) {
	case transmodel.TransAccount:
		account, err := i.accountDecode(entry)
		if err != nil {
			return fmt.Errorf("inputEntry: error decoding entry into account: %s", err)
		}
		if err := i.putInDB(ctx, account); err != nil {
			return fmt.Errorf("inputEntry: error adding account to database: %s", err)
		}
		log.Infof("inputEntry: added account with id %s", account.ID)
		return nil
	case transmodel.TransBlock:
		block, err := i.blockDecode(entry)
		if err != nil {
			return fmt.Errorf("inputEntry: error decoding entry into block: %s", err)
		}
		if err := i.putInDB(ctx, block); err != nil {
			return fmt.Errorf("inputEntry: error adding block to database: %s", err)
		}
		log.Infof("inputEntry: added block with id %s", block.ID)
		return nil
	case transmodel.TransDomainBlock:
		block, err := i.domainBlockDecode(entry)
		if err != nil {
			return fmt.Errorf("inputEntry: error decoding entry into domain block: %s", err)
		}
		if err := i.putInDB(ctx, block); err != nil {
			return fmt.Errorf("inputEntry: error adding domain block to database: %s", err)
		}
		log.Infof("inputEntry: added domain block with id %s", block.ID)
		return nil
	case transmodel.TransFollow:
		follow, err := i.followDecode(entry)
		if err != nil {
			return fmt.Errorf("inputEntry: error decoding entry into follow: %s", err)
		}
		if err := i.putInDB(ctx, follow); err != nil {
			return fmt.Errorf("inputEntry: error adding follow to database: %s", err)
		}
		log.Infof("inputEntry: added follow with id %s", follow.ID)
		return nil
	case transmodel.TransFollowRequest:
		fr, err := i.followRequestDecode(entry)
		if err != nil {
			return fmt.Errorf("inputEntry: error decoding entry into follow request: %s", err)
		}
		if err := i.putInDB(ctx, fr); err != nil {
			return fmt.Errorf("inputEntry: error adding follow request to database: %s", err)
		}
		log.Infof("inputEntry: added follow request with id %s", fr.ID)
		return nil
	case transmodel.TransInstance:
		inst, err := i.instanceDecode(entry)
		if err != nil {
			return fmt.Errorf("inputEntry: error decoding entry into instance: %s", err)
		}
		if err := i.putInDB(ctx, inst); err != nil {
			return fmt.Errorf("inputEntry: error adding instance to database: %s", err)
		}
		log.Infof("inputEntry: added instance with id %s", inst.ID)
		return nil
	case transmodel.TransUser:
		user, err := i.userDecode(entry)
		if err != nil {
			return fmt.Errorf("inputEntry: error decoding entry into user: %s", err)
		}
		if err := i.putInDB(ctx, user); err != nil {
			return fmt.Errorf("inputEntry: error adding user to database: %s", err)
		}
		log.Infof("inputEntry: added user with id %s", user.ID)
		return nil
	}

	log.Errorf("inputEntry: didn't recognize transtype '%s', skipping it", t)
	return nil
}

func (i *importer) putInDB(ctx context.Context, entry interface{}) error {
	return i.db.Put(ctx, entry)
}

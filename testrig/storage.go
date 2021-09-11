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

package testrig

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"

	"git.iim.gay/grufwub/go-store/kv"
	"git.iim.gay/grufwub/go-store/storage"
	"git.iim.gay/grufwub/go-store/util"
)

// NewTestStorage returns a new in memory storage with the default test config
func NewTestStorage() *kv.KVStore {
	storage, err := kv.OpenStorage(&inMemStorage{storage: map[string][]byte{}, overwrite: false})
	if err != nil {
		panic(err)
	}
	return storage
}

// StandardStorageSetup populates the storage with standard test entries from the given directory.
func StandardStorageSetup(s *kv.KVStore, relativePath string) {
	storedA := newTestStoredAttachments()
	a := NewTestAttachments()
	for k, paths := range storedA {
		attachmentInfo, ok := a[k]
		if !ok {
			panic(fmt.Errorf("key %s not found in test attachments", k))
		}
		filenameOriginal := paths.Original
		filenameSmall := paths.Small
		pathOriginal := attachmentInfo.File.Path
		pathSmall := attachmentInfo.Thumbnail.Path
		bOriginal, err := os.ReadFile(fmt.Sprintf("%s/%s", relativePath, filenameOriginal))
		if err != nil {
			panic(err)
		}
		if err := s.Put(pathOriginal, bOriginal); err != nil {
			panic(err)
		}
		bSmall, err := os.ReadFile(fmt.Sprintf("%s/%s", relativePath, filenameSmall))
		if err != nil {
			panic(err)
		}
		if err := s.Put(pathSmall, bSmall); err != nil {
			panic(err)
		}
	}

	storedE := newTestStoredEmoji()
	e := NewTestEmojis()
	for k, paths := range storedE {
		emojiInfo, ok := e[k]
		if !ok {
			panic(fmt.Errorf("key %s not found in test emojis", k))
		}
		filenameOriginal := paths.Original
		filenameStatic := paths.Static
		pathOriginal := emojiInfo.ImagePath
		pathStatic := emojiInfo.ImageStaticPath
		bOriginal, err := os.ReadFile(fmt.Sprintf("%s/%s", relativePath, filenameOriginal))
		if err != nil {
			panic(err)
		}
		if err := s.Put(pathOriginal, bOriginal); err != nil {
			panic(err)
		}
		bStatic, err := os.ReadFile(fmt.Sprintf("%s/%s", relativePath, filenameStatic))
		if err != nil {
			panic(err)
		}
		if err := s.Put(pathStatic, bStatic); err != nil {
			panic(err)
		}
	}
}

// StandardStorageTeardown deletes everything in storage so that it's clean for the next test
func StandardStorageTeardown(s *kv.KVStore) {
	iter, err := s.Iterator(nil)
	if err != nil {
		panic(err)
	}
	keys := []string{}
	for iter.Next() {
		keys = append(keys, iter.Key())
	}
	iter.Release()
	for _, k := range keys {
		if err := s.Delete(k); err != nil {
			panic(err)
		}
	}
}

type inMemStorage struct {
	storage   map[string][]byte
	overwrite bool
}

func (s *inMemStorage) Clean() error {
	return nil
}

func (s *inMemStorage) ReadBytes(key string) ([]byte, error) {
	b, ok := s.storage[key]
	if !ok {
		return nil, errors.New("key not found")
	}
	return b, nil
}

func (s *inMemStorage) ReadStream(key string) (io.ReadCloser, error) {
	b, err := s.ReadBytes(key)
	if err != nil {
		return nil, err
	}
	return util.NopReadCloser(bytes.NewReader(b)), nil
}

func (s *inMemStorage) WriteBytes(key string, value []byte) error {
	if _, ok := s.storage[key]; ok && !s.overwrite {
		return errors.New("key already in storage")
	}
	s.storage[key] = copyBytes(value)
	return nil
}

func (s *inMemStorage) WriteStream(key string, r io.Reader) error {
	b, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	return s.WriteBytes(key, b)
}

func (s *inMemStorage) Stat(key string) (bool, error) {
	_, ok := s.storage[key]
	return ok, nil
}

func (s *inMemStorage) Remove(key string) error {
	if _, ok := s.storage[key]; !ok {
		return errors.New("key not found")
	}
	delete(s.storage, key)
	return nil
}

func (s *inMemStorage) WalkKeys(opts *storage.WalkKeysOptions) error {
	if opts == nil || opts.WalkFn == nil {
		return errors.New("invalid walkfn")
	}
	for key := range s.storage {
		opts.WalkFn(entry(key))
	}
	return nil
}

type entry string

func (e entry) Key() string {
	return string(e)
}

func copyBytes(b []byte) []byte {
	p := make([]byte, len(b))
	copy(p, b)
	return p
}

/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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
	"fmt"
	"os"
	"path"

	"codeberg.org/gruf/go-store/kv"
	"codeberg.org/gruf/go-store/storage"
)

// NewTestStorage returns a new in memory storage with the default test config
func NewTestStorage() *kv.KVStore {
	storage, err := kv.OpenStorage(storage.OpenMemory(200, false))
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
	defer os.RemoveAll(path.Join(os.TempDir(), "gotosocial"))

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

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
	"context"
	"fmt"
	"os"
	"path"

	"codeberg.org/gruf/go-store/v2/kv"
	"codeberg.org/gruf/go-store/v2/storage"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	gtsstorage "github.com/superseriousbusiness/gotosocial/internal/storage"
)

// NewInMemoryStorage returns a new in memory storage with the default test config
func NewInMemoryStorage() *gtsstorage.Local {
	storage, err := kv.OpenStorage(storage.OpenMemory(200, false))
	if err != nil {
		panic(err)
	}
	return &gtsstorage.Local{KVStore: storage}
}

func NewS3Storage() gtsstorage.Driver {
	mc, err := minio.New(os.Getenv("GTS_STORAGE_S3_ENDPOINT"), &minio.Options{
		Creds:  credentials.NewStaticV4(os.Getenv("GTS_STORAGE_S3_ACCESS_KEY"), os.Getenv("GTS_STORAGE_S3_SECRET_KEY"), ""),
		Secure: false,
	})
	if err != nil {
		panic(err)
	}
	return gtsstorage.NewS3(mc, os.Getenv("GTS_STORAGE_S3_BUCKET"))
}

// StandardStorageSetup populates the storage with standard test entries from the given directory.
func StandardStorageSetup(s gtsstorage.Driver, relativePath string) {
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
		if err := s.Put(context.TODO(), pathOriginal, bOriginal); err != nil {
			panic(err)
		}
		bSmall, err := os.ReadFile(fmt.Sprintf("%s/%s", relativePath, filenameSmall))
		if err != nil {
			panic(err)
		}
		if err := s.Put(context.TODO(), pathSmall, bSmall); err != nil {
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
		if err := s.Put(context.TODO(), pathOriginal, bOriginal); err != nil {
			panic(err)
		}
		bStatic, err := os.ReadFile(fmt.Sprintf("%s/%s", relativePath, filenameStatic))
		if err != nil {
			panic(err)
		}
		if err := s.Put(context.TODO(), pathStatic, bStatic); err != nil {
			panic(err)
		}
	}
}

// StandardStorageTeardown deletes everything in storage so that it's clean for
// the next test
// nolint:gocritic // complains about the type switch, but it's the cleanest solution
func StandardStorageTeardown(s gtsstorage.Driver) {
	defer os.RemoveAll(path.Join(os.TempDir(), "gotosocial"))

	switch st := s.(type) {
	case *gtsstorage.Local:
		iter, err := st.KVStore.Iterator(context.Background(), nil)
		if err != nil {
			panic(err)
		}
		keys := []string{}
		for iter.Next() {
			keys = append(keys, iter.Key())
		}
		iter.Release()
		for _, k := range keys {
			if err := s.Delete(context.TODO(), k); err != nil {
				panic(err)
			}
		}
	}
}

package media

import (
	"fmt"
	"sync"

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

type Media struct {
	mu         sync.Mutex
	attachment *gtsmodel.MediaAttachment
	rawData    []byte
}

func (m *Media) Thumb() (*ImageMeta, error) {
	m.mu.Lock()
	thumb, err := deriveThumbnail(m.rawData, m.attachment.File.ContentType)
	if err != nil {
		return nil, fmt.Errorf("error deriving thumbnail: %s", err)
	}
	m.attachment.Blurhash = thumb.blurhash
	aaaaaaaaaaaaaaaa
}

func (m *Media) PreLoad() {
	m.mu.Lock()
	defer m.mu.Unlock()
}

func (m *Media) Load() {
	m.mu.Lock()
	defer m.mu.Unlock()
}

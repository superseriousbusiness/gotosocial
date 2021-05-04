package message

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

func (p *processor) AdminEmojiCreate(authed *oauth.Auth, form *apimodel.EmojiCreateRequest) (*apimodel.Emoji, error) {
	if !authed.User.Admin {
		return nil, fmt.Errorf("user %s not an admin", authed.User.ID)
	}

	// open the emoji and extract the bytes from it
	f, err := form.Image.Open()
	if err != nil {
		return nil, fmt.Errorf("error opening emoji: %s", err)
	}
	buf := new(bytes.Buffer)
	size, err := io.Copy(buf, f)
	if err != nil {
		return nil, fmt.Errorf("error reading emoji: %s", err)
	}
	if size == 0 {
		return nil, errors.New("could not read provided emoji: size 0 bytes")
	}

	// allow the mediaHandler to work its magic of processing the emoji bytes, and putting them in whatever storage backend we're using
	emoji, err := p.mediaHandler.ProcessLocalEmoji(buf.Bytes(), form.Shortcode)
	if err != nil {
		return nil, fmt.Errorf("error reading emoji: %s", err)
	}

	mastoEmoji, err := p.tc.EmojiToMasto(emoji)
	if err != nil {
		return nil, fmt.Errorf("error converting emoji to mastotype: %s", err)
	}

	if err := p.db.Put(emoji); err != nil {
		return nil, fmt.Errorf("database error while processing emoji: %s", err)
	}

	return &mastoEmoji, nil
}

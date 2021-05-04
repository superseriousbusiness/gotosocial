package message

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

func (p *processor) MediaCreate(authed *oauth.Auth, form *apimodel.AttachmentRequest) (*apimodel.Attachment, error) {
	// First check this user/account is permitted to create media
	// There's no point continuing otherwise.
	if authed.User.Disabled || !authed.User.Approved || !authed.Account.SuspendedAt.IsZero() {
		return nil, errors.New("not authorized to post new media")
	}

	// open the attachment and extract the bytes from it
	f, err := form.File.Open()
	if err != nil {
		return nil, fmt.Errorf("error opening attachment: %s", err)
	}
	buf := new(bytes.Buffer)
	size, err := io.Copy(buf, f)
	if err != nil {
		return nil, fmt.Errorf("error reading attachment: %s", err)

	}
	if size == 0 {
		return nil, errors.New("could not read provided attachment: size 0 bytes")
	}

	// allow the mediaHandler to work its magic of processing the attachment bytes, and putting them in whatever storage backend we're using
	attachment, err := p.mediaHandler.ProcessLocalAttachment(buf.Bytes(), authed.Account.ID)
	if err != nil {
		return nil, fmt.Errorf("error reading attachment: %s", err)
	}

	// now we need to add extra fields that the attachment processor doesn't know (from the form)
	// TODO: handle this inside mediaHandler.ProcessAttachment (just pass more params to it)

	// first description
	attachment.Description = form.Description

	// now parse the focus parameter
	// TODO: tidy this up into a separate function and just return an error so all the c.JSON and return calls are obviated
	var focusx, focusy float32
	if form.Focus != "" {
		spl := strings.Split(form.Focus, ",")
		if len(spl) != 2 {
			return nil, fmt.Errorf("improperly formatted focus %s", form.Focus)
		}
		xStr := spl[0]
		yStr := spl[1]
		if xStr == "" || yStr == "" {
			return nil, fmt.Errorf("improperly formatted focus %s", form.Focus)
		}
		fx, err := strconv.ParseFloat(xStr, 32)
		if err != nil {
			return nil, fmt.Errorf("improperly formatted focus %s: %s", form.Focus, err)
		}
		if fx > 1 || fx < -1 {
			return nil, fmt.Errorf("improperly formatted focus %s", form.Focus)
		}
		focusx = float32(fx)
		fy, err := strconv.ParseFloat(yStr, 32)
		if err != nil {
			return nil, fmt.Errorf("improperly formatted focus %s: %s", form.Focus, err)
		}
		if fy > 1 || fy < -1 {
			return nil, fmt.Errorf("improperly formatted focus %s", form.Focus)
		}
		focusy = float32(fy)
	}
	attachment.FileMeta.Focus.X = focusx
	attachment.FileMeta.Focus.Y = focusy

	// prepare the frontend representation now -- if there are any errors here at least we can bail without
	// having already put something in the database and then having to clean it up again (eugh)
	mastoAttachment, err := p.tc.AttachmentToMasto(attachment)
	if err != nil {
		return nil, fmt.Errorf("error parsing media attachment to frontend type: %s", err)
	}

	// now we can confidently put the attachment in the database
	if err := p.db.Put(attachment); err != nil {
		return nil, fmt.Errorf("error storing media attachment in db: %s", err)
	}

	return &mastoAttachment, nil
}

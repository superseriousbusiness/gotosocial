package generates

import (
	"bytes"
	"context"
	"encoding/base64"
	"strconv"
	"strings"

	"code.superseriousbusiness.org/oauth2/v4"
	"github.com/google/uuid"
)

// NewAccessGenerate create to generate the access token instance
func NewAccessGenerate() *AccessGenerate {
	return &AccessGenerate{}
}

// AccessGenerate generate the access token
type AccessGenerate struct {
}

// Token based on the UUID generated token
func (ag *AccessGenerate) Token(ctx context.Context, data *oauth2.GenerateBasic, isGenRefresh bool) (string, string, error) {
	buf := bytes.NewBufferString(data.Client.GetID())
	buf.WriteString(data.UserID)
	buf.WriteString(strconv.FormatInt(data.CreateAt.UnixNano(), 10))

	access := base64.URLEncoding.EncodeToString([]byte(uuid.NewMD5(uuid.Must(uuid.NewRandom()), buf.Bytes()).String()))
	access = strings.ToUpper(strings.TrimRight(access, "="))
	refresh := ""
	if isGenRefresh {
		refresh = base64.URLEncoding.EncodeToString([]byte(uuid.NewSHA1(uuid.Must(uuid.NewRandom()), buf.Bytes()).String()))
		refresh = strings.ToUpper(strings.TrimRight(refresh, "="))
	}

	return access, refresh, nil
}

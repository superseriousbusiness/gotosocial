package models

import (
	"net/url"
	"time"

	"code.superseriousbusiness.org/oauth2/v4"
)

// NewToken create to token model instance
func NewToken() *Token {
	return &Token{Extension: make(url.Values)}
}

// Token token model
type Token struct {
	ClientID            string        `bson:"ClientID"`
	UserID              string        `bson:"UserID"`
	RedirectURI         string        `bson:"RedirectURI"`
	Scope               string        `bson:"Scope"`
	Code                string        `bson:"Code"`
	CodeChallenge       string        `bson:"CodeChallenge"`
	CodeChallengeMethod string        `bson:"CodeChallengeMethod"`
	CodeCreateAt        time.Time     `bson:"CodeCreateAt"`
	CodeExpiresIn       time.Duration `bson:"CodeExpiresIn"`
	Access              string        `bson:"Access"`
	AccessCreateAt      time.Time     `bson:"AccessCreateAt"`
	AccessExpiresIn     time.Duration `bson:"AccessExpiresIn"`
	Refresh             string        `bson:"Refresh"`
	RefreshCreateAt     time.Time     `bson:"RefreshCreateAt"`
	RefreshExpiresIn    time.Duration `bson:"RefreshExpiresIn"`
	Extension           url.Values    `bson:"Extension"`
}

// New create to token model instance
func (t *Token) New() oauth2.TokenInfo {
	return NewToken()
}

// GetClientID the client id
func (t *Token) GetClientID() string {
	return t.ClientID
}

// SetClientID the client id
func (t *Token) SetClientID(clientID string) {
	t.ClientID = clientID
}

// GetUserID the user id
func (t *Token) GetUserID() string {
	return t.UserID
}

// SetUserID the user id
func (t *Token) SetUserID(userID string) {
	t.UserID = userID
}

// GetRedirectURI redirect URI
func (t *Token) GetRedirectURI() string {
	return t.RedirectURI
}

// SetRedirectURI redirect URI
func (t *Token) SetRedirectURI(redirectURI string) {
	t.RedirectURI = redirectURI
}

// GetScope get scope of authorization
func (t *Token) GetScope() string {
	return t.Scope
}

// SetScope get scope of authorization
func (t *Token) SetScope(scope string) {
	t.Scope = scope
}

// GetCode authorization code
func (t *Token) GetCode() string {
	return t.Code
}

// SetCode authorization code
func (t *Token) SetCode(code string) {
	t.Code = code
}

// GetCodeCreateAt create Time
func (t *Token) GetCodeCreateAt() time.Time {
	return t.CodeCreateAt
}

// SetCodeCreateAt create Time
func (t *Token) SetCodeCreateAt(createAt time.Time) {
	t.CodeCreateAt = createAt
}

// GetCodeExpiresIn the lifetime in seconds of the authorization code
func (t *Token) GetCodeExpiresIn() time.Duration {
	return t.CodeExpiresIn
}

// SetCodeExpiresIn the lifetime in seconds of the authorization code
func (t *Token) SetCodeExpiresIn(exp time.Duration) {
	t.CodeExpiresIn = exp
}

// GetCodeChallenge challenge code
func (t *Token) GetCodeChallenge() string {
	return t.CodeChallenge
}

// SetCodeChallenge challenge code
func (t *Token) SetCodeChallenge(code string) {
	t.CodeChallenge = code
}

// GetCodeChallengeMethod challenge method
func (t *Token) GetCodeChallengeMethod() oauth2.CodeChallengeMethod {
	return oauth2.CodeChallengeMethod(t.CodeChallengeMethod)
}

// SetCodeChallengeMethod challenge method
func (t *Token) SetCodeChallengeMethod(method oauth2.CodeChallengeMethod) {
	t.CodeChallengeMethod = string(method)
}

// GetAccess access Token
func (t *Token) GetAccess() string {
	return t.Access
}

// SetAccess access Token
func (t *Token) SetAccess(access string) {
	t.Access = access
}

// GetAccessCreateAt create Time
func (t *Token) GetAccessCreateAt() time.Time {
	return t.AccessCreateAt
}

// SetAccessCreateAt create Time
func (t *Token) SetAccessCreateAt(createAt time.Time) {
	t.AccessCreateAt = createAt
}

// GetAccessExpiresIn the lifetime in seconds of the access token
func (t *Token) GetAccessExpiresIn() time.Duration {
	return t.AccessExpiresIn
}

// SetAccessExpiresIn the lifetime in seconds of the access token
func (t *Token) SetAccessExpiresIn(exp time.Duration) {
	t.AccessExpiresIn = exp
}

// GetRefresh refresh Token
func (t *Token) GetRefresh() string {
	return t.Refresh
}

// SetRefresh refresh Token
func (t *Token) SetRefresh(refresh string) {
	t.Refresh = refresh
}

// GetRefreshCreateAt create Time
func (t *Token) GetRefreshCreateAt() time.Time {
	return t.RefreshCreateAt
}

// SetRefreshCreateAt create Time
func (t *Token) SetRefreshCreateAt(createAt time.Time) {
	t.RefreshCreateAt = createAt
}

// GetRefreshExpiresIn the lifetime in seconds of the refresh token
func (t *Token) GetRefreshExpiresIn() time.Duration {
	return t.RefreshExpiresIn
}

// SetRefreshExpiresIn the lifetime in seconds of the refresh token
func (t *Token) SetRefreshExpiresIn(exp time.Duration) {
	t.RefreshExpiresIn = exp
}

// GetExtension extension of token
func (t *Token) GetExtension() url.Values {
	return t.Extension
}

// SetExtension set extension of token
func (t *Token) SetExtension(e url.Values) {
	t.Extension = e
}

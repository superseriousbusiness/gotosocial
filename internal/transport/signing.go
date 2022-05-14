package transport

import (
	"github.com/go-fed/httpsig"
)

var (
	// http signer preferences
	prefs       = []httpsig.Algorithm{httpsig.RSA_SHA256}
	digestAlgo  = httpsig.DigestSha256
	getHeaders  = []string{httpsig.RequestTarget, "host", "date"}
	postHeaders = []string{httpsig.RequestTarget, "host", "date", "digest"}
)

// NewGETSigner returns a new httpsig.Signer instance initialized with GTS GET preferences.
func NewGETSigner(expiresIn int64) (httpsig.Signer, error) {
	sig, _, err := httpsig.NewSigner(prefs, digestAlgo, getHeaders, httpsig.Signature, expiresIn)
	return sig, err
}

// NewPOSTSigner returns a new httpsig.Signer instance initialized with GTS POST preferences.
func NewPOSTSigner(expiresIn int64) (httpsig.Signer, error) {
	sig, _, err := httpsig.NewSigner(prefs, digestAlgo, postHeaders, httpsig.Signature, expiresIn)
	return sig, err
}

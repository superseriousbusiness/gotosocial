// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sessions

import "net/http"

// newCookieFromOptions returns an http.Cookie with the options set.
func newCookieFromOptions(name, value string, options *Options) *http.Cookie {
	return &http.Cookie{
		Name:        name,
		Value:       value,
		Path:        options.Path,
		Domain:      options.Domain,
		MaxAge:      options.MaxAge,
		Secure:      options.Secure,
		HttpOnly:    options.HttpOnly,
		Partitioned: options.Partitioned,
		SameSite:    options.SameSite,
	}

}

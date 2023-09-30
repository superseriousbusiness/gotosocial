// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package web

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	robotsPath          = "/robots.txt"
	robotsMetaAllowSome = "nofollow, noarchive, nositelinkssearchbox, max-image-preview:standard" // https://developers.google.com/search/docs/crawling-indexing/robots-meta-tag#robotsmeta
	robotsTxt           = `# GoToSocial robots.txt -- to edit, see internal/web/robots.go
# More info @ https://developers.google.com/search/docs/crawling-indexing/robots/intro

# Before we commence, a giant fuck you to ChatGPT in particular.
# https://platform.openai.com/docs/gptbot
User-agent: GPTBot
Disallow: /

# As of September 2023, GPTBot and ChatGPT-User are equivalent. But there's no telling
# when OpenAI might decide to change that, so block this one too.
User-agent: ChatGPT-User
Disallow: /

# And a giant fuck you to Google Bard and their other generative AI ventures too.
# https://developers.google.com/search/docs/crawling-indexing/overview-google-crawlers
User-agent: Google-Extended
Disallow: /

# Block CommonCrawl. Used in training LLMs and specifically GPT-3.
# https://commoncrawl.org/faq
User-agent: CCBot
Disallow: /

# Block Omgilike/Webz.io, a "Big Web Data" engine.
# https://webz.io/blog/web-data/what-is-the-omgili-bot-and-why-is-it-crawling-your-website/
User-agent: Omgilibot
Disallow: /

# Block Faceboobot, because Meta.
# https://developers.facebook.com/docs/sharing/bot
User-agent: FacebookBot
Disallow: /

# Rules for everything else.
User-agent: *
Crawl-delay: 500

# API endpoints.
Disallow: /api/

# Auth/login endpoints.
Disallow: /auth/
Disallow: /oauth/
Disallow: /check_your_email
Disallow: /wait_for_approval
Disallow: /account_disabled

# Well-known endpoints.
Disallow: /.well-known/

# Fileserver/media.
Disallow: /fileserver/

# Fedi S2S API endpoints.
Disallow: /users/
Disallow: /emoji/

# Settings panels.
Disallow: /admin
Disallow: /user
Disallow: /settings/

# Domain blocklist.
Disallow: /about/suspended`
)

// robotsGETHandler returns a decent robots.txt that prevents crawling
// the api, auth pages, settings pages, etc.
//
// More granular robots meta tags are then applied for web pages
// depending on user preferences (see internal/web).
func (m *Module) robotsGETHandler(c *gin.Context) {
	c.String(http.StatusOK, robotsTxt)
}

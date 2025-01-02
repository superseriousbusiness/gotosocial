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

# AI scrapers and the like.
# https://github.com/ai-robots-txt/ai.robots.txt/
User-agent: AI2Bot
User-agent: Ai2Bot-Dolma
User-agent: Amazonbot
User-agent: anthropic-ai
User-agent: Applebot
User-agent: Applebot-Extended
User-agent: Bytespider
User-agent: CCBot
User-agent: ChatGPT-User
User-agent: ClaudeBot
User-agent: Claude-Web
User-agent: cohere-ai
User-agent: cohere-training-data-crawler
User-agent: Diffbot
User-agent: DuckAssistBot
User-agent: FacebookBot
User-agent: FriendlyCrawler
User-agent: Google-Extended
User-agent: GoogleOther
User-agent: GoogleOther-Image
User-agent: GoogleOther-Video
User-agent: GPTBot
User-agent: iaskspider/2.0
User-agent: ICC-Crawler
User-agent: ImagesiftBot
User-agent: img2dataset
User-agent: ISSCyberRiskCrawler
User-agent: Kangaroo Bot
User-agent: Meta-ExternalAgent
User-agent: Meta-ExternalFetcher
User-agent: OAI-SearchBot
User-agent: omgili
User-agent: omgilibot
User-agent: PanguBot
User-agent: PerplexityBot
User-agent: PetalBot
User-agent: Scrapy
User-agent: Sidetrade indexer bot
User-agent: Timpibot
User-agent: VelenPublicWebCrawler
User-agent: Webzio-Extended
User-agent: YouBot
Disallow: /

# Marketing/SEO "intelligence" data scrapers
User-agent: AwarioRssBot
User-agent: AwarioSmartBot
User-agent: DataForSeoBot
User-agent: magpie-crawler
User-agent: Meltwater
User-agent: peer39_crawler
User-agent: peer39_crawler/1.0
User-agent: PiplBot
User-agent: scoop.it
User-agent: Seekr
Disallow: /

# Well-known.dev crawler. Indexes stuff under /.well-known.
# https://well-known.dev/about/
User-agent: WellKnownBot     
Disallow: /   

# Rules for everything else.
User-agent: *
Crawl-delay: 500

# API endpoints.
Disallow: /api/

# Auth/Sign in endpoints.
Disallow: /auth/
Disallow: /oauth/
Disallow: /check_your_email
Disallow: /wait_for_approval
Disallow: /account_disabled
Disallow: /signup

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

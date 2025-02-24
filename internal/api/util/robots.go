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

package util

// See:
//
//   - https://developers.google.com/search/docs/crawling-indexing/robots-meta-tag#robotsmeta
//   - https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Robots-Tag
//   - https://www.rfc-editor.org/rfc/rfc9309.html
const (
	RobotsDirectivesDisallow  = "noindex, nofollow"
	RobotsDirectivesAllowSome = "nofollow, noarchive, nositelinkssearchbox, max-image-preview:standard"
	RobotsTxt                 = `# GoToSocial robots.txt -- to edit, see internal/api/util/robots.go
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
Disallow: /about/suspended

# Webfinger endpoint.
Disallow: /.well-known/webfinger
`
	RobotsTxtDisallowNodeInfo = RobotsTxt + `
# Disallow nodeinfo
Disallow: /.well-known/nodeinfo
Disallow: /nodeinfo/
`

	// MD5 hash of basic robots.txt.
	RobotsTxtETag = `ce6729aacbb16fae3628210c04b462b7`
	// MD5 hash of robots.txt with NodeInfo disallowed.
	RobotsTxtDisallowNodeInfoETag = `a1e4ce6342978bc8d6c3e3dfab07cab4`
)

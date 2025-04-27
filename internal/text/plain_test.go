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

package text_test

import (
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/text"
	"github.com/stretchr/testify/suite"
)

const (
	simple                     = "this is a plain and simple status"
	simpleExpected             = "<p>this is a plain and simple status</p>"
	simpleExpectedNoParagraph  = "this is a plain and simple status"
	withTag                    = "here's a simple status that uses hashtag #welcome!"
	withTagExpected            = "<p>here's a simple status that uses hashtag <a href=\"http://localhost:8080/tags/welcome\" class=\"mention hashtag\" rel=\"tag nofollow noreferrer noopener\" target=\"_blank\">#<span>welcome</span></a>!</p>"
	withTagExpectedNoParagraph = "here's a simple status that uses hashtag <a href=\"http://localhost:8080/tags/welcome\" class=\"mention hashtag\" rel=\"tag nofollow noreferrer noopener\" target=\"_blank\">#<span>welcome</span></a>!"
	withHTML                   = "<div>blah this should just be html escaped blah</div>"
	withHTMLExpected           = "<p>&lt;div>blah this should just be html escaped blah&lt;/div></p>"
	moreComplex                = "Another test @foss_satan@fossbros-anonymous.io\n\n#Hashtag\n\nText\n\n:rainbow:"
	moreComplexExpected        = "<p>Another test <span class=\"h-card\"><a href=\"http://fossbros-anonymous.io/@foss_satan\" class=\"u-url mention\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">@<span>foss_satan</span></a></span><br><br><a href=\"http://localhost:8080/tags/hashtag\" class=\"mention hashtag\" rel=\"tag nofollow noreferrer noopener\" target=\"_blank\">#<span>Hashtag</span></a><br><br>Text<br><br>:rainbow:</p>"
	withUTF8Link               = "here's a link with utf-8 characters in it: https://example.org/söme_url"
	withUTF8LinkExpected       = "<p>here's a link with utf-8 characters in it: <a href=\"https://example.org/s%C3%B6me_url\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">https://example.org/söme_url</a></p>"
	withFunkyTags              = "#hashtag1 pee #hashtag2\u200Bpee #hashtag3|poo #hashtag4\uFEFFpoo"
	withFunkyTagsExpected      = "<p><a href=\"http://localhost:8080/tags/hashtag1\" class=\"mention hashtag\" rel=\"tag nofollow noreferrer noopener\" target=\"_blank\">#<span>hashtag1</span></a> pee <a href=\"http://localhost:8080/tags/hashtag2\" class=\"mention hashtag\" rel=\"tag nofollow noreferrer noopener\" target=\"_blank\">#<span>hashtag2</span></a>\u200bpee <a href=\"http://localhost:8080/tags/hashtag3\" class=\"mention hashtag\" rel=\"tag nofollow noreferrer noopener\" target=\"_blank\">#<span>hashtag3</span></a>|poo <a href=\"http://localhost:8080/tags/hashtag4\" class=\"mention hashtag\" rel=\"tag nofollow noreferrer noopener\" target=\"_blank\">#<span>hashtag4</span></a>\ufeffpoo</p>"
)

type PlainTestSuite struct {
	TextStandardTestSuite
}

func (suite *PlainTestSuite) TestParseSimple() {
	formatted := suite.FromPlain(simple)
	suite.Equal(simpleExpected, formatted.HTML)
}

func (suite *PlainTestSuite) TestParseSimpleNoParagraph() {
	formatted := suite.FromPlainNoParagraph(simple)
	suite.Equal(simpleExpectedNoParagraph, formatted.HTML)
}

func (suite *PlainTestSuite) TestParseWithTag() {
	formatted := suite.FromPlain(withTag)
	suite.Equal(withTagExpected, formatted.HTML)
}

func (suite *PlainTestSuite) TestParseWithTagNoParagraph() {
	formatted := suite.FromPlainNoParagraph(withTag)
	suite.Equal(withTagExpectedNoParagraph, formatted.HTML)
}

func (suite *PlainTestSuite) TestParseWithHTML() {
	formatted := suite.FromPlain(withHTML)
	suite.Equal(withHTMLExpected, formatted.HTML)
}

func (suite *PlainTestSuite) TestParseMoreComplex() {
	formatted := suite.FromPlain(moreComplex)
	suite.Equal(moreComplexExpected, formatted.HTML)
}

func (suite *PlainTestSuite) TestWithUTF8Link() {
	formatted := suite.FromPlain(withUTF8Link)
	suite.Equal(withUTF8LinkExpected, formatted.HTML)
}

func (suite *PlainTestSuite) TestLinkNoMention() {
	statusText := `here's a link to a post by zork

https://example.com/@the_mighty_zork/statuses/01FGVP55XMF2K6316MQRX6PFG1

that link shouldn't come out formatted as a mention!`

	menchies := suite.FromPlain(statusText).Mentions
	suite.Empty(menchies)
}

func (suite *PlainTestSuite) TestDeriveMentionsEmpty() {
	statusText := ``
	menchies := suite.FromPlain(statusText).Mentions
	suite.Len(menchies, 0)
}

func (suite *PlainTestSuite) TestDeriveHashtagsOK() {
	statusText := `weeeeeeee #testing123 #also testing

# testing this one shouldn't work

			#thisshouldwork #dupe #dupe!! #dupe

	here's a link with a fragment: https://example.org/whatever#ahhh
	here's another link with a fragment: https://example.org/whatever/#ahhh

(#ThisShouldAlsoWork) #this_should_not_be_split

#__ <- just underscores, shouldn't work

#111111 thisalsoshouldn'twork#### ##

#alimentación, #saúde, #lävistää, #ö, #네
#ThisOneIsOneHundredAndOneCharactersLongWhichIsReallyJustWayWayTooLongDefinitelyLongerThanYouWouldNeed...
#ThisOneIsThirteyCharactersLong
`

	tags := suite.FromPlain(statusText).Tags
	if suite.Len(tags, 12) {
		suite.Equal("testing123", tags[0].Name)
		suite.Equal("also", tags[1].Name)
		suite.Equal("thisshouldwork", tags[2].Name)
		suite.Equal("dupe", tags[3].Name)
		suite.Equal("ThisShouldAlsoWork", tags[4].Name)
		suite.Equal("this_should_not_be_split", tags[5].Name)
		suite.Equal("alimentación", tags[6].Name)
		suite.Equal("saúde", tags[7].Name)
		suite.Equal("lävistää", tags[8].Name)
		suite.Equal("ö", tags[9].Name)
		suite.Equal("네", tags[10].Name)
		suite.Equal("ThisOneIsThirteyCharactersLong", tags[11].Name)
	}

	statusText = `#올빼미 hej`
	tags = suite.FromPlain(statusText).Tags
	suite.Equal("올빼미", tags[0].Name)
}

func (suite *PlainTestSuite) TestFunkyTags() {
	formatted := suite.FromPlain(withFunkyTags)
	suite.Equal(withFunkyTagsExpected, formatted.HTML)

	tags := formatted.Tags
	suite.Equal("hashtag1", tags[0].Name)
	suite.Equal("hashtag2", tags[1].Name)
	suite.Equal("hashtag3", tags[2].Name)
	suite.Equal("hashtag4", tags[3].Name)
}

func (suite *PlainTestSuite) TestDeriveMultiple() {
	statusText := `Another test @foss_satan@fossbros-anonymous.io

	#Hashtag

	Text`

	f := suite.FromPlain(statusText)

	suite.Len(f.Mentions, 1)
	suite.Equal("@foss_satan@fossbros-anonymous.io", f.Mentions[0].NameString)

	suite.Len(f.Tags, 1)
	suite.Equal("hashtag", f.Tags[0].Name)

	suite.Len(f.Emojis, 0)
}

func (suite *PlainTestSuite) TestZalgoHashtag() {
	statusText := `yo who else loves #praying to #z̸͉̅a̸͚͋l̵͈̊g̸̫͌ỏ̷̪?`
	f := suite.FromPlain(statusText)
	if suite.Len(f.Tags, 2) {
		suite.Equal("praying", f.Tags[0].Name)
		// NFC doesn't do much for Zalgo text, but it's difficult to strip marks without affecting non-Latin text.
		suite.Equal("z̸͉̅a̸͚͋l̵͈̊g̸̫͌ỏ̷̪", f.Tags[1].Name)
	}
}

func (suite *PlainTestSuite) TestNumbersAreNotHashtags() {
	statusText := `yo who else thinks #19_98 is #1?`
	f := suite.FromPlain(statusText)
	suite.Len(f.Tags, 0)
}

func (suite *PlainTestSuite) TestParseHTMLToPlain() {
	for _, t := range []struct {
		html          string
		expectedPlain string
	}{
		{
			// Check newlines between paras preserved.
			html: "<p>butting into a serious discussion about programming languages*: \"elixir? I barely know 'er! honk honk!\"</p><p><small>*insofar as any discussion about programming languages can truly be considered \"serious\" since programmers are fucking clowns</small></p>",
			expectedPlain: `butting into a serious discussion about programming languages*: "elixir? I barely know 'er! honk honk!"

*insofar as any discussion about programming languages can truly be considered "serious" since programmers are fucking clowns`,
		},
		{
			// This one looks a bit wacky but nobody should
			// be putting definition lists in summaries *really*.
			html:          "<dl class=\"status-stats\"><div class=\"stats-grouping\"><div class=\"stats-item published-at text-cutoff\"><dt class=\"sr-only\">Published</dt><dd><time class=\"dt-published\" datetime=\"2025-01-15T23:49:59.299Z\">Jan 16, 2025, 00:49</time></dd></div><div class=\"stats-grouping\"><div class=\"stats-item\" title=\"Replies\"><dt><span class=\"sr-only\">Replies</span><i class=\"fa fa-reply-all\" aria-hidden=\"true\"></i></dt><dd>0</dd></div><div class=\"stats-item\" title=\"Faves\"><dt><span class=\"sr-only\">Favourites</span><i class=\"fa fa-star\" aria-hidden=\"true\"></i></dt><dd>4</dd></div><div class=\"stats-item\" title=\"Boosts\"><dt><span class=\"sr-only\">Reblogs</span><i class=\"fa fa-retweet\" aria-hidden=\"true\"></i></dt><dd>0</dd></div></div></div><div class=\"stats-item language\" title=\"English\"><dt class=\"sr-only\">Language</dt><dd><span class=\"sr-only\">English</span><span aria-hidden=\"true\">en</span></dd></div></dl>",
			expectedPlain: `PublishedJan 16, 2025, 00:49Replies0Favourites4Reblogs0LanguageEnglishen`,
		},
		{
			// Check <br> converted to newlines and leading / trailing space removed.
			html: "     <p>i'm a milf,<br>i'm a lover,<br>do your mom,<br>do your brother</p><p>i'm a sinner,<br>i'm a saint,<br>i will not be ashamed!</p><br>    <br>",
			expectedPlain: `i'm a milf,
i'm a lover,
do your mom,
do your brother

i'm a sinner,
i'm a saint,
i will not be ashamed!`,
		},
		{
			// Check newlines, links, lists still more or less readable as such.
			html: "<p>Hello everyone, after a week or two down the release candidate mines, we've emerged blinking into the light carrying with us <a href=\"https://gts.superseriousbusiness.org/tags/gotosocial\" class=\"mention hashtag\" rel=\"tag nofollow noreferrer noopener\" target=\"_blank\">#<span>GoToSocial</span></a> <strong>v0.18.0 Scroingly Sloth</strong>!</p><p><a href=\"https://codeberg.org/superseriousbusiness/gotosocial/releases/tag/v0.18.0\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">https://codeberg.org/superseriousbusiness/gotosocial/releases/tag/v0.18.0</a></p><p>Please read the migration notes carefully for instructions on how to upgrade to this version. <strong>This version contains several very long migrations so you will need to be patient when upgrading, and backup your database first!!</strong></p><p><strong>Release highlights</strong></p><ul><li><strong>Status edit support</strong>: one of our most-requested features! You can now edit your own statuses, and see instance edit history from other accounts too (if your instance has them stored).</li><li><strong>Push notifications</strong>: probably the second most-requested feature! GoToSocial can now send push notifications to clients via their configured push providers.<br>You may need to uninstall / reinstall client applications, or log out and back in again, for this feature to work. (And if you're using Tusky, <a href=\"https://tusky.app/faq/#why-are-notifications-less-frequent-with-tusky\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">make sure you've got ntfy installed</a>).</li><li><strong>Global instance css customization</strong>: admins can now apply custom CSS across their entire instance via the settings panel.</li><li><strong>Domain permission subscriptions</strong>: it's now possible to configure your instance to subscribe to CSV, JSON, or plaintext lists of domain permissions.<br>Each night, your instance will fetch and automatically create domain permissions (or permission drafts) based on what it finds in a subscribed list.<br>See the <a href=\"https://docs.gotosocial.org/en/latest/admin/domain_permission_subscriptions/\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">domain permission subscription documentation</a> for more information.</li><li><strong>Trusted-proxies helper</strong>: instances with improperly configured trusted-proxies settings will now show a warning on the homepage, so admins can make sure their instance is configured correctly. Check your own instance homepage after updating to see if you need to do anything.</li><li><strong>Better outbox sorting</strong>: messages from GoToSocial are now delivered more quickly to people you mention, so conversations across instances should feel a bit snappier.</li><li><strong>Log in button</strong>: there's now a login button in the top right of the instance homepage, which leads to a helpful page about clients, with a link to the settings panel. Should make things less confusing for new users!</li><li><strong>Granular stats controls</strong>: with the <code>instance-stats-mode</code> setting, admins can now choose if and how their instance serves stats via the nodeinfo endpoints. Existing behavior from v0.17.0 is the default.</li><li><strong>Post backdating</strong>: via the API you can now backdate posts (if enabled in config.yaml). This is our first step towards making it possible to import your post history from elsewhere into your GoToSocial instance. While there's no way to do this in the settings panel yet, you can already use third-party tools like Slurp to import posts from a Mastodon export (see <a href=\"https://github.com/VyrCossont/slurp\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">Slurp</a>).</li><li><strong>Configurable sign-up limits</strong>: you can now configure your sign-up backlog length and sign-up throttling (defaults remain the same).</li><li><strong>NetBSD and FreeBSD builds</strong>: yep!</li><li><strong>Respect users <code>prefers-color-scheme</code> preference</strong>: there's now a light mode default theme to complement our trusty dark mode theme, and the theme will switch based on a visitor's <code>prefers-color-scheme</code> configuration. This applies to all page and profiles, with the exception of some custom themes. Works in the settings panel too!</li></ul><p>Thanks for reading! And seriously back up your database.</p>",
			expectedPlain: `Hello everyone, after a week or two down the release candidate mines, we've emerged blinking into the light carrying with us #GoToSocial <https://gts.superseriousbusiness.org/tags/gotosocial> v0.18.0 Scroingly Sloth!

https://codeberg.org/superseriousbusiness/gotosocial/releases/tag/v0.18.0 <https://codeberg.org/superseriousbusiness/gotosocial/releases/tag/v0.18.0>

Please read the migration notes carefully for instructions on how to upgrade to this version. This version contains several very long migrations so you will need to be patient when upgrading, and backup your database first!!

Release highlights


 - Status edit support: one of our most-requested features! You can now edit your own statuses, and see instance edit history from other accounts too (if your instance has them stored).
 - Push notifications: probably the second most-requested feature! GoToSocial can now send push notifications to clients via their configured push providers.
You may need to uninstall / reinstall client applications, or log out and back in again, for this feature to work. (And if you're using Tusky, make sure you've got ntfy installed <https://tusky.app/faq/#why-are-notifications-less-frequent-with-tusky>).
 - Global instance css customization: admins can now apply custom CSS across their entire instance via the settings panel.
 - Domain permission subscriptions: it's now possible to configure your instance to subscribe to CSV, JSON, or plaintext lists of domain permissions.
Each night, your instance will fetch and automatically create domain permissions (or permission drafts) based on what it finds in a subscribed list.
See the domain permission subscription documentation <https://docs.gotosocial.org/en/latest/admin/domain_permission_subscriptions/> for more information.
 - Trusted-proxies helper: instances with improperly configured trusted-proxies settings will now show a warning on the homepage, so admins can make sure their instance is configured correctly. Check your own instance homepage after updating to see if you need to do anything.
 - Better outbox sorting: messages from GoToSocial are now delivered more quickly to people you mention, so conversations across instances should feel a bit snappier.
 - Log in button: there's now a login button in the top right of the instance homepage, which leads to a helpful page about clients, with a link to the settings panel. Should make things less confusing for new users!
 - Granular stats controls: with the instance-stats-mode setting, admins can now choose if and how their instance serves stats via the nodeinfo endpoints. Existing behavior from v0.17.0 is the default.
 - Post backdating: via the API you can now backdate posts (if enabled in config.yaml). This is our first step towards making it possible to import your post history from elsewhere into your GoToSocial instance. While there's no way to do this in the settings panel yet, you can already use third-party tools like Slurp to import posts from a Mastodon export (see Slurp <https://github.com/VyrCossont/slurp>).
 - Configurable sign-up limits: you can now configure your sign-up backlog length and sign-up throttling (defaults remain the same).
 - NetBSD and FreeBSD builds: yep!
 - Respect users prefers-color-scheme preference: there's now a light mode default theme to complement our trusty dark mode theme, and the theme will switch based on a visitor's prefers-color-scheme configuration. This applies to all page and profiles, with the exception of some custom themes. Works in the settings panel too!


Thanks for reading! And seriously back up your database.`,
		},
	} {
		plain := text.ParseHTMLToPlain(t.html)
		suite.Equal(t.expectedPlain, plain)
	}
}

func (suite *PlainTestSuite) TestStripCaption1() {
	dodgyCaption := "<script>console.log('haha!')</script>this is just a normal caption ;)"
	stripped := text.StripHTMLFromText(dodgyCaption)
	suite.Equal("this is just a normal caption ;)", stripped)
}

func (suite *PlainTestSuite) TestStripCaption2() {
	dodgyCaption := "<em>here's a LOUD caption</em>"
	stripped := text.StripHTMLFromText(dodgyCaption)
	suite.Equal("here's a LOUD caption", stripped)
}

func (suite *PlainTestSuite) TestStripCaption3() {
	dodgyCaption := ""
	stripped := text.StripHTMLFromText(dodgyCaption)
	suite.Equal("", stripped)
}

func (suite *PlainTestSuite) TestStripCaption4() {
	dodgyCaption := `


here is
a multi line
caption
with some newlines



`
	stripped := text.StripHTMLFromText(dodgyCaption)
	suite.Equal("here is\na multi line\ncaption\nwith some newlines", stripped)
}

func (suite *PlainTestSuite) TestStripCaption5() {
	// html-escaped: "<script>console.log('aha!')</script> hello world"
	dodgyCaption := `&lt;script&gt;console.log(&apos;aha!&apos;)&lt;/script&gt; hello world`
	stripped := text.StripHTMLFromText(dodgyCaption)
	suite.Equal("hello world", stripped)
}

func (suite *PlainTestSuite) TestStripCaption6() {
	// html-encoded: "<script>console.log('aha!')</script> hello world"
	dodgyCaption := `&lt;&#115;&#99;&#114;&#105;&#112;&#116;&gt;&#99;&#111;&#110;&#115;&#111;&#108;&#101;&period;&#108;&#111;&#103;&lpar;&apos;&#97;&#104;&#97;&excl;&apos;&rpar;&lt;&sol;&#115;&#99;&#114;&#105;&#112;&#116;&gt;&#32;&#104;&#101;&#108;&#108;&#111;&#32;&#119;&#111;&#114;&#108;&#100;`
	stripped := text.StripHTMLFromText(dodgyCaption)
	suite.Equal("hello world", stripped)
}

func (suite *PlainTestSuite) TestStripCustomCSS() {
	customCSS := `.toot .username {
	color: var(--link_fg);
	line-height: 2rem;
	margin-top: -0.5rem;
	align-self: start;
	
	white-space: nowrap;
	overflow: hidden;
	text-overflow: ellipsis;
}`
	stripped := text.StripHTMLFromText(customCSS)
	suite.Equal(customCSS, stripped) // should be the same as it was before
}

func (suite *PlainTestSuite) TestStripNaughtyCustomCSS1() {
	// try to break out of <style> into <head> and change the document title
	customCSS := "</style><title>pee pee poo poo</title><style>"
	stripped := text.StripHTMLFromText(customCSS)
	suite.Empty(stripped)
}

func (suite *PlainTestSuite) TestStripNaughtyCustomCSS2() {
	// try to break out of <style> into <head> and change the document title
	customCSS := "pee pee poo poo</style><title></title><style>"
	stripped := text.StripHTMLFromText(customCSS)
	suite.Equal("pee pee poo poo", stripped)
}

func TestPlainTestSuite(t *testing.T) {
	suite.Run(t, new(PlainTestSuite))
}

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

package spam

import (
	"context"
	"errors"
	"net/url"
	"slices"
	"strings"

	"code.superseriousbusiness.org/gotosocial/internal/ap"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtscontext"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"code.superseriousbusiness.org/gotosocial/internal/regexes"
	"code.superseriousbusiness.org/gotosocial/internal/util"
	"github.com/miekg/dns"
)

// preppedMention represents a partially-parsed
// mention, prepared for spam checking purposes.
type preppedMention struct {
	*gtsmodel.Mention
	uri    *url.URL
	domain string
	user   string
	local  bool
}

// StatusableOK returns no error if the given statusable looks OK,
// ie., relevant to the receiver, and not spam.
//
// This should only be used for Creates of statusables, NOT Announces!
//
// If the statusable does not pass relevancy or spam checks, either
// a Spam or NotRelevant error will be returned. Callers should use
// gtserror.IsSpam() and gtserror.IsNotRelevant() to check for this.
//
// If the returned error is not nil, but neither Spam or NotRelevant,
// then it's an actual database error.
//
// The decision is made based on the following heuristics, in order:
//
//  1. Receiver follow requester. Return nil.
//  2. Statusable doesn't mention receiver. Return NotRelevant.
//
// If instance-federation-spam-filter = false, then return nil now.
// Otherwise check:
//
//  3. Receiver is locked and is followed by requester. Return nil.
//  4. Five or more people are mentioned. Return Spam.
//  5. Receiver follow (requests) a mentioned account. Return nil.
//  6. Statusable has a media attachment. Return Spam.
//  7. Statusable contains non-mention, non-hashtag links. Return Spam.
func (f *Filter) StatusableOK(
	ctx context.Context,
	receiver *gtsmodel.Account,
	requester *gtsmodel.Account,
	statusable ap.Statusable,
) error {
	// HEURISTIC 1: Check whether receiving account follows the requesting account.
	// If so, we know it's OK and don't need to do any other checks.
	follows, err := f.state.DB.IsFollowing(ctx, receiver.ID, requester.ID)
	if err != nil {
		return gtserror.Newf("db error checking follow status: %w", err)
	}

	if follows {
		// Looks fine.
		return nil
	}

	// HEURISTIC 2: Check whether statusable mentions the
	// receiver. If not, we don't want to process this message.
	rawMentions, _ := ap.ExtractMentions(statusable)
	mentions := prepMentions(ctx, rawMentions)
	mentioned := f.isMentioned(ctx, receiver, mentions)
	if !mentioned {
		// This is a random message fired
		// into our inbox, just drop it.
		err := errors.New("receiver does not follow requester, and is not mentioned")
		return gtserror.SetNotRelevant(err)
	}

	// Receiver is mentioned, but not by someone
	// they follow. Check if we need to do more
	// granular spam filtering.
	if !config.GetInstanceFederationSpamFilter() {
		// Filter is not enabled, allow it
		// through without further checks.
		return nil
	}

	// More granular spam filtering time!
	//
	// HEURISTIC 3: Does requester follow locked receiver?
	followedBy, err := f.lockedFollowedBy(ctx, receiver, requester)
	if err != nil {
		return gtserror.Newf("db error checking follow status: %w", err)
	}

	// If receiver is locked, and is followed
	// by requester, this likely means they're
	// interested in the message. Allow it.
	if followedBy {
		return nil
	}

	// HEURISTIC 4: How many people are mentioned?
	// If it's 5 or more we can assume this is spam.
	mentionsLen := len(mentions)
	if mentionsLen >= 5 {
		err := errors.New("status mentions 5 or more people")
		return gtserror.SetSpam(err)
	}

	// HEURISTIC 5: Four or fewer people are mentioned,
	// do we follow (request) at least one of them?
	// If so, we're probably interested in the message.
	knowsOne := f.knowsOneMentioned(ctx, receiver, mentions)
	if knowsOne {
		return nil
	}

	// HEURISTIC 6: Are there any media attachments?
	attachments, err := ap.ExtractAttachments(statusable)
	if err != nil {
		log.Warnf(ctx,
			"error(s) extracting attachments for %s: %v",
			ap.GetJSONLDId(statusable), err,
		)
	}

	hasAttachments := len(attachments) != 0
	if hasAttachments {
		err := errors.New("status has attachment(s)")
		return gtserror.SetSpam(err)
	}

	// HEURISTIC 7: Are there any links in the post
	// aside from mentions and hashtags? Include the
	// summary/content warning when checking.
	hashtags, _ := ap.ExtractHashtags(statusable)
	hasErrantLinks := f.errantLinks(ctx, statusable, mentions, hashtags)
	if hasErrantLinks {
		err := errors.New("status has one or more non-mention, non-hashtag links")
		return gtserror.SetSpam(err)
	}

	// Looks OK.
	return nil
}

// prepMentions prepares a slice of mentions
// for spam checking by parsing out the namestring
// and targetAccountURI values, if present.
func prepMentions(
	ctx context.Context,
	mentions []*gtsmodel.Mention,
) []preppedMention {
	var (
		host          = config.GetHost()
		accountDomain = config.GetAccountDomain()
	)

	parsedMentions := make([]preppedMention, 0, len(mentions))
	for _, mention := range mentions {
		// Start by just embedding
		// the original mention.
		parsedMention := preppedMention{
			Mention: mention,
		}

		// Try to parse namestring if present.
		if mention.NameString != "" {
			user, domain, err := util.ExtractNamestringParts(mention.NameString)
			if err != nil {
				// Malformed mention,
				// just log + ignore.
				log.Debugf(ctx,
					"malformed mention namestring: %v",
					err,
				)
				continue
			}

			parsedMention.domain = domain
			parsedMention.user = user
		}

		// Try to parse URI if present.
		if mention.TargetAccountURI != "" {
			targetURI, err := url.Parse(mention.TargetAccountURI)
			if err != nil {
				// Malformed mention,
				// just log + ignore.
				log.Debugf(ctx,
					"malformed mention uri: %v",
					err,
				)
				continue
			}

			parsedMention.uri = targetURI

			// Set host from targetURI if
			// it wasn't set by namestring.
			if parsedMention.domain == "" {
				parsedMention.domain = targetURI.Host
			}
		}

		// It's a mention of a local account if the target host is us.
		parsedMention.local = parsedMention.domain == host || parsedMention.domain == accountDomain

		// Done with this one.
		parsedMentions = append(parsedMentions, parsedMention)
	}

	return parsedMentions
}

// isMentioned returns true if the
// receiver is targeted by at least
// one of the given mentions.
func (f *Filter) isMentioned(
	ctx context.Context,
	receiver *gtsmodel.Account,
	mentions []preppedMention,
) bool {
	return slices.ContainsFunc(
		mentions,
		func(mention preppedMention) bool {
			// Check if receiver mentioned by URI.
			if accURI := mention.TargetAccountURI; accURI != "" &&
				(accURI == receiver.URI || accURI == receiver.URL) {
				return true
			}

			// Check if receiver mentioned by namestring.
			if mention.local && strings.EqualFold(mention.user, receiver.Username) {
				return true
			}

			// Mention doesn't
			// target receiver.
			return false
		},
	)
}

// lockedFollowedBy returns true
// if receiver account is locked,
// and requester follows receiver.
func (f *Filter) lockedFollowedBy(
	ctx context.Context,
	receiver *gtsmodel.Account,
	requester *gtsmodel.Account,
) (bool, error) {
	// If receiver is not locked,
	// return early to avoid a db call.
	if !*receiver.Locked {
		return false, nil
	}

	return f.state.DB.IsFollowing(ctx, requester.ID, receiver.ID)
}

// knowsOneMentioned returns true if the
// receiver follows or has follow requested
// at least one of the mentioned accounts.
func (f *Filter) knowsOneMentioned(
	ctx context.Context,
	receiver *gtsmodel.Account,
	mentions []preppedMention,
) bool {
	return slices.ContainsFunc(
		mentions,
		func(mention preppedMention) bool {
			var (
				acc *gtsmodel.Account
				err error
			)

			// Try to get target account without
			// dereffing. After all, if they're not
			// in our db we definitely don't know them.
			if mention.TargetAccountURI != "" {
				acc, err = f.state.DB.GetAccountByURI(
					gtscontext.SetBarebones(ctx),
					mention.TargetAccountURI,
				)
			} else if mention.user != "" {
				acc, err = f.state.DB.GetAccountByUsernameDomain(
					gtscontext.SetBarebones(ctx),
					mention.user,
					mention.domain,
				)
			}

			if err != nil && !errors.Is(err, db.ErrNoEntries) {
				// Proper error.
				log.Errorf(ctx, "db error getting mentioned account: %v", err)
				return false
			}

			if acc == nil {
				// We don't know this nerd!
				return false
			}

			if acc.ID == receiver.ID {
				// This is us, doesn't count.
				return false
			}

			follows, err := f.state.DB.IsFollowing(ctx, receiver.ID, acc.ID)
			if err != nil {
				// Proper error.
				log.Errorf(ctx, "db error checking follow status: %v", err)
				return false
			}

			if follows {
				// We follow this nerd.
				return true
			}

			// We don't follow this nerd, but
			// have we requested to follow them?
			followRequested, err := f.state.DB.IsFollowRequested(ctx, receiver.ID, acc.ID)
			if err != nil {
				// Proper error.
				log.Errorf(ctx, "db error checking follow req status: %v", err)
				return false
			}

			return followRequested
		},
	)
}

// errantLinks returns true if any http/https
// link discovered in the statusable content + cw
// is not either a mention link, or a hashtag link.
func (f *Filter) errantLinks(
	ctx context.Context,
	statusable ap.Statusable,
	mentions []preppedMention,
	hashtags []*gtsmodel.Tag,
) bool {
	// Concatenate the cw with the
	// content to check for links in both.
	cw := ap.ExtractSummary(statusable)
	content := ap.ExtractContent(statusable)
	concat := cw + " " + content.Content

	// Store link string alongside link
	// URI to avoid stringifying twice.
	type preppedLink struct {
		*url.URL
		str string
	}

	// Find + parse every http/https link in the status.
	rawLinks := regexes.URLLike.FindAllString(concat, -1)
	links := make([]preppedLink, 0, len(rawLinks))
	for _, rawLink := range rawLinks {
		linkURI, err := url.Parse(rawLink)
		if err != nil {
			log.Debugf(ctx,
				"malformed link in status: %v",
				err,
			)
			// Ignore bad links
			// for spam checking.
			continue
		}

		links = append(links, preppedLink{
			URL: linkURI,
			str: rawLink,
		})
	}

	// For each link in the status, try to
	// match it to a hashtag or a mention.
	// If we can't, we have an errant link.
	for _, link := range links {
		hashtagLink := slices.ContainsFunc(
			hashtags,
			func(hashtag *gtsmodel.Tag) bool {
				// If a link is to the href
				// of a hashtag, it's fine.
				return strings.EqualFold(
					link.str,
					hashtag.Href,
				)
			},
		)

		if hashtagLink {
			// This link is accounted for.
			// Move to the next one.
			continue
		}

		mentionLink := slices.ContainsFunc(
			mentions,
			func(mention preppedMention) bool {
				// If link is straight up to the URI
				// of a mentioned account, it's fine.
				if strings.EqualFold(
					link.str,
					mention.TargetAccountURI,
				) {
					return true
				}

				// Link might be to an account URL rather
				// than URI. This is a bit trickier because
				// we can't predict the format of such URLs,
				// and it's difficult to reconstruct them
				// while also taking account of different
				// host + account-domain values.
				//
				// So, just check if this link is on the same
				// host as the mentioned account, or at least
				// shares a host with it.
				if link.Host == mention.domain {
					// Same host.
					return true
				}

				// Shares a host if it has at least two
				// components from the right in common.
				common := dns.CompareDomainName(
					link.Host,
					mention.domain,
				)
				return common >= 2
			},
		)

		if mentionLink {
			// This link is accounted for.
			// Move to the next one.
			continue
		}

		// Not a hashtag link
		// or a mention link,
		// so it's errant.
		return true
	}

	// All links OK, or
	// no links found.
	return false
}

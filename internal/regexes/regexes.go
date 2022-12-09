/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package regexes

import (
	"bytes"
	"fmt"
	"regexp"
	"sync"

	"mvdan.cc/xurls/v2"
)

const (
	users     = "users"
	actors    = "actors"
	statuses  = "statuses"
	inbox     = "inbox"
	outbox    = "outbox"
	followers = "followers"
	following = "following"
	liked     = "liked"
	// collections = "collections"
	// featured    = "featured"
	publicKey = "main-key"
	follow    = "follow"
	// update      = "updates"
	blocks = "blocks"
)

const (
	maximumUsernameLength       = 64
	maximumEmojiShortcodeLength = 30
)

var (
	schemes = `(http|https)://`
	// LinkScheme captures http/https schemes in URLs.
	LinkScheme = func() *regexp.Regexp {
		rgx, err := xurls.StrictMatchingScheme(schemes)
		if err != nil {
			panic(err)
		}
		return rgx
	}()

	mentionName = `^@([\w\-\.]+)(?:@([\w\-\.:]+))?$`
	// MentionName captures the username and domain part from a mention string
	// such as @whatever_user@example.org, returning whatever_user and example.org (without the @ symbols)
	MentionName = regexp.MustCompile(mentionName)

	// mention regex can be played around with here: https://regex101.com/r/P0vpYG/1
	mentionFinder = `(?:^|\s)(@\w+(?:@[a-zA-Z0-9_\-\.]+)?)`
	// MentionFinder extracts mentions from a piece of text.
	MentionFinder = regexp.MustCompile(mentionFinder)

	emojiShortcode = fmt.Sprintf(`\w{2,%d}`, maximumEmojiShortcodeLength)
	// EmojiShortcode validates an emoji name.
	EmojiShortcode = regexp.MustCompile(fmt.Sprintf("^%s$", emojiShortcode))

	// emoji regex can be played with here: https://regex101.com/r/478XGM/1
	emojiFinderString = fmt.Sprintf(`(?:\b)?:(%s):(?:\b)?`, emojiShortcode)
	// EmojiFinder extracts emoji strings from a piece of text.
	EmojiFinder = regexp.MustCompile(emojiFinderString)

	// usernameString defines an acceptable username on this instance
	usernameString = fmt.Sprintf(`[a-z0-9_]{2,%d}`, maximumUsernameLength)
	// Username can be used to validate usernames of new signups
	Username = regexp.MustCompile(fmt.Sprintf(`^%s$`, usernameString))

	userPathString = fmt.Sprintf(`^?/%s/(%s)$`, users, usernameString)
	// UserPath parses a path that validates and captures the username part from eg /users/example_username
	UserPath = regexp.MustCompile(userPathString)

	publicKeyPath = fmt.Sprintf(`^?/%s/(%s)/%s`, users, usernameString, publicKey)
	// PublicKeyPath parses a path that validates and captures the username part from eg /users/example_username/main-key
	PublicKeyPath = regexp.MustCompile(publicKeyPath)

	inboxPath = fmt.Sprintf(`^/?%s/(%s)/%s$`, users, usernameString, inbox)
	// InboxPath parses a path that validates and captures the username part from eg /users/example_username/inbox
	InboxPath = regexp.MustCompile(inboxPath)

	outboxPath = fmt.Sprintf(`^/?%s/(%s)/%s$`, users, usernameString, outbox)
	// OutboxPath parses a path that validates and captures the username part from eg /users/example_username/outbox
	OutboxPath = regexp.MustCompile(outboxPath)

	actorPath = fmt.Sprintf(`^?/%s/(%s)$`, actors, usernameString)
	// ActorPath parses a path that validates and captures the username part from eg /actors/example_username
	ActorPath = regexp.MustCompile(actorPath)

	followersPath = fmt.Sprintf(`^/?%s/(%s)/%s$`, users, usernameString, followers)
	// FollowersPath parses a path that validates and captures the username part from eg /users/example_username/followers
	FollowersPath = regexp.MustCompile(followersPath)

	followingPath = fmt.Sprintf(`^/?%s/(%s)/%s$`, users, usernameString, following)
	// FollowingPath parses a path that validates and captures the username part from eg /users/example_username/following
	FollowingPath = regexp.MustCompile(followingPath)

	followPath = fmt.Sprintf(`^/?%s/(%s)/%s/(%s)$`, users, usernameString, follow, ulid)
	// FollowPath parses a path that validates and captures the username part and the ulid part
	// from eg /users/example_username/follow/01F7XT5JZW1WMVSW1KADS8PVDH
	FollowPath = regexp.MustCompile(followPath)

	ulid = `[0123456789ABCDEFGHJKMNPQRSTVWXYZ]{26}`
	// ULID parses and validate a ULID.
	ULID = regexp.MustCompile(fmt.Sprintf(`^%s$`, ulid))

	likedPath = fmt.Sprintf(`^/?%s/(%s)/%s$`, users, usernameString, liked)
	// LikedPath parses a path that validates and captures the username part from eg /users/example_username/liked
	LikedPath = regexp.MustCompile(likedPath)

	likePath = fmt.Sprintf(`^/?%s/(%s)/%s/(%s)$`, users, usernameString, liked, ulid)
	// LikePath parses a path that validates and captures the username part and the ulid part
	// from eg /users/example_username/like/01F7XT5JZW1WMVSW1KADS8PVDH
	LikePath = regexp.MustCompile(likePath)

	statusesPath = fmt.Sprintf(`^/?%s/(%s)/%s/(%s)$`, users, usernameString, statuses, ulid)
	// StatusesPath parses a path that validates and captures the username part and the ulid part
	// from eg /users/example_username/statuses/01F7XT5JZW1WMVSW1KADS8PVDH
	// The regex can be played with here: https://regex101.com/r/G9zuxQ/1
	StatusesPath = regexp.MustCompile(statusesPath)

	blockPath = fmt.Sprintf(`^/?%s/(%s)/%s/(%s)$`, users, usernameString, blocks, ulid)
	// BlockPath parses a path that validates and captures the username part and the ulid part
	// from eg /users/example_username/blocks/01F7XT5JZW1WMVSW1KADS8PVDH
	BlockPath = regexp.MustCompile(blockPath)

	filePath = fmt.Sprintf(`^(%s)/([a-z]+)/([a-z]+)/(%s)\.([a-z]+)$`, ulid, ulid)
	// FilePath parses a file storage path of the form [ACCOUNT_ID]/[MEDIA_TYPE]/[MEDIA_SIZE]/[FILE_NAME]
	// eg 01F8MH1H7YV1Z7D2C8K2730QBF/attachment/small/01F8MH8RMYQ6MSNY3JM2XT1CQ5.jpeg
	// It captures the account id, media type, media size, file name, and file extension, eg
	// `01F8MH1H7YV1Z7D2C8K2730QBF`, `attachment`, `small`, `01F8MH8RMYQ6MSNY3JM2XT1CQ5`, `jpeg`.
	FilePath = regexp.MustCompile(filePath)
)

// bufpool is a memory pool of byte buffers for use in our regex utility functions.
var bufpool = sync.Pool{
	New: func() any {
		buf := bytes.NewBuffer(make([]byte, 0, 512))
		return buf
	},
}

// ReplaceAllStringFunc will call through to .ReplaceAllStringFunc in the provided regex, but provide you a clean byte buffer for optimized string writes.
func ReplaceAllStringFunc(rgx *regexp.Regexp, src string, repl func(match string, buf *bytes.Buffer) string) string {
	buf := bufpool.Get().(*bytes.Buffer) //nolint
	defer bufpool.Put(buf)
	return rgx.ReplaceAllStringFunc(src, func(match string) string {
		buf.Reset() // reset use
		return repl(match, buf)
	})
}

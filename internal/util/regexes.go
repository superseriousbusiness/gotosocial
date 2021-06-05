/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

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

package util

import (
	"fmt"
	"regexp"
)

const (
	minimumPasswordEntropy      = 60 // dictates password strength. See https://github.com/wagslane/go-password-validator
	minimumReasonLength         = 40
	maximumReasonLength         = 500
	maximumEmailLength          = 256
	maximumUsernameLength       = 64
	maximumPasswordLength       = 64
	maximumEmojiShortcodeLength = 30
	maximumHashtagLength        = 30
)

var (
	mentionNameRegexString = `^@([a-zA-Z0-9_]+)(?:@([a-zA-Z0-9_\-\.]+)?)$`
	// mention name regex captures the username and domain part from a mention string
	// such as @whatever_user@example.org, returning whatever_user and example.org (without the @ symbols)
	mentionNameRegex = regexp.MustCompile(mentionNameRegexString)

	// mention regex can be played around with here: https://regex101.com/r/qwM9D3/1
	mentionFinderRegexString = `(?: |^|\W)(@[a-zA-Z0-9_]+(?:@[a-zA-Z0-9_\-\.]+)?)(?:[^a-zA-Z0-9]|\W)`
	mentionFinderRegex       = regexp.MustCompile(mentionFinderRegexString)

	// hashtag regex can be played with here: https://regex101.com/r/Vhy8pg/1
	hashtagFinderRegexString = fmt.Sprintf(`(?: |^|\W)?#([a-zA-Z0-9]{1,%d})(?:\b|\r)`, maximumHashtagLength)
	hashtagFinderRegex       = regexp.MustCompile(hashtagFinderRegexString)

	// emoji shortcode regex can be played with here: https://regex101.com/r/zMDRaG/1
	emojiShortcodeRegexString     = fmt.Sprintf(`[a-z0-9_]{2,%d}`, maximumEmojiShortcodeLength)
	emojiShortcodeValidationRegex = regexp.MustCompile(fmt.Sprintf("^%s$", emojiShortcodeRegexString))

	// emoji regex can be played with here: https://regex101.com/r/478XGM/1
	emojiFinderRegexString = fmt.Sprintf(`(?: |^|\W)?:(%s):(?:\b|\r)?`, emojiShortcodeRegexString)
	emojiFinderRegex       = regexp.MustCompile(emojiFinderRegexString)

	// usernameRegexString defines an acceptable username on this instance
	usernameRegexString = fmt.Sprintf(`[a-z0-9_]{2,%d}`, maximumUsernameLength)
	// usernameValidationRegex can be used to validate usernames of new signups
	usernameValidationRegex = regexp.MustCompile(fmt.Sprintf(`^%s$`, usernameRegexString))

	userPathRegexString = fmt.Sprintf(`^?/%s/(%s)$`, UsersPath, usernameRegexString)
	// userPathRegex parses a path that validates and captures the username part from eg /users/example_username
	userPathRegex = regexp.MustCompile(userPathRegexString)

	inboxPathRegexString = fmt.Sprintf(`^/?%s/(%s)/%s$`, UsersPath, usernameRegexString, InboxPath)
	// inboxPathRegex parses a path that validates and captures the username part from eg /users/example_username/inbox
	inboxPathRegex = regexp.MustCompile(inboxPathRegexString)

	outboxPathRegexString = fmt.Sprintf(`^/?%s/(%s)/%s$`, UsersPath, usernameRegexString, OutboxPath)
	// outboxPathRegex parses a path that validates and captures the username part from eg /users/example_username/outbox
	outboxPathRegex = regexp.MustCompile(outboxPathRegexString)

	actorPathRegexString = fmt.Sprintf(`^?/%s/(%s)$`, ActorsPath, usernameRegexString)
	// actorPathRegex parses a path that validates and captures the username part from eg /actors/example_username
	actorPathRegex = regexp.MustCompile(actorPathRegexString)

	followersPathRegexString = fmt.Sprintf(`^/?%s/(%s)/%s$`, UsersPath, usernameRegexString, FollowersPath)
	// followersPathRegex parses a path that validates and captures the username part from eg /users/example_username/followers
	followersPathRegex = regexp.MustCompile(followersPathRegexString)

	followingPathRegexString = fmt.Sprintf(`^/?%s/(%s)/%s$`, UsersPath, usernameRegexString, FollowingPath)
	// followingPathRegex parses a path that validates and captures the username part from eg /users/example_username/following
	followingPathRegex = regexp.MustCompile(followingPathRegexString)

	// see https://ihateregex.io/expr/uuid/
	uuidRegexString = `[0-9a-fA-F]{8}\b-[0-9a-fA-F]{4}\b-[0-9a-fA-F]{4}\b-[0-9a-fA-F]{4}\b-[0-9a-fA-F]{12}`

	likedPathRegexString = fmt.Sprintf(`^/?%s/(%s)/%s$`, UsersPath, usernameRegexString, LikedPath)
	// likedPathRegex parses a path that validates and captures the username part from eg /users/example_username/liked
	likedPathRegex = regexp.MustCompile(likedPathRegexString)

	likePathRegexString = fmt.Sprintf(`^/?%s/(%s)/%s/(%s)$`, UsersPath, usernameRegexString, LikedPath, uuidRegexString)
	// likePathRegex parses a path that validates and captures the username part and the uuid part
	// from eg /users/example_username/liked/123e4567-e89b-12d3-a456-426655440000.
	likePathRegex = regexp.MustCompile(likePathRegexString)

	statusesPathRegexString = fmt.Sprintf(`^/?%s/(%s)/%s/(%s)$`, UsersPath, usernameRegexString, StatusesPath, uuidRegexString)
	// statusesPathRegex parses a path that validates and captures the username part and the uuid part
	// from eg /users/example_username/statuses/123e4567-e89b-12d3-a456-426655440000.
	// The regex can be played with here: https://regex101.com/r/G9zuxQ/1
	statusesPathRegex = regexp.MustCompile(statusesPathRegexString)
)

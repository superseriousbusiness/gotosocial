# Roadmap <!-- omit in toc -->

This document contains the roadmap for GoToSocial to be considered eligible for its first [beta release](https://en.wikipedia.org/wiki/Software_release_life_cycle#Beta), and a summary of progress made towards that goal.

All the dates contained in this document are best-guess only. It's useful to have a rough timeline, but things will undoubtedly change along the way.

## Table of Contents <!-- omit in toc -->

- [Beta Aims](#beta-aims)
- [Timeline](#timeline)
  - [Q3 2022](#q3-2022)
  - [Q4 2022](#q4-2022)
  - [Q1 2023](#q1-2023)
  - [Q2 2023](#q2-2023)
  - [Q3 2023](#q3-2023)
- [Detailed To-do List](#detailed-to-do-list)
  - [API Groups + Endpoints](#api-groups--endpoints)
  - [Non-API tasks](#non-api-tasks)

## Beta Aims

The milestone for beta in GoToSocial's case is to have a feature set that roughly compares to existing popular ActivityPub server implementations (minus features which we don't see as particularly important or useful). The beta milestone also includes GoToSocial-specific features which, we believe, are vital for user safety, such as local-only posting, block list subscriptions, and so on.

Once feature parity is roughly in place, we will use beta time to start adding and polishing bonus features like slow federation, group moderation decisions, migration tooling, etc.

We currently foresee entering beta phase around the middle of 2023, though this is only an estimate and is subject to change.

## Timeline

What follows is a per-quarter timeline of features that will be implemented on the road to beta. The timeline is calculated on the following assumptions:

- We will continue to develop at a pace that is similar to what we've done over the previous year.
- Our combined bandwidth is roughly equivalent to one person working full time on the project.
- One distinct 'feature' takes one person 2-4 weeks to develop and test, depending on the size of the feature.
- There will be other bugs to fix in between implementing things, so we shouldn't pack in features too tightly.

Each quarter contains one 'big feature' which will probably take the longest amount of time that quarter.

**This timeline is a best-guess about when things will be implemented. The order of feature releases may change. It may go faster or slower depending on the number of hurdles we run into, and the amount of help we receive from community contributions of code. The timeline also does not include background tasks like admin, polishing existing features, refactoring code, and ensuring compatibility with other AP implementations.**

### Q3 2022

- **Big Feature** -- User settings page: allow users to edit their profile page and settings through a web page served by the GtS instance, for cases where client apps don't implement their own settings page.
- Block list subscription support: allow instance admins to subscribe their instance to plaintext domain block lists (much of the work for this is already in place).
- Tag support: implement federating hashtags and viewing hashtags to allow users to discover posts that they might be interested in.

### Q4 2022

- **Big Feature** — Video support: allow users to view + post videos (before we only implemented images and gifs).
- Custom emoji support: allow users to use custom emojis in posts. Fetch custom emojis from remote instances and display them properly.
- Pinned posts: allow users to 'feature' or 'pin' posts on their profile, and serve these featured posts via AP for other servers to see.
- Profile fields: allow users to set 'fields' on their profile: short key/value items that can display pronouns, links to websites, etc.

### Q1 2023

- **Big Feature** — Reports: allow users to file reports for abusive behavior etc. Expose the API for admins to view + act on reports. Handle federation of reports.
- List support: allow users to create lists of other users, which they can view as separate timelines.
- Polls support: allow users to create polls and vote in existing polls; federate the polls correctly via AP.

### Q2 2023

- **Big Feature** — Sign-up flow. Allow users to submit a sign-up request to an instance. Allow admins to moderate sign-up requests.
- Direct conversations: Allow users to see all direct-message conversations they're a part of.
- Muting conversations: Allow users to mute notifications for conversations they're no longer interested in.

### Q3 2023

- **Big Feature** — Support the `Move` Activity, to allow users to move across instances and Fediverse implementations.
- More to be confirmed.

## Detailed To-do List

### API Groups + Endpoints

Unfilled check - not yet implemented.
Filled check   - implemented.
Crossed out    - will not be implemented / will be stubbed only.

- [ ] Client-To-Server (Client REST API)
  - [ ] Token and sign-in
    - [x] /api/v1/apps POST                                 (Create an application)
    - [ ] /api/v1/apps/verify_credentials GET               (Verify an application works)
    - [x] /oauth/authorize GET                              (Show authorize page to user)
    - [x] /oauth/authorize POST                             (Get an OAuth access code for an app/user)
    - [x] /oauth/token POST                                 (Obtain a user-level access token)
    - [ ] /oauth/revoke POST                                (Revoke a user-level access token)
    - [x] /auth/sign_in GET                                 (Show form for user sign-in)
    - [x] /auth/sign_in POST                                (Validate username and password and sign user in)
  - [ ] Accounts
    - [x] /api/v1/accounts POST                             (Register a new account)
    - [x] /api/v1/accounts/verify_credentials GET           (Verify account credentials with a user token)
    - [x] /api/v1/accounts/update_credentials PATCH         (Update user's display name/preferences)
    - [x] /api/v1/accounts/:id GET                          (Get account information)
    - [x] /api/v1/accounts/:id/statuses GET                 (Get an account's statuses)
    - [x] /api/v1/accounts/:id/followers GET                (Get an account's followers)
    - [x] /api/v1/accounts/:id/following GET                (Get an account's following)
    - [ ] /api/v1/accounts/:id/featured_tags GET            (Get an account's featured tags)
    - [ ] /api/v1/accounts/:id/lists GET                    (Get lists containing this account)
    - [x] /api/v1/accounts/:id/follow POST                  (Follow this account)
    - [x] /api/v1/accounts/:id/unfollow POST                (Unfollow this account)
    - [x] /api/v1/accounts/:id/block POST                   (Block this account)
    - [x] /api/v1/accounts/:id/unblock POST                 (Unblock this account)
    - [ ] /api/v1/accounts/:id/mute POST                    (Mute this account)
    - [ ] /api/v1/accounts/:id/unmute POST                  (Unmute this account)
    - [ ] /api/v1/accounts/:id/pin POST                     (Feature this account on profile)
    - [ ] /api/v1/accounts/:id/unpin POST                   (Remove this account from profile)
    - [ ] /api/v1/accounts/:id/note POST                    (Make a personal note about this account)
    - [x] /api/v1/accounts/relationships GET                (Check relationships with accounts)
    - [ ] /api/v1/accounts/search GET                       (Search for an account)
    - ~~/api/v1/accounts/:id/identity_proofs GET            (Get identity proofs for this account)~~
  - [x] Favourites
    - [x] /api/v1/favourites GET                            (See faved statuses)
  - [ ] Mutes
    - [ ] /api/v1/mutes GET                                 (See list of muted accounts)
  - [x] Blocks
    - [x] /api/v1/blocks GET                                (See list of blocked accounts)
  - [x] Domain Blocks
    - [x] /api/v1/domain_blocks GET                         (See list of domain blocks)
    - [x] /api/v1/domain_blocks POST                        (Create a domain block)
    - [x] /api/v1/domain_blocks DELETE                      (Remove a domain block)
  - [ ] Filters
    - [ ] /api/v1/filters GET                               (Get list of filters)
    - [ ] /api/v1/filters/:id GET                           (View a filter)
    - [ ] /api/v1/filters POST                              (Create a filter)
    - [ ] /api/v1/filters/:id PUT                           (Update a filter)
    - [ ] /api/v1/filters/:id DELETE                        (Remove a filter)
  - [ ] Reports
    - [ ] /api/v1/reports POST                              (File a report)
  - [x] Follow Requests
    - [x] /api/v1/follow_requests GET                       (View pending follow requests)
    - [x] /api/v1/follow_requests/:id/authorize POST        (Accept a follow request)
    - [x] /api/v1/follow_requests/:id/reject POST           (Reject a follow request)
  - [ ] Featured Tags
    - [ ] /api/v1/featured_tags GET                         (View featured tags)
    - [ ] /api/v1/featured_tags POST                        (Feature a tag)
    - [ ] /api/v1/featured_tags/:id DELETE                  (Unfeature a tag)
    - ~~/api/v1/featured_tags/suggestions GET               (See most used tags)~~
  - [ ] Preferences
    - [ ] /api/v1/preferences GET                           (Get user preferences)
  - [ ] Statuses
    - [x] /api/v1/statuses POST                             (Create a new status)
    - [x] /api/v1/statuses/:id GET                          (View an existing status)
    - [x] /api/v1/statuses/:id DELETE                       (Delete a status)
    - [x] /api/v1/statuses/:id/context GET                  (View statuses above and below status ID)
    - [x] /api/v1/statuses/:id/reblogged_by GET             (See who has reblogged a status)
    - [x] /api/v1/statuses/:id/favourited_by GET            (See who has faved a status)
    - [x] /api/v1/statuses/:id/favourite POST               (Fave a status)
    - [x] /api/v1/statuses/:id/unfavourite POST             (Unfave a status)
    - [x] /api/v1/statuses/:id/reblog POST                  (Reblog a status)
    - [x] /api/v1/statuses/:id/unreblog POST                (Undo a reblog)
    - [ ] /api/v1/statuses/:id/mute POST                    (Mute notifications on a status)
    - [ ] /api/v1/statuses/:id/unmute POST                  (Unmute notifications on a status)
    - [ ] /api/v1/statuses/:id/pin POST                     (Pin a status to profile)
    - [ ] /api/v1/statuses/:id/unpin POST                   (Unpin a status from profile)
    - [x] /api/v1/statuses/:id/bookmark POST                  (Bookmark a status)
    - [x] /api/v1/statuses/:id/unbookmark POST                (Undo a bookmark)
  - [x] Media
    - [x] /api/v1/media POST                                (Upload a media attachment)
    - [x] /api/v1/media/:id GET                             (Get a media attachment)
    - [x] /api/v1/media/:id PUT                             (Update an attachment)
  - [ ] Polls
    - [ ] /api/v1/polls/:id GET                             (Show a poll)
    - [ ] /api/v1/polls/:id/votes POST                      (Vote on a poll)
  - [ ] Timelines
    - [x] /api/v1/timelines/public GET                      (See the public/federated timeline)
    - [ ] /api/v1/timelines/tag/:hashtag GET                (Get public statuses that use hashtag)
    - [x] /api/v1/timelines/home GET                        (View statuses from followed users)
    - [ ] /api/v1/timelines/list/:list_id GET               (Get statuses in given list)
  - [ ] Conversations
    - [ ] /api/v1/conversations GET                         (Get a list of direct message convos)
    - [ ] /api/v1/conversations/:id DELETE                  (Delete a direct message convo)
    - [ ] /api/v1/conversations/:id POST                    (Mark a conversation as read)
  - [ ] Lists
    - [ ] /api/v1/lists GET                                 (Show a list of lists)
    - [ ] /api/v1/lists/:id GET                             (Show a single list)
    - [ ] /api/v1/lists POST                                (Create a new list)
    - [ ] /api/v1/lists/:id PUT                             (Update a list)
    - [ ] /api/v1/lists/:id DELETE                          (Delete a list)
    - [ ] /api/v1/lists/:id/accounts GET                    (View which accounts are in a list)
    - [ ] /api/v1/lists/:id/accounts POST                   (Add accounts to a list)
    - [ ] /api/v1/lists/:id/accounts DELETE                 (Remove accounts from a list)
  - [ ] Markers
    - [ ] /api/v1/markers GET                               (Get saved timeline position)
    - [ ] /api/v1/markers POST                              (Save timeline position)
  - [x] Streaming
    - [x] /api/v1/streaming WEBSOCKETS                      (Stream live events to user via websockets)
  - [ ] Notifications
    - [x] /api/v1/notifications GET                         (Get list of notifications)
    - [x] /api/v1/notifications/:id GET                     (Get a single notification)
    - [ ] /api/v1/notifications/clear POST                  (Clear all notifications)
    - [ ] /api/v1/notifications/:id POST                    (Clear a single notification)
  - [x] Search
    - [x] /api/v2/search GET                                (Get search query results)
  - [ ] Instance
    - [x] /api/v1/instance GET                              (Get instance information)
    - [x] /api/v1/instance PATCH                            (Update instance information)
    - [x] /api/v1/instance/peers GET                        (Get list of federated servers)
    - ~~ /api/v1/instance/activity GET                      (Instance activity over the last 3 months, binned weekly.)~~
  - [x] Custom Emojis
    - [x] /api/v1/custom_emojis GET                         (Show this server's custom emoji)
  - [ ] Admin
    - [x] /api/v1/admin/custom_emojis POST                  (Upload a custom emoji for instance-wide usage)
    - [ ] /api/v1/admin/accounts GET                        (View accounts filtered by criteria)
    - [ ] /api/v1/admin/accounts/:id GET                    (View admin level info about an account)
    - [x] /api/v1/admin/accounts/:id/action POST            (Perform an admin action on account)
    - [ ] /api/v1/admin/accounts/:id/approve POST           (Approve pending account)
    - [ ] /api/v1/admin/accounts/:id/reject POST            (Deny pending account)
    - [ ] /api/v1/admin/accounts/:id/enable POST            (Re-enable a disabled account)
    - [ ] /api/v1/admin/accounts/:id/unsilence POST         (Unsilence a silenced account)
    - [ ] /api/v1/admin/accounts/:id/unsuspend POST         (Unsuspend a suspended account)
    - [ ] /api/v1/admin/reports GET                         (View all reports)
    - [ ] /api/v1/admin/reports/:id GET                     (View a single report)
    - [ ] /api/v1/admin/reports/:id/assign_to_self POST     (Assign a report to the current admin account)
    - [ ] /api/v1/admin/reports/:id/unassign POST           (Unassign a report)
    - [ ] /api/v1/admin/reports/:id/resolve POST            (Mark a report as resolved)
    - [ ] /api/v1/admin/reports/:id/reopen POST             (Reopen a closed report)
  - [ ] Announcements
    - [ ] /api/v1/announcements GET                         (Show all current announcements)
    - [ ] /api/v1/announcements/:id/dismiss POST            (Mark an announcement as read)
    - [ ] /api/v1/announcements/:id/reactions/:name PUT     (Add a reaction to an announcement)
    - [ ] /api/v1/announcements/:id/reactions/:name DELETE  (Remove a reaction from an announcement)
  - [ ] Oembed
    - [ ] /api/oembed GET                                   (Get oembed metadata for a status URL)
  - ~~Proofs~~
    - ~~/api/proofs GET                                     (View identity proofs)~~
  - ~~Trends~~
    - ~~/api/v1/trends GET                                  (Get a list of trending tags for the last week)~~
  - ~~Directory~~
    - ~~/api/v1/directory GET                               (Show profiles this server is aware of.)~~
  - ~~Endorsements~~
    - ~~/api/v1/endorsements GET                            (View existing endorsements)~~
  - ~~Push~~
    - ~~/api/v1/push/subscription POST                      (Subscribe to push notifications)~~
    - ~~/api/v1/push/subscription GET                       (Get current subscription)~~
    - ~~/api/v1/push/subscription PUT                       (Change notification types)~~
    - ~~/api/v1/push/subscription DELETE                    (Delete current subscription)~~
  - ~~Scheduled Statuses~~
    - ~~/api/v1/scheduled_statuses GET                      (View scheduled statuses)~~
    - ~~/api/v1/scheduled_statuses/:id GET                  (View a scheduled status)~~
    - ~~/api/v1/scheduled_statuses/:id PUT                  (Schedule a status)~~
    - ~~/api/v1/scheduled_statuses/:id DELETE               (Cancel a scheduled status)~~
  - ~~Suggestions~~
    - ~~/api/v1/suggestions GET                             (Get suggested accounts to follow)~~
    - ~~/api/v1/suggestions/:account_id DELETE              (Delete a suggestion)~~
  - [x] Bookmarks
    - [x] /api/v1/bookmarks GET                               (See bookmarked statuses)

### Non-API tasks

- [ ] Server-To-Server (Federation protocol)
  - [x] Mechanism to trigger side effects from client AP
  - [x] Webfinger account lookups
  - [ ] Federation modes
    - [ ] 'Slow' federation
      - [ ] Reputation scoring system for instances
    - [x] 'Greedy' federation
    - [ ] No federation (insulate this instance from the Fediverse)
      - [ ] Allow list
  - [x] Secure HTTP signatures (creation and validation)
- [x] Storage
  - [x] Internal/statuses/preferences etc
    - [x] Postgres interface
    - [x] SQLite interface
  - [x] Media storage
    - [x] Local storage interface
    - [x] S3 storage interface
- [x] Cache
  - [x] In-memory cache
- [ ] Security features
  - [x] Authorization middleware
  - [ ] Rate limiting middleware
  - [ ] Scope middleware
  - [x] Permissions/ACL middleware for admins+moderators
- [ ] Documentation
  - [x] Swagger API documentation
  - [x] ReadTheDocs.io documentation
  - [x] Deployment documentation
  - [ ] App creation guide
- [ ] Tooling
  - [ ] Database migration tool
  - [x] Admin CLI tool
- [x] Build
  - [x] Docker containerization
    - [x] Dockerfile
    - [x] docker-compose.yml
- [ ] Tests
  - [ ] Unit/integration
    - [x] 25% coverage
    - [ ] 50% coverage
    - [ ] 90%+ coverage
  - [ ] Benchmarking

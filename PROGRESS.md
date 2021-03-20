# Progress

* [ ] Client-To-Server (Client REST API)
  * [ ] Token and sign-in
    * [x] /api/v1/apps POST                                 (Create an application)
    * [ ] /api/v1/apps/verify_credentials GET               (Verify an application works)
    * [x] /oauth/authorize GET                              (Show authorize page to user)
    * [x] /oauth/authorize POST                             (Get an oauth access code for an app/user)
    * [ ] /oauth/token POST                                 (Obtain a user-level access token)
    * [ ] /oauth/revoke POST                                (Revoke a user-level access token)
    * [x] /auth/sign_in GET                                 (Show form for user signin)
    * [x] /auth/sign_in POST                                (Validate username and password and sign user in)
  * [ ] Accounts
    * [ ] /api/v1/accounts POST                             (Register a new account)
    * [ ] /api/v1/accounts/verify_credentials GET           (Verify account credentials with a user token)
    * [ ] /api/v1/accounts/update_credentials PATCH         (Update user's display name/preferences)
    * [ ] /api/v1/accounts/:id GET                          (Get account information)
    * [ ] /api/v1/accounts/:id/statuses GET                 (Get an account's statuses)
    * [ ] /api/v1/accounts/:id/followers GET                (Get an account's followers)
    * [ ] /api/v1/accounts/:id/following GET                (Get an account's following)
    * [ ] /api/v1/accounts/:id/featured_tags GET            (Get an account's featured tags)
    * [ ] /api/v1/accounts/:id/lists GET                    (Get lists containing this account)
    * [ ] /api/v1/accounts/:id/identity_proofs GET          (Get identity proofs for this account)
    * [ ] /api/v1/accounts/:id/follow POST                  (Follow this account)
    * [ ] /api/v1/accounts/:id/unfollow POST                (Unfollow this account)
    * [ ] /api/v1/accounts/:id/block POST                   (Block this account)
    * [ ] /api/v1/accounts/:id/unblock POST                 (Unblock this account)
    * [ ] /api/v1/accounts/:id/mute POST                    (Mute this account)
    * [ ] /api/v1/accounts/:id/unmute POST                  (Unmute this account)
    * [ ] /api/v1/accounts/:id/pin POST                     (Feature this account on profile)
    * [ ] /api/v1/accounts/:id/unpin POST                   (Remove this account from profile)
    * [ ] /api/v1/accounts/:id/note POST                    (Make a personal note about this account)
    * [ ] /api/v1/accounts/relationships GET                (Check relationships with accounts)
    * [ ] /api/v1/accounts/search GET                       (Search for an account)
  * [ ] Bookmarks
    * [ ] /api/v1/bookmarks GET                             (See bookmarked statuses)
  * [ ] Favourites
    * [ ] /api/v1/favourites GET                            (See faved statuses)
  * [ ] Mutes
    * [ ] /api/v1/mutes GET                                 (See list of muted accounts)
  * [ ] Blocks
    * [ ] /api/v1/blocks GET                                (See list of blocked accounts)
  * [ ] Domain Blocks
    * [ ] /api/v1/domain_blocks GET                         (See list of domain blocks)
    * [ ] /api/v1/domain_blocks POST                        (Create a domain block)
    * [ ] /api/v1/domain_blocks DELETE                      (Remove a domain block)
  * [ ] Filters
    * [ ] /api/v1/filters GET                               (Get list of filters)
    * [ ] /api/v1/filters/:id GET                           (View a filter)
    * [ ] /api/v1/filters POST                              (Create a filter)
    * [ ] /api/v1/filters/:id PUT                           (Update a filter)
    * [ ] /api/v1/filters/:id DELETE                        (Remove a filter)
  * [ ] Reports
    * [ ] /api/v1/reports POST                              (File a report)
  * [ ] Follow Requests
    * [ ] /api/v1/follow_requests GET                       (View pending follow requests)
    * [ ] /api/v1/follow_requests/:id/authorize POST        (Accept a follow request)
    * [ ] /api/v1/follow_requests/:id/reject POST           (Reject a follow request)
  * [ ] Endorsements
    * [ ] /api/v1/endorsements GET                          (View existing endorsements)
  * [ ] Featured Tags
    * [ ] /api/v1/featured_tags GET                         (View featured tags)
    * [ ] /api/v1/featured_tags POST                        (Feature a tag)
    * [ ] /api/v1/featured_tags/:id DELETE                  (Unfeature a tag)
    * [ ] /api/v1/featured_tags/suggestions GET             (See most used tags)
  * [ ] Preferences
    * [ ] /api/v1/preferences GET                           (Get user preferences)
  * [ ] Suggestions
    * [ ] /api/v1/suggestions GET                           (Get suggested accounts to follow)
    * [ ] /api/v1/suggestions/:account_id DELETE            (Delete a suggestion)
  * [ ] Statuses
    * [ ] /api/v1/statuses POST                             (Create a new status)
    * [ ] /api/v1/statuses/:id GET                          (View an existing status)
    * [ ] /api/v1/statuses/:id DELETE                       (Delete a status)
    * [ ] /api/v1/statuses/:id/context GET                  (View statuses above and below status ID)
    * [ ] /api/v1/statuses/:id/reblogged_by GET             (See who has reblogged a status)
    * [ ] /api/v1/statuses/:id/favourited_by GET            (See who has faved a status)
    * [ ] /api/v1/statuses/:id/favourite POST               (Fave a status)
    * [ ] /api/v1/statuses/:id/favourite POST               (Unfave a status)
    * [ ] /api/v1/statuses/:id/reblog POST                  (Reblog a status)
    * [ ] /api/v1/statuses/:id/unreblog POST                (Undo a reblog)
    * [ ] /api/v1/statuses/:id/bookmark POST                (Bookmark a status)
    * [ ] /api/v1/statuses/:id/unbookmark POST              (Undo a bookmark)
    * [ ] /api/v1/statuses/:id/mute POST                    (Mute notifications on a status)
    * [ ] /api/v1/statuses/:id/unmute POST                  (Unmute notifications on a status)
    * [ ] /api/v1/statuses/:id/pin POST                     (Pin a status to profile)
    * [ ] /api/v1/statuses/:id/unpin POST                   (Unpin a status from profile)
  * [ ] Media
    * [ ] /api/v1/media POST                                (Upload a media attachment)
    * [ ] /api/v1/media/:id GET                             (Get a media attachment)
    * [ ] /api/v1/media/:id PUT                             (Update an attachment)
  * [ ] Polls
    * [ ] /api/v1/polls/:id GET                             (Show a poll)
    * [ ] /api/v1/polls/:id/votes POST                      (Vote on a poll)
  * [ ] Scheduled Statuses
    * [ ] /api/v1/scheduled_statuses GET                    (View scheduled statuses)
    * [ ] /api/v1/scheduled_statuses/:id GET                (View a scheduled status)
    * [ ] /api/v1/scheduled_statuses/:id PUT                (Schedule a status)
    * [ ] /api/v1/scheduled_statuses/:id DELETE             (Cancel a scheduled status)
  * [ ] Timelines
    * [ ] /api/v1/timelines/public GET                      (See the public/federated timeline)
    * [ ] /api/v1/timelines/tag/:hashtag GET                (Get public statuses that use hashtag)
    * [ ] /api/v1/timelines/home GET                        (View statuses from followed users)
    * [ ] /api/v1/timelines/list/:list_id GET               (Get statuses in given list)
  * [ ] Conversations
    * [ ] /api/v1/conversations GET                         (Get a list of direct message convos)
    * [ ] /api/v1/conversations/:id DELETE                  (Delete a direct message convo)
    * [ ] /api/v1/conversations/:id POST                    (Mark a conversation as read)
  * [ ] Lists
    * [ ] /api/v1/lists GET                                 (Show a list of lists)
    * [ ] /api/v1/lists/:id GET                             (Show a single list)
    * [ ] /api/v1/lists POST                                (Create a new list)
    * [ ] /api/v1/lists/:id PUT                             (Update a list)
    * [ ] /api/v1/lists/:id DELETE                          (Delete a list)
    * [ ] /api/v1/lists/:id/accounts GET                    (View which accounts are in a list)
    * [ ] /api/v1/lists/:id/accounts POST                   (Add accounts to a list)
    * [ ] /api/v1/lists/:id/accounts DELETE                 (Remove accounts from a list)
  * [ ] Markers
    * [ ] /api/v1/markers GET                               (Get saved timeline position)
    * [ ] /api/v1/markers POST                              (Save timeline position)
  * [ ] Streaming
    * [ ] /api/v1/streaming WEBSOCKETS                      (Stream live events to user via websockets)
  * [ ] Notifications
    * [ ] /api/v1/notifications GET                         (Get list of notifications)
    * [ ] /api/v1/notifications/:id GET                     (Get a single notification)
    * [ ] /api/v1/notifications/clear POST                  (Clear all notifications)
    * [ ] /api/v1/notifications/:id POST                    (Clear a single notification)
  * [ ] Push
    * [ ] /api/v1/push/subscription POST                    (Subscribe to push notifications)
    * [ ] /api/v1/push/subscription GET                     (Get current subscription)
    * [ ] /api/v1/push/subscription PUT                     (Change notification types)
    * [ ] /api/v1/push/subscription DELETE                  (Delete current subscription)
  * [ ] Search
    * [ ] /api/v2/search GET                                (Get search query results)
  * [ ] Instance
    * [ ] /api/v1/instance GET                              (Get instance information)
    * [ ] /api/v1/instance PATCH                            (Update instance information)
    * [ ] /api/v1/instance/peers GET                        (Get list of federated servers)
    * [ ] /api/v1/instance/activity GET                     (Instance activity over the last 3 months, binned weekly.)
  * [ ] Trends
    * [ ] /api/v1/trends GET                                (Get a list of trending tags for the last week)
  * [ ] Directory
    * [ ] /api/v1/directory GET                             (Show profiles this server is aware of.)
  * [ ] Custom Emojis
    * [ ] /api/v1/custom_emojis GET                         (Show this server's custom emoji)
  * [ ] Admin
    * [ ] /api/v1/admin/accounts GET                        (View accounts filtered by criteria)
    * [ ] /api/v1/admin/accounts/:id GET                    (View admin level info about an account)
    * [ ] /api/v1/admin/accounts/:id/action POST            (Perform an admin action on account)
    * [ ] /api/v1/admin/accounts/:id/approve POST           (Approve pending account)
    * [ ] /api/v1/admin/accounts/:id/reject POST            (Deny pending account)
    * [ ] /api/v1/admin/accounts/:id/enable POST            (Reenable a disabled account)
    * [ ] /api/v1/admin/accounts/:id/unsilence POST         (Unsilence a silenced account)
    * [ ] /api/v1/admin/accounts/:id/unsuspend POST         (Unsuspend a suspended account)
    * [ ] /api/v1/admin/reports GET                         (View all reports)
    * [ ] /api/v1/admin/reports/:id GET                     (View a single report)
    * [ ] /api/v1/admin/reports/:id/assign_to_self POST     (Assign a report to the current admin account)
    * [ ] /api/v1/admin/reports/:id/unassign POST           (Unassign a report)
    * [ ] /api/v1/admin/reports/:id/resolve POST            (Mark a report as resolved)
    * [ ] /api/v1/admin/reports/:id/reopen POST             (Reopen a closed report)
  * [ ] Announcements
    * [ ] /api/v1/announcements GET                         (Show all current announcements)
    * [ ] /api/v1/announcements/:id/dismiss POST            (Mark an announcement as read)
    * [ ] /api/v1/announcements/:id/reactions/:name PUT     (Add a reaction to an announcement)
    * [ ] /api/v1/announcements/:id/reactions/:name DELETE  (Remove a reaction from an announcement)
  * [ ] Proofs
    * [ ] /api/proofs GET                                   (View identity proofs)
  * [ ] Oembed
    * [ ] /api/oembed GET                                   (Get oembed metadata for a status URL)
* [ ] Server-To-Server (Federation protocol)
  * [ ] Mechanism to trigger side effects from client AP
  * [ ] Federation modes
    * [ ] 'Slow' federation
      * [ ] Reputation scoring system for instances
    * [ ] 'Greedy' federation
    * [ ] No federation (insulate this instance from the Fediverse)
      * [ ] Allowlist
* [ ] Storage
  * [x] Internal/statuses/preferences etc
    * [x] Postgres interface
  * [ ] Media storage
    * [ ] Local storage interface
    * [ ] S3 storage interface
* [ ] Cache
  * [ ] In-memory cache
* [ ] Security features
  * [ ] Authorization middleware
  * [ ] Rate limiting middleware
  * [ ] Scope middleware
  * [ ] Permissions/acl middleware for admins+moderators
* [ ] Documentation
  * [ ] Swagger API documentation
  * [ ] ReadTheDocs.io documentation
  * [ ] Deployment documentation
  * [ ] App creation guide
* [ ] Tooling
  * [ ] Database migration tool
  * [ ] Admin CLI tool
* [ ] Build
  * [ ] Docker containerization
    * [ ] Dockerfile
    * [ ] docker-compose.yml
* [ ] Tests
  * [ ] Unit/integration
    * [ ] 25% coverage
    * [ ] 50% coverage
    * [ ] 90%+ coverage
  * [ ] Benchmarking

# Instance

## Settings

```yaml
###########################
##### INSTANCE CONFIG #####
###########################

# Config pertaining to instance federation settings, pages to hide/expose, etc.

# Array of string. BCP47 language tags to indicate preferred languages of users on this instance.
#
# If you provide these, you should provide these in order from most-preferred to least-preferred,
# but note that leaving out a language from this array doesn't mean it can't be used on this instance,
# it only means it won't be advertised as a preferred instance language.
#
# It is valid to provide no entries here; your instance will then have no particular preferred language.
#
# See here for commonly-used tags: https://en.wikipedia.org/wiki/IETF_language_tag#List_of_common_primary_language_subtags
# See here for all current tags: https://www.iana.org/assignments/language-subtag-registry/language-subtag-registry
#
# Example: ["nl", "en-gb", "fr"]
# Default: []
instance-languages: []

# String. Federation mode to use for this instance.
#
# "blocklist" -- open federation by default. Only instances that are explicitly 
#                blocked will be denied (unless they are also explicitly allowed).
#
# "allowlist" -- closed federation by default. Only instances that are explicitly
#                allowed will be able to interact with this instance.
#
# For more details on blocklist and allowlist modes, check the documentation at:
# https://docs.gotosocial.org/en/latest/admin/federation_modes
#
# Options: ["blocklist", "allowlist"]
# Default: "blocklist"
instance-federation-mode: "blocklist"

# Bool. Enable spam filtering heuristics for messages entering your instance
# via the federation API. Regardless of what you set here, basic checks
# for message relevancy will still be performed, but you can try enabling
# this setting if you are being spammed with unwanted messages from other
# instances, and want to more strictly filter out spam messages.
#
# THIS IS CURRENTLY AN EXPERIMENTAL SETTING, AND MAY FILTER OUT LEGITIMATE
# MESSAGES, OR FAIL TO FILTER OUT SPAMMY MESSAGES. It is recommended to
# only enable this setting when the fediverse is in the midst of a spam
# wave, and you need to batten down the hatches to keep your instance usable.
#
# The decision of whether a message counts as spam or not is made based on
# the following heuristics, in order, where receiver = the account on your
# instance that received a message in their inbox, and requester = the
# account on a remote instance that sent the message.
#
# First, basic relevancy checks
#
#  1. Receiver follows requester. Return OK.
#  2. Statusable doesn't mention receiver. Return NotRelevant.
#
# If instance-federation-spam-filter = false, then return OK now.
# Otherwise check:
#
#  3. Receiver is locked and is followed by requester. Return OK.
#  4. Five or more people are mentioned. Return Spam.
#  5. Receiver follow (requests) a mentioned account. Return OK.
#  6. Statusable has a media attachment. Return Spam.
#  7. Statusable contains non-mention, non-hashtag links. Return Spam.
#
# Messages identified as spam will be dropped from your instance, and not
# inserted into the database, or into home timelines or notifications.
#
# Options: [true, false]
# Default: false
instance-federation-spam-filter: false

# Bool. Allow unauthenticated users to make queries to /api/v1/instance/peers?filter=open in order
# to see a list of instances that this instance 'peers' with. Even if set to 'false', then authenticated
# users (members of the instance) will still be able to query the endpoint.
# Options: [true, false]
# Default: false
instance-expose-peers: false

# Bool. Allow unauthenticated users to make queries to /api/v1/instance/peers?filter=suspended in order
# to see a list of instances that this instance blocks/suspends. Even if set to 'false', then authenticated
# users (members of the instance) will still be able to query the endpoint.
#
# WARNING: Setting this variable to 'true' may result in your instance being scraped by blocklist scrapers.
# See: https://docs.gotosocial.org/en/latest/admin/domain_blocks/#block-announce-bots
#
# Options: [true, false]
# Default: false
instance-expose-suspended: false

# Bool. Allow unauthenticated users to view /about/suspended,
# showing the HTML rendered list of instances that this instance blocks/suspends.
# Options: [true, false]
# Default: false
instance-expose-suspended-web: false

# Bool. Allow unauthenticated users to make queries to /api/v1/timelines/public in order
# to see a list of public posts on this server. Even if set to 'false', then authenticated
# users (members of the instance) will still be able to query the endpoint.
# Options: [true, false]
# Default: false
instance-expose-public-timeline: false

# Bool. This flag tweaks whether GoToSocial will deliver ActivityPub messages
# to the shared inbox of a recipient, if one is available, instead of delivering
# each message to each actor who should receive a message individually.
#
# Shared inbox delivery can significantly reduce network load when delivering
# to multiple recipients share an inbox (eg., on large Mastodon instances).
#
# See: https://www.w3.org/TR/activitypub/#shared-inbox-delivery
#
# Options: [true, false]
# Default: true
instance-deliver-to-shared-inboxes: true

# Bool. This flag will inject a Mastodon version into the version field that
# is included in /api/v1/instance. This version is often used by Mastodon clients
# to do API feature detection. By injecting a Mastodon compatible version, it is
# possible to cajole those clients to behave correctly with GoToSocial.
#
# Options: [true, false]
# Default: false
instance-inject-mastodon-version: false

# String. 24hr time of day formatted as hh:mm.
# Examples: ["14:30", "00:00", "04:00"]
# Default: "23:00" (11pm).
instance-subscriptions-process-from: "23:00"

# Duration. Period between subscription updates.
# Examples: ["24h", "72h", "12h"]
# Default: "24h" (once per day).
instance-subscriptions-process-every: "24h"

# String. Allows you to customize if and how stats are served to
# crawlers at the /api/v1|v2/instance and /nodeinfo endpoints.
#
# Note that no matter what you set below, the /api/v1|v2/instance
# endpoints will not be allowed by robots.txt, as these are client
# API endpoints.
#
# "" / empty string (default mode): Serve accurate stats at instance
# and nodeinfo endpoints, and DISALLOW crawlers from crawling
# those endpoints in robots.txt. This mode is equivalent to politely
# asking crawlers not to crawl, but there's no guarantee they will obey,
# as unfortunately many crawlers don't even check robots.txt.
#
# "zero": Serve zeroed-out stats at instance and nodeinfo endpoints,
# and DISALLOW crawlers from crawling those endpoints in robots.txt.
# This mode prevents even ill-behaved crawlers from gathering stats
# about your instance, as all gathered values will be 0. This is the
# safest way of preserving your instance's privacy in terms of stats.
#
# "serve": Serve accurate stats at instance and nodeinfo endpoints,
# and ALLOW crawlers to crawl those endpoints. This mode is useful
# if you want to contribute to fediverse statistics collection projects.
#
# "baffle": Serve randomized, preposterous stats at instance and nodeinfo
# endpoints, and DISALLOW crawlers from crawling those endpoints in robots.txt.
# This mode can be useful to annoy crawlers that don't respect robots.txt.
# Warning that this may draw the ire of crawler implementers who don't
# respect robots.txt, and may therefore put a target on your instance.
#
# Options: ["", "zero", "serve", "baffle"]
# Default: ""
instance-stats-mode: ""
```

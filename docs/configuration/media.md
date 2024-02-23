# Media

## Settings

```yaml
########################
##### MEDIA CONFIG #####
########################

# Config pertaining to media uploads (videos, image, image descriptions, emoji).

# Int. Maximum allowed image upload size in bytes.
# Examples: [2097152, 10485760]
# Default: 10485760 -- aka 10MB
media-image-max-size: 10485760

# Int. Maximum allowed video upload size in bytes.
# Examples: [2097152, 10485760]
# Default: 41943040 -- aka 40MB
media-video-max-size: 41943040

# Int. Minimum amount of characters required as an image or video description.
# Examples: [500, 1000, 1500]
# Default: 0 (not required)
media-description-min-chars: 0

# Int. Maximum amount of characters permitted in an image or video description.
# Examples: [1000, 1500, 3000]
# Default: 1500
media-description-max-chars: 1500

# Int. Max size in bytes of emojis uploaded to this instance via the admin API.
# The default is the same as the Mastodon size limit for emojis (50kb), which allows
# for good interoperability. Raising this limit may cause issues with federation
# of your emojis to other instances, so beware.
# Examples: [51200, 102400]
# Default: 51200
media-emoji-local-max-size: 51200

# Int. Max size in bytes of emojis to download from other instances.
# By default this is 100kb, or twice the size of the default for media-emoji-local-max-size.
# This strikes a good balance between decent interoperability with instances that have
# higher emoji size limits, and not taking up too much space in storage.
# Examples: [51200, 102400]
# Default: 102400
media-emoji-remote-max-size: 102400

# The below media cleanup settings allow admins to customize when and
# how often media cleanup + prune jobs run, while being set to a fairly
# sensible default (every night @ midnight). For more information on exactly
# what these settings do, with some customization examples, see the docs:
# https://docs.gotosocial.org/en/latest/admin/media_caching#cleanup

# Int. Number of days to cache media from remote instances before
# they are removed from the cache. When remote media is removed from
# the cache, it is deleted from storage but the database entries for
# the media are kept so that it can be fetched again if requested by a user.
#
# If this is set to 0, then media from remote instances will be cached indefinitely.
#
# Examples: [30, 60, 7, 0]
# Default: 7
media-remote-cache-days: 7

# String. 24hr time of day formatted as hh:mm.
# Examples: ["14:30", "00:00", "04:00"]
# Default: "00:00" (midnight). 
media-cleanup-from: "00:00"

# Duration. Period between media cleanup runs.
# More than once per 24h is not recommended
# is likely overkill. Setting this to something
# very low like once every 10 minutes will probably
# cause lag and possibly other issues.
# Examples: ["24h", "72h", "12h"]
# Default: "24h" (once per day).
media-cleanup-every: "24h"
```

# Media

## Settings

```yaml
########################
##### MEDIA CONFIG #####
########################

# Config pertaining to user media uploads (videos, image, image descriptions).

# Int. Maximum allowed image upload size in bytes.
# Examples: [2097152, 10485760]
# Default: 2097152 -- aka 2MB
media-image-max-size: 2097152

# Int. Maximum allowed video upload size in bytes.
# Examples: [2097152, 10485760]
# Default: 10485760 -- aka 10MB
media-video-max-size: 10485760

# Int. Minimum amount of characters required as an image or video description.
# Examples: [500, 1000, 1500]
# Default: 0 (not required)
media-description-min-chars: 0

# Int. Maximum amount of characters permitted in an image or video description.
# Examples: [500, 1000, 1500]
# Default: 500
media-description-max-chars: 500

# Int. Number of days to cache media from remote instances before they are removed from the cache.
# A job will run every day at midnight to clean up any remote media older than the given amount of days.
#
# When remote media is removed from the cache, it is deleted from storage but the database entries for the media
# are kept so that it can be fetched again if requested by a user.
#
# If this is set to 0, then media from remote instances will be cached indefinitely.
# Examples: [30, 60, 7, 0]
# Default: 30
media-remote-cache-days: 30
```

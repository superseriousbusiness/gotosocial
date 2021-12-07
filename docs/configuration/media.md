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
```

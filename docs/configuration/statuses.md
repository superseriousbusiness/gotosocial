# Statuses

## Settings

```yaml
###########################
##### STATUSES CONFIG #####
###########################

# Config pertaining to the creation of statuses/posts, and permitted limits.

# Int. Maximum amount of characters permitted for a new status.
# Note that going way higher than the default might break federation.
# Examples: [140, 500, 5000]
# Default: 5000
statuses-max-chars: 5000

# Int. Maximum amount of characters allowed in the CW/subject header of a status.
# Note that going way higher than the default might break federation.
# Examples: [100, 200]
# Default: 100
statuses-cw-max-chars: 100

# Int. Maximum amount of options to permit when creating a new poll.
# Note that going way higher than the default might break federation.
# Examples: [4, 6, 10]
# Default: 6
statuses-poll-max-options: 6

# Int. Maximum amount of characters to permit per poll option when creating a new poll.
# Note that going way higher than the default might break federation.
# Examples: [50, 100, 150]
# Default: 50
statuses-poll-option-max-chars: 50

# Int. Maximum amount of media files that can be attached to a new status.
# Note that going way higher than the default might break federation.
# Examples: [4, 6, 10]
# Default: 6
statuses-media-max-files: 6
```

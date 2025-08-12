# Statuses

## Settings

```yaml
###########################
##### STATUSES CONFIG #####
###########################

# Config pertaining to the creation of statuses/posts, and permitted limits.

# Int. Maximum amount of characters permitted for a new status,
# including the content warning (if set).
#
# Note that going way higher than the default might break federation.
#
# Examples: [140, 500, 5000]
# Default: 5000
statuses-max-chars: 5000

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

# Int. Maximum number of statuses a user can schedule at time.
# Examples: [300]
# Default: 300
scheduled-statuses-max-total: 300

# Int. Maximum number of statuses a user can schedule for a single day.
# Examples: [25]
# Default: 25
scheduled-statuses-max-daily: 25
```

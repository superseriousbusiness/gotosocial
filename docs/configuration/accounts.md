# Accounts

## Settings

```yaml
###########################
##### ACCOUNTS CONFIG #####
###########################

# Config pertaining to creation and maintenance of accounts on the server, as well as defaults for new accounts.

# Bool. Allow people to submit new sign-up / registration requests via the form at /signup.
#
# Options: [true, false]
# Default: false
accounts-registration-open: false

# Bool. Are sign up requests required to submit a reason for the request (eg., an explanation of why they want to join the instance)?
# Options: [true, false]
# Default: true
accounts-reason-required: true

# Int. Number of approved sign-ups allowed within
# 24hrs before new account registration is closed.
#
# Leaving this count at the default essentially limits
# your instance to growing by 10 accounts per day.
#
# Setting this number to 0 or less removes the limit.
#
# Default: 10
accounts-registration-daily-limit: 10

# Int. Number of new account sign-ups allowed in the pending
# approval queue before new account registration is closed.
#
# This can be used to essentially "throttle" the sign-up
# queue to prevent instance admins becoming overwhelmed.
#
# Setting this number to 0 or less removes the limit.
#
# Default: 20
accounts-registration-backlog-limit: 20

# Bool. Allow accounts on this instance to set custom CSS for their profile pages and statuses.
# Enabling this setting will allow accounts to upload custom CSS via the /user settings page,
# which will then be rendered on the web view of the account's profile and statuses.
#
# For instances with public sign ups, it is **HIGHLY RECOMMENDED** to leave this setting on 'false',
# since setting it to true allows malicious accounts to make their profile pages misleading, unusable
# or even dangerous to visitors. In other words, you should only enable this setting if you trust
# the users on your instance not to produce harmful CSS.
#
# Regardless of what this value is set to, any uploaded CSS will not be federated to other instances,
# it will only be shown on profiles and statuses on *this* instance.
#
# Options: [true, false]
# Default: false
accounts-allow-custom-css: false

# Int. If accounts-allow-custom-css is true, this is the permitted length in characters for
# CSS uploaded by accounts on this instance. No effect if accounts-allow-custom-css is false.
#
# Examples: [500, 5000, 9999]
# Default: 10000
accounts-custom-css-length: 10000
```

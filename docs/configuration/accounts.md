# Accounts

## Settings

```yaml
###########################
##### ACCOUNTS CONFIG #####
###########################

# Config pertaining to creation and maintenance of accounts on the server, as well as defaults for new accounts.

# Bool. Do we want people to be able to just submit sign up requests, or do we want invite only?
# Options: [true, false]
# Default: true
accounts-registration-open: true

# Bool. Do sign up requests require approval from an admin/moderator before an account can sign in/use the server?
# Options: [true, false]
# Default: true
accounts-approval-required: true

# Bool. Are sign up requests required to submit a reason for the request (eg., an explanation of why they want to join the instance)?
# Options: [true, false]
# Default: true
accounts-reason-required: true

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
```

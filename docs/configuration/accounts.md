# Accounts

## Settings

```yaml
###########################
##### ACCOUNTS CONFIG #####
###########################

# Config pertaining to creation and maintenance of accounts on the server, as well as defaults for new accounts.
accounts:

  # Bool. Do we want people to be able to just submit sign up requests, or do we want invite only?
  # Options: [true, false]
  # Default: true
  openRegistration: true

  # Bool. Do sign up requests require approval from an admin/moderator before an account can sign in/use the server?
  # Options: [true, false]
  # Default: true
  requireApproval: true

  # Bool. Are sign up requests required to submit a reason for the request (eg., an explanation of why they want to join the instance)?
  # Options: [true, false]
  # Default: true
  reasonRequired: true
```

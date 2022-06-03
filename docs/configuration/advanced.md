# Advanced

Advanced settings options are provided for the sake of allowing admins to tune their instance to their liking.

These are set to sensible defaults, so most server admins won't need to touch them or think about them.

**Changing these settings if you don't know what you're doing may break your instance**.

## Settings

```yaml
#############################
##### ADVANCED SETTINGS #####
#############################

# Advanced settings pertaining to http timeouts, security, cookies, and more.
#
# ONLY ADJUST THESE SETTINGS IF YOU KNOW WHAT YOU ARE DOING!
#
# Most users will not need to (and should not) touch these settings, since
# they are set to sensible defaults, and may break if they are changed.
#
# Nevertheless, they are provided for the sake of allowing server admins to
# tweak their instance for performance or security reasons.

# String. Value of the SameSite attribute of cookies set by GoToSocial.
# Defaults to 'lax' to ensure that the OIDC flow does not break, which is
# fine in most cases. If you want to harden your instance against CSRF attacks
# and don't mind if some login-related things might break, you can set this
# to 'strict' instead.
#
# For an overview of what this does, see:
# https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Set-Cookie/SameSite
#
# Options: ["lax", "strict"]
# Default: "lax"
advanced-cookies-samesite: "lax"
```

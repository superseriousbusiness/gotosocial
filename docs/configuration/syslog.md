# Syslog

GoToSocial can be configured to mirror logs to [syslog](https://en.wikipedia.org/wiki/Syslog), either via udp/tcp, or a local syslog (eg., `/var/log/syslog`).

This is useful if you want to daemonize GtS and not handle log rotations etc yourself but rely on a proven implementation.

Logs in syslog will look something like this:

```text
Dec 12 17:44:03 dilettante ./gotosocial[246860]: time=2021-12-12T17:44:03+01:00 level=info msg=connected to SQLITE database
Dec 12 17:44:03 dilettante ./gotosocial[246860]: time=2021-12-12T17:44:03+01:00 level=info msg=there are no new migrations to run func=doMigration
```

## Settings

```yaml
#########################
##### SYSLOG CONFIG #####
#########################

# Config for additional syslog log hooks. See https://en.wikipedia.org/wiki/Syslog,
# and https://github.com/sirupsen/logrus/tree/master/hooks/syslog.
#
# These settings are useful when one wants to daemonize GoToSocial and send logs
# to a specific place, either a local location or a syslog server. Most users will
# not need to touch these settings.

# Bool. Enable the syslog logging hook. Logs will be mirrored to the configured destination.
# Options: [true, false]
# Default: false
syslog-enabled: false

# String. Protocol to use when directing logs to syslog. Leave empty to connect to local syslog.
# Options: ["udp", "tcp", ""]
# Default: "tcp"
syslog-protocol: "udp"

# String. Address:port to send syslog logs to. Leave empty to connect to local syslog.
# Default: "localhost:514"
syslog-address: "localhost:514"
```

# Email Config (smtp)

GoToSocial supports sending emails to users via the [Simple Mail Transfer Protocol](https://wikipedia.org/wiki/Simple_Mail_Transfer_Protocol) or **smtp**.

Configuring GoToSocial to send emails is **not required** in order to have a properly running instance. Still, it's very useful for doing things like sending confirmation emails and notifications, and handling password reset requests.

In order to make GoToSocial email sending work, you need an smtp-compatible mail service running somewhere, either as a server on the same machine that GoToSocial is running on, or via an external service like [Mailgun](https://mailgun.com). It may also be possible to use a free personal email address for sending emails, if your email provider supports smtp (check with them--most do), but you might run into trouble sending lots of emails.

## Settings

The configuration options for smtp are as follows:

```yaml
#######################
##### SMTP CONFIG #####
#######################

# Config for sending emails via an smtp server. See https://en.wikipedia.org/wiki/Simple_Mail_Transfer_Protocol

# String. The hostname of the smtp server you want to use.
# If this is not set, smtp will not be used to send emails, and you can ignore the other settings.
# Examples: ["mail.example.org", "localhost"]
# Default: ""
smtp-host: ""

# Int. Port to use to connect to the smtp server.
# Examples: []
# Default: 0
smtp-port: 0

# String. Username to use when authenticating with the smtp server.
# This should have been provided to you by your smtp host.
# This is often, but not always, an email address.
# Examples: ["maillord@example.org"]
# Default: ""
smtp-username: ""

# String. Password to use when authenticating with the smtp server.
# This should have been provided to you by your smtp host.
# Examples: ["1234", "password"]
# Default: ""
smtp-password: ""

# String. 'From' address for sent emails.
# Examples: ["mail@example.org"]
# Default: ""
smtp-from: ""

# Bool. If true, when an email is sent that has multiple recipients, each recipient
# will be included in the To field, so that each recipient can see who else got the
# email, and they can 'reply all' to the other recipients if they want to.
#
# If false, email will be sent to Undisclosed Recipients, and each recipient will not
# be able to see who else received the email.
#
# It might be useful to change this setting to 'true' if you want to be able to discuss
# new moderation reports with other admins by 'replying-all' to the notification email.
# Default: false
smtp-disclose-recipients: false
```

Note that if you don't set `Host`, then email sending via smtp will be disabled, and the other settings will be ignored. GoToSocial will still log (at trace level) emails that *would* have been sent if smtp was enabled.

## Behavior

### SSL

GoToSocial requires your smtp server to present valid SSL certificates. Most of the big services like Mailgun do this anyway, but if you're running your own mail server without SSL for some reason, and you're trying to connect GoToSocial to it, it will not work.

The exception to this requirement is if you're running your mail server (or bridge to a mail server) on `localhost`, in which case SSL certs are not required.

### When are emails sent?

Currently, emails are sent:

- To the provided email address of a new user to request email confirmation when a new account is created via the API.
- To all active instance moderators + admins when a new moderation report is received. By default, recipients are Bcc'd, but you can change this behavior with the setting `smtp-disclose-recipients`.
- To the creator of a report (on this instance) when the report is closed by a moderator.

### Can I test if my SMTP configuration is correct?

Yes, you can use the API to send a test email to yourself. Check the API documentation for the `/api/v1/admin/email/test` endpoint.

### HTML versus Plaintext

Emails are sent in plaintext by default. At this point, there is no option to send emails in html, but this is something that might be added later if there's enough demand for it.

## Customization

If you like, you can customize the templates that are used for generating emails. Follow the examples in `web/templates`.

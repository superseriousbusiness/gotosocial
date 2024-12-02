# New Account Sign-Ups

If you want to allow more people than just you to have an account on your instance, you can open your instance to new account sign-ups / registrations.

Be wary that as instance admin, like it or not, you are responsible for what people post on your instance. If users on your instance harass or annoy other people on the fediverse, you may find your instance gets a bad reputation and becomes blocked by others. Moderating a space properly takes work. As such, you should carefully consider whether or not you are willing and able to do moderation, and consider accepting sign-ups on your instance only from friends and people that you really trust.

!!! warning
    For the sign-up flow to work as intended, your instance [should be configured to send emails](../configuration/smtp.md).
    
    As mentioned below, several emails are sent during the sign-up flow, both to you (as admin/moderator) and to the applicant, including an email asking them to confirm their email address.
    
    If they cannot receive this email (because your instance is not configured to send emails), you will have to manually confirm the account by [using the CLI tool](../admin/cli.md#gotosocial-admin-account-confirm).

## Opening Sign-Ups

You can open new account sign-ups for your instance by changing the variable `accounts-registration-open` to `true` in your [configuration](../configuration/accounts.md), and restarting your GoToSocial instance.

A sign-up form for your instance will be available at the `/signup` endpoint. For example, `https://your-instance.example.org/signup`.

![Sign-up form, showing email, password, username, and reason fields.](../public/signup-form.png)

Also, your instance homepage and "about" pages will be updated to reflect that registrations are open.

When someone submits a new sign-up, they'll receive an email at the provided email address, giving them a link to confirm that the address really belongs to them.

In the meantime, admins and moderators on your instance will receive an email and a notification that a new sign-up has been submitted.

## Handling Sign-Ups

Instance admins and moderators can handle a new sign-up by either approving or rejecting it via the "accounts" -> "pending" section in the admin panel.

![Admin settings panel open to "accounts" -> "pending", showing one account in a list.](../public/signup-pending.png)

If you have no sign-ups, the list pictured above will be empty. If you have a pending account sign-up, however, you can click on it to open that account in the account details screen:

![Details of a new pending account, giving options to approve or reject the sign-up.](../public/signup-account.png)

At the bottom, you will find actions that let you approve or reject the sign-up.

If you **approve** the sign-up, the account will be marked as "approved", and an email will be sent to the applicant informing them their sign-up has been approved, and reminding them to confirm their email address if they haven't already done so. If they have already confirmed their email address, they will be able to log in and start using their account.

If you **reject** the sign-up, you may wish to inform the applicant that their sign-up has been rejected, which you can do by ticking the "send email" checkbox. This will send a short email to the applicant informing them of the rejection. If you wish, you can add a custom message, which will be added at the bottom of the email. You can also add a private note that will be visible to other admins only.

!!! warning
    You may want to hold off on approving a sign-up until they have confirmed their email address, in case the applicant made a typo when submitting, or the email address they provided does not actually belong to them. If they cannot confirm their email address, they will not be able to log in and use their account.

## Sign-Up Limits

To avoid sign-up backlogs overwhelming admins and moderators, GoToSocial limits the sign-up pending backlog to 20 accounts. Once there are 20 accounts pending in the backlog waiting to be handled by an admin or moderator, new sign-ups will not be accepted via the form.

New sign-ups will also not be accepted via the form if 10 or more new account sign-ups have been approved in the last 24 hours, to avoid instances rapidly expanding beyond the capabilities of moderators.

In both cases, applicants will be shown an error message explaining why they could not submit the form, and inviting them to try again later.

To combat spam accounts, GoToSocial account sign-ups **always** require manual approval by an administrator, and applicants must **always** confirm their email address before they are able to log in and post.

## Sign-Up Via Invite

NOT IMPLEMENTED YET: in a future update, admins and moderators will be able to create and send invites that allow accounts to be created even when public sign-up is closed, and to pre-approve accounts created via invitation, and/or allow them to override the sign-up limits described above.

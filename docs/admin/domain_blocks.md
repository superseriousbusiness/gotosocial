# Domain Blocks

GoToSocial supports 'blocking'/'suspending' domains that you don't want your instance to federate with. In our documentation, the two terms 'block' and 'suspend' are used interchangeably with regard to domains, because they mean the same thing: preventing your instance and the instance running on the target domain from communicating with one another, effectively cutting off federation between the two instances.

You can view, create, and remove domain blocks and domain allows using the [instance admin panel](./settings.md#domain-permissions).

This document focuses on what domain blocks actually *do* and what side effects are processed when you create a new domain block.

## How does a domain block work

A domain block works by doing two things:

Firstly, it instructs your instance to refuse any requests made to it from the target domain:

- All incoming requests from the blocked domain to your instance will be responded to with HTTP status code `403 Forbidden`.
- This makes it impossible for an account on the target domain to interact with an account on your instance, or any statuses created by that account, since your instance will simply refuse to process the request.
- This also extends to GET requests: your instance will no longer serve an ActivityPub response to a request by a blocked instance to fetch, say, an account's bio, or pinned statuses, etc.
- Boosts of statuses from accounts on your instance should also not be visible to accounts on blocked instances, since those instances will not be able to fetch the content of the status that has been boosted.

Secondly, a domain block instructs your instance to no longer make any requests to the target instance. This means:

- Your instance will not deliver any messages to an instance on a blocked domain.
- Nor will it fetch statuses, accounts, media, or emojis from that instance.

## Safety concerns

### Block evasion

Domain blocking is not airtight. GoToSocial *can* ensure that it will neither serve requests from nor make requests to instances on blocked domains. Unfortunately it *cannot* guarantee that accounts on your instance will never be visible in any way to users with accounts on blocked instances. Consider the following circumstances, all of which represent a form of [block evasion](https://en.wikipedia.org/wiki/Block_(Internet)#Evasion):

- You've domain blocked `blocked.instance.org`. A user on `blocked.instance.org` makes an account on `not-blocked.domain`, so that they can use their new account to interact with your posts or send messages to you. They may be upfront about who they are, or they may use a false identity.
- You've domain blocked `blocked.instance.org`. A user on `not-blocked.domain` screenshots a post of yours and sends it to someone on `blocked.instance.org`.
- You've domain blocked `blocked.instance.org`. A user on `blocked.instance.org` visits the web view of your profile to read your public posts.
- You've domain blocked `blocked.instance.org`. You have RSS enabled for your profile. A user from `blocked.instance.org` subscribes to your RSS feed to read your public posts.

In the above cases, `blocked.instance.org` remains blocked, but users from that instance may still have other ways of seeing your posts and possibly reaching you.

With this in mind, you should only ever treat domain blocking as *one layer* of your privacy onion. That is, domain blocking should be deployed alongside other layers in order to achieve a level of privacy that you are comfortable with. This ought to include things like not posting sensitive information publicly, not accidentally doxxing yourself in photos, etc.

### Block announce bots

Unfortunately, the Fediverse has its share of trolls, many of whom see domain blocking as an adversary to be defeated. To achieve this, they often target instances which use domain blocks to protect users.

As such, there are bots on the Fediverse which scrape instance domain blocks and announce any discovered blocks to the followers of the bot, opening the admin of the blocking instance up to harassment. These bots use the `api/v1/instance/peers?filter=suspended` endpoint of GoToSocial instances to gather domain block information.

By default, GoToSocial does not expose this endpoint publicly, so your instance will be safe from such scraping. However, if you set `instance-expose-suspended` to `true` in your config.yaml file, you may find that this endpoint gets scraped occasionally, and you may see your blocks being announced by troll bots.

## What are the side effects of creating a domain block

When you create a new domain block (or resubmit an existing domain block), your instance will process side effects for the block. These side effects are:

1. Mark all accounts stored in your database from the target domain as suspended, and remove most information (bio, display name, fields, etc) from each account marked this way.
2. Clear all mutual and one-way relationships between local accounts and suspended accounts (followed, following, follow requests, bookmarks, etc).
3. Delete all statuses from suspended accounts.
4. Delete all media from suspended accounts and their statuses, including media attachments, avatars, headers, and emojis.

!!! danger
    Currently, most of the above side effects are **irreversible**. If you unblock a domain after blocking it, all accounts on that domain will be marked as no longer suspended, and you will be able to interact with them again, but all relationships will still be wiped out, and all statuses and media will be gone.
    
    Think carefully before blocking a domain.

## Blocking a domain and all subdomains

When you add a new domain block, GoToSocial will also block all subdomains of the blocked domain. This allows you to block specific subdomains, if you wish, or to block a domain more generally if you don't trust the domain owner.

Some examples:

1. You block `example.org`. This blocks the following domains (not exhaustive): `example.org`, `subdomain.example.org`, `another-subdomain.example.org`, `sub.sub.sub.domain.example.org`.
2. You block `baddies.example.org`. This blocks the following domains (not exhaustive): `baddies.example.org`, `really-bad.baddies.example.org`. However the following domains are not blocked (not exhaustive): `example.org`, `subdomain.example.org`, `not-baddies.example.org`.

A more practical example:

Some absolute jabroni owns the domain `fossbros-anonymous.io`. Not only do they run a Mastodon instance at `mastodon.fossbros-anonymous.io`, they also have a GoToSocial instance at `gts.fossbros-anonymous.io`, and an Akkoma instance at `akko.fossbros-anonymous.io`. You want to block all of these instances at once (and any future instances they might create at, say, `pl.fossbros-anonymous.io`, etc). You can do this by simply creating a domain block for `fossbros-anonymous.io`. None of the instances at subdomains will be able to communicate with your instance. Yeet!

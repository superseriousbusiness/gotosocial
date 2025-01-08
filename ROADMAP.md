# Roadmap to Beta <!-- omit in toc -->

This document contains the roadmap for GoToSocial to be considered eligible for its first proper [stable release](https://en.wikipedia.org/wiki/Software_release_life_cycle#Stable_release).

All the info contained in this document is best-guess only. It's useful to have a rough timeline we can direct people to, but things will undoubtedly change along the way; don't hold us to anything in this doc!

Thank you to [NLnet](https://nlnet.nl) for helping to fund the alpha and beta phases of GoToSocial development!

Big thank you to all of our [Open Collective](https://opencollective.com/gotosocial) and [Liberapay](https://liberapay.com/gotosocial) contributors, who've helped us keep the lights on! 💕 

## Table of Contents <!-- omit in toc -->

- [Beta Aims](#beta-aims)
- [Timeline](#timeline)
  - [Mid 2023](#mid-2023)
  - [Mid/late 2023](#midlate-2023)
  - [Early 2024](#early-2024)
  - [BETA milestone](#beta-milestone)
  - [Remainder 2024 - early 2025](#remainder-2024---early-2025)
  - [On the way out of BETA to STABLE RELEASE](#on-the-way-out-of-beta-to-stable-release)
- [Wishlist](#wishlist)

## Beta Aims

Every software project has a different idea of what it means to be "in beta". In our case, we want the beta version of GoToSocial to provide a feature set that roughly compares to existing popular ActivityPub server implementations.

In other words, you should be able to use the beta version of GoToSocial as your go-to(!) fedi instance for following people and making posts, without running into too many issues with things being missing or not working properly.

Our target for beta also includes features which, we believe, are vital for user safety and wellbeing, such as non-replyable posting, block list subscription, allow-list support, and so on.

Once we have the features we want in order to say "this is now a beta", we will use beta time to work on fixing bugs, tuning performance, and adding extra features which require a stable base to build on.

Our hope for beta is also that, once we're there, the client API will stay relatively stable, and folks can confidently start building applications on top of GoToSocial without worrying that the API will change overnight and necessitate major rewrites.

We currently foresee entering beta phase around the start of 2024, though this is only an estimate and is subject to change.

## Timeline

What follows is a rough timeline of features that will be implemented on the road to beta. The timeline is calculated on the following assumptions:

- We will continue to develop at a pace that is similar to what we've done over the previous two years.
- Our combined bandwidth is roughly equivalent to one person working full time on the project.
- One distinct 'feature' takes one person 2-4 weeks to develop and test, depending on the size of the feature.
- There will be other bugs to fix in between implementing things, so we shouldn't pack in features too tightly.

**This timeline is a best-guess about when things will be implemented. The order of feature releases is not fixed. Things may go faster or slower depending on the number of hurdles we run into, and the amount of help we receive from community contributions of code. The timeline also does not include background tasks like admin, polishing existing features, refactoring code, release management, and ensuring compatibility with other AP implementations.**

### Mid 2023

- [x] **Hashtags** -- implement federating hashtags and viewing hashtags to allow users to discover posts that they might be interested in. (Done! https://github.com/superseriousbusiness/gotosocial/pull/2032).

### Mid/late 2023

- [x] **Polls** -- implementing parsing, creating, and voting in polls. (Done! https://github.com/superseriousbusiness/gotosocial/pull/2330)
- [x] **Mute posts/threads** -- opt-out of notifications for replies to a thread; no longer show a given post in your timeline. (Done! https://github.com/superseriousbusiness/gotosocial/pull/2278)
- [x] **Limited peering/allowlists** -- allow instance admins to limit federation with other instances by default. (Done! https://github.com/superseriousbusiness/gotosocial/pull/2200)

### Early 2024

- [x] **Move activity** -- use the ActivityPub `Move` activity to support migration of a user's profile across servers.
- [x] **Sign-up flow** -- allow users to submit a sign-up request to an instance; allow admins to moderate sign-up requests.

### BETA milestone

Completion of all above features indicates that we are now in the BETA phase of GoToSocial. We foresee this happening around Feb/March 2024. EDIT: It ended up happening in September/October 2024, whoops!

### Remainder 2024 - early 2025

These are provided in no specific order.

- [x] **Filters v2** -- implement v2 of the filters API.
- [x] **Mute accounts** -- mute accounts to prevent their posts showing up in your home timeline (optional: for limited period of time).
- [x] **Non-replyable posts** -- design a non-replyable post path for GoToSocial based on https://github.com/mastodon/mastodon/issues/14762#issuecomment-1196889788; allow users to create non-replyable posts.
- [x] **Block + allow list subscriptions** -- allow instance admins to subscribe their instance to domain block/allow lists.
- [x] **Direct conversation view** -- allow users to easily page through all direct-message conversations they're a part of.
- [ ] **Oauth token management** -- create / view / invalidate OAuth tokens via the settings panel.
- [ ] **Status EDIT support** -- edit statuses that you've created, without having to delete + redraft. Federate edits out properly.
- [ ] **Fediverse relay support** -- publish posts to relays, pull posts from relays.
- [ ] **Two factor authentication (2fa)** -- allow users to enable 2FA for their account via the settings panel, enforce 2FA on login.
- [ ] **Moderation: Append content warning / mark-as-sensitive all content from an instance/account**.

More tbd!

### On the way out of BETA to STABLE RELEASE

Tbd.

## Wishlist

These cool things will be implemented if time allows (because we really want them):

- **Groups** and group posting!
- Reputation-based 'slow' federation.
- Community decision-making for federation and moderation actions.
- User-selectable custom templates for rendering public posts:
  - Twitter-style
  - Blogpost
  - Gallery
  - Etc.

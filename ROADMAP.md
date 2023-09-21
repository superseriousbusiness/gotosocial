# Roadmap to Beta <!-- omit in toc -->

This document contains the roadmap for GoToSocial to be considered eligible for its first [beta release](https://en.wikipedia.org/wiki/Software_release_life_cycle#Beta).

All the info contained in this document is best-guess only. It's useful to have a rough timeline we can direct people to, but things will undoubtedly change along the way; don't hold us to anything in this doc!

Thank you to [NLnet](https://nlnet.nl) for helping to fund the alpha phase of GoToSocial development and get us moving towards beta!

Big thank you to all of our [Open Collective](https://opencollective.com/gotosocial) and [Liberapay](https://liberapay.com/gotosocial) contributors, who've helped us keep the lights on! 💕 

## Table of Contents <!-- omit in toc -->

- [Beta Aims](#beta-aims)
- [Timeline](#timeline)
  - [Mid 2023](#mid-2023)
  - [Mid/late 2023](#midlate-2023)
  - [Late 2023](#late-2023)
  - [Early 2024](#early-2024)
  - [And then...](#and-then)

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
- [ ] **Block list subscriptions** -- allow instance admins to subscribe their instance to plaintext domain block lists (much of the work for this is already in place).
- [ ] **Direct conversation view** -- allow users to easily page through all direct-message conversations they're a part of.

### Mid/late 2023

- [ ] **Polls** -- implementing parsing, creating, and voting in polls.
- [ ] **Mute posts/threads** -- opt-out of notifications for replies to a thread; no longer show a given post in your timeline.
- [x] **Limited peering/allowlists** -- allow instance admins to limit federation with other instances by default. (Done! https://github.com/superseriousbusiness/gotosocial/pull/2200)

### Late 2023

- [ ] **Move activity** -- use the ActivityPub `Move` activity to support migration of a user's profile across servers.
- [ ] **Sign-up flow** -- allow users to submit a sign-up request to an instance; allow admins to moderate sign-up requests.

### Early 2024

- [ ] **Non-replyable posts** -- design a non-replyable post path for GoToSocial based on https://github.com/mastodon/mastodon/issues/14762#issuecomment-1196889788; allow users to create non-replyable posts.

### And then...

BETA TIME baby!

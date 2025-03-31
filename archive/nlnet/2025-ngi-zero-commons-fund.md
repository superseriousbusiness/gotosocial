# NLnet Grant Application - NGI Zero Commons 2025

This document is the application on behalf of GoToSocial for funding from the [NLnet](https://nlnet.nl) [NGI Zero Commons fund](https://nlnet.nl/commonsfund/), April 2025. See [here](https://nlnet.nl/propose/).

## General Project Information

> Project Name

GoToSocial

> Website / wiki

https://github.com/superseriousbusiness/gotosocial / https://docs.gotosocial.org

> Abstract: Can you explain the whole project and its expected outcome(s). (you have 1200 characters)

GoToSocial (GtS) is an ActivityPub social network server, which provides a lightweight, simple entryway into Fediverse hosting. It is comparable to (but distinct from) projects such as Mastodon, Friendica, and PixelFed.

GtS emphasizes user safety and privacy. Unlike other AP servers, it always requires http signatures, and makes a strong differentiation between 'public' and other kinds of posts. It also makes it very easy for admins to block instances they don't want to interact with, by allowing them to subscribe to block lists or allow lists, and to import blocks, ensuring that users stay safe.

GtS values ease of deployment and maintenance; this means low system requirements, simple configuration, minimal external dependencies, and clear documentation. GtS makes it easy + affordable for self-hosting newcomers to set up a Fediverse server on low- (or even solar-) powered equipment they might have lying around at home.

GtS began development in Feb 2021, and entered Beta in 2024. We hope to use NLnet funding to continue tuning performance and adding features as we work towards a 1.0 release.

> Have you been involved with projects or organisations relevant to this project before? And if so, can you tell us a bit about your contributions? (Optional) This can help us determine if you are the right person to undertake this effort. (max 2500 characters)

I have been working on GoToSocial since the beginning of the project, first independently, paying myself from my savings, and then thanks to two previous grants from NLnet. My colleague Kim has a similar trajectory, as they now work full time on the project, again thanks to NLnet.

Over the last years we have put many thousands of hours of work into the project: writing code and documentation, fixing bugs, communicating with contributors and doing code review, deploying infrastructure for project builds and discussion, doing project planning, and answering user questions.

Aside from our work on GoToSocial, we also maintain a fork of the Activity library (https://codeberg.org/superseriousbusiness/activity), a fork of the standalone Mastodon frontend customized for GoToSocial (https://codeberg.org/superseriousbusiness/masto-fe-standalone, deployed at https://masto-fe.superseriousbusiness.org), and Kim in particular maintains a large amount of libraries used by the project (https://codeberg.org/gruf), particularly go-ffmpreg (https://codeberg.org/gruf/go-ffmpreg).

## Requested Support

> Explain what the requested budget will be used for?
> Does the project have other funding sources, both past and present?
> (If you want, you can in addition attach a budget at the bottom of the form)
> Explain costs for hardware, human labor (including rates used), travel cost to technical meetings, etc. (max 2500 characters, be concise)

GoToSocial has received two NLnet grants previously, one in 2022-2023, for 50k euros, and another one in 2024-2025, for 75k euros, which we shared between the two of us. These grants went towards paying living costs for myself and Kim (rent, groceries, utilities, taxes, etc), while we both worked full time on the project.

For our first grant we underbudgeted our own costs and ended up underpaying ourselves. We better estimated the second grant, but still had some issues doing more work than we budgeted for, taking account of bug fixes, extra features, and release coordination.

With the benefit of experience, this time we intend to budget more sensibly so that we don't end up going into our overdrafts before the end of each milestone, hence we are asking for a larger amount of funding: 100k euros in total, or 50k euros each per year. This amount of money is comparable to the rate for a mid/senior-level developer in the countries we live in, and should allow us to pay our costs without panic.

Happily, we receive a decent amount of money per month via OpenCollective (https://opencollective.com/gotosocial), which allows us to pay for costs (Greenhost.net) for our CI/CD + snapshot distribution server. It also allows us to pay our freelance colleague f0x for the contributions they are able to make when school is not too busy. As such, NLnet money does not need to be used for anything besides mine and Kim's living costs.

This year, we would like to use the NGI Zero Commons grant to fund development of the following efforts (more tbd):

- Add functionality to allow users and admins to configure cleanup of old statuses and accounts from the database, to keep database sizes smaller.
- Implement better threading support when statuses are deleted (ie., store + show status tombstones).
- Improve search performance and add full-text-search (SQLite, Postgres).
- Add additional in-memory caches for frontend object types (statuses, notifications, etc) to reduce database calls and improve response times.
- Add support for subscribing to relays, and allowing GoToSocial itself to act as a relay, improving connectivity with other instances.
- Add additional federation controls (silence/mute/limit instances).
- Add an opt-in profile directory to make it easier to discover accounts on an instance.
- Implement admin panel section to track unreachable instances, so that admins can tell whether another instance has shut down, and take appropriate action.

> Compare your own project with existing or historical efforts
> E.g. what is new, more thorough or otherwise different. (max 4000 characters, be concise)

ActivityPub is now a popular protocol, with a proliferation of AP implementations like Mastodon, Akkoma, Lemmy, Misskey (and its many forks), Pixelfed, WriteFreely, Wordpress, and more, each with its own focus and purpose. As strong as many of these implementations are, however, they are not without their downsides.

For instance, Mastodon and Pixelfed both lean more towards replicating existing corporate social media efforts (Twitter and Instagram respectively), which is offputting for many users. These two implementations are also complex to install and maintain, which puts many would-be admins off hosting their own servers, and leads to a concentration of users on developer-run servers like mastodon.social and pixelfed.social, which breaks the promise of decentralization.

Other ActivityPub-enabled microblogging softwares like Honk and Snac2 have fewer moving parts and lower system requirements, which has led to a surge of deployments of Snac2 in particular. However, their implementation of the Mastodon client API (the de-facto client API of the fediverse, for better or worse), misses some features, and the barebones admin functionality is not user-friendly for people unaccustomed to using the command line.

In addition, some other fediverse projects have a heavy front-end UI which results in a poor experience on low-bandwidth connections or low-powered devices.

GoToSocial's continued aim is to strike a balance between featureful but fairly resource-intensive AP implementations like Mastodon, and lighter and simpler AP implementations like Snac2 and Honk.

Firstly, GtS is small, easy to run, and very well documented. It uses about 200-300MiB of memory on average, so it can be deployed on tiny VPSs for cheap, and on single-board computers. It has no dependencies on things like ffmpeg or on-system SQLite, as everything is virtualized inside the single binary using WASM/Wazero. This means that people who want to run GoToSocial don't need to install or maintain anything else on the host, and it is very easy for third parties like Yunohost to package and distribute.

Secondly, the admin/settings panel offers admins and users the ability to easily customize their profiles and adjust the way that their instance looks, feels, and federates, and to handle day-to-day admin tasks like reviewing reports, blocking or allowing other instances, etc. The web view of profiles is rendered in simple HTML + CSS; Javascript is not required, but is used for progressive enhancement if available. This makes it relatively easy to view a GoToSocial profile on a mobile device with poor connection, on a lower-end laptop etc.

Thirdly, it provides strong safety features thanks to strict block implementation, always-on HTTP signature authentication, interaction controls, and allowing users to choose what visibility of level of posts can be viewed on their profile. The allowlist subscription functionality we added in 2024 is also critical here, as it allows groups of instances to easily federate only with each other, forming "islands" that can be more easily moderated than fully open federation.

Fourthly, we have been -- and continue to be -- fastidious with our Mastodon client API implementation, which means that GoToSocial can be used with a wide variety of clients that provide different experiences to the user. It even works with apps for Pixelfed, so user's can use a GoToSocial account for 'gram-style media-only posting.

> What are significant technical challenges you expect to solve during the project, if any? (optional but recommended, max 5000 characters)

Many of the technical challenges we expect to overcome during this development period are specific to the development efforts we will undertake as part of this grant:

- Status cleanup: Ensure that cleanup processes are capable of being regularly scheduled asynchronously, while not consuming all available server resources (with what will likely involve scanning the entirety of our largest database table!). This may require some clever indexing and/or marking of statuses as expirable in advance.
- Status tombstones: figure out whether we can write a migration to retroactively rethread old statuses that have become unlinked due to deletes.
- Relay support: figure out whether adding support for relay mode will require a separate relay binary, and if so, refactor sections of the codebase into libraries that can be shared by that binary.
- Relay support: make sure GoToSocial can subscribe to both existing relay services, as well as GoToSocial relays (this will probably involve lots of dipping into the codebases of existing relay services like fedi buzz, to figure out what they're doing). And vice versa: ensure that existing fedi implementations with relay support can subscribe to GoToSocial relays.
- Relay support: make sure that DB sizes and memory usage don't become too burdensome given the amount of statuses that relays are likely to process compared to a vanilla instance.
- Profile directory: ensure "discoverable" flag is respected; optimize required new database queries to ensure they use existing indexes (or figure out which new ones need to be created).
- Search support: ensure that the same functionality and performance is offered by both Postgres and SQLite; possibly refactor our database wrappers and migration code for this.
- Unreachable instances: develop a reasonable heuristic to determine whether an instance is unreachable; work out the best way of storing this information in the database and presenting it to admins via the settings panel.

Other technical challenges we will (continue to) address in the near future are not related to any specific milestone task:

- Ensure continued compatibility of GoToSocial with other fedi software projects.
- Ensure continued compatibility of GoToSocial with Mastodon API client apps.
- Refactor old parts of the codebase to increase readability and remove bugs.
- Make performance tweaks to the codebase wherever necessary (reduce CPU usage, improve memory usage).
- Increase coverage of our test suites.
- Continue to support seamless database migration from old versions of GoToSocial to newer ones.
- Refactor some of the frontend settings panel code to maximize code reuse and minimize compiled javascript size, before adding lots of new functionality.
- Move our CI/CD infrastructure and code repository from Github to Codeberg with minimal disruption to our work.

> Describe the ecosystem of the project, and how you will engage with relevant actors and promote the outcomes. E.g. which actors will you involve? Who should run or deploy your solution to make it a success? (max 2500 characters, be concise)

Much of the work we do involves debugging and solving interoperability issues with other federated softwares, which requires keeping communication channels open with the maintainers of those, and figuring out who needs to change what in order for the issue to be resolved. We've done that a lot over the last year or so:

- Fixed interop with Bandwagon: https://github.com/EmissarySocial/bandwagon/issues/152
- Fixed interop with Iceshrimp: https://github.com/superseriousbusiness/gotosocial/issues/1947
- Coordinated interop with Mastodon: https://github.com/superseriousbusiness/gotosocial/pull/3703
- Fixed federation with Gancio: https://github.com/superseriousbusiness/gotosocial/issues/3875
- Alerted Pixelfed of AP serialization issues: https://github.com/pixelfed/pixelfed/issues/5642
- Cajoled Bluesky into adding user-agent headers: https://github.com/bluesky-social/atproto/issues/3504
- Help out Writefreely with http signature request issues: https://github.com/writefreely/writefreely/issues/661#issuecomment-1951367449
- Debug federation with Lemmy along with one of the Lemmy devs: https://github.com/superseriousbusiness/gotosocial/issues/2697

For GoToSocial-specific extensions to ActivityPub, we've also diligently documented what we've done so far, and exposed a GoToSocial namespace so that remote softwares can easily incorporate GtS extensions if they want to: https://docs.gotosocial.org/en/latest/federation/interaction_policy/, https://gotosocial.org/ns.

This is all part and parcel of our goal for GoToSocial to be a "good citizen" in terms of how it federates, and how we fit into the larger ActivityPub microblogging ecosystem. We intend to keep doing this!

## Attachments

> Attachments: add any additional information about the project that may help us to gain more insight into the proposed effort, for instance a more detailed task description, a justification of costs or relevant endorsements. Attachments should only contain background information, please make sure that the proposal without attachments is self-contained and concise. Don't waste too much time on this. Really.

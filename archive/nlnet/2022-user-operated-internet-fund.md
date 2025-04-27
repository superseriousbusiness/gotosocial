# NLnet Grant Application - User Operated Internet Fund

This document is the application on behalf of GoToSocial for funding from the [NLnet](https://nlnet.nl) [User Operated Internet Fund](https://nlnet.nl/useroperated/), June 2022 edition. See [here](https://nlnet.nl/propose/).

## General Project Information

> Project Name

GoToSocial

> Website / wiki

https://codeberg.org/superseriousbusiness/gotosocial / https://docs.gotosocial.org

> Abstract: Can you explain the whole project and its expected outcome(s). (you have 1200 characters)

GoToSocial (GtS) is an ActivityPub social network server, which provides a lightweight and approachable entryway into Fediverse hosting. It is comparable to (but distinct from) projects such as Mastodon, Friendica, and PixelFed.

GtS emphasizes user safety and privacy. Unlike other AP servers, it always requires http signatures, and makes a strong differentiation between 'public' and other kinds of posts. It also makes it very easy for admins to block instances they don't want to interact with, by allowing them to subscribe to block lists and to mass import blocks, ensuring that users stay safe.

GtS values ease of deployment and maintenance; this means low system requirements, simple configuration, minimal external dependencies, and clear documentation. GtS makes it easy + affordable for self-hosting newcomers to set up a Fediverse server on low- (or even solar-) powered equipment they might have lying around at home.

GtS began development in Feb 2021. It is still in Alpha, and we hope to use NLNet funding to bring it up to the Beta phase. The project roadmap (https://codeberg.org/superseriousbusiness/gotosocial/src/branch/main/ROADMAP.md) gives more information on what we have planned.

> Have you been involved with projects or organisations relevant to this project before? And if so, can you tell us a bit about your contributions? (Optional) This can help us determine if you are the right person to undertake this effort.

I am the founder of the GoToSocial project.

Before starting work on GoToSocial, I worked at a corporate IT company as a data engineer, but I became disillusioned quickly with the company's lack of open source contributions, and its reliance on AWS and other exploitative corporate technologies. Around the same time, I started looking into open source social networking software like Mastodon, and span up a Mastodon server. Shortly after, I kicked off GoToSocial as a way to address some of the issues I saw with Mastodon, and to learn more about ActivityPub. At the end of last year, I left my corporate job to give more attention to GtS.

Over the last year or so I have put (something like) 1,000+ hours of work into the project. This involved many, many hours of writing code and documentation, communicating with other contributors and project maintainers, deploying infrastructure for project builds and discussion, project planning, and answering user questions.

Aside from GoToSocial, I've also made small PRs upstream to the ActivityPub library that we use -- https://github.com/go-fed/activity -- and a small fix to Pinafore, one of the clients that we recommend for GoToSocial -- https://github.com/nolanlawson/pinafore. I'm also in active communication with the developers of Tusky, another Mastodon client that works with GoToSocial as well, and will likely make PRs to them to implement GtS-specific features.

## Requested Support

> Requested Amount (between 5,000 and 50,000 euro)

42,950

> Explain what the requested budget will be used for?
> Does the project have other funding sources, both past and present?
> (If you want, you can in addition attach a budget at the bottom of the form)
> Explain costs for hardware, human labor (including rates used), travel cost to technical meetings, etc.

Currently, GoToSocial receives about €22/week from LiberaPay donations - https://liberapay.com/gotosocial. I have been paying my own costs for working on the project from my savings, which is unfortunately not sustainable for a lot longer.

The requested NLNet budget will be used to fund the remaining Alpha portion of development, and bring GoToSocial into the Beta phase (see the roadmap - https://codeberg.org/superseriousbusiness/gotosocial/src/branch/main/ROADMAP.md). In practical terms, this means paying myself to work full time on the project for one year, and paying for contributions from other developers as well.

To pay my living costs + rent I need to make about €2,000/month after tax, working full time. In Belgium, that equates to about €3,000/month, which is €36,000 for one year of work. Naively calculated at 40 hours / week, that's €18.75 per hour.

The remaining budget will be divided as follows:

€3750 - Pay contributors €18.75/hour for code submitted to the project towards features, bugfixes, and security patches. It's difficult to foresee in advance how much code we will receive in this way, and how many people will want to just donate code rather than asking for payment, so this number may need to be adjusted up or down. €3750 allows us to pay for 200 hours of contributed work over the coming year.

€2000 - Pay our longtime contributor f0x €25/hour for an estimated 80 hours of upcoming frontend development work.

€2100 - Pay our admin/fundraiser Maloki €175/month for one year, for part time contributions to GoToSocial admin, and work on the OpenCollective page.

€1200 - One year of hosting costs for https://gts.superseriousbusiness.org, the main GoToSocial testing instance, which is also used for CI/CD and building/pushing Docker containers to Docker hub. This is currently hosted on a small Greenhost.net instance, but ideally we could use some of this budget to take a larger instance (to speed up testing and builds) with more disk space on it.

We would like to put the NLNet grant money into the GoToSocial OpenCollective fund, so that we can make our incomings + outgoings as transparent as possible, and use the Open Collective Europe fiscal host to make payouts. See https://opencollective.com/gotosocial.

> Compare your own project with existing or historical efforts. (e.g. what is new, more thorough, or otherwise different)

ActivityPub has become a popular protocol, with a proliferation of AP implementations like Mastodon, Pleroma, Lemmy, Pixelfed, WriteFreely, and many more, each with its own focus and purpose. In particular, the Mastodon API has become a de-facto standard for fediverse client APIs (GoToSocial implements the Mastodon API!). As strong as many of these implementations are, however, they are not without their downsides.

For instance, Mastodon and Pixelfed both lean more towards replicating existing corporate social media efforts (Twitter and Instagram respectively), which is offputting for many users. These two implementations are also complex and resource intensive to install and maintain, which puts many would-be admins off hosting their own servers, and leads to a concentration of users on developer-run servers like mastodon.social and pixelfed.social, which breaks the promise of decentralization.

While Pleroma offers the benefit of small-footprint installs like GoToSocial, many users who would like to switch to something other than Mastodon are deterred by its lack of security features, its project governance, and its reputation as an engine for disseminating hate speech (it is not uncommon to see people describe Pleroma instances as being block-on-sight, since they feel unsafe interacting with those who use the software).

Finally, most of the AP implementations mentioned above have opinionated web applications attached to them, which tend to funnel users into interacting with the server implementation in specific ways, and make it difficult for users and admins to customize the look and feel of their profiles and posts.

GoToSocial is specifically designed to mitigate the above concerns.

Firstly, it's small, easy to run, and well documented.

Secondly, it provides strong safety features thanks to strict block implementation, and the alternative federation modes we have planned for the future.

Thirdly, we want to make GtS as customizable as possible by allowing admins to easily modify html templates and css. Eventually, in Beta, we'd like to allow users to select templates for their own profiles and posts, and modify their own css, in a hark back to the days of Web 1.0 and early Web 2.0. We aim to populate the fediverse with many small, weird, specialized instances, each of which looks and feels different.

> What are significant technical challenges you expect to solve during the project, if any? (optional but recommended)

The main technical challenges we foresee on the project are:

1. Ensuring compatibility with other AP servers (see here: https://codeberg.org/superseriousbusiness/gotosocial/projects/4).
2. Ensuring compatibility with clients that use the Mastodon API (see here: https://codeberg.org/superseriousbusiness/gotosocial/projects/5).
3. Designing nuanced federation safety features that allow instance admins to screen federation without totally breaking it. This will require careful design discussions and lots of testing.
4. Implementing our own open-source http signature library with a reference implementation of the latest draft of the http signature proposal: https://httpwg.org/http-extensions/draft-ietf-httpbis-message-signatures.html.
5. Writing + maintaining our own extensions to the AP protocol (see below).

> Describe the ecosystem of the project, and how you will engage with relevant actors and promote the outcomes?

One of the biggest challenges for an ActivityPub project is ensuring that federation works smoothly with other servers. While it's easy for GoToSocial to federate with other GtS instances, other softwares interpret the protocol differently. This means that we will have to spend a lot of time reading through code for Mastodon, Pixelfed, etc, and trying to reach a consensus with the developers of those projects regarding where fixes/workarounds ought to be implemented, and by whom.

Additionally, many of the federation issues we've seen so far have been due to GtS' strict implementation of http signatures, so much of the work we'll have to do in communication with other projects will also involve advocating for http signatures as an authentication method.

If possible, we would like to be able to join a discussion space for these issues with other AP devs, although it's currently unclear whether such a space exists, or would have to be formed.

Finally, since the go-fed library project has become unmaintained, much of our effort will go towards maintaining the GoToSocial fork of the library, and carefully implementing and documenting our own extensions to it. Ideally, we would also host our own AP namespace comparable to the toot: Mastodon JSON-LD namespace, to make it easier for future developers to integrate GoToSocial extensions into their own projects.

## Attachments

> Attachments: add any additional information about the project that may help us to gain more insight into the proposed effort, for instance a more detailed task description, a justification of costs or relevant endorsements. Attachments should only contain background information, please make sure that the proposal without attachments is self-contained and concise. Don't waste too much time on this. Really.

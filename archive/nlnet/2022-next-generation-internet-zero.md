# NLnet Grant Application - NGI Zero Entrust

This document is the application on behalf of GoToSocial for funding from the [NLnet](https://nlnet.nl) [NGI Zero Entrust](https://nlnet.nl/entrust/), August 2022 edition. See [here](https://nlnet.nl/propose/).

## General Project Information

> Project Name

GoToSocial

> Website / wiki

https://codeberg.org/superseriousbusiness/gotosocial / https://docs.gotosocial.org

> Abstract: Can you explain the whole project and its expected outcome(s). (you have 1200 characters)

GoToSocial (GtS) is an ActivityPub (AP) social network project. It complements existing AP implementations by providing a lightweight, customizable entryway into decentralized social media hosting. Though GtS is still in Alpha, there are already 100+ servers up and running (https://gotosocial.fediverse.observer/stats).

GtS values ease of deployment/maintenance; this means low system requirements, minimal external dependencies, and clear documentation. GtS empowers self-hosting newcomers to deploy small, personalized instances, from which they connect to others across the Fediverse, using low- (or even solar-) powered equipment lying around at home.

GtS protects user data: it always requires http signatures, it strongly differentiates between 'public' and other kinds of posts, and it allows server admins to subscribe to instance block lists and to mass import blocks, ensuring that users feel confident about where their data ends up.

NLnet/NGI0 funding would be used to bring GtS to the end of the Alpha phase (feature parity with other AP implementations), at which point we can implement the AP `Move` activity, ensuring data/identity portability for users across AP implementations.

> Have you been involved with projects or organisations relevant to this project before? And if so, can you tell us a bit about your contributions? (Optional) This can help us determine if you are the right person to undertake this effort.

I am the founder of the GoToSocial project.

Before starting work on GoToSocial, I worked at a for-profit IT company as a data engineer, but I became disillusioned quickly with the company's lack of open source contributions, and its reliance on AWS and other exploitative corporate technologies. Around the same time, I began looking into open source social networking software like Mastodon, and span up a Mastodon server. Shortly after, I kicked off GoToSocial as a way to address some of the issues I saw with Mastodon, and to learn more about ActivityPub. At the end of last year, I left my corporate job to give more attention to GtS.

Over the last year or so I have put (something like) 1,000+ hours of work into the project. This involved many, many hours of writing code and documentation, communicating with other contributors and project maintainers, deploying infrastructure for project builds and discussion, project planning, and answering user questions.

Aside from GoToSocial, I've also made PRs upstream to the ActivityPub library that we use (https://github.com/go-fed/activity), a small fix to Pinafore, one of the clients that we recommend for GoToSocial (https://github.com/nolanlawson/pinafore, and I have also committed code to OwnCast, a self-hosted livestreaming service which uses the ActivityPub protocol (https://owncast.online/). I'm also in active communication with the developers of Tusky, another Mastodon client that works with GoToSocial as well, and will likely make PRs to them to implement GtS-specific features.

I have a keen interest in user privacy and decentralized networking.

## Requested Support

> Requested Amount (between 5,000 and 50,000 euro)

43850

> Explain what the requested budget will be used for?
> Does the project have other funding sources, both past and present?
> (If you want, you can in addition attach a budget at the bottom of the form)
> Explain costs for hardware, human labor (including rates used), travel cost to technical meetings, etc.

Currently, GoToSocial receives about €34/week from LiberaPay donations (https://liberapay.com/gotosocial) and about €100/week from OpenCollective (https://opencollective.com/gotosocial). That aside, I have been paying my own costs for working on the project from my savings, which is unfortunately not sustainable for a lot longer.

The requested NLnet budget will be used to fund the remaining Alpha portion of development (please see the attached ROADMAP.md document) -- in practical terms, this means paying myself to work full time on the project for one year, and paying for contributions from other developers as well.

We see the end of Alpha development, and the beginning of Beta, as an important milestone, since that is the point where GoToSocial can be considered as having feature parity with other AP implementations, and therefore being 'ready' for daily use. Once GtS reaches that point, we will implement the `Move` ActivityPub activity, which will allow users to move their data and identity from other servers onto GoToSocial, or indeed from GoToSocial to other servers.

Detailed breakdown:

To pay my living costs + rent I need to make about €2,000/month after tax, working full time. In Belgium, that equates to about €3,000/month, which is €36,000 for one year of work. Naively calculated at 40 hours / week, that's €18.75 per hour.

The remaining budget will be divided as follows:

€3750 - Pay contributors €18.75/hour for code submitted to the project towards features, bugfixes, and security patches. It's difficult to foresee in advance how much code we will receive in this way, and how many people will want to just donate code rather than asking for payment, so this number may need to be adjusted up or down. €3750 allows us to pay for 200 hours of contributed work over the coming year.

€2000 - Pay our longtime contributor f0x €25/hour for an estimated 80 hours of upcoming frontend development work.

€2100 - Pay our admin/fundraiser Maloki €175/month for one year, for part time contributions to GoToSocial admin, and work on the OpenCollective page.

We would like to put the NLnet grant money into the GoToSocial OpenCollective fund, so that we can make our incomings + outgoings as transparent as possible, and use the Open Collective Europe fiscal host to make payouts. See https://opencollective.com/gotosocial.

> Compare your own project with existing or historical efforts. (e.g. what is new, more thorough, or otherwise different)

ActivityPub has become a popular social protocol, with a proliferation of AP implementations like Mastodon, Pleroma, Lemmy, Pixelfed, WriteFreely, and many more, each with its own focus and purpose. In particular, the Mastodon API has become a de-facto standard for fediverse client APIs (GoToSocial implements the Mastodon API!). As strong as many of these implementations are, however, they are not without their downsides.

For instance, Mastodon and Pixelfed both lean more towards replicating existing corporate social media efforts (Twitter and Instagram respectively), which is offputting for many users. These two implementations are also complex and resource intensive to install and maintain, which puts many would-be admins off hosting their own servers, and leads to a concentration of users on developer-run servers like mastodon.social and pixelfed.social, which breaks the promise of decentralization by exposing a single point of failure.

While Pleroma offers the benefit of small-footprint installs like GoToSocial, many users who would like to switch to something other than Mastodon are deterred by its lack of security features, its project governance, and its reputation as an engine for disseminating hate speech (it is not uncommon to see people describe Pleroma instances as being block-on-sight, since they feel unsafe interacting with those who use the software).

Finally, most of the AP implementations mentioned above have opinionated web applications attached to them, which tend to funnel users into interacting with the server implementation in specific ways, and make it difficult for users and admins to customize the look and feel of their profiles and posts.

GoToSocial is specifically designed to mitigate the above concerns.

Firstly, it's small, light, and well documented. It has been tested (and works well) on Raspberry Pi, which opens the door to potentially running it on solar powered servers, though this needs more investigation. Low requirements and simple set up means more instances with fewer users, rather than fewer instances with more users. This ensures a more resilient and decentralized social web.

Secondly, it provides strong safety features thanks to strict block implementation, blocklists, and the alternative federation modes we have planned for the future. Users should not have to worry about private messages being exposed, nor should they have to worry about trolls and harassment, or painstakingly blocking bad actor after bad actor. Blocklists allow the community to pool their effort to ensure user safety.

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

## Follow-up Questions

> You mention the difficulties of interoperating with various server and client softwares, and highlight that this also requires contributing to other projects’ implementations (e.g. of HTTP signatures). How much of your work do you estimate to be specific to developing GoToSocial, and how much to contributions to the broader ActivityPub ecosystem?

I think the ratio so far has been about 90% GoToSocial, 10% looking through other projects, contributing code, and talking to other project developers. I think if we were to take over maintenance of go-fed, or fork the library (more on this below) this would probably change to something more like 70% GoToSocial, 30% broader ActivityPub contributions.

> With the go-fed library you depend on being unmaintained (as you say), why do you plan to fork it rather than contributing to its maintenance? Have you discussed with other users like Owncast? Could you reflect on go-ap, which seems to be in a healthier state? Would it be possible to merge the functionality of these two, given some additional budget?

When I started sketching out ideas for GoToSocial, one of the first steps was to look around for Go libraries that implemented ActivityPub, in order to avoid reinventing the wheel. At that time (~ January 2021), go-fed was the best option: well-documented, easily extensible, well-maintained, with an active Matrix channel, and contributed to (and used) by lots of people. By contrast, go-ap came across as more of a work-in-progress: it wasn't as well documented, didn't have any releases or tags (which is still the case), and didn't have as many contributors, nor was it used by as many people as go-fed.

In the year and a half since then, this situation has changed somewhat. go-fed is no longer being actively maintained by its original developer (CJ), and the Matrix channel is no longer active either. By contrast, go-ap is still being actively developed, judging by the commit history on Github (https://github.com/go-ap/activitypub/commits/master) -- not to mention the fact that the Gitea project has chosen to use go-ap for its federation extensions.

Nevertheless, even if GoToSocial was just started up today, I believe we would still pick go-fed over go-ap, mostly for the same reasons that motivated our initial choice: it's a more mature library with wider adoption that has been used to federate in production by several different successful projects, it has proper versioned releases and great docs, and by now it's had lots and lots of eyes on the code.

Now that I have more experience with go-fed--and the ActivityPub ecosystem in general--I can also better understand why go-fed code is generated based on jsonld documents, which is something that confused me a bit at first: it makes writing custom extensions relatively easy, and you essentially get the proper jsonld schema document for your namespace for free, because you need to write it anyway in order to generate your Go code (see https://github.com/go-fed/activity/tree/master/astool). This is very useful for compatibility with ActivityPub implementations that require the jsonld namespace to be properly dereferenceable, like Friendica; if GoToSocial wants to add a gts namespace for custom extensions (more on this below), we can then host the jsonld document at the appropriate URL so that it can dereferenced by other AP implementations.

That said, go-fed is not without its downsides, and go-ap has its advantages too. The main reason Gitea picked go-ap over go-fed is because of the extra overhead that go-fed adds to compiled binary sizes (something like 40MB if I remember rightly). This is less of a problem in go-ap, which doesn't add a lot of size to binaries, and can be considered a much lighter library in this regard. This is well explained here by Marius himself: https://lists.sr.ht/~mariusor/activitypub-go/%3Ca4bb4473-1379-453a-93aa-0c6aa8af49e4@dachary.org%3E#%3C20220512072502.mltep3wqjk2v5ste@slate%3E

As for the reasoning behind forking go-fed rather than contributing to its maintenance, the short answer is that I haven't been able to get in touch with CJ for a long time, and the last time I submitted a PR to go-fed, it took a long time to get it reviewed + merged. I got the idea that CJ was unfortunately rather burned out on the project, so I didn't want to keep badgering him about it. Ultimately it doesn't really make much difference whether we fork or contribute to the existing project repo instead; it's more of a logistical issue than an ideological one. If we can reach CJ, and he's conducive to it, we'd happily work on the existing repo instead with his guidance!

Re: Owncast -- Forking/maintaining go-fed is not something we've discussed with Owncast yet, though we are on good terms with Gabe. If we were to take over maintenance of go-fed, or indeed decide to proceed forwards with the fork instead, we would absolutely want to involve Owncast and other users like WriteFreely as contributors, to make sure that any changes we made wouldn't break their use-cases, and to incorporate any ideas they have about easing some of the existing pain points with go-fed. The goal is not to make go-fed hyper-specific to GoToSocial, but to keep it as a community library, in a healthy state, usable by the widest number of projects.

Re: merging the functionality of go-fed and go-ap, given additional budget -- I believe this would take some more detailed investigation about exactly what the differences are between go-fed and go-ap in terms of code and ActivityPub API coverage, and what people find beneficial about each library compared to the other. My understanding is that while go-fed implements almost all of the ActivityPub vocabulary, go-ap focuses on a smaller subset of the vocabulary. I'm not sure at this point if additional budget for GoToSocial would be well spent trying to pull the two libraries together in one package, but I do think it might make for another interesting NLnet application in future.

> Related to that, how much of the implementation work will be directly to go-fed (or its fork) itself, and (how) might this also help other activitypub-based server software?

This is tricky to foresee, beyond the basic 70/30 estimate given in the answer above.

In the short term, we would like to polish, gently extend, and fix bugs in the existing codebase, using the existing code generation paradigm. We also foresee some possible performance improvements to the library, which are relatively easy to implement (things like not parsing URLs back and forth quite so much). This sort of low-key, incremental stewardship of go-fed would already provide benefit to developers using the library.

In the long run, some of the wishlist changes I've discussed with Kim (one of the other GoToSocial developers) would involve modifying the underlying code generation methods of go-fed, and using Go 1.18 generics in place of generating lots of structs; this might necessitate changes to some of the library's API, and would likely eventually lead to something more like a v2 of go-fed. This would provide a lot of benefit for AP developers using Go, since it would probably fix some of the binary size issues, and make understanding + maintaining go-fed code generation much easier. This amount of work, however, would probably require a full-time developer to work on go-fed for at least six months or so. This is beyond the scope of this NLnet funding application, but it would make for a very exciting application further down the line if Kim (or someone else) has time and energy for it!

Given the scope of this current application, it's probably best to stick with the low-hanging fruit and make gentle incremental tweaks and improvements to go-fed either on the main repo or a fork, but we would love to see the library reworked in the future. Unfortunately, there are only so many hours in the day!

> Regarding “Writing + maintaining our own extensions to the AP protocol”: what extensions would GoToSocial require? And how would this impact the interoperability with other software that has not implemented these extensions?

The extensions required by GoToSocial primarily involve things like adding flags to ActivityPub Objects to indicate whether a post can be boosted, replied-to, or liked. Currently, setting such flags on posts is supported internally in GoToSocial, and it is enforced by accepting or refusing Activities coming in via the GtS federation API. However, this does not solve the problem of how to federate these preferences out to remote servers.

In an issue on the Mastodon repository here -- https://github.com/mastodon/mastodon/issues/14762 -- Claire and trwnh propose a few different possibilities for how to handle no-reply posts over ActivityPub, using flags added to the posts (or the owner's profile) and federating those to other instances. This requires extending the AP spec to support these new flags. If GoToSocial were to implement this, in a GtS AP namespace, it would require us to extend the go-fed library to support + document the new types, and to host the schema for the namespace at something like eg https://gotosocial.org/ns#, similar to Mastodon's toot namespace at http://joinmastodon.org/ns# (unfortunately not online anymore), and the as namespace at https://www.w3.org/ns/activitystreams.

Alternatively, it could be the case that Mastodon implements the extensions first in their toot namespace, in which case we will extend go-fed to implement these new extensions in the existing toot jsonld document.

From an interoperability perspective, we would work to ensure that any extensions to AP on our part represent progressive enhancement rather than breaking changes that prevent different servers from talking with each other. That is, if a GoToSocial server with GtS specific extensions communicates with another server without the extensions, both servers should be able to process the interaction. To take the example above of non-replyable posts, if a GoToSocial server with these extensions sends a post to a server without these extensions, the GtS server will fall back to the existing best-effort behavior for preventing post replies, which is non-breaking. Conversely, if a GoToSocial server receives a post with extensions it cannot parse, it will simply read the message to the best of its ability using the basic as namespace, without dropping the message.

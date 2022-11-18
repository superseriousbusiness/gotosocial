# Posts

## Privacy Settings

GoToSocial offers Mastodon-style privacy settings for posts. In order from most to least private, these are:

* Direct
* Mutuals-only
* Private/Followers-only
* Unlisted
* Public

Whatever privacy setting you choose for a post, GoToSocial will do the best it can to ensure that your posts don't appear to users on instances that you've blocked, or to users that you've blocked directly.

Unlike with some other fediverse server implementations, GoToSocial uses a default post setting of `unlisted` rather than `public` for new accounts. Our philosophy here is that posting something public should always be a conscious decision rather than a default.

Please note that while GoToSocial respects these privacy settings very strictly, other server implementations cannot necessarily be trusted to do so: there are bad actors on the fediverse. As with any social media, you should think carefully about what you post and to whom.

### Direct

Posts with a visibility of `direct` will only appear to the post author, and to users who are mentioned in the post. Take the following post for example:

```text
Hey @whoever@example.org, this is a private/direct post! Only we can see this!
```

If this message was written by `@someone@server.com` then only `@whoever@example.org` and `@someone@server.com` would be able to see it.

As the name implies, `direct` posts are best used when you want to communicate directly with one or more people.

However, `direct` posts are **not** a suitable replacement for end-to-end encrypted messaging offered by things like [Signal](https://signal.org/) and [Matrix](https://matrix.org/). If you want to communicate directly, but you're not communicating sensitive information, then direct posts are fine. If you need to have a sensitive + secure conversation, use something else!

Direct posts can be liked/faved, but they cannot be boosted.

Direct posts are **not** accessible via a web URL on your GoToSocial instance.

### Mutuals-only

Posts with a visibility of `mutuals_only` will only appear to the post author, and to *mutual follows* of the post author. In other words, they can only be seen by others if two conditions are met:

1. The other account follows the post author.
2. The post author follows the other account back.

This is useful for when you want to post something that you only want friends to see.

Mutuals-only posts can be liked/faved, but they cannot be boosted.

Mutuals-only posts are **not** accessible via a web URL on your GoToSocial instance.

### Private/Followers-only

Posts with a visibility of `private` will only be visible to the post author, and to people who follow the post author. This is similar to `mutuals_only`, but only the first condition needs to met; the post author doesn't need to follow the other account back.

This is useful for when you want to make announcements to people who follow you, or share something slightly less private than `mutuals_only`.

Private/followers-only posts can be liked/faved, but they cannot be boosted.

Private/followers-only posts are **not** accessible via a web URL on your GoToSocial instance.

### Unlisted

Posts with a visibility of `unlisted` (sometimes called `unlocked` posts) are semi-public. They will be sent to anyone who follows you, and they can be boosted into the timelines of people who don't follow you, but they won't appear on Federated or Local timelines, and they won't appear on your public profile.

Unlisted posts are useful when you want to allow a post to spread, but you don't want it to be immediately visible to everyone. They are also useful when you want to make public-ish posts, but without clogging up Federated/Local timelines.

Unlisted posts can be liked/faved, and they can be boosted.

Unlike with Mastodon, unlisted posts are **not** accessible via a web URL on your GoToSocial instance!

### Public

Posts with a visibility of `public` are *fully* public. That is, they can be seen via the web, and they will appear in Local and Federated timelines, and they are fully boostable. `public` is the ultimate 'let my post be seen everywhere' setting, for when you want something to be widely available and easy to distribute.

Public posts can be liked/faved, and they can be boosted.

**Public posts are accessible via a web URL on your GoToSocial instance!**

## Extra Flags

GoToSocial offers four extra flags on posts, which can be used to tweak how your post can be interacted with by others. These are:

* `federated`
* `boostable`
* `replyable`
* `likeable`

By default, all these flags are set to `true`.

Please note that while GoToSocial strictly respects these settings, other fediverse server implementations might not be aware of them. A consequence of this is that users on non-GoToSocial servers might think they are replying/boosting/liking your post, and their instance might behave as though that behavior was allowed, but those interactions will be denied by your GoToSocial server and you won't see them.

### Federated

When set to `false`, this post will not be federated out to other fediverse servers, and will be viewable only to accounts on your GoToSocial instance. This is sometimes called 'local-only' posting.

### Boostable

When set to `false`, your post will not be boostable, even if it is unlisted or public. GoToSocial enforces this by refusing dereferencing requests from remote servers in the event that someone tries to boost the post.

### Replyable

When set to `false`, replies to your post will not be accepted by your GoToSocial server, and will not appear in your timeline or create notifications. GoToSocial enforces this by giving an error message to attempted replies to the post from federated servers.

### Likeable

When set to `false`, likes/faves of your post will not be accepted by your GoToSocial server, and will not create notifications. GoToSocial enforces this by giving an error message to attempted likes/faves on the post from federated servers.

## Input Types

GoToSocial currently accepts two different types of input for posts (and user bio). These are:

* `plain`
* `markdown`

Plain is the default method of posting: GtS accepts some plain looking text, and converts it into some nice HTML by parsing links and mentions etc. If you're used to Mastodon or Twitter or most other social media platforms, this way of writing posts will be immediately familiar.

Markdown is a more complex way of organizing text, which gives you more control over how your text is parsed and formatted.

GoToSocial supports the [Basic Markdown Syntax](https://www.markdownguide.org/basic-syntax), and some of the [Extended Markdown Syntax](https://www.markdownguide.org/extended-syntax/) as well, including fenced code blocks, footnotes, strikethrough, subscript, superscript, and automated URL linking.

You can also include snippets of basic HTML in your markdown!

For more information on Markdown, see [The Markdown Guide](https://www.markdownguide.org/).

For a quick reference on Markdown syntax, see the [Markdown Cheat Sheet](https://www.markdownguide.org/cheat-sheet).

## Formatting

When a post is submitted in `plain` format, GoToSocial automatically does some tidying up and formatting of the post in order to convert it to HTML, as described below.

### Whitespace

Any leading or trailing whitespaces and newlines are removed from the post. So for example:

```text


this post starts with some newlines
```

will become:

```text
this post starts with some newlines
```

### Wrapping

The whole post will be wrapped in `<p></p>`.

So the following text:

```text
Hi here's a little post!
```

Will become:

```html
<p>Hi here's a little post!</p>
```

### Linebreaks

Any newlines will be replaced with `<br />`

So to continue the above example:

```text
Hi here's a little post!

And here's another line.
```

Will become:

```html
<p>Hi here's a little post!<br /><br />And here's another line</p>
```

### Links

Any recognizable links in the text will be shortened and turned into proper hyperlinks, and have some additional attributes added to them.

For example:

```text
Here's a link to something: https://example.org/some/link/address
```

will become:

```html
Here's a link to something: <a href="https://example.org/some/link/address" rel="nofollow" rel="noreferrer" rel="noopener">example.org/some/link/address</a>
```

which will be rendered as:

> Here's a link to something: [example.org/some/link/address](https://example.org/some/link/address)

Note that this will only work for `http` and `https` links; other schemes are not supported.

### Mentions

You can 'mention' another account by referring to the account in the following way:

> @some_account@example.org

In this example, `some_account` is the username of the account you want to mention, and `example.org` is the domain that hosts their account.

The mentioned account will get a notification that you've mentioned them, and be able to see the post in which they were mentioned.

Mentions are formatted in a similar way to links, so:

```text
hi @some_account@example.org how's it going?
```

will become:

```html
hi <span class="h-card"><a href="https://example.org/@some_account" class="u-url mention">@<span>some_account</span></a></span> how's it going?
```

which will be rendered as:

> hi <span class="h-card"><a href="https://example.org/@some_account" class="u-url mention">@<span>some_account</span></a></span> how's it going?

When mentioning local accounts (ie., accounts on your instance), the second part of the mention is not necessary. If there's an account with username `local_account_person` on your instance, you can mention them just by writing:

```text
hey @local_account_person you're my neighbour
```

This will become:

```html
hey <span class="h-card"><a href="https://my.instance.org/@local_account_person" class="u-url mention">@<span>local_account_person</span></a></span> you're my neighbour
```

which will be rendered as:

> hey <span class="h-card"><a href="https://my.instance.org/@local_account_person" class="u-url mention">@<span>local_account_person</span></a></span> you're my neighbour

## Input Sanitization

In order not to spread scripts, vulnerabilities, and glitchy HTML all over the place, GoToSocial performs the following types of input sanitization:

`plain` input type:

* Before parsing, any existing HTML is completely removed from the post body and content-warning fields.
* After parsing, all generated HTML is run through a sanitizer to remove harmful elements.

`markdown` input type:

* Before parsing, any existing HTML is completely removed from the content-warning field.
* Before parsing, any existing HTML in the post body is run through a sanitizer to remove harmful elements.
* After parsing, all generated HTML is run through a sanitizer to remove harmful elements.

GoToSocial uses [bluemonday](https://github.com/microcosm-cc/bluemonday) for HTML sanitization.

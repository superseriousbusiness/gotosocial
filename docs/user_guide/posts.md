# Posts

## Input Types

GoToSocial currently accepts two different types of input for posts. These are:

* `plain`
* `markdown`

Plain is the default method of posting: GtS accepts some plain looking text, and converts it into some nice HTML by parsing links and mentions etc. If you're used to Mastodon or Twitter or most other social media platforms, this way of writing posts will be immediately familiar.

Markdown is a more complex way of organizing text, which gives you more control over how your text is parsed and formatted.

For more information on markdown, see [The Markdown Guide](https://www.markdownguide.org/).

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

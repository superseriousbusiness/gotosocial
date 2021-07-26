# Writing Posts

TODO

## Formatting

This section describes the different post input types accepted by GoToSocial, and the method GtS uses to parse text into HTML.

### Links

Any recognized links in the text will be shortened and turned into proper hyperlinks. For example:

> Here's a link to something: https://example.org/some/link/address

will become:

> Here's a link to something: [example.org/some/link/address](https://example.org/some/link/address)

### Mentions

You can 'mention' another account by referring to the account in the following way:

> @some_account@example.org

In this example, `some_account` is the username of the account you want to mention, and `example.org` is the domain that hosts their account.

The mentioned account will get a notification that you've mentioned them, and be able to see the post in which they were mentioned.

Mentions are formatted in a similar way to links, so:

> @some_account@example.org

will become:

> <span class="h-card"><a href="https://example.org/@some_account" class="u-url mention">@<span>some_account</span></a></span>

## Input Types

GoToSocial currently accepts two different types of input. These are:

* `plain`
* `markdown`

Plain is the default method of posting: GtS accepts some plain looking text, and converts it into some nice HTML by parsing links and mentions etc.

Markdown is a more complex way of organizing text, which gives you more control over how your text is parsed and formatted.

For more information on markdown, see [The Markdown Guide](https://www.markdownguide.org/).

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

!!! warning
    Mutuals-only posts are not currently functional.

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

## Input Types

GoToSocial currently accepts two different types of input for posts (and user bio). The [user settings page](./settings.md) allows you to select between them. These are:

* `plain`
* `markdown`

Plain is the default method of posting: GtS accepts some plain looking text, and converts it into some nice HTML by parsing links and mentions etc. If you're used to Mastodon or Twitter or most other social media platforms, this way of writing posts will be immediately familiar.

Markdown is a more complex way of organizing text, which gives you more control over how your text is parsed and formatted.

GoToSocial supports the [Basic Markdown Syntax](https://www.markdownguide.org/basic-syntax), and some of the [Extended Markdown Syntax](https://www.markdownguide.org/extended-syntax/) as well, including fenced code blocks, footnotes, strikethrough, subscript, superscript, and automated URL linking.

You can also include snippets of basic HTML in your markdown!

For more information on Markdown, see [The Markdown Guide](https://www.markdownguide.org/).

For a quick reference on Markdown syntax, see the [Markdown Cheat Sheet](https://www.markdownguide.org/cheat-sheet).

## Media Attachments

GoToSocial allows you to attach media files to your posts, which most clients will then render in a gallery view at the bottom of your post. By default, you can attach 6 pieces of media to a post, but this may vary depending on the client that you're using, and the configuration of your instance.

The following file types are (currently) supported:

- image/jpeg
- image/gif
- image/png
- image/webp
- video/mp4 (most types)

By default, the size limit of uploaded media is 40MB, but again this may vary depending on your instance configuration.

### Image Descriptions (alt text)

When you attach a piece of media to a post, like an image or a video, most clients will give you the option to provide a description of what the image or video depicts. This description will be provided as alt text for all users viewing the media. This is useful for everyone, but especially for blind or partially-sighted folks. Without an image description, it may be unclear what is contained in a piece of media, and why it was attached to a given post.

Writing a good image description can be difficult, but it is worthwhile!

> image descriptions are a gesture of care and an essential part of accessibility. Without them, content would be completely unavailable to Blind/low vision folks. By writing image descriptions, we show support of cross-disability solidarity and cross-movement solidarity.

-- Alex Chen, [How to write an image description](https://uxdesign.cc/how-to-write-an-image-description-2f30d3bf5546).

### Exif Data

When a photo or video is taken, most traditional cameras and phone cameras encode [Exif data tags](https://en.wikipedia.org/wiki/Exif) into the resulting media as metadata. This Exif data contains things like:

- Make and model of the camera.
- Color and pixel information for the image or video.
- Dimensions and orientation of the image or video.
- Data and time information.
- Location information (if enabled).

Traditionally, these Exif data points are used by photographers to help them catalogue their own images. Unfortunately, though, they also have [privacy and security implications](https://en.wikipedia.org/wiki/Exif#Privacy_and_security), especially where location data is concerned. If you've ever posted an image online to a platform like Facebook, you may have wondered how Facebook knows where and when the image was taken; this is largely thanks to the location information and timestamp embedded in the Exif data, which Facebook reads from the image in order to assemble a timeline of "places you've been".

To avoid leaking information about your location, GoToSocial makes a best-effort attempt to remove Exif information from media when you upload it, by zeroing out Exif data points.

!!! danger
    For your convenience and privacy, GoToSocial currently removes Exif tags from image files when they are uploaded. However, **automated removal of Exif data from mp4 videos is not currently supported** (see [#2577](https://codeberg.org/superseriousbusiness/gotosocial/issues/2577)).
    
    Before you upload a video to GoToSocial, we recommend ensuring that Exif data tags are already removed from the video. You can find various tools and services online for doing this.
    
    To prevent Exif location data being encoded into an image or video in the first place, you can also turn off location tagging (often called geotagging) in the camera app of your device.

!!! tip
    Even if you fully remove all Exif metadata from an image or video before uploading it, there are many ways that malicious jerks can infer your location anyway based on the contents of the media itself.
    
    If you are part of an organization that has an operational requirement for secrecy, or if you are being stalked or surveilled, you may want to consider not posting any media that could contain clues as to your whereabouts.

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

### Hashtags

You can use one or more hashtags in your post to indicate subject matter, and to allow the post to be grouped together with other posts using the same hashtag in order to aid discoverability of your posts.

Most ActivityPub server implementations like Mastodon and similar only group together **Public** posts by the hashtags they use, but there is no guarantee about that. Generally speaking then, it is better to only use hashtags for Public visibility posts where you want the post to be able to spread more widely than it would otherwise. A good example of this is the `#introduction` hashtag, which tends to be used by new accounts who want to introduce themselves to the fediverse!

Including hashtags in your post works like most other social media software: just add a `#` symbol before the word you want to use as a hashtag.

Some examples:

* `#introduction`
* `#Mosstodon`
* `#LichenSubscribe`

Hashtags in GoToSocial are case-insensitive, so it doesn't matter if you use uppercase, lowercase, or a mixture of both when writing your hashtag, it will still count as the same hashtag. For example, `#Introduction` and `#introduction` are treated exactly the same.

For accessibility reasons, it is considerate to use upper camel case when you're writing hashtags. In other words: capitalize the first letter of every word in the hashtag. So rather than writing `#thisisahashtag`, which is difficult to read visually, and difficult for screenreaders to read out loud, consider writing `#ThisIsAHashtag` instead.

You can include as many hashtags as you like within a GoToSocial post, and each hashtag has a length limit of 100 characters.

!!! tip
    To end a hashtag, you can simply use a space, for example in the text `this #soup rules`, the hashtag is terminated by a space so `#soup` becomes the hashtag. However, you can also use a pipe character `|`, or the unicode characters `\u200B` (zero-width no-break space) or `\uFEFF` (zero-width space), to create "partial-word" hashtags. For example, with input text `this #so|up rules`, only the `#so` part becomes the hashtag. Likewise, with the input text `this #so​up rules`, which contains an invisible zero-width space after the o and before the u, only the `#so` part becomes the hashtag. See here for more information on zero-width spaces: https://en.wikipedia.org/wiki/Zero-width_space.

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

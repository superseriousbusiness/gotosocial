# Search

## Query formats

GotoSocial accepts several kinds of search query:

- `@username`: search for an account with the given username on any domain. Can return multiple results.
- `@username@domain`: search for a remote account with exact username and domain. Will only ever return 1 result at most.
- `https://example.org/some/arbitrary/url`: search for an account or post with the given URL. If the account or post hasn't already federated to GotoSocial, it will try to retrieve it. Will only ever return 1 result at most.
- `#hashtag_name`: search for a hashtag with the given hashtag name, or starting with the given hashtag name. Case insensitive. Can return multiple results.
- `any arbitrary text`: search for posts containing the text, hashtags containing the text, and accounts with usernames, display names, or bios containing the text, exactly as written. Can return multiple results. Bios will only be searched for accounts that you follow.

## Search operators

Arbitrary text queries may include the following search operators:

- `from:username`: restrict results to statuses created by the specified *local* account.
- `from:username@domain`: restrict results to statuses created by the specified remote account.

For example, you can search for `sloth from:yourusername` to find your own posts about sloths.

# Domain Permission Subscriptions

Via the [admin settings panel](./settings.md#subscriptions), you can create and manage domain permission subscriptions.

Domain permission subscriptions allow you to specify a URL at which a permission list is hosted. Every 24hrs at 11pm (by default), your instance will fetch and parse each list you're subscribed to, in order of priority (highest to lowest), and create domain permissions (or domain permission drafts) based on entries discovered in the lists.

Each domain permission subscription can be used to create domain allow or domain block entries.

!!! warning
    Currently, via blocklist subscriptions it is only possible to create "suspend" level domain blocks; other severities are not yet supported. Entries of severity "silence" or "limit" etc. on subscribed blocklists will be skipped.

## Priority

When you specify multiple domain permission subscriptions, they will be fetched and parsed in order of priority, from highest priority (255) to lowest priority (0).

Permissions discovered on lists higher up in the priority ranking will override permissions on lists lower down in the priority ranking.

For example, an instance admin subscribes to two allow lists, "Important List" at priority 255, and "Less Important List" at priority 128. Each of these subscribed lists contain an entry for `good-eggs.example.org`.

The subscription with the higher priority is the one that now creates and manages the domain allow entry for `good-eggs.example.org`.

If the subscription with the higher priority is removed, then the next time all the subscriptions are fetched, "Less Important List" will create (or take ownership of) the domain allow instead.

## Retractions

Sometimes, an entry that was present on a subscribed block or allow list will be removed later by the curator(s) of that list. When this happens, the removed domain permission entry can be said to have been "retracted".

For example, say your instance subscribes to one block list, and that block list contains an entry for `baddies.example.org`. A corresponding domain block for `baddies.example.org` has therefore been created in your database, with the subscription ID of your block list. In other words, the domain block is in force, and is managed by your block list subscription.

At some point, your instance fetches the list again, and this time it sees that the entry for `baddies.example.org` is no longer present in the list, because it has been removed by the list curator(s) (perhaps the admins turned their policies around, or the instance was shut down, etc). Thus, according to your instance, the block for `baddies.example.org` is now a "retracted" domain permission entry.

If the domain permission subscription is set to "Remove retracted permissions," then the now-retracted domain block will be removed from the database, and will no longer be enforced. In this example, that means your instance will start federating (again) with `baddies.example.org`.

If the domain permission subscription is *not* set to "Remove retracted permissions," then instead of the retracted block being removed from the database, it will be kept in the database but "orphaned" -- ie., it will still be in force, but it will be marked as no longer being managed by the subscription. In this example, that means your instance will keep blocking `baddies.example.org`.

!!!! Note "Retracted permissions and other subscriptions"
    When a permission is retracted and removed from the database, but an entry for it exists on the list of *another* subscription of a lower priority than the one it was retracted from, then the permission will be recreated as an entry managed by the lower priority list.

    For example, say you subscribe to List1 at priority 255, and List2 at priority 128, and `baddies.example.org` is present on both lists. That means the domain block entry will be managed by List1. If List1 later *retracts* the entry, it will be removed from your database (assuming you have "Remove retracted permissions" set). However, as soon as List2 is checked (usually seconds after List1), then an entry for `baddies.example.org` will be created again, but managed by List2 this time.

    In other words, it is only when an entry is retracted from *every list you subscribe to* that it will truly be removed.

## Orphan Permissions

Domain permissions (blocks or allows) that are not currently managed by a domain permission subscription are considered "orphan" permissions. This includes permissions that an admin created in the settings panel by hand, entries which were imported manually via the import/export page, or entries that belonged to a subscription but have since been retracted but not removed.

If you wish, when creating a domain permission subscription, you can set ["adopt orphans"](./settings.md#adopt-orphan-permissions) to true for that subscription. If a domain permission subscription that is set to adopt orphans encounters an orphan permission which is *also present on the list at the subscription's URI*, then it will "adopt" the orphan by setting the orphan's subscription ID to its own ID.

For example, an instance admin manually creates a domain block for the domain `horrid-trolls.example.org`. Later, they create a domain permission subscription for a block list that contains an entry for `horrid-trolls.example.org`, and they set "adopt orphans" to true. When their instance fetches and parses the list, and creates domain permission entries from it, then the orphan domain block for `horrid-trolls.example.org` gets adopted by the domain permission subscription. Now, if the domain permission subscription is removed, and the option to remove all permissions owned by the subscription is checked, then the domain block for `horrid-trolls.example.org` will also be removed.

## Fun Stuff To Do With Domain Permission Subscriptions

### 1. Create an allowlist-federation cluster.

Domain permission subscriptions make it possible to easily create allowlist-federation clusters, ie., a group of instances can essentially form their own mini-fediverse, wherein each instance runs in [allowlist federation mode](./federation_modes.md#allowlist-federation-mode), and subscribes to a cooperatively-managed allowlist hosted somewhere.

For example, instances `instance-a.example.org`, `instance-b.example.org`, and `instance-c.example.org` decide that they only want to federate with each other.

Using some version management platform like Codeberg, they host a plaintext-formatted allowlist at something like `https://codeberg.org/our-cluster/allowlist/raw/branch/main/allows.txt`.

The contents of the plaintext-formatted allowlist are as follows:

```text
instance-a.example.org
instance-b.example.org
instance-c.example.org
```

Each instance admin sets their federation mode to `allowlist`, and creates a subscription to create allows from `https://codeberg.org/our-cluster/allowlist/raw/branch/main/allows.txt`, which results in domain allow entries being created for their own domain, and for each other domain in the cluster.

At some point, someone from `instance-d.example.org` asks (out of band) whether they can be added to the cluster. The existing admins agree, and update their plaintext-formatted allowlist to read:

```text
instance-a.example.org
instance-b.example.org
instance-c.example.org
instance-d.example.org
```

The next time each instance fetches the list, a new domain allow entry will be created for `instance-d.example.org`, and it will be able to federate with the other domains on the list.

### 2. Cooperatively manage a blocklist.

Domain permission subscriptions make it easy to collaborate on and subscribe to shared blocklists of domains that host illegal / fashy / otherwise undesired accounts and content.

For example, the admins of instances `instance-e.example.org`, `instance-f.example.org`, and `instance-g.example.org` decide that they are tired of duplicating work by playing whack-a-mole with bad actors. To make their lives easier, they decide to collaborate on a shared blocklist.

Using some version management platform like Codeberg, they host a blocklist at something like `https://codeberg.org/our-cluster/allowlist/raw/branch/main/blocks.csv`.

When someone discovers a new domain hosting an instance they don't like, they can open a pull request or similar against the list, to add the questionable instance to the domain.

For example, someone gets an unpleasant reply from a new instance `fashy-arseholes.example.org`. Using their collaboration tools, they propose adding `fashy-arseholes.example.org` to the blocklist. After some deliberation and discussion, the domain is added to the list.

The next time each of `instance-e.example.org`, `instance-f.example.org`, and `instance-g.example.org` fetch the block list, a block entry will be created for ``fashy-arseholes.example.org``.

### 3. Subscribe to a blocklist, but ignore some of it.

Say that `instance-g.example.org` in the previous section decides that they agree with most of the collaboratively-curated blocklist, but they actually would like to keep federating with ``fashy-arseholes.example.org`` for some godforsaken reason.

This can be done in one of three ways:

1. The admin of `instance-g.example.org` subscribes to the shared blocklist, but they do so with the ["create as drafts"](./settings.md#create-permissions-as-drafts) option set to true. When their instance fetches the blocklist, a draft block is created for `fashy-arseholes.example.org`. The admin of `instance-g` just leaves the permission as a draft, or rejects it, so it never comes into force.
2. Before the blocklist is re-fetched, the admin of `instance-g.example.org` creates a [domain permission exclude](./settings.md#excludes) entry for ``instance-g.example.org``. The domain ``instance-g.example.org`` then becomes exempt/excluded from automatic permission creation, and so the block for ``instance-g.example.org`` on the shared blocklist does not get created in the database of ``instance-g.example.org`` the next time the list is fetched.
3. The admin of `instance-g.example.org` creates an explicit domain allow entry for `fashy-arseholes.example.org` on their own instance. Because their instance is running in `blocklist` federation mode, [the explicit allow overrides the domain block entry](./federation_modes.md#in-blocklist-mode), and so the domain remains unblocked.

### 4. Subscribe directly to another instance's blocklist.

Because GoToSocial is able to fetch and parse JSON-formatted lists of domain permissions, it is possible to subscribe directly to another instance's list of blocked domains via their `/api/v1/instance/domain_blocks` (Mastodon) or `/api/v1/instance/peers?filter=suspended` (GoToSocial) endpoint (if exposed).

For example, the Mastodon instance `peepee.poopoo.example.org` exposes their block list publicly, and the owner of the GoToSocial instance `instance-h.example.org` decides they quite like the cut of the Mastodon moderator's jib. They create a domain permission subscription of type JSON, and set the URI to `https://peepee.poopoo.example.org/api/v1/instance/domain_blocks`. Every 24 hours, their instance will go fetch the blocklist JSON from the Mastodon instance, and create permissions based on entries discovered therein.

## Example lists per content type

Shown below are examples of the different permission list formats that GoToSocial is able to understand and parse.

Each list contains three domains, `bumfaces.net`, `peepee.poopoo`, and `nothanks.com`.

### CSV

CSV lists use content type `text/csv`.

Mastodon domain permission exports generally use this format.

```csv
#domain,#severity,#reject_media,#reject_reports,#public_comment,#obfuscate
bumfaces.net,suspend,false,false,big jerks,false
peepee.poopoo,suspend,false,false,harassment,false
nothanks.com,suspend,false,false,,false
```

### JSON (application/json)

JSON lists use content type `application/json`.

```json
[
  {
    "domain": "bumfaces.net",
    "suspended_at": "2020-05-13T13:29:12.000Z",
    "comment": "big jerks"
  },
  {
    "domain": "peepee.poopoo",
    "suspended_at": "2020-05-13T13:29:12.000Z",
    "comment": "harassment"
  },
  {
    "domain": "nothanks.com",
    "suspended_at": "2020-05-13T13:29:12.000Z"
  }
]
```

As an alternative to `"comment"`, `"public_comment"` will also work:

```json
[
  {
    "domain": "bumfaces.net",
    "suspended_at": "2020-05-13T13:29:12.000Z",
    "public_comment": "big jerks"
  },
  {
    "domain": "peepee.poopoo",
    "suspended_at": "2020-05-13T13:29:12.000Z",
    "public_comment": "harassment"
  },
  {
    "domain": "nothanks.com",
    "suspended_at": "2020-05-13T13:29:12.000Z"
  }
]
```

### Plaintext (text/plain)

Plaintext lists use content type `text/plain`.

Note that it is not possible to include any fields like "obfuscate" or "public comment" in plaintext lists, as they are simply a newline-separated list of domains.

```text
bumfaces.net
peepee.poopoo
nothanks.com
```

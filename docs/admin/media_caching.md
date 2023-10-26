# Media Caching

GoToSocial uses the configured [storage backend](https://docs.gotosocial.org/en/latest/configuration/storage/) in order to store media (images, videos, etc) uploaded to the instance by local users, as well as to cache media attached to posts and profiles federated in from remote instances.

Media uploaded by local instance users will be kept in storage forever (unless the post or profile it's attached to is deleted), so that it's always available to be served in response to requests coming from remote instances.

Remote media, on the other hand, is cached only temporarily. After a certain amount of time (see below), it will be removed from storage to help alleviate storage space usage. Remote media uncached this way will be re-fetched automatically from the remote instance if it's needed again.

!!! info "Why cache?"
    There is an argument to be made for not caching remote media at all, since it's always available on the origin server. Why not just forego caching entirely, and rely on the remote instance to serve everything on demand?
    
    While this is a straightforward approach to saving storage space, it can cause other problems and is generally considered to be rather impolite.
    
    For example, say someone from a small instance makes a funny post with an image attached. The post gets boosted by an account that's followed by 1,000 people across 5 different instances (200 on each instance). Each of those 1,000 people then have the image put in their timeline at once.
    
    With no remote media caching in place, this may cause up to 1,000 requests to hit the small instance simultaneously, as the browser of each recipient of the post must go and make a unique request to fetch the image from the small instance. This causes a large traffic spike for the small instance. In extreme scenarios, this can cause the instance to become unresponsive or crash, essentially DDOS'ing it.
    
    With remote media caching in place, however, boosting a post to 1,000 people across 5 different instances will cause only 5 requests to the small instance: 1 request for each instance. Each instance will then serve 200 requests to its local users from the cached version of the remote image, effectively spreading the load and sparing the smaller instance.

## Cleanup

Cleanup of the remote media cache occurs as a scheduled background process, and no manual intervention is required by admins. Cleanup takes somewhere between 5-30 minutes depending on the speed of the server, the speed of the configured storage, and the amount of media to work through.

GoToSocial exposes three variables that let you, the admin, tune when and how this work is performed: `media-remote-cache-days`, `media-cleanup-from` and `media-cleanup-every`.

By default, these variables are set to the following values:

| Variable name             | Default      | Meaning  |
|---------------------------|--------------|----------|
| `media-remote-cache-days` | `7`          | 7 days   |
| `media-cleanup-from`      | `"00:00:00"` | midnight |
| `media-cleanup-every`     | `"24h"`      | daily    |

In other words, the default settings mean that every night at midnight, remote media older than a week will be uncached and removed from storage.

You can achieve different results by tuning these variables. For example, say you wanted to prune at 4.30am instead of midnight, you could change `media-cleanup-from` to `"04:30:00"`.

If you only want to prune every couple of days instead of every night, you could set `media-cleanup-every` to a higher value, like `"48h"` or `"72h"`.

If you wanted to adopt a more aggressive cleanup strategy to minimize storage usage, you could set the following values:

| Variable name             | Setting      | Meaning     |
|---------------------------|--------------|-------------|
| `media-remote-cache-days` | `1`          | 1 day       |
| `media-cleanup-from`      | `"00:00:00"` | midnight    |
| `media-cleanup-every`     | `"8h"`       | every 8 hrs |

The above settings would mean that every 8 hours starting from midnight, GoToSocial would prune any media older than 1 day (24hrs). The prune jobs would run at 00:00, 08:00, and 16:00, ie., midnight, 8am, and 4pm. With this configuration, the longest amount of time you could possibly keep remote media in your storage would be about 32 hours.

!!! tip
    Setting `media-remote-cache-days` to 0 or less means that remote media will never be uncached. However, cleanup jobs for orphaned local media and other consistency checks will still be run using the schedule defined by the other variables.

!!! tip
    You can also run cleanup manually as a one-off action through the admin panel, if you so wish ([see docs](./settings.md#media)).

!!! warning
    Setting `media-cleanup-every` to a very small value like `"30m"` or less will probably cause your instance to just constantly iterate through attachments, causing high database use for very little benefit. We don't recommend setting this value to less than about `"8h"` and even that is probably overkill.

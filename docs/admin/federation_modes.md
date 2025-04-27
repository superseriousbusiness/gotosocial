# Federation Modes

GoToSocial currently offers both 'blocklist' and 'allowlist' federation modes, which can be set using the `instance-federation-mode` setting in the config.yaml, or using the `GTS_INSTANCE_FEDERATION_MODE` environment variable. These are described below.

## Blocklist federation mode (default)

When `instance-federation-mode` is set to `blocklist`, your instance will federate openly with other instances, without restriction, with the exception of instances you have explicitly created domain blocks for via the settings panel.

When your instance receives a new request from an instance that is not blocked via a domain block entry, it will serve the request if the request is valid, and if the requester is permitted to view the resource that's being requested (taking account of status visibility, and any user-level blocks).

When your instance encounters a mention or an announce of a status or account it hasn't seen before, it will go fetch the resource if the domain of the resource is not blocked via a domain block entry.

!!! info
    Blocklist federation mode is the default federation mode for GoToSocial. It's also the default for most other ActivityPub server implementations.

## Allowlist federation mode

!!! warning
    Allowlist federation mode is still considered "experimental" while we work out how well it actually works in practice. It should do what it says on the box, but it may cause bugs or edge cases to appear elsewhere, we're not sure yet!

When `instance-federation-mode` is set to `allowlist`, your instance will federate only with instances for which an explicit allow has been created via the settings panel, and will restrict access by any instances for which an allow has not been created.

When your instance receives a new request from an instance that is not explicitly allowed via a domain allow entry, it will refuse to serve the request. If the request comes from a domain that is on the allowlist, your instance will serve the request (taking account of status visibility, and any user-level blocks).

When your instance encounters a mention or an announce of a status or account it hasn't seen before, it will only go fetch the resource if the domain of the resource is explicitly allowed via a domain allow entry.

!!! tip
    Allowlist federation mode is useful in cases where you want to federate only with select 'trusted' instances. However, this comes at the cost of hampering discovery. Under blocklist federation mode, you will organically encounter posts and accounts from instances you were not yet aware of, via boosts and replies, but in allowlist federation mode no such serendipity will occur.
    
    As such, it is recommended that you *either* start with blocklist federation mode and switch over to allowlist federation later on once you've established which other instances you 'like', *or* you start with allowlist federation mode, and have an allowlist populated and ready to import after first booting up your instance, in order to 'bootstrap' it.

## Combining blocks and allows

!!! danger
    Combining blocks and allows is a tricky business!
    
    When importing lists of allows and blocks, you should always review the list manually to make sure that you do not inadvertently block a domain that you would prefer not to block, since this can have **very annoying side effects** like removing follows/following, statuses, etc.
    
    When in doubt, always add an explicit allow first as an insurance policy!

It is possible to both block and allow the same domain, and the effect of combining these two things depends on which federation mode your instance is currently using, as explained in the following subsections, which are summarised in a flowchart that you can find below.

### In blocklist mode

As the chart below shows, in blocklist mode (the left-hand part of the diagram), an explicit domain allow can be used to override a domain block.

This is useful in cases where you are importing a blocklist from someone else, but the imported blocklist contains some instances you would actually prefer not to block. To avoid blocking those instances, you can create explicit domain allows for those instances first. Then, when you import the block list, the explicitly allowed domains will not be blocked, and the side effects of creating a block (deleting statuses, media, relationships etc) will not be processed.

If you later remove an explicit allow for a domain that also has a block, the instance will become blocked, and side effects of block creation will be processed.

Conversely, if you add an explicit allow for a domain that was blocked, the side effects of block *deletion* will be processed.

### In allowlist mode

As the chart below shows, in allowlist mode (the right-hand part of the diagram) an explicit domain block trumps an explicit domain allow. The following two things must be true in order for an instance to be allowed through, when running in allowlist mode:

1. An explicit domain block **does not exist** for the instance.
2. An explicit domain allow **does exist** for the instance.

If either of the above conditions are not met, the request will be denied.

![A flow chart diagram showing how the two different federation modes treat incoming requests.](../public/diagrams/federation_modes.png)

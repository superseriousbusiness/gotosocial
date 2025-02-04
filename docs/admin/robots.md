# Robots.txt

GoToSocial serves a `robots.txt` file on the host domain. This file contains rules that attempt to block known AI scrapers, as well as some other indexers. It also includes some rules to ensure things like API endpoints aren't indexed by search engines since there really isn't any point to them.

## Allow/disallow stats collection

You can allow or disallow crawlers from collecting stats about your instance from the `/nodeinfo/2.0` and `/nodeinfo/2.1` endpoints by changing the setting `instance-stats-mode`, which modifies the `robots.txt` file. See [instance configuration](../configuration/instance.md) for more details.

## AI scrapers

The AI scrapers come from a [community maintained repository][airobots]. It's manually kept in sync for the time being. If you know of any missing robots, please send them a PR!

A number of AI scrapers are known to ignore entries in `robots.txt` even if it explicitly matches their User-Agent. This means the `robots.txt` file is not a foolproof way of ensuring AI scrapers don't grab your content.
    
If you want to block these things fully, you'll need to block based on the User-Agent header in a reverse proxy until GoToSocial can filter requests by User-Agent header.

[airobots]: https://github.com/ai-robots-txt/ai.robots.txt/

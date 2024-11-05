# Robots.txt

GoToSocial 在主域名上提供一个 `robots.txt` 文件。该文件包含试图屏蔽已知 AI 爬虫的一些规则，以及其他一些索引器。它还包括一些规则，以确保诸如 API 端点之类的内容不会被搜索引擎索引，因为这些内容没有被索引的必要。

## AI 爬虫

AI 爬虫来自一个[社区维护的仓库][airobots]。目前是手动保持同步的。如果你知道有任何遗漏的爬虫，请给他们提交一个 PR！

已知有许多 AI 爬虫即便明确匹配其 User-Agent，也会忽略 `robots.txt` 中的条目。这意味着 `robots.txt` 文件并不是确保 AI 爬虫不抓取你的内容的万无一失的方法。

如果你想完全屏蔽这些爬虫，需要在反向代理中根据 User-Agent 头进行屏蔽，直到 GoToSocial 能够根据 User-Agent 头过滤请求。

[airobots]: https://github.com/ai-robots-txt/ai.robots.txt/

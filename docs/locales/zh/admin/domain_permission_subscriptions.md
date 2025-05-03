# 域名权限订阅

你可以通过[管理设置面板](./settings.md#订阅)创建和管理域名权限订阅。

域名权限订阅允许你指定一个域名权限列表托管的URL。默认情况下，每24小时在当前时区晚上11点进行自动更新，你的实例将获取并解析你订阅的每个列表，基于在列表中发现的条目，按照优先级（从高到低）顺序创建域（或域名权限草稿）。

每个域名权限订阅可以用来创建域名允许或域名阻止条目。

!!! warning "警告"
    目前，通过阻止列表订阅只能创建“屏蔽”级别的域名阻止条目；其他严重程度尚不支持。订阅阻止列表中严重程度为“隐藏”或“限制”等的条目将被跳过。

## 优先级

当存在多个域名权限订阅时，它们将按照优先级顺序（从最高优先级（255）到最低优先级（0））被获取和解析。

在优先级较高的列表上发现的权限条目将覆盖优先级较低的列表上的权限条目。

例如，一名实例管理员订阅了两个允许列表，“重要列表”优先级为255，“不太重要的列表”优先级为128。每个订阅列表都包含了`good-eggs.example.org`的条目。

那么优先级较高的订阅会负责创建和管理`good-eggs.example.org`的域名允许条目。

如果移除了优先级较高的订阅，那么下次获取所有订阅时，“不太重要的列表”将创建（或接管）该域名允许条目。

## 孤立权限

目前没有被域名权限订阅管理的域名权限条目（阻止条目和允许条目）被认为是“孤立”权限。这包括管理员手动在设置面板中创建的权限，或者是通过导入/导出页面手动导入的权限。

如果你愿意，在创建域名权限订阅时，可以将该订阅的[“接管孤立权限条目”](./settings.md#接管孤立权限条目)设置为 true。如果一个启用了“接管孤立权限条目”的域名权限订阅遇到一个孤立权限，并且该条目 *也在该订阅地址指向的列表中*，那么它将把该孤立条目的订阅ID设置为其自身ID，来“接收”此孤立条目。

例如，一个实例管理员手动为域名`horrid-trolls.example.org`创建了域名阻止条目。稍后，他们创建了一个域名阻止列表订阅，并将“收养孤儿”设置为真，且该订阅包含`horrid-trolls.example.org`。当实例获取并解析列表,并从中创建域名权限条目时，`horrid-trolls.example.org`这个孤立的域名阻止条目将被刚刚配置的域名权限订阅接收。现在，如果域名权限订阅被移除，且在移除时勾选了移除订阅所拥有的所有权限选项，那么`horrid-trolls.example.org`这个域名阻止条目也将被移除。

## 域名权限订阅的几种有趣的应用场景

### 1. 创建白名单联合实例集群

域名权限订阅使得创建白名单联合实例集群集群变得更加容易，也就是说，一组实例理论上可以形成自己的迷你联邦宇宙，每个实例在[白名单联合模式](./federation_modes.md#白名单联合模式)下运行，并订阅同一个合作管理的、托管在某处的允许列表。

例如，实例 `instance-a.example.org`、`instance-b.example.org` 和 `instance-c.example.org` 决定他们只想彼此联合。

他们可以使用像 Codeberg 这样的版本管理平台托管一个纯文本格式的允许列表，比如在 `https://codeberg.org/our-cluster/allowlist/raw/branch/main/allows.txt`。

纯文本格式的允许列表内容如下:

```text
instance-a.example.org
instance-b.example.org
instance-c.example.org
```

每个实例管理员都将他们的联合模式设置为`allowlist`，并创建一个类型为“允许”，订阅地址为 `https://codeberg.org/our-cluster/allowlist/raw/branch/main/allows.txt` 的订阅，这会为他们自己的域名以及集群中的其他域名创建域名允许条目。

在某个时候，来自 `instance-d.example.org` 的某人（在站外）申请被添加到集群中。现有的管理员同意，并更新他们的纯文本格式允许列表为:

```text
instance-a.example.org
instance-b.example.org
instance-c.example.org
instance-d.example.org
```

下次每个实例获取列表时，将为 `instance-d.example.org` 创建一个新的域名允许条目,它将能够与该列表中的其他域进行联合。

### 2. 合作管理阻止列表

域名权限订阅使得合作管理和订阅共享的、包含非法/极右/其他不良账户和内容的域名的阻止列表变得容易。

例如，实例 `instance-e.example.org`、`instance-f.example.org` 和 `instance-g.example.org` 的管理员认定：他们厌倦了通过与坏人玩打地鼠游戏来重复工作。为了让生活更轻松，他们决定合作开发一个共享的阻止列表。

他们使用像 Codeberg 这样的版本管理平台在类似 `https://codeberg.org/our-cluster/allowlist/raw/branch/main/blocks.csv` 的地方托管一个阻止列表。

当有人发现另一个他们不喜欢的实例时，他们可以通过合并请求或类似方法添加这个有问题的实例到域名列表中。

例如，有人从一个新实例 `fashy-arseholes.example.org` 收到一个不愉快的回复。他们使用他们的协作工具,建议将 `fashy-arseholes.example.org` 添加到阻止列表。经过一些审议和讨论后,该域被添加到列表中。

下次 `instance-e.example.org`、`instance-f.example.org` 和 `instance-g.example.org` 获取阻止列表时，将为 `fashy-arseholes.example.org` 创建一个阻止条目。

### 3. 订阅阻止列表，但忽略其中的一部分

假设上一节中的 `instance-g.example.org` 认定他们同意大部分协作策划的阻止列表，但出于某种原因，他们实际上希望继续与 `fashy-arseholes.example.org` 联合。

这可以通过以下三种方法实现:

1. `instance-g.example.org` 的管理员订阅共享阻止列表，但他们将其["创建为草稿"](./settings.md#将此条目设为草稿)选项设置为 true。当他们的实例获取阻止列表时，会为 `fashy-arseholes.example.org` 创建一个阻止条目草稿。`instance-g` 的管理员只需将权限保留为草稿或拒绝它，因此它永远不会生效。
2. 在重新获取阻止列表之前，`instance-g.example.org` 的管理员为 `instance-g.example.org` 创建一个[域名权限例外](./settings.md#例外)条目。设置保存后，域名权限订阅将无法`instance-g.example.org` 域名创建权限，因此在列表下次被获取时，共享阻止列表上对于 `instance-g.example.org` 的阻止不会在 `instance-g.example.org` 的实例数据库中创建。
3. `instance-g.example.org` 的管理员在其实例上为 `fashy-arseholes.example.org` 创建一个显式的域名允许条目。`instance-g` 实例在`黑名单`联合模式下运行，因此[显式允许条目将覆盖域名阻止条目](./federation_modes.md#黑名单模式)。`fashy-arseholes` 域名将保持未被阻止的状态。

### 4. 直接订阅另一个实例的阻止列表

GoToSocial 能够获取和解析 JSON 格式的域名权限列表，所以可以通过他们的 `/api/v1/instance/domain_blocks` （Mastodon） 或 `/api/v1/instance/peers?filter=suspended` （GoToSocial）端点（如果已公开）直接订阅另一个实例的屏蔽列表。

例如，Mastodon 实例 `peepee.poopoo.example.org` 公开他们的阻止列表，而GoToSocial实例的所有者 `instance-h.example.org` 认定他们非常喜欢该 Mastodon 管理员的标准。他们创建一个JSON类型的域名权限订阅，并将地址设为 `https://peepee.poopoo.example.org/api/v1/instance/domain_blocks`。他们的实例将每24小时获取一次对方 Mastodon 实例的阻止列表JSON，并根据其中发现的条目创建权限。

## 域名权限订阅列表的格式示例

以下是 GoToSocial 能够解析的不同权限列表格式的示例。

每个列表包含三个域，`bumfaces.net`、`peepee.poopoo` 和 `nothanks.com`。

### CSV

CSV列表使用内容类型 `text/csv`。

Mastodon域名权限通常使用这种格式导出。

```csv
#domain,#severity,#reject_media,#reject_reports,#public_comment,#obfuscate
bumfaces.net,suspend,false,false,这个实例上有坏蛋,false
peepee.poopoo,suspend,false,false,骚扰,false
nothanks.com,suspend,false,false,,false
```

### JSON (application/json)

JSON列表使用内容类型 `application/json`。

```json
[
  {
    "domain": "bumfaces.net",
    "suspended_at": "2020-05-13T13:29:12.000Z",
    "comment": "这个实例上有坏蛋"
  },
  {
    "domain": "peepee.poopoo",
    "suspended_at": "2020-05-13T13:29:12.000Z",
    "comment": "骚扰"
  },
  {
    "domain": "nothanks.com",
    "suspended_at": "2020-05-13T13:29:12.000Z"
  }
]
```

可以使用 `"comment"` 字段替代 `"public_comment"`:

```json
[
  {
    "domain": "bumfaces.net",
    "suspended_at": "2020-05-13T13:29:12.000Z",
    "public_comment": "这个实例上有坏蛋"
  },
  {
    "domain": "peepee.poopoo",
    "suspended_at": "2020-05-13T13:29:12.000Z",
    "public_comment": "骚扰"
  },
  {
    "domain": "nothanks.com",
    "suspended_at": "2020-05-13T13:29:12.000Z"
  }
]
```

### 纯文本 (text/plain)

纯文本列表使用内容类型 `text/plain`。

注意在纯文本列表中无法包含像“obfuscate”或“public comment”这样的字段，因为它们只是一个以换行符分隔的域名列表。

```text
bumfaces.net
peepee.poopoo
nothanks.com
```

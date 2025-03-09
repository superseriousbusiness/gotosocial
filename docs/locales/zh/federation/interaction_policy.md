# 互动规则

GoToSocial 在帖文中使用 `interactionPolicy` 属性向外站实例描述：对于任意给定的帖文，本站服务器允许处理和存储哪些类型的互动及其条件。

`interactionPolicy` 及相关对象和属性的 `@context` 文档位于 `https://gotosocial.org/ns`。

!!! danger 警告
    互动规则旨在限制帖文作者不希望看到的回复以及其它可能对用户造成有害影响的互动（例如，来自“回复狂”的互动）。
    
    然而，这一技术方案远未足以解决此问题，因为在用户最初的希望设定的回复范围之外，仍存在许多“常规途径之外”的分发或回复方式。
    
    例如，用户可能会创建一个具有非常严格的互动规则的帖文，但发现其他软件和实例并不遵守这一规则，其他实例上的用户可能正在他们各自实例的范围讨论这个帖文并回复。原帖文作者所在的实例会自动将这些作者不希望出现的互动从视图中丢弃，但外站实例可能仍会显示它们。
    
    再举一个例子：有人可能会看到一个规定“任何人都不能回复”的帖文，但他截屏了该帖文，然后在自己的新帖文中发布该截屏，并提及原帖作者。另外，用户也可能发布链接到该帖文的 URL，并以提及形式标记原帖作者。在这种情况下，他们通过创建一个新帖文有效地“回复”了原来的帖文。
    
    考虑到上述情形，GoToSocial 只能提供一种尽最大努力、部分解决问题的技术方案，该方案之外的情形更多是关乎社交行为和边界问题。

## 概览

`interactionPolicy` 是附加在类似帖文的 `Object`（例如 `Note`、`Article`、`Question` 等）上的一个对象属性，其格式如下：

```json
{
  "@context": [
    "https://gotosocial.org/ns",
    [...]
  ],
  [...],
  "interactionPolicy": {
    "canLike": {
      "always": [ "零个或多个总是允许此操作的 URI" ],
      "approvalRequired": [ "零个或多个需要批准才能执行此操作的 URI" ]
    },
    "canReply": {
      "always": [ "零个或多个总是允许此操作的 URI" ],
      "approvalRequired": [ "零个或多个需要批准才能执行此操作的 URI" ]
    },
    "canAnnounce": {
      "always": [ "零个或多个总是允许此操作的 URI" ],
      "approvalRequired": [ "零个或多个需要批准才能执行此操作的 URI" ]
    }
  },
  [...]
}
```

在 `interactionPolicy` 对象中：

- `canLike` 是一个子规则，用于表明哪些对象（或集合）被允许以帖文 URI 作为 `Like` 的 `object` 来创建一个点赞活动。
- `canReply` 是一个子规则，用于表明哪些对象（或集合）被允许以将 `inReplyTo` 设置为该帖文的 URI/ID 来创建一个回复帖文。
- `canAnnounce` 是一个子规则，用于表明哪些对象（或集合）被允许以帖文 URI/ID 作为 `Announce` 的 `object` 来创建一个转发活动。

另外：

- `always` 表示 ActivityPub URI/ID 中的 `Actor` 或者 `Actor` 集合，它们被允许在不需要帖文作者手动批准的情况下创建并分发针对帖文的互动。
- `approvalRequired` 表示 ActivityPub URI/ID 中的 `Actor` 或者 `Actor` 集合，它们被允许创建针对帖文的互动，但应当等待帖文作者手动批准后再进行分发（详见[请求、获取和验证批准](#请求获取和验证批准)）。

在 `always` 和 `approvalRequired` 中，合法的 URI 项目包括：

- 特殊的 ActivityStreams 公共 URI `https://www.w3.org/ns/activitystreams#Public`
- 帖文创建者的 `Following` 和／或 `Followers` 集合的 URI
- 单个行为体的 URI

例如：

```json
[
    "https://www.w3.org/ns/activitystreams#Public",
    "https://example.org/users/someone/followers",
    "https://example.org/users/someone/following",
    "https://example.org/users/someone_else",
    "https://somewhere.else.example.org/users/someone_on_a_different_instance"
]
```

!!! info "注意"
    请注意，根据 JSON-LD 规范，`always` 与 `approvalRequired` 的值可以是单个字符串，也可以是字符串数组。也就是说，以下几种写法都是合法的：
    
    - 单个字符串：`"always": "https://example.org/users/someone"`
    - 单一条目数组：`"always": [ "https://example.org/users/someone" ]`
    - 多个条目的数组：`"always": [ "https://example.org/users/someone", "https://example.org/users/someone_else" ]`

## 指定“没有人”

要指定除作者之外**没有人**可以对帖文进行互动（而作者始终是被允许互动），实现者应将 `always` 数组设置为**仅包含帖文的作者 URI**，而 `approvalRequired` 则可以不包含、设为 `null` 或者留空。

例如，下面的 `canLike` 值表示除帖文作者以外，**没有人**可以对该帖文执行点赞操作：

```json
"canLike": {
  "always": "帖文作者的ActivityPub URI"
},
```

再举一个例子。对于下面这条由 `https://example.org/users/someone` 发布的帖文，它的 `interactionPolicy` 表明任何人都可以点赞该帖文，但只有作者本人可以回复或转发：

```json
{
  "@context": [
    "https://gotosocial.org/ns",
    [...]
  ],
  [...],
  "interactionPolicy": {
    "canLike": {
      "always": "https://www.w3.org/ns/activitystreams#Public"
    },
    "canReply": {
      "always": "https://example.org/users/someone"
    },
    "canAnnounce": {
      "always": "https://example.org/users/someone"
    }
  },
  [...]
}
```

!!! note "注意"
    为了防止恶意行为，GoToSocial 即使在规则中指定“没有人”时，也会对谁能否进行互动做隐含假设，详见[隐含假设](#隐含假设)。

## 冲突 / 重复的值

如果一个用户既存在于某个集合 URI 中，又以单个行为体 URI 的形式被明确指定，则**更为具体**的值将优先采用。

例如：

```json
[...],
"canReply": {
  "always": "https://example.org/users/someone",
  "approvalRequired": "https://www.w3.org/ns/activitystreams#Public"
},
[...]
```

在这里，`@someone@example.org` 出现在 `always` 中，同时又实际包含在 `approvalRequired` 中的特殊公共集合里。此时，由于 `always` 中的值更明确，他们总是可以回复。

另一个例子：

```json
[...],
"canReply": {
  "always": "https://www.w3.org/ns/activitystreams#Public",
  "approvalRequired": "https://example.org/users/someone"
},
[...]
```

这里，`@someone@example.org` 出现在 `approvalRequired` 中，但也实际存在于 `always` 中的特殊公共集合里。在这种情况下，所有人都可以在不需要批准的情况下回复，但 `@someone@example.org` **除外**，它需要批准。

如果相同的 URI 同时存在于 `always` 和 `approvalRequired` 中，则**权限较高的**（即出现在 `always` 中的）值将会优先。

## 默认的 `interactionPolicy`

当帖文中完全没有包含 `interactionPolicy` 属性，或者 `interactionPolicy` 键存在但其值为 `null` 或 `{}` 时，实现者可以假定该帖文具有下面隐含的、默认的 `interactionPolicy`：

```json
{
  "@context": [
    "https://gotosocial.org/ns",
    [...]
  ],
  [...],
  "interactionPolicy": {
    "canLike": {
      "always": "https://www.w3.org/ns/activitystreams#Public"
    },
    "canReply": {
      "always": "https://www.w3.org/ns/activitystreams#Public"
    },
    "canAnnounce": {
      "always": "https://www.w3.org/ns/activitystreams#Public"
    }
  },
  [...]
}
```

默认情况下各子规则中没有任何 `approvalRequired` 属性，也就是说 `approvalRequired` 的默认值为空数组。

该默认 `interactionPolicy` 旨在反映撰写时所有版本低于 v0.17.0 的 GoToSocial 以及其他 ActivityPub 服务器软件实际采用的互动规则。也就是说，这正是那些不支持互动规则的服务器*已默认假定*的互动权限。

!!! info "行为体只能与他们有权查看的帖文进行互动"
    请注意，即使对帖文假定了默认 `interactionPolicy`，帖文的**可见性**仍需通过查看 `to`、`cc` 及／或 `audience` 属性来确认，以确保那些无权*查看*帖文的行为体也无法*互动*。例如，如果一条帖文仅面向粉丝，并且假定了默认 `interactionPolicy`，那么不关注帖文作者的人仍然*不能*看到或者互动该帖文。

!!! tip "提示"
    与其他 ActivityPub 实现规范类似，实现者通常仍希望对针对帖文的转发（`Announce`）操作做限制，当帖文仅粉丝可见时，仅允许作者本人进行相关操作。

## 各子规则的默认值

当某个互动规则仅被*部分*定义（例如，仅设置了 `canReply`，而没有设置 `canLike` 或 `canAnnounce` 键）时，实现者应对 `interactionPolicy` 对象中未定义的每个子规则做如下假定。

!!! tip "未来的扩展可能具有不同默认值"
    请注意，**下述列表并非详尽无遗**，未来对 `interactionPolicy` 的扩展可能希望为其他类型的互动定义**不同的默认值**。

### `canLike`

如果 `interactionPolicy` 中缺失 `canLike`，或 `canLike` 的值为 `null` 或 `{}`，则实现者应假定：

```json
"canLike": {
  "always": "https://www.w3.org/ns/activitystreams#Public"
}
```

换言之，默认情况下**任何能看到帖文的人都可以对其点赞**。

### `canReply`

如果 `interactionPolicy` 中缺失 `canReply`，或 `canReply` 的值为 `null` 或 `{}`，则实现者应假定：

```json
"canReply": {
  "always": "https://www.w3.org/ns/activitystreams#Public"
}
```

换言之，默认情况下**任何能看到帖文的人都可以回复**。

### `canAnnounce`

如果 `interactionPolicy` 中缺失 `canAnnounce`，或 `canAnnounce` 的值为 `null` 或 `{}`，则实现者应假定：

```json
"canAnnounce": {
  "always": "https://www.w3.org/ns/activitystreams#Public"
}
```

换言之，默认情况下**任何能看到帖文的人都可以转发**。

## 描述子规则是否需要验证

在本文撰写时，并非所有服务器都已经实现了互动规则，因此有必要提供一种方法，使实现者可以表明他们**既知道又会执行**下文[互动验证](#互动验证)部分中描述的互动规则。

这种参与互动规则的描述方式，要求服务器在外发帖文时显式设置 `interactionPolicy` 及其子规则，而不依赖于上述默认值。

也就是说，**一个实例通过在帖文上设置 `interactionPolicy.*`即可向其它实例表明其会对每个显式设置的子规则进行互动验证。**

这意味着，如果服务端自己实现了互动规则控制，并希望其他服务端遵循，则应总是显式设置 `interactionPolicy` 上其已实现的所有子规则，即使这些值与隐含默认值没有区别。

例如，如果一个服务器理解并希望强制执行 `canLike`、`canReply` 和 `canAnnounce` 子规则（正如 GoToSocial 的情况），那么他们应当在外发帖文时显式为这些子规则赋值，即使这些值与隐含默认值相同。这让外站服务器知道本站服务器会执行规则，并了解如何处理每个子规则的相应 `Reject`／`Accept` 消息。

另一个例子：如果某个服务器只实现了 `canReply` 的互动子规则，而没有实现 `canLike` 或 `canAnnounce`，那么他们应总是设置 `interactionPolicy.canReply`，并将另外两个子规则排除在 `interactionPolicy` 外，以表明他们无法理解或执行它们。

这种通过键的存在与否来表明参与互动规则的方式，就是为了让大部分未设置 `interactionPolicy` 的服务器（因为它们尚未实现该功能）无需更改行为。已实现互动规则的服务器则可以通过帖文上没有 `interactionPolicy` 键这一特征了解到原始服务器不支持互动规则，并作出相应处理。

## 隐含假设

出于常识性的安全考虑，GoToSocial 做出并始终应用两条关于互动规则的隐含假设。

### 1. 被提及和被回复的行为体总是可以回复

无论帖文的可见性和 `interactionPolicy` 如何，被提及或者被帖文回复中的行为体**总是**可以无需批准即对该帖进行回复，**除非**提及或回复它们的帖文本身正处于待批准状态。

这样设计是为了防止潜在的骚扰者在滥用帖文时提及某人，从而使被提及的用户无从回复。

因此，在发出互动规则时，GoToSocial **总是**会将被提及的用户 URI 加入 `canReply.always` 数组中，除非这些用户已经被 ActivityStreams 的特殊公共 URI 所覆盖。

同样，在执行接收到的互动规则时，GoToSocial 会**始终**将行为体当作已出现在 `canReply.always` 数组中，即使实际数据中没有包含他们的 URI。

### 2. 行为体始终可以对自己的帖文进行任何形式的互动

**其次**，行为体**始终**应该能够对自己的帖文进行回复、点赞和转发（boost），而无需批准，**除非**该帖文本身正处于待批准状态。

因此，在发出互动规则时，GoToSocial **总是**会将帖文作者的 URI 加入到 `canLike.always`、`canReply.always` 和 `canAnnounce.always` 数组中，**除非**这些 URI 已经被 ActivityStreams 的特殊公共 URI 所涵盖。

同样，在执行接收到的互动规则时，GoToSocial 会**始终**将帖文作者视作出现在每个 `always` 字段中，即使实际数据中不存在。

## 示例

这里给出了一些有关互动规则允许用户操作的示例。

### 1. 限制讨论范围

在下面的示例中，用户 `@the_mighty_zork` 希望与用户 `@booblover6969` 和 `@hodor` 开启一段讨论。

为了防止讨论被其他人插话而偏离，Ta 希望帖文的回复（除这三位参与者外）必须经过 `@the_mighty_zork` 的批准才能生效。

此外，Ta 希望只允许自己的粉丝以及这三位讨论参与者转发（announce）他们的帖文。

然而，任何人都可以对 `@the_mighty_zork` 的帖文进行点赞。

这可以通过为一篇可见行为“公开”的帖文设置如下 `interactionPolicy` 来实现：

```json
{
  "@context": [
    "https://gotosocial.org/ns",
    [...]
  ],
  [...],
  "interactionPolicy": {
    "canLike": {
      "always": "https://www.w3.org/ns/activitystreams#Public"
    },
    "canReply": {
      "always": [
        "https://example.org/users/the_mighty_zork",
        "https://example.org/users/booblover6969",
        "https://example.org/users/hodor"
      ],
      "approvalRequired": "https://www.w3.org/ns/activitystreams#Public"
    },
    "canAnnounce": {
      "always": [
        "https://example.org/users/the_mighty_zork",
        "https://example.org/users/the_mighty_zork/followers",
        "https://example.org/users/booblover6969",
        "https://example.org/users/hodor"
      ]
    }
  },
  [...]
}
```

### 2. 单人长篇讨论串

在这个示例中，用户 `@the_mighty_zork` 想要写一段长篇讨论。

他们不介意别人转发和点赞讨论串中的帖文，但不想收到任何回复，因为他们没有精力去管理讨论；他们只是想发发牢骚。

这可以通过在讨论串中的每个帖文都设置如下 `interactionPolicy` 实现：

```json
{
  "@context": [
    "https://gotosocial.org/ns",
    [...]
  ],
  [...],
  "interactionPolicy": {
    "canLike": {
      "always": "https://www.w3.org/ns/activitystreams#Public"
    },
    "canReply": {
      "always": "https://example.org/users/the_mighty_zork"
    },
    "canAnnounce": {
      "always": "https://www.w3.org/ns/activitystreams#Public"
    }
  },
  [...]
}
```

在这里，任何人都允许点赞或转发，但除了 `@the_mighty_zork` 自己以外，没人允许回复。

### 3. 完全开放

在这个示例中，`@the_mighty_zork` 希望发表一条完全开放的帖文，以便任何能看到它的人都可以回复、转发或点赞：

```json
{
  "@context": [
    "https://gotosocial.org/ns",
    [...]
  ],
  [...],
  "interactionPolicy": {
    "canLike": {
      "always": "https://www.w3.org/ns/activitystreams#Public"
    },
    "canReply": {
      "always": "https://www.w3.org/ns/activitystreams#Public"
    },
    "canAnnounce": {
      "always": "https://www.w3.org/ns/activitystreams#Public"
    }
  },
  [...]
}
```

## 后续回复／扩大范围

讨论中的每条后续回复都有其各自的互动规则，由创建该回复的用户设定。换言之，整个*讨论串*或*主题*并不由单一的 `interactionPolicy` 控制，每个帖文作者可以为其后续帖文设置不同的规则。

不幸的是，这意味着即使设置了 `interactionPolicy`，讨论串的范围有时也会无意中超出第一帖作者的预期。

例如，在上面的[示例 1 - 限制讨论范围](#1-限制讨论范围)中，`@the_mighty_zork` 在首个帖文中设置了如下 `canReply.always` 值：

```json
[
  "https://example.org/users/the_mighty_zork",
  "https://example.org/users/booblover6969",
  "https://example.org/users/hodor"
]
```

而在随后的某个回复中，可能因疏忽或刻意为之，`@booblover6969` 将 `canReply.always` 值设置为：

```json
[
  "https://www.w3.org/ns/activitystreams#Public"
]
```

此举扩大了讨论范围，因为现在任何人都可以回复 `@booblover6969` 的帖文，并可能在回复中提及 `@the_mighty_zork`。

为了避免这种情况，建议外站实例防止用户扩大讨论范围（具体的实现机制有待确定）。

同时，实例也应把任何处于待批准状态的帖文（包含互动）也视作待批准状态。

换言之，实例应将所有处于待批准状态的上级帖文之下的互动也标记为待批准状态，不论该待批准的上级帖文的互动规则是否允许那些互动。

这样可以避免以下情况：某人回复一个帖文，即便他们的回复正待批准，但随后他们可对自己的回复继续回复，从而利用自身作为作者的[隐含允许回复的权限](#隐含假设)使其回复被标记为允许。

## 互动验证

[互动规则](#互动规则)部分描述了互动规则的格式、假定默认值以及相关假设。

本节描述互动规则的执行和验证，即设置互动规则的服务端如何发送批准或拒绝消息，以回应请求／待批准的互动，以及外站服务器如何证明互动者已获得互动对象对互动目标帖文的批准。

### 请求、获取和验证批准

当某个行为体的 URI 存在于某种互动类型的 `approvalRequired` 数组中，**或者**需要通过验证其在某集合中的存在（参见[验证在粉丝或关注集合中的存在](#验证在粉丝或关注集合中的存在)），在行为体希望对某条受互动规则限制的帖文请求批准互动时，服务端实现应当执行以下步骤：

1. 按常规构造该互动 `Activity`（例如 `Like`、`Create`，或 `Announce`）。
2. 按常规将该 `Activity` 的 `to` 和 `cc` 指定为预期的活动接收方。
3. 仅将该 `Activity` 以 `POST` 方式发送至互动目标帖文作者的 `Inbox`（或 `sharedInbox`）。
4. **此时不要再对该 Activity 执行进一步分发**。

在此阶段，该互动可视为*待批准*状态，不应显示在被互动帖文的回复或点赞等集合中。

它可以以“互动待批准”的模式显示给发送该互动的用户，但理想情况下不应显示给与该用户同一实例的其他用户。

从这一点开始，可能出现以下三种情况之一：

#### 拒绝

在这种情况下，互动目标帖文的作者所在的服务器将会发送一个 `Reject` 类型的 `Activity`，其 `object` 属性为待批准互动的 URI/ID。

例如，下面这个 JSON 对象拒绝了 `@someone@somewhere.else.example.org` 试图回复 `@post_author@example.org` 帖文的请求：

```json
{
  "@context": "https://www.w3.org/ns/activitystreams",
  "actor": "https://example.org/users/post_author",
  "to": "https://somewhere.else.example.org/users/someone",
  "id": "https://example.org/users/post_author/activities/reject/01J0K2YXP9QCT5BE1JWQSAM3B6",
  "object": "https://somewhere.else.example.org/users/someone/statuses/01J17XY2VXGMNNPH1XR7BG2524",
  "type": "Reject"
}
```

若发生这种情况，`@someone@somewhere.else.example.org`（以及其所在实例）应将该互动视为已被拒绝。实例应将其内部存储（例如数据库）中的对应活动删除，或以其他方式标记为已拒绝，并且不再进一步分发或重试该互动。服务器可能希望通知互动发起者他们的互动已被拒绝。

#### 无应答

在这种情况下，被互动帖文的作者既没有返回 `Reject` 也没有返回 `Accept` 类型的 `Activity`。在这种情况下，该互动将无限期地处于“待批准”状态。实现者可以考虑实现某种清理机制，将达到一定时间而依然处于待批准状态的发送或待批准互动视为已失效或被拒绝，然后以之前提到的方式移除。

#### 接受

在这种情况下，互动目标帖文的作者会发送一个 `Accept` 类型的 `Activity`，其 `object` 属性为待批准互动的 URI/ID，同时其 `result` 属性中包含一个可解引用的批准对象 URI（详见[批准对象](#批准对象)）。

例如，下面这个 JSON 对象接受了 `@someone@somewhere.else.example.org` 试图回复 `@post_author@example.org` 帖文的请求：

```json
{
  "@context": "https://www.w3.org/ns/activitystreams",
  "actor": "https://example.org/users/post_author",
  "cc": [
    "https://www.w3.org/ns/activitystreams#Public",
    "https://example.org/users/post_author/followers"
  ],
  "to": "https://somewhere.else.example.org/users/someone",
  "id": "https://example.org/users/post_author/activities/accept/01J0K2YXP9QCT5BE1JWQSAM3B6",
  "object": "https://somewhere.else.example.org/users/someone/statuses/01J17XY2VXGMNNPH1XR7BG2524",
  "result": "https://example.org/users/post_author/reply_approvals/01JMMGABRDNA9G9BDNYJR7TC8D",
  "type": "Accept"
}
```

若发生这种情况，`@someone@somewhere.else.example.org`（以及其所在实例）应将该互动视为已获得被互动者的批准／接受。

此时，`somewhere.else.example.org` 应再次分发该互动，但有以下区别：

1. 这次需要在发送的 `Create` 活动中包含 `Accept` 消息中给出的 `result` 的 URI/ID，并将其放入 `approvedBy` 字段中。
2. 这次将该互动分发给 `to`、`cc` 等所有目标收件人。

!!! note "注意"
    虽然不是严格必须，但在上例中，行为体 `https://example.org/users/post_author` 不仅将 `Accept` 消息的接收方设为互动发起者 `https://somewhere.else.example.org/users/someone`，还额外包含了他们的粉丝集合（以及隐含地包含了公共地址）。这使得其他服务器上的 `https://example.org/users/post_author` 的粉丝，也可以标记该互动为已接受，并在不必解引用 `approvedBy` URI 的情况下，将该互动与被互动帖文一同展示。

### 批准对象

批准对象是基本 ActivityStreams 对象的扩展，其类型可以是 `LikeApproval`、`ReplyApproval` 或 `AnnounceApproval`。每种类型对应一个特定互动类型的批准。

例如，`LikeApproval`：

```json
{
  "@context": [
    "https://www.w3.org/ns/activitystreams",
    "https://gotosocial.org/ns"
  ],
  "attributedTo": "https://example.org/users/post_author",
  "id": "https://example.org/users/post_author/approvals/01J0K2YXP9QCT5BE1JWQSAM3B6",
  "object": "https://somewhere.else.example.org/users/someone/likes/01JMPKG79EAH0NB04BHEM9D20N",
  "target": "https://example.org/users/post_author/statuses/01JJYV141Y5M4S65SC1XCP65NT",
  "type": "LikeApproval"
}
```

`ReplyApproval`：

```json
{
  "@context": [
    "https://www.w3.org/ns/activitystreams",
    "https://gotosocial.org/ns"
  ],
  "attributedTo": "https://example.org/users/post_author",
  "id": "https://example.org/users/post_author/approvals/01J0K2YXP9QCT5BE1JWQSAM3B6",
  "object": "https://somewhere.else.example.org/users/someone/statuses/01J17XY2VXGMNNPH1XR7BG2524",
  "target": "https://example.org/users/post_author/statuses/01JJYV141Y5M4S65SC1XCP65NT",
  "type": "ReplyApproval"
}
```

`AnnounceApproval`：

```json
{
  "@context": [
    "https://www.w3.org/ns/activitystreams",
    "https://gotosocial.org/ns"
  ],
  "attributedTo": "https://example.org/users/post_author",
  "id": "https://example.org/users/post_author/approvals/01J0K2YXP9QCT5BE1JWQSAM3B6",
  "object": "https://somewhere.else.example.org/users/someone/boosts/01JMPKG79EAH0NB04BHEM9D20N",
  "target": "https://example.org/users/post_author/statuses/01JJYV141Y5M4S65SC1XCP65NT",
  "type": "AnnounceApproval"
}
```

在一个批准对象中：

- `attributedTo`：应为发出 `Accept` 消息的行为体，也即互动对象。
- `object`：应为进行互动的 `Like`、`Announce` 或帖文类别的 `Object`。
- `target`（可选）：如果包含，应为被互动的帖文类别的 `Object`。

!!! info "批准对象应当可以被解引用"
    根据验证机制（参见[验证 `approvedBy`](#验证-approvedby)），各实例应确保对批准对象 URI 的解引用返回合法的 ActivityPub 响应。否则，外站实例在分发帖文时可能会受限。

### `approvedBy`

`approvedBy` 是附加在 `Like`、`Announce` 活动中，以及任何被视为“帖文”（例如 `Note`、`Article` 等）的对象上的额外属性。

`approvedBy` 表示该互动（或回复对象）已获得目标帖文作者的批准／接受，从而现在可以分发给预期的受众。

`approvedBy` 的值应为在 `Accept` 消息中发送的 `result` URI/ID，该 URI 指向一个可解引用的批准对象。

例如，下面这个 `Announce` 活动通过存在 `approvedBy` 表明其已被 `@post_author@example.org` 接受：

```json
{
  "actor": "https://somewhere.else.example.org/users/someone",
  "to": [
    "https://somewhere.else.example.org/users/someone/followers"
  ],
  "cc": [
    "https://example.org/users/post_author"
  ],
  "id": "https://somewhere.else.example.org/users/someone/activities/announce/01J0K2YXP9QCT5BE1JWQSAM3B6",
  "object": "https://example.org/users/post_author/statuses/01J17ZZFK6W82K9MJ9SYQ33Y3D",
  "approvedBy": "https://example.org/users/post_author/reply_approvals/01JMMGABRDNA9G9BDNYJR7TC8D",
  "type": "Announce"
}
```

#### 验证 `approvedBy`

当接收到带有 `approvedBy` 值的活动或帖文对象时，外站实例应：

1. 验证 `approvedBy` URI 的主机名／域名与被互动帖文作者的主机名／域名一致。
2. 解引用 `approvedBy` URI/ID 以获得批准对象（见[批准对象](#approval-objects)）。
3. 检查批准对象的类型，确保其正确，即例如一个 `Announce` 消息的 `approvedBy` URI 应指向 `AnnounceApproval`，而不是 `ReplyApproval` 或 `LikeApproval`。
4. 检查批准对象中的 `attributedTo` 值是否与进行互动的行为体的 URI/ID 一致。
5. 检查批准对象中的 `object` 值是否与互动 `Activity` 或对象的 `id` 一致。

如果批准对象无法解引用，或者未通过上述有效性检查，则应将该互动视为无效并丢弃。

!!! warning "警告"
    GoToSocial 版本 0.17.x 和 0.18.x 没有包含指向批准对象的 `result`，而是直接在 `approvedBy` 中发送了 `Accept` 的 URI/ID。
    
    GoToSocial 版本 0.18.x 部分地向批准对象类型提供了向前兼容，因为它可以对解引用得到的 `Accept` 或批准对象进行验证，同时仍在 `approvedBy` 字段中发送 `Accept` 的 URI。
    
    GoToSocial 版本 0.19.x 及更高版本将按照本文档所述发送指向批准对象的 `approvedBy`，而不是发送 `Accept` 的 URI。

### 验证在粉丝或关注集合中的存在

如果一个行为体（通过 `Like`、`inReplyTo` 或 `Announce`）对一个对象进行互动，而其权限依赖于其出现在 `interactionPolicy` 中 `always` 字段里 Followers 或 Following 集合，则其服务器**仍应等待**目标行为体的服务器发出 `Accept` 消息后，再将该互动以带有 `approvedBy` 属性（值为批准 URI/ID）的形式广泛分发。

这是为了防止第三方服务器需要以某种方式验证进行互动的行为体是否存在于被互动行为体的粉丝或关注集合中。让目标服务器来做验证，并信任其隐式批准互动的行为体存在于相应集合中会更简单。

同理，当接收到一个行为体的互动，且其权限与 `always` 属性中关注或粉丝集合中的某一项匹配时，被互动行为体所在的服务器应**总是**确保尽快发送 `Accept` 消息，以便发起互动行为体所在的服务器可以带上适当的批准证明分发该互动。

这一过程应该绕过正常的"待批准"阶段，也就是说，被互动行为体所在的服务器无需通知被互动行为体有待处理的交互，并等待行为体接受或拒绝，因为行为体实际上已经明确同意这些交互。在 GoToSocial 的代码库中，这一过程称为“预先批准”。

### 可选行为

本节描述了在发送 `Accept` 和 `Reject` 消息时实现者*可能*使用、以及在接收时应考虑的可选行为。

#### 总是发送 `Accept` 消息

实现者可能希望：即使根据所在 `always` 数组，互动行为已被默认或显式允许，也要向外站互动者发送一个 `Accept` 消息。当接收到这样的 `Accept` 时，实现者可能仍希望更新其互动记录，将 `approvedBy` URI 更新为指向批准对象。这在以后处理撤回（TODO）时可能会有所帮助。

#### 类型提示：`Accept` 和 `Reject` 的内联 `object` 属性

如果需要，实现者可以部分展开／内联 `Accept` 或 `Reject` 消息的 `object` 属性，以向外站服务器提示即将被接受或拒绝的互动类型。当以这种方式内联时，`object` 中至少必须定义 `type` 和 `id`。例如：

```json
{
  "@context": "https://www.w3.org/ns/activitystreams",
  "actor": "https://example.org/users/post_author",
  "cc": [
    "https://www.w3.org/ns/activitystreams#Public",
    "https://example.org/users/post_author/followers"
  ],
  "to": "https://somewhere.else.example.org/users/someone",
  "id": "https://example.org/users/post_author/activities/reject/01J0K2YXP9QCT5BE1JWQSAM3B6",
  "object": {
    "type": "Note",
    "id": "https://somewhere.else.example.org/users/someone/statuses/01J17XY2VXGMNNPH1XR7BG2524",
    [...]
  },
  "result": "https://example.org/users/post_author/approvals/01JMPS01E54DG9JCF2ZK3JDMXE",
  "type": "Accept"
}
```

#### 在 `Accept` 和 `Reject` 消息中设定 `target` 属性

如果需要，实现者可以在发出的 `Accept` 或 `Reject` 消息中设置 `target` 属性，其值为互动目标帖文的 `id`，以便外站服务器更容易理解所接受或拒绝的互动的形状和关联性。

例如，下面这个 JSON 对象接受了 `@someone@somewhere.else.example.org` 试图回复 id 为 `https://example.org/users/post_author/statuses/01JJYV141Y5M4S65SC1XCP65NT` 的帖文的互动：

```json
{
  "@context": "https://www.w3.org/ns/activitystreams",
  "actor": "https://example.org/users/post_author",
  "cc": [
    "https://www.w3.org/ns/activitystreams#Public",
    "https://example.org/users/post_author/followers"
  ],
  "to": "https://somewhere.else.example.org/users/someone",
  "id": "https://example.org/users/post_author/activities/reject/01J0K2YXP9QCT5BE1JWQSAM3B6",
  "object": "https://somewhere.else.example.org/users/someone/statuses/01J17XY2VXGMNNPH1XR7BG2524",
  "target": "https://example.org/users/post_author/statuses/01JJYV141Y5M4S65SC1XCP65NT",
  "result": "https://example.org/users/post_author/approvals/01JMPS01E54DG9JCF2ZK3JDMXE",
  "type": "Accept"
}
```

如果需要，`target` 属性也可部分展开／内联以提示互动目标帖文的类型。在内联时，`target` 至少必须定义 `type` 和 `id`。例如：

```json
{
  "@context": "https://www.w3.org/ns/activitystreams",
  "actor": "https://example.org/users/post_author",
  "cc": [
    "https://www.w3.org/ns/activitystreams#Public",
    "https://example.org/users/post_author/followers"
  ],
  "to": "https://somewhere.else.example.org/users/someone",
  "id": "https://example.org/users/post_author/activities/reject/01J0K2YXP9QCT5BE1JWQSAM3B6",
  "object": "https://somewhere.else.example.org/users/someone/statuses/01J17XY2VXGMNNPH1XR7BG2524",
  "target": {
    "type": "Note",
    "id": "https://example.org/users/post_author/statuses/01JJYV141Y5M4S65SC1XCP65NT"
    [ ... ]
  },
  "result": "https://example.org/users/post_author/approvals/01JMPS01E54DG9JCF2ZK3JDMXE",
  "type": "Accept"
}
```

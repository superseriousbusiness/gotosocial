# HTTP Request Header Filtering Modes

GoToSocial currently offers 'block', 'allow' and disabled HTTP request header filtering modes, which can be set using the `advanced-header-filter-mode` setting in the config.yaml, or using the `GTS_ADVANCED_HEADER_FILTER_MODE` environment variable. These are described below.

!!! warning
    HTTP request header filtering is an advanced setting. If you are not well versed in the uses and intricacies of HTTP request headers, you may break federation or even access to your own instance by changing these.

    HTTP request header filtering is also still considered "experimental". It should do what it says on the box, but it may cause bugs or edge cases to appear elsewhere, we're not sure yet!

    Management via settings panel is TBA. Until then you will need to manage these directly via API endpoints.

## Disabled header filtering mode (default)

When `advanced-header-filter-mode` is set to `""`, i.e. an empty string, all request header filtering will be disabled.

## Block filtering mode

When `advanced-header-filter-mode` is set to `"block"`, your instance will accept HTTP requests as normal (pending API token checks, HTTP signature checks etc), with the exception of matching block header filters you have explicitly created via the settings panel.

In block mode, an allow header filter can be used to override an existing block filter, providing an extra level of granularity.

A request in block mode will be accepted if it is EXPLICITLY ALLOWED OR NOT EXPLICITLY BLOCKED.

## Allow filtering mode

When `advanced-header-filter-mode` is set to `"allow"`, your instance will only accept HTTP requests for which a matching allow header filter has been explicitly created via the settings panel. All other requests will be refused.

In allow mode, a block header filter can be used to override an existing allow filter, providing an extra level of granularity.

A request in allow mode will only be accepted if it is EXPLICITLY ALLOWED AND NOT EXPLICITLY BLOCKED.

!!! danger
    Allow filtering mode is an extremely restrictive mode that will almost certainly prevent many (legitimate) clients from being able to access your instance, including yourself. You should only enable this mode if you know exactly what you're trying to achieve.
# Feature Gates

The runtime package contains a feature gate used to ease the migration
from the [previous runtime metrics conventions] to the new [OpenTelemetry Go
Runtime conventions].

Note that the new runtime metrics conventions are still experimental, and may
change in backwards incompatible ways as feedback is applied.

## Features

- [Include Deprecated Metrics](#include-deprecated-metrics)

### Include Deprecated Metrics

Once new experimental runtime metrics are added, they will be produced
**in addition to** the existing runtime metrics. Users that migrate right away
can disable the old runtime metrics:

```console
export OTEL_GO_X_DEPRECATED_RUNTIME_METRICS=false
```

In a later release, the deprecated runtime metrics will stop being produced by
default. To temporarily re-enable the deprecated metrics:

```console
export OTEL_GO_X_DEPRECATED_RUNTIME_METRICS=true
```

After two additional releases, the deprecated runtime metrics will be removed,
and setting the environment variable will no longer have any effect.

The value set must be the case-insensitive string of `"true"` to enable the
feature, and `"false"` to disable the feature. All other values are ignored.

[previous runtime metrics conventions]: https://pkg.go.dev/go.opentelemetry.io/contrib/instrumentation/runtime@v0.52.0
[OpenTelemetry Go Runtime conventions]: https://github.com/open-telemetry/semantic-conventions/blob/main/docs/runtime/go-metrics.md

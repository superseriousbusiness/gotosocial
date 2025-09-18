# Metrics

GoToSocial uses the [OpenTelemetry][otel] Go SDK to enable instance admins to expose runtime metrics in the Prometheus metrics format.

Currently, the following metrics are collected:

* Go performance and runtime metrics
* Gin (HTTP server) metrics
* Bun (database) metrics

## Enabling metrics

To enable metrics, first set the `metrics-enabled` configuration value to `true` in your config.yaml file:

```yaml
metrics-enabled: true
```

Then, you will need to set some additional environment variables on the GoToSocial process in order to configure OpenTelemetry to expose metrics in the Prometheus format:

```env
OTEL_METRICS_PRODUCERS=prometheus
OTEL_METRICS_EXPORTER=prometheus
```

By default, this configuration will instantiate an additional HTTP server running alongside the standard GoToSocial HTTP server, which exposes a Prometheus metrics endpoint at `localhost:9464/metrics`.

!!! tip
    If you are running GoToSocial using the [example systemd service definition](../../example/gotosocial.service), you can easily set these environment variables by uncommenting the relevant two lines in that file, and reloading + restarting the service.

If you wish, you can further customize this metrics HTTP server by using the following environment variables to change the host and port:

```env
OTEL_EXPORTER_PROMETHEUS_HOST=example.org
OTEL_EXPORTER_PROMETHEUS_PORT=9999
```

## Serving metrics to the outside world

If you have deployed GoToSocial in a "bare-metal" fashion without a reverse proxy, you can expose the metrics endpoint to the outside world by setting `OTEL_EXPORTER_PROMETHEUS_HOST` to your host value. For example, if your GtS instance `host` configuration value is set to `example.org`, you should set `OTEL_EXPORTER_PROMETHEUS_HOST=example.org`. You should then be able to access your metrics at `http://example.org:9464/metrics`. GoToSocial running in this fashion will not serve LetsEncrypt certificates at the metrics endpoint, so you will be limited to using HTTP rather than HTTPS.

If you are using a reverse proxy like Nginx, you can expose the metrics endpoint to the outside world with HTTPS certificates, by putting an additional location stanza in your Nginx configuration above the catch-all `location /` stanza:

```nginx
location /metrics {
  proxy_pass http://127.0.0.1:9464;
}
```

This will instruct Nginx to forward requests to `example.org/metrics` to the separate Prometheus server running on port 9464.

## Enabling basic authentication

Although there is no sensitive data contained in the OTEL runtime statistics exported by Prometheus, you may nevertheless wish to gate access to the `/metrics` endpoint behind some kind of authentication, to prevent every Tom, Dick, and Harry from looking at your runtime stats.

You can do this by configuring your reverse proxy to require basic authentication for access to `/metrics`.

In Nginx, for example, you could do this by creating an `htpasswd` file alongside your site in the `sites-available` directory of Nginx, and instructing Nginx to use that file to gate access.

Assuming you followed the [guide for setting up Nginx as your reverse proxy](../getting_started/reverse_proxy/nginx.md), you will already have a file for your Nginx service definition at `/etc/nginx/sites-available/example.org`, where `example.org` is the hostname of your instance.

You can create an `htpasswd` file alongside this file using the following command:

```bash
htpasswd -c /etc/nginx/sites-available/example.org.htpasswd username
```

In the command, be sure to replace `example.org` with your hostname, and `username` with whatever username you want to use.

Now, edit `/etc/nginx/sites-available/example.org` and update your `/metrics` stanza to use the `httpasswd` file. After editing it should look something like this:

```nginx
location /metrics {
  proxy_pass http://127.0.0.1:9464;
  auth_basic           "Metrics";
  auth_basic_user_file /etc/nginx/sites-available/example.org.htpasswd;
}
```

Again, replace `example.org` in that snippet with your instance hostname.

When you're finished editing, reload + restart Nginx, and you should see a basic authentication prompt when visiting the `/metrics` endpoint of your instance in your browser.

## Prometheus scrape configuration 

You can scrape your `/metrics` endpoint with a Prometheus instance using the following configuration in your `scrape_configs`:

```yaml
- job_name: gotosocial
  metrics_path: /metrics
  scheme: https
  basic_auth:
    username: some_username
    password: some_password
  static_configs:
  - targets:
    - example.org
```

Change `example.org` to your hostname in the above snippet. If you are not using HTTPS, change the `scheme` value to `http`. If you are not using basic authentication, you can remove the `basic_auth` section. If you are not using a reverse proxy, and metrics are exposed on port 9464, add the port to the host (eg., `example.org` -> `example.org:9464`).

## Viewing metrics on Grafana

Instructions on how to set up Grafana are beyond the scope of this document. However, once you have set up a Grafana to pull from your Prometheus instance, you can import the [example Grafana dashboard](https://codeberg.org/superseriousbusiness/gotosocial/raw/branch/main/example/metrics/gotosocial_grafana_dashboard.json) into your Grafana frontend to easily view GoToSocial Go runtime and HTTP metrics.

[otel]: https://opentelemetry.io/
[prom]: https://prometheus.io/docs/instrumenting/exposition_formats/
[obs]: ../configuration/observability_and_metrics.md

# Health Checks

GoToSocial exposes two health check HTTP endpoints: `/readyz` and `/livez`.

These can be used to check whether GoToSocial is reachable and able to make simple database queries.

`/livez` will always return a 200 OK response with no body, in response to both GET and HEAD requests. This is useful to check if the GoToSocial service is alive.

`/readyz` will return a 200 OK response with no body, in response to both GET and HEAD requests, if GoToSocial is able to run a very simple SELECT query against the configured database backend. If an error occurs while running the SELECT, the error will be logged, and 500 Internal Server Error will be returned, with no body.

You can use the above endpoints to implement health checks in container runtimes / orchestration systems.

For example, in a Docker setup, you could add the following to your docker-compose.yaml:

```yaml
healthcheck:
  test: wget --no-verbose --tries=1 --spider http://localhost:8080/readyz || exit 1
  interval: 120s
  retries: 5
  start_period: 30s
  timeout: 10s
```

The above health check will start after 30 seconds, and check every two minutes whether the service is available by doing a HEAD request to `/readyz`. If the check fails five times in a row, the service will be reported as unhealthy. You can use this in whatever orchestration system you are using to force the container to restart.

!!! warning
    When doing database migrations on slow hardware, migration might take longer than the 10 minutes afforded by the above health check.
    
    On such a system, you may want to increase the interval or number of retries of the health check to ensure that you don't stop GoToSocial in the middle of a migration (which is a very bad thing to do!).

!!! tip
    Though the health check endpoints don't reveal any sensitive info, and run only very simple queries, you may want to avoid exposing them to the outside world. You could do this in nginx, for example, by adding the following snippet to your `server` stanza:
    
    ```nginx
    location /livez {
      return 404;
    }
    location /readyz {
      return 404;
    }
    ```
    
    This will cause nginx to intercept these requests *before* they are passed to GoToSocial, and just return 404 Not Found.

References:

- [Dockerfile reference](https://docs.docker.com/reference/dockerfile/#healthcheck)
- [Compose file reference](https://docs.docker.com/compose/compose-file/compose-file-v3/#healthcheck)

Create http server listening on unix sockets and systemd socket activated fds

## Quick Usage

    go get go.balki.me/anyhttp

Just replace `http.ListenAndServe` with `anyhttp.ListenAndServe`.

```diff
- http.ListenAndServe(addr, h)
+ anyhttp.ListenAndServe(addr, h)
```

## Address Syntax

### Unix socket

Syntax

    unix/<path to socket>

Examples

    unix/relative/path.sock
    unix//var/run/app/absolutepath.sock

### Systemd Socket activated fd:

Syntax

    sysd/fdidx/<fd index starting at 0>
    sysd/fdname/<fd name set using FileDescriptorName socket setting >

Examples:
    
    # First (or only) socket fd passed to app
    sysd/fdidx/0

    # Socket with FileDescriptorName
    sysd/fdname/myapp

    # Using default name
    sysd/fdname/myapp.socket

### TCP port

If the address is a number less than 65536, it is assumed as a port and passed
as `http.ListenAndServe(":<port>",...)` Anything else is directly passed to
`http.ListenAndServe` as well. Below examples should work

    :http
    :8888
    127.0.0.1:8080

## Idle server auto shutdown

When using systemd socket activation, idle servers can be shut down to save on
resources.  They will be restarted with socket activation when new request
arrives. Quick example for the case. (Error checking skipped for brevity)

```go
addrType, httpServer, done, _ := anyhttp.Serve(addr, idle.WrapHandler(nil))
if addrType == anyhttp.SystemdFD {
    idle.Wait(30 * time.Minute)
    httpServer.Shutdown(context.TODO())
}
<-done
```

## Documentation

https://pkg.go.dev/go.balki.me/anyhttp

### Related links

  * https://gist.github.com/teknoraver/5ffacb8757330715bcbcc90e6d46ac74#file-unixhttpd-go
  * https://github.com/coreos/go-systemd/tree/main/activation

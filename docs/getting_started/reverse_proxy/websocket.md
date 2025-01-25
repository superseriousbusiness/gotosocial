# WebSocket

GoToSocial uses the secure [WebSocket protocol](https://en.wikipedia.org/wiki/WebSocket) (aka `wss`) to allow for streaming updates of statuses and notifications via client apps like Pinafore.

In order to use this functionality, you need to ensure that whatever proxy you've configured GoToSocial to run behind allows WebSocket connections through.

The WebSocket endpoint is located at `wss://example.org/api/v1/streaming` where `example.org` is the hostname of your GoToSocial instance.

The WebSocket endpoint uses the same port as configured in the `port` section of your [general config](../../configuration/general.md).

Typical WebSocket **request** headers as sent by Pinafore look like the following:

```text
Host: example.org
User-Agent: Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:99.0) Gecko/20100101 Firefox/99.0
Accept: */*
Accept-Language: en-US,en;q=0.5
Accept-Encoding: gzip, deflate, br
Sec-WebSocket-Version: 13
Origin: https://pinafore.social
Sec-WebSocket-Protocol: null
Sec-WebSocket-Extensions: permessage-deflate
Sec-WebSocket-Key: YWFhYWFhYm9vYmllcwo=
DNT: 1
Connection: keep-alive, Upgrade
Sec-Fetch-Dest: websocket
Sec-Fetch-Mode: websocket
Sec-Fetch-Site: cross-site
Pragma: no-cache
Cache-Control: no-cache
Upgrade: websocket
```

Typical WebSocket **response** headers as returned by GoToSocial look like the following:

```text
HTTP/1.1 101 Switching Protocols
Upgrade: websocket
Connection: Upgrade
Sec-WebSocket-Accept: WVdGaFlXRmhZbTl2WW1sbGN3bz0K
```

Whatever your setup, you need to ensure that these headers are allowed through your proxy, which may require extra configuration depending on the exact proxy being used.
